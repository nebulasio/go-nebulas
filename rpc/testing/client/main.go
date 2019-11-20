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
	"time"

	"github.com/nebulasio/go-nebulas/crypto/keystore"

	"fmt"

	"github.com/nebulasio/go-nebulas/rpc"
	rpcpb "github.com/nebulasio/go-nebulas/rpc/pb"
	"github.com/nebulasio/go-nebulas/util"
	"golang.org/x/net/context"
)

// TODO: add command line flag.
const (
	//config = "../../../../config.pb.txt"
	from  = "n1QZMXSZtW7BUerroSms4axNfyBGyFGkrh5"
	to    = "n1Zn6iyyQRhqthmCfqGBzWfip1Wx8wEvtrJ"
	value = 2
)

// RPC testing client.
func main() {
	// Set up a connection to the server.
	//cfg := neblet.LoadConfig(config).Rpc
	addr := fmt.Sprintf("127.0.0.1:%d", uint32(8684))
	conn, err := rpc.Dial(addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ac := rpcpb.NewApiServiceClient(conn)
	adc := rpcpb.NewAdminServiceClient(conn)

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
		var val *util.Uint128
		r, err := ac.GetAccountState(context.Background(), &rpcpb.GetAccountStateRequest{Address: from})
		if err != nil {
			log.Println("GetAccountState", from, "failed", err)
		} else if val, err = util.NewUint128FromString(r.GetBalance()); err != nil {
			log.Println("GetAccountState", from, "failed to get balance", err)
		} else {
			nonce = r.Nonce
			// nonce = r.Nonce
			log.Println("GetAccountState", from, "nonce", r.Nonce, "value", val)
		}
	}

	{
		var val *util.Uint128
		r, err := ac.GetAccountState(context.Background(), &rpcpb.GetAccountStateRequest{Address: to})
		if err != nil {
			log.Println("GetAccountState", to, "failed", err)
		} else if val, err = util.NewUint128FromString(r.GetBalance()); err != nil {
			log.Println("GetAccountState", from, "failed to get balance", err)
		} else {
			// nonce = r.Nonce
			log.Println("GetAccountState", to, "nonce", r.Nonce, "value", val)
		}
	}

	//admin := rpcpb.NewAdminServiceClient(conn)

	{
		v, err := util.NewUint128FromInt(value)
		if err != nil {
			log.Println("newUint128 failed:", err)
		}

		_, err = adc.UnlockAccount(context.Background(), &rpcpb.UnlockAccountRequest{
			Address: from, Passphrase: "passphrase", Duration: uint64(keystore.DefaultUnlockDuration),
		})
		if err != nil {
			log.Println("UnlockAccount failed:", err)
		} else {
			log.Println("UnlockAccount", from)
		}

		r, err := adc.SendTransaction(context.Background(), &rpcpb.TransactionRequest{
			From: from, To: to, Value: v.String(), Nonce: nonce + 1,
			GasPrice: "2000000", GasLimit: "1000000",
		})
		if err != nil {
			log.Println("SendTransaction failed:", err)
		} else {
			log.Println("SendTransaction", from, "->", to, "value", value, r)
		}
	}

	time.Sleep(40 * time.Second)

	{
		var val *util.Uint128
		r, err := ac.GetAccountState(context.Background(), &rpcpb.GetAccountStateRequest{Address: to})
		if err != nil {
			log.Println("GetAccountState", to, "failed", err)
		} else if val, err = util.NewUint128FromString(r.GetBalance()); err != nil {
			log.Println("GetAccountState", from, "failed to get balance", err)
		} else {
			nonce = r.Nonce
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
			log.Println("recv notification: ", reply.Topic, reply.Data)
		}
	}
}
