package storage

import (
	"errors"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testData struct {
	name   string
	cfName string
	key    []byte
	value  []byte
	err    error
}

func TestNewRocksStorage(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *RocksStorage
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRocksStorage(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRocksStorage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRocksStorage() = %v, want %v", got, tt.want)
			}
		})
	}

	s, err := NewRocksStorage("./rock.db/")
	assert.Nil(t, err)

	key := []byte("key")
	value := []byte("value")
	err = s.Put(key, value)
	assert.Nil(t, err)

	val, err := s.Get(key)
	assert.Equal(t, val, value)
}

func TestNewRocksStorageWithCF(t *testing.T) {
	dbName := "rocksdb"
	cfNames := []string{"default", "tx", "acc", "event"}

	tests := []testData{
		{"default", cfNames[0], []byte("key1"), []byte("value1"), nil},
		{"tx", cfNames[1], []byte("key2"), []byte("value2"), nil},
		{"tx redo", cfNames[1], []byte("key2"), []byte("value2"), nil},
		{"acc", cfNames[2], []byte("key3"), []byte("value3"), nil},
		{"event", cfNames[3], []byte("key4"), []byte("value4"), nil},
		{"err", "err", []byte("key5"), []byte("value5"), errors.New("column family not found")},
	}

	testsDefault := []testData{
		{"noCF1", "default", []byte("noCFkey111"), []byte("noCFvalue111"), nil},
		{"noCF2", "default", []byte("noCFkey222"), []byte("noCFvalue222"), nil},
	}

	testsDefaultRedo := []testData{
		{"noCF1 redo", "default", []byte("noCFkey111"), []byte("noCFvalue111 redo"), nil},
		{"noCF2 redo", "default", []byte("noCFkey222"), []byte("noCFvalue222 redo"), nil},
	}
	//no column families read and wirte
	dbold, err := NewRocksStorage(dbName)
	assert.Nil(t, err)
	wg := new(sync.WaitGroup)
	wg.Add(len(testsDefault))
	for _, tt := range testsDefault {
		t.Run(tt.name, func(t *testing.T) {
			defer wg.Done()

			err := dbold.Put(tt.key, tt.value)
			assert.Equal(t, tt.err, err)

			if err == nil {
				val, err := dbold.Get(tt.key)
				assert.Nil(t, err)
				assert.Equal(t, tt.value, val)
			}

		})
	}
	wg.Wait()
	dbold.Close()

	// using column families
	db, err := NewRocksStorageWithCF(dbName, cfNames)
	assert.Nil(t, err)
	wg.Add(len(tests))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer wg.Done()

			err := db.PutCF(tt.cfName, tt.key, tt.value)
			assert.Equal(t, tt.err, err)

			if err == nil {
				val, err := db.GetCF(tt.cfName, tt.key)
				assert.Nil(t, err)
				assert.Equal(t, tt.value, val)
			}

		})
	}

	wg.Wait()
	db.Close()

	//Compatibility test
	db, err = NewRocksStorageWithCF(dbName, cfNames)
	assert.Nil(t, err)
	wg.Add(len(testsDefault))
	for _, tt := range testsDefault {
		t.Run(tt.name, func(t *testing.T) {
			defer wg.Done()
			val, err := db.GetCF(tt.cfName, tt.key)
			assert.Nil(t, err)
			assert.Equal(t, tt.value, val)
		})
	}
	wg.Wait()

	wg.Add(len(testsDefaultRedo))
	for _, tt := range testsDefaultRedo {
		t.Run(tt.name, func(t *testing.T) {
			defer wg.Done()
			err := db.PutCF(tt.cfName, tt.key, tt.value)
			assert.Equal(t, tt.err, err)

			if err == nil {
				val, err := db.GetCF(tt.cfName, tt.key)
				assert.Nil(t, err)
				assert.Equal(t, tt.value, val)
			}
		})
	}

	wg.Wait()
	db.Close()

	os.RemoveAll(dbName)
}
