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

package bytes

import (
	"reflect"
	"testing"
)

func TestEncode(t *testing.T) {
	type args struct {
		s   interface{}
		enc Encoder
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Encode(tt.args.s, tt.args.enc)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	type args struct {
		data []byte
		dec  Decoder
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Decode(tt.args.data, tt.args.dec); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Decode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHex(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a",
			args{[]byte{167, 255, 198, 248, 191, 30, 215, 102, 81, 193, 71, 86, 160, 97, 214, 98, 245, 128, 255, 77, 228, 59, 73, 250, 130, 216, 10, 75, 128, 248, 67, 74}},
			"a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a",
		},
		{
			"3550aba97492de38af3066f0157fc532db6791b37d53262ce7688dcc5d461856",
			args{[]byte{53, 80, 171, 169, 116, 146, 222, 56, 175, 48, 102, 240, 21, 127, 197, 50, 219, 103, 145, 179, 125, 83, 38, 44, 231, 104, 141, 204, 93, 70, 24, 86}},
			"3550aba97492de38af3066f0157fc532db6791b37d53262ce7688dcc5d461856",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Hex(tt.args.data); got != tt.want {
				t.Errorf("Hex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromHex(t *testing.T) {
	type args struct {
		data string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			"a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a",
			args{"a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a"},
			[]byte{167, 255, 198, 248, 191, 30, 215, 102, 81, 193, 71, 86, 160, 97, 214, 98, 245, 128, 255, 77, 228, 59, 73, 250, 130, 216, 10, 75, 128, 248, 67, 74},
			false,
		},
		{
			"3550aba97492de38af3066f0157fc532db6791b37d53262ce7688dcc5d461856",
			args{"3550aba97492de38af3066f0157fc532db6791b37d53262ce7688dcc5d461856"},
			[]byte{53, 80, 171, 169, 116, 146, 222, 56, 175, 48, 102, 240, 21, 127, 197, 50, 219, 103, 145, 179, 125, 83, 38, 44, 231, 104, 141, 204, 93, 70, 24, 86},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromHex(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromHex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromHex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUint64(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want uint64
	}{
		{
			"0",
			args{[]byte{0, 0, 0, 0, 0, 0, 0, 0}},
			uint64(0),
		},
		{
			"1",
			args{[]byte{0, 0, 0, 0, 0, 0, 0, 1}},
			uint64(1),
		},
		{
			"923",
			args{[]byte{0, 0, 0, 0, 0, 0, 0, 1}},
			uint64(1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Uint64(tt.args.data); got != tt.want {
				t.Errorf("Uint64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromUint64(t *testing.T) {
	type args struct {
		v uint64
	}
	tests := []struct {
		name  string
		args  args
		wantB [8]byte
	}{
		{
			"0",
			args{uint64(0)},
			[8]byte{0, 0, 0, 0, 0, 0, 0, 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotB := FromUint64(tt.args.v); !reflect.DeepEqual(gotB, tt.wantB) {
				t.Errorf("FromUint64() = %v, want %v", gotB, tt.wantB)
			}
		})
	}
}

func TestUint32(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Uint32(tt.args.data); got != tt.want {
				t.Errorf("Uint32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromUint32(t *testing.T) {
	type args struct {
		v uint32
	}
	tests := []struct {
		name  string
		args  args
		wantB [4]byte
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotB := FromUint32(tt.args.v); !reflect.DeepEqual(gotB, tt.wantB) {
				t.Errorf("FromUint32() = %v, want %v", gotB, tt.wantB)
			}
		})
	}
}

func TestUint16(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want uint16
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Uint16(tt.args.data); got != tt.want {
				t.Errorf("Uint16() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromUint16(t *testing.T) {
	type args struct {
		v uint16
	}
	tests := []struct {
		name  string
		args  args
		wantB [2]byte
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotB := FromUint16(tt.args.v); !reflect.DeepEqual(gotB, tt.wantB) {
				t.Errorf("FromUint16() = %v, want %v", gotB, tt.wantB)
			}
		})
	}
}

func TestInt64(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Int64(tt.args.data); got != tt.want {
				t.Errorf("Int64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromInt64(t *testing.T) {
	type args struct {
		v int64
	}
	tests := []struct {
		name  string
		args  args
		wantB [8]byte
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotB := FromInt64(tt.args.v); !reflect.DeepEqual(gotB, tt.wantB) {
				t.Errorf("FromInt64() = %v, want %v", gotB, tt.wantB)
			}
		})
	}
}

func TestInt32(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want int32
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Int32(tt.args.data); got != tt.want {
				t.Errorf("Int32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromInt32(t *testing.T) {
	type args struct {
		v int32
	}
	tests := []struct {
		name  string
		args  args
		wantB [4]byte
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotB := FromInt32(tt.args.v); !reflect.DeepEqual(gotB, tt.wantB) {
				t.Errorf("FromInt32() = %v, want %v", gotB, tt.wantB)
			}
		})
	}
}

func TestInt16(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want int16
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Int16(tt.args.data); got != tt.want {
				t.Errorf("Int16() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromInt16(t *testing.T) {
	type args struct {
		v int16
	}
	tests := []struct {
		name  string
		args  args
		wantB [2]byte
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotB := FromInt16(tt.args.v); !reflect.DeepEqual(gotB, tt.wantB) {
				t.Errorf("FromInt16() = %v, want %v", gotB, tt.wantB)
			}
		})
	}
}
