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
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/nebulasio/go-nebulas/rpc"
	"github.com/nebulasio/go-nebulas/rpc/pb"
	"github.com/nebulasio/go-nebulas/util"
	"golang.org/x/net/context"
)

// TODO: add command line flag.
const (
	from      = "1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c"
	to        = "fbcef590704577fa307198bf6edc6272df451610b5744bfe"
	value     = 2
	sendtimes = 10000
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

	admin := rpcpb.NewAdminServiceClient(conn)
	api := rpcpb.NewApiServiceClient(conn)
	var nonce uint64
	var lastnonce uint64

	{
		r, err := api.GetAccountState(context.Background(), &rpcpb.GetAccountStateRequest{Address: from})
		if err != nil {
			log.Println("GetAccountState", from, "failed", err)
		} else {
			val := util.NewUint128FromString(r.GetBalance())
			nonce, _ = strconv.ParseUint(r.Nonce, 10, 64)
			lastnonce = nonce
			log.Println("GetAccountState", from, "nonce", r.Nonce, "value", val)
		}
	}

	{
		_, err := admin.UnlockAccount(context.Background(), &rpcpb.UnlockAccountRequest{Address: from, Passphrase: "passphrase"})
		if err != nil {
			log.Println("Unlock failed:", err)
		} else {
			log.Println("Unlock")
		}
	}

	start := time.Now().Unix()
	nonce++
	{
		v := util.NewUint128FromInt(value)
		for i := 0; i < sendtimes; i++ {
			_, err := api.SendTransaction(context.Background(), &rpcpb.TransactionRequest{From: from, To: to, Value: v.String(), Nonce: nonce})
			if err != nil {
				log.Println("SendTransaction failed:", err)
			}
			nonce++
		}
	}
	log.Println("SendTransaction ", sendtimes)

	{
		for true {
			a, err := api.GetAccountState(context.Background(), &rpcpb.GetAccountStateRequest{Address: from})
			if err != nil {
				log.Println("GetAccountState", from, "failed", err)
			} else {
				val := util.NewUint128FromString(a.GetBalance())
				nonce, _ := strconv.ParseUint(a.Nonce, 10, 64)
				if nonce == lastnonce+sendtimes {
					log.Println("Tps: ", sendtimes/(time.Now().Unix()-start))
					return
				}
				log.Println("GetAccountState", from, "nonce", nonce, "value", val, "Unix", time.Now().Unix(), "Start", start)
			}
			time.Sleep(1 * time.Second)
		}
	}
}
