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

package main

import (
	"io"
	"log"
	"strconv"

	"fmt"

	"github.com/nebulasio/go-nebulas/rpc"
	"github.com/nebulasio/go-nebulas/rpc/pb"
	"github.com/nebulasio/go-nebulas/util"
	"golang.org/x/net/context"
)

// TODO: add command line flag.
const (
	//config = "../../../../config.pb.txt"
	from  = "8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf"
	to    = "22ac3a9a2b1c31b7a9084e46eae16e761f83f02324092b09"
	value = 2
)

// RPC testing client.
func main() {
	// Set up a connection to the server.
	//cfg := neblet.LoadConfig(config).Rpc
	addr := fmt.Sprintf("127.0.0.1:%d", uint32(51510))
	conn, err := rpc.Dial(addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ac := rpcpb.NewApiServiceClient(conn)
	var nonce uint64

	{
		r, err := ac.GetNebState(context.Background(), &rpcpb.NonParamsRequest{})
		if err != nil {
			log.Println("GetNebState", "failed", err)
		} else {
			//tail := r.GetTail()
			log.Println("GetNebState tail", r)
		}
	}

	{
		r, err := ac.GetAccountState(context.Background(), &rpcpb.GetAccountStateRequest{Address: from})
		if err != nil {
			log.Println("GetAccountState", from, "failed", err)
		} else {
			val := util.NewUint128FromString(r.GetBalance())
			nonce, _ = strconv.ParseUint(r.Nonce, 10, 64)
			// nonce = r.Nonce
			log.Println("GetAccountState", from, "nonce", r.Nonce, "value", val)
		}
	}

	{
		v := util.NewUint128FromInt(value)
		r, err := ac.SendTransaction(context.Background(), &rpcpb.TransactionRequest{From: from, To: to, Value: v.String(), Nonce: nonce + 1})
		if err != nil {
			log.Println("SendTransaction failed:", err)
		} else {
			log.Println("SendTransaction", from, "->", to, "value", value, r)
		}
	}

	{
		r, err := ac.GetAccountState(context.Background(), &rpcpb.GetAccountStateRequest{Address: to})
		if err != nil {
			log.Println("GetAccountState", to, "failed", err)
		} else {
			val := util.NewUint128FromString(r.GetBalance())
			nonce, _ = strconv.ParseUint(r.Nonce, 10, 64)
			// nonce = r.Nonce
			log.Println("GetAccountState", to, "nonce", r.Nonce, "value", val)
		}
	}

	{
		stream, err := ac.Subscribe(context.Background(), &rpcpb.SubscribeRequest{})

		if err != nil {
			log.Fatalf("could not subscribe: %v", err)
		}
		for {
			reply, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("failed to recv: %v", err)
			}
			log.Println("recv notification: ", reply.MsgType, reply.Data)
		}
	}
}
