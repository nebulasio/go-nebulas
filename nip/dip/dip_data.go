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

package dip

import (
	"encoding/json"

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/nip/dip/dipdata"
)

var (
	dipData = dipdata.MustAsset("dip_data.json")
)

// rewardData
func (d *Dip) loadCache() error {
	var (
		dips []*DIPData
	)
	if d.neb.BlockChain().ChainID() == core.MainNetID {
		err := json.Unmarshal(dipData, &dips)
		if err != nil {
			return err
		}
	}
	for _, item := range dips {
		key := (item.StartHeight - core.NrStartHeight) / core.NrIntervalHeight
		d.cache.Add(uint64(key), item)
	}
	return nil
}
