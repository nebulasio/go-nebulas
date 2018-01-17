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
	"os"
	"strconv"
	"time"

	"github.com/nebulasio/go-nebulas/util/byteutils"

	"github.com/nebulasio/go-nebulas/neblet"
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/net/p2p"

	"github.com/nebulasio/go-nebulas/util/logging"
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

	// mode
	mode := os.Args[1]
	configPath := os.Args[2]
	delayMs := int64(0)

	if len(os.Args) >= 4 {
		delayMs, _ = strconv.ParseInt(os.Args[3], 10, 64)
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

	// first trigger.
	if mode == "client" {
		go func() {
			time.Sleep(10 * time.Second)
			netService.SendMessageToPeers(PingMessage, GenerateTime(), net.MessagePriorityNormal, new(p2p.ChainSyncPeersFilter))
		}()
	}

	for {
		select {
		case message := <-messageCh:
			messageName := message.MessageType()
			switch messageName {
			case PingMessage:
				netService.SendMessageToPeer(PongMessage, message.Data().([]byte), net.MessagePriorityNormal, message.MessageFrom())
			case PongMessage:
				sendAt := ParseTime(message.Data().([]byte))
				nowAt := time.Now().UnixNano()
				logging.CLog().Infof("Duration(ms): |%9d|, from %d to %d", (nowAt-sendAt)/int64(1000000), sendAt, nowAt)

				if delayMs > 0 {
					time.Sleep(time.Millisecond * time.Duration(delayMs))
				}

				netService.SendMessageToPeers(PingMessage, GenerateTime(), net.MessagePriorityNormal, new(p2p.ChainSyncPeersFilter))
			}
		}
	}
}

func ParseTime(data []byte) int64 {
	return byteutils.Int64(data)
}

func GenerateTime() []byte {
	return byteutils.FromInt64(time.Now().UnixNano())
}

func help() {
	fmt.Printf("%s [server|client] [config] [(bytes)]\n", os.Args[0])
	os.Exit(1)
}
