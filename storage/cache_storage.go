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

package storage

import (
	"github.com/hashicorp/golang-lru"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)


// const val
const (
	LRUCacheSize  = 500 * 1024 *1024
)

// CacheStorage the cache storage
type CacheStorage struct {
	storage Storage
	cache *lru.Cache
}

// NewCacheStorageWithPath returns a cache storage
func NewCacheStorageWithPath(path string) (*CacheStorage, error) {
	storage, err := NewRocksStorage(path)
	if err != nil {
		return nil, err
	}

	cache, err := lru.New(LRUCacheSize)
	if err != nil {
		return nil, err
	}
	return &CacheStorage{storage:storage, cache:cache}, nil
}

// NewCacheStorage returns a cache storage
func NewCacheStorage(storage Storage, size int) (*CacheStorage, error) {
	cache, err := lru.New(size)
	if err != nil {
		return nil, err
	}
	return &CacheStorage{storage:storage, cache:cache}, nil
}

// Get return the value to the key in Storage.
func (s *CacheStorage) Get(key []byte) ([]byte, error) {
	value, result := s.cache.Get(byteutils.Hex(key))
	if !result {
		return s.storage.Get(key)
	}
	return value.([]byte), nil
}

// Put put the key-value entry to Storage.
func (s *CacheStorage) Put(key []byte, value []byte) error {
	s.cache.Add(byteutils.Hex(key), value)
	return s.storage.Put(key, value)
}

// Del delete the key entry in Storage.
func (s *CacheStorage) Del(key []byte) error {
	s.cache.Remove(byteutils.Hex(key))
	return s.storage.Del(key)
}

// EnableBatch enable batch write.
func (s *CacheStorage) EnableBatch() {
	s.storage.EnableBatch()
}

// DisableBatch disable batch write.
func (s *CacheStorage) DisableBatch() {
	s.storage.DisableBatch()
}

// Flush write and flush pending batch write.
func (s *CacheStorage) Flush() error {
	return s.storage.Flush()
}
