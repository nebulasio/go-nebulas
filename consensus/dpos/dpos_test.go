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
	config  *nebletpb.Config
	chain   *core.BlockChain
	ns      net.Service
	am      *account.Manager
	genesis *corepb.Genesis
	storage storage.Storage
	emitter *core.EventEmitter
}

func mockNeb(t *testing.T) *Neb {
	storage, _ := storage.NewMemoryStorage()
	eventEmitter := core.NewEventEmitter(1024)
	genesisConf := MockGenesisConf()
	neb := &Neb{
		genesis: genesisConf,
		storage: storage,
		emitter: eventEmitter,
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
	am := account.NewManager(neb)
	var ns MockNetService
	chain, err := core.NewBlockChain(neb)
	assert.Nil(t, err)
	neb.chain = chain
	neb.am = am
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

func (n *Neb) AccountManager() *account.Manager {
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

func (n *Neb) StartActiveSync() {}

var (
	DefaultOpenDynasty = []string{
		"1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c",
		"2fe3f9f51f9a05dd5f7c5329127f7c917917149b4e16b0b8",
		"333cb3ed8c417971845382ede3cf67a0a96270c05fe2f700",
		"48f981ed38910f1232c1bab124f650c482a57271632db9e3",
		"59fc526072b09af8a8ca9732dae17132c4e9127e43cf2232",
		"75e4e5a71d647298b88928d8cb5da43d90ab1a6c52d0905f",
		"7da9dabedb4c6e121146fb4250a9883d6180570e63d6b080",
		"a8f1f53952c535c6600c77cf92b65e0c9b64496a8a328569",
		"b040353ec0f2c113d5639444f7253681aecda1f8b91f179f",
		"b414432e15f21237013017fa6ee90fc99433dec82c1c8370",
		"b49f30d0e5c9c88cade54cd1adecf6bc2c7e0e5af646d903",
		"b7d83b44a3719720ec54cdb9f54c0202de68f1ebcb927b4f",
		"ba56cc452e450551b7b9cffe25084a069e8c1e94412aad22",
		"c5bcfcb3fa8250be4f2bf2b1e70e1da500c668377ba8cd4a",
		"c79d9667c71bb09d6ca7c3ed12bfe5e7be24e2ffe13a833d",
		"d1abde197e97398864ba74511f02832726edad596775420a",
		"d86f99d97a394fa7a623fdf84fdc7446b99c3cb335fca4bf",
		"e0f78b011e639ce6d8b76f97712118f3fe4a12dd954eba49",
		"f38db3b6c801dddd624d6ddc2088aa64b5a24936619e4848",
		"fc751b484bd5296f8d267a8537d33f25a848f7f7af8cfcf6",
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
		},
	}
}

var (
	received = []byte{}
)

type MockNetService struct{}

func (n MockNetService) Start() error { return nil }
func (n MockNetService) Stop()        {}

func (n MockNetService) Node() *net.Node { return nil }

func (n MockNetService) Sync(net.Serializable) error { return nil }

func (n MockNetService) Register(...*net.Subscriber)   {}
func (n MockNetService) Deregister(...*net.Subscriber) {}

func (n MockNetService) Broadcast(name string, msg net.Serializable, priority int) {
	pb, _ := msg.ToProto()
	bytes, _ := proto.Marshal(pb)
	received = bytes
}
func (n MockNetService) Relay(name string, msg net.Serializable, priority int) {
	pb, _ := msg.ToProto()
	bytes, _ := proto.Marshal(pb)
	received = bytes
}
func (n MockNetService) SendMsg(name string, msg []byte, target string, priority int) error {
	received = msg
	return nil
}

func (n MockNetService) SendMessageToPeers(messageName string, data []byte, priority int, filter net.PeerFilterAlgorithm) []string {
	return make([]string, 0)
}
func (n MockNetService) SendMessageToPeer(messageName string, data []byte, priority int, peerID string) error {
	return nil
}

func (n MockNetService) ClosePeer(peerID string, reason error) {}

func (n MockNetService) BroadcastNetworkID([]byte) {}

func (n MockNetService) BuildRawMessageData([]byte, string) []byte { return nil }

func TestDpos_New(t *testing.T) {
	neb := mockNeb(t)
	_, err := NewDpos(neb)
	assert.Nil(t, err)
	coinbase := neb.config.Chain.Coinbase
	neb.config.Chain.Coinbase += "0"
	_, err = NewDpos(neb)
	assert.NotNil(t, err)
	neb.config.Chain.Coinbase = coinbase
	neb.config.Chain.Miner += "0"
	_, err = NewDpos(neb)
	assert.NotNil(t, err)
}

func TestDpos_VerifySign(t *testing.T) {
	dpos, err := NewDpos(mockNeb(t))
	assert.Nil(t, err)
	dpos.chain.SetConsensusHandler(dpos)
	tail := dpos.chain.TailBlock()

	elapsedSecond := int64(core.DynastySize*core.BlockInterval + core.DynastyInterval)
	context, err := tail.NextDynastyContext(dpos.chain, elapsedSecond)
	assert.Nil(t, err)
	coinbase, err := core.AddressParse("1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c")
	assert.Nil(t, err)
	block, err := core.NewBlock(dpos.chain.ChainID(), coinbase, tail)
	assert.Nil(t, err)
	block.LoadDynastyContext(context)
	block.SetMiner(coinbase)
	block.Seal()
	manager := account.NewManager(nil)
	miner, err := core.AddressParseFromBytes(context.Proposer)
	assert.Nil(t, err)
	assert.Nil(t, manager.Unlock(miner, []byte("passphrase"), keystore.DefaultUnlockDuration))
	assert.Nil(t, manager.SignBlock(miner, block))
	assert.Nil(t, dpos.VerifyBlock(block, tail))

	miner, err = core.AddressParse("fc751b484bd5296f8d267a8537d33f25a848f7f7af8cfcf6")
	assert.Nil(t, err)
	assert.Nil(t, manager.Unlock(miner, []byte("passphrase"), keystore.DefaultUnlockDuration))
	assert.Nil(t, manager.SignBlock(miner, block))
	assert.Equal(t, dpos.VerifyBlock(block, tail), ErrInvalidBlockProposer)
}

func GetUnlockAddress(t *testing.T, am *account.Manager, addr string) *core.Address {
	address, err := core.AddressParse(addr)
	assert.Nil(t, err)
	assert.Nil(t, am.Unlock(address, []byte("passphrase"), time.Second*60*60*24*365))
	return address
}

func TestForkChoice(t *testing.T) {
	neb := mockNeb(t)
	dpos, err := NewDpos(neb)
	assert.Nil(t, err)
	dpos.chain.SetConsensusHandler(dpos)

	am := account.NewManager(neb)

	/*
		genesis -- 0 -- 11 -- 111 -- 1111
					 \_ 12 -- 221
	*/

	addr0 := GetUnlockAddress(t, am, "2fe3f9f51f9a05dd5f7c5329127f7c917917149b4e16b0b8")
	block0, _ := dpos.chain.NewBlock(addr0)
	block0.SetTimestamp(core.BlockInterval)
	block0.SetMiner(addr0)
	block0.Seal()
	am.SignBlock(addr0, block0)
	assert.Nil(t, dpos.chain.BlockPool().Push(block0))
	assert.Equal(t, block0.Hash(), dpos.chain.TailBlock().Hash())

	addr1 := GetUnlockAddress(t, am, "333cb3ed8c417971845382ede3cf67a0a96270c05fe2f700")
	block11, err := dpos.chain.NewBlock(addr1)
	assert.Nil(t, err)
	block11.SetTimestamp(core.BlockInterval * 2)
	block11.SetMiner(addr1)
	block11.Seal()
	am.SignBlock(addr1, block11)
	assert.Nil(t, dpos.chain.BlockPool().Push(block11))

	block12, _ := dpos.chain.NewBlock(addr1)
	block12.SetTimestamp(core.BlockInterval * 2)
	block12.SetMiner(addr1)
	block12.Seal()
	am.SignBlock(addr1, block12)
	assert.Error(t, dpos.chain.BlockPool().Push(block12), core.ErrDoubleBlockMinted)

	assert.Equal(t, len(dpos.chain.DetachedTailBlocks()), 1)
	assert.Equal(t, dpos.chain.TailBlock().Hash(), block11.Hash())

	addr2 := GetUnlockAddress(t, am, "48f981ed38910f1232c1bab124f650c482a57271632db9e3")
	block111, _ := dpos.chain.NewBlockFromParent(addr2, block11)
	block111.SetTimestamp(core.BlockInterval * 3)
	block111.SetMiner(addr2)
	block111.Seal()
	am.SignBlock(addr2, block111)
	assert.Equal(t, len(dpos.chain.DetachedTailBlocks()), 1)

	addr3 := GetUnlockAddress(t, am, "59fc526072b09af8a8ca9732dae17132c4e9127e43cf2232")
	block1111, _ := dpos.chain.NewBlockFromParent(addr3, block111)
	block1111.SetTimestamp(core.BlockInterval * 4)
	block1111.SetMiner(addr3)
	block1111.Seal()
	am.SignBlock(addr3, block1111)
	assert.Error(t, dpos.chain.BlockPool().Push(block1111), core.ErrMissingParentBlock)
	assert.Equal(t, len(dpos.chain.DetachedTailBlocks()), 1)
	assert.Nil(t, dpos.chain.BlockPool().Push(block111))
	assert.Equal(t, len(dpos.chain.DetachedTailBlocks()), 1)
	assert.Equal(t, dpos.chain.TailBlock().Hash(), block1111.Hash())
}

func TestCanMining(t *testing.T) {
	dpos, err := NewDpos(mockNeb(t))
	assert.Nil(t, err)
	assert.Equal(t, dpos.Pending(), true)
	dpos.SuspendMining()
	assert.Equal(t, dpos.Pending(), true)
	dpos.ResumeMining()
	assert.Equal(t, dpos.Pending(), false)
}

func TestFastVerifyBlock(t *testing.T) {
	dpos, err := NewDpos(mockNeb(t))
	assert.Nil(t, err)
	dpos.chain.SetConsensusHandler(dpos)
	tail := dpos.chain.TailBlock()

	coinbase, err := core.AddressParse("1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c")
	assert.Nil(t, err)
	manager := account.NewManager(nil)
	assert.Nil(t, dpos.EnableMining("passphrase"))

	elapsedSecond := int64(core.DynastyInterval)
	context, err := tail.NextDynastyContext(dpos.chain, elapsedSecond)
	assert.Nil(t, err)
	block, err := core.NewBlock(dpos.chain.ChainID(), coinbase, tail)
	block.SetTimestamp(block.Timestamp() + 1)
	assert.Nil(t, err)
	block.LoadDynastyContext(context)
	block.SetMiner(coinbase)
	block.Seal()
	assert.Nil(t, manager.SignBlock(coinbase, block))
	assert.Nil(t, dpos.FastVerifyBlock(block))

	elapsedSecond = int64(core.DynastyInterval)
	context, err = tail.NextDynastyContext(dpos.chain, elapsedSecond)
	block, err = core.NewBlock(dpos.chain.ChainID(), coinbase, tail)
	assert.Nil(t, err)
	block.LoadDynastyContext(context)
	block.SetMiner(coinbase)
	block.Seal()
	assert.Nil(t, manager.SignBlock(coinbase, block))
	assert.Nil(t, dpos.FastVerifyBlock(block))

	elapsedSecond = int64(core.DynastySize*core.BlockInterval + core.DynastyInterval)
	context, err = tail.NextDynastyContext(dpos.chain, elapsedSecond)
	block, err = core.NewBlock(dpos.chain.ChainID(), coinbase, tail)
	assert.Nil(t, err)
	block.LoadDynastyContext(context)
	block.SetMiner(coinbase)
	block.Seal()
	assert.Nil(t, manager.SignBlock(coinbase, block))
	assert.Nil(t, dpos.FastVerifyBlock(block))
}

func TestDpos_MintBlock(t *testing.T) {
	dpos, err := NewDpos(mockNeb(t))
	assert.Nil(t, err)
	dpos.chain.SetConsensusHandler(dpos)

	assert.Equal(t, dpos.mintBlock(0), ErrCannotMintWhenDiable)

	assert.Nil(t, dpos.EnableMining("passphrase"))
	dpos.SuspendMining()
	assert.Equal(t, dpos.mintBlock(0), ErrCannotMintWhenPending)

	dpos.ResumeMining()
	assert.Equal(t, dpos.mintBlock(core.BlockInterval), ErrInvalidBlockProposer)

	received = []byte{}
	assert.Equal(t, dpos.mintBlock(core.DynastyInterval), nil)
	assert.NotEqual(t, received, []byte{})
}

func TestContracts(t *testing.T) {
	dpos, err := NewDpos(mockNeb(t))
	assert.Nil(t, err)
	dpos.chain.SetConsensusHandler(dpos)
	tail := dpos.chain.TailBlock()

	coinbase, err := core.AddressParse("1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c")
	assert.Nil(t, err)
	manager := account.NewManager(nil)
	assert.Nil(t, dpos.EnableMining("passphrase"))

	elapsedSecond := int64(core.DynastyInterval)
	context, err := tail.NextDynastyContext(dpos.chain, elapsedSecond)
	assert.Nil(t, err)
	block, err := core.NewBlock(dpos.chain.ChainID(), coinbase, tail)
	block.SetTimestamp(block.Timestamp() + 1)
	assert.Nil(t, err)
	block.LoadDynastyContext(context)

	source := `
	"use strict";var DepositeContent = function (text) {if (text) {
		var o = JSON.parse(text);
		this.balance = new BigNumber(o.balance);
		this.expiryHeight = new BigNumber(o.expiryHeight);
	} else {
		this.balance = new BigNumber(0);
		this.expiryHeight = new BigNumber(0);
	}
};

DepositeContent.prototype = {
	toString: function () {
		return JSON.stringify(this);
	}
};

var BankVaultContract = function () {
	LocalContractStorage.defineMapProperty(this, "bankVault", {
		parse: function (text) {
			return new DepositeContent(text);
		},
		stringify: function (o) {
			return o.toString();
		}
	});
};

// save value to contract, only after height of block, users can takeout
BankVaultContract.prototype = {
	init: function () {
		//TODO:
	},

	save: function (height) {
		var from = Blockchain.transaction.from;
		var value = Blockchain.transaction.value;
		var bk_height = new BigNumber(Blockchain.block.height);

		var orig_deposit = this.bankVault.get(from);
		if (orig_deposit) {
			value = value.plus(orig_deposit.balance);
		}

		var deposit = new DepositeContent();
		deposit.balance = value;
		deposit.expiryHeight = bk_height.plus(height);

		this.bankVault.put(from, deposit);
	},takeout: function (value) {var from = Blockchain.transaction.from;var bk_height = new BigNumber(Blockchain.block.height);var amount = new BigNumber(value);var deposit = this.bankVault.get(from);if (!deposit) {throw new Error("No deposit before.");}if (bk_height.lt(deposit.expiryHeight)) {throw new Error("Can not takeout before expiryHeight.");}if (amount.gt(deposit.balance)) {throw new Error("Insufficient balance.");}var result = Blockchain.transfer(from, amount);if (result != 0) {throw new Error("transfer failed.");}Event.Trigger("BankVault", {Transfer: {from: Blockchain.transaction.to,to: from,value: amount.toString()}});deposit.balance = deposit.balance.sub(amount);this.bankVault.put(from, deposit);},balanceOf: function () {var from = Blockchain.transaction.from;return this.bankVault.get(from);}}; module.exports = BankVaultContract;
`
	sourceType := "js"
	argsDeploy := ""
	payloadDeploy, _ := core.NewDeployPayload(source, sourceType, argsDeploy).ToBytes()
	txDeploy := core.NewTransaction(dpos.chain.ChainID(), coinbase, coinbase, util.NewUint128FromInt(1), 1, core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, util.NewUint128FromInt(200000))
	manager.SignTransaction(coinbase, txDeploy)
	dpos.chain.TransactionPool().Push(txDeploy)

	function := "save"
	argsCall := "[1]"
	payloadCall, _ := core.NewCallPayload(function, argsCall).ToBytes()
	txCall := core.NewTransaction(dpos.chain.ChainID(), coinbase, coinbase, util.NewUint128FromInt(1), 1, core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, util.NewUint128FromInt(200000))
	manager.SignTransaction(coinbase, txCall)
	dpos.chain.TransactionPool().Push(txCall)

	block.CollectTransactions(time.Now().Unix() + 1)

	block.SetMiner(coinbase)
	block.Seal()
	assert.Nil(t, manager.SignBlock(coinbase, block))
	assert.Nil(t, dpos.chain.BlockPool().Push(block))
	assert.Equal(t, block.Hash(), dpos.chain.TailBlock().Hash())
}
