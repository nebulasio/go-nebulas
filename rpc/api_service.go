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
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/nebulasio/go-nebulas/common/trie"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/core"
	corepb "github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/crypto/hash"
	nnet "github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/net/p2p"
	"github.com/nebulasio/go-nebulas/rpc/pb"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

// APIService implements the RPC API service interface.
type APIService struct {
	server Server
}

// GetNebState is the RPC API handler.
func (s *APIService) GetNebState(ctx context.Context, req *rpcpb.NonParamsRequest) (*rpcpb.GetNebStateResponse, error) {
	logging.VLog().WithFields(logrus.Fields{
		"api": "/v1/user/nebstate",
	}).Info("Rpc request.")

	neb := s.server.Neblet()

	tail := neb.BlockChain().TailBlock()

	resp := &rpcpb.GetNebStateResponse{}
	resp.ChainId = neb.BlockChain().ChainID()
	resp.Tail = tail.Hash().String()
	resp.Coinbase = tail.Coinbase().String()
	resp.Synchronized = neb.NetManager().Node().GetSynchronizing()
	resp.PeerCount = getStreamCount(neb.NetManager().Node().GetStream())
	resp.ProtocolVersion = p2p.ProtocolID

	return resp, nil
}

// NodeInfo is the PRC API handler
func (s *APIService) NodeInfo(ctx context.Context, req *rpcpb.NonParamsRequest) (*rpcpb.NodeInfoResponse, error) {
	logging.VLog().WithFields(logrus.Fields{
		"api": "/v1/user/nodeinfo",
	}).Info("Rpc request.")

	neb := s.server.Neblet()
	resp := &rpcpb.NodeInfoResponse{}
	node := neb.NetManager().Node()
	resp.Id = node.ID()
	resp.ChainId = node.Config().ChainID
	resp.BucketSize = int32(node.Config().Bucketsize)
	resp.Version = uint32(node.Config().Version)
	resp.StreamStoreSize = int32(node.Config().StreamStoreSize)
	resp.StreamStoreExtendSize = int32(node.Config().StreamStoreExtendSize)
	resp.RelayCacheSize = int32(node.Config().RelayCacheSize)
	resp.PeerCount = getStreamCount(node.GetStream())
	resp.ProtocolVersion = p2p.ProtocolID
	for _, v := range node.PeerStore().Peers() {
		routeTable := &rpcpb.RouteTable{}
		routeTable.Id = v.Pretty()
		if len(node.PeerStore().Addrs(v)) > 0 {
			var addrs []string
			for _, val := range node.PeerStore().Addrs(v) {
				addrs = append(addrs, val.String())
			}
			routeTable.Address = addrs
			resp.RouteTable = append(resp.RouteTable, routeTable)
		}
	}
	return resp, nil
}

// StatisticsNodeInfo is the RPC API handler.
func (s *APIService) StatisticsNodeInfo(ctx context.Context, req *rpcpb.NonParamsRequest) (*rpcpb.StatisticsNodeInfoResponse, error) {
	logging.VLog().WithFields(logrus.Fields{
		"api": "/v1/admin/statistics/nodeInfo",
	}).Info("Rpc request.")

	neb := s.server.Neblet()
	node := neb.NetManager().Node()
	tail := neb.BlockChain().TailBlock()
	resp := &rpcpb.StatisticsNodeInfoResponse{}
	resp.NodeID = node.ID()
	resp.Height = tail.Height()
	resp.Hash = byteutils.Hex(tail.Hash())
	resp.PeerCount = getStreamCount(node.GetStream())
	return resp, nil
}

func getStreamCount(m *sync.Map) uint32 {
	length := 0
	m.Range(func(_, _ interface{}) bool {
		length++
		return true
	})
	return uint32(length)
}

// Accounts is the RPC API handler.
func (s *APIService) Accounts(ctx context.Context, req *rpcpb.NonParamsRequest) (*rpcpb.AccountsResponse, error) {
	logging.VLog().WithFields(logrus.Fields{
		"api": "/v1/user/accounts",
	}).Info("Rpc request.")

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

// GetAccountState is the RPC API handler.
func (s *APIService) GetAccountState(ctx context.Context, req *rpcpb.GetAccountStateRequest) (*rpcpb.GetAccountStateResponse, error) {
	logging.VLog().WithFields(logrus.Fields{
		"address": req.Address,
		"block":   req.Block,
		"api":     "/v1/user/accountstate",
	}).Info("Rpc request.")

	neb := s.server.Neblet()

	addr, err := core.AddressParse(req.Address)
	if err != nil {
		return nil, err
	}

	block := neb.BlockChain().TailBlock()
	if len(req.Block) > 0 {
		blockHash, err := byteutils.FromHex(req.Block)
		if err != nil {
			return nil, err
		}
		block = neb.BlockChain().GetBlock(blockHash)
		if block == nil {
			return nil, errors.New("block hash not found")
		}
	}

	balance := block.GetBalance(addr.Bytes())
	nonce := block.GetNonce(addr.Bytes())

	return &rpcpb.GetAccountStateResponse{Balance: balance.String(), Nonce: fmt.Sprintf("%d", nonce)}, nil
}

// GetDynasty is the RPC API handler.
func (s *APIService) GetDynasty(ctx context.Context, req *rpcpb.NonParamsRequest) (*rpcpb.GetDynastyResponse, error) {
	logging.VLog().WithFields(logrus.Fields{
		"api": "/v1/admin/dynasty",
	}).Info("Rpc request.")

	neb := s.server.Neblet()
	dynastyRoot := neb.BlockChain().TailBlock().DposContext().DynastyRoot
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

// GetDelegateVoters is the RPC API handler.
func (s *APIService) GetDelegateVoters(ctx context.Context, req *rpcpb.GetDelegateVotersRequest) (*rpcpb.GetDelegateVotersResponse, error) {
	logging.VLog().WithFields(logrus.Fields{
		"delegatee": req.Delegatee,
		"api":       "/v1/admin/delegateVoters",
	}).Info("Rpc request.")

	neb := s.server.Neblet()
	delegatee, err := core.AddressParse(req.Delegatee)
	if err != nil {
		return nil, err
	}
	delegateRoot := neb.BlockChain().TailBlock().DposContext().DelegateRoot
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

// SendTransaction is the RPC API handler.
func (s *APIService) SendTransaction(ctx context.Context, req *rpcpb.TransactionRequest) (*rpcpb.SendTransactionResponse, error) {
	logging.VLog().WithFields(logrus.Fields{
		"api": "/v1/user/transaction",
	}).Info("Rpc request.")

	return s.sendTransaction(req)
}

// Call is the RPC API handler.
func (s *APIService) Call(ctx context.Context, req *rpcpb.TransactionRequest) (*rpcpb.SendTransactionResponse, error) {
	logging.VLog().WithFields(logrus.Fields{
		"api": "/v1/user/call",
	}).Info("Rpc request.")

	return s.sendTransaction(req)
}

func (s *APIService) sendTransaction(req *rpcpb.TransactionRequest) (*rpcpb.SendTransactionResponse, error) {
	neb := s.server.Neblet()
	tail := neb.BlockChain().TailBlock()
	addr, err := core.AddressParse(req.From)
	if err != nil {
		return nil, err
	}
	if req.Nonce <= tail.GetNonce(addr.Bytes()) {
		return nil, errors.New("nonce is invalid")
	}

	tx, err := parseTransaction(neb, req)
	if err != nil {
		return nil, err
	}
	if err := neb.AccountManager().SignTransaction(tx.From(), tx); err != nil {
		return nil, err
	}
	if err := neb.BlockChain().TransactionPool().PushAndBroadcast(tx); err != nil {
		return nil, err
	}
	if tx.Type() == core.TxPayloadDeployType {
		address, _ := core.NewContractAddressFromHash(hash.Sha3256(tx.From().Bytes(), byteutils.FromUint64(tx.Nonce())))
		return &rpcpb.SendTransactionResponse{Txhash: tx.Hash().String(), ContractAddress: address.String()}, nil
	}

	return &rpcpb.SendTransactionResponse{Txhash: tx.Hash().String()}, nil
}

func parseTransaction(neb Neblet, reqTx *rpcpb.TransactionRequest) (*core.Transaction, error) {
	fromAddr, err := core.AddressParse(reqTx.From)
	if err != nil {
		return nil, err
	}
	toAddr, err := core.AddressParse(reqTx.To)
	if err != nil {
		return nil, err
	}

	value := util.NewUint128FromString(reqTx.Value)
	gasPrice := util.NewUint128FromString(reqTx.GasPrice)
	gasLimit := util.NewUint128FromString(reqTx.GasLimit)

	var (
		payloadType string
		payload     []byte
	)
	if reqTx.Contract != nil && len(reqTx.Contract.Source) > 0 {
		payloadType = core.TxPayloadDeployType
		payload, err = core.NewDeployPayload(reqTx.Contract.Source, reqTx.Contract.SourceType, reqTx.Contract.Args).ToBytes()
	} else if reqTx.Contract != nil && len(reqTx.Contract.Function) > 0 {
		payloadType = core.TxPayloadCallType
		payload, err = core.NewCallPayload(reqTx.Contract.Function, reqTx.Contract.Args).ToBytes()
	} else if reqTx.Candidate != nil {
		payloadType = core.TxPayloadCandidateType
		payload, err = core.NewCandidatePayload(reqTx.Candidate.Action).ToBytes()
	} else if reqTx.Delegate != nil {
		payloadType = core.TxPayloadDelegateType
		payload, err = core.NewDelegatePayload(reqTx.Delegate.Action, reqTx.Delegate.Delegatee).ToBytes()
	} else {
		payloadType = core.TxPayloadBinaryType
	}
	if err != nil {
		return nil, err
	}

	tx := core.NewTransaction(neb.BlockChain().ChainID(), fromAddr, toAddr, value, reqTx.Nonce, payloadType, payload, gasPrice, gasLimit)
	return tx, nil
}

// SendRawTransaction submit the signed transaction raw data to txpool
func (s *APIService) SendRawTransaction(ctx context.Context, req *rpcpb.SendRawTransactionRequest) (*rpcpb.SendTransactionResponse, error) {
	logging.VLog().WithFields(logrus.Fields{
		"api": "/v1/user/rawtransaction",
	}).Info("Rpc request.")

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

	if tx.Type() == core.TxPayloadDeployType {
		address, _ := core.NewContractAddressFromHash(hash.Sha3256(tx.From().Bytes(), byteutils.FromUint64(tx.Nonce())))
		return &rpcpb.SendTransactionResponse{Txhash: tx.Hash().String(), ContractAddress: address.String()}, nil
	}

	return &rpcpb.SendTransactionResponse{Txhash: tx.Hash().String()}, nil
}

// GetBlockByHash get block info by the block hash
func (s *APIService) GetBlockByHash(ctx context.Context, req *rpcpb.GetBlockByHashRequest) (*corepb.Block, error) {
	logging.VLog().WithFields(logrus.Fields{
		"hash": req.Hash,
		"api":  "/v1/user/getBlockByHash",
	}).Info("Rpc request.")

	neb := s.server.Neblet()

	bhash, _ := byteutils.FromHex(req.GetHash())
	block := neb.BlockChain().GetBlock(bhash)
	if block == nil {
		return nil, errors.New("block not found")
	}
	pbBlock, err := block.ToProto()
	if err != nil {
		return nil, err
	}
	return pbBlock.(*corepb.Block), nil
}

// BlockDump is the RPC API handler.
func (s *APIService) BlockDump(ctx context.Context, req *rpcpb.BlockDumpRequest) (*rpcpb.BlockDumpResponse, error) {
	logging.VLog().WithFields(logrus.Fields{
		"count": req.Count,
		"api":   "/v1/user/transaction",
	}).Info("Rpc request.")

	neb := s.server.Neblet()
	data := neb.BlockChain().Dump(int(req.Count))
	return &rpcpb.BlockDumpResponse{Data: data}, nil
}

// GetTransactionReceipt get transaction info by the transaction hash
func (s *APIService) GetTransactionReceipt(ctx context.Context, req *rpcpb.GetTransactionByHashRequest) (*rpcpb.TransactionReceiptResponse, error) {
	logging.VLog().WithFields(logrus.Fields{
		"hash": req.Hash,
		"api":  "/v1/user/getTransactionReceipt",
	}).Info("Rpc request.")

	neb := s.server.Neblet()
	bhash, _ := byteutils.FromHex(req.GetHash())
	tx := neb.BlockChain().GetTransaction(bhash)
	if tx == nil {
		return nil, errors.New("transaction not found")
	}

	receipt := &rpcpb.TransactionReceiptResponse{
		ChainId:   tx.ChainID(),
		Hash:      byteutils.Hex(tx.Hash()),
		From:      tx.From().String(),
		To:        tx.To().String(),
		Value:     tx.Value().String(),
		Nonce:     tx.Nonce(),
		Timestamp: tx.Timestamp(),
		Type:      tx.Type(),
		Data:      byteutils.Hex(tx.Data()),
		GasPrice:  tx.GasPrice().String(),
		GasLimit:  tx.GasLimit().String(),
	}
	if tx.Type() == core.TxPayloadDeployType {
		contractAddr, err := tx.GenerateContractAddress()
		if err != nil {
			return nil, err
		}
		receipt.ContractAddress = contractAddr.String()
	}
	return receipt, nil
}

// NewAccount generate a new address with passphrase
func (s *APIService) NewAccount(ctx context.Context, req *rpcpb.NewAccountRequest) (*rpcpb.NewAccountResponse, error) {
	logging.VLog().WithFields(logrus.Fields{
		"api": "/v1/admin/account/new",
	}).Info("Rpc request.")

	neb := s.server.Neblet()
	addr, err := neb.AccountManager().NewAccount([]byte(req.Passphrase))
	if err != nil {
		return nil, err
	}
	return &rpcpb.NewAccountResponse{Address: addr.String()}, nil
}

// UnlockAccount unlock address with the passphrase
func (s *APIService) UnlockAccount(ctx context.Context, req *rpcpb.UnlockAccountRequest) (*rpcpb.UnlockAccountResponse, error) {
	logging.VLog().WithFields(logrus.Fields{
		"api": "/v1/admin/account/unlock",
	}).Info("Rpc request.")

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
func (s *APIService) LockAccount(ctx context.Context, req *rpcpb.LockAccountRequest) (*rpcpb.LockAccountResponse, error) {
	logging.VLog().WithFields(logrus.Fields{
		"api": "/v1/admin/account/lock",
	}).Info("Rpc request.")

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
func (s *APIService) SignTransaction(ctx context.Context, req *rpcpb.TransactionRequest) (*rpcpb.SignTransactionResponse, error) {
	logging.VLog().WithFields(logrus.Fields{
		"api": "/v1/admin/sign",
	}).Info("Rpc request.")

	neb := s.server.Neblet()
	tx, err := parseTransaction(neb, req)
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
func (s *APIService) SendTransactionWithPassphrase(ctx context.Context, req *rpcpb.SendTransactionPassphraseRequest) (*rpcpb.SendTransactionPassphraseResponse, error) {
	logging.VLog().WithFields(logrus.Fields{
		"api": "/v1/admin/transactionWithPassphrase",
	}).Info("Rpc request.")

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

// Subscribe ..
func (s *APIService) Subscribe(req *rpcpb.SubscribeRequest, gs rpcpb.ApiService_SubscribeServer) error {
	logging.VLog().WithFields(logrus.Fields{
		"topic": req.Topic,
		"api":   "/v1/user/subscribe",
	}).Info("Rpc request.")

	neb := s.server.Neblet()

	chainEventCh := make(chan *core.Event, 128)
	emitter := neb.EventEmitter()
	for _, v := range req.Topic {
		emitter.Register(v, chainEventCh)
	}

	defer (func() {
		for _, v := range req.Topic {
			emitter.Deregister(v, chainEventCh)
		}
	})()

	netEventCh := make(chan nnet.Message, 128)
	net := neb.NetManager()
	net.Register(nnet.NewSubscriber(s, netEventCh, core.MessageTypeNewBlock))
	net.Register(nnet.NewSubscriber(s, netEventCh, core.MessageTypeNewTx))
	defer net.Deregister(nnet.NewSubscriber(s, netEventCh, core.MessageTypeNewBlock))
	defer net.Deregister(nnet.NewSubscriber(s, netEventCh, core.MessageTypeNewTx))

	var err error
	for {
		select {
		case event := <-chainEventCh:
			err = gs.Send(&rpcpb.SubscribeResponse{MsgType: event.Topic, Data: event.Data})
			if err != nil {
				return err
			}
		case event := <-netEventCh:
			switch event.MessageType() {
			case core.MessageTypeNewBlock:
				block := new(core.Block)
				pbblock := new(corepb.Block)
				if err := proto.Unmarshal(event.Data().([]byte), pbblock); err != nil {
					return err
				}
				if err := block.FromProto(pbblock); err != nil {
					return err
				}
				blockjson, err := json.Marshal(block)
				if err != nil {
					return err
				}
				err = gs.Send(&rpcpb.SubscribeResponse{MsgType: event.MessageType(), Data: string(blockjson)})
			case core.MessageTypeNewTx:
				tx := new(core.Transaction)
				pbTx := new(corepb.Transaction)
				if err := proto.Unmarshal(event.Data().([]byte), pbTx); err != nil {
					return err
				}
				if err := tx.FromProto(pbTx); err != nil {
					return err
				}
				txjson, err := json.Marshal(tx)
				if err != nil {
					return err
				}
				err = gs.Send(&rpcpb.SubscribeResponse{MsgType: event.MessageType(), Data: string(txjson)})
			}
		}
	}
}

// GetGasPrice get gas price from chain.
func (s *APIService) GetGasPrice(ctx context.Context, req *rpcpb.NonParamsRequest) (*rpcpb.GasPriceResponse, error) {
	logging.VLog().WithFields(logrus.Fields{
		"api": "/v1/user/getGasPrice",
	}).Info("Rpc request.")

	neb := s.server.Neblet()
	gasPrice := neb.BlockChain().GasPrice()
	return &rpcpb.GasPriceResponse{GasPrice: gasPrice.String()}, nil
}

// EstimateGas Compute the smart contract gas consumption.
func (s *APIService) EstimateGas(ctx context.Context, req *rpcpb.TransactionRequest) (*rpcpb.EstimateGasResponse, error) {
	logging.VLog().WithFields(logrus.Fields{
		"api": "/v1/user/estimateGas",
	}).Info("Rpc request.")

	neb := s.server.Neblet()
	tail := neb.BlockChain().TailBlock()
	addr, err := core.AddressParse(req.From)
	if err != nil {
		return nil, err
	}
	if req.Nonce <= tail.GetNonce(addr.Bytes()) {
		return nil, errors.New("nonce is invalid")
	}

	tx, err := parseTransaction(neb, req)
	if err != nil {
		return nil, err
	}
	estimateGas, err := neb.BlockChain().EstimateGas(tx)
	if err != nil {
		return nil, err
	}
	return &rpcpb.EstimateGasResponse{EstimateGas: estimateGas.String()}, nil
}

// GetEventsByHash return events by tx hash.
func (s *APIService) GetEventsByHash(ctx context.Context, req *rpcpb.GetTransactionByHashRequest) (*rpcpb.EventsResponse, error) {
	logging.VLog().WithFields(logrus.Fields{
		"api": "/v1/user/getEventsByHash",
	}).Info("Rpc request.")

	neb := s.server.Neblet()
	bhash, _ := byteutils.FromHex(req.GetHash())
	tx, err := neb.BlockChain().TailBlock().GetTransaction(bhash)
	if err != nil {
		return nil, err
	}
	if tx != nil {
		result, err := neb.BlockChain().TailBlock().FetchEvents(tx.Hash())
		if err != nil {
			return nil, err
		}
		events := []*rpcpb.Event{}
		for _, v := range result {
			event := &rpcpb.Event{Topic: v.Topic, Data: v.Data}
			events = append(events, event)
		}

		return &rpcpb.EventsResponse{Events: events}, nil
	}

	return nil, nil

}

// ChangeNetworkID change the network id
func (s *APIService) ChangeNetworkID(ctx context.Context, req *rpcpb.ChangeNetworkIDRequest) (*rpcpb.ChangeNetworkIDResponse, error) {
	logging.VLog().WithFields(logrus.Fields{
		"api": "/v1/admin/changeNetworkID",
	}).Info("Rpc request.")

	neb := s.server.Neblet()
	neb.NetManager().Node().Config().NetworkID = req.NetworkId
	// broadcast to all the node in the routetable.
	neb.NetManager().BroadcastNetworkID(byteutils.FromUint32(req.NetworkId))
	return &rpcpb.ChangeNetworkIDResponse{Result: true}, nil
}
