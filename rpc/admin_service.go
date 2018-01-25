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
	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/rpc/pb"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

// AdminService implements the RPC admin service interface.
type AdminService struct {
	server GRPCServer
}

// NewAccount generate a new address with passphrase
func (s *AdminService) NewAccount(ctx context.Context, req *rpcpb.NewAccountRequest) (*rpcpb.NewAccountResponse, error) {
	logging.VLog().WithFields(logrus.Fields{
		"api": "/v1/admin/account/new",
	}).Info("Rpc request.")
	metricsRPCCounter.Mark(1)

	neb := s.server.Neblet()
	addr, err := neb.AccountManager().NewAccount([]byte(req.Passphrase))
	if err != nil {
		return nil, err
	}
	return &rpcpb.NewAccountResponse{Address: addr.String()}, nil
}

// UnlockAccount unlock address with the passphrase
func (s *AdminService) UnlockAccount(ctx context.Context, req *rpcpb.UnlockAccountRequest) (*rpcpb.UnlockAccountResponse, error) {
	logging.VLog().WithFields(logrus.Fields{
		"api": "/v1/admin/account/unlock",
	}).Info("Rpc request.")
	metricsRPCCounter.Mark(1)

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
	logging.VLog().WithFields(logrus.Fields{
		"api": "/v1/admin/account/lock",
	}).Info("Rpc request.")
	metricsRPCCounter.Mark(1)

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
	logging.VLog().WithFields(logrus.Fields{
		"api": "/v1/admin/sign",
	}).Info("Rpc request.")
	metricsRPCCounter.Mark(1)

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
	logging.VLog().WithFields(logrus.Fields{
		"api": "/v1/admin/transactionWithPassphrase",
	}).Info("Rpc request.")
	metricsRPCCounter.Mark(1)

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
	logging.VLog().WithFields(logrus.Fields{
		"api": "/v1/admin/statistics/nodeInfo",
	}).Info("Rpc request.")
	metricsRPCCounter.Mark(1)

	neb := s.server.Neblet()
	node := neb.NetManager().Node()
	tail := neb.BlockChain().TailBlock()
	resp := &rpcpb.StatisticsNodeInfoResponse{}
	resp.NodeID = node.ID()
	resp.Height = tail.Height()
	resp.Hash = byteutils.Hex(tail.Hash())
	resp.PeerCount = uint32(node.PeersCount())
	return resp, nil
}

// GetDynasty is the RPC API handler.
func (s *AdminService) GetDynasty(ctx context.Context, req *rpcpb.ByBlockHeightRequest) (*rpcpb.GetDynastyResponse, error) {
	logging.VLog().WithFields(logrus.Fields{
		"api":    "/v1/admin/dynasty",
		"height": req.Height,
	}).Info("Rpc request.")
	metricsRPCCounter.Mark(1)

	neb := s.server.Neblet()
	block := neb.BlockChain().GetBlockOnCanonicalChainByHeight(req.Height)
	if block == nil {
		block = neb.BlockChain().TailBlock()
	}
	dynastyRoot := block.DposContext().DynastyRoot
	dynastyTrie, err := trie.NewBatchTrie(dynastyRoot, neb.BlockChain().Storage())
	if err != nil {
		return nil, err
	}
	delegatees, err := core.TraverseDynasty(dynastyTrie)
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, v := range delegatees {
		result = append(result, string(v.Hex()))
	}
	return &rpcpb.GetDynastyResponse{Delegatees: result}, nil
}

// GetCandidates is the RPC API handler.
func (s *AdminService) GetCandidates(ctx context.Context, req *rpcpb.ByBlockHeightRequest) (*rpcpb.GetCandidatesResponse, error) {
	logging.VLog().WithFields(logrus.Fields{
		"api":    "/v1/admin/candidates",
		"height": req.Height,
	}).Info("Rpc request.")
	metricsRPCCounter.Mark(1)

	neb := s.server.Neblet()
	block := neb.BlockChain().GetBlockOnCanonicalChainByHeight(req.Height)
	if block == nil {
		block = neb.BlockChain().TailBlock()
	}
	candidateRoot := block.DposContext().CandidateRoot
	candidateTrie, err := trie.NewBatchTrie(candidateRoot, neb.BlockChain().Storage())
	if err != nil {
		return nil, err
	}
	candidates, err := core.TraverseDynasty(candidateTrie)
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, v := range candidates {
		result = append(result, string(v.Hex()))
	}
	return &rpcpb.GetCandidatesResponse{Candidates: result}, nil
}

// GetDelegateVoters is the RPC API handler.
func (s *AdminService) GetDelegateVoters(ctx context.Context, req *rpcpb.GetDelegateVotersRequest) (*rpcpb.GetDelegateVotersResponse, error) {
	logging.VLog().WithFields(logrus.Fields{
		"delegatee": req.Delegatee,
		"api":       "/v1/admin/delegateVoters",
	}).Info("Rpc request.")
	metricsRPCCounter.Mark(1)

	neb := s.server.Neblet()
	delegatee, err := core.AddressParse(req.Delegatee)
	if err != nil {
		return nil, err
	}
	block := neb.BlockChain().GetBlockOnCanonicalChainByHeight(req.Height)
	if block == nil {
		block = neb.BlockChain().TailBlock()
	}
	delegateRoot := block.DposContext().DelegateRoot
	delegateTrie, _ := trie.NewBatchTrie(delegateRoot, neb.BlockChain().Storage())
	iter, err := delegateTrie.Iterator(delegatee.Bytes())
	if err != nil {
		return nil, err
	}
	voters := []string{}
	exist, err := iter.Next()
	if err != nil {
		return nil, err
	}
	for exist {
		voter := byteutils.Hex(iter.Value())
		voters = append(voters, voter)
		exist, err = iter.Next()
		if err != nil {
			return nil, err
		}
	}
	return &rpcpb.GetDelegateVotersResponse{Voters: voters}, nil
}

// ChangeNetworkID change the network id
func (s *AdminService) ChangeNetworkID(ctx context.Context, req *rpcpb.ChangeNetworkIDRequest) (*rpcpb.ChangeNetworkIDResponse, error) {
	logging.VLog().WithFields(logrus.Fields{
		"api": "/v1/admin/changeNetworkID",
	}).Info("Rpc request.")
	metricsRPCCounter.Mark(1)

	neb := s.server.Neblet()
	neb.NetManager().Node().Config().NetworkID = req.NetworkId
	// broadcast to all the node in the routetable.
	neb.NetManager().BroadcastNetworkID(byteutils.FromUint32(req.NetworkId))
	return &rpcpb.ChangeNetworkIDResponse{Result: true}, nil
}

// StartMining start mining
func (s *AdminService) StartMining(ctx context.Context, req *rpcpb.StartMiningRequest) (*rpcpb.MiningResponse, error) {
	logging.VLog().WithFields(logrus.Fields{
		"api": "/v1/admin/startMining",
	}).Info("Rpc request.")
	metricsRPCCounter.Mark(1)

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
	logging.VLog().WithFields(logrus.Fields{
		"api": "/v1/admin/stopMining",
	}).Info("Rpc request.")
	metricsRPCCounter.Mark(1)

	neb := s.server.Neblet()

	if !neb.Consensus().Enable() {
		return nil, errors.New("consensus not start yet")
	}

	if err := neb.Consensus().DisableMining(); err != nil {
		return nil, err
	}
	return &rpcpb.MiningResponse{Result: true}, nil
}
