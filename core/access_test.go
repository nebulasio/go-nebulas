// Copyright (C) 2018 go-nebulas authors
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
	"strings"
	"testing"

	corepb "github.com/nebulasio/go-nebulas/core/pb"
	"github.com/stretchr/testify/assert"
)

func TestNewAccess(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "no path",
			path:    "",
			wantErr: false,
		},
		{
			name:    "invalid path",
			path:    "invalid path",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			access, err := NewAccess(testNeb(t))
			assert.Equal(t, tt.wantErr, err != nil)
			if err == nil {
				assert.NotNil(t, access)
			}
		})
	}
}

func TestAccess_CheckTransaction(t *testing.T) {
	access := &Access{
		access: &corepb.Access{
			Blacklist: &corepb.Blacklist{
				From: []string{
					"n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE",
					"n1GmkKH6nBMw4rrjt16RrJ9WcgvKUtAZP1s",
					"n1H4MYms9F55ehcvygwWE71J8tJC4CRr2so",
				},
				To: []string{
					"n1JAy4X6KKLCNiTd7MWMRsVBjgdVq5WCCpf",
					"n1LkDi2gGMqPrjYcczUiweyP4RxTB6Go1qS",
					"n1LmP9K8pFF33fgdgHZonFEMsqZinJ4EUqk",
				},
				Contracts: []*corepb.Contract{
					{
						Address: "n1UM7z6MqnGyKEPvUpwrfxZpM1eB7UpzmLJ",
						Functions: []string{
							"transferEvent",
							"approveEvent",
						},
						Keywords: []string{
							"random",
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name     string
		from     string
		to       string
		function string
		contract string
		err      error
	}{
		{
			name: "normal",
			from: mockAddress().String(),
			to:   mockAddress().String(),
			err:  nil,
		},
		{
			name: "from in blacklist",
			from: "n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE",
			to:   mockAddress().String(),
			err:  ErrRestrictedFromAddress,
		},
		{
			name: "to in blacklist",
			from: mockAddress().String(),
			to:   "n1JAy4X6KKLCNiTd7MWMRsVBjgdVq5WCCpf",
			err:  ErrRestrictedToAddress,
		},
		{
			name:     "func in blacklist but address not in",
			from:     mockAddress().String(),
			to:       mockAddress().String(),
			function: "transferEvent",
			err:      nil,
		},
		{
			name:     "func and from address in blacklist",
			from:     "n1UM7z6MqnGyKEPvUpwrfxZpM1eB7UpzmLJ",
			to:       mockAddress().String(),
			function: "transferEvent",
			err:      nil,
		},
		{
			name:     "func and to address in blacklist",
			from:     mockAddress().String(),
			to:       "n1UM7z6MqnGyKEPvUpwrfxZpM1eB7UpzmLJ",
			function: "transferEvent",
			err:      ErrUnsupportedFunction,
		},
		{
			name:     "address in blacklist but func not in",
			from:     mockAddress().String(),
			to:       "n1UM7z6MqnGyKEPvUpwrfxZpM1eB7UpzmLJ",
			function: "transferEventTest",
			err:      nil,
		},
		{
			name:     "keyword in blacklist",
			from:     mockAddress().String(),
			to:       mockAddress().String(),
			contract: "random",
			err:      ErrUnsupportedKeyword,
		},
		{
			name:     "keyword not in blacklist",
			from:     mockAddress().String(),
			to:       mockAddress().String(),
			contract: "ran",
			err:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payloadType := TxPayloadBinaryType
			payloadBytes := []byte{}
			if tt.function != "" {
				payloadType = TxPayloadCallType
				payload, _ := NewCallPayload(tt.function, "")
				payloadBytes, _ = payload.ToBytes()
			}
			if tt.contract != "" {
				payloadType = TxPayloadDeployType
				payload, _ := NewDeployPayload(tt.contract, "js", "")
				payloadBytes, _ = payload.ToBytes()
			}
			tx := mockTransaction(100, 1, payloadType, payloadBytes)
			tx.from, _ = AddressParse(tt.from)
			tx.to, _ = AddressParse(tt.to)

			err := access.CheckTransaction(tx)
			if tt.err != ErrUnsupportedKeyword {
				assert.Equal(t, tt.err, err, tt.name)
			} else {
				assert.True(t, strings.HasPrefix(err.Error(), tt.err.Error()))
			}
		})
	}

}

func TestAccess_CheckNRC20Transaction(t *testing.T) {
	nrc20Contract := "n1qGqcvPWE45VGYZYNPd9VgA1KAAV1sRqfo"
	access := &Access{
		access: &corepb.Access{
			Nrc20List: &corepb.Nrc20List{
				Contracts: []string{nrc20Contract},
			},
		},
	}

	tests := []struct {
		name     string
		to       string
		function string
		args     string
		err      error
	}{
		{
			name: "normal not nrc20",
			to:   mockAddress().String(),
			err:  nil,
		},
		{
			name:     "nrc20 func not check",
			to:       nrc20Contract,
			function: "balanceOf",
			err:      nil,
		},
		{
			name:     "transfer args err",
			to:       nrc20Contract,
			function: NRC20FuncTransfer,
			args:     "[\"n1qGqcvPWE45VGYZYNPd9VgA1KAAV1sRqfo\",0]",
			err:      ErrNrc20ArgsCheckFailed,
		},
		{
			name:     "transfer addr err",
			to:       nrc20Contract,
			function: NRC20FuncTransfer,
			args:     "[\"1\",\"0\"]",
			err:      ErrNrc20AddressCheckFailed,
		},
		{
			name:     "transfer value err",
			to:       nrc20Contract,
			function: NRC20FuncTransfer,
			args:     "[\"n1qGqcvPWE45VGYZYNPd9VgA1KAAV1sRqfo\",\"0.1\"]",
			err:      ErrNrc20ValueCheckFailed,
		},
		{
			name:     "transfer value err",
			to:       nrc20Contract,
			function: NRC20FuncTransfer,
			args:     "[\"n1qGqcvPWE45VGYZYNPd9VgA1KAAV1sRqfo\",\"-1\"]",
			err:      ErrNrc20ValueCheckFailed,
		},
		{
			name:     "transfer value err",
			to:       nrc20Contract,
			function: NRC20FuncTransfer,
			args:     "[\"n1qGqcvPWE45VGYZYNPd9VgA1KAAV1sRqfo\",\"1e2\"]",
			err:      ErrNrc20ValueCheckFailed,
		},
		{
			name:     "transfer success",
			to:       nrc20Contract,
			function: NRC20FuncTransfer,
			args:     "[\"n1qGqcvPWE45VGYZYNPd9VgA1KAAV1sRqfo\",\"1\"]",
			err:      nil,
		},
		{
			name:     "transferFrom addr err",
			to:       nrc20Contract,
			function: NRC20FuncTransferFrom,
			args:     "[\"1\",\"0\",\"0\"]",
			err:      ErrNrc20AddressCheckFailed,
		},
		{
			name:     "transferFrom value err",
			to:       nrc20Contract,
			function: NRC20FuncTransferFrom,
			args:     "[\"n1qGqcvPWE45VGYZYNPd9VgA1KAAV1sRqfo\",\"0\",\"0.1\"]",
			err:      ErrNrc20ValueCheckFailed,
		},
		{
			name:     "transferFrom value err",
			to:       nrc20Contract,
			function: NRC20FuncTransferFrom,
			args:     "[\"n1qGqcvPWE45VGYZYNPd9VgA1KAAV1sRqfo\",\"0\",\"-1\"]",
			err:      ErrNrc20ValueCheckFailed,
		},
		{
			name:     "transferFrom value err",
			to:       nrc20Contract,
			function: NRC20FuncTransferFrom,
			args:     "[\"n1qGqcvPWE45VGYZYNPd9VgA1KAAV1sRqfo\",\"0\",\"1e2\"]",
			err:      ErrNrc20ValueCheckFailed,
		},
		{
			name:     "transferFrom success",
			to:       nrc20Contract,
			function: NRC20FuncTransferFrom,
			args:     "[\"n1qGqcvPWE45VGYZYNPd9VgA1KAAV1sRqfo\",\"0\",\"0\"]",
			err:      nil,
		},
		{
			name:     "approve value err",
			to:       nrc20Contract,
			function: NRC20FuncApprove,
			args:     "[\"n1qGqcvPWE45VGYZYNPd9VgA1KAAV1sRqfo\",\"0\",\"1e2\"]",
			err:      ErrNrc20ValueCheckFailed,
		},
		{
			name:     "approve",
			to:       nrc20Contract,
			function: NRC20FuncApprove,
			args:     "[\"n1qGqcvPWE45VGYZYNPd9VgA1KAAV1sRqfo\",\"0\",\"0\"]",
			err:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payloadType := TxPayloadBinaryType
			payloadBytes := []byte{}
			if tt.function != "" {
				payloadType = TxPayloadCallType
				payload, _ := NewCallPayload(tt.function, tt.args)
				payloadBytes, _ = payload.ToBytes()
			}
			tx := mockTransaction(100, 1, payloadType, payloadBytes)
			tx.to, _ = AddressParse(tt.to)

			err := access.CheckTransaction(tx)
			assert.Equal(t, tt.err, err, tt.name)
		})
	}

}
