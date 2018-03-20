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

package core

import (
	"reflect"
	"testing"
	"time"

	"github.com/nebulasio/go-nebulas/consensus/pb"
	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/util/byteutils"

	pb "github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/crypto"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/keystore/secp256k1"
	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/stretchr/testify/assert"
)

const (
	BlockInterval        = 5
	AcceptedNetWorkDelay = 2
)

var (
	stor, _ = storage.NewMemoryStorage()
)

var (
	MockDynasty = []string{
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
	return &corepb.Genesis{
		Meta: &corepb.GenesisMeta{ChainId: 100},
		Consensus: &corepb.GenesisConsensus{
			Dpos: &corepb.GenesisConsensusDpos{
				Dynasty: MockDynasty,
			},
		},
		TokenDistribution: []*corepb.GenesisTokenDistribution{
			&corepb.GenesisTokenDistribution{
				Address: "n1UZtMgi94oE913L2Sa2C9XwvAzNTQ82v64",
				Value:   "10000000000000000000000",
			},
			&corepb.GenesisTokenDistribution{
				Address: "n1coJhpn8QXvKFogVG93wx49eCQ6aPQHSAN",
				Value:   "10000000000000000000000",
			},
		},
	}
}

type mockConsensusState struct {
	timestamp int64
}

func newMockConsensusState(timestamp int64) (*mockConsensusState, error) {
	return &mockConsensusState{
		timestamp: timestamp,
	}, nil
}

func (cs *mockConsensusState) Begin()    {}
func (cs *mockConsensusState) Commit()   {}
func (cs *mockConsensusState) Rollback() {}

func (cs *mockConsensusState) RootHash() (*consensuspb.ConsensusRoot, error) {
	return &consensuspb.ConsensusRoot{}, nil
}
func (cs *mockConsensusState) String() string { return "" }
func (cs *mockConsensusState) Clone() (state.ConsensusState, error) {
	return &mockConsensusState{
		timestamp: cs.timestamp,
	}, nil
}

func (cs *mockConsensusState) Proposer() byteutils.Hash { return nil }
func (cs *mockConsensusState) TimeStamp() int64         { return cs.timestamp }
func (cs *mockConsensusState) NextState(elapsed int64) (state.ConsensusState, error) {
	return &mockConsensusState{
		timestamp: cs.timestamp + elapsed,
	}, nil
}

func (cs *mockConsensusState) Dynasty() ([]byteutils.Hash, error) { return nil, nil }
func (cs *mockConsensusState) DynastyRoot() byteutils.Hash        { return nil }

type mockConsensus struct {
	chain *BlockChain
}

func (c *mockConsensus) Setup(neb Neblet) error {
	c.chain = neb.BlockChain()
	return nil
}

func (c *mockConsensus) Start() {}
func (c *mockConsensus) Stop()  {}

func (c *mockConsensus) VerifyBlock(block *Block) error {
	return nil
}

func mockLess(a *Block, b *Block) bool {
	if a.Height() != b.Height() {
		return a.Height() < b.Height()
	}
	return byteutils.Less(a.Hash(), b.Hash())
}

// ForkChoice select new tail
func (c *mockConsensus) ForkChoice() error {
	bc := c.chain
	tailBlock := bc.TailBlock()
	detachedTailBlocks := bc.DetachedTailBlocks()

	// find the max depth.
	newTailBlock := tailBlock

	for _, v := range detachedTailBlocks {
		if mockLess(newTailBlock, v) {
			newTailBlock = v
		}
	}

	if newTailBlock.Hash().Equals(tailBlock.Hash()) {
		return nil
	}

	err := bc.SetTailBlock(newTailBlock)
	if err != nil {
		return err
	}
	return nil
}

func (c *mockConsensus) UpdateLIB() {}

func (c *mockConsensus) SuspendMining() {}
func (c *mockConsensus) ResumeMining()  {}
func (c *mockConsensus) Pending() bool  { return false }

func (c *mockConsensus) EnableMining(passphrase string) error { return nil }
func (c *mockConsensus) DisableMining() error                 { return nil }
func (c *mockConsensus) Enable() bool                         { return true }

func (c *mockConsensus) CheckTimeout(block *Block) bool {
	return time.Now().Unix()-block.Timestamp() > AcceptedNetWorkDelay
}
func (c *mockConsensus) NewState(root *consensuspb.ConsensusRoot, storage storage.Storage) (state.ConsensusState, error) {
	return newMockConsensusState(root.Timestamp)
}
func (c *mockConsensus) GenesisState(*BlockChain, *corepb.Genesis) (state.ConsensusState, error) {
	return newMockConsensusState(0)
}

type mockManager struct{}

func (m mockManager) NewAccount([]byte) (*Address, error) { return nil, nil }
func (m mockManager) Accounts() []*Address                { return nil }

func (m mockManager) Unlock(addr *Address, passphrase []byte, expire time.Duration) error { return nil }
func (m mockManager) Lock(addr *Address) error                                            { return nil }

func (m mockManager) SignBlock(addr *Address, block *Block) error                        { return nil }
func (m mockManager) SignTransaction(*Address, *Transaction) error                       { return nil }
func (m mockManager) SignTransactionWithPassphrase(*Address, *Transaction, []byte) error { return nil }

func (m mockManager) Update(*Address, []byte, []byte) error   { return nil }
func (m mockManager) Load([]byte, []byte) (*Address, error)   { return nil, nil }
func (m mockManager) Import([]byte, []byte) (*Address, error) { return nil, nil }
func (m mockManager) Delete(*Address, []byte) error           { return nil }

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

func (n mockNetService) Broadcast(name string, msg net.Serializable, priority int) {}
func (n mockNetService) Relay(name string, msg net.Serializable, priority int)     {}
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

type mockNeb struct {
	config    *nebletpb.Config
	chain     *BlockChain
	ns        net.Service
	am        AccountManager
	genesis   *corepb.Genesis
	storage   storage.Storage
	consensus Consensus
	emitter   *EventEmitter
	nvm       Engine
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

func (n *mockNeb) EventEmitter() *EventEmitter {
	return n.emitter
}

func (n *mockNeb) Consensus() Consensus {
	return n.consensus
}

func (n *mockNeb) BlockChain() *BlockChain {
	return n.chain
}

func (n *mockNeb) NetService() net.Service {
	return n.ns
}

func (n *mockNeb) AccountManager() AccountManager {
	return n.am
}

func (n *mockNeb) Nvm() Engine {
	return n.nvm
}

func (n *mockNeb) StartPprof(string) error {
	return nil
}

func (n *mockNeb) SetGenesis(genesis *corepb.Genesis) {
	n.genesis = genesis
}

type mockNvm struct {
}

func (nvm *mockNvm) CreateEngine(block *Block, tx *Transaction, owner, contract state.Account, state state.AccountState) error {
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

func (nvm *mockNvm) Clone() Engine {
	return &mockNvm{}
}

func testNeb(t *testing.T) *mockNeb {
	storage, _ := storage.NewMemoryStorage()
	eventEmitter := NewEventEmitter(1024)
	consensus := new(mockConsensus)
	nvm := &mockNvm{}
	var am mockManager
	var ns mockNetService
	neb := &mockNeb{
		genesis:   MockGenesisConf(),
		config:    &nebletpb.Config{Chain: &nebletpb.ChainConfig{ChainId: MockGenesisConf().Meta.ChainId}},
		storage:   storage,
		emitter:   eventEmitter,
		consensus: consensus,
		am:        am,
		ns:        ns,
		nvm:       nvm,
	}
	chain, err := NewBlockChain(neb)
	assert.Nil(t, err)
	chain.bkPool.RegisterInNetwork(ns)
	neb.chain = chain
	assert.Nil(t, consensus.Setup(neb))
	assert.Nil(t, chain.Setup(neb))
	return neb
}
func TestBlock(t *testing.T) {
	type fields struct {
		header       *BlockHeader
		miner        *Address
		height       uint64
		transactions Transactions
	}
	from1, _ := NewAddress(AccountAddress, []byte("eb693e1438fce79f5cb2"))
	from2, _ := NewAddress(AccountAddress, []byte("eb692e1438fce79f5cb2"))
	to1, _ := NewAddress(AccountAddress, []byte("eb691e1438fce79f5cb2"))
	to2, _ := NewAddress(AccountAddress, []byte("eb690e1438fce79f5cb2"))
	coinbase, _ := NewAddress(AccountAddress, []byte("5425730430bc2d63f257"))

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			"full struct",
			fields{
				&BlockHeader{
					hash:       []byte("a6e5eb190e1438fce79f5cb8774a72621637c2c9654c8b2525ed1d7e4e73653f"),
					parentHash: []byte("a6e5eb240e1438fce79f5cb8774a72621637c2c9654c8b2525ed1d7e4e73653f"),
					stateRoot:  []byte("43656"),
					txsRoot:    []byte("43656"),
					eventsRoot: []byte("43656"),
					consensusRoot: &consensuspb.ConsensusRoot{
						DynastyRoot: []byte("43656"),
					},
					coinbase:  coinbase,
					timestamp: time.Now().Unix(),
					chainID:   100,
				},
				&Address{address: []byte("hello")},
				1,
				Transactions{
					&Transaction{
						[]byte("123452"),
						from1,
						to1,
						util.NewUint128(),
						456,
						1516464510,
						&corepb.Data{Type: TxPayloadBinaryType, Payload: []byte("hello")},
						1,
						util.NewUint128(),
						util.NewUint128(),
						keystore.SECP256K1,
						nil,
					},
					&Transaction{
						[]byte("123455"),
						from2,
						to2,
						util.NewUint128(),
						446,
						1516464511,
						&corepb.Data{Type: TxPayloadBinaryType, Payload: []byte("hllo")},
						2,
						util.NewUint128(),
						util.NewUint128(),
						keystore.SECP256K1,
						nil,
					},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Block{
				header:       tt.fields.header,
				height:       tt.fields.height,
				transactions: tt.fields.transactions,
			}
			proto, _ := b.ToProto()
			ir, _ := pb.Marshal(proto)
			nb := new(Block)
			pb.Unmarshal(ir, proto)
			err := nb.FromProto(proto)
			assert.Nil(t, err)
			b.header.timestamp = nb.header.timestamp

			if !reflect.DeepEqual(*b.header, *nb.header) {
				t.Errorf("Transaction.Serialize() = %v, want %v", *b.header, *nb.header)
			}
			if !reflect.DeepEqual(*b.transactions[0], *nb.transactions[0]) {
				t.Errorf("Transaction.Serialize() = %v, want %v", *b.transactions[0], *nb.transactions[0])
			}
			if !reflect.DeepEqual(*b.transactions[1], *nb.transactions[1]) {
				t.Errorf("Transaction.Serialize() = %v, want %v", *b.transactions[1], *nb.transactions[1])
			}
		})
	}
}

func TestBlock_LinkParentBlock(t *testing.T) {
	bc := testNeb(t).chain
	genesis := bc.genesisBlock
	assert.Equal(t, genesis.Height(), uint64(1))
	block1 := &Block{
		header: &BlockHeader{
			hash:       []byte("124546"),
			parentHash: GenesisHash,
			stateRoot:  []byte("43656"),
			txsRoot:    []byte("43656"),
			eventsRoot: []byte("43656"),
			consensusRoot: &consensuspb.ConsensusRoot{
				DynastyRoot: []byte("43656"),
			},
			coinbase:  &Address{address: []byte("hello")},
			timestamp: BlockInterval,
			chainID:   100,
		},
		transactions: []*Transaction{},
	}
	assert.Equal(t, block1.Height(), uint64(0))
	assert.Equal(t, block1.LinkParentBlock(bc, genesis), nil)
	assert.Equal(t, block1.Height(), uint64(2))
	assert.Equal(t, block1.ParentHash(), genesis.Hash())
	block2 := &Block{
		header: &BlockHeader{
			hash:       []byte("124546"),
			parentHash: []byte("344543"),
			stateRoot:  []byte("43656"),
			txsRoot:    []byte("43656"),
			eventsRoot: []byte("43656"),
			consensusRoot: &consensuspb.ConsensusRoot{
				DynastyRoot: []byte("43656"),
			},
			coinbase:  &Address{address: []byte("hello")},
			timestamp: BlockInterval * 2,
			chainID:   100,
		},
		transactions: []*Transaction{},
	}
	assert.Equal(t, block2.LinkParentBlock(bc, genesis), ErrLinkToWrongParentBlock)
	assert.Equal(t, block2.Height(), uint64(0))
}

func TestBlock_CollectTransactions(t *testing.T) {
	bc := testNeb(t).chain

	tail := bc.tailBlock

	ks := keystore.DefaultKS
	priv := secp256k1.GeneratePrivateKey()
	pubdata, _ := priv.PublicKey().Encoded()
	from, _ := NewAddressFromPublicKey(pubdata)
	ks.SetKey(from.String(), priv, []byte("passphrase"))
	ks.Unlock(from.String(), []byte("passphrase"), time.Second*60*60*24*365)

	key, _ := ks.GetUnlocked(from.String())
	signature, _ := crypto.NewSignature(keystore.SECP256K1)
	signature.InitSign(key.(keystore.PrivateKey))

	priv1 := secp256k1.GeneratePrivateKey()
	pubdata1, _ := priv1.PublicKey().Encoded()
	to, _ := NewAddressFromPublicKey(pubdata1)
	priv2 := secp256k1.GeneratePrivateKey()
	pubdata2, _ := priv2.PublicKey().Encoded()
	coinbase, _ := NewAddressFromPublicKey(pubdata2)

	block0, err := NewBlock(bc.ChainID(), from, tail)
	assert.Nil(t, err)
	consensusState, err := tail.NextConsensusState(BlockInterval)
	assert.Nil(t, err)
	block0.LoadConsensusState(consensusState)
	block0.Seal()
	assert.Nil(t, bc.BlockPool().Push(block0))

	block, _ := NewBlock(bc.ChainID(), coinbase, block0)
	block.header.timestamp = BlockInterval * 2

	value, _ := util.NewUint128FromInt(1)
	gasLimit, _ := util.NewUint128FromInt(200000)
	tx1, _ := NewTransaction(bc.ChainID(), from, to, value, 1, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, gasLimit)
	tx1.Sign(signature)
	tx2, _ := NewTransaction(bc.ChainID(), from, to, value, 2, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, gasLimit)
	tx2.Sign(signature)
	tx3, _ := NewTransaction(bc.ChainID(), from, to, value, 5, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, gasLimit)
	tx3.Sign(signature)
	tx4, _ := NewTransaction(bc.ChainID(), from, to, value, 4, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, gasLimit)
	tx4.Sign(signature)
	tx5, _ := NewTransaction(bc.ChainID(), from, to, value, 3, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, gasLimit)
	tx5.Sign(signature)
	tx6, _ := NewTransaction(bc.ChainID()+1, from, to, value, 1, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, gasLimit)
	tx6.Sign(signature)

	assert.Nil(t, bc.txPool.Push(tx1))
	assert.Nil(t, bc.txPool.Push(tx2))
	assert.Nil(t, bc.txPool.Push(tx3))
	assert.Nil(t, bc.txPool.Push(tx4))
	assert.Nil(t, bc.txPool.Push(tx5))
	assert.NotNil(t, bc.txPool.Push(tx6), ErrInvalidChainID)

	assert.Equal(t, len(block.transactions), 0)
	assert.Equal(t, len(bc.txPool.all), 5)
	block.CollectTransactions(time.Now().Unix() + 2)
	assert.Equal(t, len(block.transactions), 5)
	assert.Equal(t, len(bc.txPool.all), 0)

	assert.Equal(t, block.Sealed(), false)
	balance, err := block.GetBalance(block.header.coinbase.address)
	assert.Nil(t, err)
	assert.Equal(t, balance.Cmp(util.NewUint128()), 1)
	block.Seal()
	assert.Equal(t, block.Sealed(), true)
	assert.Equal(t, block.transactions[0], tx1)
	assert.Equal(t, block.transactions[1], tx2)
	stateRoot, err := block.accState.RootHash()
	assert.Nil(t, err)
	assert.Equal(t, block.StateRoot().Equals(stateRoot), true)
	assert.Equal(t, block.TxsRoot().Equals(block.txsState.RootHash()), true)
	balance, err = block.GetBalance(block.header.coinbase.address)
	assert.Nil(t, err)
	// balance > BlockReward (BlockReward + gas)
	//gas, _ := bc.EstimateGas(tx1)
	logging.CLog().Info(balance.String())
	logging.CLog().Info(BlockReward.String())
	assert.NotEqual(t, balance.Cmp(BlockReward), 0)
	// mock net message
	block, _ = deepCopyBlock(block)
	assert.Equal(t, block.LinkParentBlock(bc, bc.tailBlock), nil)
	assert.Nil(t, block.VerifyExecution())
}

func TestBlock_fetchEvents(t *testing.T) {
	bc := testNeb(t).chain
	tail := bc.tailBlock
	events := []*Event{
		&Event{Topic: "chain.block", Data: "hello"},
		&Event{Topic: "chain.tx", Data: "hello"},
		&Event{Topic: "chain.block", Data: "hello"},
		&Event{Topic: "chain.block", Data: "hello"},
	}
	tx := &Transaction{hash: []byte("tx")}
	for _, event := range events {
		assert.Nil(t, tail.recordEvent(tx.Hash(), event))
	}
	es, err := tail.FetchEvents(tx.Hash())
	assert.Nil(t, err)
	for idx, event := range es {
		assert.Equal(t, events[idx], event)
	}
}

func TestBlockSign(t *testing.T) {
	bc := testNeb(t).chain
	block := bc.tailBlock
	ks := keystore.DefaultKS
	signature, _ := crypto.NewSignature(keystore.SECP256K1)
	signer := mockAddress()
	key, _ := ks.GetUnlocked(signer.String())
	signature.InitSign(key.(keystore.PrivateKey))
	assert.Nil(t, block.Sign(signature))
	assert.Equal(t, block.Alg(), keystore.Algorithm(keystore.SECP256K1))
	assert.Equal(t, block.Signature(), block.header.sign)
}

func TestGivebackInvalidTx(t *testing.T) {
	bc := testNeb(t).chain
	from := mockAddress()
	ks := keystore.DefaultKS
	gasLimit, _ := util.NewUint128FromInt(200000)
	tx, _ := NewTransaction(bc.ChainID(), from, from, util.NewUint128(), 2, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, gasLimit)
	key, err := ks.GetUnlocked(from.String())
	assert.Nil(t, err)
	signature, err := crypto.NewSignature(keystore.SECP256K1)
	assert.Nil(t, err)
	signature.InitSign(key.(keystore.PrivateKey))
	tx.Sign(signature)
	assert.Nil(t, bc.txPool.Push(tx))
	assert.Equal(t, len(bc.txPool.all), 1)
	block, err := bc.NewBlock(from)
	assert.Nil(t, err)
	block.CollectTransactions(time.Now().Unix() + 2)
	timer := time.NewTimer(time.Second).C
	<-timer
	assert.Equal(t, len(bc.txPool.all), 1)
}

func TestRecordEvent(t *testing.T) {
	bc := testNeb(t).chain
	txHash := []byte("hello")
	assert.Nil(t, bc.tailBlock.RecordEvent(txHash, TopicSendTransaction, "world"))
	events, err := bc.tailBlock.FetchEvents(txHash)
	assert.Nil(t, err)
	assert.Equal(t, len(events), 1)
	assert.Equal(t, events[0].Topic, TopicSendTransaction)
	assert.Equal(t, events[0].Data, "world")
}

func TestBlockVerifyIntegrity(t *testing.T) {
	bc := testNeb(t).chain
	assert.Equal(t, bc.tailBlock.VerifyIntegrity(0, bc.ConsensusHandler()), ErrInvalidChainID)
	bc.tailBlock.header.hash[0] = 1
	assert.Equal(t, bc.tailBlock.VerifyIntegrity(bc.ChainID(), bc.ConsensusHandler()), ErrInvalidBlockHash)
	ks := keystore.DefaultKS
	from := mockAddress()
	key, err := ks.GetUnlocked(from.String())
	assert.Nil(t, err)
	signature, err := crypto.NewSignature(keystore.SECP256K1)
	assert.Nil(t, err)
	signature.InitSign(key.(keystore.PrivateKey))
	block, err := bc.NewBlock(from)
	assert.Nil(t, err)
	gasLimit, _ := util.NewUint128FromInt(200000)
	tx1, _ := NewTransaction(bc.ChainID(), from, from, util.NewUint128(), 1, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, gasLimit)
	tx1.Sign(signature)
	tx2, _ := NewTransaction(bc.ChainID(), from, from, util.NewUint128(), 2, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, gasLimit)
	tx2.Sign(signature)
	tx2.hash[0]++
	block.transactions = append(block.transactions, tx1)
	block.transactions = append(block.transactions, tx2)
	block.Seal()
	block.Sign(signature)
	assert.NotNil(t, block.VerifyIntegrity(bc.ChainID(), bc.ConsensusHandler()))
}

func TestBlockVerifyIntegrityDup(t *testing.T) {
	bc := testNeb(t).chain
	assert.Equal(t, bc.tailBlock.VerifyIntegrity(0, bc.ConsensusHandler()), ErrInvalidChainID)
	bc.tailBlock.header.hash[0] = 1
	assert.Equal(t, bc.tailBlock.VerifyIntegrity(bc.ChainID(), bc.ConsensusHandler()), ErrInvalidBlockHash)
	ks := keystore.DefaultKS
	from := mockAddress()
	key, err := ks.GetUnlocked(from.String())
	assert.Nil(t, err)
	signature, err := crypto.NewSignature(keystore.SECP256K1)
	assert.Nil(t, err)
	signature.InitSign(key.(keystore.PrivateKey))
	block, err := bc.NewBlock(from)
	assert.Nil(t, err)
	gasLimit, _ := util.NewUint128FromInt(200000)
	tx1, _ := NewTransaction(bc.ChainID(), from, from, util.NewUint128(), 1, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, gasLimit)
	tx1.Sign(signature)
	block.transactions = append(block.transactions, tx1)
	block.transactions = append(block.transactions, tx1)
	block.Seal()
	block.Sign(signature)
	assert.Equal(t, block.VerifyExecution(), ErrSmallTransactionNonce)
}

func TestBlockVerifyExecution(t *testing.T) {
	bc := testNeb(t).chain
	assert.Equal(t, bc.tailBlock.VerifyIntegrity(0, bc.ConsensusHandler()), ErrInvalidChainID)
	bc.tailBlock.header.hash[0] = 1
	assert.Equal(t, bc.tailBlock.VerifyIntegrity(bc.ChainID(), bc.ConsensusHandler()), ErrInvalidBlockHash)
	ks := keystore.DefaultKS
	from := mockAddress()
	key, err := ks.GetUnlocked(from.String())
	assert.Nil(t, err)
	signature, err := crypto.NewSignature(keystore.SECP256K1)
	assert.Nil(t, err)
	signature.InitSign(key.(keystore.PrivateKey))
	block, err := bc.NewBlock(from)
	assert.Nil(t, err)
	gasLimit, _ := util.NewUint128FromInt(200000)
	tx1, _ := NewTransaction(bc.ChainID(), from, from, util.NewUint128(), 1, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, gasLimit)
	tx1.Sign(signature)
	tx2, _ := NewTransaction(bc.ChainID(), from, from, util.NewUint128(), 3, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, gasLimit)
	tx2.Sign(signature)
	block.transactions = append(block.transactions, tx1)
	block.transactions = append(block.transactions, tx2)
	block.Seal()
	block.Sign(signature)
	assert.Nil(t, block.VerifyIntegrity(bc.ChainID(), bc.ConsensusHandler()))
	root1, err := block.accState.RootHash()
	assert.Nil(t, err)
	assert.Equal(t, block.VerifyExecution(), ErrLargeTransactionNonce)
	root2, err := block.accState.RootHash()
	assert.Nil(t, err)
	assert.Equal(t, root1, root2)
}

func TestBlockVerifyState(t *testing.T) {
	bc := testNeb(t).chain
	assert.Equal(t, bc.tailBlock.VerifyIntegrity(0, bc.ConsensusHandler()), ErrInvalidChainID)
	bc.tailBlock.header.hash[0] = 1
	assert.Equal(t, bc.tailBlock.VerifyIntegrity(bc.ChainID(), bc.ConsensusHandler()), ErrInvalidBlockHash)
	ks := keystore.DefaultKS
	from := mockAddress()
	key, err := ks.GetUnlocked(from.String())
	assert.Nil(t, err)
	signature, err := crypto.NewSignature(keystore.SECP256K1)
	assert.Nil(t, err)
	signature.InitSign(key.(keystore.PrivateKey))
	block, err := bc.NewBlock(from)
	assert.Nil(t, err)
	gasLimit, _ := util.NewUint128FromInt(200000)
	tx1, _ := NewTransaction(bc.ChainID(), from, from, util.NewUint128(), 1, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, gasLimit)
	tx1.Sign(signature)
	tx2, _ := NewTransaction(bc.ChainID(), from, from, util.NewUint128(), 2, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, gasLimit)
	tx2.Sign(signature)
	block.transactions = append(block.transactions, tx1)
	block.transactions = append(block.transactions, tx2)
	block.Seal()
	block.Sign(signature)
	assert.Nil(t, block.VerifyIntegrity(bc.ChainID(), bc.ConsensusHandler()))
	block.header.stateRoot[0]++
	assert.NotNil(t, block.VerifyExecution())
}
