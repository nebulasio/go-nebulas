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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nebulasio/go-nebulas/storage"
)

func TestMVCCDB_NewMVCCDB(t *testing.T) {
	storage, _ := storage.NewMemoryStorage()
	db, err := NewMVCCDB(storage)
	assert.Nil(t, err)

	assert.False(t, db.isInTransaction)
	assert.False(t, db.isPreparedDB)
}

func TestMVCCDB_DirectOpts(t *testing.T) {
	storage, _ := storage.NewMemoryStorage()
	db, _ := NewMVCCDB(storage)

	key := []byte("key")
	val := []byte("val")

	v, err := db.getFromStorage(key)
	assert.Nil(t, v)
	assert.NotNil(t, err)

	err = db.putToStorage(key, val)
	assert.Nil(t, err)

	v, err = db.getFromStorage(key)
	assert.Nil(t, err)
	assert.Equal(t, val, v)

	err = db.delFromStorage(key)
	assert.Nil(t, err)

	v, err = db.getFromStorage(key)
	assert.Nil(t, v)
	assert.NotNil(t, err)
}

func TestMVCCDB_OptsWithoutTransaction(t *testing.T) {
	storage, _ := storage.NewMemoryStorage()
	db, _ := NewMVCCDB(storage)

	key := []byte("key")
	val := []byte("val")

	v, err := db.Get(key)
	assert.Nil(t, v)
	assert.NotNil(t, err)

	err = db.Put(key, val)
	assert.Nil(t, err)

	v, err = db.Get(key)
	assert.Nil(t, err)
	assert.Equal(t, val, v)

	err = db.Del(key)
	assert.Nil(t, err)

	v, err = db.Get(key)
	assert.Nil(t, v)
	assert.NotNil(t, err)
}

func TestMVCCDB_OptsWithinTransaction(t *testing.T) {
	store, _ := storage.NewMemoryStorage()
	db, _ := NewMVCCDB(store)

	key := []byte("key")
	val := []byte("val")

	err := db.Begin()
	assert.Nil(t, err)
	assert.True(t, db.isInTransaction)

	// unsupported nested transaction.
	err = db.Begin()
	assert.Equal(t, err, ErrUnsupportedNestedTransaction)

	v, err := db.Get(key)
	assert.Nil(t, v)
	assert.Equal(t, err, storage.ErrKeyNotFound)

	err = db.Put(key, val)
	assert.Nil(t, err)

	{
		// other MVCCDB can't read before commit.
		db2, _ := NewMVCCDB(store)
		v, err := db2.Get(key)
		assert.Nil(t, v)
		assert.Equal(t, err, storage.ErrKeyNotFound)
	}

	// commit.
	db.Commit()

	// read.
	v, err = db.Get(key)
	assert.Nil(t, err)
	assert.Equal(t, val, v)

	// begin.
	err = db.Begin()
	assert.Nil(t, err)

	err = db.Del(key)
	assert.Nil(t, err)

	{
		// other MVCCDB read old value.
		db2, _ := NewMVCCDB(store)
		v, err := db2.Get(key)
		assert.Equal(t, val, v)
		assert.Nil(t, err)
	}

	v, err = db.Get(key)
	assert.Nil(t, v)
	assert.Equal(t, err, storage.ErrKeyNotFound)

	// rollback.
	db.RollBack()

	// read.
	v, err = db.Get(key)
	assert.Nil(t, err)
	assert.Equal(t, val, v)

	{
		// other MVCCDB read old value.
		db2, _ := NewMVCCDB(store)
		v, err := db2.Get(key)
		assert.Equal(t, val, v)
		assert.Nil(t, err)
	}

	// begin.
	err = db.Begin()
	assert.Nil(t, err)

	err = db.Del(key)
	assert.Nil(t, err)

	// commit.
	db.Commit()

	// read.
	v, err = db.Get(key)
	assert.Nil(t, v)
	assert.Equal(t, err, storage.ErrKeyNotFound)

	{
		// other MVCCDB read nil.
		db2, _ := NewMVCCDB(store)
		v, err := db2.Get(key)
		assert.Nil(t, v)
		assert.Equal(t, err, storage.ErrKeyNotFound)
	}
}
