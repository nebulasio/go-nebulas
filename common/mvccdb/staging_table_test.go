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
	"reflect"
	"testing"

	"github.com/nebulasio/go-nebulas/storage"
	"github.com/stretchr/testify/assert"
)

func TestNewDefaultVersionizedValueItem(t *testing.T) {
	key := make([]byte, 0)
	val := []byte("value")
	value := NewDefaultVersionizedValueItem(key, val, "tid")

	assert.Equal(t, key, value.key)
	assert.Equal(t, val, value.val)
	assert.Equal(t, "tid", value.tid)
	assert.Equal(t, 0, value.version)
	assert.False(t, value.deleted)
	assert.False(t, value.dirty)
}

func TestIncrVersionizedValueItem(t *testing.T) {
	tid := "tid"
	key := make([]byte, 0)
	val := make([]byte, 0)

	type args struct {
		tid      interface{}
		oldValue *VersionizedValueItem
	}
	tests := []struct {
		name string
		args args
		want *VersionizedValueItem
	}{
		{"1",
			args{
				tid: tid,
				oldValue: &VersionizedValueItem{
					tid:     tid,
					key:     key,
					val:     val,
					version: 0,
					deleted: false,
					dirty:   false,
				},
			},
			&VersionizedValueItem{
				tid:     tid,
				key:     key,
				val:     val,
				version: 0,
				deleted: false,
				dirty:   false,
			},
		},
		{"2",
			args{
				tid: tid,
				oldValue: &VersionizedValueItem{
					tid:     tid,
					key:     key,
					val:     val,
					version: 1,
					deleted: true,
					dirty:   false,
				},
			},
			&VersionizedValueItem{
				tid:     tid,
				key:     key,
				val:     val,
				version: 1,
				deleted: true,
				dirty:   false,
			},
		},
		{"3",
			args{
				tid: tid,
				oldValue: &VersionizedValueItem{
					tid:     tid,
					key:     key,
					val:     val,
					version: 1,
					deleted: false,
					dirty:   true,
				},
			},
			&VersionizedValueItem{
				tid:     tid,
				key:     key,
				val:     val,
				version: 1,
				deleted: false,
				dirty:   false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IncrVersionizedValueItem(tt.args.tid, tt.args.oldValue); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IncrVersionizedValueItem() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionizedValueItem_CloneForMerge(t *testing.T) {
	tid := "tid"
	key := make([]byte, 0)
	val := make([]byte, 0)

	tests := []struct {
		name  string
		value *VersionizedValueItem
		want  *VersionizedValueItem
	}{
		{"1",
			&VersionizedValueItem{
				tid:     tid,
				key:     key,
				val:     val,
				version: 0,
				deleted: false,
				dirty:   false,
			},
			&VersionizedValueItem{
				tid:     tid,
				key:     key,
				val:     val,
				version: 1,
				deleted: false,
				dirty:   true,
			},
		},
		{"2",
			&VersionizedValueItem{
				tid:     tid,
				key:     key,
				val:     val,
				version: 1,
				deleted: false,
				dirty:   true,
			},
			&VersionizedValueItem{
				tid:     tid,
				key:     key,
				val:     val,
				version: 2,
				deleted: false,
				dirty:   true,
			},
		},
		{"3",
			&VersionizedValueItem{
				tid:     tid,
				key:     key,
				val:     val,
				version: 0,
				deleted: true,
				dirty:   true,
			},
			&VersionizedValueItem{
				tid:     tid,
				key:     key,
				val:     val,
				version: 1,
				deleted: true,
				dirty:   true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := tt.value
			if got := value.CloneForMerge(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VersionizedValueItem.CloneForFinal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStagingTable_SingleTidAction(t *testing.T) {
	stor, _ := storage.NewMemoryStorage()
	tid := "tid"

	tbl := NewStagingTable(stor, tid, false)

	// Get non-exist key.
	{
		key := []byte("key1")
		value, err := tbl.Get(key)
		assert.Nil(t, err)
		assert.Nil(t, value.val)
	}

	// Put the key.
	{
		key := []byte("key1")
		val := []byte("val of key1")
		value, _ := tbl.Put(key, val)
		assert.NotNil(t, value)
		assert.Equal(t, tid, value.tid)
		assert.Equal(t, key, value.key)
		assert.Equal(t, val, value.val)
		assert.Equal(t, 0, value.version)
		assert.False(t, value.deleted)
		assert.True(t, value.dirty)

		ret, _ := tbl.Get(key)
		assert.Equal(t, value, ret)
	}

	// Set the key with dirty = false.
	{
		key := []byte("key1")
		val := []byte("val of key1")
		value, _ := tbl.Set(key, val, false, false)
		assert.NotNil(t, value)
		assert.Equal(t, tid, value.tid)
		assert.Equal(t, key, value.key)
		assert.Equal(t, val, value.val)
		assert.Equal(t, 0, value.version)
		assert.False(t, value.deleted)
		assert.False(t, value.dirty)

		ret, _ := tbl.Get(key)
		assert.Equal(t, value, ret)
	}

	// Put the key again.
	{
		// same value, expect dirty = false.
		key := []byte("key1")
		val := []byte("val of key1")
		value, _ := tbl.Put(key, val)
		assert.NotNil(t, value)
		assert.Equal(t, tid, value.tid)
		assert.Equal(t, key, value.key)
		assert.Equal(t, val, value.val)
		assert.Equal(t, 0, value.version)
		assert.False(t, value.deleted)
		assert.False(t, value.dirty)

		ret, _ := tbl.Get(key)
		assert.Equal(t, value, ret)

		// changed value, expect dirty = true.
		val = []byte("new val of key1")
		value, _ = tbl.Put(key, val)
		assert.NotNil(t, value)
		assert.Equal(t, tid, value.tid)
		assert.Equal(t, key, value.key)
		assert.Equal(t, val, value.val)
		assert.Equal(t, 0, value.version)
		assert.False(t, value.deleted)
		assert.True(t, value.dirty)

		ret, _ = tbl.Get(key)
		assert.Equal(t, value, ret)

		// same value, expected dirty = true.
		val = []byte("new val of key1")
		value, _ = tbl.Put(key, val)
		assert.NotNil(t, value)
		assert.Equal(t, tid, value.tid)
		assert.Equal(t, key, value.key)
		assert.Equal(t, val, value.val)
		assert.Equal(t, 0, value.version)
		assert.False(t, value.deleted)
		assert.True(t, value.dirty)

		ret, _ = tbl.Get(key)
		assert.Equal(t, value, ret)
	}

	// Del the key.
	{
		key := []byte("key1")
		value, _ := tbl.Del(key)
		assert.NotNil(t, value)
		assert.Equal(t, tid, value.tid)
		assert.Equal(t, key, value.key)
		assert.NotNil(t, value.val)
		assert.Equal(t, 0, value.version)
		assert.True(t, value.deleted)
		assert.True(t, value.dirty)

		ret, _ := tbl.Get(key)
		assert.Equal(t, value, ret)

		// restore dirty flag.
		value, _ = tbl.Set(key, value.val, false, false)
		assert.False(t, value.deleted)
		assert.False(t, value.dirty)

		// del again.
		value, _ = tbl.Del(key)
		assert.NotNil(t, value)
		assert.Equal(t, tid, value.tid)
		assert.Equal(t, key, value.key)
		assert.NotNil(t, value.val)
		assert.Equal(t, 0, value.version)
		assert.True(t, value.deleted)
		assert.True(t, value.dirty)

		ret, _ = tbl.Get(key)
		assert.Equal(t, value, ret)
	}
}

func TestStagingTable_MultiTidAction(t *testing.T) {
	stor, _ := storage.NewMemoryStorage()

	tid1 := "tid1"
	tid2 := "tid2"
	key := []byte("key1")
	val := []byte("val of key1")

	tbl1 := NewStagingTable(stor, tid1, false)
	tbl2 := NewStagingTable(stor, tid2, false)

	// tid1 put the key.
	value, _ := tbl1.Put(key, val)
	ret, _ := tbl1.Get(key)
	assert.Equal(t, value, ret)

	// read from tid2.
	value2, _ := tbl2.Get(key)
	assert.Nil(t, value2.val)

	// delete.
	tbl2.Del(key)

	ret, _ = tbl2.Get(key)
	assert.True(t, ret.deleted)
	ret, _ = tbl1.Get(key)
	assert.False(t, ret.deleted)
}

func TestStagingTable_MergeToParent(t *testing.T) {
	tid1 := "tid1"
	tid2 := "tid2"

	stor, _ := storage.NewMemoryStorage()

	tbl1 := NewStagingTable(stor, tid1, false)
	tbl2, err := tbl1.Prepare(tid2)
	assert.Nil(t, err)

	// Init.
	key1 := []byte("key1")
	val1_1 := []byte("v1_1")
	val1_2 := []byte("v1_2")

	key2 := []byte("key2")
	key3 := []byte("key3")

	tbl1.Get(key1)
	tbl2.Get(key1)

	// Put keys.
	tbl1.Put(key1, val1_1)
	tbl1.Del(key2)

	tbl2.Put(key1, val1_2)
	tbl2.Del(key3)

	// Merge tid2, success.
	dependencies, err := tbl2.MergeToParent()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(dependencies))
	ret, _ := tbl1.Get(key1)
	assert.Equal(t, val1_2, ret.val)
	ret, _ = tbl2.Get(key1)
	assert.Equal(t, val1_2, ret.val)

	// Merge tid2, fail.
	dependencies, err = tbl2.MergeToParent()
	assert.Equal(t, ErrStagingTableKeyConfliction, err)
	assert.Equal(t, 0, len(dependencies))

	// tid3 read.
	tid3 := "tid3"
	tbl3, err := tbl1.Prepare(tid3)
	assert.Nil(t, err)

	value, _ := tbl3.Get(key1)
	assert.NotNil(t, value)
	assert.Equal(t, tid3, value.tid)
	assert.Equal(t, key1, value.key)
	assert.Equal(t, val1_2, value.val)
	assert.Equal(t, 1, value.version)
	assert.False(t, value.deleted)
	assert.False(t, value.dirty)

	// tid3 put.
	tbl3.Put(key2, []byte("val of deled key2"))

	// Merge tid3.
	dependencies, err = tbl3.MergeToParent()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(dependencies))
	assert.Equal(t, tid2, dependencies[0])

	// tid4 read/put.
	tid4 := "tid4"
	tbl4, err := tbl1.Prepare(tid4)
	assert.Nil(t, err)

	value, _ = tbl4.Get(key1)
	assert.Equal(t, val1_2, value.val)
	assert.Equal(t, 1, value.version)

	value, _ = tbl4.Del(key2)
	assert.Equal(t, 1, value.version)

	value, _ = tbl4.Put(key3, []byte("val of deleted key3"))
	assert.Equal(t, 1, value.version)

	// Merge tid4.
	dependencies, err = tbl4.MergeToParent()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(dependencies))

	expectedDepends := make(map[interface{}]bool)
	expectedDepends[tid2] = true
	expectedDepends[tid3] = true

	for _, v := range dependencies {
		assert.True(t, expectedDepends[v])
	}
}

func TestStagingTable_Purge(t *testing.T) {
	key1 := []byte("key1")
	key2 := []byte("key2")

	parentTid := "tid0"
	tid := "tid1"

	stor, _ := storage.NewMemoryStorage()
	parentTbl := NewStagingTable(stor, parentTid, false)
	tbl, err := parentTbl.Prepare(tid)
	assert.Nil(t, err)

	// tid put.
	tbl.Put(key1, []byte("value of key1"))
	tbl.Put(key2, []byte("value of key2"))

	ret, _ := tbl.Get(key1)
	assert.Equal(t, []byte("value of key1"), ret.val)
	ret, _ = tbl.Get(key2)
	assert.Equal(t, []byte("value of key2"), ret.val)

	// purge.
	tbl.Purge()

	// verify.
	ret, _ = tbl.Get(key1)
	assert.Nil(t, ret.val)
	ret, _ = tbl.Get(key2)
	assert.Nil(t, ret.val)

	// merge.
	tbl.Put(key1, []byte("value of key1"))
	_, err = tbl.MergeToParent()
	assert.Nil(t, err)

	// tid put.
	tbl.Put(key1, []byte("value of key1: 2"))
	tbl.Put(key2, []byte("value of key2: 2"))

	// purge again.
	tbl.Purge()

	// verify.
	ret, _ = tbl.Get(key1)
	assert.Equal(t, []byte("value of key1"), ret.val)
	ret, _ = tbl.Get(key2)
	assert.Nil(t, ret.val)
}

func TestStagingTable_PrepareAndClose(t *testing.T) {
	stor, _ := storage.NewMemoryStorage()

	{
		rootTbl := NewStagingTable(stor, "tid0", false)
		assert.Equal(t, ErrDisallowedCallingInNoPreparedDB, rootTbl.Close())
	}

	{
		tbl := NewStagingTable(stor, "tid0", false)

		tbl1, err := tbl.Prepare("tid1")
		assert.Nil(t, err)
		assert.Nil(t, tbl1.Close())

		tbl1, err = tbl.Prepare("tid1")
		assert.Nil(t, err)
		assert.Nil(t, tbl1.Close())
	}

	{
		tbl := NewStagingTable(stor, "tid0", false)

		tbl1, err := tbl.Prepare("tid1")
		assert.Nil(t, err)
		assert.NotNil(t, tbl1)

		tbl2, err := tbl.Prepare("tid1")
		assert.Equal(t, ErrTidIsExist, err)
		assert.Nil(t, tbl2)
	}
}

func TestStagingTable_TrieSameKeyCompatibility(t *testing.T) {
	stor, _ := storage.NewMemoryStorage()
	tbl := NewStagingTable(stor, "tid0", true)

	tbl1, _ := tbl.Prepare("tid1")
	tbl2, _ := tbl.Prepare("tid2")

	// same key/val.
	key := []byte("key")
	val := []byte("value")
	key1 := []byte("key1")
	val1 := []byte("val1")
	key2 := []byte("key2")
	val2 := []byte("val2")
	keyDel := []byte("key_del")

	// prepare key_del.
	tbl.Put(keyDel, []byte("should deleted."))

	// update prepared staging table.
	tbl1.Put(key, val)
	tbl1.Put(key1, val1)
	tbl1.Del(keyDel)

	tbl2.Put(key, val)
	tbl2.Put(key2, val2)
	tbl2.Del(keyDel)

	// merge tbl1, succeed.
	depends, err := tbl1.MergeToParent()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(depends))

	// merge tbl2, succeed.
	depends, err = tbl2.MergeToParent()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(depends))

	// verify.
	value, err := tbl.Get(key)
	assert.Nil(t, err)
	assert.Equal(t, val, value.val)
	assert.Equal(t, 1, value.version)

	value, err = tbl.Get(key1)
	assert.Nil(t, err)
	assert.Equal(t, val1, value.val)
	assert.Equal(t, 1, value.version)

	value, err = tbl.Get(key2)
	assert.Nil(t, err)
	assert.Equal(t, val2, value.val)
	assert.Equal(t, 1, value.version)
}

func TestStagingTable_TrieSameKeyCompatibility1(t *testing.T) {
	stor, _ := storage.NewMemoryStorage()

	// same key/val.
	key1 := []byte("key1")
	val1 := []byte("val1")
	val2 := []byte("val2")

	{
		tbl := NewStagingTable(stor, "tid0", true)
		tbl1, _ := tbl.Prepare("tid1")
		tbl2, _ := tbl.Prepare("tid2")

		// different value.
		tbl1.Put(key1, val1)
		tbl2.Put(key1, val2)

		// merge tbl1, succeed.
		depends, err := tbl1.MergeToParent()
		assert.Nil(t, err)
		assert.Equal(t, 0, len(depends))

		// merge tbl2, failed.
		depends, err = tbl2.MergeToParent()
		assert.Equal(t, ErrStagingTableKeyConfliction, err)
	}

	{
		tbl := NewStagingTable(stor, "tid0", true)
		tbl1, _ := tbl.Prepare("tid1")
		tbl2, _ := tbl.Prepare("tid2")

		// one delete.
		tbl1.Put(key1, val1)
		tbl2.Del(key1)

		// merge tbl1, succeed.
		depends, err := tbl1.MergeToParent()
		assert.Nil(t, err)
		assert.Equal(t, 0, len(depends))

		// merge tbl2, failed.
		depends, err = tbl2.MergeToParent()
		assert.Equal(t, ErrStagingTableKeyConfliction, err)
	}

}
