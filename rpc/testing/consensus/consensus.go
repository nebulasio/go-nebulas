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

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/rpc"
	"github.com/nebulasio/go-nebulas/rpc/pb"
	"golang.org/x/net/context"
)

// TODO: add command line flag.
const (
	from  = "48f981ed38910f1232c1bab124f650c482a57271632db9e3"
	to    = "48f981ed38910f1232c1bab124f650c482a57271632db9e3"
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

	admin := rpcpb.NewAdminServiceClient(conn)
	api := rpcpb.NewApiServiceClient(conn)
	{
		_, err := admin.UnlockAccount(context.Background(), &rpcpb.UnlockAccountRequest{Address: from, Passphrase: "passphrase"})
		if err != nil {
			log.Println("Unlock failed:", err)
		} else {
			log.Println("Unlock")
		}
	}

	{
		resp, err := admin.SignTransaction(context.Background(), &rpcpb.TransactionRequest{
			From:      from,
			To:        from,
			Value:     "0",
			Nonce:     1,
			Candidate: &rpcpb.CandidateRequest{Action: core.LogoutAction},
			GasLimit:  "400000",
		})
		r, err := api.SendRawTransaction(context.Background(), &rpcpb.SendRawTransactionRequest{Data: resp.Data})
		if err != nil {
			log.Println("SendRawTransaction failed:", err)
		} else {
			log.Println("SendRawTransaction", from, "->", to, "value", value, r)
		}
	}
}
