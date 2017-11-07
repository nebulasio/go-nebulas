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
	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/rpc/pb"
	"golang.org/x/net/context"
)

// ManagementService implements the RPC management service interface.
type ManagementService struct {
	server Server
}

// NewAccount generate a new address with passphrase
func (s *ManagementService) NewAccount(ctx context.Context, req *rpcpb.NewAccountRequest) (*rpcpb.NewAccountResponse, error) {
	neb := s.server.Neblet()
	addr, err := neb.AccountManager().NewAccount([]byte(req.Passphrase))
	if err != nil {
		return nil, err
	}
	return &rpcpb.NewAccountResponse{Address: addr.ToHex()}, nil
}

// UnlockAccount unlock address with the passphrase
func (s *ManagementService) UnlockAccount(ctx context.Context, req *rpcpb.UnlockAccountRequest) (*rpcpb.UnlockAccountResponse, error) {
	neb := s.server.Neblet()
	addr, err := core.AddressParse(req.Address)
	if err != nil {
		return nil, err
	}
	err = neb.AccountManager().Unlock(addr, []byte(req.Passphrase))
	if err != nil {
		return nil, err
	}
	return &rpcpb.UnlockAccountResponse{Result: true}, nil
}

// LockAccount lock address
func (s *ManagementService) LockAccount(ctx context.Context, req *rpcpb.LockAccountRequest) (*rpcpb.LockAccountResponse, error) {
	neb := s.server.Neblet()
	addr, err := core.AddressParse(req.Address)
	if err != nil {
		return nil, err
	}
	err = neb.AccountManager().Lock(addr)
	if err != nil {
		return nil, err
	}
	return &rpcpb.LockAccountResponse{Result: true}, nil
}

// SignTransaction sign transaction with the from addr passphrase
func (s *ManagementService) SignTransaction(ctx context.Context, req *rpcpb.SignTransactionRequest) (*rpcpb.SignTransactionResponse, error) {
	neb := s.server.Neblet()
	tx, err := parseTransaction(neb, req.From, req.To, req.Value, req.Nonce, nil)
	if err != nil {
		return nil, err
	}
	if err := neb.AccountManager().SignTransaction(tx.From(), tx); err != nil {
		return nil, err
	}
	pbMsg, err := tx.ToProto()
	if err != nil {
		return nil, err
	}
	data, err := proto.Marshal(pbMsg)
	if err != nil {
		return nil, err
	}
	return &rpcpb.SignTransactionResponse{Data: data}, nil
}

// SendTransactionWithPassphrase send transaction with the from addr passphrase
func (s *ManagementService) SendTransactionWithPassphrase(ctx context.Context, req *rpcpb.SendTransactionPassphraseRequest) (*rpcpb.SendTransactionPassphraseResponse, error) {
	neb := s.server.Neblet()
	tx, err := parseTransaction(neb, req.From, req.To, req.Value, req.Nonce, nil)
	if err != nil {
		return nil, err
	}
	if err := neb.AccountManager().SignTransactionWithPassphrase(tx.From(), tx, []byte(req.Passphrase)); err != nil {
		return nil, err
	}
	if err := neb.BlockChain().TransactionPool().PushAndBroadcast(tx); err != nil {
		return nil, err
	}
	return &rpcpb.SendTransactionPassphraseResponse{Hash: tx.Hash().String()}, nil
}
