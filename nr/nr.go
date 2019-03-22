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

package nr

import (
	"errors"

	"github.com/hashicorp/golang-lru"

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/nf/nbre"
)

type NR struct {
	neb Neblet

	cache *lru.Cache
}

// NewNR create nr
func NewNR(neb Neblet) (*NR, error) {
	cache, err := lru.New(8)
	if err != nil {
		return nil, err
	}

	nr := &NR{
		neb:   neb,
		cache: cache,
	}
	return nr, nil
}

// GetNRHandler returns the nr query handler
func (n *NR) GetNRByAddress(addr *core.Address) (core.Data, error) {

	height := n.neb.BlockChain().TailBlock().Height()
	data, err := n.GetNRListByHeight(height)
	if err != nil {
		return nil, err
	}
	nrData := data.(*NRData)
	for _, nr := range nrData.Nrs {
		if nr.Address == addr.String() {
			return nr, nil
		}
	}
	return nil, ErrNRNotFound
}

// Since the calculation of NR takes time,
// in order to ensure that all machines can get consistent NR value,
// it is necessary to add a buffer height when querying the NR value,
// and query the NR value corresponding to the block height
func (n *NR) delayHeight() uint64 {
	chainID := n.neb.BlockChain().ChainID()
	// for Mainnet and Testnet, delay 1 day to query nr value.
	if chainID == core.MainNetID {
		return 24 * 60 * 60 / 15
	} else if chainID == core.TestNetID {
		return 24 * 60 * 60 / 15
	} else {
		return 33
	}
}

// GetNRListByHeight return nr list, which subtract the deplay height, ensure all node is equal.
func (n *NR) GetNRListByHeight(height uint64) (core.Data, error) {
	var (
		nrData *NRData
	)

	height = height - n.delayHeight()
	if height < 1 {
		return nil, ErrNRNotFound
	}

	if data, ok := n.cache.Get(height); ok {
		nrData = data.(*NRData)
	} else {
		data, err := n.neb.Nbre().Execute(nbre.CommandNRListByHeight, height)
		if err != nil {
			return nil, err
		}
		nrData = &NRData{}
		if err := nrData.FromBytes([]byte(data.(string))); err != nil {
			return nil, err
		}
		if len(nrData.Err) > 0 {
			return nil, errors.New(nrData.Err)
		}
		n.cache.Add(height, nrData)
	}
	return nrData, nil
}

// GetNRHandler returns the nr query handler
func (n *NR) GetNRHandle(start, end, version uint64) (string, error) {
	if start < n.neb.Config().Nbre.StartHeight {
		return "", ErrInvalidStartHeight
	}
	if start >= end {
		return "", ErrInvalidHeightInterval
	}
	if end <= 0 || end > n.neb.BlockChain().TailBlock().Height() {
		return "", ErrInvalidEndHeight
	}
	data, err := n.neb.Nbre().Execute(nbre.CommandNRHandler, start, end, version)
	if err != nil {
		return "", err
	}
	return data.(string), nil
}

// GetNRList returns the nr list
func (n *NR) GetNRListByHandle(handle []byte) (core.Data, error) {
	data, err := n.neb.Nbre().Execute(nbre.CommandNRListByHandle, string(handle))
	if err != nil {
		return nil, err
	}
	nrData := &NRData{}
	if err := nrData.FromBytes([]byte(data.(string))); err != nil {
		return nil, err
	}
	if len(nrData.Err) > 0 {
		return nil, errors.New(nrData.Err)
	}
	return nrData, nil
}
