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
	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/nf/nbre"
)

type NR struct {
	conf *nebletpb.Config
	nbre core.Nbre

	chain *core.BlockChain

	cache *lru.Cache
}

// NewNR create nr
func NewNR(neb Neblet) (*NR, error) {
	cache, err := lru.New(8)

	if err != nil {
		return nil, err
	}

	nr := &NR{
		conf:  neb.Config(),
		nbre:  neb.Nbre(),
		chain: neb.BlockChain(),
		cache: cache,
	}
	return nr, nil
}

// GetNRHandler returns the nr query handler
func (n *NR) GetNRByAddress(addr *core.Address, height uint64) (core.Data, error) {
	nrdata, err := n.getLatestNRList(height)
	if err != nil {
		return nil, err
	}
	for _, nr := range nrdata.Nrs {
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
	chainID := n.chain.ChainID()
	// for Mainnet and Testnet, delay 1 day to query nr value.
	if chainID == core.MainNetID {
		return 12 * 60 * 60 / 15
	} else if chainID == core.TestNetID {
		return 12 * 60 * 60 / 15
	} else {
		return 33
	}
}

func (n *NR) getLatestNRList(height uint64) (*NRData, error) {
	height = height - n.delayHeight()
	if height < 1 {
		return nil, ErrNRNotFound
	}
	var (
		nrdata *NRData
	)
	if data, ok := n.cache.Get(height); ok {
		nrdata = data.(*NRData)
	} else {
		data, err := n.nbre.Execute(nbre.CommandNRListByHeight, height)
		if err != nil {
			return nil, err
		}
		nrdata = data.(*NRData)
		n.cache.Add(height, nrdata)
	}
	return nrdata, nil
}

// GetNRHandler returns the nr query handler
func (n *NR) GetNRHandler(start, end, version uint64) (string, error) {
	if start < n.conf.Nbre.StartHeight {
		return "", ErrInvalidStartHeight
	}
	if start >= end {
		return "", ErrInvalidHeightInterval
	}
	if end <= 0 || end > n.chain.TailBlock().Height() {
		return "", ErrInvalidEndHeight
	}
	data, err := n.nbre.Execute(nbre.CommandNRHandler, start, end, version)
	if err != nil {
		return "", err
	}
	return data.(string), nil
}

// GetNRList returns the nr list
func (n *NR) GetNRList(hash []byte) (core.Data, error) {
	data, err := n.nbre.Execute(nbre.CommandNRListByHandle, string(hash))
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
