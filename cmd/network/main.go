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
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/nebulasio/go-nebulas/util/byteutils"

	"github.com/nebulasio/go-nebulas/neblet"
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/net/p2p"

	"github.com/nebulasio/go-nebulas/util/logging"
	metrics "github.com/rcrowley/go-metrics"
)

var (
	PingMessage = "ping"
	PongMessage = "pong"
	messageCh   = make(chan net.Message, 4096)
)

func main() {

	if len(os.Args) < 3 {
		help()
		return
	}

	// rand.
	rand.Seed(time.Now().UnixNano())

	// mode
	mode := os.Args[1]
	configPath := os.Args[2]
	packageSize := int64(0)

	if len(os.Args) >= 4 {
		packageSize, _ = strconv.ParseInt(os.Args[3], 10, 64)
	}

	// config.
	config := neblet.LoadConfig(configPath)

	// init log.
	logging.Init(config.App.LogFile, config.App.LogLevel, config.App.LogAge)

	// neblet.
	neblet, _ := neblet.New(config)
	netService, err := p2p.NewNetService(neblet)

	if err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(1)
	}

	// register dispatcher.
	netService.Register(net.NewSubscriber(netService, messageCh, PingMessage))
	netService.Register(net.NewSubscriber(netService, messageCh, PongMessage))

	// start server.
	netService.Start()

	// metrics.
	tps := metrics.GetOrRegisterMeter("tps", nil)
	throughput := metrics.GetOrRegisterMeter("throughput", nil)
	latency := metrics.GetOrRegisterHistogram("latency", nil, metrics.NewUniformSample(100))

	// first trigger.
	if mode == "client" {
		go func() {
			time.Sleep(10 * time.Second)
			netService.SendMessageToPeers(PingMessage, GenerateData(packageSize), net.MessagePriorityNormal, new(p2p.ChainSyncPeersFilter))
		}()
	}

	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case message := <-messageCh:
			messageName := message.MessageType()
			switch messageName {
			case PingMessage:
				data := message.Data().([]byte)
				sendAt := ParseData(data)
				nowAt := time.Now().UnixNano()

				latencyVal := (nowAt - sendAt) / int64(1000000)

				// metrics.
				tps.Mark(2)
				throughput.Mark(2 * int64(p2p.NebMessageHeaderLength+len(data)))
				latency.Update(latencyVal)

				netService.SendMessageToPeer(PongMessage, message.Data().([]byte), net.MessagePriorityNormal, message.MessageFrom())
			case PongMessage:
				data := message.Data().([]byte)

				sendAt := ParseData(data)
				nowAt := time.Now().UnixNano()
				latencyVal := (nowAt - sendAt) / int64(1000000)

				// metrics.
				tps.Mark(2)
				throughput.Mark(2 * int64(p2p.NebMessageHeaderLength+len(data)))
				latency.Update(latencyVal)

				// if latencyVal > 10 {
				// 	logging.CLog().Infof("Duration(ms): |%9d|, from %d to %d", (nowAt-sendAt)/int64(1000000), sendAt, nowAt)
				// }

				netService.SendMessageToPeer(PingMessage, GenerateData(packageSize), net.MessagePriorityNormal, message.MessageFrom())
			}
		case <-ticker.C:
			fmt.Printf("[Perf] tps: %6.2f/s; throughput: %6.2fk/s; latency p95: %6.2f\n", tps.Rate1(), throughput.Rate1()/1000, latency.Percentile(float64(0.50)))
		}
	}
}

func ParseData(data []byte) int64 {
	return byteutils.Int64(data)
}

func GenerateData(packageSize int64) []byte {
	data := make([]byte, 8+packageSize)
	copy(data, byteutils.FromInt64(time.Now().UnixNano()))
	return data
}

func help() {
	fmt.Printf("%s [server|client] [config] [package size]\n", os.Args[0])
	os.Exit(1)
}
