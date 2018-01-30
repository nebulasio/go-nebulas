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

package net

import (
	"fmt"
	"net"
	"time"

	"github.com/multiformats/go-multiaddr"
	"github.com/nebulasio/go-nebulas/neblet/pb"
)

// const
const (
	DefaultBucketCapacity         = 64
	DefaultRoutingTableMaxLatency = 10
	DefaultPrivateKeyPath         = "conf/network.key"
	DefaultMaxSyncNodes           = 64
	DefaultChainID                = 1
	DefaultVersion                = 0
	DefaultRelayCacheSize         = 65536
	DefaultStreamStoreSize        = 128
	DefaultStreamStoreExtendSize  = 32
	DefaultNetworkID              = 1
	DefaultRoutingTableDir        = ""
)

// Default Configuration in P2P network
var (
	DefaultListen = []string{"0.0.0.0:8680"}

	RouteTableSyncLoopInterval   = 30 * time.Second
	RouteTableSaveToDiskInterval = 3 * 60 * time.Second
	RouteTableCacheFileName      = "routetable.cache"

	MaxPeersCountForSyncResp = 32
)

// Config TODO: move to proto config.
type Config struct {
	Bucketsize            int
	Latency               time.Duration
	BootNodes             []multiaddr.Multiaddr
	PrivateKeyPath        string
	Listen                []string
	MaxSyncNodes          int
	ChainID               uint32
	Version               uint8
	RelayCacheSize        int
	StreamStoreSize       int
	StreamStoreExtendSize int
	NetworkID             uint32
	RoutingTableDir       string
}

// Neblet interface breaks cycle import dependency.
type Neblet interface {
	Config() *nebletpb.Config
}

// NewP2PConfig return new config object.
func NewP2PConfig(n Neblet) *Config {
	chainConf := n.Config().Chain
	networkConf := n.Config().Network
	config := NewConfigFromDefaults()

	// listen.
	if len(networkConf.Listen) == 0 {
		panic("Missing network.listen config.")
	}
	if err := verifyListenAddress(networkConf.Listen); err != nil {
		panic(fmt.Sprintf("Invalid network.listen config: err is %s, config value is %s.", err, networkConf.Listen))
	}
	config.Listen = networkConf.Listen

	// private key path.
	if checkPathConfig(networkConf.PrivateKey) == false {
		panic(fmt.Sprintf("The network private key path %s is not exist.", networkConf.PrivateKey))
	}
	config.PrivateKeyPath = networkConf.PrivateKey

	// Chain ID.
	config.ChainID = chainConf.ChainId

	// TODO: @robin set networkid when --debug.
	config.NetworkID = networkConf.NetworkId

	// routing table dir.
	// TODO: @robin using diff dir for temp files.
	if checkPathConfig(chainConf.Datadir) == false {
		panic(fmt.Sprintf("The chain data directory %s is not exist.", chainConf.Datadir))
	}
	config.RoutingTableDir = chainConf.Datadir

	// seed server address.
	seeds := networkConf.Seed
	if len(seeds) > 0 {
		config.BootNodes = make([]multiaddr.Multiaddr, len(seeds))
		for i, v := range seeds {
			addr, err := multiaddr.NewMultiaddr(v)
			if err != nil {
				panic(fmt.Sprintf("Invalid seed address config: err is %s, config value is %s.", err, v))
			}
			config.BootNodes[i] = addr
		}
	}

	return config
}

func localHost() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return ""
}

// NewConfigFromDefaults return new config from defaults.
func NewConfigFromDefaults() *Config {
	return &Config{
		DefaultBucketCapacity,
		DefaultRoutingTableMaxLatency,
		[]multiaddr.Multiaddr{},
		DefaultPrivateKeyPath,
		DefaultListen,
		DefaultMaxSyncNodes,
		DefaultChainID,
		DefaultVersion,
		DefaultRelayCacheSize,
		DefaultStreamStoreSize,
		DefaultStreamStoreExtendSize,
		DefaultNetworkID,
		DefaultRoutingTableDir,
	}
}
