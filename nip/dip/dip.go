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
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
	"time"
	"github.com/nebulasio/go-nebulas/nf/nbre"
)

type Dip struct {

	accountManager core.AccountManager
	blockchain *core.BlockChain
	nbre core.Nbre

	cache *lru.Cache

	// reward address
	address *core.Address

	quitCh            chan int
}

// NewDIP create a dip
func NewDIP(neb core.Neblet) (*Dip, error) {
	cache, err := lru.New(CacheSize)
	if err != nil {
		return nil, err
	}
	dip := &Dip{
		accountManager:neb.AccountManager(),
		blockchain:neb.BlockChain(),
		nbre: neb.Nbre(),
		cache:cache,
		quitCh:make(chan int, 1),
	}
	return dip, nil
}

func (d *Dip)RewardAddress() *core.Address {
	if d.address == nil {
		//TODO(larry): The current award address is constant and visible to all.
		priv, _ := byteutils.FromHex(DipRewardAddressPrivate)
		d.address, _ = d.accountManager.LoadPrivate(priv, []byte(DipRewardAddressPassphrase))
	}
	return d.address
}

// Start start dip.
func (d *Dip) Start() {
	logging.CLog().WithFields(logrus.Fields{
	}).Info("Starting Dip...")

	go d.loop()
}

// Stop stop dip.
func (d *Dip) Stop() {
	logging.CLog().WithFields(logrus.Fields{
	}).Info("Stopping Dip...")

	d.quitCh <- 1
}

func (d *Dip) loop() {
	logging.CLog().Info("Started Dip.")

	timerChan := time.NewTicker(time.Second * 15).C
	for {
		select {
		case <-d.quitCh:
			logging.CLog().Info("Stopped Dip.")
			return
		case <-timerChan:
			d.publishReward()
		}
	}
}

// publishReward generate dip transactions and push to tx pool
func (d *Dip) publishReward()  {
	height := d.blockchain.TailBlock().Height() - uint64(DipDelayRewardHeight)
	if height < 1 {
		return
	}
	data, err := d.GetDipList(height)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Debug("Failed to get dip reward list.")
		return
	}

	dipData := data.(*DIPData)
	endBlock := d.blockchain.GetBlockOnCanonicalChainByHeight(dipData.End)
	endAccount, err := endBlock.GetAccount(d.RewardAddress().Bytes())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Warn("Failed to load reward account in dip end account.")
		return
	}

	tailAccount, err := d.blockchain.TailBlock().GetAccount(d.RewardAddress().Bytes())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Warn("Failed to load reward account in tail block.")
		return
	}

	for idx, v := range dipData.Dips {
		nonce := endAccount.Nonce()+uint64(idx)+1
		// only not reward tx can be pushed to tx pool.
		if nonce > tailAccount.Nonce() {
			tx, err := d.generateRewardTx(v, nonce, endBlock)
			if err != nil {
				logging.VLog().WithFields(logrus.Fields{
					"err": err,
				}).Warn("Failed to generate reward transaction.")
				continue
			}

			d.blockchain.TransactionPool().Push(tx)
		}
	}
}

func (d *Dip) generateRewardTx(item *DIPItem, nonce uint64, block *core.Block) (*core.Transaction, error) {
	var (
		to *core.Address
		value *util.Uint128
		payload []byte
		err error
	)
	if to, err = core.AddressParse(item.Address); err != nil {
		return nil, err
	}
	if value, err = util.NewUint128FromString(item.Reward); err != nil {
		return nil, err
	}
	if payload, err = core.NewDipPayload(nil).ToBytes(); err != nil {
		return nil, err
	}

	tx, err := core.NewTransaction(
		block.ChainID(),
		d.RewardAddress(),
		to,
		value,
		nonce,
		core.TxPayloadDipType,
		payload,
		core.TransactionGasPrice,
		core.MinGasCountPerTransaction)
	if err != nil {
		return nil, err
	}
	// update all reward timestamp to last calculated block timestamp, generate hash equal every loop.
	tx.SetTimestamp(block.Timestamp())
	if err = d.accountManager.SignTransactionWithPassphrase(d.RewardAddress(), tx, []byte(DipRewardAddressPassphrase)); err != nil {
		return nil, err
	}
	return tx, nil
}

// GetDipList returns dip info list
func (d *Dip) GetDipList(height uint64) (core.Data, error) {
	data, ok := d.checkCache(height)
	if !ok {
		dipData, err := d.nbre.Execute(nbre.CommandDIPList, byteutils.FromUint64(height))
		if err != nil {
			return nil, err
		}
		data = &DIPData{}
		if err := data.FromBytes(dipData); err != nil {
			return nil, err
		}
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
	for _, v := range dip.Dips {
		if v.Address == addr && value.String() == v.Reward {
			return nil
		}
	}
	return ErrDipNotFound
}