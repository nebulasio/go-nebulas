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
	"log"
	/* 	"strconv" */

	"fmt"

	"strconv"

	"time"

	"github.com/nebulasio/go-nebulas/rpc"
	"github.com/nebulasio/go-nebulas/rpc/pb"
	"github.com/nebulasio/go-nebulas/util"
	"golang.org/x/net/context"
)

// TODO: add command line flag.
const (
	from  = "333cb3ed8c417971845382ede3cf67a0a96270c05fe2f700"
	to    = "fbcef590704577fa307198bf6edc6272df451610b5744bfe"
	value = 0
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

	ac := rpcpb.NewAPIServiceClient(conn)
	start := time.Now().Unix()
	{
		_, err := ac.UnlockAccount(context.Background(), &rpcpb.UnlockAccountRequest{Address: from, Passphrase: "passphrase"})
		if err != nil {
			log.Println("Unlock failed:", err)
		} else {
			log.Println("Unlock")
		}
	}

	{
		v := util.NewUint128FromInt(value)
		nonce := uint64(1)
		for i := 0; i < 10000; i++ {
			r, err := ac.SendTransaction(context.Background(), &rpcpb.SendTransactionRequest{From: from, To: to, Value: v.String(), Nonce: nonce})
			if err != nil {
				log.Println("SendTransaction failed:", err)
			} else {
				log.Println("SendTransaction", from, "->", to, "value", value, r)
			}
			nonce++
		}
	}

	{
		for true {
			a, err := ac.GetAccountState(context.Background(), &rpcpb.GetAccountStateRequest{Address: from})
			if err != nil {
				log.Println("GetAccountState", from, "failed", err)
			} else {
				val := util.NewUint128FromString(a.GetBalance())
				nonce, _ := strconv.ParseUint(a.Nonce, 10, 64)
				log.Println("GetAccountState", from, "nonce", nonce, "value", val, "Unix", time.Now().Unix(), "Start", start)
			}
			time.Sleep(1 * time.Second)
		}
	}
}
