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
	"errors"
)

type Dip struct {

	neb Neblet

	cache *lru.Cache

	// reward rewardAddress
	rewardAddress *core.Address
	rewardValue   *util.Uint128

	isLooping		bool
	quitCh            chan int
}

// NewDIP create a dip
func NewDIP(neb Neblet) (*Dip, error) {
	cache, err := lru.New(CacheSize)
	if err != nil {
		return nil, err
	}

	// dip reward address.
	priv, err := byteutils.FromHex(DipRewardAddressPrivate)
	if err != nil {
		return nil, err
	}
	//TODO(larry): The current award rewardAddress is constant and visible to all.
	addr, err := neb.AccountManager().LoadPrivate(priv, []byte(DipRewardAddressPassphrase))
	if err != nil {
		return nil, err
	}

	dip := &Dip{
		neb: neb,
		cache:cache,
		quitCh:make(chan int, 1),
		isLooping:false,
		rewardAddress: addr,
		rewardValue: DipRewardValue,
	}
	return dip, nil
}

// RewardAddress return dip reward rewardAddress.
func (d *Dip)RewardAddress() *core.Address {
	return d.rewardAddress
}

func (d *Dip)RewardValue() *util.Uint128 {
	return d.rewardValue
}

// Start start dip.
func (d *Dip) Start() {
	if !d.isLooping {
		d.isLooping = true

		logging.CLog().WithFields(logrus.Fields{
		}).Info("Starting Dip...")

		go d.loop()
	}
}

// Stop stop dip.
func (d *Dip) Stop() {
	if d.isLooping {
		logging.CLog().WithFields(logrus.Fields{
		}).Info("Stopping Dip...")

		d.quitCh <- 1
		d.isLooping = false
	}
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
	height := d.neb.BlockChain().TailBlock().Height() - uint64(DipDelayRewardHeight)
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
	endBlock := d.neb.BlockChain().GetBlockOnCanonicalChainByHeight(dipData.EndHeight)
	endAccount, err := endBlock.GetAccount(d.RewardAddress().Bytes())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Warn("Failed to load reward account in dip end account.")
		return
	}

	tailAccount, err := d.neb.BlockChain().TailBlock().GetAccount(d.RewardAddress().Bytes())
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

			d.neb.BlockChain().TransactionPool().Push(tx)
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
	if err = d.neb.AccountManager().SignTransactionWithPassphrase(d.RewardAddress(), tx, []byte(DipRewardAddressPassphrase)); err != nil {
		return nil, err
	}
	return tx, nil
}

// GetDipList returns dip info list
func (d *Dip) GetDipList(height uint64) (core.Data, error) {
	if height <= 0 || height > d.neb.BlockChain().TailBlock().Height()  {
		return nil, ErrInvalidHeight
	}
	data, ok := d.checkCache(height)
	if !ok {
		dipData, err := d.neb.Nbre().Execute(nbre.CommandDIPList, byteutils.FromUint64(height))
		if err != nil {
			return nil, err
		}
		data = &DIPData{}
		if err := data.FromBytes(dipData); err != nil {
			return nil, err
		}
		if len(data.Err) > 0 {
			return nil, errors.New(data.Err)
		}
		key := append(byteutils.FromUint64(data.StartHeight), byteutils.FromUint64(data.EndHeight)...)
		d.cache.Add(byteutils.Hex(key), data)
	}
	return data, nil
}

func (d *Dip) checkCache(height uint64) (*DIPData, bool) {
	keys:= d.cache.Keys()
	for _, v := range keys {
		bytes, err := byteutils.FromHex(v.(string))
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err": err,
			}).Fatal("Failed to parse dip cache.")
		}
		start := byteutils.Uint64(bytes[:8])
		end := byteutils.Uint64(bytes[8:])
		if height >= start && height <= end {
			data, _ := d.cache.Get(v)
			return data.(*DIPData), true
		}
	}
	return nil, false
}

func (d *Dip) CheckReward(height uint64, tx *core.Transaction) error {
	//if tx type is not dip, can't use the reward address send tx.
	if tx.Type() != core.TxPayloadDipType && tx.From().Equals(d.rewardAddress) {
		return ErrUnsupportedTransactionFromDipAddress
	}

	if tx.Type() == core.TxPayloadDipType {
		if !tx.From().Equals(d.rewardAddress) {
			return ErrInvalidDipAddress
		}

		data, err := d.GetDipList(height - uint64(DipDelayRewardHeight))
		if err != nil {
			return err
		}
		dip := data.(*DIPData)
		for _, v := range dip.Dips {
			if tx.To().String() == v.Address && tx.Value().String() == v.Reward {
				return nil
			}
		}
		return ErrDipNotFound
	}
	return nil
}