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

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/nf/nbre"
)

type NR struct {
	conf *nebletpb.Config
	nbre core.Nbre

	chain *core.BlockChain
}

// NewNR create nr
func NewNR(neb Neblet) *NR {
	nr := &NR{
		conf:  neb.Config(),
		nbre:  neb.Nbre(),
		chain: neb.BlockChain(),
	}
	return nr
}

// GetNRHandler returns the nr query handler
func (n *NR) GetNRByAddress(addr *core.Address) (core.Data, error) {
	data, err := n.GetNRList(nil)
	if err != nil {
		return nil, err
	}
	nrdata := data.(*NRData)
	for _, nr := range nrdata.Nrs {
		if nr.Address == addr.String() {
			return nr, nil
		}
	}
	return nil, ErrNRNotFound
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
	data, err := n.nbre.Execute(nbre.CommandNRList, string(hash))
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
