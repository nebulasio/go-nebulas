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
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/rpc/pb"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"golang.org/x/net/context"
)

// AdminService implements the RPC admin service interface.
type AdminService struct {
	server GRPCServer
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

// UnlockAccount unlock address with the passphrase
func (s *AdminService) UnlockAccount(ctx context.Context, req *rpcpb.UnlockAccountRequest) (*rpcpb.UnlockAccountResponse, error) {

	neb := s.server.Neblet()
	addr, err := core.AddressParse(req.Address)
	if err != nil {
		metricsUnlockFailed.Mark(1)
		return nil, err
	}
	duration := time.Duration(req.Duration)
	if duration == 0 {
		duration = keystore.DefaultUnlockDuration
	}
	err = neb.AccountManager().Unlock(addr, []byte(req.Passphrase), duration)
	if err != nil {
		metricsUnlockFailed.Mark(1)
		return nil, err
	}

	metricsUnlockSuccess.Mark(1)
	return &rpcpb.UnlockAccountResponse{Result: true}, nil
}

// LockAccount lock address
func (s *AdminService) LockAccount(ctx context.Context, req *rpcpb.LockAccountRequest) (*rpcpb.LockAccountResponse, error) {

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
func (s *AdminService) SignTransaction(ctx context.Context, req *rpcpb.TransactionRequest) (*rpcpb.SignTransactionResponse, error) {

	neb := s.server.Neblet()
	tx, err := parseTransaction(neb, req)
	if err != nil {
		metricsSignTxFailed.Mark(1)
		return nil, err
	}
	if err := neb.AccountManager().SignTransaction(tx.From(), tx); err != nil {
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
	return &rpcpb.SignTransactionResponse{Data: data}, nil
}

// SendTransactionWithPassphrase send transaction with the from addr passphrase
func (s *AdminService) SendTransactionWithPassphrase(ctx context.Context, req *rpcpb.SendTransactionPassphraseRequest) (*rpcpb.SendTransactionPassphraseResponse, error) {

	neb := s.server.Neblet()
	tx, err := parseTransaction(neb, req.Transaction)
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

// StatisticsNodeInfo is the RPC API handler.
func (s *AdminService) StatisticsNodeInfo(ctx context.Context, req *rpcpb.NonParamsRequest) (*rpcpb.StatisticsNodeInfoResponse, error) {

	neb := s.server.Neblet()
	node := neb.NetService().Node()
	tail := neb.BlockChain().TailBlock()
	resp := &rpcpb.StatisticsNodeInfoResponse{}
	resp.NodeID = node.ID()
	resp.Height = tail.Height()
	resp.Hash = byteutils.Hex(tail.Hash())
	resp.PeerCount = uint32(node.PeersCount())
	return resp, nil
}

// ChangeNetworkID change the network id
func (s *AdminService) ChangeNetworkID(ctx context.Context, req *rpcpb.ChangeNetworkIDRequest) (*rpcpb.ChangeNetworkIDResponse, error) {

	neb := s.server.Neblet()
	neb.NetService().Node().Config().NetworkID = req.NetworkId
	// broadcast to all the node in the routetable.
	neb.NetService().BroadcastNetworkID(byteutils.FromUint32(req.NetworkId))
	return &rpcpb.ChangeNetworkIDResponse{Result: true}, nil
}

// StartMining start mining
func (s *AdminService) StartMining(ctx context.Context, req *rpcpb.StartMiningRequest) (*rpcpb.MiningResponse, error) {

	neb := s.server.Neblet()

	if neb.Consensus().Enable() {
		return nil, errors.New("consensus has already been started")
	}

	err := neb.Consensus().EnableMining(req.Passphrase)
	if err != nil {
		return nil, err
	}
	return &rpcpb.MiningResponse{Result: true}, nil
}

// StopMining stop mining
func (s *AdminService) StopMining(ctx context.Context, req *rpcpb.NonParamsRequest) (*rpcpb.MiningResponse, error) {

	neb := s.server.Neblet()

	if !neb.Consensus().Enable() {
		return nil, errors.New("consensus not start yet")
	}

	if err := neb.Consensus().DisableMining(); err != nil {
		return nil, err
	}
	return &rpcpb.MiningResponse{Result: true}, nil
}

// StartPprof start pprof
func (s *AdminService) StartPprof(ctx context.Context, req *rpcpb.PprofRequest) (*rpcpb.PprofResponse, error) {

	neb := s.server.Neblet()

	if err := neb.StartPprof(req.Listen); err != nil {
		return nil, err
	}
	return &rpcpb.PprofResponse{Result: true}, nil
}
