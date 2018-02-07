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

	"github.com/nebulasio/go-nebulas/crypto"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
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
			want:      NewCallPayload("", "[0]"),
			wantEqual: true,
		},

		{
			name:      "not args",
			bytes:     []byte(`{"function":"func"}`),
			parse:     true,
			want:      NewCallPayload("func", ""),
			wantEqual: true,
		},

		{
			name:      "normal",
			bytes:     []byte(`{"function":"func","args":"[0]"}`),
			parse:     true,
			want:      NewCallPayload("func", "[0]"),
			wantEqual: true,
		},

		{
			name:      "not equal",
			bytes:     []byte(`{"function":"func", "args":"[1]"}`),
			parse:     true,
			want:      NewCallPayload("func", "[0]"),
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

func TestLoadCandidatePayload(t *testing.T) {
	tests := []struct {
		name      string
		bytes     []byte
		parse     bool
		want      *CandidatePayload
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
			name:      LoginAction,
			bytes:     []byte(`{"action": "login"}`),
			parse:     true,
			want:      NewCandidatePayload(LoginAction),
			wantEqual: true,
		},
		{
			name:      LogoutAction,
			bytes:     []byte(`{"action": "logout"}`),
			parse:     true,
			want:      NewCandidatePayload(LogoutAction),
			wantEqual: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadCandidatePayload(tt.bytes)
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

func TestLoadDelegatePayload(t *testing.T) {
	tests := []struct {
		name      string
		bytes     []byte
		parse     bool
		want      *DelegatePayload
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
			name:      DelegateAction,
			bytes:     []byte(`{"action": "do", "delegatee": "1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c"}`),
			parse:     true,
			want:      NewDelegatePayload(DelegateAction, "1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c"),
			wantEqual: true,
		},
		{
			name:      UnDelegateAction,
			bytes:     []byte(`{"action": "undo", "delegatee": "1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c"}`),
			parse:     true,
			want:      NewDelegatePayload(UnDelegateAction, "1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c"),
			wantEqual: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadDelegatePayload(tt.bytes)
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
	deployPayload, _ := deployTx.LoadPayload(nil)
	deployData, _ := deployPayload.ToBytes()

	tests := []struct {
		name      string
		bytes     []byte
		parse     bool
		want      TxPayload
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
			name:      LoginAction,
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
		name    string
		payload TxPayload
		tx      *Transaction
		block   *Block
		want    *util.Uint128
		wantErr error
	}

	neb := testNeb()
	bc, _ := NewBlockChain(neb)
	block := bc.tailBlock
	block.begin()

	tests := []testPayload{
		{
			name:    "normal none",
			payload: NewBinaryPayload(nil),
			tx:      mockNormalTransaction(bc.chainID, 0),
			block:   block,
			want:    util.NewUint128(),
			wantErr: nil,
		},
		{
			name:    "normal",
			payload: NewBinaryPayload([]byte("data")),
			tx:      mockNormalTransaction(bc.chainID, 0),
			block:   block,
			want:    util.NewUint128(),
			wantErr: nil,
		},
	}

	deployTx := mockDeployTransaction(bc.chainID, 0)
	deployPayload, _ := deployTx.LoadPayload(nil)
	tests = append(tests, testPayload{
		name:    "deploy",
		payload: deployPayload,
		tx:      deployTx,
		block:   block,
		want:    util.NewUint128FromInt(189),
		wantErr: nil,
	})

	callTx := mockCallTransaction(bc.chainID, 1, "totalSupply", "")
	callTx.to, _ = deployTx.GenerateContractAddress()
	callPayload, _ := callTx.LoadPayload(nil)
	tests = append(tests, testPayload{
		name:    "call",
		payload: callPayload,
		tx:      callTx,
		block:   block,
		want:    util.NewUint128(),
		wantErr: ErrContractDeployFailed,
	})

	delegateTx := mockDelegateTransaction(bc.chainID, 0, DelegateAction, mockAddress().String())
	delegatePayload, _ := delegateTx.LoadPayload(nil)
	tests = append(tests, testPayload{
		name:    "delegate no candidate",
		payload: delegatePayload,
		tx:      delegateTx,
		block:   block,
		want:    ZeroGasCount,
		wantErr: ErrInvalidDelegateToNonCandidate,
	})

	candidateInTx := mockCandidateTransaction(bc.chainID, 0, LoginAction)
	candidateInPayload, _ := candidateInTx.LoadPayload(nil)
	tests = append(tests, testPayload{
		name:    "candidate login",
		payload: candidateInPayload,
		tx:      candidateInTx,
		block:   block,
		want:    util.NewUint128(),
		wantErr: nil,
	})

	delegateCandidateTx := mockDelegateTransaction(bc.chainID, 0, DelegateAction, candidateInTx.from.String())
	delegateCandidatePayload, _ := delegateCandidateTx.LoadPayload(nil)
	tests = append(tests, testPayload{
		name:    "delegate candidate",
		payload: delegateCandidatePayload,
		tx:      delegateCandidateTx,
		block:   block,
		want:    ZeroGasCount,
		wantErr: nil,
	})

	candidateOutTx := mockCandidateTransaction(bc.chainID, 0, LogoutAction)
	candidateOutTx.from = candidateInTx.from
	candidateOutPayload, _ := candidateOutTx.LoadPayload(nil)
	tests = append(tests, testPayload{
		name:    "candidate logout",
		payload: candidateOutPayload,
		tx:      candidateOutTx,
		block:   block,
		want:    util.NewUint128(),
		wantErr: nil,
	})

	ks := keystore.DefaultKS
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, _ := ks.GetUnlocked(tt.tx.from.String())
			signature, _ := crypto.NewSignature(keystore.SECP256K1)
			signature.InitSign(key.(keystore.PrivateKey))

			err := tt.tx.Sign(signature)
			assert.Nil(t, err)

			block.acceptTransaction(tt.tx)

			txblock, _ := block.Clone()

			got, _, err := tt.payload.Execute(txblock, tt.tx)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)

			if err != nil {
				txblock.rollback()
			} else {
				block.Merge(txblock)
			}
		})
	}

	block.rollback()
}
