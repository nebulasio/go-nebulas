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
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/rpc/pb"
	"github.com/nebulasio/go-nebulas/util"
	"golang.org/x/net/context"
)

// APIService implements the RPC API service interface.
type APIService struct {
	server *Server
}

// GetNebState is the RPC API handler.
func (s *APIService) GetNebState(ctx context.Context, req *rpcpb.GetNebStateRequest) (*rpcpb.GetNebStateResponse, error) {
	neb := s.server.Neblet()

	tailHash := neb.BlockChain().TailBlock().Hash()

	resp := &rpcpb.GetNebStateResponse{Tail: string(tailHash.Hex())}

	return resp, nil
}

// GetAccountState is the RPC API handler.
func (s *APIService) GetAccountState(ctx context.Context, req *rpcpb.GetAccountStateRequest) (*rpcpb.GetAccountStateResponse, error) {
	neb := s.server.Neblet()

	addr, err := core.AddressParse(req.Address)
	if err != nil {
		return nil, err
	}

	// TODO: handle specific block number.
	acct := neb.BlockChain().TailBlock().FindAccount(addr)

	fsb, err := acct.UserBalance.ToFixedSizeByteSlice()
	if err != nil {
		return nil, err
	}
	return &rpcpb.GetAccountStateResponse{Balance: fsb, Nonce: acct.UserNonce}, nil
}

// SendTransaction is the RPC API handler.
func (s *APIService) SendTransaction(ctx context.Context, req *rpcpb.SendTransactionRequest) (*rpcpb.SendTransactionResponse, error) {
	// Validate and sign the tx, then submit it to the tx pool.
	neb := s.server.Neblet()

	fromAddr, err := core.AddressParse(req.From)
	if err != nil {
		return nil, err
	}
	toAddr, err := core.AddressParse(req.To)
	if err != nil {
		return nil, err
	}

	value, err := util.NewUint128FromFixedSizeByteSlice(req.Value)
	if err != nil {
		return nil, err
	}

	tx := core.NewTransaction(neb.BlockChain().ChainID(), fromAddr, toAddr, value, req.Nonce /*req.Data */, nil)
	if err := neb.AccountManager().SignTransaction(fromAddr, tx); err != nil {
		return nil, err
	}

	if err := neb.BlockChain().TransactionPool().PushAndBroadcast(tx); err != nil {
		return nil, err
	}

	return &rpcpb.SendTransactionResponse{Hash: tx.Hash().String()}, nil
}
