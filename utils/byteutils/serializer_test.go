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

package byteutils

import (
	"reflect"
	"testing"
)

func TestRLPSerializerStruct(t *testing.T) {
	type args struct {
		val interface{}
		res interface{}
	}
	type Message struct {
		Name string
		Time uint
		Tags []string
		Body [][]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"full struct",
			args{Message{"Alice", 10, []string{"cat", "dog"}, [][]string{[]string{"world"}}}, &Message{}},
			false,
		},
		{
			"empty struct",
			args{Message{"", 10, []string{}, [][]string{}}, &Message{}},
			false,
		},
		{
			"incomplete struct",
			args{Message{"Alice", 10, []string{}, [][]string{[]string{"world"}, []string{}}}, &Message{}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s Serializable = &RLPSerializer{}
			got, err1 := s.Serialize(tt.args.val)
			err2 := s.Deserialize(got, tt.args.res)
			if (err1 != nil) != tt.wantErr {
				t.Errorf("JSONSerializer.EncodeToBytes() error = %v, wantErr %v", err1, tt.wantErr)
				return
			}
			if (err2 != nil) != tt.wantErr {
				t.Errorf("JSONSerializer.EncodeToBytes() error = %v, wantErr %v", err2, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(*(tt.args.res.(*Message)), tt.args.val) {
				t.Errorf("JSONSerializer.EncodeToBytes() = %v, want %v", *(tt.args.res.(*Message)), tt.args.val)
			}

		})
	}
}
func TestRLPSerializerArray(t *testing.T) {
	type args struct {
		val interface{}
		res interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"full array",
			args{[][]byte{[]byte("car"), []byte("dog")}, &[][]byte{}},
			false,
		},
		{
			"empty array",
			args{[][]byte{}, &[][]byte{}},
			false,
		},
		{
			"incomplete array",
			args{[][]byte{[]byte("car"), []byte{}, []byte("dog")}, &[][]byte{}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s Serializable = &RLPSerializer{}
			got, err1 := s.Serialize(tt.args.val)
			err2 := s.Deserialize(got, tt.args.res)
			if (err1 != nil) != tt.wantErr {
				t.Errorf("JSONSerializer.EncodeToBytes() error = %v, wantErr %v", err1, tt.wantErr)
				return
			}
			if (err2 != nil) != tt.wantErr {
				t.Errorf("JSONSerializer.EncodeToBytes() error = %v, wantErr %v", err2, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(*(tt.args.res.(*[][]byte)), tt.args.val) {
				t.Errorf("JSONSerializer.EncodeToBytes() = %v, want %v", *(tt.args.res.(*[][]byte)), tt.args.val)
			}
		})
	}
}

func TestJSONSerializerStruct(t *testing.T) {
	type args struct {
		val interface{}
		res interface{}
	}
	type Message struct {
		Name string
		Time int64
		Tags []string
		Map  map[int]string
		Body [][]string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"full struct",
			args{Message{"Alice", 1294706395881547000, []string{"cat", "dog"}, map[int]string{1: "hello"}, [][]string{[]string{"world"}}}, &Message{}},
			false,
		},
		{
			"empty struct",
			args{Message{}, &Message{}},
			false,
		},
		{
			"incomplete struct",
			args{Message{"Alice", 1294706395881547000, nil, map[int]string{1: "hello"}, [][]string{[]string{"world"}, nil}}, &Message{}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s Serializable = &JSONSerializer{}
			got, err1 := s.Serialize(tt.args.val)
			err2 := s.Deserialize(got, tt.args.res)
			if (err1 != nil) != tt.wantErr {
				t.Errorf("JSONSerializer.EncodeToBytes() error = %v, wantErr %v", err1, tt.wantErr)
				return
			}
			if (err2 != nil) != tt.wantErr {
				t.Errorf("JSONSerializer.EncodeToBytes() error = %v, wantErr %v", err2, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(*(tt.args.res.(*Message)), tt.args.val) {
				t.Errorf("JSONSerializer.EncodeToBytes() = %v, want %v", tt.args.res, tt.args.val)
			}
		})
	}
}

func TestJSONSerializerArray(t *testing.T) {
	type args struct {
		val interface{}
		res interface{}
	}
	type Message struct {
		Name string
		Time int64
		Tags []string
		Map  map[int]string
		Body [][]string
	}
	emptyBranchVal := [16][]byte{}
	emptyBranch := emptyBranchVal[:]
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"full array",
			args{[][]byte{[]byte("car"), []byte("dog")}, &[][]byte{}},
			false,
		},
		{
			"empty array",
			args{emptyBranch, &[][]byte{}},
			false,
		},
		{
			"incomplete array",
			args{[][]byte{[]byte("car"), nil, []byte("dog")}, &[][]byte{}},
			false,
		},
		{
			"nil",
			args{[][]byte{nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}, &[][]byte{}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s Serializable = &JSONSerializer{}
			got, err1 := s.Serialize(tt.args.val)
			err2 := s.Deserialize(got, tt.args.res)
			if (err1 != nil) != tt.wantErr {
				t.Errorf("JSONSerializer.EncodeToBytes() error = %v, wantErr %v", err1, tt.wantErr)
				return
			}
			if (err2 != nil) != tt.wantErr {
				t.Errorf("JSONSerializer.EncodeToBytes() error = %v, wantErr %v", err2, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(*(tt.args.res.(*[][]byte)), tt.args.val) {
				t.Errorf("JSONSerializer.EncodeToBytes() = %v, want %v", tt.args.res, tt.args.val)
			}
		})
	}
}

func Test_compress(t *testing.T) {
	type args struct {
		val []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"case 1",
			args{[]byte("this is a compress test case")},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err1 := compress(tt.args.val)
			res, err2 := uncompress(got)
			if (err1 != nil) != tt.wantErr {
				t.Errorf("compress() error = %v, wantErr %v", err1, tt.wantErr)
				return
			}
			if (err2 != nil) != tt.wantErr {
				t.Errorf("compress() error = %v, wantErr %v", err2, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(res, tt.args.val) {
				t.Errorf("compress() = %v, want %v", res, tt.args.val)
			}
		})
	}
}
