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

	"github.com/nebulasio/go-nebulas/crypto/keystore"
	metrics "github.com/nebulasio/go-nebulas/metrics"

	"github.com/nebulasio/go-nebulas/account"

	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/net"

	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// Errors in PoW Consensus
var (
	ErrInvalidBlockInterval   = errors.New("invalid block interval")
	ErrMissingConfigForDpos   = errors.New("missing configuration for Dpos")
	ErrInvalidBlockProposer   = errors.New("invalid block proposer")
	ErrCannotMintWhenPending  = errors.New("cannot mint block now, waiting for cancel pending again")
	ErrCannotMintWhenDiable   = errors.New("cannot mint block now, waiting for enable it again")
	ErrWaitingBlockInLastSlot = errors.New("cannot mint block now, waiting for last block")
	ErrBlockMintedInNextSlot  = errors.New("cannot mint block now, there is a block minted in current slot")
)

// Metrics
var (
	metricsBlockPackingTime = metrics.NewGauge("neb.block.packing")
)

// Neblet interface breaks cycle import dependency and hides unused services.
type Neblet interface {
	Config() *nebletpb.Config
	BlockChain() *core.BlockChain
	NetService() net.Service
	AccountManager() *account.Manager
}

// Dpos Delegate Proof-of-Stake
type Dpos struct {
	quitCh chan bool

	chain *core.BlockChain
	ns    net.Service
	am    *account.Manager

	coinbase *core.Address
	miner    *core.Address

	blockInterval   int64
	dynastyInterval int64
	txsPerBlock     int

	enable  bool
	pending bool
}

// NewDpos create Dpos instance.
func NewDpos(neblet Neblet) (*Dpos, error) {
	p := &Dpos{
		quitCh: make(chan bool, 5),

		chain: neblet.BlockChain(),
		ns:    neblet.NetService(),
		am:    neblet.AccountManager(),

		blockInterval:   core.BlockInterval,
		dynastyInterval: core.DynastyInterval,
		txsPerBlock:     10000,

		enable:  false,
		pending: true,
	}

	config := neblet.Config().Chain
	coinbase, err := core.AddressParse(config.Coinbase)
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"address": config.Coinbase,
			"err":     err,
		}).Error("Failed to parse coinbase address.")
		return nil, err
	}
	miner, err := core.AddressParse(config.Miner)
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"address": config.Miner,
			"err":     err,
		}).Error("Failed to parse miner address.")
		return nil, err
	}
	p.coinbase = coinbase
	p.miner = miner
	return p, nil
}

// Start start pow service.
func (p *Dpos) Start() {
	logging.CLog().Info("Starting Dpos Mining...")
	go p.blockLoop()
}

// Stop stop pow service.
func (p *Dpos) Stop() {
	logging.CLog().Info("Stopping Dpos Mining...")
	p.DisableMining()
	p.quitCh <- true
}

// EnableMining start the consensus
func (p *Dpos) EnableMining(passphrase string) error {
	if err := p.am.Unlock(p.miner, []byte(passphrase), keystore.YearUnlockDuration); err != nil {
		return err
	}
	p.enable = true
	logging.CLog().Info("Enabled Dpos Mining...")
	return nil
}

// DisableMining stop the consensus
func (p *Dpos) DisableMining() error {
	if err := p.am.Lock(p.miner); err != nil {
		return err
	}
	p.enable = false
	logging.CLog().Info("Disable Dpos Mining...")
	return nil
}

// Enable returns is mining
func (p *Dpos) Enable() bool {
	return p.enable
}

func less(a *core.Block, b *core.Block) bool {
	if a.Height() != b.Height() {
		return a.Height() < b.Height()
	}
	return byteutils.Less(a.Hash(), b.Hash())
}

// ForkChoice select new tail
func (p *Dpos) ForkChoice() error {
	bc := p.chain
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

// Pending return if consensus can do mining now
func (p *Dpos) Pending() bool {
	return p.pending
}

// SuspendMining pend dpos mining
func (p *Dpos) SuspendMining() {
	logging.CLog().Info("Suspended Dpos Mining.")
	p.pending = true
}

// ResumeMining continue dpos mining
func (p *Dpos) ResumeMining() {
	logging.CLog().Info("Resumed Dpos Mining.")
	p.pending = false
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

// FastVerifyBlock verify the block before its parent found
// can be verified if the block's dynasty == tail's dynasty
// can be verified if the block's dynasty == tails's next dynasty
func (p *Dpos) FastVerifyBlock(block *core.Block) error {
	tail := p.chain.TailBlock()
	// check timestamp
	elapsedSecond := block.Timestamp() - tail.Timestamp()
	if elapsedSecond%p.blockInterval != 0 {
		return ErrInvalidBlockInterval
	}
	// check proposer
	currentHour := block.Timestamp() / core.DynastyInterval
	tailHour := tail.Timestamp() / core.DynastyInterval
	var dynastyRoot byteutils.Hash
	if currentHour == tailHour {
		dynastyRoot = tail.DposContext().DynastyRoot
	} else if currentHour == tailHour+1 {
		dynastyRoot = tail.DposContext().NextDynastyRoot
	} else {
		return nil
	}
	dynasty, err := trie.NewBatchTrie(dynastyRoot, p.chain.Storage())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":   err,
			"root":  dynastyRoot,
			"block": block,
		}).Debug("Failed to create new trie.")
		return err
	}
	proposer, err := core.FindProposer(block.Timestamp(), dynasty)
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

// VerifyBlock verify the block with its parent found
func (p *Dpos) VerifyBlock(block *core.Block, parent *core.Block) error {
	// check proposer
	dynasty, err := trie.NewBatchTrie(block.DposContext().DynastyRoot, block.Storage())
	if err != nil {
		return err
	}
	proposer, err := core.FindProposer(block.Timestamp(), dynasty)
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

func (p *Dpos) newBlock(tail *core.Block, context *core.DynastyContext, deadline int64) (*core.Block, error) {
	block, err := core.NewBlock(p.chain.ChainID(), p.coinbase, tail)
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"tail":     tail,
			"coinbase": p.coinbase,
			"chainid":  p.chain.ChainID(),
			"err":      err,
		}).Error("Failed to create new block")
		return nil, err
	}

	logging.CLog().WithFields(logrus.Fields{
		"coinbase": p.coinbase,
		"reward":   core.BlockReward,
	}).Info("Rewarded the coinbase.")

	if err := block.LoadDynastyContext(context); err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"block": block,
			"err":   err,
		}).Error("Failed to load dynasty context")
		return nil, err
	}
	block.CollectTransactions(deadline)
	block.SetMiner(p.miner)
	if err = block.Seal(); err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"block": block,
			"err":   err,
		}).Error("Failed to seal new block")
		return nil, err
	}
	if err = p.am.SignBlock(p.miner, block); err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"miner": p.miner,
			"block": block,
			"err":   err,
		}).Error("Failed to sign new block")
		return nil, err
	}

	logging.CLog().WithFields(logrus.Fields{
		"now":      time.Now().Unix(),
		"deadline": deadline,
		"txs":      len(block.Transactions()),
	}).Info("Packed txs.")

	return block, nil
}

func lastSlot(now int64) int64 {
	return int64((now-1)/core.BlockInterval) * core.BlockInterval
}

func nextSlot(now int64) int64 {
	return int64((now+core.BlockInterval-1)/core.BlockInterval) * core.BlockInterval
}

func deadline(now int64) int64 {
	nextSlot := nextSlot(now)
	remain := nextSlot - now
	if core.MaxMintDuration > remain {
		return nextSlot
	}
	return now + core.MaxMintDuration
}

func (p *Dpos) checkDeadline(tail *core.Block, now int64) (int64, error) {
	lastSlot := lastSlot(now)
	nextSlot := nextSlot(now)

	if tail.Timestamp() == nextSlot {
		return 0, ErrBlockMintedInNextSlot
	}
	if tail.Timestamp() == lastSlot {
		return deadline(now), nil
	}
	if nextSlot-now <= core.MinMintDuration {
		return deadline(now), nil
	}
	return 0, ErrWaitingBlockInLastSlot
}

func (p *Dpos) checkProposer(tail *core.Block, now int64) (*core.DynastyContext, error) {
	slot := nextSlot(now)
	elapsed := slot - tail.Timestamp()
	context, err := tail.NextDynastyContext(p.chain, elapsed)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"tail":    tail,
			"elapsed": elapsed,
			"err":     err,
		}).Debug("Failed to generate next dynasty context.")
		return nil, core.ErrGenerateNextDynastyContext
	}
	if context.Proposer == nil || !context.Proposer.Equals(p.miner.Bytes()) {
		proposer := "nil"
		if context.Proposer != nil {
			proposer = string(context.Proposer.Hex())
		}
		logging.VLog().WithFields(logrus.Fields{
			"tail":     tail,
			"now":      now,
			"slot":     slot,
			"expected": proposer,
			"actual":   p.miner,
		}).Debug("Not my turn, waiting...")
		return nil, ErrInvalidBlockProposer
	}
	return context, nil
}

func (p *Dpos) broadcast(tail *core.Block, block *core.Block) error {
	if err := p.chain.BlockPool().PushAndBroadcast(block); err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"tail":  tail,
			"block": block,
			"err":   err,
		}).Error("Failed to push new minted block into block pool")
		return err
	}
	return nil
}

func (p *Dpos) mintBlock(now int64) error {
	metricsBlockPackingTime.Update(0)

	// check mining enable
	if !p.enable {
		return ErrCannotMintWhenDiable
	}

	// check mining pending
	if p.pending {
		return ErrCannotMintWhenPending
	}

	tail := p.chain.TailBlock()

	deadline, err := p.checkDeadline(tail, now)
	if err != nil {
		return err
	}

	context, err := p.checkProposer(tail, now)
	if err != nil {
		return err
	}

	logging.CLog().WithFields(logrus.Fields{
		"tail":     tail,
		"start":    now,
		"deadline": deadline,
		"expected": context.Proposer.Hex(),
		"actual":   p.coinbase,
	}).Info("My turn to mint block")
	metricsBlockPackingTime.Update(deadline - now)

	block, err := p.newBlock(tail, context, deadline)
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

	if err := p.broadcast(tail, block); err != nil {
		return err
	}

	logging.CLog().WithFields(logrus.Fields{
		"tail":  tail,
		"block": block,
	}).Info("Broadcasted new block")
	return nil
}

func (p *Dpos) blockLoop() {
	logging.CLog().Info("Started Dpos Mining.")
	timeChan := time.NewTicker(time.Second).C
	for {
		select {
		case now := <-timeChan:
			p.mintBlock(now.Unix())
		case <-p.quitCh:
			logging.CLog().Info("Stopped Dpos Mining.")
			return
		}
	}
}
