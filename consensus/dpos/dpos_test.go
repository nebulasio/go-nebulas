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

	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/account"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/crypto"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/keystore/secp256k1"
	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/net/p2p"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/stretchr/testify/assert"
)

type Neb struct {
	config  nebletpb.Config
	chain   *core.BlockChain
	ns      p2p.Manager
	am      *account.Manager
	genesis *corepb.Genesis
	storage storage.Storage
	emitter *core.EventEmitter
}

func mockNeb() *Neb {
	storage, _ := storage.NewMemoryStorage()
	eventEmitter := core.NewEventEmitter(1024)
	genesisConf := MockGenesisConf()
	neb := &Neb{
		genesis: genesisConf,
		storage: storage,
		emitter: eventEmitter,
		config: nebletpb.Config{
			Chain: &nebletpb.ChainConfig{
				ChainId:    genesisConf.Meta.ChainId,
				Coinbase:   "1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c",
				Miner:      "1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c",
				Passphrase: "passphrase",
			},
		},
	}
	am := account.NewManager(neb)
	var nm MockNetManager
	chain, _ := core.NewBlockChain(neb)
	neb.chain = chain
	neb.am = am
	neb.ns = nm
	neb.chain.BlockPool().RegisterInNetwork(nm)
	return neb
}

func (n *Neb) Config() nebletpb.Config {
	return n.config
}

func (n *Neb) BlockChain() *core.BlockChain {
	return n.chain
}

func (n *Neb) NetManager() p2p.Manager {
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

func (n *Neb) StartSync() {}

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

type MockConsensus struct {
	storage storage.Storage
}

func (c MockConsensus) FastVerifyBlock(block *core.Block) error {
	block.SetMiner(block.Coinbase())
	return nil
}
func (c MockConsensus) VerifyBlock(block *core.Block, parent *core.Block) error {
	block.SetMiner(block.Coinbase())
	return nil
}

var (
	received = []byte{}
)

type MockNetManager struct{}

func (n MockNetManager) Start() error { return nil }
func (n MockNetManager) Stop()        {}

func (n MockNetManager) Node() *p2p.Node { return nil }

func (n MockNetManager) Sync(net.Serializable) error            { return nil }
func (n MockNetManager) SendSyncReply(string, net.Serializable) {}

func (n MockNetManager) Register(...*net.Subscriber)   {}
func (n MockNetManager) Deregister(...*net.Subscriber) {}

func (n MockNetManager) Broadcast(name string, msg net.Serializable) {
	pb, _ := msg.ToProto()
	bytes, _ := proto.Marshal(pb)
	received = bytes
}
func (n MockNetManager) Relay(name string, msg net.Serializable) {
	pb, _ := msg.ToProto()
	bytes, _ := proto.Marshal(pb)
	received = bytes
}
func (n MockNetManager) SendMsg(name string, msg []byte, target string) error {
	received = msg
	return nil
}

func (n MockNetManager) BroadcastNetworkID([]byte) {}

func (n MockNetManager) BuildData([]byte, string) []byte { return nil }

func TestDpos_New(t *testing.T) {
	neb := mockNeb()
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
	dpos, err := NewDpos(mockNeb())
	assert.Nil(t, err)
	var c MockConsensus
	dpos.chain.SetConsensusHandler(c)
	tail := dpos.chain.TailBlock()

	elapsedSecond := int64(core.DynastySize*core.BlockInterval + core.DynastyInterval)
	context, err := tail.NextDynastyContext(elapsedSecond)
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
	assert.Nil(t, manager.Unlock(miner, []byte("passphrase")))
	assert.Nil(t, manager.SignBlock(miner, block))
	assert.Nil(t, dpos.VerifyBlock(block, tail))

	miner, err = core.AddressParse("fc751b484bd5296f8d267a8537d33f25a848f7f7af8cfcf6")
	assert.Nil(t, err)
	assert.Nil(t, manager.Unlock(miner, []byte("passphrase")))
	assert.Nil(t, manager.SignBlock(miner, block))
	assert.Equal(t, dpos.VerifyBlock(block, tail), ErrInvalidBlockProposer)
}

func TestForkChoice(t *testing.T) {
	dpos, err := NewDpos(mockNeb())
	assert.Nil(t, err)
	var c MockConsensus
	dpos.chain.SetConsensusHandler(c)

	ks := keystore.DefaultKS
	priv := secp256k1.GeneratePrivateKey()
	pubdata, _ := priv.PublicKey().Encoded()
	from, _ := core.NewAddressFromPublicKey(pubdata)
	ks.SetKey(from.String(), priv, []byte("passphrase"))
	ks.Unlock(from.String(), []byte("passphrase"), time.Second*60*60*24*365)
	key, _ := ks.GetUnlocked(from.String())
	signature, _ := crypto.NewSignature(keystore.SECP256K1)
	signature.InitSign(key.(keystore.PrivateKey))

	/*
		genesis -- 0 -- 11 -- 111 -- 1111
					 \_ 12 -- 221
	*/

	block0, _ := dpos.chain.NewBlock(from)
	block0.SetTimestamp(core.BlockInterval)
	block0.SetMiner(from)
	block0.Seal()
	assert.Nil(t, dpos.chain.BlockPool().Push(block0))
	dpos.forkChoice()
	assert.Equal(t, block0.Hash(), dpos.chain.TailBlock().Hash())

	block11, _ := dpos.chain.NewBlock(from)
	block11.SetTimestamp(core.BlockInterval * 2)
	block11.SetMiner(from)
	block11.Seal()
	assert.Nil(t, dpos.chain.BlockPool().Push(block11))

	block12, _ := dpos.chain.NewBlock(from)
	block12.SetTimestamp(core.BlockInterval * 3)
	block12.SetMiner(from)
	block12.Seal()
	assert.Nil(t, dpos.chain.BlockPool().Push(block12))

	assert.Equal(t, len(dpos.chain.DetachedTailBlocks()), 2)
	dpos.forkChoice()
	tail := block11
	if core.Less(block11, block12) {
		tail = block12
	}
	assert.Equal(t, dpos.chain.TailBlock().Hash(), tail.Hash())
	assert.Equal(t, len(dpos.chain.DetachedTailBlocks()), 2)

	block111, _ := dpos.chain.NewBlockFromParent(from, block11)
	block111.SetTimestamp(core.BlockInterval * 4)
	block111.SetMiner(from)
	block111.Seal()

	block1111, _ := dpos.chain.NewBlockFromParent(from, block111)
	block1111.SetTimestamp(core.BlockInterval * 5)
	block1111.SetMiner(from)
	block1111.Seal()
	assert.Error(t, dpos.chain.BlockPool().Push(block1111), core.ErrMissingParentBlock)
	assert.Equal(t, len(dpos.chain.DetachedTailBlocks()), 2)
	assert.Nil(t, dpos.chain.BlockPool().Push(block111))
	assert.Equal(t, len(dpos.chain.DetachedTailBlocks()), 2)
	dpos.forkChoice()
	assert.Equal(t, len(dpos.chain.DetachedTailBlocks()), 2)
	assert.Equal(t, dpos.chain.TailBlock().Hash(), block1111.Hash())

	block221, _ := dpos.chain.NewBlockFromParent(from, block12)
	block221.SetTimestamp(core.BlockInterval * 6)
	block221.SetMiner(from)
	block221.Seal()
	assert.Nil(t, dpos.chain.BlockPool().Push(block221))
	assert.Equal(t, len(dpos.chain.DetachedTailBlocks()), 2)
	dpos.forkChoice()
	assert.Equal(t, len(dpos.chain.DetachedTailBlocks()), 2)
	assert.Equal(t, dpos.chain.TailBlock().Hash(), block1111.Hash())
}

func TestCanMining(t *testing.T) {
	dpos, err := NewDpos(mockNeb())
	assert.Nil(t, err)
	assert.Equal(t, dpos.CanMining(), false)
	dpos.SetCanMining(true)
	assert.Equal(t, dpos.CanMining(), true)
}

func TestFastVerifyBlock(t *testing.T) {
	dpos, err := NewDpos(mockNeb())
	assert.Nil(t, err)
	var c MockConsensus
	dpos.chain.SetConsensusHandler(c)
	tail := dpos.chain.TailBlock()

	coinbase, err := core.AddressParse("1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c")
	assert.Nil(t, err)
	manager := account.NewManager(nil)
	assert.Nil(t, manager.Unlock(coinbase, []byte("passphrase")))

	elapsedSecond := int64(core.DynastyInterval)
	context, err := tail.NextDynastyContext(elapsedSecond)
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
	context, err = tail.NextDynastyContext(elapsedSecond)
	block, err = core.NewBlock(dpos.chain.ChainID(), coinbase, tail)
	assert.Nil(t, err)
	block.LoadDynastyContext(context)
	block.SetMiner(coinbase)
	block.Seal()
	assert.Nil(t, manager.SignBlock(coinbase, block))
	assert.Nil(t, dpos.FastVerifyBlock(block))

	elapsedSecond = int64(core.DynastySize*core.BlockInterval + core.DynastyInterval)
	context, err = tail.NextDynastyContext(elapsedSecond)
	block, err = core.NewBlock(dpos.chain.ChainID(), coinbase, tail)
	assert.Nil(t, err)
	block.LoadDynastyContext(context)
	block.SetMiner(coinbase)
	block.Seal()
	assert.Nil(t, manager.SignBlock(coinbase, block))
	assert.Nil(t, dpos.FastVerifyBlock(block))
}

func TestDpos_MintBlock(t *testing.T) {
	dpos, err := NewDpos(mockNeb())
	assert.Nil(t, err)
	var c MockConsensus
	dpos.chain.SetConsensusHandler(c)

	coinbase, err := core.AddressParse("1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c")
	assert.Nil(t, err)
	manager := account.NewManager(nil)
	assert.Nil(t, manager.Unlock(coinbase, []byte("passphrase")))

	assert.Equal(t, dpos.mintBlock(0), ErrCannotMintBlockNow)

	dpos.SetCanMining(true)
	assert.Equal(t, dpos.mintBlock(core.BlockInterval), ErrInvalidBlockProposer)

	received = []byte{}
	assert.Equal(t, dpos.mintBlock(core.DynastyInterval), nil)
	assert.NotEqual(t, received, []byte{})
}
