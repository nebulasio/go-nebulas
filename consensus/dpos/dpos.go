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
	"time"

	"github.com/nebulasio/go-nebulas/rpc"
	rpcpb "github.com/nebulasio/go-nebulas/rpc/pb"
	"golang.org/x/net/context"

	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/crypto/keystore"

	lru "github.com/hashicorp/golang-lru"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// Dpos Delegate Proof-of-Stake
type Dpos struct {
	quitCh chan bool

	chain *core.BlockChain
	ns    net.Service
	am    core.AccountManager

	dynasty *Dynasty

	coinbase               *core.Address
	miner                  *core.Address
	enableRemoteSignServer bool
	remoteSignServer       string

	slot *lru.Cache

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

	dynasty, err := NewDynasty(neblet)
	if err != nil {
		return err
	}
	dpos.dynasty = dynasty

	chainConfig := neblet.Config().Chain
	if chainConfig.StartMine {
		coinbase, err := core.AddressParse(chainConfig.Coinbase)
		if err != nil {
			logging.CLog().WithFields(logrus.Fields{
				"address": chainConfig.Coinbase,
				"err":     err,
			}).Error("Failed to parse coinbase address.")
			return err
		}
		miner, err := core.AddressParse(chainConfig.Miner)
		if err != nil {
			logging.CLog().WithFields(logrus.Fields{
				"address": chainConfig.Miner,
				"err":     err,
			}).Error("Failed to parse miner address.")
			return err
		}
		dpos.coinbase = coinbase
		dpos.miner = miner
		dpos.enableRemoteSignServer = chainConfig.EnableRemoteSignServer
		dpos.remoteSignServer = chainConfig.RemoteSignServer
	}

	slot, err := lru.New(128)
	if err != nil {
		return err
	}
	dpos.slot = slot
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
	if err := dpos.unlock(passphrase); err != nil {
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
		}).Debug("Current tail is best, no need to change.")
		return nil
	}

	err := bc.SetTailBlock(newTailBlock)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"new tail": newTailBlock,
			"old tail": tailBlock,
			"err":      err,
		}).Debug("Failed to set new tail block.")
		return err
	}

	logging.VLog().WithFields(logrus.Fields{
		"new tail": newTailBlock,
		"old tail": tailBlock,
	}).Info("change to new tail.")
	return nil
}

// UpdateLIB update the latest irrversible block
func (dpos *Dpos) UpdateLIB(rversibleBlocks []byteutils.Hash) {
	lib := dpos.chain.LIB()
	tail := dpos.chain.TailBlock()
	cur := tail
	miners := make(map[string]bool)
	dynasty := int64(-1)
	for !cur.Hash().Equals(lib.Hash()) {
		curDynasty := cur.Timestamp() * SecondInMs / DynastyIntervalInMs
		if curDynasty != dynasty {
			miners = make(map[string]bool)
			dynasty = curDynasty
		}
		// fast prune
		if int(cur.Height())-int(lib.Height()) < ConsensusSize-len(miners) {
			return
		}
		miners[byteutils.Hex(cur.ConsensusRoot().Proposer)] = true
		if len(miners) >= ConsensusSize {
			if err := dpos.chain.StoreLIBHashToStorage(cur); err != nil {
				logging.VLog().WithFields(logrus.Fields{
					"tail": tail,
					"lib":  cur,
				}).Debug("Failed to store latest irreversible block.")
				return
			}
			logging.CLog().WithFields(logrus.Fields{
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
	}).Debug("Failed to update latest irreversible block.")
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
	signer, err := core.RecoverSignerFromSignature(block.Alg(), block.Hash(), block.Signature())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"signer": signer,
			"err":    err,
			"block":  block,
		}).Debug("Failed to recover block's miner.")
		return err
	}
	if !miner.Equals(signer) {
		logging.VLog().WithFields(logrus.Fields{
			"signer": signer,
			"miner":  miner,
			"block":  block,
		}).Debug("Failed to verify block's sign.")
		return ErrInvalidBlockProposer
	}
	return nil
}

// CheckDoubleMint if double mint exists
func (dpos *Dpos) CheckDoubleMint(block *core.Block) bool {
	if preBlock, exist := dpos.slot.Get(block.Timestamp()); exist {
		if preBlock.(*core.Block).Hash().Equals(block.Hash()) == false {
			logging.VLog().WithFields(logrus.Fields{
				"curBlock": block,
				"preBlock": preBlock.(*core.Block),
			}).Warn("Found someone minted multiple blocks at same time.")
			return true
		}
	}
	return false
}

// Serial return dynasty serial number
func (pod *Dpos) Serial(timestamp int64) int64 {
	return GenesisDynastySerial
}

// VerifyBlock verify the block
func (dpos *Dpos) VerifyBlock(block *core.Block) error {
	tail := dpos.chain.TailBlock()
	// check timestamp
	if block.Timestamp() != block.ConsensusRoot().Timestamp {
		return ErrInvalidBlockTimestamp
	}
	elapsedSecondInMs := block.Timestamp() * SecondInMs
	if elapsedSecondInMs <= 0 || (elapsedSecondInMs%BlockIntervalInMs) != 0 {
		return ErrInvalidBlockInterval
	}

	var (
		miners []byteutils.Hash
		err    error
	)

	cs, err := dpos.dynasty.getDynasty(block.Timestamp())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":   err,
			"tail":  tail,
			"block": block,
		}).Error("Failed to retrieve dynasty trie.")
		return err
	}
	miners, err = TraverseDynasty(cs)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":   err,
			"block": block,
		}).Debug("Failed to get miners from dynasty.")
		return err
	}
	proposer, err := FindProposer(block.Timestamp(), miners)
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
	// check signature
	if err := verifyBlockSign(miner, block); err != nil {
		return err
	}

	// check block random
	if core.RandomAvailableAtHeight(block.Height()) && !block.HasRandomSeed() {
		logging.VLog().WithFields(logrus.Fields{
			"blockHeight":      block.Height(),
			"compatibleHeight": core.NebCompatibility.RandomAvailableHeight(),
		}).Debug("No random found in block header.")
		return core.ErrInvalidBlockRandom
	}

	dpos.slot.Add(block.Timestamp(), block)
	return nil
}

func (dpos *Dpos) generateRandomSeed(block *core.Block, adminService rpcpb.AdminServiceClient) error {

	ancestorHash, parentSeed, err := dpos.chain.GetInputForVRFSigner(block.ParentHash(), block.Height())
	if err != nil {
		return err
	}

	if dpos.enableRemoteSignServer == true {
		if adminService == nil {
			return ErrInvalidArgument
		}
		// generate VRF hash,proof
		random, err := adminService.GenerateRandomSeed(
			context.Background(),
			&rpcpb.GenerateRandomSeedRequest{
				Address:      dpos.miner.String(),
				ParentSeed:   parentSeed,
				AncestorHash: ancestorHash,
			})
		if err != nil {
			return err
		}
		block.SetRandomSeed(random.VrfSeed, random.VrfProof)
		return nil
	}

	// generate VRF hash,proof
	vrfSeed, vrfProof, err := dpos.am.GenerateRandomSeed(dpos.miner, ancestorHash, parentSeed)
	if err != nil {
		return err
	}
	block.SetRandomSeed(vrfSeed, vrfProof)

	return nil
}

func (dpos *Dpos) remoteSignBlock(block *core.Block, adminService rpcpb.AdminServiceClient) error {
	if adminService == nil {
		return ErrInvalidArgument
	}
	alg := keystore.SECP256K1
	resp, err := adminService.SignHash(
		context.Background(),
		&rpcpb.SignHashRequest{
			Address: dpos.miner.String(),
			Hash:    block.Hash(),
			Alg:     uint32(alg),
		})
	if err != nil {
		return err
	}

	block.SetSignature(alg, resp.Data)
	return nil
}

func (dpos *Dpos) unlock(passphrase string) error {
	if dpos.enableRemoteSignServer == false {
		return dpos.am.Unlock(dpos.miner, []byte(passphrase), DefaultMaxUnlockDuration)
	}
	return nil

}

func (dpos *Dpos) newBlock(tail *core.Block, consensusState state.ConsensusState, deadlineInMs int64) (*core.Block, error) {
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

	var adminService rpcpb.AdminServiceClient
	if dpos.enableRemoteSignServer == true {
		conn, err := rpc.Dial(dpos.remoteSignServer)
		defer func() {
			if conn != nil {
				conn.Close()
			}
		}()
		if err != nil {
			return nil, err
		}
		adminService = rpcpb.NewAdminServiceClient(conn)
	}

	if core.RandomAvailableAtHeight(block.Height()) {
		err := dpos.generateRandomSeed(block, adminService)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"block": block,
				"err":   err,
			}).Error("Failed to generate random seed from remote.")
			return nil, err
		}
	}

	block.WorldState().SetConsensusState(consensusState)
	block.SetTimestamp(consensusState.TimeStamp())
	block.CollectTransactions(deadlineInMs)
	if err = block.Seal(); err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"block": block,
			"err":   err,
		}).Error("Failed to seal new block")
		go block.ReturnTransactions()
		return nil, err
	}

	if dpos.enableRemoteSignServer == true {
		err = dpos.remoteSignBlock(block, adminService)
	} else {
		err = dpos.am.SignBlock(dpos.miner, block)
	}
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"miner": dpos.miner,
			"block": block,
			"err":   err,
		}).Error("Failed to sign new block")
		go block.ReturnTransactions()
		return nil, err
	}
	endAt := time.Now().Unix()

	logging.VLog().WithFields(logrus.Fields{
		"start": startAt,
		"end":   endAt,
		"diff":  endAt - startAt,
		"block": block,
		"txs":   len(block.Transactions()),
	}).Debug("Packed txs.")

	return block, nil
}

func lastSlot(nowInMs int64) int64 {
	return int64((nowInMs-SecondInMs)/BlockIntervalInMs) * BlockIntervalInMs
}

func nextSlot(nowInMs int64) int64 {
	return int64((nowInMs+BlockIntervalInMs-SecondInMs)/BlockIntervalInMs) * BlockIntervalInMs
}

func deadline(nowInMs int64) int64 {
	nextSlotInMs := nextSlot(nowInMs)
	remainInMs := nextSlotInMs - nowInMs
	if MaxMintDurationInMs > remainInMs {
		return nextSlotInMs
	}
	return nowInMs + MaxMintDurationInMs
}

func (dpos *Dpos) checkDeadline(tail *core.Block, nowInMs int64) (int64, error) {
	lastSlotInMs := lastSlot(nowInMs)
	nextSlotInMs := nextSlot(nowInMs)

	if tail.Timestamp()*SecondInMs >= nextSlotInMs {
		return 0, ErrBlockMintedInNextSlot
	}
	if tail.Timestamp()*SecondInMs == lastSlotInMs {
		return deadline(nowInMs), nil
	}
	if nextSlotInMs-nowInMs <= MinMintDurationInMs {
		return deadline(nowInMs), nil
	}
	return 0, ErrWaitingBlockInLastSlot
}

func (dpos *Dpos) checkProposer(tail *core.Block, nowInMs int64) (state.ConsensusState, error) {
	slotInMs := nextSlot(nowInMs)
	elapsedInMs := slotInMs - tail.Timestamp()*SecondInMs
	consensusState, err := tail.WorldState().NextConsensusState(elapsedInMs / SecondInMs)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"tail":    tail,
			"elapsed": elapsedInMs,
			"err":     err,
		}).Debug("Failed to generate next dynasty context.")
		return nil, ErrGenerateNextConsensusState
	}
	if consensusState.Proposer() == nil || !consensusState.Proposer().Equals(dpos.miner.Bytes()) {
		proposer := "nil"
		if consensusState.Proposer() != nil {
			proposer = consensusState.Proposer().Base58()
		}
		logging.VLog().WithFields(logrus.Fields{
			"tail":     tail,
			"now":      nowInMs,
			"slot":     slotInMs,
			"expected": proposer,
			"actual":   dpos.miner,
		}).Debug("Not my turn, waiting...")
		return nil, ErrInvalidBlockProposer
	}
	return consensusState, nil
}

func (dpos *Dpos) pushAndBroadcast(tail *core.Block, block *core.Block) error {
	if err := dpos.chain.BlockPool().PushAndBroadcast(block); err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"tail":  tail,
			"block": block,
			"err":   err,
		}).Error("Failed to push new minted block into block pool")
		return err
	}

	if !dpos.chain.TailBlock().Hash().Equals(block.Hash()) {
		return ErrAppendNewBlockFailed
	}

	logging.CLog().WithFields(logrus.Fields{
		"tail":  tail,
		"block": block,
	}).Info("Broadcasted new block")
	return nil
}

func (dpos *Dpos) mintBlock(now int64) error {
	metricsBlockPackingTime.Update(0)
	metricsBlockWaitingTime.Update(0)

	nowInMs := now * SecondInMs
	// check mining enable
	if !dpos.enable {
		return ErrCannotMintWhenDisable
	}

	// check mining pending
	if dpos.pending {
		return ErrCannotMintWhenPending
	}

	tail := dpos.chain.TailBlock()

	deadlineInMs, err := dpos.checkDeadline(tail, nowInMs)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"tail": tail,
			"now":  nowInMs,
			"err":  err,
		}).Debug("checkDeadline")
		return err
	}

	consensusState, err := dpos.checkProposer(tail, nowInMs)
	if err != nil {
		return err
	}

	miner := "nil"
	if dpos.miner != nil {
		miner = dpos.miner.String()
	}
	logging.CLog().WithFields(logrus.Fields{
		"tail":     tail,
		"start":    nowInMs,
		"deadline": deadlineInMs,
		"expected": consensusState.Proposer().Hex(),
		"actual":   miner,
	}).Info("My turn to mint block")
	metricsBlockPackingTime.Update(deadlineInMs - nowInMs)

	block, err := dpos.newBlock(tail, consensusState, deadlineInMs)
	if err != nil {
		return err
	}

	slotInMs := nextSlot(nowInMs)
	currentInMs := time.Now().Unix() * SecondInMs
	if slotInMs > currentInMs {
		timer := time.NewTimer(time.Duration(slotInMs-currentInMs) * time.Millisecond).C
		<-timer
		metricsBlockWaitingTime.Update(slotInMs - currentInMs)
	}

	logging.CLog().WithFields(logrus.Fields{
		"tail":     tail,
		"block":    block,
		"start":    nowInMs,
		"packed":   currentInMs,
		"deadline": deadlineInMs,
		"slot":     slotInMs,
		"end":      time.Now().Unix(),
	}).Info("Minted new block")

	metricsMintBlock.Inc(1)
	// try to push the new block on chain
	// if failed, return all txs back

	if err := dpos.pushAndBroadcast(tail, block); err != nil {
		go block.ReturnTransactions()
		return err
	}

	return nil
}

func (dpos *Dpos) blockLoop() {
	logging.CLog().Info("Started Dpos Mining.")
	timeChan := time.NewTicker(time.Second).C
	for { // ToRefine: change loop logic, try more times second
		select {
		case now := <-timeChan:
			metricsLruPoolSlotBlock.Update(int64(dpos.slot.Len()))
			dpos.mintBlock(now.Unix())
		case <-dpos.quitCh:
			logging.CLog().Info("Stopped Dpos Mining.")
			return
		}
	}
}

func (dpos *Dpos) findProposer(now int64) (proposer byteutils.Hash, err error) {
	miners, err := dpos.chain.TailBlock().WorldState().Dynasty()
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Debug("Failed to get miners from dynasty.")
		return nil, err
	}
	proposer, err = FindProposer(now, miners)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"proposer": proposer,
			"err":      err,
		}).Debug("Failed to find proposer.")
		return nil, err
	}
	return proposer, nil
}

// NumberOfBlocksInDynasty number of blocks in one dynasty
func (dpos *Dpos) NumberOfBlocksInDynasty() uint64 {
	return uint64(DynastyIntervalInMs) / uint64(BlockIntervalInMs)
}
