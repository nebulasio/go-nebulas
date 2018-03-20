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
	"github.com/nebulasio/go-nebulas/core/state"
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
	nvm       core.Engine
}

func mockNeb(t *testing.T) *Neb {
	storage, _ := storage.NewMemoryStorage()
	eventEmitter := core.NewEventEmitter(1024)
	genesisConf := MockGenesisConf()
	dpos := NewDpos()
	nvm := &mockNvm{}
	neb := &Neb{
		genesis:   genesisConf,
		storage:   storage,
		emitter:   eventEmitter,
		consensus: dpos,
		nvm:       nvm,
		config: &nebletpb.Config{
			Chain: &nebletpb.ChainConfig{
				ChainId:    genesisConf.Meta.ChainId,
				Keydir:     "keydir",
				Coinbase:   "n1K4rWU3YrhZmU1GHHYqnES8CcypTYQa9oJ",
				Miner:      "n1PtnbfQcC9EZpr2LS2vLUCKf2UtkyArzVr",
				Passphrase: "passphrase",
			},
		},
		ns: mockNetService{},
	}

	am := account.NewManager(neb)
	neb.am = am

	chain, err := core.NewBlockChain(neb)
	assert.Nil(t, err)
	neb.chain = chain
	assert.Nil(t, dpos.Setup(neb))
	assert.Nil(t, chain.Setup(neb))

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

func (n *Neb) AccountManager() core.AccountManager {
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

func (n *Neb) Nvm() core.Engine {
	return n.nvm
}

func (n *Neb) StartActiveSync() {}

func (n *Neb) StartPprof(string) error { return nil }

func (n *Neb) SetGenesis(genesis *corepb.Genesis) {
	n.genesis = genesis
}

type mockNvm struct {
}

func (nvm *mockNvm) CreateEngine(block *core.Block, tx *core.Transaction, owner, contract state.Account, state state.AccountState) error {
	return nil
}
func (nvm *mockNvm) SetEngineExecutionLimits(limitsOfExecutionInstructions uint64) error {
	return nil
}
func (nvm *mockNvm) DeployAndInitEngine(source, sourceType, args string) (string, error) {
	return "", nil
}
func (nvm *mockNvm) CallEngine(source, sourceType, function, args string) (string, error) {
	return "", nil
}
func (nvm *mockNvm) ExecutionInstructions() (uint64, error) {
	return uint64(100), nil
}
func (nvm *mockNvm) DisposeEngine() {

}

func (nvm *mockNvm) Clone() core.Engine {
	return &mockNvm{}
}

var (
	DefaultOpenDynasty = []string{
		"n1LQxBdAtxcfjUazHeK94raKdxRsNpujUyU",
		"n1PtnbfQcC9EZpr2LS2vLUCKf2UtkyArzVr",
		"n1SRGKRFrF6DHK4Ym4MoXbbUHYkV5W2MZPw",
		"n1TRySsvYmAU8ChPZyYyvrPpDYJ1Z5DFoxo",
		"n1aoyV8M2g79pFXxdZEK9GfU7fzuJcCN75X",
		"n1beo9QAjhhJX6tjpjHyinoorbqdi6UKAEb",
	}
)

// MockGenesisConf return mock genesis conf
func MockGenesisConf() *corepb.Genesis {
	dynasty := []string{}
	for _, v := range DefaultOpenDynasty {
		dynasty = append(dynasty, v)
	}
	return &corepb.Genesis{
		Meta: &corepb.GenesisMeta{ChainId: 0},
		Consensus: &corepb.GenesisConsensus{
			Dpos: &corepb.GenesisConsensusDpos{
				Dynasty: dynasty,
			},
		},
		TokenDistribution: []*corepb.GenesisTokenDistribution{
			&corepb.GenesisTokenDistribution{
				Address: "n1LQxBdAtxcfjUazHeK94raKdxRsNpujUyU",
				Value:   "10000000000000000000000",
			},
			&corepb.GenesisTokenDistribution{
				Address: "n1SRGKRFrF6DHK4Ym4MoXbbUHYkV5W2MZPw",
				Value:   "10000000000000000000000",
			},
			&corepb.GenesisTokenDistribution{
				Address: "n1UZtMgi94oE913L2Sa2C9XwvAzNTQ82v64",
				Value:   "10000000000000000000000",
			},
			&corepb.GenesisTokenDistribution{
				Address: "n1aoyV8M2g79pFXxdZEK9GfU7fzuJcCN75X",
				Value:   "10000000000000000000000",
			},
			&corepb.GenesisTokenDistribution{
				Address: "n1coJhpn8QXvKFogVG93wx49eCQ6aPQHSAN",
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

/* func TestDpos_New(t *testing.T) {
	neb := mockNeb(t)
	coinbase := neb.config.Chain.Coinbase
	neb.config.Chain.Coinbase += "0"
	assert.NotNil(t, neb.Consensus().Setup(neb))
	neb.config.Chain.Coinbase = coinbase
	neb.config.Chain.Miner += "0"
	assert.NotNil(t, neb.Consensus().Setup(neb))
} */

func TestDpos_VerifySign(t *testing.T) {
	neb := mockNeb(t)
	dpos := neb.consensus
	chain := neb.chain
	tail := chain.TailBlock()

	elapsedSecond := int64(DynastySize*BlockInterval + DynastyInterval)
	consensusState, err := tail.NextConsensusState(elapsedSecond)
	assert.Nil(t, err)
	coinbase, err := core.AddressParse("n1K4rWU3YrhZmU1GHHYqnES8CcypTYQa9oJ")
	assert.Nil(t, err)
	block, err := core.NewBlock(chain.ChainID(), coinbase, tail)
	assert.Nil(t, err)
	block.LoadConsensusState(consensusState)
	block.Seal()
	manager := account.NewManager(nil)
	miner, err := core.AddressParseFromBytes(consensusState.Proposer())
	assert.Nil(t, err)
	assert.Nil(t, manager.Unlock(miner, []byte("passphrase"), keystore.DefaultUnlockDuration))
	assert.Nil(t, manager.SignBlock(miner, block))
	assert.Nil(t, dpos.VerifyBlock(block))

	miner, err = core.AddressParse("n1PtnbfQcC9EZpr2LS2vLUCKf2UtkyArzVr")
	assert.Nil(t, err)
	assert.Nil(t, manager.Unlock(miner, []byte("passphrase"), keystore.DefaultUnlockDuration))
	assert.Nil(t, manager.SignBlock(miner, block))
	assert.Equal(t, dpos.VerifyBlock(block), ErrDoubleBlockMinted)
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
	chain := neb.chain

	/*
		genesis -- 0 -- 11 -- 111 -- 1111
					 \_ 12 -- 221
	*/

	addr0 := GetUnlockAddress(t, am, "n1PtnbfQcC9EZpr2LS2vLUCKf2UtkyArzVr")
	block0, _ := chain.NewBlock(addr0)
	consensusState, err := chain.TailBlock().NextConsensusState(BlockInterval)
	assert.Nil(t, err)
	block0.LoadConsensusState(consensusState)
	block0.Seal()
	am.SignBlock(addr0, block0)
	assert.Nil(t, chain.BlockPool().Push(block0))
	assert.Equal(t, block0.Hash(), chain.TailBlock().Hash())

	addr1 := GetUnlockAddress(t, am, "n1SRGKRFrF6DHK4Ym4MoXbbUHYkV5W2MZPw")
	block11, err := chain.NewBlock(addr1)
	assert.Nil(t, err)
	consensusState, err = block0.NextConsensusState(BlockInterval)
	assert.Nil(t, err)
	block11.LoadConsensusState(consensusState)
	block11.Seal()
	am.SignBlock(addr1, block11)
	assert.Nil(t, chain.BlockPool().Push(block11))

	addr2 := GetUnlockAddress(t, am, "n1TRySsvYmAU8ChPZyYyvrPpDYJ1Z5DFoxo")
	block12, _ := chain.NewBlockFromParent(addr2, block0)
	consensusState, err = block0.NextConsensusState(BlockInterval * 2)
	assert.Nil(t, err)
	block12.LoadConsensusState(consensusState)
	block12.Seal()
	am.SignBlock(addr2, block12)
	assert.Nil(t, chain.BlockPool().Push(block12))
	assert.Equal(t, len(neb.chain.DetachedTailBlocks()), 2)
	tail := block11.Hash()
	if less(block11, block12) {
		tail = block12.Hash()
	}
	assert.Equal(t, neb.chain.TailBlock().Hash(), tail)

	addr3 := GetUnlockAddress(t, am, "n1aoyV8M2g79pFXxdZEK9GfU7fzuJcCN75X")
	block111, _ := chain.NewBlockFromParent(addr3, block11)
	consensusState, err = block11.NextConsensusState(BlockInterval * 2)
	assert.Nil(t, err)
	block111.LoadConsensusState(consensusState)
	block111.Seal()
	am.SignBlock(addr3, block111)
	assert.Nil(t, chain.BlockPool().Push(block111))
	assert.Equal(t, len(chain.DetachedTailBlocks()), 2)
}

func TestCanMining(t *testing.T) {
	dpos := mockNeb(t).consensus
	assert.Equal(t, dpos.Pending(), true)
	dpos.SuspendMining()
	assert.Equal(t, dpos.Pending(), true)
	dpos.ResumeMining()
	assert.Equal(t, dpos.Pending(), false)
}

func TestDpos_MintBlock(t *testing.T) {
	neb := mockNeb(t)
	dpos := neb.consensus.(*Dpos)

	assert.Equal(t, dpos.mintBlock(0), ErrCannotMintWhenDisable)

	assert.Nil(t, dpos.EnableMining("passphrase"))
	dpos.SuspendMining()
	assert.Equal(t, dpos.mintBlock(0), ErrCannotMintWhenPending)

	dpos.ResumeMining()
	assert.Equal(t, dpos.mintBlock(DynastyInterval), ErrInvalidBlockProposer)

	received = []byte{}
	assert.Equal(t, dpos.mintBlock(BlockInterval), nil)
	assert.NotEqual(t, received, []byte{})
}

func TestContracts(t *testing.T) {
	neb := mockNeb(t)
	dpos := neb.consensus.(*Dpos)
	chain := neb.chain
	tail := chain.TailBlock()

	coinbase, err := core.AddressParse("n1PtnbfQcC9EZpr2LS2vLUCKf2UtkyArzVr")
	assert.Nil(t, err)
	manager := account.NewManager(nil)
	assert.Nil(t, dpos.EnableMining("passphrase"))

	elapsedSecond := int64(DynastyInterval / 12)
	context, err := tail.NextConsensusState(elapsedSecond)
	assert.Nil(t, err)
	block, err := core.NewBlock(dpos.chain.ChainID(), coinbase, tail)
	assert.Nil(t, err)
	block.LoadConsensusState(context)

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

	value, _ := util.NewUint128FromInt(1)
	gasLimit, _ := util.NewUint128FromInt(200000)
	txDeploy, _ := core.NewTransaction(dpos.chain.ChainID(), coinbase, coinbase, value, 1, core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
	manager.SignTransaction(coinbase, txDeploy)
	dpos.chain.TransactionPool().Push(txDeploy)

	function := "save"
	argsCall := "[1]"
	payloadCall, _ := core.NewCallPayload(function, argsCall).ToBytes()
	txCall, _ := core.NewTransaction(dpos.chain.ChainID(), coinbase, coinbase, value, 1, core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
	manager.SignTransaction(coinbase, txCall)
	dpos.chain.TransactionPool().Push(txCall)

	block.CollectTransactions(time.Now().Unix() + 1)
	block.Seal()
	assert.Nil(t, manager.SignBlock(coinbase, block))
	assert.Nil(t, dpos.chain.BlockPool().Push(block))
	assert.Equal(t, block.Hash(), dpos.chain.TailBlock().Hash())
}

func TestDoubleMint(t *testing.T) {
	neb := mockNeb(t)
	chain := neb.chain
	am := neb.am

	addr0 := GetUnlockAddress(t, am, "n1PtnbfQcC9EZpr2LS2vLUCKf2UtkyArzVr")
	block0, _ := chain.NewBlock(addr0)
	consensusState, err := chain.TailBlock().NextConsensusState(BlockInterval)
	assert.Nil(t, err)
	block0.LoadConsensusState(consensusState)
	block0.Seal()
	am.SignBlock(addr0, block0)
	assert.Nil(t, chain.BlockPool().Push(block0))
	assert.Equal(t, block0.Hash(), chain.TailBlock().Hash())

	block11, err := chain.NewBlock(addr0)
	assert.Nil(t, err)
	consensusState, err = block0.NextConsensusState(0)
	assert.Nil(t, err)
	block11.LoadConsensusState(consensusState)
	block11.Seal()
	am.SignBlock(addr0, block11)
	assert.Equal(t, chain.BlockPool().Push(block11), ErrDoubleBlockMinted)
}
