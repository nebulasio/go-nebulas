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
	"testing"

	"github.com/nebulasio/go-nebulas/util"
	"github.com/stretchr/testify/assert"
)

func TestLoadBinaryPayload(t *testing.T) {

	tests := []struct {
		name      string
		bytes     []byte
		want      *BinaryPayload
		wantEqual bool
	}{
		{
			name:      "none",
			bytes:     nil,
			want:      NewBinaryPayload(nil),
			wantEqual: true,
		},

		{
			name:      "normal",
			bytes:     []byte("data"),
			want:      NewBinaryPayload([]byte("data")),
			wantEqual: true,
		},

		{
			name:      "not equal",
			bytes:     []byte("data"),
			want:      NewBinaryPayload([]byte("data1")),
			wantEqual: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadBinaryPayload(tt.bytes)
			assert.Nil(t, err)
			if tt.wantEqual {
				assert.Equal(t, tt.want, got)
			} else {
				assert.NotEqual(t, tt.want, got)
			}
		})
	}

}

func TestLoadCallPayload(t *testing.T) {
	want1, _ := NewCallPayload("", "[0]")
	want2, _ := NewCallPayload("func", "")
	want3, _ := NewCallPayload("func", "[0]")
	want4, _ := NewCallPayload("func", "[0]")
	tests := []struct {
		name      string
		bytes     []byte
		parse     bool
		want      *CallPayload
		wantEqual bool
	}{
		{
			name:      "none",
			bytes:     nil,
			parse:     false,
			want:      nil,
			wantEqual: false,
		},

		{
			name:      "parse faild",
			bytes:     []byte("data"),
			parse:     false,
			want:      nil,
			wantEqual: false,
		},

		{
			name:      "no func",
			bytes:     []byte(`{"args": "[0]"}`),
			parse:     true,
			want:      want1,
			wantEqual: true,
		},

		{
			name:      "not args",
			bytes:     []byte(`{"function":"func"}`),
			parse:     true,
			want:      want2,
			wantEqual: true,
		},

		{
			name:      "normal",
			bytes:     []byte(`{"function":"func","args":"[0]"}`),
			parse:     true,
			want:      want3,
			wantEqual: true,
		},

		{
			name:      "not equal",
			bytes:     []byte(`{"function":"func", "args":"[1]"}`),
			parse:     true,
			want:      want4,
			wantEqual: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadCallPayload(tt.bytes)
			if tt.parse {
				assert.Nil(t, err)
				if tt.wantEqual {
					assert.Equal(t, tt.want, got)
				} else {
					assert.NotEqual(t, tt.want, got)
				}
			} else {
				assert.NotNil(t, err)
			}
		})
	}

}

func TestLoadDeployPayload(t *testing.T) {

	deployTx := mockDeployTransaction(0, 0)
	deployPayload, _ := deployTx.LoadPayload()
	deployData, _ := deployPayload.ToBytes()

	tests := []struct {
		name      string
		bytes     []byte
		parse     bool
		want      TxPayload
		wantEqual bool
		limit     *util.Uint128
	}{
		{
			name:      "none",
			bytes:     nil,
			parse:     false,
			want:      nil,
			wantEqual: false,
		},

		{
			name:      "parse faild",
			bytes:     []byte("data"),
			parse:     false,
			want:      nil,
			wantEqual: false,
		},

		{
			name:      "deploy",
			bytes:     deployData,
			parse:     true,
			want:      deployPayload,
			wantEqual: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadDeployPayload(tt.bytes)
			if tt.parse {
				assert.Nil(t, err)
				if tt.wantEqual {
					assert.Equal(t, tt.want, got)
				} else {
					assert.NotEqual(t, tt.want, got)
				}
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestPayload_Execute(t *testing.T) {

	type testPayload struct {
		name     string
		payload  TxPayload
		tx       *Transaction
		block    *Block
		want     *util.Uint128
		wantErr  error
		giveback bool
		limit    *util.Uint128
	}

	neb := testNeb(t)
	bc := neb.chain
	block := bc.tailBlock
	block.Begin()

	tests := []testPayload{
		{
			name:     "normal none",
			payload:  NewBinaryPayload(nil),
			tx:       mockNormalTransaction(bc.chainID, 0),
			block:    block,
			want:     util.NewUint128(),
			wantErr:  nil,
			giveback: false,
		},
		{
			name:     "normal",
			payload:  NewBinaryPayload([]byte("data")),
			tx:       mockNormalTransaction(bc.chainID, 0),
			block:    block,
			want:     util.NewUint128(),
			wantErr:  nil,
			giveback: false,
		},
	}

	deployTx := mockDeployTransaction(bc.chainID, 0)
	deployPayload, _ := deployTx.LoadPayload()
	want, _ := util.NewUint128FromInt(100)
	tests = append(tests, testPayload{
		name:     "deploy",
		payload:  deployPayload,
		tx:       deployTx,
		block:    block,
		want:     want,
		wantErr:  nil,
		giveback: false,
	})

	callTx := mockCallTransaction(bc.chainID, 1, "totalSupply", "")
	callTx.to, _ = deployTx.GenerateContractAddress()
	callPayload, _ := callTx.LoadPayload()
	tests = append(tests, testPayload{
		name:     "call",
		payload:  callPayload,
		tx:       callTx,
		block:    block,
		want:     util.NewUint128(),
		wantErr:  ErrContractCheckFailed,
		giveback: false,
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := tt.payload.Execute(tt.limit, tt.tx, block, block.WorldState())
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
			giveback, err := AcceptTransaction(tt.tx, block.WorldState())
			assert.Nil(t, err)
			assert.Equal(t, giveback, tt.giveback)
		})
	}

	block.RollBack()
}
