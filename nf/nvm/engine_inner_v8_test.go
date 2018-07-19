package nvm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/account"
	"github.com/nebulasio/go-nebulas/consensus/dpos"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/net"

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/stretchr/testify/assert"
)

// const contractStr = "n218MQSwc7hcXvM7rUkr6smMoiEf2VbGuYr"

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
	// storage, _ := storage.NewDiskStorage("test.db")
	// storage, err := storage.NewRocksStorage("rocks.db")
	// assert.Nil(t, err)
	storage, _ := storage.NewMemoryStorage()
	eventEmitter := core.NewEventEmitter(1024)
	genesisConf := MockGenesisConf()
	consensus := dpos.NewDpos()
	nvm := NewNebulasVM()
	neb := &Neb{
		genesis:   genesisConf,
		storage:   storage,
		emitter:   eventEmitter,
		consensus: consensus,
		nvm:       nvm,
		config: &nebletpb.Config{
			Chain: &nebletpb.ChainConfig{
				ChainId:    genesisConf.Meta.ChainId,
				Keydir:     "keydir",
				StartMine:  true,
				Coinbase:   "n1dYu2BXgV3xgUh8LhZu8QDDNr15tz4hVDv",
				Miner:      "n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE",
				Passphrase: "passphrase",
			},
		},
		ns: mockNetService{},
	}

	am, _ := account.NewManager(neb)
	neb.am = am

	chain, err := core.NewBlockChain(neb)
	assert.Nil(t, err)

	neb.chain = chain
	assert.Nil(t, consensus.Setup(neb))
	assert.Nil(t, chain.Setup(neb))

	var ns mockNetService
	neb.ns = ns
	neb.chain.BlockPool().RegisterInNetwork(ns)

	eventEmitter.Start()
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

func (n *Neb) Nvm() core.NVM {
	return n.nvm
}

func (n *Neb) StartActiveSync() {}

func (n *Neb) StartPprof(string) error { return nil }

func (n *Neb) SetGenesis(genesis *corepb.Genesis) {
	n.genesis = genesis
}

var (
	DefaultOpenDynasty = []string{
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

type contract struct {
	contractPath string
	sourceType   string
	initArgs     string
}

type call struct {
	function   string
	args       string
	exceptArgs []string //[congractA,B,C, AccountA, B]
}

func TestInnerTransactions(t *testing.T) {
	core.NebCompatibility = core.NewCompatibilityLocal()
	tests := []struct {
		name      string
		contracts []contract
		calls     []call
	}{
		{
			"deploy test_require_module.js",
			[]contract{
				contract{
					"./test/test_inner_transaction.js",
					"js",
					"",
				},
				contract{
					"./test/bank_vault_contract_second.js",
					"js",
					"",
				},
				contract{
					"./test/bank_vault_contract.js",
					"js",
					"",
				},
			},
			[]call{
				call{
					"save",
					"[1]",
					[]string{"1", "3", "2", "4999999999999833224999994", "5000004280820166775000000"},
				},
			},
		},
	}
	tt := tests[0]
	for _, call := range tt.calls {

		neb := mockNeb(t)
		tail := neb.chain.TailBlock()
		manager, err := account.NewManager(neb)
		assert.Nil(t, err)

		a, _ := core.AddressParse("n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE")
		assert.Nil(t, manager.Unlock(a, []byte("passphrase"), keystore.YearUnlockDuration))
		b, _ := core.AddressParse("n1GmkKH6nBMw4rrjt16RrJ9WcgvKUtAZP1s")
		assert.Nil(t, manager.Unlock(b, []byte("passphrase"), keystore.YearUnlockDuration))
		c, _ := core.AddressParse("n1H4MYms9F55ehcvygwWE71J8tJC4CRr2so")
		assert.Nil(t, manager.Unlock(c, []byte("passphrase"), keystore.YearUnlockDuration))
		d, _ := core.AddressParse("n1JAy4X6KKLCNiTd7MWMRsVBjgdVq5WCCpf")
		assert.Nil(t, manager.Unlock(d, []byte("passphrase"), keystore.YearUnlockDuration))

		elapsedSecond := dpos.BlockIntervalInMs / dpos.SecondInMs
		consensusState, err := tail.WorldState().NextConsensusState(elapsedSecond)
		assert.Nil(t, err)
		// mock empty block(height=2)
		block, err := core.MockBlockEx(neb.chain.ChainID(), c, tail, 2)
		fmt.Printf("mock 2, block.height:%v\n", block.Height())
		assert.Nil(t, err)
		/* ----- mock random seed for new block ------*/
		miner, err := core.AddressParseFromBytes(consensusState.Proposer())
		assert.Nil(t, err)
		seed, proof, err := manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), neb.chain.GenesisBlock().Hash()) // NOTE: 3rd arg is genesis's hash for the first block
		assert.Nil(t, err)
		block.SetRandomSeed(seed, proof)
		/* ----- mock random seed for new block END ------*/
		block.WorldState().SetConsensusState(consensusState)
		block.SetTimestamp(consensusState.TimeStamp())
		block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
		assert.Nil(t, block.Seal())
		assert.Nil(t, manager.SignBlock(b, block))
		assert.Nil(t, neb.chain.BlockPool().Push(block))
		fmt.Printf("mock 2, block.tailblock.height: %v\n", neb.chain.TailBlock().Height())

		// inner call block(height=3)
		tail = neb.chain.TailBlock()
		block, err = core.MockBlockEx(neb.chain.ChainID(), c, tail, 3)
		assert.Nil(t, err)
		consensusState, err = tail.WorldState().NextConsensusState(elapsedSecond)
		assert.Nil(t, err)
		/* ----- mock random seed for new block ------*/
		miner, err = core.AddressParseFromBytes(consensusState.Proposer())
		assert.Nil(t, err)
		seed, proof, err = manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), seed) // NOTE: 3rd arg is parent's seed
		assert.Nil(t, err)
		block.SetRandomSeed(seed, proof)
		/* ----- mock random seed for new block END ------*/
		block.WorldState().SetConsensusState(consensusState)
		block.SetTimestamp(consensusState.TimeStamp())

		contractsAddr := []string{}

		// t.Run(tt.name, func(t *testing.T) {
		for k, v := range tt.contracts {
			data, err := ioutil.ReadFile(v.contractPath)
			assert.Nil(t, err, "contract path read error")
			source := string(data)
			sourceType := "js"
			argsDeploy := ""
			deploy, _ := core.NewDeployPayload(source, sourceType, argsDeploy)
			payloadDeploy, _ := deploy.ToBytes()

			value, _ := util.NewUint128FromInt(0)
			gasLimit, _ := util.NewUint128FromInt(200000)
			txDeploy, err := core.NewTransaction(neb.chain.ChainID(), a, a, value, uint64(k+1), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
			assert.Nil(t, err)
			assert.Nil(t, manager.SignTransaction(a, txDeploy))
			assert.Nil(t, neb.chain.TransactionPool().Push(txDeploy))

			contractAddr, err := txDeploy.GenerateContractAddress()
			assert.Nil(t, err)
			contractsAddr = append(contractsAddr, contractAddr.String())
		}
		// })

		block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
		assert.Nil(t, block.Seal())
		assert.Nil(t, manager.SignBlock(c, block))
		assert.Nil(t, neb.chain.BlockPool().Push(block))

		for _, v := range contractsAddr {
			contract, err := core.AddressParse(v)
			assert.Nil(t, err)
			_, err = neb.chain.TailBlock().CheckContract(contract)
			assert.Nil(t, err)
		}

		elapsedSecond = dpos.BlockIntervalInMs / dpos.SecondInMs
		tail = neb.chain.TailBlock()
		consensusState, err = tail.WorldState().NextConsensusState(elapsedSecond)
		assert.Nil(t, err)
		block, err = core.NewBlock(neb.chain.ChainID(), c, tail)
		assert.Nil(t, err)
		/* ----- mock random seed for new block ------*/
		miner, err = core.AddressParseFromBytes(consensusState.Proposer())
		assert.Nil(t, err)
		seed, proof, err = manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), seed) // NOTE: 3rd arg is parent's seed
		assert.Nil(t, err)
		block.SetRandomSeed(seed, proof)
		/* ----- mock random seed for new block END ------*/
		block.WorldState().SetConsensusState(consensusState)
		block.SetTimestamp(consensusState.TimeStamp())
		//accountA, err := tail.GetAccount(a.Bytes())
		//accountB, err := tail.GetAccount(b.Bytes())
		assert.Nil(t, err)

		calleeContract := contractsAddr[1]
		callToContract := contractsAddr[2]
		callPayload, _ := core.NewCallPayload(call.function, fmt.Sprintf("[\"%s\", \"%s\", 1]", calleeContract, callToContract))
		payloadCall, _ := callPayload.ToBytes()

		value, _ := util.NewUint128FromInt(6)
		gasLimit, _ := util.NewUint128FromInt(200000)

		proxyContractAddress, err := core.AddressParse(contractsAddr[0])
		txCall, err := core.NewTransaction(neb.chain.ChainID(), a, proxyContractAddress, value,
			uint64(len(contractsAddr)+1), core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
		assert.Nil(t, err)
		assert.Nil(t, manager.SignTransaction(a, txCall))
		assert.Nil(t, neb.chain.TransactionPool().Push(txCall))

		block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
		assert.Nil(t, block.Seal())
		assert.Nil(t, manager.SignBlock(d, block))
		assert.Nil(t, neb.chain.BlockPool().Push(block))

		// check
		tail = neb.chain.TailBlock()
		// event, err := tail.FetchExecutionResultEvent(txCall.Hash())
		// assert.Nil(t, err)
		// txEvent := core.TransactionEvent{}
		// err = json.Unmarshal([]byte(event.Data), &txEvent)
		// assert.Nil(t, err)
		// // if txEvent.Status != 1 {
		// // 	fmt.Println(txEvent)
		// // }
		// fmt.Println("=====================", txEvent)

		events, err := tail.FetchEvents(txCall.Hash())
		assert.Nil(t, err)
		for _, event := range events {

			fmt.Println("==============", event.Data)
		}
		contractAddrA, err := core.AddressParse(contractsAddr[0])
		accountAAcc, err := tail.GetAccount(contractAddrA.Bytes())
		assert.Nil(t, err)
		fmt.Printf("account :%v\n", accountAAcc)
		assert.Equal(t, call.exceptArgs[0], accountAAcc.Balance().String())

		contractAddrB, err := core.AddressParse(contractsAddr[1])
		accountBAcc, err := tail.GetAccount(contractAddrB.Bytes())
		assert.Nil(t, err)
		fmt.Printf("accountB :%v\n", accountBAcc)
		assert.Equal(t, call.exceptArgs[1], accountBAcc.Balance().String())

		contractAddrC, err := core.AddressParse(contractsAddr[2])
		accountAccC, err := tail.GetAccount(contractAddrC.Bytes())
		assert.Nil(t, err)
		fmt.Printf("accountC :%v\n", accountAccC)
		assert.Equal(t, call.exceptArgs[2], accountAccC.Balance().String())

		aUser, err := tail.GetAccount(a.Bytes())
		assert.Equal(t, call.exceptArgs[3], aUser.Balance().String())
		fmt.Printf("aI:%v\n", aUser)
		cUser, err := tail.GetAccount(c.Bytes())
		fmt.Printf("b:%v\n", cUser)
		assert.Equal(t, call.exceptArgs[4], cUser.Balance().String())

		// cUser, err := tail.GetAccount(c.Bytes())
		// fmt.Printf("c:%v\n", cUser)

		// dUser, err := tail.GetAccount(d.Bytes())
		// fmt.Printf("d:%v\n", dUser)
		// assert.Equal(t, txEvent.Status, 1)
	}
}

func TestInnerTransactionsMaxMulit(t *testing.T) {
	core.NebCompatibility = core.NewCompatibilityLocal()
	tests := []struct {
		name                   string
		contracts              []contract
		call                   call
		innerExpectedErr       string
		contractExpectedErr    string
		contractExpectedResult string
	}{
		{
			"deploy test_require_module.js",
			[]contract{
				contract{
					"./test/test_inner_transaction.js",
					"js",
					"",
				},
				contract{
					"./test/bank_vault_contract_second.js",
					"js",
					"",
				},
				contract{
					"./test/bank_vault_contract.js",
					"js",
					"",
				},
			},
			call{
				"saveToLoop",
				"[1]",
				[]string{""},
			},
			"multi execution failed",
			"Call: Inner Contract: out of limit nvm count",
			"Inner Contract: out of limit nvm count",
		},
	}

	for _, tt := range tests {
		neb := mockNeb(t)
		tail := neb.chain.TailBlock()
		manager, err := account.NewManager(neb)
		assert.Nil(t, err)

		a, _ := core.AddressParse("n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE")
		assert.Nil(t, manager.Unlock(a, []byte("passphrase"), keystore.YearUnlockDuration))
		b, _ := core.AddressParse("n1GmkKH6nBMw4rrjt16RrJ9WcgvKUtAZP1s")
		assert.Nil(t, manager.Unlock(b, []byte("passphrase"), keystore.YearUnlockDuration))
		c, _ := core.AddressParse("n1H4MYms9F55ehcvygwWE71J8tJC4CRr2so")
		assert.Nil(t, manager.Unlock(c, []byte("passphrase"), keystore.YearUnlockDuration))
		d, _ := core.AddressParse("n1JAy4X6KKLCNiTd7MWMRsVBjgdVq5WCCpf")
		assert.Nil(t, manager.Unlock(d, []byte("passphrase"), keystore.YearUnlockDuration))

		elapsedSecond := dpos.BlockIntervalInMs / dpos.SecondInMs
		consensusState, err := tail.WorldState().NextConsensusState(elapsedSecond)
		assert.Nil(t, err)
		// mock empty block(height=2)
		block, err := core.MockBlockEx(neb.chain.ChainID(), c, tail, 2)
		fmt.Printf("mock 2, block.height:%v\n", block.Height())
		assert.Nil(t, err)
		/* ----- mock random seed for new block ------*/
		miner, err := core.AddressParseFromBytes(consensusState.Proposer())
		assert.Nil(t, err)
		seed, proof, err := manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), neb.chain.GenesisBlock().Hash()) // NOTE: 3rd arg is genesis's hash for the first block
		assert.Nil(t, err)
		block.SetRandomSeed(seed, proof)
		/* ----- mock random seed for new block END ------*/
		block.WorldState().SetConsensusState(consensusState)
		block.SetTimestamp(consensusState.TimeStamp())
		block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
		assert.Nil(t, block.Seal())
		assert.Nil(t, manager.SignBlock(b, block))
		assert.Nil(t, neb.chain.BlockPool().Push(block))
		fmt.Printf("mock 2, block.tailblock.height: %v\n", neb.chain.TailBlock().Height())

		// inner call block(height=3)
		tail = neb.chain.TailBlock()
		block, err = core.MockBlockEx(neb.chain.ChainID(), c, tail, 3)
		assert.Nil(t, err)
		consensusState, err = tail.WorldState().NextConsensusState(elapsedSecond)
		assert.Nil(t, err)
		/* ----- mock random seed for new block ------*/
		miner, err = core.AddressParseFromBytes(consensusState.Proposer())
		assert.Nil(t, err)
		seed, proof, err = manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), seed) // NOTE: 3rd arg is parent's seed
		assert.Nil(t, err)
		block.SetRandomSeed(seed, proof)
		/* ----- mock random seed for new block END ------*/
		block.WorldState().SetConsensusState(consensusState)
		block.SetTimestamp(consensusState.TimeStamp())

		contractsAddr := []string{}
		fmt.Printf("++++++++++++pack account")
		// t.Run(tt.name, func(t *testing.T) {
		for k, v := range tt.contracts {
			data, err := ioutil.ReadFile(v.contractPath)
			assert.Nil(t, err, "contract path read error")
			source := string(data)
			sourceType := "js"
			argsDeploy := ""
			deploy, _ := core.NewDeployPayload(source, sourceType, argsDeploy)
			payloadDeploy, _ := deploy.ToBytes()

			value, _ := util.NewUint128FromInt(0)
			gasLimit, _ := util.NewUint128FromInt(200000)
			txDeploy, err := core.NewTransaction(neb.chain.ChainID(), a, a, value, uint64(k+1), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
			assert.Nil(t, err)
			assert.Nil(t, manager.SignTransaction(a, txDeploy))
			assert.Nil(t, neb.chain.TransactionPool().Push(txDeploy))

			contractAddr, err := txDeploy.GenerateContractAddress()
			assert.Nil(t, err)
			contractsAddr = append(contractsAddr, contractAddr.String())
		}
		// })

		block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
		assert.Nil(t, block.Seal())
		assert.Nil(t, manager.SignBlock(c, block))
		assert.Nil(t, neb.chain.BlockPool().Push(block))

		for _, v := range contractsAddr {
			contract, err := core.AddressParse(v)
			assert.Nil(t, err)
			_, err = neb.chain.TailBlock().CheckContract(contract)
			assert.Nil(t, err)
		}

		elapsedSecond = dpos.BlockIntervalInMs / dpos.SecondInMs
		tail = neb.chain.TailBlock()
		consensusState, err = tail.WorldState().NextConsensusState(elapsedSecond)
		assert.Nil(t, err)
		block, err = core.NewBlock(neb.chain.ChainID(), b, tail)
		assert.Nil(t, err)
		/* ----- mock random seed for new block ------*/
		miner, err = core.AddressParseFromBytes(consensusState.Proposer())
		assert.Nil(t, err)
		seed, proof, err = manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), seed) // NOTE: 3rd arg is parent's seed
		assert.Nil(t, err)
		block.SetRandomSeed(seed, proof)
		/* ----- mock random seed for new block END ------*/
		block.WorldState().SetConsensusState(consensusState)
		block.SetTimestamp(consensusState.TimeStamp())
		//accountA, err := tail.GetAccount(a.Bytes())
		//accountB, err := tail.GetAccount(b.Bytes())
		assert.Nil(t, err)

		calleeContract := contractsAddr[0]
		callToContract := contractsAddr[2]
		fmt.Printf("++++++++++++pack payload")
		callPayload, _ := core.NewCallPayload(tt.call.function, fmt.Sprintf("[\"%s\", \"%s\", 1]", calleeContract, callToContract))
		payloadCall, _ := callPayload.ToBytes()

		value, _ := util.NewUint128FromInt(6)
		gasLimit, _ := util.NewUint128FromInt(2000000)

		proxyContractAddress, err := core.AddressParse(contractsAddr[0])
		fmt.Printf("++++++++++++pack transaction")
		txCall, err := core.NewTransaction(neb.chain.ChainID(), a, proxyContractAddress, value,
			uint64(len(contractsAddr)+1), core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
		assert.Nil(t, err)
		assert.Nil(t, manager.SignTransaction(a, txCall))
		assert.Nil(t, neb.chain.TransactionPool().Push(txCall))

		fmt.Printf("++++++++++++pack collect")
		block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
		assert.Nil(t, block.Seal())
		assert.Nil(t, manager.SignBlock(d, block))
		assert.Nil(t, neb.chain.BlockPool().Push(block))

		fmt.Printf("++++++++++++pack check\n")
		// check
		tail = neb.chain.TailBlock()

		events, err := tail.FetchEvents(txCall.Hash())
		assert.Nil(t, err)
		// assert.Equal(t, len(events), 1)
		// events.
		fmt.Printf("==events:%v\n", events)
		for _, event := range events {
			fmt.Printf("topic:%v--event:%v\n", event.Topic, event.Data)
			if event.Topic == "chain.transactionResult" {
				var jEvent SysEvent
				if err := json.Unmarshal([]byte(event.Data), &jEvent); err == nil {
					assert.Equal(t, tt.contractExpectedErr, jEvent.Err)
					assert.Equal(t, tt.contractExpectedResult, jEvent.Result)
				}
			} else {
				var jEvent InnerEvent
				if err := json.Unmarshal([]byte(event.Data), &jEvent); err == nil {
					assert.Equal(t, tt.innerExpectedErr, jEvent.Error)
				}
			}

		}

		//
	}
}

// optimized
func TestInnerTransactionsGasLimit(t *testing.T) {
	core.NebCompatibility = core.NewCompatibilityLocal()
	tests := []struct {
		name              string
		contracts         []contract
		call              call
		expectedErr       string
		gasArr            []int
		gasExpectedErr    []string
		gasExpectedResult []string
	}{
		{
			"deploy test_require_module.js",
			[]contract{
				contract{
					"./test/test_inner_transaction.js",
					"js",
					"",
				},
				contract{
					"./test/bank_vault_contract_second.js",
					"js",
					"",
				},
				contract{
					"./test/bank_vault_contract.js",
					"js",
					"",
				},
			},
			call{
				"save",
				"[1]",
				[]string{""},
			},
			"multi execution failed",
			//96093 “”	"\"\""
			//95269	insufficient gas "null"	B+C执行完毕，代码回到A执行失败
			//95105	insufficient gas "null"
			//95092	insufficient gas	"\"\""	C刚好最后一句代码gas limit
			//94710	insufficient gas "null"
			//94697	insufficient gas "null"
			//57249	insufficient gas "null"
			//57248 insufficient gas "null"
			//53000	insufficient gas "null"
			[]int{20175, 20174, 53000, 57248, 57249, 94697, 94710, 95092, 95105, 95269, 96093},
			// []int{53000},
			//95093->95105, 94698->94710
			//96093 “”	"\"\""
			//95269	insufficient gas "null"	B+C执行完毕，代码回到A执行失败
			//95105 c刚好消费殆尽,代码回到B后gas不足. Call: inner transation err [insufficient gas] engine index:0
			//95092	Inner Call: inner transation err [insufficient gas] engine index:1
			//94710 Inner Call: inner transation err [insufficient gas] engine index:1","execute_result":"inner transation err [insufficient gas] engine index:1"
			//94697 调用C的时候B消耗完毕	Inner Call: inner transation err [preparation inner nvm insufficient gas] engine index:1
			//57249 Inner Call: inner transation err [insufficient gas] engine index:0
			//57248 调用B的时候,A消耗完毕 Inner Call: inner transation err [preparation inner nvm insufficient gas] engine index:0
			//53000 Inner Call: inner transation err [preparation inner nvm insufficient gas] engine index:0
			//20174	out of gas limit 	""
			//20175 insufficient gas "null"
			//20000
			[]string{"insufficient gas", "out of gas limit",
				"insufficient gas",
				"insufficient gas",
				"insufficient gas",
				"insufficient gas",
				"insufficient gas",
				"insufficient gas",
				"insufficient gas",
				"insufficient gas",
				"",
			},
			[]string{"null", "", "Inner Contract: null",
				"Inner Contract: null",
				"Inner Contract: null",
				"Inner Contract: null",
				"Inner Contract: null",
				"Inner Contract: \"\"",
				"Inner Contract: null",
				"null",
				"\"\""},
		},
	}

	for _, tt := range tests {

		neb := mockNeb(t)
		tail := neb.chain.TailBlock()
		manager, err := account.NewManager(neb)
		assert.Nil(t, err)

		a, _ := core.AddressParse("n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE")
		assert.Nil(t, manager.Unlock(a, []byte("passphrase"), keystore.YearUnlockDuration))
		b, _ := core.AddressParse("n1GmkKH6nBMw4rrjt16RrJ9WcgvKUtAZP1s")
		assert.Nil(t, manager.Unlock(b, []byte("passphrase"), keystore.YearUnlockDuration))
		c, _ := core.AddressParse("n1H4MYms9F55ehcvygwWE71J8tJC4CRr2so")
		assert.Nil(t, manager.Unlock(c, []byte("passphrase"), keystore.YearUnlockDuration))
		d, _ := core.AddressParse("n1JAy4X6KKLCNiTd7MWMRsVBjgdVq5WCCpf")
		assert.Nil(t, manager.Unlock(d, []byte("passphrase"), keystore.YearUnlockDuration))

		elapsedSecond := dpos.BlockIntervalInMs / dpos.SecondInMs
		consensusState, err := tail.WorldState().NextConsensusState(elapsedSecond)
		assert.Nil(t, err)

		// mock empty block(height=2)
		block, err := core.MockBlockEx(neb.chain.ChainID(), c, tail, 2)
		fmt.Printf("mock 2, block.height:%v\n", block.Height())
		assert.Nil(t, err)
		/* ----- mock random seed for new block ------*/
		miner, err := core.AddressParseFromBytes(consensusState.Proposer())
		assert.Nil(t, err)
		seed, proof, err := manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), neb.chain.GenesisBlock().Hash()) // NOTE: 3rd arg is genesis's hash for the first block
		assert.Nil(t, err)
		block.SetRandomSeed(seed, proof)
		/* ----- mock random seed for new block END ------*/
		block.WorldState().SetConsensusState(consensusState)
		block.SetTimestamp(consensusState.TimeStamp())
		block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
		assert.Nil(t, block.Seal())
		assert.Nil(t, manager.SignBlock(b, block))
		assert.Nil(t, neb.chain.BlockPool().Push(block))
		fmt.Printf("mock 2, block.tailblock.height: %v\n", neb.chain.TailBlock().Height())

		// inner call block(height=3)
		tail = neb.chain.TailBlock()
		block, err = core.MockBlockEx(neb.chain.ChainID(), c, tail, 3)
		assert.Nil(t, err)
		consensusState, err = tail.WorldState().NextConsensusState(elapsedSecond)
		assert.Nil(t, err)
		/* ----- mock random seed for new block ------*/
		miner, err = core.AddressParseFromBytes(consensusState.Proposer())
		assert.Nil(t, err)
		seed, proof, err = manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), seed) // NOTE: 3rd arg is parent's seed
		assert.Nil(t, err)
		block.SetRandomSeed(seed, proof)
		/* ----- mock random seed for new block END ------*/
		block.WorldState().SetConsensusState(consensusState)
		block.SetTimestamp(consensusState.TimeStamp())

		contractsAddr := []string{}
		fmt.Printf("++++++++++++pack account")
		// t.Run(tt.name, func(t *testing.T) {
		for k, v := range tt.contracts {
			data, err := ioutil.ReadFile(v.contractPath)
			assert.Nil(t, err, "contract path read error")
			source := string(data)
			sourceType := "js"
			argsDeploy := ""
			deploy, _ := core.NewDeployPayload(source, sourceType, argsDeploy)
			payloadDeploy, _ := deploy.ToBytes()

			value, _ := util.NewUint128FromInt(0)
			gasLimit, _ := util.NewUint128FromInt(200000)
			txDeploy, err := core.NewTransaction(neb.chain.ChainID(), a, a, value, uint64(k+1), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
			assert.Nil(t, err)
			assert.Nil(t, manager.SignTransaction(a, txDeploy))
			assert.Nil(t, neb.chain.TransactionPool().Push(txDeploy))

			contractAddr, err := txDeploy.GenerateContractAddress()
			assert.Nil(t, err)
			contractsAddr = append(contractsAddr, contractAddr.String())
		}
		// })

		block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
		assert.Nil(t, block.Seal())
		assert.Nil(t, manager.SignBlock(c, block))
		assert.Nil(t, neb.chain.BlockPool().Push(block))

		for _, v := range contractsAddr {
			contract, err := core.AddressParse(v)
			assert.Nil(t, err)
			_, err = neb.chain.TailBlock().CheckContract(contract)
			assert.Nil(t, err)
		}

		proxyContractAddress, err := core.AddressParse(contractsAddr[0])
		calleeContract := contractsAddr[1]
		callToContract := contractsAddr[2]

		// call tx

		pSeed := seed
		pTail := neb.chain.TailBlock()

		fmt.Println("++++++++++++pack payload")
		callPayload, _ := core.NewCallPayload(tt.call.function, fmt.Sprintf("[\"%s\", \"%s\", 1]", calleeContract, callToContract))
		payloadCall, _ := callPayload.ToBytes()

		consensusState, err = pTail.WorldState().NextConsensusState(elapsedSecond)
		assert.Nil(t, err)
		/* ----- mock random seed for new block ------*/
		miner, err = core.AddressParseFromBytes(consensusState.Proposer())
		assert.Nil(t, err)
		tempSeed, tempProof, err := manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), pSeed) // NOTE: 3rd arg is parent's seed
		assert.Nil(t, err)
		/* ----- mock random seed for new block END ------*/

		for i := 0; i < len(tt.gasArr); i++ {

			consensusState, err = pTail.WorldState().NextConsensusState(elapsedSecond)
			assert.Nil(t, err)

			block, err = core.NewBlock(neb.chain.ChainID(), b, pTail)
			assert.Nil(t, err)
			block.SetRandomSeed(tempSeed, tempProof)
			block.WorldState().SetConsensusState(consensusState)
			block.SetTimestamp(consensusState.TimeStamp())

			value, _ := util.NewUint128FromInt(6)
			//gasLimit, _ := util.NewUint128FromInt(21300)
			//gasLimit, _ := util.NewUint128FromInt(25300)	//null                            file=logger.go func=nvm.V8Log line=32
			gasLimit, _ := util.NewUint128FromInt(int64(tt.gasArr[i]))
			fmt.Println("++++++++++++pack transaction")
			txCall, err := core.NewTransaction(neb.chain.ChainID(), a, proxyContractAddress, value,
				uint64(len(contractsAddr)+1), core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
			assert.Nil(t, err)
			assert.Nil(t, manager.SignTransaction(a, txCall))
			assert.Nil(t, neb.chain.TransactionPool().Push(txCall))

			fmt.Println("++++++++++++pack collect")
			block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
			// assert.Nil(t, block.Seal())
			// assert.Nil(t, manager.SignBlock(d, block))
			// neb.chain.RemoveCacheBlock(block.Hash())
			// err = neb.chain.Storage().Del(block.Hash())
			assert.Nil(t, err)
			// assert.Nil(t, neb.chain.BlockPool().Push(block))

			events, err := block.WorldState().FetchEvents(txCall.Hash())
			assert.Nil(t, err)
			// events.
			// fmt.Println("==========>Txs")
			for _, tx := range block.Transactions() {
				// fmt.Println("=========>Tx", tx.Hash(), tx.GasLimit(), txCall.Hash())
				assert.Equal(t, tx.Hash(), txCall.Hash())
			}
			fmt.Println("==events:", events)
			for _, event := range events {

				fmt.Println("==============", event.Data)
				var jEvent SysEvent
				if err := json.Unmarshal([]byte(event.Data), &jEvent); err == nil {
					if jEvent.Hash != "" {
						assert.Equal(t, tt.gasExpectedErr[i], jEvent.Err)
						assert.Equal(t, tt.gasExpectedResult[i], jEvent.Result)
					}
				}
			}

		}

		/*
			contractOne, err := core.AddressParse(contractsAddr[0])
			accountANew, err := tail.GetAccount(contractOne.Bytes())
			assert.Nil(t, err)
			fmt.Printf("contractA account :%v\n", accountANew)

			contractTwo, err := core.AddressParse(contractsAddr[1])
			accountBNew, err := tail.GetAccount(contractTwo.Bytes())
			assert.Nil(t, err)
			fmt.Printf("contractB account :%v\n", accountBNew)

			aI, err := tail.GetAccount(a.Bytes())
			// bI, err := tail.GetAccount(b.Bytes())
			fmt.Printf("aI:%v\n", aI)
			bI, err := tail.GetAccount(b.Bytes())
			fmt.Printf("bI:%v\n", bI)*/
		//
	}
}

type SysEvent struct {
	Hash    string `json:"hash"`
	Status  int    `json:"status"`
	GasUsed string `json:"gas_used"`
	Err     string `json:"error"`
	Result  string `json:"execute_result"`
}
type InnerEvent struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Value int    `json:"value"`
	Error string `json:"error"`
}

func TestInnerTransactionsMemLimit(t *testing.T) {
	core.NebCompatibility = core.NewCompatibilityLocal()
	tests := []struct {
		name              string
		contracts         []contract
		call              call
		expectedErr       string
		memArr            []int
		memExpectedErr    []string
		memExpectedResult []string
	}{
		{
			"deploy test_require_module.js",
			[]contract{
				contract{
					"./test/test_inner_transaction.js",
					"js",
					"",
				},
				contract{
					"./test/bank_vault_contract_second.js",
					"js",
					"",
				},
				contract{
					"./test/bank_vault_contract.js",
					"js",
					"",
				},
			},
			call{
				"saveMem",
				"[1]",
				[]string{""},
			},
			"multi execution failed",
			[]int{5 * 1024 * 1024, 10 * 1024 * 1024, 20 * 1024 * 1024, 40 * 1024 * 1024},
			// []int{40 * 1024 * 1024},
			[]string{"",
				"exceed memory limits",
				"exceed memory limits",
				"exceed memory limits"},
			[]string{"\"\"",
				"Inner Contract: null", "Inner Contract: null", "null",
			},
		},
	}

	for _, tt := range tests {
		for i := 0; i < len(tt.memArr); i++ {

			neb := mockNeb(t)
			tail := neb.chain.TailBlock()

			manager, err := account.NewManager(neb)
			assert.Nil(t, err)

			a, _ := core.AddressParse("n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE")
			assert.Nil(t, manager.Unlock(a, []byte("passphrase"), keystore.YearUnlockDuration))
			b, _ := core.AddressParse("n1GmkKH6nBMw4rrjt16RrJ9WcgvKUtAZP1s")
			assert.Nil(t, manager.Unlock(b, []byte("passphrase"), keystore.YearUnlockDuration))
			c, _ := core.AddressParse("n1H4MYms9F55ehcvygwWE71J8tJC4CRr2so")
			assert.Nil(t, manager.Unlock(c, []byte("passphrase"), keystore.YearUnlockDuration))
			d, _ := core.AddressParse("n1JAy4X6KKLCNiTd7MWMRsVBjgdVq5WCCpf")
			assert.Nil(t, manager.Unlock(d, []byte("passphrase"), keystore.YearUnlockDuration))

			elapsedSecond := dpos.BlockIntervalInMs / dpos.SecondInMs
			consensusState, err := tail.WorldState().NextConsensusState(elapsedSecond)
			assert.Nil(t, err)
			// block, err := core.NewBlock(neb.chain.ChainID(), b, tail)
			fmt.Printf("tail.height:%v\n", tail.Height())

			// mock empty block(height=2)
			block, err := core.MockBlockEx(neb.chain.ChainID(), c, tail, 2)
			fmt.Printf("mock 2, block.height:%v\n", block.Height())
			assert.Nil(t, err)
			/* ----- mock random seed for new block ------*/
			miner, err := core.AddressParseFromBytes(consensusState.Proposer())
			assert.Nil(t, err)
			seed, proof, err := manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), neb.chain.GenesisBlock().Hash()) // NOTE: 3rd arg is genesis's hash for the first block
			assert.Nil(t, err)
			block.SetRandomSeed(seed, proof)
			/* ----- mock random seed for new block END ------*/
			block.WorldState().SetConsensusState(consensusState)
			block.SetTimestamp(consensusState.TimeStamp())
			block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
			assert.Nil(t, block.Seal())
			assert.Nil(t, manager.SignBlock(b, block))
			assert.Nil(t, neb.chain.BlockPool().Push(block))
			fmt.Printf("mock 2, block.tailblock.height: %v\n", neb.chain.TailBlock().Height())

			// inner call block(height=3)
			tail = neb.chain.TailBlock()
			block, err = core.MockBlockEx(neb.chain.ChainID(), c, tail, 3)
			assert.Nil(t, err)
			consensusState, err = tail.WorldState().NextConsensusState(elapsedSecond)
			assert.Nil(t, err)

			/* ----- mock random seed for new block ------*/
			miner, err = core.AddressParseFromBytes(consensusState.Proposer())
			assert.Nil(t, err)
			seed, proof, err = manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), seed) // NOTE: 3rd arg is parent's seed
			assert.Nil(t, err)
			block.SetRandomSeed(seed, proof)
			/* ----- mock random seed for new block END ------*/
			block.WorldState().SetConsensusState(consensusState)
			block.SetTimestamp(consensusState.TimeStamp())

			contractsAddr := []string{}
			for k, v := range tt.contracts {
				data, err := ioutil.ReadFile(v.contractPath)
				assert.Nil(t, err, "contract path read error")
				source := string(data)
				sourceType := "js"
				argsDeploy := ""
				deploy, _ := core.NewDeployPayload(source, sourceType, argsDeploy)
				payloadDeploy, _ := deploy.ToBytes()

				value, _ := util.NewUint128FromInt(0)
				gasLimit, _ := util.NewUint128FromInt(200000)
				txDeploy, err := core.NewTransaction(neb.chain.ChainID(), a, a, value, uint64(k+1), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
				assert.Nil(t, err)
				assert.Nil(t, manager.SignTransaction(a, txDeploy))
				assert.Nil(t, neb.chain.TransactionPool().Push(txDeploy))

				contractAddr, err := txDeploy.GenerateContractAddress()
				assert.Nil(t, err)
				contractsAddr = append(contractsAddr, contractAddr.String())
			}

			block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
			assert.Nil(t, block.Seal())
			assert.Nil(t, manager.SignBlock(c, block))
			assert.Nil(t, neb.chain.BlockPool().Push(block))

			for _, v := range contractsAddr {
				contract, err := core.AddressParse(v)
				assert.Nil(t, err)
				_, err = neb.chain.TailBlock().CheckContract(contract)
				assert.Nil(t, err)
			}

			fmt.Printf("mock 3, block.height:%v, tail: %v\n", block.Height(), neb.chain.TailBlock().Height())

			elapsedSecond = dpos.BlockIntervalInMs / dpos.SecondInMs
			tail = neb.chain.TailBlock()
			consensusState, err = tail.WorldState().NextConsensusState(elapsedSecond)
			assert.Nil(t, err)
			block, err = core.NewBlock(neb.chain.ChainID(), b, tail)
			fmt.Printf("mock 4, block.height:%v, tail: %v\n", block.Height(), neb.chain.TailBlock().Height())
			// block, err = core.MockBlockEx(neb.chain.ChainID(), c, tail, 2)
			// block = core.MockBlock(nil, 3)
			assert.Nil(t, err)
			/* ----- mock random seed for new block ------*/
			miner, err = core.AddressParseFromBytes(consensusState.Proposer())
			assert.Nil(t, err)
			seed, proof, err = manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), seed) // NOTE: 3rd arg is parent's seed
			assert.Nil(t, err)
			block.SetRandomSeed(seed, proof)
			/* ----- mock random seed for new block END ------*/
			block.WorldState().SetConsensusState(consensusState)
			block.SetTimestamp(consensusState.TimeStamp())
			assert.Nil(t, err)

			calleeContract := contractsAddr[1]
			callToContract := contractsAddr[2]
			callPayload, _ := core.NewCallPayload(tt.call.function, fmt.Sprintf("[\"%s\", \"%s\", \"%d\"]", calleeContract, callToContract, tt.memArr[i]))
			payloadCall, _ := callPayload.ToBytes()

			value, _ := util.NewUint128FromInt(6)
			gasLimit, _ := util.NewUint128FromInt(5000000)
			proxyContractAddress, err := core.AddressParse(contractsAddr[0])
			txCall, err := core.NewTransaction(neb.chain.ChainID(), a, proxyContractAddress, value,
				uint64(len(contractsAddr)+1), core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
			assert.Nil(t, err)
			assert.Nil(t, manager.SignTransaction(a, txCall))
			assert.Nil(t, neb.chain.TransactionPool().Push(txCall))

			block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
			assert.Nil(t, block.Seal())
			assert.Nil(t, manager.SignBlock(d, block))
			assert.Nil(t, neb.chain.BlockPool().Push(block))

			tail = neb.chain.TailBlock()
			events, err := tail.FetchEvents(txCall.Hash())
			for _, event := range events {
				fmt.Printf("mem err:%v", event.Data)
				var jEvent SysEvent
				if err := json.Unmarshal([]byte(event.Data), &jEvent); err == nil {
					if jEvent.Hash != "" {
						assert.Equal(t, tt.memExpectedErr[i], jEvent.Err)
						assert.Equal(t, tt.memExpectedResult[i], jEvent.Result)
					}
				}

			}
		}
	}
}

func TestInnerTransactionsErr(t *testing.T) {
	core.NebCompatibility = core.NewCompatibilityLocal()
	tests := []struct {
		name            string
		contracts       []contract
		call            call
		errFlagArr      []uint32
		expectedErrArr  []string
		expectedAccount [][]string
	}{
		{
			"deploy TestInnerTransactionsErr.js",
			[]contract{
				contract{
					"./test/test_inner_transaction.js",
					"js",
					"",
				},
				contract{
					"./test/bank_vault_contract_second.js",
					"js",
					"",
				},
				contract{
					"./test/bank_vault_contract.js",
					"js",
					"",
				},
			},
			call{
				"saveErr",
				"[1]",
				[]string{""},
			},
			[]uint32{0, 1, 2, 3},
			// []uint32{0},
			[]string{"Call: saveErr in test_inner_transaction",
				"Call: Inner Contract: saveErr in bank_vault_contract_second",
				"Call: Inner Contract: saveErr in bank_vault_contract",
				"",
			},
			[][]string{
				{"0", "0", "0", "4999999999999903290000000", "5000000000000000000000000", "5000004280820096710000000", "5000000000000000000000000"},
				{"0", "0", "0", "4999999999999871253000000", "5000000000000000000000000", "5000004280820128747000000", "5000000000000000000000000"},
				{"0", "0", "0", "4999999999999871253000000", "5000000000000000000000000", "5000004280820128747000000", "5000000000000000000000000"},
				{"1", "2", "3", "4999999999999871253000000", "5000000000000000000000000", "5000004280820128747000000", "5000000000000000000000000"},
			},
		},
	}

	for _, tt := range tests {
		for i := 0; i < len(tt.errFlagArr); i++ {

			neb := mockNeb(t)
			tail := neb.chain.TailBlock()
			manager, err := account.NewManager(neb)
			assert.Nil(t, err)

			a, _ := core.AddressParse("n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE")
			assert.Nil(t, manager.Unlock(a, []byte("passphrase"), keystore.YearUnlockDuration))
			b, _ := core.AddressParse("n1GmkKH6nBMw4rrjt16RrJ9WcgvKUtAZP1s")
			assert.Nil(t, manager.Unlock(b, []byte("passphrase"), keystore.YearUnlockDuration))
			c, _ := core.AddressParse("n1H4MYms9F55ehcvygwWE71J8tJC4CRr2so")
			assert.Nil(t, manager.Unlock(c, []byte("passphrase"), keystore.YearUnlockDuration))
			d, _ := core.AddressParse("n1JAy4X6KKLCNiTd7MWMRsVBjgdVq5WCCpf")
			assert.Nil(t, manager.Unlock(d, []byte("passphrase"), keystore.YearUnlockDuration))

			elapsedSecond := dpos.BlockIntervalInMs / dpos.SecondInMs
			consensusState, err := tail.WorldState().NextConsensusState(elapsedSecond)
			assert.Nil(t, err)

			// mock empty block(height=2)
			block, err := core.MockBlockEx(neb.chain.ChainID(), c, tail, 2)
			fmt.Printf("mock 2, block.height:%v\n", block.Height())
			assert.Nil(t, err)
			/* ----- mock random seed for new block ------*/
			miner, err := core.AddressParseFromBytes(consensusState.Proposer())
			assert.Nil(t, err)
			seed, proof, err := manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), neb.chain.GenesisBlock().Hash()) // NOTE: 3rd arg is genesis's hash for the first block
			assert.Nil(t, err)
			block.SetRandomSeed(seed, proof)
			/* ----- mock random seed for new block END ------*/
			block.WorldState().SetConsensusState(consensusState)
			block.SetTimestamp(consensusState.TimeStamp())
			block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
			assert.Nil(t, block.Seal())
			assert.Nil(t, manager.SignBlock(b, block))
			assert.Nil(t, neb.chain.BlockPool().Push(block))
			fmt.Printf("mock 2, block.tailblock.height: %v\n", neb.chain.TailBlock().Height())

			// inner call block(height=3)
			tail = neb.chain.TailBlock()
			block, err = core.MockBlockEx(neb.chain.ChainID(), c, tail, 3)
			assert.Nil(t, err)
			consensusState, err = tail.WorldState().NextConsensusState(elapsedSecond)
			assert.Nil(t, err)
			/* ----- mock random seed for new block ------*/
			miner, err = core.AddressParseFromBytes(consensusState.Proposer())
			assert.Nil(t, err)
			seed, proof, err = manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), seed) // NOTE: 3rd arg is parent's seed
			assert.Nil(t, err)
			block.SetRandomSeed(seed, proof)
			/* ----- mock random seed for new block END ------*/

			block.WorldState().SetConsensusState(consensusState)
			block.SetTimestamp(consensusState.TimeStamp())

			contractsAddr := []string{}
			for k, v := range tt.contracts {
				data, err := ioutil.ReadFile(v.contractPath)
				assert.Nil(t, err, "contract path read error")
				source := string(data)
				sourceType := "js"
				argsDeploy := ""
				deploy, _ := core.NewDeployPayload(source, sourceType, argsDeploy)
				payloadDeploy, _ := deploy.ToBytes()

				value, _ := util.NewUint128FromInt(0)
				gasLimit, _ := util.NewUint128FromInt(200000)
				txDeploy, err := core.NewTransaction(neb.chain.ChainID(), a, a, value, uint64(k+1), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
				assert.Nil(t, err)
				assert.Nil(t, manager.SignTransaction(a, txDeploy))
				assert.Nil(t, neb.chain.TransactionPool().Push(txDeploy))

				contractAddr, err := txDeploy.GenerateContractAddress()
				assert.Nil(t, err)
				contractsAddr = append(contractsAddr, contractAddr.String())
			}

			block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
			assert.Nil(t, block.Seal())
			assert.Nil(t, manager.SignBlock(c, block))
			assert.Nil(t, neb.chain.BlockPool().Push(block))

			bUser, err := tail.GetAccount(b.Bytes())
			fmt.Printf("bbbb:%v", bUser.Balance().String())

			for _, v := range contractsAddr {
				contract, err := core.AddressParse(v)
				assert.Nil(t, err)
				_, err = neb.chain.TailBlock().CheckContract(contract)
				assert.Nil(t, err)
			}

			elapsedSecond = dpos.BlockIntervalInMs / dpos.SecondInMs
			tail = neb.chain.TailBlock()
			consensusState, err = tail.WorldState().NextConsensusState(elapsedSecond)
			assert.Nil(t, err)
			block, err = core.NewBlock(neb.chain.ChainID(), c, tail)
			assert.Nil(t, err)
			/* ----- mock random seed for new block ------*/
			miner, err = core.AddressParseFromBytes(consensusState.Proposer())
			assert.Nil(t, err)
			seed, proof, err = manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), seed) // NOTE: 3rd arg is parent's seed
			assert.Nil(t, err)
			block.SetRandomSeed(seed, proof)
			/* ----- mock random seed for new block END ------*/
			block.WorldState().SetConsensusState(consensusState)
			block.SetTimestamp(consensusState.TimeStamp())
			assert.Nil(t, err)

			calleeContract := contractsAddr[1]
			callToContract := contractsAddr[2]
			callPayload, _ := core.NewCallPayload(tt.call.function, fmt.Sprintf("[\"%s\", \"%s\", \"%d\"]", calleeContract, callToContract, tt.errFlagArr[i]))
			payloadCall, _ := callPayload.ToBytes()

			value, _ := util.NewUint128FromInt(6)
			gasLimit, _ := util.NewUint128FromInt(1000000)
			proxyContractAddress, err := core.AddressParse(contractsAddr[0])
			txCall, err := core.NewTransaction(neb.chain.ChainID(), a, proxyContractAddress, value,
				uint64(len(contractsAddr)+1), core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
			assert.Nil(t, err)
			assert.Nil(t, manager.SignTransaction(a, txCall))
			assert.Nil(t, neb.chain.TransactionPool().Push(txCall))

			block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
			assert.Nil(t, block.Seal())
			assert.Nil(t, manager.SignBlock(d, block))
			assert.Nil(t, neb.chain.BlockPool().Push(block))

			tail = neb.chain.TailBlock()
			events, err := tail.FetchEvents(txCall.Hash())
			for _, event := range events {
				fmt.Printf("event:%v\n", event.Data)
				var jEvent SysEvent
				if err := json.Unmarshal([]byte(event.Data), &jEvent); err == nil {
					if jEvent.Hash != "" {
						assert.Equal(t, tt.expectedErrArr[i], jEvent.Err)
					}
				}

			}
			//chech accout
			contractAddrA, err := core.AddressParse(contractsAddr[0])
			accountAAcc, err := tail.GetAccount(contractAddrA.Bytes())
			assert.Nil(t, err)
			// fmt.Printf("account :%v\n", accountAAcc)
			assert.Equal(t, tt.expectedAccount[i][0], accountAAcc.Balance().String())

			contractAddrB, err := core.AddressParse(contractsAddr[1])
			accountBAcc, err := tail.GetAccount(contractAddrB.Bytes())
			assert.Nil(t, err)
			// fmt.Printf("accountB :%v\n", accountBAcc)
			assert.Equal(t, tt.expectedAccount[i][1], accountBAcc.Balance().String())

			contractAddrC, err := core.AddressParse(contractsAddr[2])
			accountAccC, err := tail.GetAccount(contractAddrC.Bytes())
			assert.Nil(t, err)
			fmt.Printf("accountC :%v\n", accountAccC)
			assert.Equal(t, tt.expectedAccount[i][2], accountAccC.Balance().String())

			aUser, err := tail.GetAccount(a.Bytes())
			// assert.Equal(t, tt.expectedAccount[i][3], aUser.Balance().String())
			fmt.Printf("aI:%v\n", aUser)
			bUser, err = tail.GetAccount(b.Bytes())
			assert.Equal(t, tt.expectedAccount[i][4], bUser.Balance().String())

			cUser, err := tail.GetAccount(c.Bytes())
			fmt.Printf("cI:%v\n", cUser)
			// assert.Equal(t, tt.expectedAccount[i][5], cUser.Balance().String())

			// fmt.Printf("b:%v\n", bUser)
			// assert.Equal(t, tt.expectedAccount[i][4], cUser.Balance().String())
			dUser, err := tail.GetAccount(d.Bytes())
			assert.Equal(t, tt.expectedAccount[i][6], dUser.Balance().String())
			// fmt.Printf("d:%v\n", dUser)
			// assert.Equal(t, tt.expectedAccount[i][4], dUser.Balance().String())
		}
	}
}

func TestInnerTransactionsValue(t *testing.T) {
	core.NebCompatibility = core.NewCompatibilityLocal()
	tests := []struct {
		name            string
		contracts       []contract
		call            call
		errValueArr     []string
		expectedErrArr  []string
		expectedAccount [][]string
	}{
		{
			"deploy TestInnerTransactionsValue",
			[]contract{
				contract{
					"./test/test_inner_transaction.js",
					"js",
					"",
				},
				contract{
					"./test/bank_vault_contract_second.js",
					"js",
					"",
				},
				contract{
					"./test/bank_vault_contract.js",
					"js",
					"",
				},
			},
			call{
				"saveValue",
				"[1]",
				[]string{""},
			},
			[]string{"-1", "340282366920938463463374607431768211455", "340282366920938463463374607431768211456",
				"1.1", "NaN", "undefined", "null", "infinity", "", "\"\""},
			// []string{"0"},
			[]string{"Call: Inner Call: invalid value",
				"Call: Inner Contract: inner transfer failed",
				"Call: Inner Contract: inner transfer failed",
				"Call: Inner Call: invalid value",
				"Call: Inner Call: invalid value",
				"Call: BigNumber Error: new BigNumber() not a number: undefined",
				"Call: BigNumber Error: new BigNumber() not a number: null",
				"Call: BigNumber Error: new BigNumber() not a number: infinity",
				"",
				"invalid function of call payload",
			},
			[][]string{
				{"0", "0", "0", "4999999999999903290000000", "5000000000000000000000000", "5000004280820096710000000", "5000000000000000000000000"},
				{"0", "0", "0", "4999999999999871253000000", "5000000000000000000000000", "5000004280820128747000000", "5000000000000000000000000"},
				{"0", "0", "0", "4999999999999871253000000", "5000000000000000000000000", "5000004280820128747000000", "5000000000000000000000000"},
				{"0", "0", "0", "4999999999999871253000000", "5000000000000000000000000", "5000004280820128747000000", "5000000000000000000000000"},
				{"0", "0", "0", "4999999999999871253000000", "5000000000000000000000000", "5000004280820128747000000", "5000000000000000000000000"},
				{"0", "0", "0", "4999999999999871253000000", "5000000000000000000000000", "5000004280820128747000000", "5000000000000000000000000"},
				{"0", "0", "0", "4999999999999871253000000", "5000000000000000000000000", "5000004280820128747000000", "5000000000000000000000000"},
				{"0", "0", "0", "4999999999999871253000000", "5000000000000000000000000", "5000004280820128747000000", "5000000000000000000000000"},
				{"6", "0", "0", "4999999999999871253000000", "5000000000000000000000000", "5000004280820128747000000", "5000000000000000000000000"},
				{"0", "0", "0", "4999999999999871253000000", "5000000000000000000000000", "5000004280820128747000000", "5000000000000000000000000"},
			},
		},
	}
	for _, tt := range tests {
		for i := 0; i < len(tt.errValueArr); i++ {

			neb := mockNeb(t)
			tail := neb.chain.TailBlock()
			manager, err := account.NewManager(neb)
			assert.Nil(t, err)

			a, _ := core.AddressParse("n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE")
			assert.Nil(t, manager.Unlock(a, []byte("passphrase"), keystore.YearUnlockDuration))
			b, _ := core.AddressParse("n1GmkKH6nBMw4rrjt16RrJ9WcgvKUtAZP1s")
			assert.Nil(t, manager.Unlock(b, []byte("passphrase"), keystore.YearUnlockDuration))
			c, _ := core.AddressParse("n1H4MYms9F55ehcvygwWE71J8tJC4CRr2so")
			assert.Nil(t, manager.Unlock(c, []byte("passphrase"), keystore.YearUnlockDuration))
			d, _ := core.AddressParse("n1JAy4X6KKLCNiTd7MWMRsVBjgdVq5WCCpf")
			assert.Nil(t, manager.Unlock(d, []byte("passphrase"), keystore.YearUnlockDuration))

			elapsedSecond := dpos.BlockIntervalInMs / dpos.SecondInMs
			consensusState, err := tail.WorldState().NextConsensusState(elapsedSecond)
			assert.Nil(t, err)

			// mock empty block(height=2)
			block, err := core.MockBlockEx(neb.chain.ChainID(), c, tail, 2)
			fmt.Printf("mock 2, block.height:%v\n", block.Height())
			assert.Nil(t, err)
			/* ----- mock random seed for new block ------*/
			miner, err := core.AddressParseFromBytes(consensusState.Proposer())
			assert.Nil(t, err)
			seed, proof, err := manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), neb.chain.GenesisBlock().Hash()) // NOTE: 3rd arg is genesis's hash for the first block
			assert.Nil(t, err)
			block.SetRandomSeed(seed, proof)
			/* ----- mock random seed for new block END ------*/
			block.WorldState().SetConsensusState(consensusState)
			block.SetTimestamp(consensusState.TimeStamp())
			block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
			assert.Nil(t, block.Seal())
			assert.Nil(t, manager.SignBlock(b, block))
			assert.Nil(t, neb.chain.BlockPool().Push(block))
			fmt.Printf("mock 2, block.tailblock.height: %v\n", neb.chain.TailBlock().Height())

			// inner call block(height=3)
			tail = neb.chain.TailBlock()
			block, err = core.MockBlockEx(neb.chain.ChainID(), c, tail, 3)
			assert.Nil(t, err)
			consensusState, err = tail.WorldState().NextConsensusState(elapsedSecond)
			assert.Nil(t, err)
			/* ----- mock random seed for new block ------*/
			miner, err = core.AddressParseFromBytes(consensusState.Proposer())
			assert.Nil(t, err)
			seed, proof, err = manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), seed) // NOTE: 3rd arg is parent's seed
			assert.Nil(t, err)
			block.SetRandomSeed(seed, proof)
			/* ----- mock random seed for new block END ------*/

			block.WorldState().SetConsensusState(consensusState)
			block.SetTimestamp(consensusState.TimeStamp())

			contractsAddr := []string{}
			for k, v := range tt.contracts {
				data, err := ioutil.ReadFile(v.contractPath)
				assert.Nil(t, err, "contract path read error")
				source := string(data)
				sourceType := "js"
				argsDeploy := ""
				deploy, _ := core.NewDeployPayload(source, sourceType, argsDeploy)
				payloadDeploy, _ := deploy.ToBytes()

				value, _ := util.NewUint128FromInt(0)
				gasLimit, _ := util.NewUint128FromInt(200000)
				txDeploy, err := core.NewTransaction(neb.chain.ChainID(), a, a, value, uint64(k+1), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
				assert.Nil(t, err)
				assert.Nil(t, manager.SignTransaction(a, txDeploy))
				assert.Nil(t, neb.chain.TransactionPool().Push(txDeploy))

				contractAddr, err := txDeploy.GenerateContractAddress()
				assert.Nil(t, err)
				contractsAddr = append(contractsAddr, contractAddr.String())
			}

			block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
			assert.Nil(t, block.Seal())
			assert.Nil(t, manager.SignBlock(c, block))
			assert.Nil(t, neb.chain.BlockPool().Push(block))

			for _, v := range contractsAddr {
				contract, err := core.AddressParse(v)
				assert.Nil(t, err)
				_, err = neb.chain.TailBlock().CheckContract(contract)
				assert.Nil(t, err)
			}

			elapsedSecond = dpos.BlockIntervalInMs / dpos.SecondInMs
			tail = neb.chain.TailBlock()
			consensusState, err = tail.WorldState().NextConsensusState(elapsedSecond)
			assert.Nil(t, err)
			block, err = core.NewBlock(neb.chain.ChainID(), c, tail)
			assert.Nil(t, err)
			/* ----- mock random seed for new block ------*/
			miner, err = core.AddressParseFromBytes(consensusState.Proposer())
			assert.Nil(t, err)
			seed, proof, err = manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), seed) // NOTE: 3rd arg is parent's seed
			assert.Nil(t, err)
			block.SetRandomSeed(seed, proof)
			/* ----- mock random seed for new block END ------*/
			block.WorldState().SetConsensusState(consensusState)
			block.SetTimestamp(consensusState.TimeStamp())
			assert.Nil(t, err)

			calleeContract := contractsAddr[1]
			aa, err := core.AddressParse(contractsAddr[1])
			aaa, err := tail.GetAccount(aa.Bytes())
			fmt.Printf("aaaaaaa:%v", aaa.Address().String())

			callToContract := contractsAddr[2]
			callPayload, _ := core.NewCallPayload(tt.call.function, fmt.Sprintf("[\"%s\", \"%s\", \"%s\"]", calleeContract, callToContract, tt.errValueArr[i]))
			payloadCall, _ := callPayload.ToBytes()

			value, _ := util.NewUint128FromInt(6)
			gasLimit, _ := util.NewUint128FromInt(1000000)
			proxyContractAddress, err := core.AddressParse(contractsAddr[0])
			txCall, err := core.NewTransaction(neb.chain.ChainID(), a, proxyContractAddress, value,
				uint64(len(contractsAddr)+1), core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
			assert.Nil(t, err)
			assert.Nil(t, manager.SignTransaction(a, txCall))
			assert.Nil(t, neb.chain.TransactionPool().Push(txCall))

			block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
			assert.Nil(t, block.Seal())
			assert.Nil(t, manager.SignBlock(d, block))
			assert.Nil(t, neb.chain.BlockPool().Push(block))

			tail = neb.chain.TailBlock()
			events, err := tail.FetchEvents(txCall.Hash())
			for _, event := range events {
				fmt.Printf("event:%v\n", event.Data)
				var jEvent SysEvent
				if err := json.Unmarshal([]byte(event.Data), &jEvent); err == nil {
					if jEvent.Hash != "" {
						assert.Equal(t, tt.expectedErrArr[i], jEvent.Err)
					}
				}

			}
			//chech accout
			contractAddrA, err := core.AddressParse(contractsAddr[0])
			accountAAcc, err := tail.GetAccount(contractAddrA.Bytes())
			assert.Nil(t, err)
			// fmt.Printf("account :%v\n", accountAAcc)
			assert.Equal(t, tt.expectedAccount[i][0], accountAAcc.Balance().String())

			contractAddrB, err := core.AddressParse(contractsAddr[1])
			accountBAcc, err := tail.GetAccount(contractAddrB.Bytes())
			assert.Nil(t, err)
			// fmt.Printf("accountB :%v\n", accountBAcc)
			assert.Equal(t, tt.expectedAccount[i][1], accountBAcc.Balance().String())

			contractAddrC, err := core.AddressParse(contractsAddr[2])
			accountAccC, err := tail.GetAccount(contractAddrC.Bytes())
			assert.Nil(t, err)
			fmt.Printf("accountC :%v\n", accountAccC)
			assert.Equal(t, tt.expectedAccount[i][2], accountAccC.Balance().String())

			aUser, err := tail.GetAccount(a.Bytes())
			// assert.Equal(t, tt.expectedAccount[i][3], aUser.Balance().String())
			fmt.Printf("aI:%v\n", aUser)
			bUser, err := tail.GetAccount(b.Bytes())
			assert.Equal(t, tt.expectedAccount[i][4], bUser.Balance().String())

			cUser, err := tail.GetAccount(c.Bytes())
			fmt.Printf("cI:%v\n", cUser)
			// assert.Equal(t, tt.expectedAccount[i][5], cUser.Balance().String())

			// fmt.Printf("b:%v\n", bUser)
			// assert.Equal(t, tt.expectedAccount[i][4], cUser.Balance().String())
			dUser, err := tail.GetAccount(d.Bytes())
			assert.Equal(t, tt.expectedAccount[i][6], dUser.Balance().String())
			// fmt.Printf("d:%v\n", dUser)
			// assert.Equal(t, tt.expectedAccount[i][4], dUser.Balance().String())
		}
	}
}

func TestGetContractErr(t *testing.T) {
	core.NebCompatibility = core.NewCompatibilityLocal()
	tests := []struct {
		name      string
		contracts []contract
		calls     []call
	}{
		{
			"TestGetContractErr",
			[]contract{
				contract{
					"./test/test_inner_transaction.js",
					"js",
					"",
				},
				contract{
					"./test/bank_vault_contract_second.js",
					"js",
					"",
				},
			},
			[]call{
				call{
					"getSource",
					"[1]",
					[]string{"Call: Inner Call: no contract at this address"},
				},
			},
		},
	}

	for _, tt := range tests {
		for i := 0; i < len(tt.calls); i++ {

			neb := mockNeb(t)
			tail := neb.chain.TailBlock()
			manager, err := account.NewManager(neb)
			assert.Nil(t, err)

			a, _ := core.AddressParse("n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE")
			assert.Nil(t, manager.Unlock(a, []byte("passphrase"), keystore.YearUnlockDuration))
			b, _ := core.AddressParse("n1GmkKH6nBMw4rrjt16RrJ9WcgvKUtAZP1s")
			assert.Nil(t, manager.Unlock(b, []byte("passphrase"), keystore.YearUnlockDuration))
			c, _ := core.AddressParse("n1H4MYms9F55ehcvygwWE71J8tJC4CRr2so")
			assert.Nil(t, manager.Unlock(c, []byte("passphrase"), keystore.YearUnlockDuration))
			d, _ := core.AddressParse("n1JAy4X6KKLCNiTd7MWMRsVBjgdVq5WCCpf")
			assert.Nil(t, manager.Unlock(d, []byte("passphrase"), keystore.YearUnlockDuration))

			elapsedSecond := dpos.BlockIntervalInMs / dpos.SecondInMs
			consensusState, err := tail.WorldState().NextConsensusState(elapsedSecond)
			assert.Nil(t, err)
			// mock empty block(height=2)
			block, err := core.MockBlockEx(neb.chain.ChainID(), c, tail, 2)
			fmt.Printf("mock 2, block.height:%v\n", block.Height())
			assert.Nil(t, err)
			/* ----- mock random seed for new block ------*/
			miner, err := core.AddressParseFromBytes(consensusState.Proposer())
			assert.Nil(t, err)
			seed, proof, err := manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), neb.chain.GenesisBlock().Hash()) // NOTE: 3rd arg is genesis's hash for the first block
			assert.Nil(t, err)
			block.SetRandomSeed(seed, proof)
			/* ----- mock random seed for new block END ------*/

			block.WorldState().SetConsensusState(consensusState)
			block.SetTimestamp(consensusState.TimeStamp())
			block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
			assert.Nil(t, block.Seal())
			assert.Nil(t, manager.SignBlock(b, block))
			assert.Nil(t, neb.chain.BlockPool().Push(block))
			fmt.Printf("mock 2, block.tailblock.height: %v\n", neb.chain.TailBlock().Height())

			// inner call block(height=3)
			tail = neb.chain.TailBlock()
			block, err = core.MockBlockEx(neb.chain.ChainID(), c, tail, 3)
			assert.Nil(t, err)
			consensusState, err = tail.WorldState().NextConsensusState(elapsedSecond)
			assert.Nil(t, err)
			/* ----- mock random seed for new block ------*/
			miner, err = core.AddressParseFromBytes(consensusState.Proposer())
			assert.Nil(t, err)
			seed, proof, err = manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), seed) // NOTE: 3rd arg is parent's seed
			assert.Nil(t, err)
			block.SetRandomSeed(seed, proof)
			/* ----- mock random seed for new block END ------*/
			block.WorldState().SetConsensusState(consensusState)
			block.SetTimestamp(consensusState.TimeStamp())

			contractsAddr := []string{}
			for k, v := range tt.contracts {
				data, err := ioutil.ReadFile(v.contractPath)
				assert.Nil(t, err, "contract path read error")
				source := string(data)
				sourceType := "js"
				argsDeploy := ""
				deploy, _ := core.NewDeployPayload(source, sourceType, argsDeploy)
				payloadDeploy, _ := deploy.ToBytes()

				value, _ := util.NewUint128FromInt(0)
				gasLimit, _ := util.NewUint128FromInt(200000)
				txDeploy, err := core.NewTransaction(neb.chain.ChainID(), a, a, value, uint64(k+1), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
				assert.Nil(t, err)
				assert.Nil(t, manager.SignTransaction(a, txDeploy))
				assert.Nil(t, neb.chain.TransactionPool().Push(txDeploy))

				contractAddr, err := txDeploy.GenerateContractAddress()
				assert.Nil(t, err)
				contractsAddr = append(contractsAddr, contractAddr.String())
			}

			block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
			assert.Nil(t, block.Seal())
			assert.Nil(t, manager.SignBlock(c, block))
			assert.Nil(t, neb.chain.BlockPool().Push(block))

			for _, v := range contractsAddr {
				contract, err := core.AddressParse(v)
				assert.Nil(t, err)
				_, err = neb.chain.TailBlock().CheckContract(contract)
				assert.Nil(t, err)
			}

			elapsedSecond = dpos.BlockIntervalInMs / dpos.SecondInMs
			tail = neb.chain.TailBlock()
			consensusState, err = tail.WorldState().NextConsensusState(elapsedSecond)
			assert.Nil(t, err)
			block, err = core.NewBlock(neb.chain.ChainID(), c, tail)
			assert.Nil(t, err)
			/* ----- mock random seed for new block ------*/
			miner, err = core.AddressParseFromBytes(consensusState.Proposer())
			assert.Nil(t, err)
			seed, proof, err = manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), seed) // NOTE: 3rd arg is parent's seed
			assert.Nil(t, err)
			block.SetRandomSeed(seed, proof)
			/* ----- mock random seed for new block END ------*/
			block.WorldState().SetConsensusState(consensusState)
			block.SetTimestamp(consensusState.TimeStamp())
			assert.Nil(t, err)

			calleeContract := "123456789"
			callToContract := "123456789"
			callPayload, _ := core.NewCallPayload(tt.calls[i].function, fmt.Sprintf("[\"%s\", \"%s\"]", calleeContract, callToContract))
			payloadCall, _ := callPayload.ToBytes()

			value, _ := util.NewUint128FromInt(6)
			gasLimit, _ := util.NewUint128FromInt(1000000)
			proxyContractAddress, err := core.AddressParse(contractsAddr[0])
			txCall, err := core.NewTransaction(neb.chain.ChainID(), a, proxyContractAddress, value,
				uint64(len(contractsAddr)+1), core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
			assert.Nil(t, err)
			assert.Nil(t, manager.SignTransaction(a, txCall))
			assert.Nil(t, neb.chain.TransactionPool().Push(txCall))

			block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
			assert.Nil(t, block.Seal())
			assert.Nil(t, manager.SignBlock(d, block))
			assert.Nil(t, neb.chain.BlockPool().Push(block))

			tail = neb.chain.TailBlock()
			events, err := tail.FetchEvents(txCall.Hash())
			for _, event := range events {
				fmt.Printf("event:%v\n", events)
				var jEvent SysEvent
				if err := json.Unmarshal([]byte(event.Data), &jEvent); err == nil {
					if jEvent.Hash != "" {
						assert.Equal(t, tt.calls[i].exceptArgs[0], jEvent.Err)
					}
					fmt.Printf("event:%v\n", jEvent.Err)
				}

			}
		}
	}
}

func TestInnerTransactionsRand(t *testing.T) {
	core.NebCompatibility = core.NewCompatibilityLocal()
	tests := []struct {
		name      string
		contracts []contract
		call      call
	}{
		{
			"test TestInnerTransactionsRand",
			[]contract{
				contract{
					"./test/test_inner_transaction.js",
					"js",
					"",
				},
				contract{
					"./test/bank_vault_contract_second.js",
					"js",
					"",
				},
				contract{
					"./test/bank_vault_contract.js",
					"js",
					"",
				},
			},
			call{
				"getRandom",
				"[1]",
				[]string{""},
			},
		},
	}

	for _, tt := range tests {
		// for i := 0; i < len(tt); i++ {

		neb := mockNeb(t)
		tail := neb.chain.TailBlock()
		manager, err := account.NewManager(neb)
		assert.Nil(t, err)

		a, _ := core.AddressParse("n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE")
		assert.Nil(t, manager.Unlock(a, []byte("passphrase"), keystore.YearUnlockDuration))
		b, _ := core.AddressParse("n1GmkKH6nBMw4rrjt16RrJ9WcgvKUtAZP1s")
		assert.Nil(t, manager.Unlock(b, []byte("passphrase"), keystore.YearUnlockDuration))
		c, _ := core.AddressParse("n1H4MYms9F55ehcvygwWE71J8tJC4CRr2so")
		assert.Nil(t, manager.Unlock(c, []byte("passphrase"), keystore.YearUnlockDuration))
		d, _ := core.AddressParse("n1JAy4X6KKLCNiTd7MWMRsVBjgdVq5WCCpf")
		assert.Nil(t, manager.Unlock(d, []byte("passphrase"), keystore.YearUnlockDuration))

		elapsedSecond := dpos.BlockIntervalInMs / dpos.SecondInMs
		consensusState, err := tail.WorldState().NextConsensusState(elapsedSecond)
		assert.Nil(t, err)
		// block, err := core.NewBlock(neb.chain.ChainID(), b, tail)
		block, err := core.MockBlockEx(neb.chain.ChainID(), c, tail, 2)
		assert.Nil(t, err)
		miner, _ := core.AddressParseFromBytes(consensusState.Proposer())
		// fmt.Println("====", miner.String()) // n1GmkKH6nBMw4rrjt16RrJ9WcgvKUtAZP1s
		seed, proof, err := manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), neb.chain.GenesisBlock().Hash())
		block.SetRandomSeed(seed, proof)

		block.WorldState().SetConsensusState(consensusState)
		block.SetTimestamp(consensusState.TimeStamp())
		block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
		assert.Nil(t, block.Seal())
		assert.Nil(t, manager.SignBlock(b, block))
		assert.Nil(t, neb.chain.BlockPool().Push(block))
		fmt.Printf("mock 2, block.tailblock.height: %v\n", neb.chain.TailBlock().Height())

		// inner call block(height=3)
		tail = neb.chain.TailBlock()
		block, err = core.MockBlockEx(neb.chain.ChainID(), c, tail, 3)
		assert.Nil(t, err)
		consensusState, err = tail.WorldState().NextConsensusState(elapsedSecond)
		assert.Nil(t, err)

		/* ----- mock random seed for new block ------*/
		miner, err = core.AddressParseFromBytes(consensusState.Proposer())
		assert.Nil(t, err)
		seed, proof, err = manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), seed) // NOTE: 3rd arg is parent's seed
		assert.Nil(t, err)
		block.SetRandomSeed(seed, proof)
		/* ----- mock random seed for new block END ------*/
		block.WorldState().SetConsensusState(consensusState)
		block.SetTimestamp(consensusState.TimeStamp())

		contractsAddr := []string{}
		for k, v := range tt.contracts {
			data, err := ioutil.ReadFile(v.contractPath)
			assert.Nil(t, err, "contract path read error")
			source := string(data)
			sourceType := "js"
			argsDeploy := ""
			deploy, _ := core.NewDeployPayload(source, sourceType, argsDeploy)
			payloadDeploy, _ := deploy.ToBytes()

			value, _ := util.NewUint128FromInt(0)
			gasLimit, _ := util.NewUint128FromInt(200000)
			txDeploy, err := core.NewTransaction(neb.chain.ChainID(), a, a, value, uint64(k+1), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
			assert.Nil(t, err)
			assert.Nil(t, manager.SignTransaction(a, txDeploy))
			assert.Nil(t, neb.chain.TransactionPool().Push(txDeploy))

			contractAddr, err := txDeploy.GenerateContractAddress()
			assert.Nil(t, err)
			contractsAddr = append(contractsAddr, contractAddr.String())
		}

		block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
		assert.Nil(t, block.Seal())
		assert.Nil(t, manager.SignBlock(c, block))
		assert.Nil(t, neb.chain.BlockPool().Push(block))

		for _, v := range contractsAddr {
			contract, err := core.AddressParse(v)
			assert.Nil(t, err)
			_, err = neb.chain.TailBlock().CheckContract(contract)
			assert.Nil(t, err)
		}

		elapsedSecond = dpos.BlockIntervalInMs / dpos.SecondInMs
		tail = neb.chain.TailBlock()
		consensusState, err = tail.WorldState().NextConsensusState(elapsedSecond)
		assert.Nil(t, err)
		block, err = core.NewBlock(neb.chain.ChainID(), b, tail)
		assert.Nil(t, err)

		block.WorldState().SetConsensusState(consensusState)
		block.SetTimestamp(consensusState.TimeStamp())
		assert.Nil(t, err)

		miner, err = core.AddressParseFromBytes(consensusState.Proposer())
		assert.Nil(t, err)
		seed, proof, err = manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), seed)
		assert.Nil(t, err)
		block.SetRandomSeed(seed, proof)

		calleeContract := contractsAddr[1]
		callToContract := contractsAddr[2]
		callPayload, _ := core.NewCallPayload(tt.call.function, fmt.Sprintf("[\"%s\", \"%s\"]", calleeContract, callToContract))
		payloadCall, _ := callPayload.ToBytes()

		value, _ := util.NewUint128FromInt(6)
		gasLimit, _ := util.NewUint128FromInt(int64(100000))
		proxyContractAddress, err := core.AddressParse(contractsAddr[0])
		txCall, err := core.NewTransaction(neb.chain.ChainID(), a, proxyContractAddress, value,
			uint64(len(contractsAddr)+1), core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
		assert.Nil(t, err)
		assert.Nil(t, manager.SignTransaction(a, txCall))
		assert.Nil(t, neb.chain.TransactionPool().Push(txCall))

		block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
		assert.Nil(t, block.Seal())
		assert.Nil(t, manager.SignBlock(d, block))
		assert.Nil(t, neb.chain.BlockPool().Push(block))

		tail = neb.chain.TailBlock()
		events, err := tail.FetchEvents(txCall.Hash())
		for _, event := range events {

			var jEvent SysEvent
			if err := json.Unmarshal([]byte(event.Data), &jEvent); err == nil {
				fmt.Printf("event:%v\n", event.Data)
				if jEvent.Hash != "" {
					assert.Equal(t, "", jEvent.Err)
				}
			}

		}
		// }
	}
}

func TestMultiLibVersion(t *testing.T) {
	tests := []struct {
		filepath       string
		expectedErr    error
		expectedResult string
	}{
		{"test/test_multi_lib_version_require.js", nil, "\"\""},
		{"test/test_uint.js", nil, "\"\""},
		{"test/test_date_1.0.5.js", nil, "\"\""},
		{"test/test_crypto.js", nil, "\"\""},
		{"test/test_blockchain_1.0.5.js", nil, "\"\""},
	}

	for _, tt := range tests {
		t.Run(tt.filepath, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.filepath)
			assert.Nil(t, err, "filepath read error")
			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewWorldState(dpos.NewDpos(), mem)
			addr, _ := core.AddressParse("n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE")
			owner, err := context.GetOrCreateUserAccount(addr.Bytes())
			assert.Nil(t, err)
			owner.AddBalance(newUint128FromIntWrapper(1000000000000))
			addr, _ = core.AddressParse("n1p8cwrrfrbFe71eda1PQ6y4WnX3gp8bYze")
			contract, _ := context.CreateContractAccount(addr.Bytes(), nil, &corepb.ContractMeta{Version: "1.0.5"})
			ctx, err := NewContext(mockBlockForLib(2000000), mockTransaction(), contract, context)

			engine := NewV8Engine(ctx)
			engine.SetExecutionLimits(5000000, 10000000)
			result, err := engine.RunScriptSource(string(data), 0)
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedResult, result)
			engine.Dispose()
		})
	}
}
func TestInnerTransactionsTimeOut(t *testing.T) {
	core.NebCompatibility = core.NewCompatibilityLocal()
	tests := []struct {
		name           string
		contracts      []contract
		call           call
		errFlagArr     []uint32
		expectedErrArr []string
	}{
		{
			"deploy TestInnerTransactionsErr.js",
			[]contract{
				contract{
					"./test/test_inner_transaction.js",
					"js",
					"",
				},
				contract{
					"./test/bank_vault_contract_second.js",
					"js",
					"",
				},
				contract{
					"./test/bank_vault_contract.js",
					"js",
					"",
				},
			},
			call{
				"saveTimeOut",
				"[1]",
				[]string{""},
			},
			[]uint32{0, 1, 2},
			[]string{"insufficient gas",
				"insufficient gas",
				"insufficient gas"},
		},
	}

	for _, tt := range tests {
		for i := 0; i < len(tt.errFlagArr); i++ {

			neb := mockNeb(t)
			tail := neb.chain.TailBlock()
			manager, err := account.NewManager(neb)
			assert.Nil(t, err)

			a, _ := core.AddressParse("n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE")
			assert.Nil(t, manager.Unlock(a, []byte("passphrase"), keystore.YearUnlockDuration))
			b, _ := core.AddressParse("n1GmkKH6nBMw4rrjt16RrJ9WcgvKUtAZP1s")
			assert.Nil(t, manager.Unlock(b, []byte("passphrase"), keystore.YearUnlockDuration))
			c, _ := core.AddressParse("n1H4MYms9F55ehcvygwWE71J8tJC4CRr2so")
			assert.Nil(t, manager.Unlock(c, []byte("passphrase"), keystore.YearUnlockDuration))
			d, _ := core.AddressParse("n1JAy4X6KKLCNiTd7MWMRsVBjgdVq5WCCpf")
			assert.Nil(t, manager.Unlock(d, []byte("passphrase"), keystore.YearUnlockDuration))

			elapsedSecond := dpos.BlockIntervalInMs / dpos.SecondInMs
			consensusState, err := tail.WorldState().NextConsensusState(elapsedSecond)
			assert.Nil(t, err)

			// mock empty block(height=2)
			block, err := core.MockBlockEx(neb.chain.ChainID(), c, tail, 2)
			fmt.Printf("mock 2, block.height:%v\n", block.Height())
			assert.Nil(t, err)
			/* ----- mock random seed for new block ------*/
			miner, err := core.AddressParseFromBytes(consensusState.Proposer())
			assert.Nil(t, err)
			seed, proof, err := manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), neb.chain.GenesisBlock().Hash()) // NOTE: 3rd arg is genesis's hash for the first block
			assert.Nil(t, err)
			block.SetRandomSeed(seed, proof)
			/* ----- mock random seed for new block END ------*/
			block.WorldState().SetConsensusState(consensusState)
			block.SetTimestamp(consensusState.TimeStamp())
			block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
			assert.Nil(t, block.Seal())
			assert.Nil(t, manager.SignBlock(b, block))
			assert.Nil(t, neb.chain.BlockPool().Push(block))
			fmt.Printf("mock 2, block.tailblock.height: %v\n", neb.chain.TailBlock().Height())

			// inner call block(height=3)
			tail = neb.chain.TailBlock()
			block, err = core.MockBlockEx(neb.chain.ChainID(), c, tail, 3)
			assert.Nil(t, err)
			consensusState, err = tail.WorldState().NextConsensusState(elapsedSecond)
			assert.Nil(t, err)
			/* ----- mock random seed for new block ------*/
			miner, err = core.AddressParseFromBytes(consensusState.Proposer())
			assert.Nil(t, err)
			seed, proof, err = manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), seed) // NOTE: 3rd arg is parent's seed
			assert.Nil(t, err)
			block.SetRandomSeed(seed, proof)
			/* ----- mock random seed for new block END ------*/

			block.WorldState().SetConsensusState(consensusState)
			block.SetTimestamp(consensusState.TimeStamp())

			contractsAddr := []string{}
			for k, v := range tt.contracts {
				data, err := ioutil.ReadFile(v.contractPath)
				assert.Nil(t, err, "contract path read error")
				source := string(data)
				sourceType := "js"
				argsDeploy := ""
				deploy, _ := core.NewDeployPayload(source, sourceType, argsDeploy)
				payloadDeploy, _ := deploy.ToBytes()

				value, _ := util.NewUint128FromInt(0)
				gasLimit, _ := util.NewUint128FromInt(5000000)
				txDeploy, err := core.NewTransaction(neb.chain.ChainID(), a, a, value, uint64(k+1), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
				assert.Nil(t, err)
				assert.Nil(t, manager.SignTransaction(a, txDeploy))
				assert.Nil(t, neb.chain.TransactionPool().Push(txDeploy))

				contractAddr, err := txDeploy.GenerateContractAddress()
				assert.Nil(t, err)
				contractsAddr = append(contractsAddr, contractAddr.String())
			}

			block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
			assert.Nil(t, block.Seal())
			assert.Nil(t, manager.SignBlock(c, block))
			assert.Nil(t, neb.chain.BlockPool().Push(block))

			for _, v := range contractsAddr {
				contract, err := core.AddressParse(v)
				assert.Nil(t, err)
				_, err = neb.chain.TailBlock().CheckContract(contract)
				assert.Nil(t, err)
			}

			elapsedSecond = dpos.BlockIntervalInMs / dpos.SecondInMs
			tail = neb.chain.TailBlock()
			consensusState, err = tail.WorldState().NextConsensusState(elapsedSecond)
			assert.Nil(t, err)
			block, err = core.NewBlock(neb.chain.ChainID(), b, tail)
			assert.Nil(t, err)
			/* ----- mock random seed for new block ------*/
			miner, err = core.AddressParseFromBytes(consensusState.Proposer())
			assert.Nil(t, err)
			seed, proof, err = manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), seed) // NOTE: 3rd arg is parent's seed
			assert.Nil(t, err)
			block.SetRandomSeed(seed, proof)
			/* ----- mock random seed for new block END ------*/
			block.WorldState().SetConsensusState(consensusState)
			block.SetTimestamp(consensusState.TimeStamp())
			assert.Nil(t, err)

			calleeContract := contractsAddr[1]
			callToContract := contractsAddr[2]
			callPayload, _ := core.NewCallPayload(tt.call.function, fmt.Sprintf("[\"%s\", \"%s\", \"%d\"]", calleeContract, callToContract, tt.errFlagArr[i]))
			payloadCall, _ := callPayload.ToBytes()

			value, _ := util.NewUint128FromInt(6)
			gasLimit, _ := util.NewUint128FromInt(1000000)
			proxyContractAddress, err := core.AddressParse(contractsAddr[0])
			txCall, err := core.NewTransaction(neb.chain.ChainID(), a, proxyContractAddress, value,
				uint64(len(contractsAddr)+1), core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
			assert.Nil(t, err)
			assert.Nil(t, manager.SignTransaction(a, txCall))
			assert.Nil(t, neb.chain.TransactionPool().Push(txCall))

			block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
			assert.Nil(t, block.Seal())
			assert.Nil(t, manager.SignBlock(d, block))
			assert.Nil(t, neb.chain.BlockPool().Push(block))

			tail = neb.chain.TailBlock()
			events, err := tail.FetchEvents(txCall.Hash())
			for _, event := range events {
				fmt.Printf("event:%v\n", event.Data)
				var jEvent SysEvent
				if err := json.Unmarshal([]byte(event.Data), &jEvent); err == nil {
					if jEvent.Hash != "" {
						assert.Equal(t, tt.expectedErrArr[i], jEvent.Err)
					}
				}

			}
		}
	}
}
func TestInnerTxInstructionCounter(t *testing.T) {
	core.NebCompatibility = core.NewCompatibilityLocal()
	tests := []struct {
		name      string
		contracts []contract
		calls     []call
	}{
		{
			"deploy contracts",
			[]contract{
				contract{
					"./test/instruction_counter_tests/inner_contract_callee.js",
					"js",
					"",
				},
				contract{
					"./test/instruction_counter_tests/inner_contract_caller.js",
					"js",
					"",
				},
			},
			[]call{
				call{
					"callWhile",
					"",
					[]string{"57296000000"}, // 57286000000  for instruction_counter.js v1.0.0
				},
			},
		},
	}
	tt := tests[0]
	for _, call := range tt.calls {

		neb := mockNeb(t)
		tail := neb.chain.TailBlock()
		manager, err := account.NewManager(neb)
		assert.Nil(t, err)

		a, _ := core.AddressParse("n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE")
		assert.Nil(t, manager.Unlock(a, []byte("passphrase"), keystore.YearUnlockDuration))
		b, _ := core.AddressParse("n1GmkKH6nBMw4rrjt16RrJ9WcgvKUtAZP1s")
		assert.Nil(t, manager.Unlock(b, []byte("passphrase"), keystore.YearUnlockDuration))
		c, _ := core.AddressParse("n1H4MYms9F55ehcvygwWE71J8tJC4CRr2so")
		assert.Nil(t, manager.Unlock(c, []byte("passphrase"), keystore.YearUnlockDuration))
		d, _ := core.AddressParse("n1JAy4X6KKLCNiTd7MWMRsVBjgdVq5WCCpf")
		assert.Nil(t, manager.Unlock(d, []byte("passphrase"), keystore.YearUnlockDuration))

		elapsedSecond := dpos.BlockIntervalInMs / dpos.SecondInMs
		consensusState, err := tail.WorldState().NextConsensusState(elapsedSecond)
		assert.Nil(t, err)
		// mock empty block(height=2)
		block, err := core.MockBlockEx(neb.chain.ChainID(), c, tail, 2)
		fmt.Printf("mock 2, block.height:%v\n", block.Height())
		assert.Nil(t, err)
		/* ----- mock random seed for new block ------*/
		miner, err := core.AddressParseFromBytes(consensusState.Proposer())
		assert.Nil(t, err)
		seed, proof, err := manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), neb.chain.GenesisBlock().Hash()) // NOTE: 3rd arg is genesis's hash for the first block
		assert.Nil(t, err)
		block.SetRandomSeed(seed, proof)
		/* ----- mock random seed for new block END ------*/
		block.WorldState().SetConsensusState(consensusState)
		block.SetTimestamp(consensusState.TimeStamp())
		block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
		assert.Nil(t, block.Seal())
		assert.Nil(t, manager.SignBlock(b, block))
		assert.Nil(t, neb.chain.BlockPool().Push(block))
		fmt.Printf("mock 2, block.tailblock.height: %v\n", neb.chain.TailBlock().Height())

		// inner call block(height=3)
		tail = neb.chain.TailBlock()
		block, err = core.MockBlockEx(neb.chain.ChainID(), c, tail, 3)
		assert.Nil(t, err)
		consensusState, err = tail.WorldState().NextConsensusState(elapsedSecond)
		assert.Nil(t, err)
		/* ----- mock random seed for new block ------*/
		miner, err = core.AddressParseFromBytes(consensusState.Proposer())
		assert.Nil(t, err)
		seed, proof, err = manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), seed) // NOTE: 3rd arg is parent's seed
		assert.Nil(t, err)
		block.SetRandomSeed(seed, proof)
		/* ----- mock random seed for new block END ------*/
		block.WorldState().SetConsensusState(consensusState)
		block.SetTimestamp(consensusState.TimeStamp())

		contractsAddr := []string{}

		// t.Run(tt.name, func(t *testing.T) {
		for k, v := range tt.contracts {
			data, err := ioutil.ReadFile(v.contractPath)
			assert.Nil(t, err, "contract path read error")
			source := string(data)
			sourceType := "js"
			argsDeploy := ""
			deploy, _ := core.NewDeployPayload(source, sourceType, argsDeploy)
			payloadDeploy, _ := deploy.ToBytes()

			value, _ := util.NewUint128FromInt(0)
			gasLimit, _ := util.NewUint128FromInt(200000)
			txDeploy, err := core.NewTransaction(neb.chain.ChainID(), a, a, value, uint64(k+1), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
			assert.Nil(t, err)
			assert.Nil(t, manager.SignTransaction(a, txDeploy))
			assert.Nil(t, neb.chain.TransactionPool().Push(txDeploy))

			contractAddr, err := txDeploy.GenerateContractAddress()
			assert.Nil(t, err)
			contractsAddr = append(contractsAddr, contractAddr.String())
		}
		// })

		block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
		assert.Nil(t, block.Seal())
		assert.Nil(t, manager.SignBlock(c, block))
		assert.Nil(t, neb.chain.BlockPool().Push(block))

		for _, v := range contractsAddr {
			contract, err := core.AddressParse(v)
			assert.Nil(t, err)
			_, err = neb.chain.TailBlock().CheckContract(contract)
			assert.Nil(t, err)
		}

		elapsedSecond = dpos.BlockIntervalInMs / dpos.SecondInMs
		tail = neb.chain.TailBlock()
		consensusState, err = tail.WorldState().NextConsensusState(elapsedSecond)
		assert.Nil(t, err)
		block, err = core.NewBlock(neb.chain.ChainID(), c, tail)
		assert.Nil(t, err)
		/* ----- mock random seed for new block ------*/
		miner, err = core.AddressParseFromBytes(consensusState.Proposer())
		assert.Nil(t, err)
		seed, proof, err = manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), seed) // NOTE: 3rd arg is parent's seed
		assert.Nil(t, err)
		block.SetRandomSeed(seed, proof)
		/* ----- mock random seed for new block END ------*/
		block.WorldState().SetConsensusState(consensusState)
		block.SetTimestamp(consensusState.TimeStamp())
		//accountA, err := tail.GetAccount(a.Bytes())
		//accountB, err := tail.GetAccount(b.Bytes())
		assert.Nil(t, err)

		callPayload, _ := core.NewCallPayload(call.function, fmt.Sprintf("[\"%s\"]", contractsAddr[0]))
		payloadCall, _ := callPayload.ToBytes()

		value, _ := util.NewUint128FromInt(0)
		gasLimit, _ := util.NewUint128FromInt(200000)

		aUser, err := tail.GetAccount(a.Bytes())
		balBefore := aUser.Balance()

		proxyContractAddress, err := core.AddressParse(contractsAddr[1])
		txCall, err := core.NewTransaction(neb.chain.ChainID(), a, proxyContractAddress, value,
			uint64(len(contractsAddr)+1), core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
		assert.Nil(t, err)
		assert.Nil(t, manager.SignTransaction(a, txCall))
		assert.Nil(t, neb.chain.TransactionPool().Push(txCall))

		block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
		assert.Nil(t, block.Seal())
		assert.Nil(t, manager.SignBlock(d, block))
		assert.Nil(t, neb.chain.BlockPool().Push(block))

		// // check
		tail = neb.chain.TailBlock()
		events, err := tail.FetchEvents(txCall.Hash())
		assert.Nil(t, err)
		for _, event := range events {
			fmt.Println("==============", event.Data)
		}

		aUser, err = tail.GetAccount(a.Bytes())
		assert.Nil(t, err)
		det, err := balBefore.Sub(aUser.Balance())
		assert.Nil(t, err)
		// fmt.Println("from account balance change: ", det.String())
		assert.Equal(t, call.exceptArgs[0], det.String())
		fmt.Printf("aI:%v\n", aUser)
	}
}

func TestMultiLibVersionCall(t *testing.T) {
	core.NebCompatibility = core.NewCompatibilityLocal()
	m := core.NebCompatibility.V8JSLibVersionHeightMap().Data
	m["1.1.0"] = 4
	m["1.0.5"] = 3
	defer func() {
		m["1.1.0"] = 3
		m["1.0.5"] = 2
	}()

	tests := []struct {
		name  string
		calls []call
	}{
		{
			"call contracts",
			[]call{
				call{
					"testInnerCall",
					"[\"%s\", \"undefined\"]", // in version before 1.1.0, typeof(Blockchain.Contract)=='undefined'
					[]string{"\"\"", ""},
				},
				call{
					"testRandom",
					// before 1.1.0
					// true -- Blockchain.block.seed is not empty
					// true -- Math.random.seed exists
					//
					// since 1.1.0
					// false -- Blockchain.block.seed is empty
					// false -- Math.random.seed not exist
					"[\"%s\", true, true, false, false]",
					[]string{"\"\"", ""},
				},
			},
		},
	}
	tt := tests[0]

	neb := mockNeb(t)
	tail := neb.chain.TailBlock()
	manager, err := account.NewManager(neb)
	assert.Nil(t, err)

	a, _ := core.AddressParse("n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE")
	assert.Nil(t, manager.Unlock(a, []byte("passphrase"), keystore.YearUnlockDuration))
	b, _ := core.AddressParse("n1GmkKH6nBMw4rrjt16RrJ9WcgvKUtAZP1s")
	assert.Nil(t, manager.Unlock(b, []byte("passphrase"), keystore.YearUnlockDuration))
	c, _ := core.AddressParse("n1H4MYms9F55ehcvygwWE71J8tJC4CRr2so")
	assert.Nil(t, manager.Unlock(c, []byte("passphrase"), keystore.YearUnlockDuration))
	d, _ := core.AddressParse("n1JAy4X6KKLCNiTd7MWMRsVBjgdVq5WCCpf")
	assert.Nil(t, manager.Unlock(d, []byte("passphrase"), keystore.YearUnlockDuration))
	e, _ := core.AddressParse("n1LkDi2gGMqPrjYcczUiweyP4RxTB6Go1qS")
	assert.Nil(t, manager.Unlock(e, []byte("passphrase"), keystore.YearUnlockDuration))

	elapsedSecond := dpos.BlockIntervalInMs / dpos.SecondInMs
	consensusState, err := tail.WorldState().NextConsensusState(elapsedSecond)
	assert.Nil(t, err)
	// mock empty block(height=2)
	block, err := core.MockBlockEx(neb.chain.ChainID(), c, tail, 2)
	fmt.Printf("mock 2, block.height:%v\n", block.Height())
	assert.Nil(t, err)
	/* ----- mock random seed for new block ------*/
	miner, err := core.AddressParseFromBytes(consensusState.Proposer())
	assert.Nil(t, err)
	seed, proof, err := manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), neb.chain.GenesisBlock().Hash()) // NOTE: 3rd arg is genesis's hash for the first block
	assert.Nil(t, err)
	block.SetRandomSeed(seed, proof)
	/* ----- mock random seed for new block END ------*/
	block.WorldState().SetConsensusState(consensusState)
	block.SetTimestamp(consensusState.TimeStamp())
	// packing
	block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
	assert.Nil(t, block.Seal())
	assert.Nil(t, manager.SignBlock(b, block))
	assert.Nil(t, neb.chain.BlockPool().Push(block))
	fmt.Printf("mock 2, block.tailblock.height: %v\n", neb.chain.TailBlock().Height())

	// ===========new block(height=3), deploy callee.js=========
	tail = neb.chain.TailBlock()
	block, err = core.MockBlockEx(neb.chain.ChainID(), c, tail, 3)
	assert.Nil(t, err)
	consensusState, err = tail.WorldState().NextConsensusState(elapsedSecond)
	assert.Nil(t, err)
	/* ----- mock random seed for new block ------*/
	miner, err = core.AddressParseFromBytes(consensusState.Proposer())
	assert.Nil(t, err)
	seed, proof, err = manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), seed) // NOTE: 3rd arg is parent's seed
	assert.Nil(t, err)
	block.SetRandomSeed(seed, proof)
	/* ----- mock random seed for new block END ------*/
	block.WorldState().SetConsensusState(consensusState)
	block.SetTimestamp(consensusState.TimeStamp())

	contractsAddr := []string{}

	data, err := ioutil.ReadFile("./test/inner_call_tests/callee.js")
	assert.Nil(t, err, "contract path read error")
	source := string(data)
	sourceType := "js"
	argsDeploy := ""
	deploy, _ := core.NewDeployPayload(source, sourceType, argsDeploy)
	payloadDeploy, _ := deploy.ToBytes()

	value, _ := util.NewUint128FromInt(0)
	gasLimit, _ := util.NewUint128FromInt(200000)
	txDeploy, err := core.NewTransaction(neb.chain.ChainID(), a, a, value, uint64(1), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
	assert.Nil(t, err)
	assert.Nil(t, manager.SignTransaction(a, txDeploy))
	assert.Nil(t, neb.chain.TransactionPool().Push(txDeploy))

	contractAddr, err := txDeploy.GenerateContractAddress()
	assert.Nil(t, err)
	contractsAddr = append(contractsAddr, contractAddr.String())

	block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
	assert.Nil(t, block.Seal())
	assert.Nil(t, manager.SignBlock(c, block))
	assert.Nil(t, neb.chain.BlockPool().Push(block))
	// ===========new block(height=3), deploy callee.js END =========

	// ===========new block(height=4), deploy caller.js =========
	tail = neb.chain.TailBlock()
	block, err = core.MockBlockEx(neb.chain.ChainID(), c, tail, 4)
	assert.Nil(t, err)
	consensusState, err = tail.WorldState().NextConsensusState(elapsedSecond)
	assert.Nil(t, err)
	/* ----- mock random seed for new block ------*/
	miner, err = core.AddressParseFromBytes(consensusState.Proposer())
	assert.Nil(t, err)
	seed, proof, err = manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), seed) // NOTE: 3rd arg is parent's seed
	assert.Nil(t, err)
	block.SetRandomSeed(seed, proof)
	/* ----- mock random seed for new block END ------*/
	block.WorldState().SetConsensusState(consensusState)
	block.SetTimestamp(consensusState.TimeStamp())

	data, err = ioutil.ReadFile("./test/inner_call_tests/caller.js")
	assert.Nil(t, err, "contract path read error")
	source = string(data)
	sourceType = "js"
	argsDeploy = ""
	deploy, _ = core.NewDeployPayload(source, sourceType, argsDeploy)
	payloadDeploy, _ = deploy.ToBytes()

	txDeploy, err = core.NewTransaction(neb.chain.ChainID(), a, a, value, uint64(2), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
	assert.Nil(t, err)
	assert.Nil(t, manager.SignTransaction(a, txDeploy))
	assert.Nil(t, neb.chain.TransactionPool().Push(txDeploy))

	contractAddr, err = txDeploy.GenerateContractAddress()
	assert.Nil(t, err)
	contractsAddr = append(contractsAddr, contractAddr.String())

	block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
	assert.Nil(t, block.Seal())
	assert.Nil(t, manager.SignBlock(d, block))
	assert.Nil(t, neb.chain.BlockPool().Push(block))
	//  ===========new block(height=4), deploy caller.js END =========

	for _, v := range contractsAddr {
		contract, err := core.AddressParse(v)
		assert.Nil(t, err)
		_, err = neb.chain.TailBlock().CheckContract(contract)
		assert.Nil(t, err)
	}

	for _, call := range tt.calls {
		elapsedSecond = dpos.BlockIntervalInMs / dpos.SecondInMs
		tail = neb.chain.TailBlock()
		consensusState, err = tail.WorldState().NextConsensusState(elapsedSecond)
		assert.Nil(t, err)
		block, err = core.NewBlock(neb.chain.ChainID(), c, tail)
		assert.Nil(t, err)
		/* ----- mock random seed for new block ------*/
		miner, err = core.AddressParseFromBytes(consensusState.Proposer())
		assert.Nil(t, err)
		seed, proof, err = manager.GenerateRandomSeed(miner, neb.chain.GenesisBlock().Hash(), seed) // NOTE: 3rd arg is parent's seed
		assert.Nil(t, err)
		block.SetRandomSeed(seed, proof)
		/* ----- mock random seed for new block END ------*/
		block.WorldState().SetConsensusState(consensusState)
		block.SetTimestamp(consensusState.TimeStamp())

		callPayload, _ := core.NewCallPayload(call.function, fmt.Sprintf(call.args, contractsAddr[0]))
		payloadCall, _ := callPayload.ToBytes()

		proxyContractAddress, err := core.AddressParse(contractsAddr[1])
		txCall, err := core.NewTransaction(neb.chain.ChainID(), a, proxyContractAddress, value,
			uint64(len(contractsAddr)+1), core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
		assert.Nil(t, err)
		assert.Nil(t, manager.SignTransaction(a, txCall))
		assert.Nil(t, neb.chain.TransactionPool().Push(txCall))

		block.CollectTransactions((time.Now().Unix() + 1) * dpos.SecondInMs)
		events, err := block.WorldState().FetchEvents(txCall.Hash())
		assert.Nil(t, err)
		for _, event := range events {
			fmt.Println("==============", event.Data)

			var jEvent SysEvent
			if event.Topic == core.TopicTransactionExecutionResult {
				if err = json.Unmarshal([]byte(event.Data), &jEvent); err == nil {
					assert.Equal(t, call.exceptArgs[0], jEvent.Result)
					assert.Equal(t, call.exceptArgs[1], jEvent.Err)
				}
			}
		}

	}
}
