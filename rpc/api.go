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
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// APIService implements the RPC API service interface.
type APIService struct {
	server *Server
}

// GetBalance is the RPC API handler.
func (s *APIService) GetBalance(ctx context.Context, req *rpcpb.GetBalanceRequest) (*rpcpb.GetBalanceResponse, error) {
	if len(req.Address) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Address is empty.")
	}

	// TODO: cleanup dummy logic.
	bal := util.NewUint128FromInt(996)
	if s.server != nil {
		if neb := s.server.Neblet(); neb != nil {
			addr, _ := byteutils.FromHex(req.Address)
			// TODO: handle specific block number.
			bal = neb.BlockChain().TailBlock().GetBalance(addr)
		}
	}

	vb, err := bal.ToFixedSizeByteSlice()
	if err != nil {
		return nil, err
	}
	return &rpcpb.GetBalanceResponse{Value: vb}, nil
}

// SendTransaction is the RPC API handler.
func (s *APIService) SendTransaction(ctx context.Context, req *rpcpb.SendTransactionRequest) (*rpcpb.SendTransactionResponse, error) {
	if len(req.From) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Sender address is empty.")
	}
	// TODO: cleanup dummy logic.
	if s.server == nil {
		return &rpcpb.SendTransactionResponse{Hash: "0x07"}, nil
	}

	// Validate and sign the tx, then submit it to the tx pool.
	if neb := s.server.Neblet(); neb != nil {
		from, err := byteutils.FromHex(req.From)
		if err != nil {
			return nil, err
		}
		fromAddr, err := core.NewAddress(from)
		if err != nil {
			return nil, err
		}

		to, err := byteutils.FromHex(req.To)
		if err != nil {
			return nil, err
		}
		toAddr, err := core.NewAddress(to)
		if err != nil {
			return nil, err
		}

		value, err := util.NewUint128FromFixedSizeByteSlice(req.Value)
		if err != nil {
			return nil, err
		}

		// TODO: use uint128 instead of uint64.
		tx := core.NewTransaction(neb.BlockChain().ChainID(), fromAddr, toAddr, value, req.Nonce /*req.Data */, nil)
		if err := neb.AccountManager().SignTransaction(fromAddr, tx); err != nil {
			return nil, err
		}

		if err := neb.BlockChain().TransactionPool().Push(tx); err != nil {
			return nil, err
		}
	}

	// TODO: returns the transaction hash if available.
	return &rpcpb.SendTransactionResponse{}, nil
}
