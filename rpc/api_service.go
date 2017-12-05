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
	"fmt"
	"sync"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/core"
	corepb "github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/crypto/hash"
	"github.com/nebulasio/go-nebulas/net/p2p"
	"github.com/nebulasio/go-nebulas/rpc/pb"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
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
	resp.ChainId = neb.BlockChain().ChainID()
	resp.Tail = tail.Hash().String()
	resp.Coinbase = tail.Coinbase().ToHex()
	resp.Synchronized = neb.NetService().Node().GetSynchronized()
	resp.PeerCount = getStreamCount(neb.NetService().Node().GetStream())
	resp.ProtocolVersion = p2p.ProtocolID

	return resp, nil
}

// NodeInfo is the PRC API handler
func (s *APIService) NodeInfo(ctx context.Context, req *rpcpb.NodeInfoRequest) (*rpcpb.NodeInfoResponse, error) {
	neb := s.server.Neblet()
	resp := &rpcpb.NodeInfoResponse{}
	node := neb.NetService().Node()
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
func (s *APIService) StatisticsNodeInfo(ctx context.Context, req *rpcpb.NodeInfoRequest) (*rpcpb.StatisticsNodeInfoResponse, error) {
	neb := s.server.Neblet()
	node := neb.NetService().Node()
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
	balance := neb.BlockChain().TailBlock().GetBalance(addr.Bytes())
	nonce := neb.BlockChain().TailBlock().GetNonce(addr.Bytes())

	return &rpcpb.GetAccountStateResponse{Balance: balance.String(), Nonce: fmt.Sprintf("%d", nonce)}, nil
}

// SendTransaction is the RPC API handler.
func (s *APIService) SendTransaction(ctx context.Context, req *rpcpb.SendTransactionRequest) (*rpcpb.SendTransactionResponse, error) {
	neb := s.server.Neblet()

	var data []byte
	var err error
	payloadType := core.TxPayloadBinaryType
	tail := neb.BlockChain().TailBlock()
	addr, err := core.AddressParse(req.From)
	if err != nil {
		return nil, err
	}
	if req.Nonce <= tail.GetNonce(addr.Bytes()) {
		return nil, errors.New("nonce is invalid")
	}
	if len(req.Source) > 0 {
		data, err = core.NewDeployPayload(req.Source, req.Args).ToBytes()
		payloadType = core.TxPayloadDeployType
		if err != nil {
			return nil, err
		}
	}

	tx, err := parseTransaction(neb, req.From, req.To, req.Value, req.Nonce, payloadType, data, req.GasPrice, req.GasLimit)
	if err != nil {
		return nil, err
	}
	if err := neb.AccountManager().SignTransaction(tx.From(), tx); err != nil {
		return nil, err
	}
	if err := neb.BlockChain().TransactionPool().PushAndBroadcast(tx); err != nil {
		return nil, err
	}
	if len(req.Source) > 0 {
		address, _ := core.NewContractAddressFromHash(hash.Sha3256(tx.From().Bytes(), byteutils.FromUint64(tx.Nonce())))
		return &rpcpb.SendTransactionResponse{Txhash: tx.Hash().String(), ContractAddress: address.ToHex()}, nil
	}

	return &rpcpb.SendTransactionResponse{Txhash: tx.Hash().String()}, nil

}

// Call is the RPC API handler.
func (s *APIService) Call(ctx context.Context, req *rpcpb.CallRequest) (*rpcpb.SendTransactionResponse, error) {
	neb := s.server.Neblet()
	tail := neb.BlockChain().TailBlock()
	addr, err := core.AddressParse(req.From)
	if err != nil {
		return nil, err
	}
	if req.Nonce <= tail.GetNonce(addr.Bytes()) {
		return nil, errors.New("nonce is invalid")
	}
	data, err := core.NewCallPayload(req.Function, req.Args).ToBytes()
	if err != nil {
		return nil, err
	}
	tx, err := parseTransaction(neb, req.From, req.To, req.Value, req.Nonce, core.TxPayloadCallType, data, req.GasPrice, req.GasLimit)
	if err != nil {
		return nil, err
	}
	if err := neb.AccountManager().SignTransaction(tx.From(), tx); err != nil {
		return nil, err
	}
	if err := neb.BlockChain().TransactionPool().PushAndBroadcast(tx); err != nil {
		return nil, err
	}

	return &rpcpb.SendTransactionResponse{Txhash: tx.Hash().String()}, nil

}

func parseTransaction(neb Neblet, from, to string, v string, nonce uint64, payloadType string, payload []byte, price string, limit string) (*core.Transaction, error) {
	fromAddr, err := core.AddressParse(from)
	if err != nil {
		return nil, err
	}
	toAddr, err := core.AddressParse(to)
	if err != nil {
		return nil, err
	}

	value := util.NewUint128FromString(v)
	gasPrice := util.NewUint128FromString(price)
	gasLimit := util.NewUint128FromString(limit)

	tx := core.NewTransaction(neb.BlockChain().ChainID(), fromAddr, toAddr, value, nonce, payloadType, payload, gasPrice, gasLimit)
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

	return &rpcpb.SendTransactionResponse{Txhash: tx.Hash().String()}, nil
}

// GetBlockByHash get block info by the block hash
func (s *APIService) GetBlockByHash(ctx context.Context, req *rpcpb.GetBlockByHashRequest) (*corepb.Block, error) {
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

// // GetTransactionByHash get transaction info by the transaction hash
// func (s *APIService) GetTransactionByHash(ctx context.Context, req *rpcpb.GetTransactionByHashRequest) (*corepb.Transaction, error) {
// 	neb := s.server.Neblet()
// 	bhash, _ := byteutils.FromHex(req.GetHash())
// 	tx := neb.BlockChain().GetTransaction(bhash)
// 	if tx == nil {
// 		return nil, errors.New("transaction not found")
// 	}
// 	pbTx, err := tx.ToProto()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return pbTx.(*corepb.Transaction), nil
// }

// BlockDump is the RPC API handler.
func (s *APIService) BlockDump(ctx context.Context, req *rpcpb.BlockDumpRequest) (*rpcpb.BlockDumpResponse, error) {
	neb := s.server.Neblet()
	data := neb.BlockChain().Dump(int(req.Count))
	return &rpcpb.BlockDumpResponse{Data: data}, nil
}

// GetTransactionReceipt get transaction info by the transaction hash
func (s *APIService) GetTransactionReceipt(ctx context.Context, req *rpcpb.GetTransactionByHashRequest) (*rpcpb.TransactionReceiptResponse, error) {
	neb := s.server.Neblet()
	bhash, _ := byteutils.FromHex(req.GetHash())
	tx := neb.BlockChain().GetTransaction(bhash)
	if tx == nil {
		return nil, errors.New("transaction not found")
	}
	if tx.From().ToHex() == tx.To().ToHex() {
		contractAddr, err := tx.GenerateContractAddress()
		if err != nil {
			return nil, err
		}
		return &rpcpb.TransactionReceiptResponse{
			Hash:            byteutils.Hex(tx.Hash()),
			From:            tx.From().ToHex(),
			To:              tx.To().ToHex(),
			Nonce:           tx.Nonce(),
			Timestamp:       tx.Timestamp(),
			ChainId:         tx.ChainID(),
			ContractAddress: contractAddr.ToHex(),
			Data:            string(tx.Data()),
		}, nil
	}

	return &rpcpb.TransactionReceiptResponse{
		Hash:      byteutils.Hex(tx.Hash()),
		From:      tx.From().ToHex(),
		To:        tx.To().ToHex(),
		Nonce:     tx.Nonce(),
		Timestamp: tx.Timestamp(),
		ChainId:   tx.ChainID(),
	}, nil
}

// NewAccount generate a new address with passphrase
func (s *APIService) NewAccount(ctx context.Context, req *rpcpb.NewAccountRequest) (*rpcpb.NewAccountResponse, error) {
	neb := s.server.Neblet()
	addr, err := neb.AccountManager().NewAccount([]byte(req.Passphrase))
	if err != nil {
		return nil, err
	}
	return &rpcpb.NewAccountResponse{Address: addr.ToHex()}, nil
}

// UnlockAccount unlock address with the passphrase
func (s *APIService) UnlockAccount(ctx context.Context, req *rpcpb.UnlockAccountRequest) (*rpcpb.UnlockAccountResponse, error) {
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
func (s *APIService) SignTransaction(ctx context.Context, req *rpcpb.SignTransactionRequest) (*rpcpb.SignTransactionResponse, error) {
	neb := s.server.Neblet()
	tx, err := parseTransaction(neb, req.From, req.To, req.Value, req.Nonce, core.TxPayloadBinaryType, nil, req.GasPrice, req.GasLimit)
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
	neb := s.server.Neblet()
	tx, err := parseTransaction(neb, req.From, req.To, req.Value, req.Nonce, core.TxPayloadBinaryType, nil, req.GasPrice, req.GasLimit)
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
