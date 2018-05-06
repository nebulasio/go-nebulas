package storage

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
