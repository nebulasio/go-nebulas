package nvm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/nebulasio/go-nebulas/account"
	"github.com/nebulasio/go-nebulas/consensus/dpos"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/stretchr/testify/assert"
)

type deployTx struct {
	value    int64
	gasLimit int64
	contract contract
}

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

func mockNeb(t *testing.T) core.Neblet {
	am, err := account.NewManager(nil)
	assert.Nil(t, err)
	consensus := dpos.NewDpos()
	nvm := NewNebulasVM()
	neb := core.NewMockNeb(am, consensus, nvm)
	return neb
}

func generateRandom(t *testing.T, neb core.Neblet, block *core.Block) {
	ancestorHash, parentSeed, err := neb.BlockChain().GetInputForVRFSigner(block.ParentHash(), block.Height())
	assert.Nil(t, err)

	proposer := block.WorldState().ConsensusRoot().Proposer
	miner, err := core.AddressParseFromBytes(proposer)
	assert.Nil(t, err)

	// generate VRF hash,proof
	vrfSeed, vrfProof, err := neb.AccountManager().GenerateRandomSeed(miner, ancestorHash, parentSeed)
	assert.Nil(t, err)
	block.SetRandomSeed(vrfSeed, vrfProof)
}

func unlockAccount(t *testing.T, neb core.Neblet) []*core.Address {
	manager := neb.AccountManager()
	dynasty := neb.Genesis().Consensus.Dpos.Dynasty
	addrs := []*core.Address{}
	for _, v := range dynasty {
		addr, err := core.AddressParse(v)
		assert.Nil(t, err)
		assert.Nil(t, manager.Unlock(addr, []byte("passphrase"), keystore.YearUnlockDuration))
		addrs = append(addrs, addr)
	}
	//a, _ := core.AddressParse("n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE")
	//assert.Nil(t, manager.Unlock(a, []byte("passphrase"), keystore.YearUnlockDuration))
	//b, _ := core.AddressParse("n1GmkKH6nBMw4rrjt16RrJ9WcgvKUtAZP1s")
	//assert.Nil(t, manager.Unlock(b, []byte("passphrase"), keystore.YearUnlockDuration))
	//c, _ := core.AddressParse("n1H4MYms9F55ehcvygwWE71J8tJC4CRr2so")
	//assert.Nil(t, manager.Unlock(c, []byte("passphrase"), keystore.YearUnlockDuration))
	//d, _ := core.AddressParse("n1JAy4X6KKLCNiTd7MWMRsVBjgdVq5WCCpf")
	//assert.Nil(t, manager.Unlock(d, []byte("passphrase"), keystore.YearUnlockDuration))
	return addrs
}

func mintBlock(t *testing.T, neb core.Neblet, txs []*core.Transaction) {
	mintBlockWithDuration(t, neb, txs, int64(1))
}

func mintBlockWithDuration(t *testing.T, neb core.Neblet, txs []*core.Transaction, duration int64) {
	tail := neb.BlockChain().TailBlock()
	manager := neb.AccountManager()
	elapsedSecond := dpos.BlockIntervalInMs / dpos.SecondInMs
	consensusState, err := tail.WorldState().NextConsensusState(elapsedSecond)
	assert.Nil(t, err)

	coinbase, err := core.AddressParse(neb.Config().Chain.Coinbase)
	assert.Nil(t, err)
	block, err := core.NewBlock(neb.BlockChain().ChainID(), coinbase, tail)
	assert.Nil(t, err)
	block.WorldState().SetConsensusState(consensusState)
	block.SetTimestamp(consensusState.TimeStamp())
	generateRandom(t, neb, block)

	for _, tx := range txs {
		assert.Nil(t, neb.BlockChain().TransactionPool().Push(tx))
	}

	block.CollectTransactions((time.Now().Unix() + duration) * dpos.SecondInMs)
	assert.Nil(t, block.Seal())
	err = manager.SignBlock(block.Miner(), block)
	assert.Nil(t, err)
	assert.Nil(t, neb.BlockChain().BlockPool().Push(block))
	assert.Equal(t, block.Hash(), neb.BlockChain().TailBlock().Hash())
}

//block->a empty coinbase c . block->b deploy and accout is a and coinbase is c . block->c run js
func TestInnerTransactions(t *testing.T) {
	tests := []struct {
		name  string
		txs   []deployTx
		calls []call
	}{
		{
			"inner transaction",
			[]deployTx{{
				0, 200000,
				contract{
					"./test/inner_call_tests/test_inner_transaction.js",
					"js",
					"",
				},
			},
				{
					0, 200000,
					contract{
						"./test/inner_call_tests/bank_vault_inner_contract.js",
						"js",
						"",
					},
				},
				{
					0, 200000,
					contract{
						"./test/inner_call_tests/bank_vault_final_contract.js",
						"js",
						"",
					},
				},
			},
			[]call{
				call{
					"save",
					"[1]",
					[]string{"1", "3", "2", "4999999999999833225999994"},
				},
			},
		},
	}

	core.NebCompatibility = core.NewCompatibilityLocal()
	neb := mockNeb(t)
	manager := neb.AccountManager()

	addrs := unlockAccount(t, neb)

	// mint height 2, inner contracts >= 3
	mintBlock(t, neb, nil)

	tt := tests[0]
	for _, call := range tt.calls {

		txs := []*core.Transaction{}
		contracts := []*core.Address{}
		for k, tx := range tt.txs {
			data, err := ioutil.ReadFile(tx.contract.contractPath)
			assert.Nil(t, err, "contract path read error")
			source := string(data)
			sourceType := "js"
			argsDeploy := ""
			deploy, _ := core.NewDeployPayload(source, sourceType, argsDeploy)
			payloadDeploy, _ := deploy.ToBytes()

			value, _ := util.NewUint128FromInt(tx.value)
			gasLimit, _ := util.NewUint128FromInt(tx.gasLimit)
			from := addrs[0]
			txDeploy, err := core.NewTransaction(neb.BlockChain().ChainID(), from, from, value, uint64(k+1), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
			assert.Nil(t, err)
			assert.Nil(t, manager.SignTransaction(from, txDeploy))
			txs = append(txs, txDeploy)

			contractAddr, err := txDeploy.GenerateContractAddress()
			assert.Nil(t, err)
			contracts = append(contracts, contractAddr)
		}
		//package contract txs
		mintBlock(t, neb, txs)

		for _, v := range contracts {
			_, err := neb.BlockChain().TailBlock().CheckContract(v)
			assert.Nil(t, err)
		}

		contractB := contracts[1]
		contractC := contracts[2]
		callPayload, _ := core.NewCallPayload(call.function, fmt.Sprintf("[\"%s\", \"%s\", 1]", contractB.String(), contractC.String()))
		payloadCall, _ := callPayload.ToBytes()
		value, _ := util.NewUint128FromInt(6)
		gasLimit, _ := util.NewUint128FromInt(200000)
		from := neb.BlockChain().TailBlock().Transactions()[0].From()
		proxyContractAddress := contracts[0]
		txCall, err := core.NewTransaction(neb.BlockChain().ChainID(), from, proxyContractAddress, value,
			uint64(len(contracts)+1), core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
		assert.Nil(t, err)
		assert.Nil(t, manager.SignTransaction(txCall.From(), txCall))

		// package call tx
		mintBlock(t, neb, []*core.Transaction{txCall})

		tail := neb.BlockChain().TailBlock()
		events, err := tail.FetchEvents(txCall.Hash())
		assert.Nil(t, err)
		for _, event := range events {
			t.Logf("call tpoic: %s, data:%s", event.Topic, event.Data)
		}
		contractAddrA := contracts[0]
		accountAAcc, err := tail.GetAccount(contractAddrA.Bytes())
		assert.Nil(t, err)
		fmt.Printf("account :%v\n", accountAAcc)
		assert.Equal(t, call.exceptArgs[0], accountAAcc.Balance().String())

		contractAddrB := contracts[1]
		accountBAcc, err := tail.GetAccount(contractAddrB.Bytes())
		assert.Nil(t, err)
		fmt.Printf("accountB :%v\n", accountBAcc)
		assert.Equal(t, call.exceptArgs[1], accountBAcc.Balance().String())

		contractAddrC := contracts[2]
		accountAccC, err := tail.GetAccount(contractAddrC.Bytes())
		assert.Nil(t, err)
		fmt.Printf("accountC :%v\n", accountAccC)
		assert.Equal(t, call.exceptArgs[2], accountAccC.Balance().String())

		aUser, err := tail.GetAccount(from.Bytes())
		assert.Equal(t, call.exceptArgs[3], aUser.Balance().String())
		fmt.Printf("from:%v\n", aUser)

	}
}

func TestInnerTransactionsInMulitithCoroutine(t *testing.T) {
	core.NebCompatibility = core.NewCompatibilityLocal()
	core.PackedParallelNum = 2
	tests := []struct {
		name       string
		txs        []deployTx
		calls      []call
		paralleNum uint32
	}{
		{
			"inner transaction",
			[]deployTx{
				{
					0, 200000,
					contract{
						"./test/inner_call_tests/test_inner_transaction.js",
						"js",
						"",
					},
				},
				{
					0, 200000,
					contract{
						"./test/inner_call_tests/bank_vault_inner_contract.js",
						"js",
						"",
					},
				},
				{
					0, 200000,
					contract{
						"./test/inner_call_tests/bank_vault_final_contract.js",
						"js",
						"",
					},
				},
			},
			[]call{
				call{
					"save",
					"[1]",
					[]string{"1", "3", "2", "4999999999999833224999994", "5000004280820166775000000"},
				},
			},
			4,
		},
	}
	neb := mockNeb(t)
	manager := neb.AccountManager()

	addrs := unlockAccount(t, neb)

	// mint height 2, inner contracts >= 3
	mintBlock(t, neb, nil)

	tt := tests[0]
	for _, call := range tt.calls {

		txs := []*core.Transaction{}
		var contractsAddr [][]string
		// var nonce uint64
		// t.Run(tt.name, func(t *testing.T) {
		for i := uint32(0); i < tt.paralleNum; i++ {
			// contractsAddr[i] = []
			var tmp []string
			nonce := uint64(0)
			for _, v := range tt.txs {
				data, err := ioutil.ReadFile(v.contract.contractPath)
				assert.Nil(t, err, "contract path read error")
				source := string(data)
				sourceType := "js"
				argsDeploy := ""
				deploy, _ := core.NewDeployPayload(source, sourceType, argsDeploy)
				payloadDeploy, _ := deploy.ToBytes()

				value, _ := util.NewUint128FromInt(v.value)
				gasLimit, _ := util.NewUint128FromInt(v.gasLimit)
				// nonce := uint32(k+1)*（i+1） + 1
				nonce++
				idx := int(i) % len(addrs)
				from := addrs[idx]
				txDeploy, err := core.NewTransaction(neb.BlockChain().ChainID(), from, from, value, nonce, core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
				assert.Nil(t, err)
				assert.Nil(t, manager.SignTransaction(from, txDeploy))
				txs = append(txs, txDeploy)

				contractAddr, err := txDeploy.GenerateContractAddress()
				assert.Nil(t, err)
				fmt.Printf("index;%v", i)
				// if contractsAddr[i] = nil {
				// 	contractsAddr[i] = ""
				// }
				// contractsAddr[i] = append(contractsAddr[i], contractAddr.String())
				tmp = append(tmp, contractAddr.String())
				// contractsAddr[i]
			}
			contractsAddr = append(contractsAddr, tmp)
		}
		// })

		// mint block for contracts
		mintBlock(t, neb, txs)

		txCalls := []*core.Transaction{}
		for i := uint32(0); i < tt.paralleNum; i++ {

			calleeContract := contractsAddr[i][1]
			calleeContractAddr, _ := core.AddressParse(calleeContract)
			_, err := neb.BlockChain().TailBlock().CheckContract(calleeContractAddr)
			assert.Nil(t, err)
			callToContract := contractsAddr[i][2]
			callToContractAddr, _ := core.AddressParse(callToContract)
			_, err = neb.BlockChain().TailBlock().CheckContract(callToContractAddr)
			assert.Nil(t, err)
			callPayload, _ := core.NewCallPayload(call.function, fmt.Sprintf("[\"%s\", \"%s\", 1]", calleeContract, callToContract))
			payloadCall, _ := callPayload.ToBytes()
			value, _ := util.NewUint128FromInt(6)
			gasLimit, _ := util.NewUint128FromInt(200000)

			proxyContractAddress, err := core.AddressParse(contractsAddr[i][0])
			idx := int(i) % len(addrs)
			from := addrs[idx]
			account, err := neb.BlockChain().TailBlock().GetAccount(from.Bytes())
			assert.Nil(t, err)
			txCall, err := core.NewTransaction(neb.BlockChain().ChainID(), from, proxyContractAddress, value,
				account.Nonce()+1, core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
			assert.Nil(t, err)
			assert.Nil(t, manager.SignTransaction(from, txCall))
			// fmt.Printf("++tx:%v\n", txCall.JSONString())
			txCalls = append(txCalls, txCall)
			// txCalls[i] = txCall
		}

		mintBlock(t, neb, txCalls)

		// check
		tail := neb.BlockChain().TailBlock()
		for i := uint32(0); i < tt.paralleNum; i++ {
			fmt.Printf("xxxxxxxxxxxx i:%v", i)
			events, err := tail.FetchEvents(txCalls[i].Hash())
			assert.Nil(t, err)
			for _, event := range events {

				fmt.Println("==============", event.Data)
			}
			contractAddrA, err := core.AddressParse(contractsAddr[i][0])
			accountAAcc, err := tail.GetAccount(contractAddrA.Bytes())
			assert.Nil(t, err)
			fmt.Printf("account :%v\n", accountAAcc)
			assert.Equal(t, call.exceptArgs[0], accountAAcc.Balance().String())

			contractAddrB, err := core.AddressParse(contractsAddr[i][1])
			accountBAcc, err := tail.GetAccount(contractAddrB.Bytes())
			assert.Nil(t, err)
			fmt.Printf("accountB :%v\n", accountBAcc)
			assert.Equal(t, call.exceptArgs[1], accountBAcc.Balance().String())

			contractAddrC, err := core.AddressParse(contractsAddr[i][2])
			accountAccC, err := tail.GetAccount(contractAddrC.Bytes())
			assert.Nil(t, err)
			fmt.Printf("accountC :%v\n", accountAccC)
			assert.Equal(t, call.exceptArgs[2], accountAccC.Balance().String())

			aUser, err := tail.GetAccount(addrs[0].Bytes())
			// assert.Equal(t, call.exceptArgs[3], aUser.Balance().String())
			fmt.Printf("a:%v\n", aUser)
			cUser, err := tail.GetAccount(addrs[2].Bytes())
			fmt.Printf("c:%v\n", cUser)
			// assert.Equal(t, call.exceptArgs[4], cUser.Balance().String())

			// cUser, err := tail.GetAccount(c.Bytes())
			// fmt.Printf("c:%v\n", cUser)

			// dUser, err := tail.GetAccount(d.Bytes())
			// fmt.Printf("d:%v\n", dUser)
			// assert.Equal(t, txEvent.Status, 1)
		}

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
					"./test/inner_call_tests/test_inner_transaction.js",
					"js",
					"",
				},
				contract{
					"./test/inner_call_tests/bank_vault_inner_contract.js",
					"js",
					"",
				},
				contract{
					"./test/inner_call_tests/bank_vault_final_contract.js",
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

	neb := mockNeb(t)
	manager, err := account.NewManager(neb)
	assert.Nil(t, err)

	addrs := unlockAccount(t, neb)
	from := addrs[0]

	// mint height 2, inner contracts >= 3
	mintBlock(t, neb, nil)

	for _, tt := range tests {

		txs := []*core.Transaction{}
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
			txDeploy, err := core.NewTransaction(neb.BlockChain().ChainID(), from, from, value, uint64(k+1), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
			assert.Nil(t, err)
			assert.Nil(t, manager.SignTransaction(from, txDeploy))
			txs = append(txs, txDeploy)

			contractAddr, err := txDeploy.GenerateContractAddress()
			assert.Nil(t, err)
			contractsAddr = append(contractsAddr, contractAddr.String())
		}
		// })

		// mint for contract deploy
		mintBlock(t, neb, txs)

		for _, v := range contractsAddr {
			contract, err := core.AddressParse(v)
			assert.Nil(t, err)
			_, err = neb.BlockChain().TailBlock().CheckContract(contract)
			assert.Nil(t, err)
		}

		calleeContract := contractsAddr[0]
		callToContract := contractsAddr[2]
		callPayload, _ := core.NewCallPayload(tt.call.function, fmt.Sprintf("[\"%s\", \"%s\", 1]", calleeContract, callToContract))
		payloadCall, _ := callPayload.ToBytes()

		value, _ := util.NewUint128FromInt(6)
		gasLimit, _ := util.NewUint128FromInt(2000000)

		proxyContractAddress, err := core.AddressParse(contractsAddr[0])
		txCall, err := core.NewTransaction(neb.BlockChain().ChainID(), from, proxyContractAddress, value,
			uint64(len(contractsAddr)+1), core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
		assert.Nil(t, err)
		assert.Nil(t, manager.SignTransaction(from, txCall))

		// mint for inner call
		mintBlock(t, neb, []*core.Transaction{txCall})

		// check
		tail := neb.BlockChain().TailBlock()

		events, err := tail.FetchEvents(txCall.Hash())
		assert.Nil(t, err)
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
					"./test/inner_call_tests/test_inner_transaction.js",
					"js",
					"",
				},
				contract{
					"./test/inner_call_tests/bank_vault_inner_contract.js",
					"js",
					"",
				},
				contract{
					"./test/inner_call_tests/bank_vault_final_contract.js",
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

	neb := mockNeb(t)
	manager, err := account.NewManager(neb)
	assert.Nil(t, err)

	addrs := unlockAccount(t, neb)
	from := addrs[0]

	// mint height 2, inner contracts >= 3
	mintBlock(t, neb, nil)

	for _, tt := range tests {

		txs := []*core.Transaction{}
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
			txDeploy, err := core.NewTransaction(neb.BlockChain().ChainID(), from, from, value, uint64(k+1), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
			assert.Nil(t, err)
			assert.Nil(t, manager.SignTransaction(from, txDeploy))
			txs = append(txs, txDeploy)

			contractAddr, err := txDeploy.GenerateContractAddress()
			assert.Nil(t, err)
			contractsAddr = append(contractsAddr, contractAddr.String())
		}
		// })

		// mint for contract deploy
		mintBlock(t, neb, txs)

		for _, v := range contractsAddr {
			contract, err := core.AddressParse(v)
			assert.Nil(t, err)
			_, err = neb.BlockChain().TailBlock().CheckContract(contract)
			assert.Nil(t, err)
		}

		proxyContractAddress, _ := core.AddressParse(contractsAddr[0])
		calleeContract := contractsAddr[1]
		callToContract := contractsAddr[2]

		callPayload, _ := core.NewCallPayload(tt.call.function, fmt.Sprintf("[\"%s\", \"%s\", 1]", calleeContract, callToContract))
		payloadCall, _ := callPayload.ToBytes()

		// call tx
		for i := 0; i < len(tt.gasArr); i++ {

			value, _ := util.NewUint128FromInt(6)
			gasLimit, _ := util.NewUint128FromInt(int64(tt.gasArr[i]))
			txCall, err := core.NewTransaction(neb.BlockChain().ChainID(), from, proxyContractAddress, value,
				uint64(len(contractsAddr)+1), core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
			assert.Nil(t, err)
			assert.Nil(t, manager.SignTransaction(from, txCall))

			// mint for contract call
			mintBlock(t, neb, []*core.Transaction{txCall})

			tail := neb.BlockChain().TailBlock()
			events, err := tail.WorldState().FetchEvents(txCall.Hash())
			assert.Nil(t, err)
			for _, tx := range tail.Transactions() {
				// fmt.Println("=========>Tx", tx.Hash(), tx.GasLimit(), txCall.Hash())
				assert.Equal(t, tx.Hash(), txCall.Hash())
			}
			for _, event := range events {
				var jEvent SysEvent
				if err := json.Unmarshal([]byte(event.Data), &jEvent); err == nil {
					if jEvent.Hash != "" {
						assert.Equal(t, tt.gasExpectedErr[i], jEvent.Err)
						assert.Equal(t, tt.gasExpectedResult[i], jEvent.Result)
					}
				}
			}

		}
	}
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
					"./test/inner_call_tests/test_inner_transaction.js",
					"js",
					"",
				},
				contract{
					"./test/inner_call_tests/bank_vault_inner_contract.js",
					"js",
					"",
				},
				contract{
					"./test/inner_call_tests/bank_vault_final_contract.js",
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

	neb := mockNeb(t)
	manager, err := account.NewManager(neb)
	assert.Nil(t, err)

	addrs := unlockAccount(t, neb)
	from := addrs[0]

	// mint height 2, inner contracts >= 3
	mintBlock(t, neb, nil)

	for _, tt := range tests {
		for i := 0; i < len(tt.memArr); i++ {

			txs := []*core.Transaction{}
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
				txDeploy, err := core.NewTransaction(neb.BlockChain().ChainID(), from, from, value, uint64(k+1), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
				assert.Nil(t, err)
				assert.Nil(t, manager.SignTransaction(from, txDeploy))
				txs = append(txs, txDeploy)

				contractAddr, err := txDeploy.GenerateContractAddress()
				assert.Nil(t, err)
				contractsAddr = append(contractsAddr, contractAddr.String())
			}

			// mint for contract deploy
			mintBlock(t, neb, txs)

			for _, v := range contractsAddr {
				contract, err := core.AddressParse(v)
				assert.Nil(t, err)
				_, err = neb.BlockChain().TailBlock().CheckContract(contract)
				assert.Nil(t, err)
			}

			calleeContract := contractsAddr[1]
			callToContract := contractsAddr[2]
			callPayload, _ := core.NewCallPayload(tt.call.function, fmt.Sprintf("[\"%s\", \"%s\", \"%d\"]", calleeContract, callToContract, tt.memArr[i]))
			payloadCall, _ := callPayload.ToBytes()

			value, _ := util.NewUint128FromInt(6)
			gasLimit, _ := util.NewUint128FromInt(5000000)
			proxyContractAddress, err := core.AddressParse(contractsAddr[0])
			txCall, err := core.NewTransaction(neb.BlockChain().ChainID(), from, proxyContractAddress, value,
				uint64(len(contractsAddr)+1), core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
			assert.Nil(t, err)
			assert.Nil(t, manager.SignTransaction(from, txCall))

			// mint for contract call
			mintBlock(t, neb, []*core.Transaction{txCall})

			tail := neb.BlockChain().TailBlock()
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
					"./test/inner_call_tests/test_inner_transaction.js",
					"js",
					"",
				},
				contract{
					"./test/inner_call_tests/bank_vault_inner_contract.js",
					"js",
					"",
				},
				contract{
					"./test/inner_call_tests/bank_vault_final_contract.js",
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
				"Call: Inner Contract: saveErr in bank_vault_inner_contract",
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
			manager, err := account.NewManager(neb)
			assert.Nil(t, err)

			addrs := unlockAccount(t, neb)
			from := addrs[0]

			// mint height 2, inner contracts >= 3
			mintBlock(t, neb, nil)

			txs := []*core.Transaction{}
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
				txDeploy, err := core.NewTransaction(neb.BlockChain().ChainID(), from, from, value, uint64(k+1), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
				assert.Nil(t, err)
				assert.Nil(t, manager.SignTransaction(from, txDeploy))
				txs = append(txs, txDeploy)

				contractAddr, err := txDeploy.GenerateContractAddress()
				assert.Nil(t, err)
				contractsAddr = append(contractsAddr, contractAddr.String())
			}

			// mint for deploy contracts
			mintBlock(t, neb, txs)

			for _, v := range contractsAddr {
				contract, err := core.AddressParse(v)
				assert.Nil(t, err)
				_, err = neb.BlockChain().TailBlock().CheckContract(contract)
				assert.Nil(t, err)
			}

			calleeContract := contractsAddr[1]
			callToContract := contractsAddr[2]
			callPayload, _ := core.NewCallPayload(tt.call.function, fmt.Sprintf("[\"%s\", \"%s\", \"%d\"]", calleeContract, callToContract, tt.errFlagArr[i]))
			payloadCall, _ := callPayload.ToBytes()

			value, _ := util.NewUint128FromInt(6)
			gasLimit, _ := util.NewUint128FromInt(1000000)
			proxyContractAddress, err := core.AddressParse(contractsAddr[0])
			txCall, err := core.NewTransaction(neb.BlockChain().ChainID(), from, proxyContractAddress, value,
				uint64(len(contractsAddr)+1), core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
			assert.Nil(t, err)
			assert.Nil(t, manager.SignTransaction(from, txCall))

			// mint for contract call
			mintBlock(t, neb, []*core.Transaction{txCall})

			tail := neb.BlockChain().TailBlock()
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

			aUser, err := tail.GetAccount(addrs[0].Bytes())
			// assert.Equal(t, tt.expectedAccount[i][3], aUser.Balance().String())
			fmt.Printf("aI:%v\n", aUser)
			bUser, err := tail.GetAccount(addrs[1].Bytes())
			assert.Equal(t, tt.expectedAccount[i][4], bUser.Balance().String())

			cUser, err := tail.GetAccount(addrs[2].Bytes())
			fmt.Printf("cI:%v\n", cUser)
			// assert.Equal(t, tt.expectedAccount[i][5], cUser.Balance().String())

			// fmt.Printf("b:%v\n", bUser)
			// assert.Equal(t, tt.expectedAccount[i][4], cUser.Balance().String())
			dUser, err := tail.GetAccount(addrs[3].Bytes())
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
					"./test/inner_call_tests/test_inner_transaction.js",
					"js",
					"",
				},
				contract{
					"./test/inner_call_tests/bank_vault_inner_contract.js",
					"js",
					"",
				},
				contract{
					"./test/inner_call_tests/bank_vault_final_contract.js",
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
				"Call: Inner Contract: uint128: overflow",
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
			manager, err := account.NewManager(neb)
			assert.Nil(t, err)

			addrs := unlockAccount(t, neb)
			from := addrs[0]

			// mint height 2, inner contracts >= 3
			mintBlock(t, neb, nil)

			txs := []*core.Transaction{}
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
				txDeploy, err := core.NewTransaction(neb.BlockChain().ChainID(), from, from, value, uint64(k+1), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
				assert.Nil(t, err)
				assert.Nil(t, manager.SignTransaction(from, txDeploy))
				txs = append(txs, txDeploy)

				contractAddr, err := txDeploy.GenerateContractAddress()
				assert.Nil(t, err)
				contractsAddr = append(contractsAddr, contractAddr.String())
			}

			// mint block for contract deploy
			mintBlock(t, neb, txs)

			for _, v := range contractsAddr {
				contract, err := core.AddressParse(v)
				assert.Nil(t, err)
				_, err = neb.BlockChain().TailBlock().CheckContract(contract)
				assert.Nil(t, err)
			}

			calleeContract := contractsAddr[1]
			callToContract := contractsAddr[2]
			callPayload, _ := core.NewCallPayload(tt.call.function, fmt.Sprintf("[\"%s\", \"%s\", \"%s\"]", calleeContract, callToContract, tt.errValueArr[i]))
			payloadCall, _ := callPayload.ToBytes()

			value, _ := util.NewUint128FromInt(6)
			gasLimit, _ := util.NewUint128FromInt(1000000)
			proxyContractAddress, err := core.AddressParse(contractsAddr[0])
			txCall, err := core.NewTransaction(neb.BlockChain().ChainID(), from, proxyContractAddress, value,
				uint64(len(contractsAddr)+1), core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
			assert.Nil(t, err)
			assert.Nil(t, manager.SignTransaction(from, txCall))

			// mint for contract call
			mintBlock(t, neb, []*core.Transaction{txCall})

			tail := neb.BlockChain().TailBlock()
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

			aUser, err := tail.GetAccount(addrs[0].Bytes())
			// assert.Equal(t, tt.expectedAccount[i][3], aUser.Balance().String())
			fmt.Printf("aI:%v\n", aUser)
			bUser, err := tail.GetAccount(addrs[1].Bytes())
			assert.Equal(t, tt.expectedAccount[i][4], bUser.Balance().String())

			cUser, err := tail.GetAccount(addrs[2].Bytes())
			fmt.Printf("cI:%v\n", cUser)
			// assert.Equal(t, tt.expectedAccount[i][5], cUser.Balance().String())

			// fmt.Printf("b:%v\n", bUser)
			// assert.Equal(t, tt.expectedAccount[i][4], cUser.Balance().String())
			dUser, err := tail.GetAccount(addrs[3].Bytes())
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
					"./test/inner_call_tests/test_inner_transaction.js",
					"js",
					"",
				},
				contract{
					"./test/inner_call_tests/bank_vault_inner_contract.js",
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
			manager, err := account.NewManager(neb)
			assert.Nil(t, err)

			addrs := unlockAccount(t, neb)
			from := addrs[0]

			// mint height 2, inner contracts >= 3
			mintBlock(t, neb, nil)

			txs := []*core.Transaction{}
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
				txDeploy, err := core.NewTransaction(neb.BlockChain().ChainID(), from, from, value, uint64(k+1), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
				assert.Nil(t, err)
				assert.Nil(t, manager.SignTransaction(from, txDeploy))
				txs = append(txs, txDeploy)

				contractAddr, err := txDeploy.GenerateContractAddress()
				assert.Nil(t, err)
				contractsAddr = append(contractsAddr, contractAddr.String())
			}

			// mint for contract deploy
			mintBlock(t, neb, txs)

			for _, v := range contractsAddr {
				contract, err := core.AddressParse(v)
				assert.Nil(t, err)
				_, err = neb.BlockChain().TailBlock().CheckContract(contract)
				assert.Nil(t, err)
			}

			calleeContract := "123456789"
			callToContract := "123456789"
			callPayload, _ := core.NewCallPayload(tt.calls[i].function, fmt.Sprintf("[\"%s\", \"%s\"]", calleeContract, callToContract))
			payloadCall, _ := callPayload.ToBytes()

			value, _ := util.NewUint128FromInt(6)
			gasLimit, _ := util.NewUint128FromInt(1000000)
			proxyContractAddress, err := core.AddressParse(contractsAddr[0])
			txCall, err := core.NewTransaction(neb.BlockChain().ChainID(), from, proxyContractAddress, value,
				uint64(len(contractsAddr)+1), core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
			assert.Nil(t, err)
			assert.Nil(t, manager.SignTransaction(from, txCall))

			// mint for contract call
			mintBlock(t, neb, []*core.Transaction{txCall})

			tail := neb.BlockChain().TailBlock()
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
					"./test/inner_call_tests/test_inner_transaction.js",
					"js",
					"",
				},
				contract{
					"./test/inner_call_tests/bank_vault_inner_contract.js",
					"js",
					"",
				},
				contract{
					"./test/inner_call_tests/bank_vault_final_contract.js",
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
		manager, err := account.NewManager(neb)
		assert.Nil(t, err)

		addrs := unlockAccount(t, neb)
		from := addrs[0]

		// mint height 2, inner contracts >= 3
		mintBlock(t, neb, nil)

		txs := []*core.Transaction{}
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
			txDeploy, err := core.NewTransaction(neb.BlockChain().ChainID(), from, from, value, uint64(k+1), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
			assert.Nil(t, err)
			assert.Nil(t, manager.SignTransaction(from, txDeploy))
			txs = append(txs, txDeploy)

			contractAddr, err := txDeploy.GenerateContractAddress()
			assert.Nil(t, err)
			contractsAddr = append(contractsAddr, contractAddr.String())
		}

		// mint for contract deploy
		mintBlock(t, neb, txs)

		for _, v := range contractsAddr {
			contract, err := core.AddressParse(v)
			assert.Nil(t, err)
			_, err = neb.BlockChain().TailBlock().CheckContract(contract)
			assert.Nil(t, err)
		}

		calleeContract := contractsAddr[1]
		callToContract := contractsAddr[2]
		callPayload, _ := core.NewCallPayload(tt.call.function, fmt.Sprintf("[\"%s\", \"%s\"]", calleeContract, callToContract))
		payloadCall, _ := callPayload.ToBytes()

		value, _ := util.NewUint128FromInt(6)
		gasLimit, _ := util.NewUint128FromInt(int64(100000))
		proxyContractAddress, err := core.AddressParse(contractsAddr[0])
		txCall, err := core.NewTransaction(neb.BlockChain().ChainID(), from, proxyContractAddress, value,
			uint64(len(contractsAddr)+1), core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
		assert.Nil(t, err)
		assert.Nil(t, manager.SignTransaction(from, txCall))

		// mint for contract call
		mintBlock(t, neb, []*core.Transaction{txCall})

		tail := neb.BlockChain().TailBlock()
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
					"./test/inner_call_tests/test_inner_transaction.js",
					"js",
					"",
				},
				contract{
					"./test/inner_call_tests/bank_vault_inner_contract.js",
					"js",
					"",
				},
				contract{
					"./test/inner_call_tests/bank_vault_final_contract.js",
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
			manager, err := account.NewManager(neb)
			assert.Nil(t, err)

			addrs := unlockAccount(t, neb)
			from := addrs[0]

			// mint height 2, inner contracts >= 3
			mintBlock(t, neb, nil)

			txs := []*core.Transaction{}
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
				txDeploy, err := core.NewTransaction(neb.BlockChain().ChainID(), from, from, value, uint64(k+1), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
				assert.Nil(t, err)
				assert.Nil(t, manager.SignTransaction(from, txDeploy))
				txs = append(txs, txDeploy)

				contractAddr, err := txDeploy.GenerateContractAddress()
				assert.Nil(t, err)
				contractsAddr = append(contractsAddr, contractAddr.String())
			}

			// mint for contract deploy
			mintBlock(t, neb, txs)

			for _, v := range contractsAddr {
				contract, err := core.AddressParse(v)
				assert.Nil(t, err)
				_, err = neb.BlockChain().TailBlock().CheckContract(contract)
				assert.Nil(t, err)
			}

			calleeContract := contractsAddr[1]
			callToContract := contractsAddr[2]
			callPayload, _ := core.NewCallPayload(tt.call.function, fmt.Sprintf("[\"%s\", \"%s\", \"%d\"]", calleeContract, callToContract, tt.errFlagArr[i]))
			payloadCall, _ := callPayload.ToBytes()

			value, _ := util.NewUint128FromInt(6)
			gasLimit, _ := util.NewUint128FromInt(1000000)
			proxyContractAddress, err := core.AddressParse(contractsAddr[0])
			txCall, err := core.NewTransaction(neb.BlockChain().ChainID(), from, proxyContractAddress, value,
				uint64(len(contractsAddr)+1), core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
			assert.Nil(t, err)
			assert.Nil(t, manager.SignTransaction(from, txCall))

			// mint for contract call
			mintBlock(t, neb, []*core.Transaction{txCall})

			tail := neb.BlockChain().TailBlock()
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
		manager, err := account.NewManager(neb)
		assert.Nil(t, err)

		addrs := unlockAccount(t, neb)
		from := addrs[0]

		// mint height 2, inner contracts >= 3
		mintBlock(t, neb, nil)

		txs := []*core.Transaction{}
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
			txDeploy, err := core.NewTransaction(neb.BlockChain().ChainID(), from, from, value, uint64(k+1), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
			assert.Nil(t, err)
			assert.Nil(t, manager.SignTransaction(from, txDeploy))
			txs = append(txs, txDeploy)

			contractAddr, err := txDeploy.GenerateContractAddress()
			assert.Nil(t, err)
			contractsAddr = append(contractsAddr, contractAddr.String())
		}
		// })

		// mint for contract deploy
		mintBlock(t, neb, txs)

		for _, v := range contractsAddr {
			contract, err := core.AddressParse(v)
			assert.Nil(t, err)
			_, err = neb.BlockChain().TailBlock().CheckContract(contract)
			assert.Nil(t, err)
		}

		callPayload, _ := core.NewCallPayload(call.function, fmt.Sprintf("[\"%s\"]", contractsAddr[0]))
		payloadCall, _ := callPayload.ToBytes()

		value, _ := util.NewUint128FromInt(0)
		gasLimit, _ := util.NewUint128FromInt(200000)

		tail := neb.BlockChain().TailBlock()
		aUser, err := tail.GetAccount(from.Bytes())
		balBefore := aUser.Balance()

		proxyContractAddress, err := core.AddressParse(contractsAddr[1])
		txCall, err := core.NewTransaction(neb.BlockChain().ChainID(), from, proxyContractAddress, value,
			uint64(len(contractsAddr)+1), core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
		assert.Nil(t, err)
		assert.Nil(t, manager.SignTransaction(from, txCall))

		// mint for contract call
		mintBlock(t, neb, []*core.Transaction{txCall})

		// // check
		tail = neb.BlockChain().TailBlock()
		events, err := tail.FetchEvents(txCall.Hash())
		assert.Nil(t, err)
		for _, event := range events {
			fmt.Println("==============", event.Data)
		}

		aUser, err = tail.GetAccount(from.Bytes())
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
	manager, err := account.NewManager(neb)
	assert.Nil(t, err)

	addrs := unlockAccount(t, neb)
	from := addrs[0]

	// mint height 2, inner contracts >= 3
	mintBlock(t, neb, nil)

	txs := []*core.Transaction{}
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
	txDeploy, err := core.NewTransaction(neb.BlockChain().ChainID(), from, from, value, uint64(1), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
	assert.Nil(t, err)
	assert.Nil(t, manager.SignTransaction(from, txDeploy))
	txs = append(txs, txDeploy)

	contractAddr, err := txDeploy.GenerateContractAddress()
	assert.Nil(t, err)
	contractsAddr = append(contractsAddr, contractAddr.String())

	// mint for contract deploy
	mintBlock(t, neb, txs)

	data, err = ioutil.ReadFile("./test/inner_call_tests/caller.js")
	assert.Nil(t, err, "contract path read error")
	source = string(data)
	sourceType = "js"
	argsDeploy = ""
	deploy, _ = core.NewDeployPayload(source, sourceType, argsDeploy)
	payloadDeploy, _ = deploy.ToBytes()

	txDeploy, err = core.NewTransaction(neb.BlockChain().ChainID(), from, from, value, uint64(2), core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
	assert.Nil(t, err)
	assert.Nil(t, manager.SignTransaction(from, txDeploy))

	contractAddr, err = txDeploy.GenerateContractAddress()
	assert.Nil(t, err)
	contractsAddr = append(contractsAddr, contractAddr.String())

	// mint for contract deploy
	mintBlock(t, neb, []*core.Transaction{txDeploy})

	for _, v := range contractsAddr {
		contract, err := core.AddressParse(v)
		assert.Nil(t, err)
		_, err = neb.BlockChain().TailBlock().CheckContract(contract)
		assert.Nil(t, err)
	}

	for _, call := range tt.calls {
		callPayload, _ := core.NewCallPayload(call.function, fmt.Sprintf(call.args, contractsAddr[0]))
		payloadCall, _ := callPayload.ToBytes()

		proxyContractAddress, err := core.AddressParse(contractsAddr[1])
		txCall, err := core.NewTransaction(neb.BlockChain().ChainID(), from, proxyContractAddress, value,
			uint64(len(contractsAddr)+1), core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
		assert.Nil(t, err)
		assert.Nil(t, manager.SignTransaction(from, txCall))

		// mint for contract call
		mintBlock(t, neb, []*core.Transaction{txCall})

		tail := neb.BlockChain().TailBlock()
		events, err := tail.WorldState().FetchEvents(txCall.Hash())
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
