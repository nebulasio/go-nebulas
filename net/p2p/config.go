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

package p2p

import (
	"net"
	"time"

	"github.com/multiformats/go-multiaddr"
	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/util/logging"
)

// const
const (
	DefaultBucketsize = 64
	DefaultLatency    = 10
	// DefaultIP         = "127.0.0.1"
	DefaultPrivateKey = ""
	// DefaultRandseed              = 12345
	// DefaultPort                  = 9999
	DefaultMaxSyncNodes          = 16
	DefaultChainID               = 1
	DefaultVersion               = 0
	DefaultRelayCacheSize        = 65536
	DefaultStreamStoreSize       = 128
	DefaultStreamStoreExtendSize = 32
	DefaultNetworkID             = 1
)

// Config TODO: move to proto config.
type Config struct {
	Bucketsize int
	Latency    time.Duration
	BootNodes  []multiaddr.Multiaddr
	PrivateKey string
	// IP                    string
	// Port                  uint32
	Listen                []string
	MaxSyncNodes          int
	ChainID               uint32
	Version               uint8
	RelayCacheSize        int
	StreamStoreSize       int
	StreamStoreExtendSize int
	NetworkID             uint32
}

// Neblet interface breaks cycle import dependency.
type Neblet interface {
	Config() nebletpb.Config
}

// NewP2PConfig new p2p network config
func NewP2PConfig(n Neblet) *Config {
	config := DefautConfig()
	config.Listen = n.Config().Network.Listen

	seeds := n.Config().Network.Seed
	if len(seeds) > 0 {
		config.BootNodes = []multiaddr.Multiaddr{}
		for _, v := range seeds {
			seed, err := multiaddr.NewMultiaddr(v)
			if err != nil {
				logging.VLog().Error("param seed error, creating seed node fail", err)
				return nil
			}
			config.BootNodes = append(config.BootNodes, seed)
		}
	}

	config.PrivateKey = n.Config().Network.PrivateKey

	if chainID := n.Config().Chain.ChainId; chainID > 0 {
		config.ChainID = chainID
	}

	if networkID := n.Config().Network.NetworkId; networkID > 0 {
		config.NetworkID = networkID
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

// DefautConfig defautConfig is the p2p network defaut config
func DefautConfig() *Config {
	defaultListen := []string{"127.0.0.1:9999"}
	return &Config{
		DefaultBucketsize,
		DefaultLatency,
		[]multiaddr.Multiaddr{},
		DefaultPrivateKey,
		defaultListen,
		DefaultMaxSyncNodes,
		DefaultChainID,
		DefaultVersion,
		DefaultRelayCacheSize,
		DefaultStreamStoreSize,
		DefaultStreamStoreExtendSize,
		DefaultNetworkID,
	}
}
