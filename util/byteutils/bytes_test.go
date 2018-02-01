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
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStringEncoderAndDecoder struct {
}

func (o *testStringEncoderAndDecoder) EncodeToBytes(s interface{}) ([]byte, error) {
	str := s.(string)

	if len(str) == 0 {
		return nil, errors.New("s must be string")
	}

	return []byte(str), nil
}

func (o *testStringEncoderAndDecoder) DecodeFromBytes(data []byte) (interface{}, error) {
	return string(data), nil
}

func TestEncode(t *testing.T) {
	o := &testStringEncoderAndDecoder{}

	src := "Hello, world"
	want := []byte{72, 101, 108, 108, 111, 44, 32, 119, 111, 114, 108, 100}

	ret, err := Encode(src, o)
	assert.Nil(t, err, "err should be nil")
	assert.Equal(t, want, ret, "Encode() = %v, want %v", ret, want)
}

func TestDecode(t *testing.T) {
	o := &testStringEncoderAndDecoder{}

	src := []byte{72, 101, 108, 108, 111, 44, 32, 119, 111, 114, 108, 100}
	want := "Hello, world"

	ret, err := Decode(src, o)
	assert.Nil(t, err, "err should be nil")
	assert.Equal(t, want, ret, "Decode() = \"%v\", want \"%v\"", ret, want)
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
		{
			"blank string",
			args{[]byte{}},
			"",
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
		{
			"blank string",
			args{""},
			[]byte{},
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
			"1024",
			args{[]byte{0, 0, 0, 0, 0, 0, 4, 0}},
			uint64(1024),
		},
		{
			"Uint64.Max",
			args{[]byte{255, 255, 255, 255, 255, 255, 255, 255}},
			uint64(18446744073709551615),
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
		wantB []byte
	}{
		{
			"0",
			args{uint64(0)},
			[]byte{0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			"1024",
			args{uint64(1024)},
			[]byte{0, 0, 0, 0, 0, 0, 4, 0},
		},
		{
			"Uint64.Max",
			args{uint64(18446744073709551615)},
			[]byte{255, 255, 255, 255, 255, 255, 255, 255},
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
		{
			"0",
			args{[]byte{0, 0, 0, 0}},
			uint32(0),
		},
		{
			"1024",
			args{[]byte{0, 0, 4, 0}},
			uint32(1024),
		},
		{
			"Uint32.Max",
			args{[]byte{255, 255, 255, 255}},
			uint32(4294967295),
		},
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
		wantB []byte
	}{
		{
			"0",
			args{uint32(0)},
			[]byte{0, 0, 0, 0},
		},
		{
			"1024",
			args{uint32(1024)},
			[]byte{0, 0, 4, 0},
		},
		{
			"Uint32.Max",
			args{uint32(4294967295)},
			[]byte{255, 255, 255, 255},
		},
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
		{
			"0",
			args{[]byte{0, 0}},
			uint16(0),
		},
		{
			"1024",
			args{[]byte{4, 0}},
			uint16(1024),
		},
		{
			"Uint16.Max",
			args{[]byte{255, 255}},
			uint16(65535),
		},
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
		wantB []byte
	}{
		{
			"0",
			args{uint16(0)},
			[]byte{0, 0},
		},
		{
			"1024",
			args{uint16(1024)},
			[]byte{4, 0},
		},
		{
			"Uint16.Max",
			args{uint16(65535)},
			[]byte{255, 255},
		},
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
		{
			"0",
			args{[]byte{0, 0, 0, 0, 0, 0, 0, 0}},
			int64(0),
		},
		{
			"1024",
			args{[]byte{0, 0, 0, 0, 0, 0, 4, 0}},
			int64(1024),
		},
		{
			"-1024",
			args{[]byte{255, 255, 255, 255, 255, 255, 252, 0}},
			int64(-1024),
		},
		{
			"Int64.Max",
			args{[]byte{127, 255, 255, 255, 255, 255, 255, 255}},
			int64(9223372036854775807),
		},
		{
			"Int64.Min",
			args{[]byte{128, 0, 0, 0, 0, 0, 0, 0}},
			int64(-9223372036854775808),
		},
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
		wantB []byte
	}{
		{
			"0",
			args{int64(0)},
			[]byte{0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			"1024",
			args{int64(1024)},
			[]byte{0, 0, 0, 0, 0, 0, 4, 0},
		},
		{
			"-1024",
			args{int64(-1024)},
			[]byte{255, 255, 255, 255, 255, 255, 252, 0},
		},
		{
			"Int64.Max",
			args{int64(9223372036854775807)},
			[]byte{127, 255, 255, 255, 255, 255, 255, 255},
		},
		{
			"Int64.Min",
			args{int64(-9223372036854775808)},
			[]byte{128, 0, 0, 0, 0, 0, 0, 0},
		},
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
		{
			"0",
			args{[]byte{0, 0, 0, 0}},
			int32(0),
		},
		{
			"1024",
			args{[]byte{0, 0, 4, 0}},
			int32(1024),
		},
		{
			"-1024",
			args{[]byte{255, 255, 252, 0}},
			int32(-1024),
		},
		{
			"Int32.Max",
			args{[]byte{127, 255, 255, 255}},
			int32(2147483647),
		},
		{
			"Int32.Min",
			args{[]byte{128, 0, 0, 0}},
			int32(-2147483648),
		},
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
		wantB []byte
	}{
		{
			"0",
			args{int32(0)},
			[]byte{0, 0, 0, 0},
		},
		{
			"1024",
			args{int32(1024)},
			[]byte{0, 0, 4, 0},
		},
		{
			"-1024",
			args{int32(-1024)},
			[]byte{255, 255, 252, 0},
		},
		{
			"Int32.Max",
			args{int32(2147483647)},
			[]byte{127, 255, 255, 255},
		},
		{
			"Int32.Min",
			args{int32(-2147483648)},
			[]byte{128, 0, 0, 0},
		},
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
		{
			"0",
			args{[]byte{0, 0}},
			int16(0),
		},
		{
			"1024",
			args{[]byte{4, 0}},
			int16(1024),
		},
		{
			"-1024",
			args{[]byte{252, 0}},
			int16(-1024),
		},
		{
			"Int16.Max",
			args{[]byte{127, 255}},
			int16(32767),
		},
		{
			"Int16.Min",
			args{[]byte{128, 0}},
			int16(-32768),
		},
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
		wantB []byte
	}{
		{
			"0",
			args{int16(0)},
			[]byte{0, 0},
		},
		{
			"1024",
			args{int16(1024)},
			[]byte{4, 0},
		},
		{
			"-1024",
			args{int16(-1024)},
			[]byte{252, 0},
		},
		{
			"Int16.Max",
			args{int16(32767)},
			[]byte{127, 255},
		},
		{
			"Int16.Min",
			args{int16(-32768)},
			[]byte{128, 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotB := FromInt16(tt.args.v); !reflect.DeepEqual(gotB, tt.wantB) {
				t.Errorf("FromInt16() = %v, want %v", gotB, tt.wantB)
			}
		})
	}
}
