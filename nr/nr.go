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
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/nf/nbre"
)

type NR struct {
	nbre core.Nbre
	// TODO(larry): add cache for nr calculate results
}

// NewNR create nr
func NewNR(neb core.Neblet) *NR {
	nr := &NR{
		nbre: neb.Nbre(),
	}
	return nr
}

// GetNRHash returns the nr query hash
func (n *NR) GetNRHash(start, end uint64) ([]byte, error) {
	return nil, nil
}

// GetNRList returns the nr list
func (n *NR) GetNRList(hash []byte) (core.Data, error) {
	//TODO(larry): give a test params
	params := &nbre.Params{
		StartBlock:10000,
		EndBlock:11000,
		Version:4294967296,
	}
	pBytes, err := params.ToBytes()
	if err != nil {
		return nil, err
	}
	data, err := n.nbre.Execute(nbre.CommandNRList, pBytes)
	if err != nil {
		return nil, err
	}
	nrData := &NRData{}
	if err := nrData.FromBytes(data); err != nil {
		return nil, err
	}
	return nrData, nil
}