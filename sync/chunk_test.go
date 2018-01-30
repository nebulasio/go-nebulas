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

package sync

import (
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/crypto"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/keystore/secp256k1"
	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/stretchr/testify/assert"

	"testing"
)

type MockConsensus struct {
	storage storage.Storage
}

func (c MockConsensus) Start() {}
func (c MockConsensus) Stop()  {}

func (c MockConsensus) EnableMining(string) error { return nil }
func (c MockConsensus) DisableMining() error      { return nil }
func (c MockConsensus) Enable() bool              { return true }

func (c MockConsensus) ResumeMining()  {}
func (c MockConsensus) SuspendMining() {}
func (c MockConsensus) Pending() bool  { return false }

func (c MockConsensus) FastVerifyBlock(block *core.Block) error {
	block.SetMiner(block.Coinbase())
	return nil
}

func (c MockConsensus) VerifyBlock(block *core.Block, parent *core.Block) error {
	block.SetMiner(block.Coinbase())
	return nil
}

func (c MockConsensus) ForkChoice() error {
	return nil
}

type mockNeb struct {
	genesis *corepb.Genesis
	config  *nebletpb.Config
	storage storage.Storage
	emitter *core.EventEmitter
}

func (n *mockNeb) Genesis() *corepb.Genesis {
	return n.genesis
}

func (n *mockNeb) Config() *nebletpb.Config {
	return n.config
}

func (n *mockNeb) Storage() storage.Storage {
	return n.storage
}

func (n *mockNeb) EventEmitter() *core.EventEmitter {
	return n.emitter
}

func (n *mockNeb) StartActiveSync() {}

func testNeb() *mockNeb {
	storage, _ := storage.NewMemoryStorage()
	eventEmitter := core.NewEventEmitter(1024)
	neb := &mockNeb{
		genesis: MockGenesisConf(),
		config:  &nebletpb.Config{Chain: &nebletpb.ChainConfig{ChainId: MockGenesisConf().Meta.ChainId}},
		storage: storage,
		emitter: eventEmitter,
	}
	return neb
}

var (
	MockDynasty = []string{
		"1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c",
		"2fe3f9f51f9a05dd5f7c5329127f7c917917149b4e16b0b8",
		"333cb3ed8c417971845382ede3cf67a0a96270c05fe2f700",
		"48f981ed38910f1232c1bab124f650c482a57271632db9e3",
		"59fc526072b09af8a8ca9732dae17132c4e9127e43cf2232",
		"75e4e5a71d647298b88928d8cb5da43d90ab1a6c52d0905f",
		"7da9dabedb4c6e121146fb4250a9883d6180570e63d6b080",
		"98a3eed687640b75ec55bf5c9e284371bdcaeab943524d51",
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
		Meta: &corepb.GenesisMeta{ChainId: 100},
		Consensus: &corepb.GenesisConsensus{
			Dpos: &corepb.GenesisConsensusDpos{
				Dynasty: MockDynasty,
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

type MockNetService struct {
}

func (n MockNetService) Start() error { return nil }
func (n MockNetService) Stop()        {}

func (n MockNetService) Node() *net.Node { return nil }

func (n MockNetService) Register(...*net.Subscriber)   {}
func (n MockNetService) Deregister(...*net.Subscriber) {}

func (n MockNetService) Broadcast(string, net.Serializable, int)   {}
func (n MockNetService) Relay(string, net.Serializable, int)       {}
func (n MockNetService) SendMsg(string, []byte, string, int) error { return nil }

func (n MockNetService) SendMessageToPeers(messageName string, data []byte, priority int, filter net.PeerFilterAlgorithm) []string {
	return make([]string, 0)
}
func (n MockNetService) SendMessageToPeer(messageName string, data []byte, priority int, peerID string) error {
	return nil
}

func (n MockNetService) ClosePeer(peerID string, reason error) {}

func (n MockNetService) BroadcastNetworkID([]byte) {}

func (n MockNetService) BuildRawMessageData([]byte, string) []byte { return nil }

func BlockFromNetwork(block *core.Block) *core.Block {
	pb, _ := block.ToProto()
	ir, _ := proto.Marshal(pb)
	proto.Unmarshal(ir, pb)
	b := new(core.Block)
	b.FromProto(pb)
	return b
}

func TestChunk_generateChunkMeta(t *testing.T) {
	var cons MockConsensus
	var n MockNetService
	chain, err := core.NewBlockChain(testNeb())
	chain.SetConsensusHandler(cons)
	chain.BlockPool().RegisterInNetwork(n)
	assert.Nil(t, err)
	ck := NewChunk(chain)

	ks := keystore.DefaultKS
	priv := secp256k1.GeneratePrivateKey()
	pubdata, _ := priv.PublicKey().Encoded()
	coinbase, _ := core.NewAddressFromPublicKey(pubdata)
	ks.SetKey(coinbase.String(), priv, []byte("passphrase"))
	ks.Unlock(coinbase.String(), []byte("passphrase"), time.Second*60*60*24*365)
	key, _ := ks.GetUnlocked(coinbase.String())
	signature, _ := crypto.NewSignature(keystore.SECP256K1)
	signature.InitSign(key.(keystore.PrivateKey))

	blocks := []*core.Block{}
	for i := 0; i < 96; i++ {
		context, err := chain.TailBlock().NextDynastyContext(chain, core.BlockInterval)
		assert.Nil(t, err)
		block, err := chain.NewBlock(coinbase)
		assert.Nil(t, err)
		block.LoadDynastyContext(context)
		block.SetTimestamp(core.BlockInterval * int64(i+1))
		block.SetMiner(coinbase)
		block.Sign(signature)
		assert.Nil(t, block.Seal())
		assert.Nil(t, chain.BlockPool().Push(BlockFromNetwork(block)))
		assert.Nil(t, chain.SetTailBlock(block))
		blocks = append(blocks, block)
	}

	meta, err := ck.generateChunkHeaders(blocks[1].Hash())
	assert.Nil(t, err)
	assert.Equal(t, len(meta.ChunkHeaders), 3)
	chunks, err := ck.generateChunkData(meta.ChunkHeaders[1])
	assert.Nil(t, err)
	assert.Equal(t, len(chunks.Blocks), core.ChunkSize)
	for i := 0; i < core.ChunkSize; i++ {
		index := core.ChunkSize + 2 + i
		assert.Equal(t, int(chunks.Blocks[i].Height), index)
	}

	meta, err = ck.generateChunkHeaders(chain.GenesisBlock().Hash())
	assert.Nil(t, err)
	assert.Equal(t, len(meta.ChunkHeaders), 3)

	meta, err = ck.generateChunkHeaders(blocks[0].Hash())
	assert.Nil(t, err)
	assert.Equal(t, len(meta.ChunkHeaders), 3)

	meta, err = ck.generateChunkHeaders(blocks[31].Hash())
	assert.Nil(t, err)
	assert.Equal(t, int(blocks[31].Height()), 33)
	assert.Equal(t, len(meta.ChunkHeaders), 2)

	meta, err = ck.generateChunkHeaders(blocks[62].Hash())
	assert.Nil(t, err)
	assert.Equal(t, int(blocks[62].Height()), 64)
	assert.Equal(t, len(meta.ChunkHeaders), 2)

	chain2, err := core.NewBlockChain(testNeb())
	chain2.SetConsensusHandler(cons)
	chain2.BlockPool().RegisterInNetwork(n)
	assert.Nil(t, err)
	meta, err = ck.generateChunkHeaders(blocks[0].Hash())
	assert.Nil(t, err)
	for _, header := range meta.ChunkHeaders {
		chunk, err := ck.generateChunkData(header)
		assert.Nil(t, err)
		for _, v := range chunk.Blocks {
			block := new(core.Block)
			assert.Nil(t, block.FromProto(v))
			assert.Nil(t, chain2.BlockPool().Push(block))
		}
	}
	tail := blocks[len(blocks)-1]
	bytes, err := chain2.Storage().Get(tail.Hash())
	assert.Nil(t, err)
	pbBlock := new(corepb.Block)
	checkBlock := new(core.Block)
	assert.Nil(t, proto.Unmarshal(bytes, pbBlock))
	assert.Nil(t, checkBlock.FromProto(pbBlock))
	assert.Equal(t, checkBlock.Hash(), tail.Hash())
}
