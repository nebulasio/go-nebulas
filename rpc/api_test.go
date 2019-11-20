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
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/nebulasio/go-nebulas/rpc/mock_pb"
	rpcpb "github.com/nebulasio/go-nebulas/rpc/pb"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestAPIService_GetNebState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_pb.NewMockAPIServiceClient(ctrl)

	{
		req := &rpcpb.NonParamsRequest{}
		expected := &rpcpb.GetNebStateResponse{Tail: "hac"}
		client.EXPECT().GetNebState(gomock.Any(), gomock.Any()).Return(expected, nil)
		resp, _ := client.GetNebState(context.Background(), req)
		assert.Equal(t, expected, resp)
	}

	// TODO: test with mock neblet.
}

func TestGetAccountState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_pb.NewMockAPIServiceClient(ctrl)

	{
		req := &rpcpb.GetAccountStateRequest{Address: "0xf"}
		tmpNumber, _ := util.NewUint128FromInt(31415926)
		bal := tmpNumber.String()
		expected := &rpcpb.GetAccountStateResponse{Balance: bal, Nonce: 1}
		client.EXPECT().GetAccountState(gomock.Any(), gomock.Any()).Return(expected, nil)
		resp, _ := client.GetAccountState(context.Background(), req)
		assert.Equal(t, expected, resp)
	}

	// TODO: test with mock neblet.
}

func TestSendTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_pb.NewMockAPIServiceClient(ctrl)

	{
		req := &rpcpb.TransactionRequest{From: "0xf"}
		expected := &rpcpb.SendTransactionResponse{Txhash: "0x2"}
		client.EXPECT().SendTransaction(gomock.Any(), gomock.Any()).Return(expected, nil)
		resp, _ := client.SendTransaction(context.Background(), req)
		assert.Equal(t, expected, resp)
	}

	// TODO: test with mock neblet.
}
