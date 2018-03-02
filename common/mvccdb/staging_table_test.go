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

	"github.com/stretchr/testify/assert"
)

func TestNewDefaultVersionizedValueItem(t *testing.T) {
	key := make([]byte, 0)
	value := NewDefaultVersionizedValueItem(key)

	assert.Equal(t, key, value.key)
	assert.Nil(t, value.val)
	assert.Nil(t, value.tid)
	assert.Equal(t, 0, value.old)
	assert.Equal(t, 0, value.new)
	assert.False(t, value.deleted)
	assert.False(t, value.dirty)
	assert.False(t, value.initialized)
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
					tid:         tid,
					key:         key,
					val:         val,
					old:         0,
					new:         0,
					deleted:     false,
					dirty:       false,
					initialized: true,
				},
			},
			&VersionizedValueItem{
				tid:         tid,
				key:         key,
				val:         val,
				old:         0,
				new:         1,
				deleted:     false,
				dirty:       false,
				initialized: true,
			},
		},
		{"2",
			args{
				tid: tid,
				oldValue: &VersionizedValueItem{
					tid:         tid,
					key:         key,
					val:         val,
					old:         0,
					new:         1,
					deleted:     true,
					dirty:       false,
					initialized: true,
				},
			},
			&VersionizedValueItem{
				tid:         tid,
				key:         key,
				val:         val,
				old:         1,
				new:         2,
				deleted:     true,
				dirty:       false,
				initialized: true,
			},
		},
		{"3",
			args{
				tid: tid,
				oldValue: &VersionizedValueItem{
					tid:         tid,
					key:         key,
					val:         val,
					old:         0,
					new:         1,
					deleted:     false,
					dirty:       true,
					initialized: true,
				},
			},
			&VersionizedValueItem{
				tid:         tid,
				key:         key,
				val:         val,
				old:         1,
				new:         2,
				deleted:     false,
				dirty:       false,
				initialized: true,
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
				tid:         tid,
				key:         key,
				val:         val,
				old:         0,
				new:         1,
				deleted:     false,
				dirty:       false,
				initialized: true,
			},
			&VersionizedValueItem{
				tid:         tid,
				key:         key,
				val:         val,
				old:         0,
				new:         1,
				deleted:     false,
				dirty:       false,
				initialized: true,
			},
		},
		{"2",
			&VersionizedValueItem{
				tid:         tid,
				key:         key,
				val:         val,
				old:         0,
				new:         1,
				deleted:     false,
				dirty:       true,
				initialized: true,
			},
			&VersionizedValueItem{
				tid:         tid,
				key:         key,
				val:         val,
				old:         0,
				new:         1,
				deleted:     false,
				dirty:       false,
				initialized: true,
			},
		},
		{"3",
			&VersionizedValueItem{
				tid:         tid,
				key:         key,
				val:         val,
				old:         0,
				new:         1,
				deleted:     true,
				dirty:       true,
				initialized: true,
			},
			&VersionizedValueItem{
				tid:         tid,
				key:         key,
				val:         val,
				old:         0,
				new:         1,
				deleted:     true,
				dirty:       false,
				initialized: true,
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
	tid := "tid"

	tbl := NewStagingTable()

	// Get non-exist key.
	{
		key := []byte("key1")
		value := tbl.Get(tid, key)
		assert.Nil(t, value)
	}

	// Put the key.
	{
		key := []byte("key1")
		val := []byte("val of key1")
		value := tbl.Put(tid, key, val)
		assert.NotNil(t, value)
		assert.Equal(t, tid, value.tid)
		assert.Equal(t, key, value.key)
		assert.Equal(t, val, value.val)
		assert.Equal(t, 0, value.old)
		assert.Equal(t, 1, value.new)
		assert.False(t, value.deleted)
		assert.True(t, value.dirty)
		assert.True(t, value.initialized)
		assert.Equal(t, value, tbl.Get(tid, key))
	}

	// Set the key with dirty = false.
	{
		key := []byte("key1")
		val := []byte("val of key1")
		value := tbl.Set(tid, key, val, false, false)
		assert.NotNil(t, value)
		assert.Equal(t, tid, value.tid)
		assert.Equal(t, key, value.key)
		assert.Equal(t, val, value.val)
		assert.Equal(t, 0, value.old)
		assert.Equal(t, 1, value.new)
		assert.False(t, value.deleted)
		assert.False(t, value.dirty)
		assert.True(t, value.initialized)
		assert.Equal(t, value, tbl.Get(tid, key))
	}

	// Put the key again.
	{
		// same value, expect dirty = false.
		key := []byte("key1")
		val := []byte("val of key1")
		value := tbl.Put(tid, key, val)
		assert.NotNil(t, value)
		assert.Equal(t, tid, value.tid)
		assert.Equal(t, key, value.key)
		assert.Equal(t, val, value.val)
		assert.Equal(t, 0, value.old)
		assert.Equal(t, 1, value.new)
		assert.False(t, value.deleted)
		assert.False(t, value.dirty)
		assert.True(t, value.initialized)
		assert.Equal(t, value, tbl.Get(tid, key))

		// changed value, expect dirty = true.
		val = []byte("new val of key1")
		value = tbl.Put(tid, key, val)
		assert.NotNil(t, value)
		assert.Equal(t, tid, value.tid)
		assert.Equal(t, key, value.key)
		assert.Equal(t, val, value.val)
		assert.Equal(t, 0, value.old)
		assert.Equal(t, 1, value.new)
		assert.False(t, value.deleted)
		assert.True(t, value.dirty)
		assert.True(t, value.initialized)
		assert.Equal(t, value, tbl.Get(tid, key))

		// same value, expected dirty = true.
		val = []byte("new val of key1")
		value = tbl.Put(tid, key, val)
		assert.NotNil(t, value)
		assert.Equal(t, tid, value.tid)
		assert.Equal(t, key, value.key)
		assert.Equal(t, val, value.val)
		assert.Equal(t, 0, value.old)
		assert.Equal(t, 1, value.new)
		assert.False(t, value.deleted)
		assert.True(t, value.dirty)
		assert.True(t, value.initialized)
		assert.Equal(t, value, tbl.Get(tid, key))
	}

	// Del the key.
	{
		key := []byte("key1")
		value := tbl.Del(tid, key)
		assert.NotNil(t, value)
		assert.Equal(t, tid, value.tid)
		assert.Equal(t, key, value.key)
		assert.NotNil(t, value.val)
		assert.Equal(t, 0, value.old)
		assert.Equal(t, 1, value.new)
		assert.True(t, value.deleted)
		assert.True(t, value.dirty)
		assert.True(t, value.initialized)
		assert.Equal(t, value, tbl.Get(tid, key))

		// restore dirty flag.
		value = tbl.Set(tid, key, value.val, false, false)
		assert.False(t, value.deleted)
		assert.False(t, value.dirty)
		assert.True(t, value.initialized)

		// del again.
		value = tbl.Del(tid, key)
		assert.NotNil(t, value)
		assert.Equal(t, tid, value.tid)
		assert.Equal(t, key, value.key)
		assert.NotNil(t, value.val)
		assert.Equal(t, 0, value.old)
		assert.Equal(t, 1, value.new)
		assert.True(t, value.deleted)
		assert.True(t, value.dirty)
		assert.True(t, value.initialized)
		assert.Equal(t, value, tbl.Get(tid, key))
	}
}

func TestStagingTable_MultiTidAction(t *testing.T) {
	tid1 := "tid1"
	tid2 := "tid2"
	key := []byte("key1")
	val := []byte("val of key1")

	tbl := NewStagingTable()

	// tid1 put the key.
	value := tbl.Put(tid1, key, val)
	assert.Equal(t, value, tbl.Get(tid1, key))

	// read from tid2.
	value2 := tbl.Get(tid2, key)
	assert.Nil(t, value2)

	// delete.
	tbl.Del(tid2, key)
	assert.True(t, tbl.Get(tid2, key).deleted)
	assert.False(t, tbl.Get(tid1, key).deleted)
}

func TestStagingTable_MergeToFinal(t *testing.T) {
	tid1 := "tid1"
	tid2 := "tid2"

	tbl := NewStagingTable()

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
	assert.Equal(t, val1_1, tbl.Get(tid1, key1).val)
	assert.Equal(t, val1_2, tbl.Get(tid2, key1).val)

	// Merge tid1, fail.
	dependencies, err = tbl.MergeToFinal(tid1)
	assert.Equal(t, ErrStagingTableKeyConfliction, err)
	assert.Nil(t, dependencies)

	// tid3 read.
	tid3 := "tid3"
	value := tbl.Get(tid3, key1)
	assert.NotNil(t, value)
	assert.Equal(t, tid3, value.tid)
	assert.Equal(t, key1, value.key)
	assert.Equal(t, val1_2, value.val)
	assert.Equal(t, 1, value.old)
	assert.Equal(t, 2, value.new)
	assert.False(t, value.deleted)
	assert.False(t, value.dirty)
	assert.True(t, value.initialized)

	// tid3 put.
	tbl.Put(tid3, key2, []byte("val of deled key2"))

	// Merge tid3.
	dependencies, err = tbl.MergeToFinal(tid3)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(dependencies))
	assert.Equal(t, tid2, dependencies[0])

	// tid4 read/put.
	tid4 := "tid4"
	value = tbl.Get(tid4, key1)
	assert.Equal(t, val1_2, value.val)
	assert.Equal(t, 1, value.old)
	assert.Equal(t, 2, value.new)

	value = tbl.Del(tid4, key2)
	assert.Equal(t, 1, value.old)
	assert.Equal(t, 2, value.new)

	value = tbl.Put(tid4, key3, []byte("val of deleted key3"))
	assert.Equal(t, 1, value.old)
	assert.Equal(t, 2, value.new)

	// Merge tid4.
	dependencies, err = tbl.MergeToFinal(tid4)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(dependencies))
	assert.Equal(t, tid2, dependencies[0])
	assert.Equal(t, tid3, dependencies[1])
}

func TestStagingTable_Purge(t *testing.T) {
	key1 := []byte("key1")
	key2 := []byte("key2")

	tid := "tid1"

	tbl := NewStagingTable()

	// tid put.
	tbl.Put(tid, key1, []byte("value of key1"))
	tbl.Put(tid, key2, []byte("value of key2"))

	assert.Equal(t, []byte("value of key1"), tbl.Get(tid, key1).val)
	assert.Equal(t, []byte("value of key2"), tbl.Get(tid, key2).val)

	// purge.
	tbl.Purge(tid)

	// verify.
	assert.Nil(t, tbl.Get(tid, key1))
	assert.Nil(t, tbl.Get(tid, key2))

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
	assert.Equal(t, []byte("value of key1"), tbl.Get(tid, key1).val)
	assert.Nil(t, tbl.Get(tid, key2))
}

func TestStagingTable_NilTidAction(t *testing.T) {
	key1 := []byte("key1")
	key2 := []byte("key2")
	tid := "tid1"

	tbl := NewStagingTable()

	// Get and Put.
	value := tbl.Get(nil, key1)
	assert.Nil(t, value)

	value = tbl.Put(nil, key1, []byte("value of key1"))
	assert.Equal(t, []byte("value of key1"), value.val)
	assert.Equal(t, 0, value.old)
	assert.Equal(t, 1, value.new)
	assert.False(t, value.deleted)
	assert.True(t, value.dirty)
	assert.True(t, value.initialized)

	value = tbl.Get(nil, key1)
	assert.Equal(t, []byte("value of key1"), value.val)
	assert.Equal(t, 0, value.old)
	assert.Equal(t, 1, value.new)
	assert.False(t, value.deleted)
	assert.True(t, value.dirty)
	assert.True(t, value.initialized)

	// Put.
	value = tbl.Put(nil, key2, []byte("value of key2"))
	assert.Equal(t, []byte("value of key2"), value.val)
	assert.Equal(t, 0, value.old)
	assert.Equal(t, 1, value.new)
	assert.False(t, value.deleted)
	assert.True(t, value.dirty)
	assert.True(t, value.initialized)

	value = tbl.Get(nil, key2)
	assert.Equal(t, []byte("value of key2"), value.val)
	assert.Equal(t, 0, value.old)
	assert.Equal(t, 1, value.new)
	assert.False(t, value.deleted)
	assert.True(t, value.dirty)
	assert.True(t, value.initialized)

	// tid get.
	value = tbl.Get(tid, key1)
	assert.Equal(t, []byte("value of key1"), value.val)
	assert.Equal(t, 1, value.old)
	assert.Equal(t, 2, value.new)
	assert.False(t, value.deleted)
	assert.False(t, value.dirty)
	assert.True(t, value.initialized)

	tbl.Del(tid, key1)

	// tid merge.
	_, err := tbl.MergeToFinal(tid)
	assert.Nil(t, err)

	// Get.
	value = tbl.Get(nil, key1)
	assert.True(t, value.deleted)
}
