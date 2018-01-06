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
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"
	mrand "math/rand"
	"reflect"
	"strings"
	"time"

	crypto "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"
)

func (node *Node) parseAddressFromMultiaddr(address ma.Multiaddr) (ma.Multiaddr, peer.ID, error) {

	addr, err := ma.NewMultiaddr(
		strings.Split(address.String(), "/ipfs/")[0],
	)
	if err != nil {
		return nil, "", err
	}

	b58, err := address.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		return nil, "", err
	}

	id, err := peer.IDB58Decode(b58)
	if err != nil {
		return nil, "", err
	}

	return addr, id, nil

}

func (node *Node) clearPeerStore(pid peer.ID, addrs []ma.Multiaddr) {
	node.peerstore.SetAddrs(pid, addrs, 0)
	if !InArray(pid.Pretty(), node.bootIds) {
		node.routeTable.Remove(pid)
	}
}

// Write write bytes to stream
func Write(writer io.Writer, data []byte) error {
	result := make(chan error, 1)
	go func(writer io.Writer, data []byte) {
		if writer == nil {
			result <- errors.New("write data occurs error, write is nil")
			return
		}
		_, err := writer.Write(data)
		result <- err
	}(writer, data)
	err := <-result
	return err
}

// ReadBytes read bytes from a stream
func ReadBytes(reader io.Reader, n uint32) ([]byte, error) {
	data := make([]byte, n)
	result := make(chan error, 1)
	go func(reader io.Reader) {
		_, err := io.ReadFull(reader, data)
		result <- err
	}(reader)
	err := <-result
	return data, err
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

func randSeed(n int) string {
	var src = mrand.NewSource(time.Now().UnixNano())
	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = mrand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

// GenerateEd25519Key generate a privKey and pubKey by ed25519.
func GenerateEd25519Key() (crypto.PrivKey, crypto.PubKey, error) {
	randseedstr := randSeed(64)
	randseed, err := hex.DecodeString(randseedstr)
	priv, pub, err := crypto.GenerateEd25519Key(
		bytes.NewReader(randseed),
	)
	return priv, pub, err
}

func getKeypairFromFile(filename string) (crypto.PrivKey, crypto.PubKey, error) {

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, nil, ErrLoadKeypairFromFile
	}

	privb, err := base64.StdEncoding.DecodeString(string(b))
	if err != nil {
		return nil, nil, err
	}
	priv, err := crypto.UnmarshalPrivateKey(privb)
	if err != nil {
		return nil, nil, err
	}
	pub := priv.GetPublic()

	return priv, pub, nil
}
