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

	"github.com/nebulasio/go-nebulas/account"
	"github.com/nebulasio/go-nebulas/consensus/dpos"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/stretchr/testify/assert"

	"testing"
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
	storage, _ := storage.NewMemoryStorage()
	eventEmitter := core.NewEventEmitter(1024)
	genesisConf := MockGenesisConf()
	dpos := dpos.NewDpos()
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

	am := account.NewManager(neb)
	neb.am = am

	chain, err := core.NewBlockChain(neb)
	assert.Nil(t, err)
	neb.chain = chain
	dpos.Setup(neb)
	chain.Setup(neb)

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

func TestChunk_generateChunkMeta(t *testing.T) {
	neb := mockNeb(t)
	chain := neb.chain
	ck := NewChunk(chain)

	blocks := []*core.Block{}
	for i := 0; i < 96; i++ {
		context, err := chain.TailBlock().WorldState().NextConsensusState(dpos.BlockIntervalInMs / dpos.SecondInMs)
		assert.Nil(t, err)
		coinbase, err := core.AddressParseFromBytes(context.Proposer())
		assert.Nil(t, err)
		assert.Nil(t, neb.am.Unlock(coinbase, []byte("passphrase"), time.Second*60*60*24*365))
		block, err := chain.NewBlock(coinbase)
		assert.Nil(t, err)
		block.WorldState().SetConsensusState(context)
		block.SetTimestamp((dpos.BlockIntervalInMs / dpos.SecondInMs) * int64(i+1))
		block.SetMiner(coinbase)
		assert.Nil(t, block.Seal())
		assert.Nil(t, neb.am.SignBlock(coinbase, block))
		assert.Nil(t, chain.BlockPool().Push(block))
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

	neb2 := mockNeb(t)
	chain2 := neb2.chain
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
