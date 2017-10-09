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

// APIService implements the RPC API service interface.
type APIService struct{}

// GetBalance is the RPC API handler.
func (s *APIService) GetBalance(ctx context.Context, req *rpcpb.GetBalanceRequest) (*rpcpb.GetBalanceResponse, error) {
	if len(req.Address) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Address is empty.")
	}
	// TODO: Invoke core manager to get balance from block state. Remove fake logic.
	return &rpcpb.GetBalanceResponse{Value: 996}, nil
}

// SendTransaction is the RPC API handler.
func (s *APIService) SendTransaction(ctx context.Context, req *rpcpb.SendTransactionRequest) (*rpcpb.SendTransactionResponse, error) {
	if len(req.From) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Sender address is empty.")
	}
	// TODO: Invoke core manager to validate and sign the tx, then submit it to the tx pool. Remove fake logic.
	return &rpcpb.SendTransactionResponse{Hash: "0x07"}, nil
}
