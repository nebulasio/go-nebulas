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
	"strconv"

	"encoding/json"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/crypto/hash"
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/rpc/pb"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"golang.org/x/net/context"
)

//the max number of block can be dumped once
const maxDumpBlockCount = 10

// APIService implements the RPC API service interface.
type APIService struct {
	server GRPCServer
}

// GetNebState is the RPC API handler.
func (s *APIService) GetNebState(ctx context.Context, req *rpcpb.NonParamsRequest) (*rpcpb.GetNebStateResponse, error) {

	neb := s.server.Neblet()

	tail := neb.BlockChain().TailBlock()

	resp := &rpcpb.GetNebStateResponse{}
	resp.ChainId = neb.BlockChain().ChainID()
	resp.Tail = tail.Hash().String() // TODO add LIB
	resp.Height = tail.Height()
	resp.Coinbase = tail.Coinbase().String()                      // TODO move to nodeinfo
	resp.Synchronized = neb.NetService().Node().IsSynchronizing() // TODO use sync.IsActiveSyncing
	resp.PeerCount = uint32(neb.NetService().Node().PeersCount()) // TODO move to nodeinfo
	resp.ProtocolVersion = net.NebProtocolID
	resp.Version = neb.Config().App.Version

	return resp, nil
}

// NodeInfo is the RPC API handler
/*
限制来自同一个ip的节点连接请求的数量 （例如来自同一个ip的节点连接不能超过10，如果当前连接的列表中来自同一个ip的节点数量为10，则拒绝所有后面来自该ip的节点的连接请求）
主动发起连接时判断目标节点的ip是否在已连接的列表中，如果已经存在，则不建立该连接
路由同步增加相应的策略
一个桶的地址不能包含两个以上节点相同的 /24 ip地址块
整个路由表不能包含十个以上节点相同的 /24 ip地址块
更改路由同步的算法，路由同步时候不再同步离目标节点最近的那些节点
*/
func (s *APIService) NodeInfo(ctx context.Context, req *rpcpb.NonParamsRequest) (*rpcpb.NodeInfoResponse, error) {

	neb := s.server.Neblet()
	resp := &rpcpb.NodeInfoResponse{}
	node := neb.NetService().Node()
	resp.Id = node.ID() // TODO check eclipse attack
	resp.ChainId = node.Config().ChainID
	resp.BucketSize = int32(node.Config().Bucketsize)
	resp.Version = uint32(node.Config().Version)                            // TODO delete
	resp.StreamStoreSize = int32(node.Config().StreamStoreSize)             // TODO delete
	resp.StreamStoreExtendSize = int32(node.Config().StreamStoreExtendSize) // TODO delete
	resp.RelayCacheSize = int32(node.Config().RelayCacheSize)               // TODO delete
	resp.PeerCount = uint32(node.PeersCount())
	resp.ProtocolVersion = net.NebProtocolID

	for k, v := range node.RouteTable().Peers() {
		routeTable := &rpcpb.RouteTable{}
		routeTable.Id = k.Pretty()
		routeTable.Address = make([]string, len(v))

		for i, addr := range v {
			routeTable.Address[i] = addr.String()
		}
		resp.RouteTable = append(resp.RouteTable, routeTable)
	}

	return resp, nil
}

// Accounts is the RPC API handler.
func (s *APIService) Accounts(ctx context.Context, req *rpcpb.NonParamsRequest) (*rpcpb.AccountsResponse, error) {

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

	neb := s.server.Neblet()

	addr, err := core.AddressParse(req.Address)
	if err != nil {
		metricsAccountStateFailed.Mark(1)
		return nil, err
	}

	block := neb.BlockChain().TailBlock()
	if req.Height > 0 {
		block = neb.BlockChain().GetBlockOnCanonicalChainByHeight(req.Height)
		if block == nil {
			metricsAccountStateFailed.Mark(1)
			return nil, errors.New("block not found")
		}
	}

	balance, err := block.GetBalance(addr.Bytes()) // TODO combine the two
	if err != nil {
		return nil, err
	}
	nonce, err := block.GetNonce(addr.Bytes())
	if err != nil {
		return nil, err
	}

	metricsAccountStateSuccess.Mark(1)
	return &rpcpb.GetAccountStateResponse{Balance: balance.String(), Nonce: nonce}, nil
}

// SendTransaction is the RPC API handler.
func (s *APIService) SendTransaction(ctx context.Context, req *rpcpb.TransactionRequest) (*rpcpb.SendTransactionResponse, error) {

	return s.sendTransaction(req)
}

// Call is the RPC API handler.
func (s *APIService) Call(ctx context.Context, req *rpcpb.TransactionRequest) (*rpcpb.CallResponse, error) { // TODO rename Simulate

	neb := s.server.Neblet()
	tx, err := parseTransaction(neb, req)
	if err != nil {
		return nil, err
	}
	_, result, err := neb.BlockChain().EstimateGas(tx)
	if err != nil {
		return nil, err
	}
	return &rpcpb.CallResponse{Result: result}, nil // TODO return gas, call result, err
}

func (s *APIService) sendTransaction(req *rpcpb.TransactionRequest) (*rpcpb.SendTransactionResponse, error) {
	neb := s.server.Neblet()
	tx, err := parseTransaction(neb, req)
	if err != nil {
		metricsSendTxFailed.Mark(1)
		return nil, err
	}
	if err := neb.AccountManager().SignTransaction(tx.From(), tx); err != nil {
		metricsSendTxFailed.Mark(1)
		return nil, err
	}

	return handleTransactionResponse(neb, tx)
}

func parseTransaction(neb core.Neblet, reqTx *rpcpb.TransactionRequest) (*core.Transaction, error) {
	fromAddr, err := core.AddressParse(reqTx.From)
	if err != nil {
		return nil, err
	}
	toAddr, err := core.AddressParse(reqTx.To)
	if err != nil {
		return nil, err
	}

	value, err := util.NewUint128FromString(reqTx.Value)
	if err != nil {
		return nil, errors.New("invalid value")
	}
	gasPrice, err := util.NewUint128FromString(reqTx.GasPrice)
	if err != nil {
		return nil, errors.New("invalid gasPrice")
	}
	gasLimit, err := util.NewUint128FromString(reqTx.GasLimit)
	if err != nil {
		return nil, errors.New("invalid gasLimit")
	}
	var (
		payloadType string
		payload     []byte
	) // TODO check payloadType, contract: deploy/call, binary, check args
	if reqTx.Contract != nil && len(reqTx.Contract.Source) > 0 { // TODO combine contract != nil
		payloadType = core.TxPayloadDeployType
		payload, err = core.NewDeployPayload(reqTx.Contract.Source, reqTx.Contract.SourceType, reqTx.Contract.Args).ToBytes()
	} else if reqTx.Contract != nil && len(reqTx.Contract.Function) > 0 {
		payloadType = core.TxPayloadCallType

		if len(reqTx.Contract.Args) > 0 { // TODO move the check into NewCallPayload
			var argsObj []interface{}
			if err := json.Unmarshal([]byte(reqTx.Contract.Args), &argsObj); err != nil {
				err = errors.New("contract arguments format error")
			}
		} // ToRefine: extract a function to check args.

		if err == nil {
			payload, err = core.NewCallPayload(reqTx.Contract.Function, reqTx.Contract.Args).ToBytes()
		}
	} else {
		payloadType = core.TxPayloadBinaryType
		payload, err = core.NewBinaryPayload(reqTx.Binary).ToBytes()
	}
	if err != nil {
		return nil, err
	}

	tx, err := core.NewTransaction(neb.BlockChain().ChainID(), fromAddr, toAddr, value, reqTx.Nonce, payloadType, payload, gasPrice, gasLimit)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func handleTransactionResponse(neb core.Neblet, tx *core.Transaction) (resp *rpcpb.SendTransactionResponse, err error) {
	defer func() {
		if err != nil {
			metricsSendTxFailed.Mark(1)
		} else {
			metricsSendTxSuccess.Mark(1)
		}
	}() // TODO add tx.VerifyIntegrity
	nonce, err := neb.BlockChain().TailBlock().GetNonce(tx.From().Bytes()) // TODO move check into core
	if err != nil {
		return nil, err
	}
	if tx.Nonce() <= nonce {

		return nil, errors.New("transaction's nonce is invalid, should bigger than the from's nonce")
	}

	if tx.Type() == core.TxPayloadDeployType {
		if !tx.From().Equals(tx.To()) {
			return nil, core.ErrContractTransactionAddressNotEqual
		}
	} else if tx.Type() == core.TxPayloadCallType {
		if _, err := neb.BlockChain().TailBlock().CheckContract(tx.To()); err != nil {
			return nil, err
		}
	}

	// push and broadcast tx
	if err := neb.BlockChain().TransactionPool().PushAndBroadcast(tx); err != nil {
		return nil, err
	}

	var contract string
	if tx.Type() == core.TxPayloadDeployType {
		addr, err := core.NewContractAddressFromHash(hash.Sha3256(tx.From().Bytes(), byteutils.FromUint64(tx.Nonce())))
		if err != nil {
			return nil, err
		}
		contract = addr.String()
	}

	return &rpcpb.SendTransactionResponse{Txhash: tx.Hash().String(), ContractAddress: contract}, nil
}

// SendRawTransaction submit the signed transaction raw data to txpool
func (s *APIService) SendRawTransaction(ctx context.Context, req *rpcpb.SendRawTransactionRequest) (*rpcpb.SendTransactionResponse, error) {

	// Validate and sign the tx, then submit it to the tx pool.
	neb := s.server.Neblet()

	pbTx := new(corepb.Transaction)
	if err := proto.Unmarshal(req.GetData(), pbTx); err != nil {
		metricsSendTxFailed.Mark(1)
		return nil, err
	}
	tx := new(core.Transaction)
	if err := tx.FromProto(pbTx); err != nil {
		metricsSendTxFailed.Mark(1)
		return nil, err
	}

	return handleTransactionResponse(neb, tx)
}

// GetBlockByHash get block info by the block hash
func (s *APIService) GetBlockByHash(ctx context.Context, req *rpcpb.GetBlockByHashRequest) (*rpcpb.BlockResponse, error) { // TODO fulltransaction -> fullfill

	neb := s.server.Neblet()

	bhash, err := byteutils.FromHex(req.GetHash())
	if err != nil {
		return nil, err
	}
	block := neb.BlockChain().GetBlock(bhash) // TODO use neb.BlockChain().GetBlockOnCanonicalChainByHash(bhash)

	return s.toBlockResponse(block, req.FullTransaction) // TODO return if the block become finality
}

// GetBlockByHeight get block info by the block hash
func (s *APIService) GetBlockByHeight(ctx context.Context, req *rpcpb.GetBlockByHeightRequest) (*rpcpb.BlockResponse, error) {

	neb := s.server.Neblet()

	block := neb.BlockChain().GetBlockOnCanonicalChainByHeight(req.Height) // TODO return if the block become finality

	return s.toBlockResponse(block, req.FullTransaction)
}

func (s *APIService) toBlockResponse(block *core.Block, fullTransaction bool) (*rpcpb.BlockResponse, error) {
	if block == nil {
		return nil, errors.New("block not found")
	}

	resp := &rpcpb.BlockResponse{
		Hash:          block.Hash().String(),
		ParentHash:    block.ParentHash().String(),
		Height:        block.Height(),
		Coinbase:      block.Coinbase().String(),
		Timestamp:     block.Timestamp(),
		ChainId:       block.ChainID(),
		StateRoot:     block.StateRoot().String(),
		TxsRoot:       block.TxsRoot().String(),
		EventsRoot:    block.EventsRoot().String(),
		ConsensusRoot: block.ConsensusRoot(),
	}

	// add block transactions
	txs := []*rpcpb.TransactionResponse{}
	for _, v := range block.Transactions() {
		var tx *rpcpb.TransactionResponse
		if fullTransaction {
			tx, _ = s.toTransactionResponse(v)
		} else {
			tx = &rpcpb.TransactionResponse{Hash: v.Hash().String()}
		}
		txs = append(txs, tx)
	}
	resp.Transactions = txs

	return resp, nil
}

// BlockDump is the RPC API handler.
func (s *APIService) BlockDump(ctx context.Context, req *rpcpb.BlockDumpRequest) (*rpcpb.BlockDumpResponse, error) { // TODO delete

	neb := s.server.Neblet()
	if req.Count > maxDumpBlockCount {
		return nil, errors.New("the max count of blocks could be dumped once is " + strconv.Itoa(maxDumpBlockCount))
	}
	if req.Count < 0 {
		return nil, errors.New("invalid count")
	}
	data := neb.BlockChain().Dump(int(req.Count))
	return &rpcpb.BlockDumpResponse{Data: data}, nil
}

// LatestIrreversibleBlock is the RPC API handler.
func (s *APIService) LatestIrreversibleBlock(ctx context.Context, req *rpcpb.NonParamsRequest) (*rpcpb.BlockResponse, error) {

	neb := s.server.Neblet()
	block := neb.BlockChain().LIB()

	return s.toBlockResponse(block, false)
}

// GetTransactionReceipt get transaction info by the transaction hash
func (s *APIService) GetTransactionReceipt(ctx context.Context, req *rpcpb.GetTransactionByHashRequest) (*rpcpb.TransactionResponse, error) {

	neb := s.server.Neblet()
	hash, err := byteutils.FromHex(req.GetHash())
	if err != nil {
		return nil, err
	}
	tx := neb.BlockChain().GetTransaction(hash)

	// if tx is nil, check it in transaction pool.
	if tx == nil {
		tx = neb.BlockChain().TransactionPool().GetTransaction(hash) // TODO make tx pending when collecttxs
		if tx == nil {
			return nil, errors.New("transaction not found")
		}
	}

	return s.toTransactionResponse(tx)
}

func (s *APIService) toTransactionResponse(tx *core.Transaction) (*rpcpb.TransactionResponse, error) {
	var (
		status  int32
		gasUsed string
	)
	neb := s.server.Neblet()
	events, _ := neb.BlockChain().TailBlock().FetchEvents(tx.Hash()) // TODO catch error

	if events != nil && len(events) > 0 {
		for _, v := range events {
			if v.Topic == core.TopicTransactionExecutionResult { // TODO use last one
				txEvent := core.TransactionEvent{}
				json.Unmarshal([]byte(v.Data), &txEvent) // TODO catch error
				status = int32(txEvent.Status)
				gasUsed = txEvent.GasUsed
				break
			}
		}
	} else {
		status = core.TxExecutionPendding
	}

	resp := &rpcpb.TransactionResponse{
		ChainId:   tx.ChainID(),
		Hash:      tx.Hash().String(),
		From:      tx.From().String(),
		To:        tx.To().String(),
		Value:     tx.Value().String(),
		Nonce:     tx.Nonce(),
		Timestamp: tx.Timestamp(),
		Type:      tx.Type(),
		Data:      tx.Data(),
		GasPrice:  tx.GasPrice().String(),
		GasLimit:  tx.GasLimit().String(),
		Status:    status,
		GasUsed:   gasUsed,
	}

	/* 	contractAddr, err := tx.TryGetContractAddress()
	   	if err != nil {
	   		return nil, err
	   	}

	   	if contractAddr != nil {
	   		resp.ContractAddress = contractAddr.String()
	   	}

	   	return resp */

	if tx.Type() == core.TxPayloadDeployType { // TODO use TryGetContractAddress
		contractAddr, err := tx.GenerateContractAddress()
		if err != nil {
			return nil, err
		}
		resp.ContractAddress = contractAddr.String()
	}
	return resp, nil
}

// Subscribe ..
func (s *APIService) Subscribe(req *rpcpb.SubscribeRequest, gs rpcpb.ApiService_SubscribeServer) error {

	neb := s.server.Neblet()

	eventSub := core.NewEventSubscriber(1024, req.Topics)
	neb.EventEmitter().Register(eventSub)
	defer neb.EventEmitter().Deregister(eventSub)

	var err error
	for {
		select {
		case <-gs.Context().Done(): // TODO add test
			return gs.Context().Err()
		case event := <-eventSub.EventChan():
			err = gs.Send(&rpcpb.SubscribeResponse{Topic: event.Topic, Data: event.Data})
			if err != nil {
				return err
			}
		}
	}
}

// GetGasPrice get gas price from chain.
func (s *APIService) GetGasPrice(ctx context.Context, req *rpcpb.NonParamsRequest) (*rpcpb.GasPriceResponse, error) {

	neb := s.server.Neblet()
	gasPrice := neb.BlockChain().GasPrice()
	return &rpcpb.GasPriceResponse{GasPrice: gasPrice.String()}, nil
}

// EstimateGas Compute the smart contract gas consumption.
func (s *APIService) EstimateGas(ctx context.Context, req *rpcpb.TransactionRequest) (*rpcpb.GasResponse, error) { // TODO update logic

	neb := s.server.Neblet()
	tx, err := parseTransaction(neb, req)
	if err != nil {
		return nil, err
	}
	estimateGas, _, err := neb.BlockChain().EstimateGas(tx)
	if err != nil {
		return nil, err
	}
	return &rpcpb.GasResponse{Gas: estimateGas.String()}, nil
}

// GetEventsByHash return events by tx hash.
func (s *APIService) GetEventsByHash(ctx context.Context, req *rpcpb.HashRequest) (*rpcpb.EventsResponse, error) {

	neb := s.server.Neblet()

	if len(req.Hash) == 0 {
		return nil, errors.New("please input valid hash")
	}

	txhash, err := byteutils.FromHex(req.Hash)
	if err != nil {
		return nil, err
	}

	tailBlock := neb.BlockChain().TailBlock()
	tx, err := tailBlock.GetTransaction(txhash)
	if err != nil {
		return nil, err
	}

	result, err := tailBlock.FetchEvents(tx.Hash())
	if err != nil {
		return nil, err
	}

	events := make([]*rpcpb.Event, len(result))
	for idx, v := range result {
		event := &rpcpb.Event{Topic: v.Topic, Data: v.Data}
		events[idx] = event
	}

	return &rpcpb.EventsResponse{Events: events}, nil
}

// GetDynasty is the RPC API handler.
func (s *APIService) GetDynasty(ctx context.Context, req *rpcpb.ByBlockHeightRequest) (*rpcpb.GetDynastyResponse, error) {
	neb := s.server.Neblet()
	var block *core.Block
	if req.Height > 0 {
		block = neb.BlockChain().GetBlockOnCanonicalChainByHeight(req.Height)
		if block == nil {
			return nil, errors.New("block not found")
		}
	} else {
		block = neb.BlockChain().TailBlock()
	}
	validators, err := block.WorldState().Dynasty() // TODO add dynasty function in block, clone first
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, v := range validators {
		result = append(result, string(v.Hex()))
	}
	return &rpcpb.GetDynastyResponse{Delegatees: result}, nil // TODO rename Delegatees -> Validators
}
