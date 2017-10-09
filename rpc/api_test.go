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

package rpc

import (
	"testing"

	"github.com/nebulasio/go-nebulas/rpc/pb"
	"github.com/stretchr/testify/assert"
)

func TestGetBalance(t *testing.T) {
	// TODO: mock service.
	s := &APIService{}
	{
		req := &rpcpb.GetBalanceRequest{}
		_, err := s.GetBalance(nil, req)
		assert.Error(t, err, "Missing address.")
	}
	{
		req := &rpcpb.GetBalanceRequest{Address: "0x1"}
		resp, _ := s.GetBalance(nil, req)
		assert.True(t, resp.Value > 0)
	}
}

func TestSendTransaction(t *testing.T) {
	// TODO: mock service.
	s := &APIService{}
	{
		req := &rpcpb.SendTransactionRequest{}
		_, err := s.SendTransaction(nil, req)
		assert.Error(t, err, "Missing sender.")
	}
	{
		req := &rpcpb.SendTransactionRequest{From: "0x1"}
		resp, _ := s.SendTransaction(nil, req)
		assert.True(t, len(resp.Hash) > 0)
	}
}
