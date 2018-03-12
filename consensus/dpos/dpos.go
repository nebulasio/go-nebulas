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

package dpos

import (
	"errors"
	"time"

	"github.com/nebulasio/go-nebulas/core/state"

	"github.com/nebulasio/go-nebulas/crypto/keystore"
	metrics "github.com/nebulasio/go-nebulas/metrics"

	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/net"

	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// Errors in PoW Consensus
var (
	ErrInvalidBlockInterval       = errors.New("invalid block interval")
	ErrMissingConfigForDpos       = errors.New("missing configuration for Dpos")
	ErrInvalidBlockProposer       = errors.New("invalid block proposer")
	ErrCannotMintWhenPending      = errors.New("cannot mint block now, waiting for cancel pending again")
	ErrCannotMintWhenDiable       = errors.New("cannot mint block now, waiting for enable it again")
	ErrWaitingBlockInLastSlot     = errors.New("cannot mint block now, waiting for last block")
	ErrBlockMintedInNextSlot      = errors.New("cannot mint block now, there is a block minted in current slot")
	ErrGenerateNextConsensusState = errors.New("Failed to generate next consensus state")
)

// Metrics
var (
	metricsBlockPackingTime = metrics.NewGauge("neb.block.packing")
)

// Dpos Delegate Proof-of-Stake
type Dpos struct {
	quitCh chan bool

	chain *core.BlockChain
	ns    net.Service
	am    core.Manager

	coinbase *core.Address
	miner    *core.Address

	enable  bool
	pending bool
}

// NewDpos create Dpos instance.
func NewDpos() *Dpos {
	dpos := &Dpos{
		quitCh:  make(chan bool, 5),
		enable:  false,
		pending: true,
	}
	return dpos
}

// Setup a dpos consensus handler
func (dpos *Dpos) Setup(neblet core.Neblet) error {
	dpos.chain = neblet.BlockChain()
	dpos.ns = neblet.NetService()
	dpos.am = neblet.AccountManager()

	config := neblet.Config().Chain
	coinbase, err := core.AddressParse(config.Coinbase)
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"address": config.Coinbase,
			"err":     err,
		}).Error("Failed to parse coinbase address.")
		return err
	}
	miner, err := core.AddressParse(config.Miner)
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"address": config.Miner,
			"err":     err,
		}).Error("Failed to parse miner address.")
		return err
	}
	dpos.coinbase = coinbase
	dpos.miner = miner
	return nil
}

// Start start pow service.
func (dpos *Dpos) Start() {
	logging.CLog().Info("Starting Dpos Mining...")
	go dpos.blockLoop()
}

// Stop stop pow service.
func (dpos *Dpos) Stop() {
	logging.CLog().Info("Stopping Dpos Mining...")
	dpos.DisableMining()
	dpos.quitCh <- true
}

// EnableMining start the consensus
func (dpos *Dpos) EnableMining(passphrase string) error {
	if err := dpos.am.Unlock(dpos.miner, []byte(passphrase), keystore.YearUnlockDuration); err != nil {
		return err
	}
	dpos.enable = true
	logging.CLog().Info("Enabled Dpos Mining...")
	return nil
}

// DisableMining stop the consensus
func (dpos *Dpos) DisableMining() error {
	if err := dpos.am.Lock(dpos.miner); err != nil {
		return err
	}
	dpos.enable = false
	logging.CLog().Info("Disable Dpos Mining...")
	return nil
}

// Enable returns is mining
func (dpos *Dpos) Enable() bool {
	return dpos.enable
}

func less(a *core.Block, b *core.Block) bool {
	if a.Height() != b.Height() {
		return a.Height() < b.Height()
	}
	if len(a.Transactions()) != len(b.Transactions()) {
		return len(a.Transactions()) < len(b.Transactions())
	}
	return byteutils.Less(a.Hash(), b.Hash())
}

// ForkChoice select new tail
func (dpos *Dpos) ForkChoice() error {
	bc := dpos.chain
	tailBlock := bc.TailBlock()
	detachedTailBlocks := bc.DetachedTailBlocks()

	// find the max depth.
	newTailBlock := tailBlock

	for _, v := range detachedTailBlocks {
		if less(newTailBlock, v) {
			newTailBlock = v
		}
	}

	if newTailBlock.Hash().Equals(tailBlock.Hash()) {
		logging.VLog().WithFields(logrus.Fields{
			"old tail": tailBlock,
			"new tail": newTailBlock,
		}).Info("Current tail is best, no need to change.")
		return nil
	}

	err := bc.SetTailBlock(newTailBlock)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"new tail": newTailBlock,
			"old tail": tailBlock,
			"err":      err,
		}).Error("Failed to set new tail block.")
		return err
	}

	logging.VLog().WithFields(logrus.Fields{
		"new tail": newTailBlock,
		"old tail": tailBlock,
	}).Info("change to new tail.")
	return nil
}

// UpdateLIB update the latest irrversible block
func (dpos *Dpos) UpdateLIB() {
	lib := dpos.chain.LIB()
	tail := dpos.chain.TailBlock()
	cur := tail
	miners := make(map[string]bool)
	dynasty := int64(0)
	for !cur.Hash().Equals(lib.Hash()) {
		curDynasty := cur.Timestamp() / DynastyInterval
		if curDynasty != dynasty {
			miners = make(map[string]bool)
			dynasty = curDynasty
		}
		// fast prune
		if int(cur.Height())-int(lib.Height()) < ConsensusSize-len(miners) {
			return
		}
		miners[cur.Miner().String()] = true
		if len(miners) >= ConsensusSize {
			if err := dpos.chain.StoreLIBToStorage(cur); err != nil {
				logging.VLog().WithFields(logrus.Fields{
					"tail": tail,
					"lib":  cur,
				}).Debug("Failed to store latest irreversible block.")
				return
			}
			logging.VLog().WithFields(logrus.Fields{
				"lib.new":          cur,
				"lib.old":          lib,
				"tail":             tail,
				"miners.limit":     ConsensusSize,
				"miners.supported": len(miners),
			}).Info("Succeed to update latest irreversible block.")
			dpos.chain.SetLIB(cur)

			e := &state.Event{
				Topic: core.TopicLibBlock,
				Data:  dpos.chain.LIB().String(),
			}
			dpos.chain.EventEmitter().Trigger(e)
			return
		}

		tmp := cur
		cur = dpos.chain.GetBlock(cur.ParentHash())
		if cur == nil || core.CheckGenesisBlock(cur) {
			logging.VLog().WithFields(logrus.Fields{
				"tail": tail,
				"cur":  tmp,
			}).Debug("Failed to find latest irreversible block.")
			return
		}
	}

	logging.VLog().WithFields(logrus.Fields{
		"cur":              cur,
		"lib":              lib,
		"tail":             tail,
		"err":              "supported miners is not enough",
		"miners.limit":     ConsensusSize,
		"miners.supported": len(miners),
	}).Warn("Failed to update latest irreversible block.")
}

// Pending return if consensus can do mining now
func (dpos *Dpos) Pending() bool {
	return dpos.pending
}

// SuspendMining pend dpos mining
func (dpos *Dpos) SuspendMining() {
	logging.CLog().Info("Suspended Dpos Mining.")
	dpos.pending = true
}

// ResumeMining continue dpos mining
func (dpos *Dpos) ResumeMining() {
	logging.CLog().Info("Resumed Dpos Mining.")
	dpos.pending = false
}

func verifyBlockSign(miner *core.Address, block *core.Block) error {
	addr, err := core.RecoverMiner(block)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"address": addr,
			"err":     err,
			"block":   block,
		}).Debug("Failed to recover block's miner.")
		return err
	}
	if !miner.Equals(addr) {
		logging.VLog().WithFields(logrus.Fields{
			"address": addr,
			"miner":   miner,
			"block":   block,
		}).Debug("Failed to verify block's sign.")
		return ErrInvalidBlockProposer
	}
	block.SetMiner(miner)
	return nil
}

// VerifyBlock verify the block
func (dpos *Dpos) VerifyBlock(block *core.Block) error {
	tail := dpos.chain.TailBlock()
	// check timestamp
	elapsedSecond := block.Timestamp() - tail.Timestamp()
	if elapsedSecond%BlockInterval != 0 {
		return ErrInvalidBlockInterval
	}
	// check proposer
	dynastyRoot := tail.WorldState().DynastyRoot()
	dynasty, err := trie.NewTrie(dynastyRoot, dpos.chain.Storage())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":   err,
			"root":  dynastyRoot,
			"block": block,
		}).Debug("Failed to create new trie.")
		return err
	}
	proposer, err := FindProposer(block.Timestamp(), dynasty)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"proposer": proposer,
			"err":      err,
			"block":    block,
		}).Debug("Failed to find proposer.")
		return err
	}
	miner, err := core.AddressParseFromBytes(proposer)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"proposer": proposer,
			"err":      err,
			"block":    block,
		}).Debug("Failed to parse proposer.")
		return err
	}
	return verifyBlockSign(miner, block)
}

func (dpos *Dpos) newBlock(tail *core.Block, consensusState state.ConsensusState, deadline int64) (*core.Block, error) {
	startAt := time.Now().Unix()

	block, err := core.NewBlock(dpos.chain.ChainID(), dpos.coinbase, tail)
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"tail":     tail,
			"coinbase": dpos.coinbase,
			"chainid":  dpos.chain.ChainID(),
			"err":      err,
		}).Error("Failed to create new block")
		return nil, err
	}

	logging.CLog().WithFields(logrus.Fields{
		"coinbase": dpos.coinbase,
		"reward":   core.BlockReward,
	}).Info("Rewarded the coinbase.")

	block.WorldState().SetConsensusState(consensusState)
	block.SetTimestamp(consensusState.TimeStamp())
	block.CollectTransactions(deadline)
	block.SetMiner(dpos.miner)
	if err = block.Seal(); err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"block": block,
			"err":   err,
		}).Error("Failed to seal new block")
		return nil, err
	}
	if err = dpos.am.SignBlock(dpos.miner, block); err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"miner": dpos.miner,
			"block": block,
			"err":   err,
		}).Error("Failed to sign new block")
		return nil, err
	}
	endAt := time.Now().Unix()

	logging.CLog().WithFields(logrus.Fields{
		"start": startAt,
		"end":   endAt,
		"diff":  endAt - startAt,
		"block": block,
		"txs":   len(block.Transactions()),
	}).Info("Packed txs.")

	return block, nil
}

func lastSlot(now int64) int64 {
	return int64((now-1)/BlockInterval) * BlockInterval
}

func nextSlot(now int64) int64 {
	return int64((now+BlockInterval-1)/BlockInterval) * BlockInterval
}

func deadline(now int64) int64 {
	nextSlot := nextSlot(now)
	remain := nextSlot - now
	if MaxMintDuration > remain {
		return nextSlot
	}
	return now + MaxMintDuration
}

func (dpos *Dpos) checkDeadline(tail *core.Block, now int64) (int64, error) {
	lastSlot := lastSlot(now)
	nextSlot := nextSlot(now)

	if tail.Timestamp() >= nextSlot {
		return 0, ErrBlockMintedInNextSlot
	}
	if tail.Timestamp() == lastSlot {
		return deadline(now), nil
	}
	if nextSlot-now <= MinMintDuration {
		return deadline(now), nil
	}
	return 0, ErrWaitingBlockInLastSlot
}

func (dpos *Dpos) checkProposer(tail *core.Block, now int64) (state.ConsensusState, error) {
	slot := nextSlot(now)
	elapsed := slot - tail.Timestamp()
	consensusState, err := tail.WorldState().NextConsensusState(elapsed)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"tail":    tail,
			"elapsed": elapsed,
			"err":     err,
		}).Debug("Failed to generate next dynasty context.")
		return nil, ErrGenerateNextConsensusState
	}
	if consensusState.Proposer() == nil || !consensusState.Proposer().Equals(dpos.miner.Bytes()) {
		proposer := "nil"
		if consensusState.Proposer() != nil {
			proposer = string(consensusState.Proposer().Hex())
		}
		logging.VLog().WithFields(logrus.Fields{
			"tail":     tail,
			"now":      now,
			"slot":     slot,
			"expected": proposer,
			"actual":   dpos.miner,
		}).Debug("Not my turn, waiting...")
		return nil, ErrInvalidBlockProposer
	}
	return consensusState, nil
}

func (dpos *Dpos) broadcast(tail *core.Block, block *core.Block) error {
	if err := dpos.chain.BlockPool().PushAndBroadcast(block); err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"tail":  tail,
			"block": block,
			"err":   err,
		}).Error("Failed to push new minted block into block pool")
		return err
	}
	return nil
}

func (dpos *Dpos) mintBlock(now int64) error {
	metricsBlockPackingTime.Update(0)

	// check mining enable
	if !dpos.enable {
		return ErrCannotMintWhenDiable
	}

	// check mining pending
	if dpos.pending {
		return ErrCannotMintWhenPending
	}

	tail := dpos.chain.TailBlock()

	deadline, err := dpos.checkDeadline(tail, now)
	if err != nil {
		return err
	}

	consensusState, err := dpos.checkProposer(tail, now)
	if err != nil {
		return err
	}

	logging.CLog().WithFields(logrus.Fields{
		"tail":     tail,
		"start":    now,
		"deadline": deadline,
		"expected": consensusState.Proposer().Hex(),
		"actual":   dpos.coinbase,
	}).Info("My turn to mint block")
	metricsBlockPackingTime.Update(deadline - now)

	block, err := dpos.newBlock(tail, consensusState, deadline)
	if err != nil {
		return err
	}

	slot := nextSlot(now)
	current := time.Now().Unix()
	if slot > current {
		timer := time.NewTimer(time.Duration(slot-current) * time.Second).C
		<-timer
	}

	logging.CLog().WithFields(logrus.Fields{
		"tail":     tail,
		"block":    block,
		"start":    now,
		"packed":   current,
		"deadline": deadline,
		"slot":     slot,
		"end":      time.Now().Unix(),
	}).Info("Minted new block")

	if err := dpos.broadcast(tail, block); err != nil {
		return err
	}

	logging.CLog().WithFields(logrus.Fields{
		"tail":  tail,
		"block": block,
	}).Info("Broadcasted new block")
	return nil
}

func (dpos *Dpos) blockLoop() {
	logging.CLog().Info("Started Dpos Mining.")
	timeChan := time.NewTicker(time.Second).C
	for {
		select {
		case now := <-timeChan:
			dpos.mintBlock(now.Unix())
		case <-dpos.quitCh:
			logging.CLog().Info("Stopped Dpos Mining.")
			return
		}
	}
}
