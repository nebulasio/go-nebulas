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
	"errors"
	"net"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	peer "github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"
)

var (
	ErrListenPortIsNotAvailable = errors.New("listen port is not available")
)

func ParseFromIPFSAddr(ipfsAddr ma.Multiaddr) (peer.ID, ma.Multiaddr, error) {
	addr, err := ma.NewMultiaddr(strings.Split(ipfsAddr.String(), "/ipfs/")[0])
	if err != nil {
		return "", nil, err
	}

	// TODO: @robin we should register neb multicodecs.
	b58, err := ipfsAddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		return "", nil, err
	}

	id, err := peer.IDB58Decode(b58)
	if err != nil {
		return "", nil, err
	}

	return id, addr, nil
}

// InArray returns whether an object exists in an array.
func InArray(obj interface{}, array interface{}) bool {
	arrayValue := reflect.ValueOf(array)
	if reflect.TypeOf(array).Kind() == reflect.Array || reflect.TypeOf(array).Kind() == reflect.Slice {
		for i := 0; i < arrayValue.Len(); i++ {
			if arrayValue.Index(i).Interface() == obj {
				return true
			}
		}
	}
	return false
}

//GetCountOfMap return the count of a map
func GetCountOfMap(m *sync.Map) uint32 {
	length := 0
	m.Range(func(_, _ interface{}) bool {
		length++
		return true
	})
	return uint32(length)
}

func verifyListenAddress(listen []string) error {
	for _, v := range listen {
		_, err := net.ResolveTCPAddr("tcp", v)
		if err != nil {
			return err
		}
	}
	return nil
}

func checkPathConfig(path string) bool {
	if path == "" {
		return true
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

func checkPortAvailable(listen []string) error {
	for _, v := range listen {
		conn, err := net.DialTimeout("tcp", v, time.Second*1)
		if err == nil {
			conn.Close()
			return ErrListenPortIsNotAvailable
		}
	}
	return nil
}
