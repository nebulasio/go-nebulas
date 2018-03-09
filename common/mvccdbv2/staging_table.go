// Copyright (C) 2018 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see <http://www.gnu.org/licenses/>.
//

package mvccdbv2

import (
	"bytes"
	"errors"
	"sync"

	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"

	"github.com/nebulasio/go-nebulas/util/byteutils"
)

var (
	ErrStagingTableKeyConfliction = errors.New("staging table key confliction")
	ErrParentStagingTableIsNil    = errors.New("parent Staging Table is nil")
)

type stagingValuesMap map[string]*VersionizedValueItem
type stagingValuesMapMap map[interface{}]stagingValuesMap

// VersionizedValueItem a struct for key/value pair, with version, dirty, deleted flags.
type VersionizedValueItem struct {
	tid     interface{}
	key     []byte
	val     []byte
	old     int
	new     int
	deleted bool
	dirty   bool
}

// StagingTable a struct to store all staging changed key/value pairs.
// There are two map to store the key/value pairs. One are stored associated with tid,
// the other is `finalVersionizedValue`, record the `ready to commit` key/value pairs.
type StagingTable struct {
	storage               storage.Storage
	allVersionizedValues  map[interface{}]stagingValuesMap
	tidMutex              sync.Mutex
	tid                   interface{}
	versionizedValues     stagingValuesMap
	mutex                 sync.Mutex
	preparedStagingTables map[interface{}]*StagingTable
	parentStagingTable    *StagingTable
}

// NewStagingTable return new instance of StagingTable.
func NewStagingTable(storage storage.Storage, tid interface{}) *StagingTable {
	tbl := &StagingTable{
		storage:               storage,
		allVersionizedValues:  make(map[interface{}]stagingValuesMap),
		tid:                   tid,
		versionizedValues:     make(stagingValuesMap),
		preparedStagingTables: make(map[interface{}]*StagingTable),
		parentStagingTable:    nil,
	}
	tbl.allVersionizedValues[tid] = tbl.versionizedValues
	return tbl
}

func (tbl *StagingTable) Prepare(tid interface{}) (*StagingTable, error) {
	tbl.mutex.Lock()
	defer tbl.mutex.Unlock()

	if tbl.allVersionizedValues[tid] != nil {
		return nil, ErrTidIsExist
	}

	preparedTbl := &StagingTable{
		storage:               tbl.storage,
		allVersionizedValues:  tbl.allVersionizedValues,
		tid:                   tid,
		versionizedValues:     make(stagingValuesMap),
		preparedStagingTables: make(map[interface{}]*StagingTable),
		parentStagingTable:    tbl,
	}

	tbl.allVersionizedValues[tid] = tbl.versionizedValues
	tbl.preparedStagingTables[tid] = preparedTbl
	return preparedTbl, nil
}

// Get return value by key. If key does not exist, copy and incr version from `parentStagingTable` to record previous version.
func (tbl *StagingTable) Get(key []byte) (*VersionizedValueItem, error) {
	tbl.mutex.Lock()
	defer tbl.mutex.Unlock()

	keyStr := byteutils.Hex(key)
	value := tbl.versionizedValues[keyStr]

	if value == nil {
		var err error
		if tbl.parentStagingTable != nil {
			value, err = tbl.parentStagingTable.Get(key)
			if err != nil {
				return nil, err
			}
			value = IncrVersionizedValueItem(tbl.tid, value)
		} else {
			// load from storage.
			value, err = tbl.loadFromStorage(key)
			if err != nil && err != storage.ErrKeyNotFound {
				return nil, err
			}
		}
		tbl.versionizedValues[keyStr] = value
	}

	return value, nil
}

// Put put the key/val pair. If key does not exist, copy and incr version from `parentStagingTable` to record previous version.
func (tbl *StagingTable) Put(key []byte, val []byte) (*VersionizedValueItem, error) {
	value, err := tbl.Get(key)
	if err != nil {
		return nil, err
	}

	value.dirty = value.dirty || bytes.Compare(value.val, val) != 0
	value.val = val
	return value, nil
}

// Set set the tid/key/val pair. If tid+key does not exist, copy and incr version from `finalVersionizedValues` to record previous version.
func (tbl *StagingTable) Set(key []byte, val []byte, deleted, dirty bool) (*VersionizedValueItem, error) {
	value, err := tbl.Get(key)
	if err != nil {
		return nil, err
	}

	value.val = val
	value.deleted = deleted
	value.dirty = dirty
	return value, nil
}

// Del del the tid/key pair. If tid+key does not exist, copy and incr version from `finalVersionizedValues` to record previous version.
func (tbl *StagingTable) Del(key []byte) (*VersionizedValueItem, error) {
	value, err := tbl.Get(key)
	if err != nil {
		return nil, err
	}

	value.deleted = true
	value.dirty = true
	return value, nil
}

// Purge purge key/value pairs of tid.
func (tbl *StagingTable) Purge() {
	tbl.mutex.Lock()
	defer tbl.mutex.Unlock()

	for _, p := range tbl.preparedStagingTables {
		p.Purge()
	}

	// purge all content.
	tbl.versionizedValues = make(stagingValuesMap)
}

// MergeToParent merge key/value pair of tid to `finalVersionizedValues` which the version of value are the same.
func (tbl *StagingTable) MergeToParent() ([]interface{}, error) {
	if tbl.parentStagingTable == nil {
		return nil, ErrParentStagingTableIsNil
	}

	tbl.parentStagingTable.mutex.Lock()
	defer tbl.parentStagingTable.mutex.Unlock()

	tbl.mutex.Lock()
	defer tbl.mutex.Unlock()

	dependentTids := make(map[interface{}]bool)
	conflictKeys := make(map[string]interface{})

	// 1. check version.
	targetValues := tbl.parentStagingTable.versionizedValues

	for keyStr, fromValueItem := range tbl.versionizedValues {
		targetValueItem := targetValues[keyStr]

		// record conflict.
		if fromValueItem.old != targetValueItem.old {
			conflictKeys[keyStr] = targetValueItem.tid
			continue
		}

		// skip default value loaded from storage.
		if targetValueItem.isDefault() {
			continue
		}

		// ignore parent tid.
		if targetValueItem.tid == tbl.parentStagingTable.tid {
			continue
		}

		// record dependentTids.
		dependentTids[targetValueItem.tid] = true
	}

	if len(conflictKeys) > 0 {
		logging.VLog().WithFields(logrus.Fields{
			"tid":          tbl.tid,
			"parentTid":    tbl.parentStagingTable.tid,
			"conflictKeys": conflictKeys,
		}).Debug("Check failed.")
		return nil, ErrStagingTableKeyConfliction
	}

	// 2. merge to final.
	for keyStr, fromValueItem := range tbl.versionizedValues {
		// ignore dirty.
		if !fromValueItem.dirty {
			continue
		}

		// merge.
		targetValues[keyStr] = fromValueItem.CloneForMerge()
	}

	tids := make([]interface{}, 0, len(dependentTids))
	for key := range dependentTids {
		tids = append(tids, key)
	}

	return tids, nil
}

func (tbl *StagingTable) GetVersionizedValues() stagingValuesMap {
	return tbl.versionizedValues
}

func (tbl *StagingTable) Lock() {
	tbl.mutex.Lock()
}

func (tbl *StagingTable) Unlock() {
	tbl.mutex.Unlock()
}

func (tbl *StagingTable) loadFromStorage(key []byte) (*VersionizedValueItem, error) {
	// get from storage.
	val, err := tbl.storage.Get(key)
	if err != nil && err != storage.ErrKeyNotFound {
		return nil, err
	}

	value := NewDefaultVersionizedValueItem(key, val, tbl.tid)
	return value, nil
}

func (value *VersionizedValueItem) isDefault() bool {
	return value.old == 0 && value.dirty == false
}

// NewDefaultVersionizedValueItem return new instance of VersionizedValueItem, old/new version are 0, dirty is false.
func NewDefaultVersionizedValueItem(key []byte, val []byte, tid interface{}) *VersionizedValueItem {
	return &VersionizedValueItem{
		tid:     tid,
		key:     key,
		val:     val,
		old:     0,
		deleted: false,
		dirty:   false,
	}
}

// IncrVersionizedValueItem copy and return the version increased VersionizedValueItem.
func IncrVersionizedValueItem(tid interface{}, oldValue *VersionizedValueItem) *VersionizedValueItem {
	return &VersionizedValueItem{
		tid:     tid,
		key:     oldValue.key,
		val:     oldValue.val,
		old:     oldValue.old,
		deleted: oldValue.deleted,
		dirty:   false,
	}
}

// CloneForMerge shadow copy of `VersionizedValueItem` with dirty is true.
func (value *VersionizedValueItem) CloneForMerge() *VersionizedValueItem {
	return &VersionizedValueItem{
		tid:     value.tid,
		key:     value.key,
		val:     value.val,
		old:     value.old + 1,
		deleted: value.deleted,
		dirty:   true,
	}
}
