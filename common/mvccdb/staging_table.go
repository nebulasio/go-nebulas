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
	"errors"
	"sync"

	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"

	"github.com/nebulasio/go-nebulas/util/byteutils"
)

var (
	ErrStagingTableKeyConfliction = errors.New("staging table key confliction")
)

type stagingValuesMap map[string]*VersionizedValueItem
type stagingValuesMapMap map[interface{}]stagingValuesMap

type VersionizedValueItem struct {
	tid     interface{}
	key     []byte
	val     []byte
	old     int
	new     int
	dirty   bool
	deleted bool
}

type StagingTable struct {
	allVersionizedValues   map[interface{}]stagingValuesMap
	finalVersionValue      stagingValuesMap
	tidMutex               sync.Mutex
	finalVersionValueMutex sync.Mutex
}

func NewStagingTable() *StagingTable {
	return &StagingTable{
		allVersionizedValues: make(map[interface{}]stagingValuesMap),
		finalVersionValue:    make(stagingValuesMap),
	}
}

func (tbl *StagingTable) Get(tid interface{}, key []byte) *VersionizedValueItem {
	return tbl.getVersionizedValueForUpdate(tid, key, false)
}

func (tbl *StagingTable) Put(tid interface{}, key []byte, val []byte, dirty bool) *VersionizedValueItem {
	value := tbl.getVersionizedValueForUpdate(tid, key, true)
	value.val = val
	value.dirty = dirty
	return value
}

func (tbl *StagingTable) Del(tid interface{}, key []byte) *VersionizedValueItem {
	value := tbl.getVersionizedValueForUpdate(tid, key, true)
	value.deleted = true
	value.dirty = true
	return value
}

func (tbl *StagingTable) Purge(tid interface{}) {
	if tid == nil {
		return
	}

	tbl.tidMutex.Lock()
	defer tbl.tidMutex.Unlock()

	delete(tbl.allVersionizedValues, tid)
}

func (tbl *StagingTable) MergeToFinal(tid interface{}) ([]interface{}, error) {
	tbl.finalVersionValueMutex.Lock()
	defer tbl.finalVersionValueMutex.Unlock()

	dependentTids := make([]interface{}, 0)
	conflictKeys := make(map[string]interface{})

	// 1. check version.
	finalValues := tbl.finalVersionValue
	tidValues := tbl.getVersionizedValuesOfTid(tid)
	for keyStr, tidValueItem := range tidValues {
		finalValueItem := finalValues[keyStr]
		if finalValueItem == nil {
			logging.VLog().WithFields(logrus.Fields{
				"keyStr": keyStr,
				"tid":    tid,
			}).Warn("Key should be in finalVersionValue.")
			continue
		}

		// ignore dirty.
		if tidValueItem.dirty == false {
			continue
		}

		// record conflict.
		if tidValueItem.old != finalValueItem.new {
			conflictKeys[keyStr] = finalValueItem.tid
			continue
		}

		// record dependentTids.
		dependentTids = append(dependentTids, finalValueItem.tid)
	}

	if len(conflictKeys) > 0 {
		logging.VLog().WithFields(logrus.Fields{
			"tid":          tid,
			"conflictKeys": conflictKeys,
		}).Debug("Check failed.")
		return nil, ErrStagingTableKeyConfliction
	}

	// 2. merge to final.
	for keyStr, tidValueItem := range tidValues {
		// ignore dirty.
		if tidValueItem.dirty == false {
			continue
		}

		// merge.
		finalValues[keyStr] = tidValueItem.CloneForFinal()
	}

	return dependentTids, nil
}

func (tbl *StagingTable) LockFinalVersionValue() {
	tbl.finalVersionValueMutex.Lock()
}

func (tbl *StagingTable) UnlockFinalVersionValue() {
	tbl.finalVersionValueMutex.Unlock()
}

func (tbl *StagingTable) getVersionizedValueForUpdate(tid interface{}, key []byte, createIfNotExist bool) *VersionizedValueItem {
	keyStr := byteutils.Hex(key)
	tidValues := tbl.getVersionizedValuesOfTid(tid)

	value := tidValues[keyStr]
	if value == nil {
		value = tbl.getAndIncrValueFromFinalVersionValue(tid, keyStr, key, createIfNotExist)
		tidValues[keyStr] = value
	}

	return value
}

func (tbl *StagingTable) getVersionizedValuesOfTid(tid interface{}) stagingValuesMap {
	if tid == nil {
		return tbl.finalVersionValue
	}

	tbl.tidMutex.Lock()
	defer tbl.tidMutex.Unlock()

	tidValues := tbl.allVersionizedValues[tid]
	if tidValues == nil {
		tidValues = make(stagingValuesMap)
		tbl.allVersionizedValues[tid] = tidValues
	}
	return tidValues
}

func (tbl *StagingTable) getAndIncrValueFromFinalVersionValue(tid interface{}, keyStr string, key []byte, createIfNotExist bool) *VersionizedValueItem {
	tbl.finalVersionValueMutex.Lock()
	defer tbl.finalVersionValueMutex.Unlock()

	latestValue := tbl.finalVersionValue[keyStr]
	if latestValue == nil && createIfNotExist == true {
		latestValue = NewDefaultVersionizedValueItem(key)
		tbl.finalVersionValue[keyStr] = latestValue
	}

	if latestValue != nil {
		// incr version.
		value := IncrVersionizedValueItem(tid, latestValue)
		return value
	}

	return nil
}

func NewDefaultVersionizedValueItem(key []byte) *VersionizedValueItem {
	return &VersionizedValueItem{
		tid:     nil,
		key:     key,
		val:     nil,
		old:     0,
		new:     0,
		deleted: false,
		dirty:   false,
	}
}

func IncrVersionizedValueItem(tid interface{}, oldValue *VersionizedValueItem) *VersionizedValueItem {
	return &VersionizedValueItem{
		tid:     tid,
		key:     oldValue.key,
		val:     oldValue.val,
		old:     oldValue.new,
		new:     oldValue.new + 1,
		deleted: false,
		dirty:   false,
	}
}

func (value *VersionizedValueItem) CloneForFinal() *VersionizedValueItem {
	return &VersionizedValueItem{
		tid:     value.tid,
		key:     value.key,
		val:     value.val,
		old:     value.old,
		new:     value.new,
		deleted: value.deleted,
		dirty:   false,
	}
}
