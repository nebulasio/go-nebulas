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
	lru "github.com/hashicorp/golang-lru"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"

	"github.com/nebulasio/go-nebulas/core"
)

type NR struct {
	neb Neblet

	nrCache  *lru.Cache
	sumCache *lru.Cache
}

// NewNR create nr
func NewNR(neb Neblet) (*NR, error) {
	nrCache, err := lru.New(CacheSize)
	if err != nil {
		return nil, err
	}
	sumCache, err := lru.New(CacheSize)
	if err != nil {
		return nil, err
	}

	nr := &NR{
		neb:      neb,
		nrCache:  nrCache,
		sumCache: sumCache,
	}
	return nr, nil
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
	if core.NbreSplitAtHeight(height) {
		return nil, core.ErrFuncDeprecated
	}

	height = height - n.delayHeight()
	if height < n.neb.Config().Nbre.StartHeight {
		return nil, ErrNRNotFound
	}

	nr, err := n.getNRListCache(height)
	if err != nil {
		return nil, err
	}
	return nr, nil
}

// GetNRSummary return nr summary info, which subtract the deplay height, ensure all node is equal.
func (n *NR) GetNRSummary(height uint64) (core.Data, error) {
	if core.NbreSplitAtHeight(height) {
		return nil, core.ErrFuncDeprecated
	}

	height = height - n.delayHeight()
	if height < n.neb.Config().Nbre.StartHeight {
		return nil, ErrNRSummaryNotFound
	}
	sum, err := n.getNRSummaryCache(height)
	if err != nil {
		return nil, err
	}
	return sum, nil
}

// getNRListCache
func (n *NR) getNRListCache(height uint64) (*NRData, error) {
	if n.nrCache.Len() == 0 {
		n.loadNRCache()
	}

	key := (height-core.NrStartHeight)/core.NrIntervalHeight - 1
	if data, ok := n.nrCache.Get(key); ok {
		nrData := data.(*NRData)
		logging.VLog().WithFields(logrus.Fields{
			"height":   height,
			"start":    nrData.StartHeight,
			"end":      nrData.EndHeight,
			"dataSize": len(nrData.Nrs),
		}).Debug("Success to find nr list in cache.")
		return nrData, nil
	}
	return nil, ErrNRNotFound
}

// getNRSummaryCache
func (n *NR) getNRSummaryCache(height uint64) (*NRSummary, error) {
	if n.sumCache.Len() == 0 {
		n.loadSumCache()
	}

	key := (height-core.NrStartHeight)/core.NrIntervalHeight - 1
	if data, ok := n.sumCache.Get(key); ok {
		sumData := data.(*NRSummary)
		logging.VLog().WithFields(logrus.Fields{
			"height": height,
			"start":  sumData.StartHeight,
			"end":    sumData.EndHeight,
			"data":   sumData.Sum,
		}).Debug("Success to find nr summary in cache.")
		return sumData, nil
	}
	return nil, ErrNRSummaryNotFound
}
