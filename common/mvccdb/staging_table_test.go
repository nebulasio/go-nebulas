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
	value := NewDefaultVersionizedValueItem(key, val)

	assert.Equal(t, key, value.key)
	assert.Equal(t, val, value.val)
	assert.Nil(t, value.tid)
	assert.Equal(t, 0, value.old)
	assert.Equal(t, 0, value.new)
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
					old:     0,
					new:     0,
					deleted: false,
					dirty:   false,
				},
			},
			&VersionizedValueItem{
				tid:     tid,
				key:     key,
				val:     val,
				old:     0,
				new:     1,
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
					old:     0,
					new:     1,
					deleted: true,
					dirty:   false,
				},
			},
			&VersionizedValueItem{
				tid:     tid,
				key:     key,
				val:     val,
				old:     1,
				new:     2,
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
					old:     0,
					new:     1,
					deleted: false,
					dirty:   true,
				},
			},
			&VersionizedValueItem{
				tid:     tid,
				key:     key,
				val:     val,
				old:     1,
				new:     2,
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

func TestVersionizedValueItem_CloneForFinal(t *testing.T) {
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
				old:     0,
				new:     1,
				deleted: false,
				dirty:   false,
			},
			&VersionizedValueItem{
				tid:     tid,
				key:     key,
				val:     val,
				old:     0,
				new:     1,
				deleted: false,
				dirty:   true,
			},
		},
		{"2",
			&VersionizedValueItem{
				tid:     tid,
				key:     key,
				val:     val,
				old:     0,
				new:     1,
				deleted: false,
				dirty:   true,
			},
			&VersionizedValueItem{
				tid:     tid,
				key:     key,
				val:     val,
				old:     0,
				new:     1,
				deleted: false,
				dirty:   true,
			},
		},
		{"3",
			&VersionizedValueItem{
				tid:     tid,
				key:     key,
				val:     val,
				old:     0,
				new:     1,
				deleted: true,
				dirty:   true,
			},
			&VersionizedValueItem{
				tid:     tid,
				key:     key,
				val:     val,
				old:     0,
				new:     1,
				deleted: true,
				dirty:   true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := tt.value
			if got := value.CloneForFinal(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VersionizedValueItem.CloneForFinal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStagingTable_SingleTidAction(t *testing.T) {
	stor, _ := storage.NewMemoryStorage()
	tid := "tid"

	tbl := NewStagingTable(stor)

	// Get non-exist key.
	{
		key := []byte("key1")
		value, err := tbl.Get(tid, key)
		assert.Nil(t, err)
		assert.Nil(t, value.val)
	}

	// Put the key.
	{
		key := []byte("key1")
		val := []byte("val of key1")
		value, _ := tbl.Put(tid, key, val)
		assert.NotNil(t, value)
		assert.Equal(t, tid, value.tid)
		assert.Equal(t, key, value.key)
		assert.Equal(t, val, value.val)
		assert.Equal(t, 0, value.old)
		assert.Equal(t, 1, value.new)
		assert.False(t, value.deleted)
		assert.True(t, value.dirty)

		ret, _ := tbl.Get(tid, key)
		assert.Equal(t, value, ret)
	}

	// Set the key with dirty = false.
	{
		key := []byte("key1")
		val := []byte("val of key1")
		value, _ := tbl.Set(tid, key, val, false, false)
		assert.NotNil(t, value)
		assert.Equal(t, tid, value.tid)
		assert.Equal(t, key, value.key)
		assert.Equal(t, val, value.val)
		assert.Equal(t, 0, value.old)
		assert.Equal(t, 1, value.new)
		assert.False(t, value.deleted)
		assert.False(t, value.dirty)

		ret, _ := tbl.Get(tid, key)
		assert.Equal(t, value, ret)
	}

	// Put the key again.
	{
		// same value, expect dirty = false.
		key := []byte("key1")
		val := []byte("val of key1")
		value, _ := tbl.Put(tid, key, val)
		assert.NotNil(t, value)
		assert.Equal(t, tid, value.tid)
		assert.Equal(t, key, value.key)
		assert.Equal(t, val, value.val)
		assert.Equal(t, 0, value.old)
		assert.Equal(t, 1, value.new)
		assert.False(t, value.deleted)
		assert.False(t, value.dirty)

		ret, _ := tbl.Get(tid, key)
		assert.Equal(t, value, ret)

		// changed value, expect dirty = true.
		val = []byte("new val of key1")
		value, _ = tbl.Put(tid, key, val)
		assert.NotNil(t, value)
		assert.Equal(t, tid, value.tid)
		assert.Equal(t, key, value.key)
		assert.Equal(t, val, value.val)
		assert.Equal(t, 0, value.old)
		assert.Equal(t, 1, value.new)
		assert.False(t, value.deleted)
		assert.True(t, value.dirty)

		ret, _ = tbl.Get(tid, key)
		assert.Equal(t, value, ret)

		// same value, expected dirty = true.
		val = []byte("new val of key1")
		value, _ = tbl.Put(tid, key, val)
		assert.NotNil(t, value)
		assert.Equal(t, tid, value.tid)
		assert.Equal(t, key, value.key)
		assert.Equal(t, val, value.val)
		assert.Equal(t, 0, value.old)
		assert.Equal(t, 1, value.new)
		assert.False(t, value.deleted)
		assert.True(t, value.dirty)

		ret, _ = tbl.Get(tid, key)
		assert.Equal(t, value, ret)
	}

	// Del the key.
	{
		key := []byte("key1")
		value, _ := tbl.Del(tid, key)
		assert.NotNil(t, value)
		assert.Equal(t, tid, value.tid)
		assert.Equal(t, key, value.key)
		assert.NotNil(t, value.val)
		assert.Equal(t, 0, value.old)
		assert.Equal(t, 1, value.new)
		assert.True(t, value.deleted)
		assert.True(t, value.dirty)

		ret, _ := tbl.Get(tid, key)
		assert.Equal(t, value, ret)

		// restore dirty flag.
		value, _ = tbl.Set(tid, key, value.val, false, false)
		assert.False(t, value.deleted)
		assert.False(t, value.dirty)

		// del again.
		value, _ = tbl.Del(tid, key)
		assert.NotNil(t, value)
		assert.Equal(t, tid, value.tid)
		assert.Equal(t, key, value.key)
		assert.NotNil(t, value.val)
		assert.Equal(t, 0, value.old)
		assert.Equal(t, 1, value.new)
		assert.True(t, value.deleted)
		assert.True(t, value.dirty)

		ret, _ = tbl.Get(tid, key)
		assert.Equal(t, value, ret)
	}
}

func TestStagingTable_MultiTidAction(t *testing.T) {
	stor, _ := storage.NewMemoryStorage()

	tid1 := "tid1"
	tid2 := "tid2"
	key := []byte("key1")
	val := []byte("val of key1")

	tbl := NewStagingTable(stor)

	// tid1 put the key.
	value, _ := tbl.Put(tid1, key, val)
	ret, _ := tbl.Get(tid1, key)
	assert.Equal(t, value, ret)

	// read from tid2.
	value2, _ := tbl.Get(tid2, key)
	assert.Nil(t, value2.val)

	// delete.
	tbl.Del(tid2, key)

	ret, _ = tbl.Get(tid2, key)
	assert.True(t, ret.deleted)
	ret, _ = tbl.Get(tid1, key)
	assert.False(t, ret.deleted)
}

func TestStagingTable_MergeToFinal(t *testing.T) {
	tid1 := "tid1"
	tid2 := "tid2"

	stor, _ := storage.NewMemoryStorage()
	tbl := NewStagingTable(stor)

	// Init.
	key1 := []byte("key1")
	val1_1 := []byte("v1_1")
	val1_2 := []byte("v1_2")

	key2 := []byte("key2")
	key3 := []byte("key3")

	tbl.Get(tid1, key1)
	tbl.Get(tid2, key1)

	// Put keys.
	tbl.Put(tid1, key1, val1_1)
	tbl.Del(tid1, key2)

	tbl.Put(tid2, key1, val1_2)
	tbl.Del(tid2, key3)

	// Merge tid2, success.
	dependencies, err := tbl.MergeToFinal(tid2)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(dependencies))
	ret, _ := tbl.Get(tid1, key1)
	assert.Equal(t, val1_1, ret.val)
	ret, _ = tbl.Get(tid2, key1)
	assert.Equal(t, val1_2, ret.val)

	// Merge tid1, fail.
	dependencies, err = tbl.MergeToFinal(tid1)
	assert.Equal(t, ErrStagingTableKeyConfliction, err)
	assert.Nil(t, dependencies)

	// tid3 read.
	tid3 := "tid3"
	value, _ := tbl.Get(tid3, key1)
	assert.NotNil(t, value)
	assert.Equal(t, tid3, value.tid)
	assert.Equal(t, key1, value.key)
	assert.Equal(t, val1_2, value.val)
	assert.Equal(t, 1, value.old)
	assert.Equal(t, 2, value.new)
	assert.False(t, value.deleted)
	assert.False(t, value.dirty)

	// tid3 put.
	tbl.Put(tid3, key2, []byte("val of deled key2"))

	// Merge tid3.
	dependencies, err = tbl.MergeToFinal(tid3)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(dependencies))
	assert.Equal(t, tid2, dependencies[0])

	// tid4 read/put.
	tid4 := "tid4"
	value, _ = tbl.Get(tid4, key1)
	assert.Equal(t, val1_2, value.val)
	assert.Equal(t, 1, value.old)
	assert.Equal(t, 2, value.new)

	value, _ = tbl.Del(tid4, key2)
	assert.Equal(t, 1, value.old)
	assert.Equal(t, 2, value.new)

	value, _ = tbl.Put(tid4, key3, []byte("val of deleted key3"))
	assert.Equal(t, 1, value.old)
	assert.Equal(t, 2, value.new)

	// Merge tid4.
	dependencies, err = tbl.MergeToFinal(tid4)
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

	tid := "tid1"

	stor, _ := storage.NewMemoryStorage()
	tbl := NewStagingTable(stor)

	// tid put.
	tbl.Put(tid, key1, []byte("value of key1"))
	tbl.Put(tid, key2, []byte("value of key2"))

	ret, _ := tbl.Get(tid, key1)
	assert.Equal(t, []byte("value of key1"), ret.val)
	ret, _ = tbl.Get(tid, key2)
	assert.Equal(t, []byte("value of key2"), ret.val)

	// purge.
	tbl.Purge(tid)

	// verify.
	ret, _ = tbl.Get(tid, key1)
	assert.Nil(t, ret.val)
	ret, _ = tbl.Get(tid, key2)
	assert.Nil(t, ret.val)

	// merge.
	tbl.Put(tid, key1, []byte("value of key1"))
	_, err := tbl.MergeToFinal(tid)
	assert.Nil(t, err)

	// tid put.
	tbl.Put(tid, key1, []byte("value of key1: 2"))
	tbl.Put(tid, key2, []byte("value of key2: 2"))

	// purge again.
	tbl.Purge(tid)

	// verify.
	ret, _ = tbl.Get(tid, key1)
	assert.Equal(t, []byte("value of key1"), ret.val)
	ret, _ = tbl.Get(tid, key2)
	assert.Nil(t, ret.val)
}

func TestStagingTable_NilTidAction(t *testing.T) {
	key1 := []byte("key1")
	key2 := []byte("key2")
	tid := "tid1"

	stor, _ := storage.NewMemoryStorage()
	tbl := NewStagingTable(stor)

	// Get and Put.
	value, _ := tbl.Get(nil, key1)
	assert.Nil(t, value.val)

	value, _ = tbl.Put(nil, key1, []byte("value of key1"))
	assert.Equal(t, []byte("value of key1"), value.val)
	assert.Equal(t, 0, value.old)
	assert.Equal(t, 1, value.new)
	assert.False(t, value.deleted)
	assert.True(t, value.dirty)

	value, _ = tbl.Get(nil, key1)
	assert.Equal(t, []byte("value of key1"), value.val)
	assert.Equal(t, 0, value.old)
	assert.Equal(t, 1, value.new)
	assert.False(t, value.deleted)
	assert.True(t, value.dirty)

	// Put.
	value, _ = tbl.Put(nil, key2, []byte("value of key2"))
	assert.Equal(t, []byte("value of key2"), value.val)
	assert.Equal(t, 0, value.old)
	assert.Equal(t, 1, value.new)
	assert.False(t, value.deleted)
	assert.True(t, value.dirty)

	value, _ = tbl.Get(nil, key2)
	assert.Equal(t, []byte("value of key2"), value.val)
	assert.Equal(t, 0, value.old)
	assert.Equal(t, 1, value.new)
	assert.False(t, value.deleted)
	assert.True(t, value.dirty)

	// tid get.
	value, _ = tbl.Get(tid, key1)
	assert.Equal(t, []byte("value of key1"), value.val)
	assert.Equal(t, 1, value.old)
	assert.Equal(t, 2, value.new)
	assert.False(t, value.deleted)
	assert.False(t, value.dirty)

	tbl.Del(tid, key1)

	// tid merge.
	_, err := tbl.MergeToFinal(tid)
	assert.Nil(t, err)

	// Get.
	value, _ = tbl.Get(nil, key1)
	assert.True(t, value.deleted)
}
