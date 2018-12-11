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

package dip

import (
	"github.com/nebulasio/go-nebulas/core"
	"github.com/hashicorp/golang-lru"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util"
)

type Dip struct {

	nbre core.Nbre

	cache *lru.Cache
}

// NewDIP create a dip
func NewDIP(neb core.Neblet) (*Dip, error) {
	cache, err := lru.New(CacheSize)
	if err != nil {
		return nil, err
	}
	dip := &Dip{
		nbre: neb.Nbre(),
		cache:cache,
	}
	return dip, nil
}

// GetDipList returns dip info list
func (d *Dip) GetDipList(height uint64) (core.Data, error) {
	data, ok := d.checkCache(height)
	if !ok {
		// TODO(larry): query from nbre and cache
		key := append(byteutils.FromUint64(data.Start), byteutils.FromUint64(data.End)...)
		d.cache.Add(key, data)
	}
	return data, nil
}

func (d *Dip) checkCache(height uint64) (*DIPData, bool) {
	keys:= d.cache.Keys()
	for _, v := range keys {
		v := v.([]byte)
		start := byteutils.Uint64(v[:8])
		end := byteutils.Uint64(v[8:])
		if height >= start && height <= end {
			data, _ := d.cache.Get(v)
			return data.(*DIPData), true
		}
	}
	return nil, false
}

func (d *Dip) CheckReward(height uint64, addr string, value *util.Uint128) error {
	data, err := d.GetDipList(height)
	if err != nil {
		return err
	}
	dip := data.(*DIPData)
	for _, v := range dip.Data {
		if v.Addr == addr && value.String() == v.Value {
			return nil
		}
	}
	return ErrDipNotFound
}