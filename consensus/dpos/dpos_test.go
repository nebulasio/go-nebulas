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
	"os"
	"runtime"
	"testing"

	"github.com/nebulasio/go-nebulas/util"

	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/account"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/stretchr/testify/assert"
)

type Neb struct {
	config    *nebletpb.Config
	chain     *core.BlockChain
	ns        net.Service
	am        *account.Manager
	genesis   *corepb.Genesis
	storage   storage.Storage
	consensus core.Consensus
	emitter   *core.EventEmitter
}

func mockNeb(t *testing.T) *Neb {
	storage, _ := storage.NewDiskStorage("test.db")
	//storage, _ := storage.NewMemoryStorage()
	eventEmitter := core.NewEventEmitter(1024)
	genesisConf := MockGenesisConf()
	dpos := NewDpos()
	neb := &Neb{
		genesis:   genesisConf,
		storage:   storage,
		emitter:   eventEmitter,
		consensus: dpos,
		config: &nebletpb.Config{
			Chain: &nebletpb.ChainConfig{
				ChainId:    genesisConf.Meta.ChainId,
				Keydir:     "keydir",
				Coinbase:   "1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c",
				Miner:      "1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c",
				Passphrase: "passphrase",
			},
		},
	}

	neb.emitter.Start()

	am := account.NewManager(neb)
	neb.am = am

	chain, err := core.NewBlockChain(neb)
	assert.Nil(t, err)
	neb.chain = chain
	err = dpos.Setup(neb)
	assert.Nil(t, err)
	err = chain.Setup(neb)
	assert.Nil(t, err)

	var ns mockNetService
	neb.ns = ns
	neb.chain.BlockPool().RegisterInNetwork(ns)
	return neb
}

func (n *Neb) Config() *nebletpb.Config {
	return n.config
}

func (n *Neb) BlockChain() *core.BlockChain {
	return n.chain
}

func (n *Neb) NetService() net.Service {
	return n.ns
}

func (n *Neb) AccountManager() core.Manager {
	return n.am
}

func (n *Neb) Genesis() *corepb.Genesis {
	return n.genesis
}

func (n *Neb) Storage() storage.Storage {
	return n.storage
}

func (n *Neb) EventEmitter() *core.EventEmitter {
	return n.emitter
}

func (n *Neb) Consensus() core.Consensus {
	return n.consensus
}

func (n *Neb) StartActiveSync() {}

var (
	DefaultOpenDynasty = []string{
		"1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c",
		"2fe3f9f51f9a05dd5f7c5329127f7c917917149b4e16b0b8",
		"333cb3ed8c417971845382ede3cf67a0a96270c05fe2f700",
		"48f981ed38910f1232c1bab124f650c482a57271632db9e3",
		"59fc526072b09af8a8ca9732dae17132c4e9127e43cf2232",
		"75e4e5a71d647298b88928d8cb5da43d90ab1a6c52d0905f",
	}
)

// MockGenesisConf return mock genesis conf
func MockGenesisConf() *corepb.Genesis {
	return &corepb.Genesis{
		Meta: &corepb.GenesisMeta{ChainId: 0},
		Consensus: &corepb.GenesisConsensus{
			Dpos: &corepb.GenesisConsensusDpos{
				Dynasty: DefaultOpenDynasty,
			},
		},
		TokenDistribution: []*corepb.GenesisTokenDistribution{
			&corepb.GenesisTokenDistribution{
				Address: "1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c",
				Value:   "10000000000000000000000",
			},
			&corepb.GenesisTokenDistribution{
				Address: "2fe3f9f51f9a05dd5f7c5329127f7c917917149b4e16b0b8",
				Value:   "10000000000000000000000",
			},
			&corepb.GenesisTokenDistribution{
				Address: "333cb3ed8c417971845382ede3cf67a0a96270c05fe2f700",
				Value:   "10000000000000000000000",
			},
			&corepb.GenesisTokenDistribution{
				Address: "48f981ed38910f1232c1bab124f650c482a57271632db9e3",
				Value:   "10000000000000000000000",
			},
			&corepb.GenesisTokenDistribution{
				Address: "59fc526072b09af8a8ca9732dae17132c4e9127e43cf2232",
				Value:   "10000000000000000000000",
			},
			&corepb.GenesisTokenDistribution{
				Address: "c79d9667c71bb09d6ca7c3ed12bfe5e7be24e2ffe13a833d",
				Value:   "10000000000000000000000",
			},
			&corepb.GenesisTokenDistribution{
				Address: "a8f1f53952c535c6600c77cf92b65e0c9b64496a8a328569",
				Value:   "10000000000000000000000",
			},
			&corepb.GenesisTokenDistribution{
				Address: "98a3eed687640b75ec55bf5c9e284371bdcaeab943524d51",
				Value:   "10000000000000000000000",
			},
			&corepb.GenesisTokenDistribution{
				Address: "75e4e5a71d647298b88928d8cb5da43d90ab1a6c52d0905f",
				Value:   "10000000000000000000000",
			},
		},
	}
}

var (
	received = []byte{}
)

type mockNetService struct{}

func (n mockNetService) Start() error { return nil }
func (n mockNetService) Stop()        {}

func (n mockNetService) Node() *net.Node { return nil }

func (n mockNetService) Sync(net.Serializable) error { return nil }

func (n mockNetService) Register(...*net.Subscriber)   {}
func (n mockNetService) Deregister(...*net.Subscriber) {}

func (n mockNetService) Broadcast(name string, msg net.Serializable, priority int) {
	pb, _ := msg.ToProto()
	bytes, _ := proto.Marshal(pb)
	received = bytes
}
func (n mockNetService) Relay(name string, msg net.Serializable, priority int) {
	pb, _ := msg.ToProto()
	bytes, _ := proto.Marshal(pb)
	received = bytes
}
func (n mockNetService) SendMsg(name string, msg []byte, target string, priority int) error {
	received = msg
	return nil
}

func (n mockNetService) SendMessageToPeers(messageName string, data []byte, priority int, filter net.PeerFilterAlgorithm) []string {
	return make([]string, 0)
}
func (n mockNetService) SendMessageToPeer(messageName string, data []byte, priority int, peerID string) error {
	return nil
}

func (n mockNetService) ClosePeer(peerID string, reason error) {}

func (n mockNetService) BroadcastNetworkID([]byte) {}

func (n mockNetService) BuildRawMessageData([]byte, string) []byte { return nil }

func mockBlockFromNetwork(block *core.Block) (*core.Block, error) {
	pbBlock, err := block.ToProto()
	if err != nil {
		return nil, err
	}
	bytes, err := proto.Marshal(pbBlock)
	if err := proto.Unmarshal(bytes, pbBlock); err != nil {
		return nil, err
	}
	block = new(core.Block)
	block.FromProto(pbBlock)
	return block, nil
}

func TestDpos_New(t *testing.T) {
	neb := mockNeb(t)
	coinbase := neb.config.Chain.Coinbase
	neb.config.Chain.Coinbase += "0"
	assert.NotNil(t, neb.Consensus().Setup(neb))
	neb.config.Chain.Coinbase = coinbase
	neb.config.Chain.Miner += "0"
	assert.NotNil(t, neb.Consensus().Setup(neb))
}

func TestDpos_VerifySign(t *testing.T) {
	neb := mockNeb(t)
	tail := neb.chain.TailBlock()

	elapsedSecond := int64(DynastySize*BlockInterval + DynastyInterval)
	consensusState, err := tail.WorldState().NextConsensusState(elapsedSecond)
	assert.Nil(t, err)
	coinbase, err := core.AddressParse("1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c")
	assert.Nil(t, err)
	block, err := core.NewBlock(neb.chain.ChainID(), coinbase, tail)
	assert.Nil(t, err)
	block.SetTimestamp(DynastySize*BlockInterval + DynastyInterval)
	block.WorldState().SetConsensusState(consensusState)
	block.SetMiner(coinbase)
	block.Seal()
	manager := account.NewManager(nil)
	miner, err := core.AddressParseFromBytes(consensusState.Proposer())
	assert.Nil(t, err)
	assert.Nil(t, manager.Unlock(miner, []byte("passphrase"), keystore.DefaultUnlockDuration))
	assert.Nil(t, manager.SignBlock(miner, block))
	assert.Nil(t, neb.consensus.VerifyBlock(block))

	miner, err = core.AddressParse("fc751b484bd5296f8d267a8537d33f25a848f7f7af8cfcf6")
	assert.Nil(t, err)
	assert.Nil(t, manager.Unlock(miner, []byte("passphrase"), keystore.DefaultUnlockDuration))
	assert.Nil(t, manager.SignBlock(miner, block))
	assert.Equal(t, neb.consensus.VerifyBlock(block), ErrInvalidBlockProposer)
}

func GetUnlockAddress(t *testing.T, am *account.Manager, addr string) *core.Address {
	address, err := core.AddressParse(addr)
	assert.Nil(t, err)
	assert.Nil(t, am.Unlock(address, []byte("passphrase"), time.Second*60*60*24*365))
	return address
}

func TestForkChoice(t *testing.T) {
	neb := mockNeb(t)
	am := account.NewManager(neb)

	/*
		genesis -- 0 -- 11 -- 111 -- 1111
					 \_ 12 -- 221
	*/

	addr0 := GetUnlockAddress(t, am, "2fe3f9f51f9a05dd5f7c5329127f7c917917149b4e16b0b8")
	block0, _ := neb.chain.NewBlock(addr0)
	block0.SetTimestamp(BlockInterval)
	consensusState, err := neb.BlockChain().TailBlock().WorldState().NextConsensusState(BlockInterval)
	assert.Nil(t, err)
	block0.WorldState().SetConsensusState(consensusState)
	block0.SetMiner(addr0)
	block0.Seal()
	am.SignBlock(addr0, block0)
	assert.Nil(t, neb.chain.BlockPool().Push(block0))
	assert.Equal(t, len(neb.chain.DetachedTailBlocks()), 1)
	assert.Equal(t, block0.Hash(), neb.chain.TailBlock().Hash())

	addr1 := GetUnlockAddress(t, am, "333cb3ed8c417971845382ede3cf67a0a96270c05fe2f700")
	block11, err := neb.chain.NewBlock(addr1)
	assert.Nil(t, err)
	consensusState, err = block0.WorldState().NextConsensusState(BlockInterval)
	assert.Nil(t, err)
	block11.WorldState().SetConsensusState(consensusState)
	state := consensusState.(*State)
	state.TimeStamp()
	block11.SetTimestamp(BlockInterval * 2)
	block11.SetMiner(addr1)
	block11.Seal()
	am.SignBlock(addr1, block11)
	assert.Nil(t, neb.chain.BlockPool().Push(block11))
	assert.Equal(t, len(neb.chain.DetachedTailBlocks()), 1)
	assert.Equal(t, block11.Hash(), neb.chain.TailBlock().Hash())

	addr2 := GetUnlockAddress(t, am, "48f981ed38910f1232c1bab124f650c482a57271632db9e3")
	block12, err := neb.chain.NewBlockFromParent(addr2, block0)
	assert.Nil(t, err)
	consensusState, err = block0.WorldState().NextConsensusState(BlockInterval * 2)
	assert.Nil(t, err)
	block12.WorldState().SetConsensusState(consensusState)
	block12.SetTimestamp(BlockInterval * 3)
	block12.SetMiner(addr2)
	block12.Seal()
	am.SignBlock(addr2, block12)
	assert.Nil(t, neb.chain.BlockPool().Push(block12))
	assert.Equal(t, len(neb.chain.DetachedTailBlocks()), 2)
	tail := block11.Hash()
	if less(block11, block12) {
		tail = block12.Hash()
	}
	assert.Equal(t, neb.chain.TailBlock().Hash(), tail)

	addr3 := GetUnlockAddress(t, am, "59fc526072b09af8a8ca9732dae17132c4e9127e43cf2232")
	block111, err := neb.chain.NewBlockFromParent(addr3, block11)
	assert.Nil(t, err)
	consensusState, err = block11.WorldState().NextConsensusState(BlockInterval * 2)
	assert.Nil(t, err)
	block111.WorldState().SetConsensusState(consensusState)
	block111.SetTimestamp(BlockInterval * 4)
	block111.SetMiner(addr3)
	block111.Seal()
	am.SignBlock(addr3, block111)
	assert.Equal(t, len(neb.chain.DetachedTailBlocks()), 2)
	assert.Nil(t, neb.chain.BlockPool().Push(block111))

}

func TestCanMining(t *testing.T) {
	neb := mockNeb(t)
	assert.Equal(t, neb.consensus.Pending(), true)
	neb.consensus.SuspendMining()
	assert.Equal(t, neb.consensus.Pending(), true)
	neb.consensus.ResumeMining()
	assert.Equal(t, neb.consensus.Pending(), false)
}

func TestVerifyBlock(t *testing.T) {
	neb := mockNeb(t)
	dpos := neb.consensus
	tail := neb.chain.TailBlock()

	coinbase, err := core.AddressParse("1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c")
	assert.Nil(t, err)
	manager := account.NewManager(nil)
	assert.Nil(t, dpos.EnableMining("passphrase"))

	elapsedSecond := int64(DynastyInterval)
	consensusState, err := tail.WorldState().NextConsensusState(elapsedSecond)
	assert.Nil(t, err)
	block, err := core.NewBlock(neb.chain.ChainID(), coinbase, tail)
	block.SetTimestamp(tail.Timestamp() + 1)
	assert.Nil(t, err)
	block.WorldState().SetConsensusState(consensusState)
	block.SetMiner(coinbase)
	block.Seal()
	assert.Nil(t, manager.SignBlock(coinbase, block))
	assert.NotNil(t, dpos.VerifyBlock(block), ErrInvalidBlockInterval)

	elapsedSecond = int64(DynastyInterval)
	consensusState, err = tail.WorldState().NextConsensusState(elapsedSecond)
	block, err = core.NewBlock(neb.chain.ChainID(), coinbase, tail)
	assert.Nil(t, err)
	block.WorldState().SetConsensusState(consensusState)
	block.SetMiner(coinbase)
	block.SetTimestamp(tail.Timestamp() + elapsedSecond)
	block.Seal()
	assert.Nil(t, manager.SignBlock(coinbase, block))
	assert.Nil(t, dpos.VerifyBlock(block))

	elapsedSecond = int64(DynastySize*BlockInterval + DynastyInterval)
	consensusState, err = tail.WorldState().NextConsensusState(elapsedSecond)
	block, err = core.NewBlock(neb.chain.ChainID(), coinbase, tail)
	assert.Nil(t, err)
	block.WorldState().SetConsensusState(consensusState)
	block.SetMiner(coinbase)
	block.SetTimestamp(tail.Timestamp() + elapsedSecond)
	block.Seal()
	assert.Nil(t, manager.SignBlock(coinbase, block))
	assert.Nil(t, dpos.VerifyBlock(block))
}

func TestDpos_MintBlock(t *testing.T) {
	neb := mockNeb(t)
	dpos := neb.consensus.(*Dpos)

	assert.Equal(t, dpos.mintBlock(0), ErrCannotMintWhenDiable)

	assert.Nil(t, dpos.EnableMining("passphrase"))
	dpos.SuspendMining()
	assert.Equal(t, dpos.mintBlock(0), ErrCannotMintWhenPending)

	dpos.ResumeMining()
	assert.Equal(t, dpos.mintBlock(BlockInterval), ErrInvalidBlockProposer)

	received = []byte{}
	assert.Equal(t, dpos.mintBlock(DynastyInterval), nil)
	assert.NotEqual(t, received, []byte{})
}

func TestDposContracts(t *testing.T) {
	// change cwd make lib accessible.
	os.Chdir("../../")
	runtime.GOMAXPROCS(runtime.NumCPU())
	neb := mockNeb(t)
	tail := neb.chain.TailBlock()
	dpos := neb.consensus

	coinbase, err := core.AddressParse("1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c")

	assert.Nil(t, err)

	elapsedSecond := int64(DynastyInterval)
	consensusState, err := tail.WorldState().NextConsensusState(elapsedSecond)
	assert.Nil(t, err)
	block, err := core.NewBlock(neb.chain.ChainID(), coinbase, tail)
	assert.Nil(t, err)
	block.SetTimestamp(consensusState.TimeStamp())
	block.WorldState().SetConsensusState(consensusState)

	manager := account.NewManager(nil)
	assert.Nil(t, dpos.EnableMining("passphrase"))

	/* 	accs := []string{
		"2fe3f9f51f9a05dd5f7c5329127f7c917917149b4e16b0b8",
		"333cb3ed8c417971845382ede3cf67a0a96270c05fe2f700",
		"48f981ed38910f1232c1bab124f650c482a57271632db9e3",
		"59fc526072b09af8a8ca9732dae17132c4e9127e43cf2232",
		"75e4e5a71d647298b88928d8cb5da43d90ab1a6c52d0905f",
		"98a3eed687640b75ec55bf5c9e284371bdcaeab943524d51",
		"a8f1f53952c535c6600c77cf92b65e0c9b64496a8a328569",
		"c79d9667c71bb09d6ca7c3ed12bfe5e7be24e2ffe13a833d",
	} */

	a, _ := core.AddressParse("2fe3f9f51f9a05dd5f7c5329127f7c917917149b4e16b0b8")
	assert.Nil(t, manager.Unlock(a, []byte("passphrase"), keystore.YearUnlockDuration))
	b, _ := core.AddressParse("333cb3ed8c417971845382ede3cf67a0a96270c05fe2f700")
	assert.Nil(t, manager.Unlock(b, []byte("passphrase"), keystore.YearUnlockDuration))
	c, _ := core.AddressParse("48f981ed38910f1232c1bab124f650c482a57271632db9e3")
	assert.Nil(t, manager.Unlock(c, []byte("passphrase"), keystore.YearUnlockDuration))
	d, _ := core.AddressParse("59fc526072b09af8a8ca9732dae17132c4e9127e43cf2232")
	assert.Nil(t, manager.Unlock(d, []byte("passphrase"), keystore.YearUnlockDuration))
	e, _ := core.AddressParse("75e4e5a71d647298b88928d8cb5da43d90ab1a6c52d0905f")
	assert.Nil(t, manager.Unlock(e, []byte("passphrase"), keystore.YearUnlockDuration))
	f, _ := core.AddressParse("98a3eed687640b75ec55bf5c9e284371bdcaeab943524d51")
	assert.Nil(t, manager.Unlock(f, []byte("passphrase"), keystore.YearUnlockDuration))
	g, _ := core.AddressParse("a8f1f53952c535c6600c77cf92b65e0c9b64496a8a328569")
	assert.Nil(t, manager.Unlock(g, []byte("passphrase"), keystore.YearUnlockDuration))
	h, _ := core.AddressParse("c79d9667c71bb09d6ca7c3ed12bfe5e7be24e2ffe13a833d")
	assert.Nil(t, manager.Unlock(h, []byte("passphrase"), keystore.YearUnlockDuration))

	source := `"use strict";var DepositeContent=function(text){if(text){var o=JSON.parse(text);this.balance=new BigNumber(o.balance);this.expiryHeight=new BigNumber(o.expiryHeight)}else{this.balance=new BigNumber(0);this.expiryHeight=new BigNumber(0)}};DepositeContent.prototype={toString:function(){return JSON.stringify(this)}};var BankVaultContract=function(){LocalContractStorage.defineMapProperty(this,"bankVault",{parse:function(text){return new DepositeContent(text)},stringify:function(o){return o.toString()}})};BankVaultContract.prototype={init:function(){},save:function(height){var from=Blockchain.transaction.from;var value=Blockchain.transaction.value;var bk_height=new BigNumber(Blockchain.block.height);var orig_deposit=this.bankVault.get(from);if(orig_deposit){value=value.plus(orig_deposit.balance)}var deposit=new DepositeContent();deposit.balance=value;deposit.expiryHeight=bk_height.plus(height);this.bankVault.put(from,deposit)},takeout:function(value){var from=Blockchain.transaction.from;var bk_height=new BigNumber(Blockchain.block.height);var amount=new BigNumber(value);var deposit=this.bankVault.get(from);if(!deposit){throw new Error("No deposit before.")}if(bk_height.lt(deposit.expiryHeight)){throw new Error("Can not takeout before expiryHeight.")}if(amount.gt(deposit.balance)){throw new Error("Insufficient balance.")}var result=Blockchain.transfer(from,amount);if(result!=0){throw new Error("transfer failed.")}Event.Trigger("BankVault",{Transfer:{from:Blockchain.transaction.to,to:from,value:amount.toString()}});deposit.balance=deposit.balance.sub(amount);this.bankVault.put(from,deposit)},balanceOf:function(){var from=Blockchain.transaction.from;return this.bankVault.get(from)}};module.exports=BankVaultContract;`
	sourceType := "js"
	argsDeploy := ""
	payloadDeploy, _ := core.NewDeployPayload(source, sourceType, argsDeploy).ToBytes()

	j := 3

	for i := 1; i < j; i++ {
		value, _ := util.NewUint128FromInt(1)
		gasLimit, _ := util.NewUint128FromInt(200000)
		txDeploy := core.NewTransaction(neb.chain.ChainID(), a, a, value, uint64(i), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
		assert.Nil(t, manager.SignTransaction(a, txDeploy))
		assert.Nil(t, neb.chain.TransactionPool().Push(txDeploy))

		txDeploy = core.NewTransaction(neb.chain.ChainID(), b, b, value, uint64(i), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
		assert.Nil(t, manager.SignTransaction(b, txDeploy))
		assert.Nil(t, neb.chain.TransactionPool().Push(txDeploy))

		txDeploy = core.NewTransaction(neb.chain.ChainID(), c, c, value, uint64(i), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
		assert.Nil(t, manager.SignTransaction(c, txDeploy))
		assert.Nil(t, neb.chain.TransactionPool().Push(txDeploy))

		txDeploy = core.NewTransaction(neb.chain.ChainID(), d, d, value, uint64(i), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
		assert.Nil(t, manager.SignTransaction(d, txDeploy))
		assert.Nil(t, neb.chain.TransactionPool().Push(txDeploy))

		txDeploy = core.NewTransaction(neb.chain.ChainID(), e, e, value, uint64(i), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
		assert.Nil(t, manager.SignTransaction(e, txDeploy))
		assert.Nil(t, neb.chain.TransactionPool().Push(txDeploy))

		txDeploy = core.NewTransaction(neb.chain.ChainID(), f, f, value, uint64(i), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
		assert.Nil(t, manager.SignTransaction(f, txDeploy))
		assert.Nil(t, neb.chain.TransactionPool().Push(txDeploy))

		txDeploy = core.NewTransaction(neb.chain.ChainID(), g, g, value, uint64(i), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
		assert.Nil(t, manager.SignTransaction(g, txDeploy))
		assert.Nil(t, neb.chain.TransactionPool().Push(txDeploy))

		txDeploy = core.NewTransaction(neb.chain.ChainID(), h, h, value, uint64(i), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
		assert.Nil(t, manager.SignTransaction(h, txDeploy))
		assert.Nil(t, neb.chain.TransactionPool().Push(txDeploy))
		/*
			function := "save"
			argsCall := "[1]"
			payloadCall, _ := core.NewCallPayload(function, argsCall).ToBytes()
			txCall := core.NewTransaction(neb.chain.ChainID(), a, a, value, 1, core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
			assert.Nil(t, manager.SignTransaction(a, txCall))
			assert.Nil(t, neb.chain.TransactionPool().Push(txCall))
		*/
	}

	block.CollectTransactions(time.Now().Unix() + 1)
	assert.Equal(t, 8*(j-1), len(block.Transactions()))
	block.SetMiner(coinbase)
	assert.Nil(t, block.Seal())
	assert.Nil(t, manager.SignBlock(coinbase, block))
	assert.Nil(t, neb.chain.BlockPool().Push(block))

	assert.Equal(t, block.Hash(), neb.chain.TailBlock().Hash())
}

func testMintBlock(t *testing.T, neb *Neb, num int) {
	manager := account.NewManager(nil)

	coinbase, err := core.AddressParse("1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c")
	assert.Nil(t, manager.Unlock(coinbase, []byte("passphrase"), keystore.YearUnlockDuration))

	a, _ := core.AddressParse("2fe3f9f51f9a05dd5f7c5329127f7c917917149b4e16b0b8")
	assert.Nil(t, manager.Unlock(a, []byte("passphrase"), keystore.YearUnlockDuration))
	b, _ := core.AddressParse("333cb3ed8c417971845382ede3cf67a0a96270c05fe2f700")
	assert.Nil(t, manager.Unlock(b, []byte("passphrase"), keystore.YearUnlockDuration))
	c, _ := core.AddressParse("48f981ed38910f1232c1bab124f650c482a57271632db9e3")
	assert.Nil(t, manager.Unlock(c, []byte("passphrase"), keystore.YearUnlockDuration))
	d, _ := core.AddressParse("59fc526072b09af8a8ca9732dae17132c4e9127e43cf2232")
	assert.Nil(t, manager.Unlock(d, []byte("passphrase"), keystore.YearUnlockDuration))
	e, _ := core.AddressParse("7da9dabedb4c6e121146fb4250a9883d6180570e63d6b080")
	assert.Nil(t, manager.Unlock(e, []byte("passphrase"), keystore.YearUnlockDuration))

	//f, _ := core.AddressParse("b040353ec0f2c113d5639444f7253681aecda1f8b91f179f")
	//assert.Nil(t, manager.Unlock(f, []byte("passphrase"), keystore.YearUnlockDuration))

	//g, _ := core.AddressParse("b414432e15f21237013017fa6ee90fc99433dec82c1c8370")
	//h, _ := core.AddressParse("b49f30d0e5c9c88cade54cd1adecf6bc2c7e0e5af646d903")
	//m, _ := core.AddressParse("fc751b484bd5296f8d267a8537d33f25a848f7f7af8cfcf6")

	elapsedSecond := int64(DynastyInterval)
	consensusState, err := neb.chain.TailBlock().WorldState().NextConsensusState(elapsedSecond)
	assert.Nil(t, err)

	block, err := core.NewBlock(neb.chain.ChainID(), coinbase, neb.chain.TailBlock())
	assert.Nil(t, err)
	block.SetTimestamp(consensusState.TimeStamp())
	block.WorldState().SetConsensusState(consensusState)
	acc, _ := block.WorldState().GetOrCreateUserAccount(a.Bytes())
	nonce := int(acc.Nonce())

	for i := 1; i < num; i++ {
		gas, _ := util.NewUint128FromInt(1000000)
		limit, _ := util.NewUint128FromInt(200000)
		tx := core.NewTransaction(neb.chain.ChainID(), a, b, util.NewUint128(), uint64(nonce+4*i-3), core.TxPayloadBinaryType, []byte("nas"), gas, limit)
		assert.Nil(t, manager.SignTransaction(a, tx))
		assert.Nil(t, neb.chain.TransactionPool().Push(tx))

		tx = core.NewTransaction(neb.chain.ChainID(), a, c, util.NewUint128(), uint64(nonce+4*i-2), core.TxPayloadBinaryType, []byte("nas"), gas, limit)
		assert.Nil(t, manager.SignTransaction(a, tx))
		assert.Nil(t, neb.chain.TransactionPool().Push(tx))

		tx = core.NewTransaction(neb.chain.ChainID(), a, d, util.NewUint128(), uint64(nonce+4*i-1), core.TxPayloadBinaryType, []byte("nas"), gas, limit)
		assert.Nil(t, manager.SignTransaction(a, tx))
		assert.Nil(t, neb.chain.TransactionPool().Push(tx))
		//assert.Equal(t, neb.chain.TransactionPool().cache.Len(), 3)

		tx = core.NewTransaction(neb.chain.ChainID(), a, e, util.NewUint128(), uint64(nonce+4*i), core.TxPayloadBinaryType, []byte("nas"), gas, limit)
		assert.Nil(t, manager.SignTransaction(a, tx))
		assert.Nil(t, neb.chain.TransactionPool().Push(tx))

		//tx = core.NewTransaction(neb.chain.ChainID(), e, h, util.NewUint128(), 2, core.TxPayloadBinaryType, []byte("nas"), gas, limit)
		//assert.Nil(t, manager.SignTransaction(e, tx))
		//assert.Nil(t, neb.chain.TransactionPool().Push(tx))
	}

	block.CollectTransactions(time.Now().Unix() + 1)
	assert.Equal(t, 4*(num-1), len(block.Transactions()))
	block.SetMiner(coinbase)
	assert.Nil(t, block.Seal())
	assert.Nil(t, manager.SignBlock(coinbase, block))
	assert.Nil(t, neb.chain.BlockPool().Push(block))

	assert.Equal(t, block.Hash(), neb.chain.TailBlock().Hash())
}

func TestDposTxBinary(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	neb := mockNeb(t)

	for i := 0; i < 10; i++ {
		testMintBlock(t, neb, 100)
	}

	return
}
