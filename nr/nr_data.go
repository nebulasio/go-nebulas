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

package nr

import (
	"encoding/json"

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/nr/nrdata"
)

var (
	nrData  = nrdata.MustAsset("nr_data.json")
	sumData = nrdata.MustAsset("sum_data.json")
)

// loadNRCache
func (d *NR) loadNRCache() error {
	var (
		nrs []*NRData
	)
	if d.neb.BlockChain().ChainID() == core.MainNetID {
		err := json.Unmarshal(nrData, &nrs)
		if err != nil {
			return err
		}
	}
	for _, item := range nrs {
		key := (item.StartHeight - core.NrStartHeight) / core.NrIntervalHeight
		d.nrCache.Add(uint64(key), item)
	}
	return nil
}

// loadSumCache
func (d *NR) loadSumCache() error {
	var (
		sums []*NRSummary
	)
	if d.neb.BlockChain().ChainID() == core.MainNetID {
		err := json.Unmarshal(sumData, &sums)
		if err != nil {
			return err
		}
	}
	for _, item := range sums {
		key := (item.StartHeight - core.NrStartHeight) / core.NrIntervalHeight
		d.sumCache.Add(uint64(key), item)
	}
	return nil
}
