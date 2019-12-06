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
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

type Dip struct {
	neb Neblet

	cache *lru.Cache

	// reward rewardAddress
	rewardAddress *core.Address
	rewardValue   *util.Uint128

	isLooping bool
	quitCh    chan int
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
	addr, err := neb.AccountManager().LoadPrivate(priv, []byte(DipRewardAddressPassphrase))
	if err != nil {
		return nil, err
	}

	dip := &Dip{
		neb:           neb,
		cache:         cache,
		quitCh:        make(chan int, 1),
		isLooping:     false,
		rewardAddress: addr,
		rewardValue:   core.DIPRewardV2,
	}
	return dip, nil
}

// RewardAddress return dip reward rewardAddress.
func (d *Dip) RewardAddress() *core.Address {
	return d.rewardAddress
}

func (d *Dip) RewardValue() *util.Uint128 {
	return d.rewardValue
}

// Start start dip.
func (d *Dip) Start() {
	if !d.isLooping {
		d.isLooping = true

		logging.CLog().WithFields(logrus.Fields{}).Info("Starting Dip...")

		go d.loop()
	}
}

// Stop stop dip.
func (d *Dip) Stop() {
	if d.isLooping {
		d.isLooping = false
		logging.CLog().WithFields(logrus.Fields{}).Info("Stopping Dip...")

		d.quitCh <- 1
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
			if d.isLooping {
				d.submitReward()
			}
		}
	}
}

func (d *Dip) DipDelayRewardHeight() uint64 {
	chainID := d.neb.BlockChain().ChainID()
	// for Mainnet and Testnet, delay 0.5 day to submit dip reward.
	if chainID == core.MainNetID {
		return 12 * 60 * 60 / 15
	} else if chainID == core.TestNetID {
		return 12 * 60 * 60 / 15
	} else {
		return 33
	}
}

// submitReward generate dip transactions and push to tx pool
func (d *Dip) submitReward() {
	height := d.neb.BlockChain().TailBlock().Height() - d.DipDelayRewardHeight()
	if height < 1 {
		return
	}
	logging.VLog().WithFields(logrus.Fields{
		"height": height,
	}).Debug("loop  to query dip for submit.")
	data, err := d.GetDipList(height, 0)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Warn("Failed to get dip reward list.")
		return
	}

	dipData := data.(*DIPData)
	endBlock := d.neb.BlockChain().GetBlockOnCanonicalChainByHeight(dipData.EndHeight)
	endAccount, err := endBlock.GetAccount(d.RewardAddress().Bytes())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to load reward account in dip end account.")
		return
	}

	tailAccount, err := d.neb.BlockChain().TailBlock().GetAccount(d.RewardAddress().Bytes())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to load reward account in tail block.")
		return
	}

	for idx, v := range dipData.Dips {
		nonce := endAccount.Nonce() + uint64(idx) + 1
		// only not reward tx can be pushed to tx pool.
		if nonce > tailAccount.Nonce() {
			tx, err := d.generateRewardTx(dipData.StartHeight, dipData.EndHeight, dipData.Version, v, nonce, endBlock)
			if err != nil {
				logging.VLog().WithFields(logrus.Fields{
					"err": err,
				}).Error("Failed to generate reward transaction.")
				return
			}

			logging.VLog().WithFields(logrus.Fields{
				"height": d.neb.BlockChain().TailBlock().Height(),
				"start":  dipData.StartHeight,
				"end":    dipData.EndHeight,
				"tx":     tx,
			}).Info("Success to push dip reward tx.")

			d.neb.BlockChain().TransactionPool().Push(tx)
		}
	}
}

func (d *Dip) generateRewardTx(start uint64, end uint64, version uint64, item *DIPItem, nonce uint64, block *core.Block) (*core.Transaction, error) {
	var (
		to           *core.Address
		value        *util.Uint128
		payload      *core.DipPayload
		payloadBytes []byte
		gasLimit     *util.Uint128
		err          error
	)
	if to, err = core.AddressParse(item.Address); err != nil {
		return nil, err
	}
	if value, err = util.NewUint128FromString(item.Reward); err != nil {
		return nil, err
	}
	if payload, err = core.NewDipPayload(start, end, version, item.Contract); err != nil {
		return nil, err
	}
	if payloadBytes, err = payload.ToBytes(); err != nil {
		return nil, err
	}

	if gasLimit, err = core.MinGasCountPerTransaction.Mul(util.NewUint128FromUint(10)); err != nil {
		return nil, err
	}

	tx, err := core.NewTransaction(
		block.ChainID(),
		d.RewardAddress(),
		to,
		value,
		nonce,
		core.TxPayloadDipType,
		payloadBytes,
		d.neb.BlockChain().TransactionPool().GetMinGasPrice(),
		gasLimit)
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
func (d *Dip) GetDipList(height, version uint64) (core.Data, error) {
	data, err := d.checkCache(height)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (d *Dip) checkCache(height uint64) (*DIPData, error) {
	if d.cache.Len() == 0 {
		d.loadCache()
	}

	key := (height-core.NrStartHeight)/core.NrIntervalHeight - 1
	if data, ok := d.cache.Get(key); ok {
		dipData := data.(*DIPData)
		logging.VLog().WithFields(logrus.Fields{
			"height":   height,
			"start":    dipData.StartHeight,
			"end":      dipData.EndHeight,
			"dataSize": len(dipData.Dips),
		}).Debug("Success to find dip list in cache.")
		return dipData, nil
	}
	return nil, ErrDipNotFound
}

func (d *Dip) CheckReward(tx *core.Transaction) error {
	//if tx type is not dip, can't use the reward address send tx.
	if tx.Type() != core.TxPayloadDipType && tx.From().Equals(d.rewardAddress) {
		return ErrUnsupportedTransactionFromDipAddress
	}

	if tx.Type() == core.TxPayloadDipType {
		if !tx.From().Equals(d.rewardAddress) {
			return ErrInvalidDipAddress
		}

		if tx.To().Equals(core.DIPRewardAddressV2) {
			return nil
		}

		payload, err := tx.LoadPayload()
		if err != nil {
			return err
		}

		dipPayload := payload.(*core.DipPayload)
		height := dipPayload.EndHeight + 1
		data, err := d.GetDipList(height, dipPayload.Version)
		if err != nil {
			return err
		}
		dip := data.(*DIPData)
		for _, v := range dip.Dips {
			if tx.To().String() == v.Address && tx.Value().String() == v.Reward {
				logging.VLog().WithFields(logrus.Fields{
					"height": height,
					"start":  dip.StartHeight,
					"end":    dip.EndHeight,
					"tx":     tx,
				}).Debug("Success to check dip reward tx.")
				return nil
			}
		}
		return ErrDipNotFound
	}
	return nil
}
