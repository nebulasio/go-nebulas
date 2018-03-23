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
	"github.com/nebulasio/go-nebulas/rpc/pb"
	"golang.org/x/net/context"
)

// AdminService implements the RPC admin service interface.
type AdminService struct {
	server GRPCServer
}

// Accounts is the RPC API handler.
func (s *AdminService) Accounts(ctx context.Context, req *rpcpb.NonParamsRequest) (*rpcpb.AccountsResponse, error) {

	neb := s.server.Neblet()
	accs := neb.AccountManager().Accounts()

	resp := new(rpcpb.AccountsResponse)
	addrs := make([]string, len(accs))
	for index, addr := range accs {
		addrs[index] = addr.String()
	}
	resp.Addresses = addrs
	return resp, nil
}

// NewAccount generate a new address with passphrase
func (s *AdminService) NewAccount(ctx context.Context, req *rpcpb.NewAccountRequest) (*rpcpb.NewAccountResponse, error) {

	neb := s.server.Neblet()
	addr, err := neb.AccountManager().NewAccount([]byte(req.Passphrase))
	if err != nil {
		return nil, err
	}
	return &rpcpb.NewAccountResponse{Address: addr.String()}, nil
}

// signTransactionWithPassphrase sign transaction with the from addr passphrase
func (s *AdminService) SignTransactionWithPassphrase(ctx context.Context, req *rpcpb.SignTransactionPassphraseRequest) (*rpcpb.SignTransactionPassphraseResponse, error) {

	neb := s.server.Neblet()
	tx, err := parseTransaction(neb, req.Transaction)
	if err != nil {
		metricsSignTxFailed.Mark(1)
		return nil, err
	}
	if err := neb.AccountManager().SignTransactionWithPassphrase(tx.From(), tx, []byte(req.Passphrase)); err != nil {
		metricsSignTxFailed.Mark(1)
		return nil, err
	}
	pbMsg, err := tx.ToProto()
	if err != nil {
		metricsSignTxFailed.Mark(1)
		return nil, err
	}
	data, err := proto.Marshal(pbMsg)
	if err != nil {
		metricsSignTxFailed.Mark(1)
		return nil, err
	}

	metricsSignTxSuccess.Mark(1)
	return &rpcpb.SignTransactionPassphraseResponse{Data: data}, nil
}

// SendTransactionWithPassphrase send transaction with the from addr passphrase
func (s *AdminService) SendTransactionWithPassphrase(ctx context.Context, req *rpcpb.SendTransactionPassphraseRequest) (*rpcpb.SendTransactionResponse, error) {

	neb := s.server.Neblet()
	tx, err := parseTransaction(neb, req.Transaction)
	if err != nil {
		return nil, err
	}
	if err := neb.AccountManager().SignTransactionWithPassphrase(tx.From(), tx, []byte(req.Passphrase)); err != nil {
		return nil, err
	}

	return handleTransactionResponse(neb, tx)
}

// StartPprof start pprof
func (s *AdminService) StartPprof(ctx context.Context, req *rpcpb.PprofRequest) (*rpcpb.PprofResponse, error) {

	neb := s.server.Neblet()

	if err := neb.StartPprof(req.Listen); err != nil {
		return nil, err
	}
	return &rpcpb.PprofResponse{Result: true}, nil
}

// GetConfig is the RPC API handler.
func (s *AdminService) GetConfig(ctx context.Context, req *rpcpb.NonParamsRequest) (*rpcpb.GetConfigResponse, error) {

	neb := s.server.Neblet()

	resp := &rpcpb.GetConfigResponse{}
	resp.Config = neb.Config()
	resp.Config.Chain.Passphrase = string("") //TODO remove passphrase to config
	return resp, nil
}
