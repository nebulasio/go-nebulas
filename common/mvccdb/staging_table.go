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

package mvccdb

import (
	"bytes"
	"errors"
	"sync"

	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"

	"github.com/nebulasio/go-nebulas/util/byteutils"
)

// Error
var (
	ErrStagingTableKeyConfliction = errors.New("staging table key confliction")
	ErrParentStagingTableIsNil    = errors.New("parent Staging Table is nil")
)

type stagingValuesMap map[string]*VersionizedValueItem
type stagingValuesMapMap map[interface{}]stagingValuesMap

// VersionizedValueItem a struct for key/value pair, with version, dirty, deleted flags.
type VersionizedValueItem struct {
	tid           interface{}
	key           []byte
	val           []byte
	version       int
	deleted       bool
	dirty         bool
	globalVersion int64
}

// StagingTable a struct to store all staging changed key/value pairs.
// There are two map to store the key/value pairs. One are stored associated with tid,
// the other is `finalVersionizedValue`, record the `ready to commit` key/value pairs.
type StagingTable struct {
	storage                         storage.Storage
	globalVersion                   int64
	parentStagingTable              *StagingTable
	versionizedValues               stagingValuesMap
	tid                             interface{}
	mutex                           sync.Mutex
	prepareingGlobalVersion         int64
	preparedStagingTables           map[interface{}]*StagingTable
	isTrieSameKeyCompatibility      bool // The `isTrieSameKeyCompatibility` is used to prevent conflict in continuous changes with same key/value.
	disableStrictGlobalVersionCheck bool // default `true`
}

// NewStagingTable return new instance of StagingTable.
func NewStagingTable(storage storage.Storage, tid interface{}, trieSameKeyCompatibility bool) *StagingTable {
	tbl := &StagingTable{
		storage:                         storage,
		globalVersion:                   0,
		parentStagingTable:              nil,
		versionizedValues:               make(stagingValuesMap),
		tid:                             tid,
		prepareingGlobalVersion:         0,
		preparedStagingTables:           make(map[interface{}]*StagingTable),
		isTrieSameKeyCompatibility:      trieSameKeyCompatibility,
		disableStrictGlobalVersionCheck: true,
	}
	return tbl
}

// SetStrictGlobalVersionCheck set global version check
func (tbl *StagingTable) SetStrictGlobalVersionCheck(flag bool) {
	tbl.disableStrictGlobalVersionCheck = !flag
}

// Prepare a independent staging table
func (tbl *StagingTable) Prepare(tid interface{}) (*StagingTable, error) {
	tbl.mutex.Lock()
	defer tbl.mutex.Unlock()

	if tbl.preparedStagingTables[tid] != nil {
		return nil, ErrTidIsExist
	}

	preparedTbl := &StagingTable{
		storage:                         tbl.storage,
		globalVersion:                   0,
		parentStagingTable:              tbl,
		prepareingGlobalVersion:         tbl.globalVersion,
		versionizedValues:               make(stagingValuesMap),
		tid:                             tid,
		preparedStagingTables:           make(map[interface{}]*StagingTable),
		isTrieSameKeyCompatibility:      tbl.isTrieSameKeyCompatibility,
		disableStrictGlobalVersionCheck: tbl.disableStrictGlobalVersionCheck,
	}

	tbl.preparedStagingTables[tid] = preparedTbl
	return preparedTbl, nil
}

// Get return value by key. If key does not exist, copy and incr version from `parentStagingTable` to record previous version.
func (tbl *StagingTable) Get(key []byte) (*VersionizedValueItem, error) {
	return tbl.GetByKey(key, true)
}

// GetByKey return value by key. If key does not exist, copy and incr version from `parentStagingTable` to record previous version.
func (tbl *StagingTable) GetByKey(key []byte, loadFromStorage bool) (*VersionizedValueItem, error) {
	// double check lock to prevent dead lock while call MergeToParent().
	tbl.mutex.Lock()
	keyStr := byteutils.Hex(key)
	value := tbl.versionizedValues[keyStr]
	tbl.mutex.Unlock()

	if value == nil {
		var err error
		if tbl.parentStagingTable != nil {
			value, err = tbl.parentStagingTable.GetByKey(key, loadFromStorage)
			if err != nil {
				return nil, err
			}

			// global version of keys are not the same, error.
			if !tbl.disableStrictGlobalVersionCheck && value.globalVersion > tbl.prepareingGlobalVersion {
				return nil, ErrStagingTableKeyConfliction
			}

			value = CloneVersionizedValueItem(tbl.tid, value)

		} else {
			if loadFromStorage {
				// load from storage.
				value, err = tbl.loadFromStorage(key)
				if err != nil && err != storage.ErrKeyNotFound {
					return nil, err
				}
			} else {
				value = NewDefaultVersionizedValueItem(key, nil, tbl.tid, 0)
				return value, nil
			}
		}

		// lock and check again.
		tbl.mutex.Lock()
		regetValue := tbl.versionizedValues[keyStr]
		if regetValue == nil {
			tbl.versionizedValues[keyStr] = value
		}
		tbl.mutex.Unlock()
	}

	return value, nil
}

// Put put the key/val pair. If key does not exist, copy and incr version from `parentStagingTable` to record previous version.
func (tbl *StagingTable) Put(key []byte, val []byte) (*VersionizedValueItem, error) {
	value, err := tbl.GetByKey(key, false)
	if err != nil {
		return nil, err
	}

	value.dirty = true
	value.val = val
	tbl.mutex.Lock()
	keyStr := byteutils.Hex(key)
	regetValue := tbl.versionizedValues[keyStr]
	if regetValue == nil {
		tbl.versionizedValues[keyStr] = value
	}
	tbl.mutex.Unlock()

	return value, nil
}

// Del del the tid/key pair. If tid+key does not exist, copy and incr version from `finalVersionizedValues` to record previous version.
func (tbl *StagingTable) Del(key []byte) (*VersionizedValueItem, error) {
	value, err := tbl.GetByKey(key, false)
	if err != nil {
		return nil, err
	}

	value.deleted = true
	value.dirty = true
	tbl.mutex.Lock()
	keyStr := byteutils.Hex(key)
	regetValue := tbl.versionizedValues[keyStr]
	if regetValue == nil {
		tbl.versionizedValues[keyStr] = value
	}
	tbl.mutex.Unlock()
	return value, nil
}

// Purge purge key/value pairs of tid.
func (tbl *StagingTable) Purge() {
	tbl.mutex.Lock()
	defer tbl.mutex.Unlock()

	for _, p := range tbl.preparedStagingTables {
		p.Purge()
	}
	tbl.preparedStagingTables = make(map[interface{}]*StagingTable)

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

		if targetValueItem == nil {
			continue
		}

		// 1. record conflict.
		if fromValueItem.isConflict(targetValueItem, tbl.isTrieSameKeyCompatibility) {
			conflictKeys[keyStr] = targetValueItem.tid
			continue
		}

		// 2. record dependentTids.

		// skip default value loaded from storage.
		if targetValueItem.isDefault() {
			continue
		}

		// ignore same parent tid for dependentTids.
		if targetValueItem.tid == tbl.parentStagingTable.tid {
			continue
		}

		// ignore version check when TrieSameKeyCompatibility is enabled.
		if tbl.isTrieSameKeyCompatibility {
			continue
		}

		dependentTids[targetValueItem.tid] = true
	}

	if len(conflictKeys) > 0 {
		logging.VLog().WithFields(logrus.Fields{
			"tid":          tbl.tid,
			"parentTid":    tbl.parentStagingTable.tid,
			"conflictKeys": conflictKeys,
		}).Debug("Failed to be merged into parent.")
		return nil, ErrStagingTableKeyConfliction
	}

	// 2. merge to final.

	// incr parentStagingTable.globalVersion.
	tbl.parentStagingTable.globalVersion++

	for keyStr, fromValueItem := range tbl.versionizedValues {
		// ignore default value item.
		if fromValueItem.isDefault() {
			continue
		}

		// ignore non-dirty.
		if !fromValueItem.dirty {
			continue
		}

		// merge.
		value := fromValueItem.CloneForMerge(tbl.parentStagingTable.globalVersion)
		targetValues[keyStr] = value
	}

	tids := make([]interface{}, 0, len(dependentTids))
	for key := range dependentTids {
		tids = append(tids, key)
	}

	return tids, nil
}

// Detach the staging table
func (tbl *StagingTable) Detach() error {
	tbl.mutex.Lock()
	if tbl.parentStagingTable == nil {
		tbl.mutex.Unlock()
		return ErrDisallowedCallingInNoPreparedDB
	}
	parentStagingTableMu := &tbl.parentStagingTable.mutex
	tbl.mutex.Unlock()

	parentStagingTableMu.Lock()
	defer parentStagingTableMu.Unlock()

	delete(tbl.parentStagingTable.preparedStagingTables, tbl.tid)
	tbl.parentStagingTable = nil
	return nil
}

func (tbl *StagingTable) getVersionizedValues() stagingValuesMap {
	return tbl.versionizedValues
}

// Lock staging table
func (tbl *StagingTable) Lock() {
	tbl.mutex.Lock()
}

// Unlock staging table
func (tbl *StagingTable) Unlock() {
	tbl.mutex.Unlock()
}

func (tbl *StagingTable) loadFromStorage(key []byte) (*VersionizedValueItem, error) {
	// get from storage.
	val, err := tbl.storage.Get(key)
	if err != nil && err != storage.ErrKeyNotFound {
		return nil, err
	}

	value := NewDefaultVersionizedValueItem(key, val, tbl.tid, 0)
	return value, nil
}

func (value *VersionizedValueItem) isDefault() bool {
	return value.version == 0 && value.dirty == false
}

func (value *VersionizedValueItem) isConflict(b *VersionizedValueItem, trieSameKeyCompatibility bool) bool {
	if b == nil {
		return true
	}

	// version same, no conflict.
	if value.version == b.version {
		return false
	}

	if trieSameKeyCompatibility == true {
		// ignore version check when TrieSameKeyCompatibility is enabled.

		if value.deleted != b.deleted {
			// deleted flag are not the same, conflict.
			return true
		}

		if !value.deleted && !bytes.Equal(value.val, b.val) {
			// both not delete, and val are not the same, conflict.
			return true
		}

		// otherwise, no conflict.
		return false
	}

	// otherwise, conflict.
	return true
}

// NewDefaultVersionizedValueItem return new instance of VersionizedValueItem, old/new version are 0, dirty is false.
func NewDefaultVersionizedValueItem(key []byte, val []byte, tid interface{}, globalVersion int64) *VersionizedValueItem {
	return &VersionizedValueItem{
		tid:           tid,
		key:           key,
		val:           val,
		version:       0,
		deleted:       false,
		dirty:         false,
		globalVersion: globalVersion,
	}
}

// CloneVersionizedValueItem copy and return the version increased VersionizedValueItem.
func CloneVersionizedValueItem(tid interface{}, oldValue *VersionizedValueItem) *VersionizedValueItem {
	return &VersionizedValueItem{
		tid:           tid,
		key:           oldValue.key,
		val:           oldValue.val,
		version:       oldValue.version,
		deleted:       oldValue.deleted,
		dirty:         false,
		globalVersion: oldValue.globalVersion,
	}
}

// CloneForMerge shadow copy of `VersionizedValueItem` with dirty is true.
func (value *VersionizedValueItem) CloneForMerge(globalVersion int64) *VersionizedValueItem {
	return &VersionizedValueItem{
		tid:           value.tid,
		key:           value.key,
		val:           value.val,
		version:       value.version + 1,
		deleted:       value.deleted,
		dirty:         true,
		globalVersion: globalVersion,
	}
}
