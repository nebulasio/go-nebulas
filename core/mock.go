// Copyright (C) 2018 go-nebulas authors
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
	"time"

	consensuspb "github.com/nebulasio/go-nebulas/consensus/pb"
	corepb "github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/keystore/secp256k1"
	nebletpb "github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

const (
	AcceptedNetWorkDelay = 2
)

var (
	MockDynasty = []string{
		"n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE",
		"n1GmkKH6nBMw4rrjt16RrJ9WcgvKUtAZP1s",
		"n1H4MYms9F55ehcvygwWE71J8tJC4CRr2so",
		"n1JAy4X6KKLCNiTd7MWMRsVBjgdVq5WCCpf",
		"n1LkDi2gGMqPrjYcczUiweyP4RxTB6Go1qS",
		"n1LmP9K8pFF33fgdgHZonFEMsqZinJ4EUqk",
		"n1MNXBKm6uJ5d76nJTdRvkPNVq85n6CnXAi",
		"n1NrMKTYESZRCwPFDLFKiKREzZKaN1nhQvz",
		"n1NwoSCDFwFL2981k6j9DPooigW33hjAgTa",
		"n1PfACnkcfJoNm1Pbuz55pQCwueW1BYs83m",
		"n1Q8mxXp4PtHaXtebhY12BnHEwu4mryEkXH",
		"n1RYagU8n3JSuV4R7q4Qs5gQJ3pEmrZd6cJ",
		"n1SAQy3ix1pZj8MPzNeVqpAmu1nCVqb5w8c",
		"n1SHufJdxt2vRWGKAxwPETYfEq3MCQXnEXE",
		"n1SSda41zGr9FKF5DJNE2ryY1ToNrndMauN",
		"n1TmQtaCn3PNpk4f4ycwrBxCZFSVKvwBtzc",
		"n1UM7z6MqnGyKEPvUpwrfxZpM1eB7UpzmLJ",
		"n1UnCsJZjQiKyQiPBr7qG27exqCLuWUf1d7",
		"n1XkoVVjswb5Gek3rRufqjKNpwrDdsnQ7Hq",
		"n1cYKNHTeVW9v1NQRWuhZZn9ETbqAYozckh",
		"n1dYu2BXgV3xgUh8LhZu8QDDNr15tz4hVDv",
	}
)

func mockAddress() *Address {
	ks := keystore.DefaultKS
	priv1 := secp256k1.GeneratePrivateKey()
	pubdata1, _ := priv1.PublicKey().Encoded()
	addr, _ := NewAddressFromPublicKey(pubdata1)
	ks.SetKey(addr.String(), priv1, []byte("passphrase"))
	ks.Unlock(addr.String(), []byte("passphrase"), time.Second*60*60*24*365)
	return addr
}

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
				Address: "n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE",
				Value:   "5000000000000000000000000",
			},
			&corepb.GenesisTokenDistribution{
				Address: "n1GmkKH6nBMw4rrjt16RrJ9WcgvKUtAZP1s",
				Value:   "5000000000000000000000000",
			},
			&corepb.GenesisTokenDistribution{
				Address: "n1H4MYms9F55ehcvygwWE71J8tJC4CRr2so",
				Value:   "5000000000000000000000000",
			},
			&corepb.GenesisTokenDistribution{
				Address: "n1JAy4X6KKLCNiTd7MWMRsVBjgdVq5WCCpf",
				Value:   "5000000000000000000000000",
			},
			&corepb.GenesisTokenDistribution{
				Address: "n1LkDi2gGMqPrjYcczUiweyP4RxTB6Go1qS",
				Value:   "5000000000000000000000000",
			},
			&corepb.GenesisTokenDistribution{
				Address: "n1LmP9K8pFF33fgdgHZonFEMsqZinJ4EUqk",
				Value:   "5000000000000000000000000",
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

func (cs *mockConsensusState) RootHash() *consensuspb.ConsensusRoot {
	return &consensuspb.ConsensusRoot{}
}
func (cs *mockConsensusState) String() string { return "" }
func (cs *mockConsensusState) Clone() (state.ConsensusState, error) {
	return &mockConsensusState{
		timestamp: cs.timestamp,
	}, nil
}
func (cs *mockConsensusState) Replay(state.ConsensusState) error { return nil }

func (cs *mockConsensusState) Proposer() byteutils.Hash { return nil }
func (cs *mockConsensusState) TimeStamp() int64         { return 0 }
func (cs *mockConsensusState) NextConsensusState(elapsed int64, ws state.WorldState) (state.ConsensusState, error) {
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

// Serial return dynasty serial number
func (pod *mockConsensus) Serial(timestamp int64) int64 {
	return 0
}

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

func (c *mockConsensus) UpdateLIB(rversibleBlocks []byteutils.Hash) {}

func (c *mockConsensus) SuspendMining() {}
func (c *mockConsensus) ResumeMining()  {}
func (c *mockConsensus) Pending() bool  { return false }

func (c *mockConsensus) EnableMining(passphrase string) error { return nil }
func (c *mockConsensus) DisableMining() error                 { return nil }
func (c *mockConsensus) Enable() bool                         { return true }

func (c *mockConsensus) CheckTimeout(block *Block) bool {
	return time.Now().Unix()-block.Timestamp() > AcceptedNetWorkDelay
}
func (c *mockConsensus) CheckDoubleMint(block *Block) bool {
	return false
}
func (c *mockConsensus) NewState(root *consensuspb.ConsensusRoot, stor storage.Storage, needChangeLog bool) (state.ConsensusState, error) {
	return newMockConsensusState(root.Timestamp)
}
func (c *mockConsensus) GenesisConsensusState(*BlockChain, *corepb.Genesis) (state.ConsensusState, error) {
	return newMockConsensusState(0)
}
func (c *mockConsensus) NumberOfBlocksInDynasty() uint64 {
	return 210
}

type mockManager struct{}

func (m mockManager) NewAccount([]byte) (*Address, error) { return nil, nil }
func (m mockManager) Accounts() []*Address                { return nil }

func (m mockManager) Unlock(addr *Address, passphrase []byte, expire time.Duration) error { return nil }
func (m mockManager) Lock(addr *Address) error                                            { return nil }

func (m mockManager) SignHash(addr *Address, hash byteutils.Hash, alg keystore.Algorithm) ([]byte, error) {
	return nil, nil
}
func (m mockManager) SignBlock(addr *Address, block *Block) error                        { return nil }
func (m mockManager) SignTransaction(*Address, *Transaction) error                       { return nil }
func (m mockManager) SignTransactionWithPassphrase(*Address, *Transaction, []byte) error { return nil }

func (m mockManager) Update(*Address, []byte, []byte) error        { return nil }
func (m mockManager) Load([]byte, []byte) (*Address, error)        { return nil, nil }
func (m mockManager) LoadPrivate([]byte, []byte) (*Address, error) { return nil, nil }
func (m mockManager) Import([]byte, []byte) (*Address, error)      { return nil, nil }
func (m mockManager) Remove(*Address, []byte) error                { return nil }
func (m mockManager) GenerateRandomSeed(addr *Address, ancestorHash, parentSeed []byte) (vrfSeed, vrfProof []byte, err error) {
	return nil, nil, nil
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

func (n MockNetService) Broadcast(name string, msg net.Serializable, priority int) {}
func (n MockNetService) Relay(name string, msg net.Serializable, priority int)     {}
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

type MockNeb struct {
	config    *nebletpb.Config
	chain     *BlockChain
	ns        net.Service
	am        AccountManager
	genesis   *corepb.Genesis
	storage   storage.Storage
	consensus Consensus
	emitter   *EventEmitter
	nvm       NVM
	dip       Dip
	nbre      Nbre
}

func (n *MockNeb) Genesis() *corepb.Genesis {
	return n.genesis
}

func (n *MockNeb) Config() *nebletpb.Config {
	return n.config
}

func (n *MockNeb) SetStorage(s storage.Storage) {
	n.storage = s
}

func (n *MockNeb) Storage() storage.Storage {
	return n.storage
}

func (n *MockNeb) EventEmitter() *EventEmitter {
	return n.emitter
}

func (n *MockNeb) SetConsensus(c Consensus) {
	n.consensus = c
}

func (n *MockNeb) Consensus() Consensus {
	return n.consensus
}

func (n *MockNeb) SetBlockChain(bc *BlockChain) {
	n.chain = bc
}

func (n *MockNeb) BlockChain() *BlockChain {
	return n.chain
}

func (n *MockNeb) NetService() net.Service {
	return n.ns
}

func (n *MockNeb) IsActiveSyncing() bool {
	return true
}

func (n *MockNeb) SetAccountManager(a AccountManager) {
	n.am = a
}

func (n *MockNeb) AccountManager() AccountManager {
	return n.am
}

func (n *MockNeb) Nvm() NVM {
	return n.nvm
}

func (n *MockNeb) Nbre() Nbre {
	return nil
}

func (n *MockNeb) Dip() Dip {
	return n.dip
}

func (n *MockNeb) Nr() NR {
	return nil
}

func (n *MockNeb) StartPprof(string) error {
	return nil
}

func (n *MockNeb) SetGenesis(genesis *corepb.Genesis) {
	n.genesis = genesis
}

type mockNvm struct{}
type mockEngine struct{}

func (nvm *mockNvm) CreateEngine(block *Block, tx *Transaction, contract state.Account, state WorldState) (SmartContractEngine, error) {
	return &mockEngine{}, nil
}
func (nvm *mockNvm) CheckV8Run() error {
	return nil
}

func (nvm *mockEngine) Dispose() {

}
func (nvm *mockEngine) SetExecutionLimits(uint64, uint64) error {
	return nil
}
func (nvm *mockEngine) DeployAndInit(source, sourceType, args string) (string, error) {
	return "", nil
}
func (nvm *mockEngine) Call(source, sourceType, function, args string) (string, error) {
	return "", nil
}
func (nvm *mockEngine) ExecutionInstructions() uint64 {
	return uint64(100)
}

type mockDip struct {
	addr *Address
}

func (m *mockDip) Start() {}
func (m *mockDip) Stop()  {}

func (m *mockDip) RewardAddress() *Address {
	if m.addr == nil {
		m.addr = mockAddress()
	}
	return m.addr
}

func (m *mockDip) RewardValue() *util.Uint128 {
	return util.NewUint128()
}

func (m *mockDip) GetDipList(height uint64, version uint64) (Data, error) {
	return nil, nil
}

func (m *mockDip) CheckReward(tx *Transaction) error {
	return nil
}

// NewMockNeb create mock neb for unit testing
func NewMockNeb(am AccountManager, consensus Consensus, nvm NVM) *MockNeb {
	storage, _ := storage.NewMemoryStorage()
	eventEmitter := NewEventEmitter(1024)
	if am == nil {
		am = new(mockManager)
	}
	if consensus == nil {
		consensus = new(mockConsensus)
	}
	if nvm == nil {
		nvm = &mockNvm{}
	}

	dip := &mockDip{}
	var ns MockNetService
	neb := &MockNeb{
		genesis: MockGenesisConf(),
		config: &nebletpb.Config{Chain: &nebletpb.ChainConfig{
			ChainId:    MockGenesisConf().Meta.ChainId,
			Keydir:     "keydir",
			StartMine:  true,
			Coinbase:   "n1dYu2BXgV3xgUh8LhZu8QDDNr15tz4hVDv",
			Miner:      "n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE",
			Passphrase: "passphrase"}},
		storage:   storage,
		emitter:   eventEmitter,
		consensus: consensus,
		am:        am,
		ns:        ns,
		nvm:       nvm,
		dip:       dip,
	}

	chain, _ := NewBlockChain(neb)
	chain.BlockPool().RegisterInNetwork(neb.ns)
	neb.chain = chain
	consensus.Setup(neb)
	chain.Setup(neb)

	return neb
}
