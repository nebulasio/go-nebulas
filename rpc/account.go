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
	"github.com/nebulasio/go-nebulas/rpc/pb"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AccountServer implements the rpc interface.
type AccountServer struct{}

// GetBalance is the rpc handler.
func (s *AccountServer) GetBalance(ctx context.Context, req *rpcpb.GetBalanceRequest) (*rpcpb.GetBalanceResponse, error) {
	if len(req.Address) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Address is empty.")
	}
	// TODO: Implement real logic -- get balance from block state.
	return &rpcpb.GetBalanceResponse{Value: 996}, nil
}

// SendTransaction is the rpc handler.
func (s *AccountServer) SendTransaction(ctx context.Context, req *rpcpb.SendTransactionRequest) (*rpcpb.SendTransactionResponse, error) {
	if len(req.From) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Sender address is empty.")
	}
	// TODO: Implement real logic -- sign the tx, and submit it to the tx pool.
	return &rpcpb.SendTransactionResponse{Hash: "0x07"}, nil
}
