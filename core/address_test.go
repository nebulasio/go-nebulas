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

package core

import (
	"reflect"
	"testing"
)

func TestTestKS(t *testing.T) {
	ks := TestKS()
	ks.GetKeyByIndex(0)
}

func TestParse(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    *Address
		wantErr bool
	}{
		{
			"sample address",
			args{"df4d22611412132d3e9bd322f82e2940674ec1bc03b20e40"},
			&Address{[]byte{223, 77, 34, 97, 20, 18, 19, 45, 62, 155, 211, 34, 248, 46, 41, 64, 103, 78, 193, 188, 3, 178, 14, 64}},
			false,
		},
		{
			"case insensitive",
			args{"DF4D22611412132D3E9BD322F82E2940674EC1BC03B20E40"},
			&Address{[]byte{223, 77, 34, 97, 20, 18, 19, 45, 62, 155, 211, 34, 248, 46, 41, 64, 103, 78, 193, 188, 3, 178, 14, 64}},
			false,
		},
		{
			"case insensitive 2",
			args{"DF4d22611412132d3e9bd322f82e2940674ec1bc03b20E40"},
			&Address{[]byte{223, 77, 34, 97, 20, 18, 19, 45, 62, 155, 211, 34, 248, 46, 41, 64, 103, 78, 193, 188, 3, 178, 14, 64}},
			false,
		},
		{
			"insufficient length",
			args{"df4d22611412132d3e9bd322f82e2940674ec1bc"},
			nil,
			true,
		},
		{
			"over length",
			args{"df4d22611412132d3e9bd322f82e2940674ec1bc03b20e4039234"},
			nil,
			true,
		},
		{
			"invalid checksum",
			args{"df4d22611412132d3e9bd322f82e2940674ec1bc03b20e41"},
			nil,
			true,
		},
		{
			"invalid data",
			args{"cf4d22611412132d3e9bd322f82e2940674ec1bc03b20e40"},
			nil,
			true,
		},
		{
			"invalid hex string",
			args{"Zf4d22611412132d3e9bd322f82e2940674ec1bc03b20e40"},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewAddress(t *testing.T) {
	type args struct {
		s []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *Address
		wantErr bool
	}{
		{
			"sample address",
			args{[]byte{223, 77, 34, 97, 20, 18, 19, 45, 62, 155, 211, 34, 248, 46, 41, 64, 103, 78, 193, 188}},
			&Address{[]byte{223, 77, 34, 97, 20, 18, 19, 45, 62, 155, 211, 34, 248, 46, 41, 64, 103, 78, 193, 188, 3, 178, 14, 64}},
			false,
		},
		{
			"insufficient length",
			args{[]byte{223, 77, 34, 97, 20, 18, 19, 45, 62, 155, 211, 34, 248, 46, 41, 64, 103, 78, 193}},
			nil,
			true,
		},
		{
			"over length",
			args{[]byte{223, 77, 34, 97, 20, 18, 19, 45, 62, 155, 211, 34, 248, 46, 41, 64, 103, 78, 193, 188, 12}},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAddress(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}
