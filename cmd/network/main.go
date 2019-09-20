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
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/metrics"
	"github.com/nebulasio/go-nebulas/neblet"
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
)

// Message Type
var (
	PingMessage = "ping"
	PongMessage = "pong"
	messageCh   = make(chan net.Message, 4096)

	cpuprofile      = flag.String("cpuprofile", "", "write cpu profile `file`")
	memprofile      = flag.String("memprofile", "", "write memory profile to `file`")
	packageSize     = flag.Int64("package_size", 0, "package size, default is 0")
	concurrentCount = flag.Int64("concurrent", 0, "concurrent count, default is 0")
	limitCount      = flag.Int64("limit", 0, "limits of sent message, default is 0, no limit")
)

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	// fmt.Printf("%v", flag.Args())
	if len(flag.Args()) < 2 {
		help()
		return
	}

	// rand.
	rand.Seed(time.Now().UnixNano())

	// mode
	mode := flag.Args()[0]
	configPath := flag.Args()[1]

	run(mode, configPath, *packageSize, *concurrentCount, *limitCount)

	fmt.Printf("done...")

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
		f.Close()
	}
}

func help() {
	fmt.Printf("%s [server|client] [config] [package size]\n", os.Args[0])
	os.Exit(1)
}

func run(mode, configPath string, packageSize, concurrentMessageCount, totalMessageCount int64) {
	// config.
	config := neblet.LoadConfig(configPath)

	// init log.
	logging.Init(config.App.LogFile, config.App.LogLevel, config.App.LogAge)

	core.SetCompatibilityOptions(config.Chain.ChainId)

	// neblet.
	neblet, _ := neblet.New(config)
	netService, err := net.NewNebService(neblet)

	if err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(1)
	}

	// register dispatcher.
	netService.Register(net.NewSubscriber(netService, messageCh, false, PingMessage, net.MessageWeightZero))
	netService.Register(net.NewSubscriber(netService, messageCh, false, PongMessage, net.MessageWeightZero))

	// start server.
	netService.Start()

	// metrics.
	tps := metrics.NewMeter("tps")
	throughput := metrics.NewMeter("throughput")
	latency := metrics.NewHistogramWithUniformSample("latency", 100)

	sentMessageCount := int64(0)

	// first trigger.
	if mode == "client" {
		sentMessageCount += concurrentMessageCount
		time.Sleep(10 * time.Second)
		go func() {
			for i := 0; i < int(concurrentMessageCount); i++ {
				netService.SendMessageToPeers(PingMessage, GenerateData(packageSize), net.MessagePriorityNormal, new(net.ChainSyncPeersFilter))
			}
		}()
	}

	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case message := <-messageCh:
			messageName := message.MessageType()
			switch messageName {
			case PingMessage:
				data := message.Data()
				sendAt := ParseData(data)
				nowAt := time.Now().UnixNano()

				latencyVal := (nowAt - sendAt) / int64(1000000)

				// metrics.
				tps.Mark(1)
				throughput.Mark(1 * int64(net.NebMessageHeaderLength+len(data)))
				latency.Update(latencyVal)

				netService.SendMessageToPeer(PongMessage, message.Data(), net.MessagePriorityNormal, message.MessageFrom())
			case PongMessage:
				data := message.Data()
				sendAt := ParseData(data)
				nowAt := time.Now().UnixNano()
				latencyVal := (nowAt - sendAt) / int64(1000000)

				// metrics.
				tps.Mark(1)
				throughput.Mark(1 * int64(net.NebMessageHeaderLength+len(data)))
				latency.Update(latencyVal)

				sentMessageCount++
				if totalMessageCount > 0 && sentMessageCount >= totalMessageCount {
					return
				}

				netService.SendMessageToPeer(PingMessage, GenerateData(packageSize), net.MessagePriorityNormal, message.MessageFrom())
			}
		case <-ticker.C:
			fmt.Printf("[Perf] tps: %6.2f/s; throughput: %6.2fk/s; latency p95: %6.2f\n", tps.Rate1(), throughput.Rate1()/1000, latency.Percentile(float64(0.50)))
		}
	}
}

// ParseData parse int64 from bytes
func ParseData(data []byte) int64 {
	return byteutils.Int64(data)
}

// GenerateData convert int64 into bytes
func GenerateData(packageSize int64) []byte {
	data := make([]byte, 8+packageSize)
	copy(data, byteutils.FromInt64(time.Now().UnixNano()))
	return data
}
