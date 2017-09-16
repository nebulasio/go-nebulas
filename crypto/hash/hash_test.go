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

package hash

import (
	"reflect"
	"testing"
)

func TestSha3256(t *testing.T) {
	type args struct {
		bytes []byte
	}
	tests := []struct {
		name       string
		args       args
		wantDigest []byte
	}{
		{
			"blank string",
			args{[]byte("")},
			[]byte{167, 255, 198, 248, 191, 30, 215, 102, 81, 193, 71, 86, 160, 97, 214, 98, 245, 128, 255, 77, 228, 59, 73, 250, 130, 216, 10, 75, 128, 248, 67, 74},
		},
		{
			"Hello, world",
			args{[]byte("Hello, world")},
			[]byte{53, 80, 171, 169, 116, 146, 222, 56, 175, 48, 102, 240, 21, 127, 197, 50, 219, 103, 145, 179, 125, 83, 38, 44, 231, 104, 141, 204, 93, 70, 24, 86},
		},
		{
			"https://nebulas.io",
			args{[]byte("https://nebulas.io")},
			[]byte{94, 159, 238, 157, 152, 227, 18, 248, 53, 8, 13, 247, 231, 21, 17, 14, 172, 34, 192, 157, 24, 175, 119, 254, 126, 208, 174, 17, 14, 77, 1, 55},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotDigest := Sha3256(tt.args.bytes); !reflect.DeepEqual(gotDigest, tt.wantDigest) {
				t.Errorf("Sha3256() = %v, want %v", gotDigest, tt.wantDigest)
			}
		})
	}
}

func TestRipemd160(t *testing.T) {
	type args struct {
		bytes []byte
	}
	tests := []struct {
		name       string
		args       args
		wantDigest []byte
	}{
		{
			"blank string",
			args{[]byte("")},
			[]byte{156, 17, 133, 165, 197, 233, 252, 84, 97, 40, 8, 151, 126, 232, 245, 72, 178, 37, 141, 49},
		},
		{
			"The quick brown fox jumps over the lazy dog",
			args{[]byte("The quick brown fox jumps over the lazy dog")},
			[]byte{55, 243, 50, 246, 141, 183, 123, 217, 215, 237, 212, 150, 149, 113, 173, 103, 28, 249, 221, 59},
		},
		{
			"The quick brown fox jumps over the lazy cog",
			args{[]byte("The quick brown fox jumps over the lazy cog")},
			[]byte{19, 32, 114, 223, 105, 9, 51, 131, 94, 184, 182, 173, 11, 119, 231, 182, 241, 74, 202, 215},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotDigest := Ripemd160(tt.args.bytes); !reflect.DeepEqual(gotDigest, tt.wantDigest) {
				t.Errorf("Ripemd160() = %v, want %v", gotDigest, tt.wantDigest)
			}
		})
	}
}

func TestSha256(t *testing.T) {
	type args struct {
		bytes []byte
	}
	tests := []struct {
		name       string
		args       args
		wantDigest []byte
	}{
		{
			"",
			args{[]byte("")},
			[]byte{227, 176, 196, 66, 152, 252, 28, 20, 154, 251, 244, 200, 153, 111, 185, 36, 39, 174, 65, 228, 100, 155, 147, 76, 164, 149, 153, 27, 120, 82, 184, 85},
		},
		{
			"Hello, world!0",
			args{[]byte("Hello, world!0")},
			[]byte{19, 18, 175, 23, 140, 37, 63, 132, 2, 141, 72, 10, 106, 220, 30, 37, 232, 28, 170, 68, 199, 73, 236, 129, 151, 97, 146, 226, 236, 147, 76, 100},
		},
		{
			"Hello, world!4250",
			args{[]byte("Hello, world!4250")},
			[]byte{0, 0, 195, 175, 66, 252, 49, 16, 63, 31, 220, 1, 81, 250, 116, 127, 248, 115, 73, 164, 113, 77, 247, 204, 82, 234, 70, 78, 18, 220, 212, 233},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotDigest := Sha256(tt.args.bytes); !reflect.DeepEqual(gotDigest, tt.wantDigest) {
				t.Errorf("Sha256() = %v, want %v", gotDigest, tt.wantDigest)
			}
		})
	}
}
