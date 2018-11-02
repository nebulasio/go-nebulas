// Copyright (C) 2018 go-nebulas authors
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
	"context"
	"crypto/rand"
	"io"
	"log"
	"time"

	"github.com/nebulasio/go-nebulas/nbre/benchmark/ipc/grpc/pb"
	"google.golang.org/grpc"
)

func main() {
	conn := DialServer()
	defer conn.Close()

	client := ipcpb.NewBenchmarkServiceClient(conn)

	dataSize := 1024 * 1024
	count := 100
	data := RandomCSPRNG(dataSize)
	timestamp := time.Now().UnixNano()
	for i := 0; i < count; i++ {
		_, err := client.Transfer(context.Background(), &ipcpb.Benchmark{
			//Timestamp:time.Now().UnixNano(),
			Data: data,
		})
		if err != nil {
			log.Println("Transfer failed:", err)
			break
		}
	}

	speed := int64(dataSize*count*2*1e9/1000) / (time.Now().UnixNano() - timestamp)
	log.Println("grpc data size:", dataSize/1024, "kb speed:", speed, "kb/s")
}

func DialServer() *grpc.ClientConn {
	maxSize := 64 * 1024 * 1024
	conn, err := grpc.Dial("127.0.0.1:8696", grpc.WithInsecure(), grpc.WithMaxMsgSize(maxSize))
	if err != nil {
		log.Println("rpc.Dial() failed: ", err)
	}
	if err != nil {
		log.Fatal(err)
	}
	return conn
}

// RandomCSPRNG a cryptographically secure pseudo-random number generator
func RandomCSPRNG(n int) []byte {
	buff := make([]byte, n)
	_, err := io.ReadFull(rand.Reader, buff)
	if err != nil {
		panic("reading from crypto/rand failed: " + err.Error())
	}
	return buff
}
