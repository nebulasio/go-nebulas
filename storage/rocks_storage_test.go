package storage

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"os"
	"sync"
	"testing"
)

const TestRocksDB  = "rocks.db"

func TestNewRocksStorage(t *testing.T) {
	s, err := NewRocksStorage(TestRocksDB)
	assert.Nil(t, err)

	key := []byte("key")
	value := []byte("value")
	err = s.Put(key, value)
	assert.Nil(t, err)

	val, err := s.Get(key)
	assert.Equal(t, val, value)

	s.Close()
	os.RemoveAll(TestRocksDB)
}

func TestNewRocksStorageWithCF(t *testing.T) {

	cfNames := []string{"account","event"}

	tests := []struct {
		name  string
		cfName string
		key   []byte
		value []byte
		err error
	}{
		{"default", "", []byte("key1"), []byte("value1"), nil},
		{"account default", "default", []byte("key1"), []byte("value1"), nil},
		{"account", cfNames[0], []byte("key1"), []byte("value11"), nil},
		{"event", cfNames[1], []byte("key11"), []byte("value111"), nil},
		{"tx ", "tx", []byte("key111"), []byte("value1111"), errors.New("column family name not found")},
	}

	db, err := NewRocksStorageWithCF(TestRocksDB, cfNames)
	assert.Nil(t, err)

	wg := new(sync.WaitGroup)
	wg.Add(len(tests))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			defer wg.Done()

			if len(tt.cfName) > 0 {
				err := db.PutCF(tt.cfName, tt.key, tt.value)
				assert.Equal(t, tt.err, err)

				if err == nil {
					val, err := db.GetCF(tt.cfName, tt.key)
					assert.Nil(t, err)
					assert.Equal(t, tt.value, val)
				}
			} else {
				err := db.Put(tt.key, tt.value)
				assert.Equal(t, tt.err, err)

				if err == nil {
					val, err := db.Get(tt.key)
					assert.Nil(t, err)
					assert.Equal(t, tt.value, val)
				}
			}
		})
	}

	wg.Wait()

	db.Close()
	os.RemoveAll(TestRocksDB)
}

func TestRocksStorage_CreateColumn(t *testing.T) {
	cfNames := []string{"account","event"}

	db, err := NewRocksStorageWithCF(TestRocksDB, cfNames)
	assert.Nil(t, err)

	err = db.CreateColumn("account1")
	assert.Nil(t, err)
	err = db.CreateColumn("account1")
	assert.Equal(t, errors.New("Invalid argument: Column family already exists"), err)

	db.Close()
	os.RemoveAll(TestRocksDB)
}