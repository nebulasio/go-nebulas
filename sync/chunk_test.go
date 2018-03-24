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

	"github.com/nebulasio/go-nebulas/nf/nvm"

	"github.com/nebulasio/go-nebulas/account"
	"github.com/nebulasio/go-nebulas/consensus/dpos"
	"github.com/nebulasio/go-nebulas/util"

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
	nvm       core.NVM
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
		nvm:       nvm.NewNebulasVM(),
		ns:        &mockNetService{},
		config: &nebletpb.Config{
			Chain: &nebletpb.ChainConfig{
				ChainId:    genesisConf.Meta.ChainId,
				Keydir:     "keydir",
				Coinbase:   "n1QZMXSZtW7BUerroSms4axNfyBGyFGkrh5",
				Miner:      "n1Kjom3J4KPsHKKzZ2xtt8Lc9W5pRDjeLcW",
				Passphrase: "passphrase",
			},
		},
	}

	am := account.NewManager(neb)
	neb.am = am

	chain, err := core.NewBlockChain(neb)
	assert.Nil(t, err)
	chain.BlockPool().RegisterInNetwork(neb.ns)
	neb.chain = chain
	dpos.Setup(neb)
	chain.Setup(neb)
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
func (n *Neb) IsActiveSyncing() bool {
	return true
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

func (n *Neb) StartActiveSync() {}

func (n *Neb) Nvm() core.NVM {
	return n.nvm
}

func (n *Neb) StartPprof(string) error {
	return nil
}

func (n *Neb) SetGenesis(genesis *corepb.Genesis) {
	n.genesis = genesis
}

var (
	// must be order by address.hash
	MockDynasty = []string{
		"n1FkntVUMPAsESuCAAPK711omQk19JotBjM",
		"n1JNHZJEUvfBYfjDRD14Q73FX62nJAzXkMR",
		"n1Kjom3J4KPsHKKzZ2xtt8Lc9W5pRDjeLcW",
		"n1TV3sU6jyzR4rJ1D7jCAmtVGSntJagXZHC",
		"n1WwqBXVMuYC3mFCEEuFFtAXad6yxqj4as4",
		"n1Zn6iyyQRhqthmCfqGBzWfip1Wx8wEvtrJ",
	}
)

// MockGenesisConf return mock genesis conf
func MockGenesisConf() *corepb.Genesis {
	return &corepb.Genesis{
		Meta: &corepb.GenesisMeta{ChainId: 0},
		Consensus: &corepb.GenesisConsensus{
			Dpos: &corepb.GenesisConsensusDpos{
				Dynasty: MockDynasty,
			},
		},
		TokenDistribution: []*corepb.GenesisTokenDistribution{
			&corepb.GenesisTokenDistribution{
				Address: "n1Z6SbjLuAEXfhX1UJvXT6BB5osWYxVg3F3",
				Value:   "10000000000000000000000",
			},
			&corepb.GenesisTokenDistribution{
				Address: "n1NHcbEus81PJxybnyg4aJgHAaSLDx9Vtf8",
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

	source := `"use strict";var DepositeContent=function(text){if(text){var o=JSON.parse(text);this.balance=new BigNumber(o.balance);this.expiryHeight=new BigNumber(o.expiryHeight)}else{this.balance=new BigNumber(0);this.expiryHeight=new BigNumber(0)}};DepositeContent.prototype={toString:function(){return JSON.stringify(this)}};var BankVaultContract=function(){LocalContractStorage.defineMapProperty(this,"bankVault",{parse:function(text){return new DepositeContent(text)},stringify:function(o){return o.toString()}})};BankVaultContract.prototype={init:function(){},save:function(height){var from=Blockchain.transaction.from;var value=Blockchain.transaction.value;var bk_height=new BigNumber(Blockchain.block.height);var orig_deposit=this.bankVault.get(from);if(orig_deposit){value=value.plus(orig_deposit.balance)}var deposit=new DepositeContent();deposit.balance=value;deposit.expiryHeight=bk_height.plus(height);this.bankVault.put(from,deposit)},takeout:function(value){var from=Blockchain.transaction.from;var bk_height=new BigNumber(Blockchain.block.height);var amount=new BigNumber(value);var deposit=this.bankVault.get(from);if(!deposit){throw new Error("No deposit before.")}if(bk_height.lt(deposit.expiryHeight)){throw new Error("Can not takeout before expiryHeight.")}if(amount.gt(deposit.balance)){throw new Error("Insufficient balance.")}var result=Blockchain.transfer(from,amount);if(result!=0){throw new Error("transfer failed.")}Event.Trigger("BankVault",{Transfer:{from:Blockchain.transaction.to,to:from,value:amount.toString()}});deposit.balance=deposit.balance.sub(amount);this.bankVault.put(from,deposit)},balanceOf:function(){var from=Blockchain.transaction.from;return this.bankVault.get(from)}};module.exports=BankVaultContract;`
	sourceType := "js"
	argsDeploy := ""
	payload, _ := core.NewDeployPayload(source, sourceType, argsDeploy)
	payloadDeploy, _ := payload.ToBytes()

	from, _ := core.AddressParse("n1FkntVUMPAsESuCAAPK711omQk19JotBjM")
	assert.Nil(t, neb.am.Unlock(from, []byte("passphrase"), time.Second*60*60*24*365))

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
		block.SetTimestamp(chain.TailBlock().Timestamp() + dpos.BlockIntervalInMs/dpos.SecondInMs)
		value, _ := util.NewUint128FromInt(1)
		gasLimit, _ := util.NewUint128FromInt(200000)
		txDeploy, _ := core.NewTransaction(neb.chain.ChainID(), from, from, value, uint64(i+1), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
		assert.Nil(t, neb.am.SignTransaction(from, txDeploy))
		assert.Nil(t, neb.chain.TransactionPool().Push(txDeploy))
		if i == 95 {
			block.CollectTransactions(time.Now().Unix()*1000 + 4000)
			assert.Equal(t, len(block.Transactions()), 96)
		}
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
			pbBlockHash, err := core.HashPbBlock(v)
			assert.Nil(t, err)
			assert.Equal(t, block.Hash(), pbBlockHash)
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
