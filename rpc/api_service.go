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
	resp.Tail = string(tail.Hash().Hex())
	resp.Coinbase = string(tail.Coinbase())
	resp.Synchronized = neb.NetService().Node().GetSynchronized()
	resp.PeerCount = uint32(len(neb.NetService().Node().GetStream()))
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
	resp.PeerCount = int32(len(node.GetStream()))
	resp.ProtocolVersion = p2p.ProtocolID
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
	balance := neb.BlockChain().TailBlock().GetBalance(addr.Bytes())
	nonce := neb.BlockChain().TailBlock().GetNonce(addr.Bytes())

	balanceBytes, err := balance.ToFixedSizeByteSlice()
	if err != nil {
		return nil, err
	}
	return &rpcpb.GetAccountStateResponse{Balance: balanceBytes, Nonce: nonce}, nil
}

// SendTransaction is the RPC API handler.
func (s *APIService) SendTransaction(ctx context.Context, req *rpcpb.SendTransactionRequest) (*rpcpb.SendTransactionResponse, error) {
	neb := s.server.Neblet()

	var data []byte
	var err error
	payloadType := core.TxPayloadBinaryType
	if len(req.Source) > 0 {
		data, err = core.NewDeployPayload(req.Source, req.Args).ToBytes()
		payloadType = core.TxPayloadDeployType
		if err != nil {
			return nil, err
		}
	}

	tx, err := parseTransaction(neb, req.From, req.To, req.Value, req.Nonce, payloadType, data)
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
		return &rpcpb.SendTransactionResponse{Hash: address.ToHex() + "$" + tx.Hash().String()}, nil
	}

	return &rpcpb.SendTransactionResponse{Hash: tx.Hash().String()}, nil

}

// Call is the RPC API handler.
func (s *APIService) Call(ctx context.Context, req *rpcpb.CallRequest) (*rpcpb.SendTransactionResponse, error) {
	neb := s.server.Neblet()
	data, err := core.NewCallPayload(req.Function, req.Args).ToBytes()
	if err != nil {
		return nil, err
	}
	tx, err := parseTransaction(neb, req.From, req.To, nil, req.Nonce, core.TxPayloadCallType, data)
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

func parseTransaction(neb Neblet, from, to string, v []byte, nonce uint64, payloadType string, payload []byte) (*core.Transaction, error) {
	fromAddr, err := core.AddressParse(from)
	if err != nil {
		return nil, err
	}
	toAddr, err := core.AddressParse(to)
	if err != nil {
		return nil, err
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

	tx := core.NewTransaction(neb.BlockChain().ChainID(), fromAddr, toAddr, value, nonce, payloadType, payload)
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
