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
	"errors"
	"io/ioutil"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/core"
	corepb "github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/rpc/pb"
	"github.com/nebulasio/go-nebulas/util"
	"golang.org/x/net/context"
)

// APIService implements the RPC API service interface.
type APIService struct {
	server Server
}

// GetNebState is the RPC API handler.
func (s *APIService) GetNebState(ctx context.Context, req *rpcpb.GetNebStateRequest) (*rpcpb.GetNebStateResponse, error) {
	neb := s.server.Neblet()

	tail := neb.BlockChain().TailBlock()

	resp := &rpcpb.GetNebStateResponse{}
	resp.ChainID = neb.BlockChain().ChainID()
	resp.Tail = string(tail.Hash().Hex())
	resp.Coinbase = tail.Coinbase().ToHex()

	return resp, nil
}

// Accounts is the RPC API handler.
func (s *APIService) Accounts(ctx context.Context, req *rpcpb.AccountsRequest) (*rpcpb.AccountsResponse, error) {
	neb := s.server.Neblet()

	accs := neb.AccountManager().Accounts()

	resp := new(rpcpb.AccountsResponse)
	addrs := make([]string, len(accs))
	for index, addr := range accs {
		addrs[index] = addr.ToHex()
	}
	resp.Addresses = addrs
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

	source, err := ioutil.ReadFile("/Users/leon/go/src/github.com/nebulasio/go-nebulas/nf/nvm/test/sample_contract.js")
	if err != nil {

	}
	args := "[\"TEST001\", 123,[{\"name\":\"robin\",\"count\":2},{\"name\":\"roy\",\"count\":3},{\"name\":\"leon\",\"count\":4}]]"

	data, err := core.NewDeploySCPayload(string(source), args)
	if err != nil {
		return nil, err
	}

	tx, err := parseTransaction(neb, req.From, req.To, req.Value, req.Nonce, data)
	if err != nil {
		return nil, err
	}
	if err := neb.AccountManager().SignTransaction(tx.From(), tx); err != nil {
		return nil, err
	}
	if err := neb.BlockChain().TransactionPool().PushAndBroadcast(tx); err != nil {
		return nil, err
	}

	return &rpcpb.SendTransactionResponse{Hash: tx.Hash().String()}, nil
}

// Call a contract
func (s *APIService) Call(ctx context.Context, req *rpcpb.SendTransactionRequest) (*rpcpb.SendTransactionResponse, error) {
	// Validate and sign the tx, then submit it to the tx pool.
	neb := s.server.Neblet()

	// data, err := core.NewDeploySCPayload(string(soruce), args)
	data, err := core.NewCallSCPayload("dump", "")
	if err != nil {
		return nil, err
	}

	tx, err := parseTransaction(neb, req.From, req.To, req.Value, req.Nonce, data)
	if err != nil {
		return nil, err
	}
	if err := neb.AccountManager().SignTransaction(tx.From(), tx); err != nil {
		return nil, err
	}

	if err := neb.BlockChain().TransactionPool().PushAndBroadcast(tx); err != nil {
		return nil, err
	}

	return &rpcpb.SendTransactionResponse{Hash: tx.Hash().String()}, nil
}

func parseTransaction(neb Neblet, from, to string, v []byte, nonce uint64, data []byte) (*core.Transaction, error) {
	fromAddr, err := core.AddressParse(from)
	if err != nil {
		return nil, err
	}
	var toAddr *core.Address
	if len(to) > 0 {
		toAddr, err = core.AddressParse(to)
		if err != nil {
			return nil, err
		}
	}

	var value *util.Uint128
	if len(v) > 0 {
		value, err = util.NewUint128FromFixedSizeByteSlice(v)
		if err != nil {
			return nil, err
		}
	} else {
		value = util.NewUint128()
	}

	tx := core.NewTransaction(neb.BlockChain().ChainID(), fromAddr, toAddr, value, nonce, data)
	return tx, nil
}

// SendRawTransaction submit the signed transaction raw data to txpool
func (s *APIService) SendRawTransaction(ctx context.Context, req *rpcpb.SendRawTransactionRequest) (*rpcpb.SendTransactionResponse, error) {
	// Validate and sign the tx, then submit it to the tx pool.
	neb := s.server.Neblet()

	pbTx := new(corepb.Transaction)
	if err := proto.Unmarshal(req.GetData(), pbTx); err != nil {
		return nil, err
	}
	tx := new(core.Transaction)
	if err := tx.FromProto(pbTx); err != nil {
		return nil, err
	}

	if err := neb.BlockChain().TransactionPool().PushAndBroadcast(tx); err != nil {
		return nil, err
	}

	return &rpcpb.SendTransactionResponse{Hash: tx.Hash().String()}, nil
}

// GetBlockByHash get block info by the block hash
func (s *APIService) GetBlockByHash(ctx context.Context, req *rpcpb.GetBlockByHashRequest) (*corepb.Block, error) {
	neb := s.server.Neblet()

	block := neb.BlockChain().GetBlock([]byte(req.GetHash()))
	if block == nil {
		return nil, errors.New("block not found")
	}
	pbBlock, err := block.ToProto()
	if err != nil {
		return nil, err
	}
	return pbBlock.(*corepb.Block), nil
}

// GetTransactionByHash get transaction info by the transaction hash
func (s *APIService) GetTransactionByHash(ctx context.Context, req *rpcpb.GetTransactionByHashRequest) (*corepb.Transaction, error) {
	neb := s.server.Neblet()

	tx := neb.BlockChain().GetTransaction([]byte(req.GetHash()))
	if tx == nil {
		return nil, errors.New("transaction not found")
	}
	pbTx, err := tx.ToProto()
	if err != nil {
		return nil, err
	}
	return pbTx.(*corepb.Transaction), nil
}

// NewDeploySCPayload new deploySCPayload
func (s *APIService) NewDeploySCPayload(ctx context.Context, req *rpcpb.NewDeploySCPayloadRequest) (*rpcpb.NewDeploySCPayloadResponse, error) {
	data, err := core.NewDeploySCPayload(req.Source, req.Args)
	if err != nil {
		return nil, err
	}
	return &rpcpb.NewDeploySCPayloadResponse{Data: data}, nil
}

// NewCallSCPayload new callSCPayload
func (s *APIService) NewCallSCPayload(ctx context.Context, req *rpcpb.NewCallSCPayloadRequest) (*rpcpb.NewCallSCPayloadResponse, error) {
	data, err := core.NewCallSCPayload(req.Function, req.Args)
	if err != nil {
		return nil, err
	}
	return &rpcpb.NewCallSCPayloadResponse{Data: data}, nil
}
