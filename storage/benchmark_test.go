// Copyright (C) 2017 go-nebulas authors
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
	"math/rand"
	"testing"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/stretchr/testify/assert"
	"time"
	"github.com/nebulasio/go-nebulas/crypto/hash"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randBytes(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return b
}

func TestLeveldbBenchmark(t *testing.T) {
	file := "leveldb_benchmark.db"
	db, err := leveldb.OpenFile(file, &opt.Options{
		OpenFilesCacheCapacity: 500,
		BlockCacheCapacity:     8 * opt.MiB,
		WriteBuffer:            4 * opt.MiB,
		Filter:                 filter.NewBloomFilter(10),
	})
	assert.Nil(t, err)

	tests := []struct {
		name  string
		count int
	}{
		//{"1", 1},
		//{"2", 10000},
		{"3", 100000},
		//{"1000000", 1000000},
		//{"4000000", 4000000},
	}

	count := int(0)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			start := time.Now().UnixNano()

			count = count + tt.count
			for i := int(0); i < tt.count; i++ {
				err := db.Put(hash.Sha3256(byteutils.FromInt32(int32(i))), randBytes(i%1024), nil)
				assert.Nil(t, err)
				db.Get(hash.Sha3256(byteutils.FromInt32(int32(rand.Intn(tt.count)))), nil)
			}

			duration := time.Now().UnixNano() - start
			assert.Nil(t, err)

			t.Log("count:", count, "duration:", duration, " Nano", "Read/write speed:",duration/int64(tt.count), "ns/count")
		})
	}
	db.Close()
	//os.Remove(file)
}

func TestRocksdbBenchmark(t *testing.T) {
	file := "rocksdb_benchmark.db"

	db, err := NewRocksStorage(file)
	assert.Nil(t, err)

	tests := []struct {
		name  string
		count int
	}{
		//{"1", 1},
		//{"2", 10000},
		//{"3", 100000},
		{"1000000", 1000000},
		//{"4000000", 4000000},
	}

	count := int(0)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			start := time.Now().UnixNano()

			count = count + tt.count
			for i := int(0); i < tt.count; i++ {
				err := db.Put(hash.Sha3256(byteutils.FromInt32(int32(i))), randBytes(i%1024))
				assert.Nil(t, err)
				db.Get(hash.Sha3256(byteutils.FromInt32(int32(rand.Intn(tt.count)))))
			}

			duration := time.Now().UnixNano() - start
			assert.Nil(t, err)

			t.Log("count:", count, "duration:", duration, " Nano", "Read/write speed:",duration/int64(tt.count), "ns/count")
		})
	}
	db.Close()
	//os.Remove(file)
}

func TestCachedbBenchmark(t *testing.T) {
	file := "cache_rocksdb_benchmark.db"

	rocks, err := NewRocksStorage(file)
	assert.Nil(t, err)

	db, err := NewCacheStorage(rocks, LRUCacheSize)
	assert.Nil(t, err)

	tests := []struct {
		name  string
		count int
	}{
		//{"1", 1},
		//{"2", 10000},
		//{"3", 100000},
		{"1000000", 1000000},
		//{"4000000", 4000000},
	}

	count := int(0)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			start := time.Now().UnixNano()

			count = count + tt.count
			for i := int(0); i < tt.count; i++ {
				key := hash.Sha3256(byteutils.FromInt32(int32(i)))
				err := db.Put(key, randBytes(i%1024))
				assert.Nil(t, err)
				db.Get(hash.Sha3256(byteutils.FromInt32(int32(rand.Intn(tt.count)))))
			}

			duration := time.Now().UnixNano() - start
			assert.Nil(t, err)

			t.Log("count:", count, "duration:", duration, " Nano", "Read/write speed:",duration/int64(tt.count), "ns/count")
		})
	}
	db.Close()
	//os.Remove(file)
}


