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
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	corepb "github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/crypto"
	"github.com/nebulasio/go-nebulas/crypto/hash"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/stretchr/testify/assert"
)

func mockNormalTransaction(chainID uint32, nonce uint64) *Transaction {
	// payload, _ := NewBinaryPayload(nil).ToBytes()
	return mockTransaction(chainID, nonce, TxPayloadBinaryType, nil)
}

func mockDeployTransaction(chainID uint32, nonce uint64) *Transaction {
	source := `
	"use strict";var StandardToken=function(){LocalContractStorage.defineProperties(this,{name:null,symbol:null,_totalSupply:null,totalIssued:null});LocalContractStorage.defineMapProperty(this,"balances")};StandardToken.prototype={init:function(name,symbol,totalSupply){this.name=name;this.symbol=symbol;this._totalSupply=totalSupply;this.totalIssued=0},totalSupply:function(){return this._totalSupply},balanceOf:function(owner){return this.balances.get(owner)||0},transfer:function(to,value){var balance=this.balanceOf(msg.sender);if(balance<value){return false}var finalBalance=balance-value;this.balances.set(msg.sender,finalBalance);this.balances.set(to,this.balanceOf(to)+value);return true},pay:function(msg,amount){if(this.totalIssued+amount>this._totalSupply){throw new Error("too much amount, exceed totalSupply")}this.balances.set(msg.sender,this.balanceOf(msg.sender)+amount);this.totalIssued+=amount}};module.exports=StandardToken;
	`
	sourceType := "js"
	args := `["NebulasToken", "NAS", 1000000000]`
	payloadObj, _ := NewDeployPayload(source, sourceType, args)
	payload, _ := payloadObj.ToBytes()
	return mockTransaction(chainID, nonce, TxPayloadDeployType, payload)
}

func mockCallTransaction(chainID uint32, nonce uint64, function, args string) *Transaction {
	callpayload, _ := NewCallPayload(function, args)
	payload, _ := callpayload.ToBytes()
	return mockTransaction(chainID, nonce, TxPayloadCallType, payload)
}

func mockTransaction(chainID uint32, nonce uint64, payloadType string, payload []byte) *Transaction {
	from := mockAddress()
	to := mockAddress()
	if payloadType == TxPayloadDeployType {
		to = from
	}
	tx, _ := NewTransaction(chainID, from, to, util.NewUint128(), nonce, payloadType, payload, TransactionGasPrice, TransactionMaxGas)
	return tx
}

func TestTransaction(t *testing.T) {
	type fields struct {
		hash      byteutils.Hash
		from      *Address
		to        *Address
		value     *util.Uint128
		nonce     uint64
		timestamp int64
		alg       uint8
		data      *corepb.Data
		gasPrice  *util.Uint128
		gasLimit  *util.Uint128
	}
	gasPrice, _ := util.NewUint128FromInt(1)
	gasLimit, _ := util.NewUint128FromInt(1)
	tests := []struct {
		name   string
		fields fields
	}{
		{
			"full struct",
			fields(fields{
				[]byte("123455"),
				mockAddress(),
				mockAddress(),
				util.NewUint128(),
				456,
				time.Now().Unix(),
				uint8(keystore.SECP256K1),
				&corepb.Data{Type: TxPayloadBinaryType, Payload: []byte("hwllo")},
				gasPrice,
				gasLimit,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &Transaction{
				hash:      tt.fields.hash,
				from:      tt.fields.from,
				to:        tt.fields.to,
				value:     tt.fields.value,
				nonce:     tt.fields.nonce,
				timestamp: tt.fields.timestamp,
				alg:       keystore.Algorithm(tt.fields.alg),
				data:      tt.fields.data,
				gasPrice:  tt.fields.gasPrice,
				gasLimit:  tt.fields.gasLimit,
			}
			msg, _ := tx.ToProto()
			ir, _ := proto.Marshal(msg)
			ntx := new(Transaction)
			nMsg := new(corepb.Transaction)
			proto.Unmarshal(ir, nMsg)
			ntx.FromProto(nMsg)
			ntx.timestamp = tx.timestamp
			if !reflect.DeepEqual(tx, ntx) {
				t.Errorf("Transaction.Serialize() = %v, want %v", *tx, *ntx)
			}
		})
	}
}

func TestTransaction_VerifyIntegrity(t *testing.T) {
	testCount := 3
	type testTx struct {
		name   string
		tx     *Transaction
		signer keystore.Signature
		count  int
	}

	tests := []testTx{}
	ks := keystore.DefaultKS

	for index := 0; index < testCount; index++ {

		from := mockAddress()
		to := mockAddress()

		key1, _ := ks.GetUnlocked(from.String())
		signature, _ := crypto.NewSignature(keystore.SECP256K1)
		signature.InitSign(key1.(keystore.PrivateKey))

		gasLimit, _ := util.NewUint128FromInt(200000)
		tx, _ := NewTransaction(1, from, to, util.NewUint128(), 10, TxPayloadBinaryType, []byte("datadata"), TransactionGasPrice, gasLimit)

		test := testTx{string(index), tx, signature, 1}
		tests = append(tests, test)
	}
	for _, tt := range tests {
		for index := 0; index < tt.count; index++ {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.tx.Sign(tt.signer)
				if err != nil {
					t.Errorf("Sign() error = %v", err)
					return
				}
				err = tt.tx.VerifyIntegrity(tt.tx.chainID)
				if err != nil {
					t.Errorf("verify failed:%s", err)
					return
				}
			})
		}
	}
}

func TestTransaction_VerifyExecutionDependency(t *testing.T) {

	neb := testNeb(t)
	bc := neb.chain

	a := mockAddress()
	b := mockAddress()
	c := mockAddress()
	e := mockAddress()
	f := mockAddress()

	ks := keystore.DefaultKS

	v, _ := util.NewUint128FromInt(1)
	tx1, _ := NewTransaction(bc.chainID, a, b, v, uint64(1), TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, TransactionMaxGas)
	tx2, _ := NewTransaction(bc.chainID, a, b, v, uint64(2), TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, TransactionMaxGas)
	tx3, _ := NewTransaction(bc.chainID, b, c, v, uint64(1), TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, TransactionMaxGas)
	tx4, _ := NewTransaction(bc.chainID, e, f, v, uint64(1), TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, TransactionMaxGas)

	txs := [4]*Transaction{tx1, tx2, tx3, tx4}
	for _, tx := range txs {
		key, _ := ks.GetUnlocked(tx.from.String())
		signature, _ := crypto.NewSignature(keystore.SECP256K1)
		signature.InitSign(key.(keystore.PrivateKey))
		assert.Nil(t, tx.Sign(signature))
	}

	balance, _ := util.NewUint128FromString("1000000000000000000")

	bc.tailBlock.Begin()
	{
		fromAcc, err := bc.tailBlock.worldState.GetOrCreateUserAccount(tx1.from.address)
		assert.Nil(t, err)
		fromAcc.AddBalance(balance)
	}
	{
		fromAcc, err := bc.tailBlock.worldState.GetOrCreateUserAccount(tx3.from.address)
		assert.Nil(t, err)
		fromAcc.AddBalance(balance)
	}
	{
		fromAcc, err := bc.tailBlock.worldState.GetOrCreateUserAccount(tx4.from.address)
		assert.Nil(t, err)
		fromAcc.AddBalance(balance)
	}

	bc.tailBlock.Commit()

	block, err := bc.NewBlock(bc.tailBlock.header.coinbase)
	assert.Nil(t, err)
	block.Begin()

	//tx1
	txWorldState1, err := block.WorldState().Prepare("1")
	assert.Nil(t, err)
	giveback, executionErr := VerifyExecution(tx1, block, txWorldState1)
	assert.Nil(t, executionErr)
	assert.False(t, giveback)
	dependency1, err := txWorldState1.CheckAndUpdate()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(dependency1))

	acc1, err := block.worldState.GetOrCreateUserAccount(tx1.from.address)
	assert.Equal(t, "999999999999999999", acc1.Balance().String())
	toacc1, err := block.worldState.GetOrCreateUserAccount(tx1.to.address)
	assert.Equal(t, "1000000000000000001", toacc1.Balance().String())

	//tx2
	txWorldState2, err := block.WorldState().Prepare("2")
	txWorldState3, err := block.WorldState().Prepare("3")
	txWorldState4, err := block.WorldState().Prepare("4")

	giveback, executionErr2 := VerifyExecution(tx2, block, txWorldState2)
	assert.Nil(t, executionErr2)
	assert.False(t, giveback)
	dependency2, err := txWorldState2.CheckAndUpdate()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(dependency2))
	assert.Equal(t, "1", dependency2[0])
	acc2, err := block.worldState.GetOrCreateUserAccount(tx2.from.address)
	assert.Equal(t, "999999999999999998", acc2.Balance().String())
	toacc2, err := block.worldState.GetOrCreateUserAccount(tx2.to.address)
	assert.Equal(t, "1000000000000000002", toacc2.Balance().String())

	// tx3
	giveback, executionErr3 := VerifyExecution(tx3, block, txWorldState3)
	assert.NotNil(t, executionErr3)
	assert.True(t, giveback)
	txWorldState3.Close()

	//tx4
	giveback, executionErr4 := VerifyExecution(tx4, block, txWorldState4)
	assert.Nil(t, executionErr4)
	assert.False(t, giveback)
	_, err4 := txWorldState4.CheckAndUpdate()
	assert.Nil(t, err4)

	txWorldState3, err = block.WorldState().Prepare("3")
	assert.Nil(t, err)
	giveback, executionErr3 = VerifyExecution(tx3, block, txWorldState3)
	assert.Nil(t, executionErr3)
	assert.False(t, giveback)
	dependency3, err := txWorldState3.CheckAndUpdate()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(dependency3))
	assert.Equal(t, "2", dependency3[0])

	acc3, err := block.worldState.GetOrCreateUserAccount(tx3.from.address)
	assert.Equal(t, "1000000000000000001", acc3.Balance().String())
	toacc3, err := block.worldState.GetOrCreateUserAccount(tx3.to.address)
	assert.Equal(t, "1", toacc3.Balance().String())

	assert.Nil(t, block.Seal())
}

func TestTransaction_VerifyExecution(t *testing.T) {
	type testTx struct {
		name            string
		tx              *Transaction
		fromBalance     *util.Uint128
		gasUsed         *util.Uint128
		wanted          error
		afterBalance    *util.Uint128
		toBalance       *util.Uint128
		coinbaseBalance *util.Uint128
		status          int
		eventErr        string
		giveback        bool
	}
	tests := []testTx{}

	neb := testNeb(t)
	bc := neb.chain

	// 1NAS = 10^18
	balance, _ := util.NewUint128FromString("1000000000000000000")
	// normal tx
	normalTx := mockNormalTransaction(bc.chainID, 0)
	normalTx.value, _ = util.NewUint128FromInt(1000000)
	gasConsume, err := normalTx.gasPrice.Mul(MinGasCountPerTransaction)
	assert.Nil(t, err)
	afterBalance, err := balance.Sub(gasConsume)
	assert.Nil(t, err)
	afterBalance, err = afterBalance.Sub(normalTx.value)
	coinbaseBalance, err := normalTx.gasPrice.Mul(MinGasCountPerTransaction)
	assert.Nil(t, err)
	tests = append(tests, testTx{
		name:            "normal tx",
		tx:              normalTx,
		fromBalance:     balance,
		gasUsed:         MinGasCountPerTransaction,
		afterBalance:    afterBalance,
		toBalance:       normalTx.value,
		coinbaseBalance: coinbaseBalance,
		wanted:          nil,
		eventErr:        "",
		status:          1,
		giveback:        false,
	})

	// contract deploy tx
	deployTx := mockDeployTransaction(bc.chainID, 0)
	deployTx.value = util.NewUint128()
	gasUsed, _ := util.NewUint128FromInt(21203)
	coinbaseBalance, err = deployTx.gasPrice.Mul(gasUsed)
	assert.Nil(t, err)
	balanceConsume, err := deployTx.gasPrice.Mul(gasUsed)
	assert.Nil(t, err)
	afterBalance, err = balance.Sub(balanceConsume)
	assert.Nil(t, err)
	tests = append(tests, testTx{
		name:            "contract deploy tx",
		tx:              deployTx,
		fromBalance:     balance,
		gasUsed:         gasUsed,
		afterBalance:    afterBalance,
		toBalance:       afterBalance,
		coinbaseBalance: coinbaseBalance,
		wanted:          nil,
		eventErr:        "",
		status:          1,
		giveback:        false,
	})

	// contract call tx
	callTx := mockCallTransaction(bc.chainID, 1, "totalSupply", "")
	callTx.value = util.NewUint128()
	gasUsed, _ = util.NewUint128FromInt(20096)
	coinbaseBalance, err = callTx.gasPrice.Mul(gasUsed)
	assert.Nil(t, err)
	balanceConsume, err = callTx.gasPrice.Mul(gasUsed)
	assert.Nil(t, err)
	afterBalance, err = balance.Sub(balanceConsume)

	tests = append(tests, testTx{
		name:            "contract call tx",
		tx:              callTx,
		fromBalance:     balance,
		gasUsed:         gasUsed,
		afterBalance:    afterBalance,
		toBalance:       callTx.value,
		coinbaseBalance: coinbaseBalance,
		wanted:          nil,
		eventErr:        ErrContractCheckFailed.Error(),
		status:          0,
		giveback:        false,
	})

	// normal tx insufficient fromBalance before execution
	insufficientBlanceTx := mockNormalTransaction(bc.chainID, 0)
	insufficientBlanceTx.value = util.NewUint128()
	tests = append(tests, testTx{
		name:         "normal tx insufficient fromBalance",
		tx:           insufficientBlanceTx,
		fromBalance:  util.NewUint128(),
		gasUsed:      util.Uint128Zero(),
		afterBalance: util.NewUint128(),
		toBalance:    util.NewUint128(),
		wanted:       ErrInsufficientBalance,
		status:       0,
		giveback:     false,
	})

	// normal tx out of  gasLimit
	outOfGasLimitTx := mockNormalTransaction(bc.chainID, 0)
	outOfGasLimitTx.value = util.NewUint128()
	outOfGasLimitTx.gasLimit, _ = util.NewUint128FromInt(1)
	tests = append(tests, testTx{
		name:         "normal tx out of gasLimit",
		tx:           outOfGasLimitTx,
		fromBalance:  balance,
		gasUsed:      util.Uint128Zero(),
		afterBalance: balance,
		toBalance:    util.NewUint128(),
		wanted:       ErrOutOfGasLimit,
		status:       0,
		giveback:     false,
	})

	// tx payload load err
	payloadErrTx := mockDeployTransaction(bc.chainID, 0)
	payloadErrTx.value = util.NewUint128()
	payloadErrTx.data.Payload = []byte("0x00")
	gasCountOfTxBase, err := payloadErrTx.GasCountOfTxBase()
	assert.Nil(t, err)
	coinbaseBalance, err = payloadErrTx.gasPrice.Mul(gasCountOfTxBase)
	assert.Nil(t, err)
	balanceConsume, err = payloadErrTx.gasPrice.Mul(gasCountOfTxBase)
	assert.Nil(t, err)
	afterBalance, err = balance.Sub(balanceConsume)
	assert.Nil(t, err)
	getUsed, err := payloadErrTx.GasCountOfTxBase()
	assert.Nil(t, err)
	tests = append(tests, testTx{
		name:            "payload error tx",
		tx:              payloadErrTx,
		fromBalance:     balance,
		gasUsed:         getUsed,
		afterBalance:    afterBalance,
		toBalance:       afterBalance,
		coinbaseBalance: coinbaseBalance,
		wanted:          nil,
		eventErr:        "invalid argument(s)",
		status:          0,
		giveback:        false,
	})

	// tx execution err
	executionErrTx := mockCallTransaction(bc.chainID, 0, "test", "")
	executionErrTx.value = util.NewUint128()
	result, err := bc.SimulateTransactionExecution(executionErrTx)
	assert.Nil(t, err)
	assert.Equal(t, ErrContractCheckFailed, result.Err)
	coinbaseBalance, err = executionErrTx.gasPrice.Mul(result.GasUsed)
	assert.Nil(t, err)
	balanceConsume, err = executionErrTx.gasPrice.Mul(result.GasUsed)
	assert.Nil(t, err)
	afterBalance, err = balance.Sub(balanceConsume)
	assert.Nil(t, err)

	tests = append(tests, testTx{
		name:            "execution err tx",
		tx:              executionErrTx,
		fromBalance:     balance,
		gasUsed:         result.GasUsed,
		afterBalance:    afterBalance,
		toBalance:       util.NewUint128(),
		coinbaseBalance: coinbaseBalance,
		wanted:          nil,
		eventErr:        ErrContractCheckFailed.Error(),
		status:          0,
		giveback:        false,
	})

	// tx execution insufficient fromBalance after execution
	executionInsufficientBalanceTx := mockDeployTransaction(bc.chainID, 0)
	executionInsufficientBalanceTx.value = balance
	gasUsed, _ = util.NewUint128FromInt(21103)
	coinbaseBalance, err = executionInsufficientBalanceTx.gasPrice.Mul(gasUsed)
	assert.Nil(t, err)
	balanceConsume, err = executionInsufficientBalanceTx.gasPrice.Mul(gasUsed)
	assert.Nil(t, err)
	afterBalance, err = balance.Sub(balanceConsume)
	assert.Nil(t, err)
	tests = append(tests, testTx{
		name:            "execution insufficient fromBalance after execution tx",
		tx:              executionInsufficientBalanceTx,
		fromBalance:     balance,
		gasUsed:         gasUsed,
		afterBalance:    afterBalance,
		toBalance:       afterBalance,
		coinbaseBalance: coinbaseBalance,
		wanted:          nil,
		eventErr:        ErrInsufficientBalance.Error(),
		status:          0,
		giveback:        false,
	})

	// tx execution equal fromBalance after execution
	executionEqualBalanceTx := mockDeployTransaction(bc.chainID, 0)
	result, err = bc.SimulateTransactionExecution(executionEqualBalanceTx)
	assert.Nil(t, err)
	assert.Equal(t, ErrInsufficientBalance, result.Err)
	executionEqualBalanceTx.gasLimit = result.GasUsed
	t.Log("gasUsed:", result.GasUsed)
	coinbaseBalance, err = executionInsufficientBalanceTx.gasPrice.Mul(result.GasUsed)
	assert.Nil(t, err)
	executionEqualBalanceTx.value = balance
	gasCost, err := executionEqualBalanceTx.gasPrice.Mul(result.GasUsed)
	assert.Nil(t, err)
	fromBalance, err := gasCost.Add(balance)
	assert.Nil(t, err)
	tests = append(tests, testTx{
		name:            "execution equal fromBalance after execution tx",
		tx:              executionEqualBalanceTx,
		fromBalance:     fromBalance,
		gasUsed:         result.GasUsed,
		afterBalance:    balance,
		toBalance:       balance,
		coinbaseBalance: coinbaseBalance,
		wanted:          nil,
		eventErr:        "",
		status:          1,
		giveback:        false,
	})

	ks := keystore.DefaultKS
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, _ := ks.GetUnlocked(tt.tx.from.String())
			signature, _ := crypto.NewSignature(keystore.SECP256K1)
			signature.InitSign(key.(keystore.PrivateKey))

			err := tt.tx.Sign(signature)
			assert.Nil(t, err)

			block, err := bc.NewBlock(mockAddress())
			block.Begin()
			fromAcc, err := block.worldState.GetOrCreateUserAccount(tt.tx.from.address)
			assert.Nil(t, err)
			fromAcc.AddBalance(tt.fromBalance)
			block.Commit()

			block, err = bc.NewBlockFromParent(bc.tailBlock.header.coinbase, block)
			assert.Nil(t, err)
			block.Begin()

			txWorldState, err := block.WorldState().Prepare(tt.tx.Hash().String())
			assert.Nil(t, err)

			giveback, executionErr := VerifyExecution(tt.tx, block, txWorldState)
			assert.Equal(t, tt.wanted, executionErr)
			assert.Equal(t, giveback, tt.giveback)
			fromAcc, err = txWorldState.GetOrCreateUserAccount(tt.tx.from.address)
			assert.Nil(t, err)
			fmt.Println(fromAcc.Balance().String())

			txWorldState.CheckAndUpdate()
			assert.Nil(t, block.rewardCoinbaseForGas())
			block.Commit()

			fromAcc, err = block.worldState.GetOrCreateUserAccount(tt.tx.from.address)
			assert.Nil(t, err)
			toAcc, err := block.worldState.GetOrCreateUserAccount(tt.tx.to.address)
			assert.Nil(t, err)
			coinbaseAcc, err := block.worldState.GetOrCreateUserAccount(block.header.coinbase.address)
			assert.Nil(t, err)

			if tt.afterBalance != nil {
				assert.Equal(t, tt.afterBalance.String(), fromAcc.Balance().String())
			}
			if tt.toBalance != nil {
				assert.Equal(t, tt.toBalance, toAcc.Balance())
			}
			if tt.coinbaseBalance != nil {
				coinbaseBalance, err := tt.coinbaseBalance.Add(BlockReward)
				assert.Nil(t, err)
				assert.Equal(t, coinbaseBalance, coinbaseAcc.Balance())
			}

			events, _ := block.worldState.FetchEvents(tt.tx.hash)

			for _, v := range events {
				if v.Topic == TopicTransactionExecutionResult {
					txEvent := TransactionEvent{}
					json.Unmarshal([]byte(v.Data), &txEvent)
					status := int(txEvent.Status)
					assert.Equal(t, tt.status, status)
					assert.Equal(t, tt.eventErr, txEvent.Error)
					assert.Equal(t, tt.gasUsed.String(), txEvent.GasUsed)
					break
				}
			}
		})
	}
}

func TestTransaction_SimulateExecution(t *testing.T) {
	type testCase struct {
		name    string
		tx      *Transaction
		gasUsed *util.Uint128
		result  string
		wanted  error
	}

	tests := []testCase{}

	neb := testNeb(t)
	bc := neb.chain

	normalTx := mockNormalTransaction(bc.chainID, 0)
	normalTx.value, _ = util.NewUint128FromInt(1000000)
	tests = append(tests, testCase{
		name:    "normal tx",
		tx:      normalTx,
		gasUsed: MinGasCountPerTransaction,
		result:  "",
		wanted:  ErrInsufficientBalance,
	})

	deployTx := mockDeployTransaction(bc.chainID, 0)
	deployTx.to = deployTx.from
	deployTx.value = util.NewUint128()
	gasUsed, _ := util.NewUint128FromInt(21203)
	tests = append(tests, testCase{
		name:    "contract deploy tx",
		tx:      deployTx,
		gasUsed: gasUsed,
		result:  "",
		wanted:  ErrInsufficientBalance,
	})

	// contract call tx
	callTx := mockCallTransaction(bc.chainID, 1, "totalSupply", "")
	callTx.value = util.NewUint128()
	gasUsed, _ = util.NewUint128FromInt(20096)
	tests = append(tests, testCase{
		name:    "contract call tx",
		tx:      callTx,
		gasUsed: gasUsed,
		result:  "",
		wanted:  ErrContractCheckFailed,
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block, err := bc.NewBlock(bc.tailBlock.header.coinbase)
			assert.Nil(t, err)

			fromAcc, err := block.worldState.GetOrCreateUserAccount(tt.tx.from.address)
			assert.Nil(t, err)
			fromBefore := fromAcc.Balance()

			toAcc, err := block.worldState.GetOrCreateUserAccount(tt.tx.to.address)
			assert.Nil(t, err)
			toBefore := toAcc.Balance()

			coinbaseAcc, err := block.worldState.GetOrCreateUserAccount(block.header.coinbase.address)
			assert.Nil(t, err)
			coinbaseBefore := coinbaseAcc.Balance()

			result, err := tt.tx.simulateExecution(block)

			assert.Equal(t, tt.wanted, result.Err)
			assert.Equal(t, tt.result, result.Msg)
			assert.Equal(t, tt.gasUsed, result.GasUsed)
			assert.Nil(t, err)

			fromAcc, err = block.worldState.GetOrCreateUserAccount(tt.tx.from.address)
			assert.Nil(t, err)
			assert.Equal(t, fromBefore, fromAcc.Balance())

			toAcc, err = block.worldState.GetOrCreateUserAccount(tt.tx.to.address)
			assert.Nil(t, err)
			assert.Equal(t, toBefore, toAcc.Balance())

			coinbaseAcc, err = block.worldState.GetOrCreateUserAccount(block.header.coinbase.address)
			assert.Nil(t, err)
			assert.Equal(t, coinbaseBefore, coinbaseAcc.Balance())

			block.Seal()
		})
	}
}

func TestDeployAndCall(t *testing.T) {
	neb := testNeb(t)
	bc := neb.chain

	coinbase := mockAddress()
	from := mockAddress()
	balance, _ := util.NewUint128FromString("1000000000000000000")

	ks := keystore.DefaultKS
	key, _ := ks.GetUnlocked(from.String())
	signature, _ := crypto.NewSignature(keystore.SECP256K1)
	signature.InitSign(key.(keystore.PrivateKey))

	block, err := bc.NewBlock(coinbase)
	fromAcc, err := block.worldState.GetOrCreateUserAccount(from.address)
	assert.Nil(t, err)
	fromAcc.AddBalance(balance)
	block.Commit()

	// contract deploy tx
	deployTx := mockDeployTransaction(bc.chainID, 0)
	deployTx.from = from
	deployTx.to = from
	deployTx.value = util.NewUint128()
	deployTx.Sign(signature)

	// contract call tx
	callTx := mockCallTransaction(bc.chainID, 1, "totalSupply", "")
	callTx.from = from
	callTx.to, _ = deployTx.GenerateContractAddress()
	callTx.value = util.NewUint128()
	callTx.Sign(signature)

	assert.Nil(t, err)

	block, err = bc.NewBlockFromParent(bc.tailBlock.header.coinbase, block)
	assert.Nil(t, err)

	txWorldState, err := block.WorldState().Prepare(deployTx.Hash().String())
	assert.Nil(t, err)
	giveback, err := VerifyExecution(deployTx, block, txWorldState)
	assert.False(t, giveback)
	assert.Nil(t, err)
	giveback, err = AcceptTransaction(deployTx, txWorldState)
	assert.False(t, giveback)
	assert.Nil(t, err)
	_, err = txWorldState.CheckAndUpdate()
	assert.Nil(t, err)

	deployEvents, _ := block.worldState.FetchEvents(deployTx.Hash())
	assert.Equal(t, len(deployEvents), 1)
	event := deployEvents[0]
	if event.Topic == TopicTransactionExecutionResult {
		txEvent := TransactionEvent{}
		json.Unmarshal([]byte(event.Data), &txEvent)
		status := int(txEvent.Status)
		assert.Equal(t, status, TxExecutionSuccess)
	}

	txWorldState, err = block.WorldState().Prepare(callTx.Hash().String())
	assert.Nil(t, err)
	giveback, err = VerifyExecution(callTx, block, txWorldState)
	assert.False(t, giveback)
	assert.Nil(t, err)
	giveback, err = AcceptTransaction(callTx, txWorldState)
	assert.False(t, giveback)
	assert.Nil(t, err)
	_, err = txWorldState.CheckAndUpdate()
	assert.Nil(t, err)

	block.Commit()

	callEvents, _ := block.worldState.FetchEvents(callTx.Hash())
	assert.Equal(t, len(callEvents), 1)
	event = callEvents[0]
	if event.Topic == TopicTransactionExecutionResult {
		txEvent := TransactionEvent{}
		json.Unmarshal([]byte(event.Data), &txEvent)
		status := int(txEvent.Status)
		assert.Equal(t, status, TxExecutionSuccess)
	}
}

func Test1(t *testing.T) {
	fmt.Println(len(hash.Sha3256([]byte("abc"))))
}
func TestTransactionString(t *testing.T) {
	neb := testNeb(t)
	bc := neb.chain

	a := mockAddress()
	b := mockAddress()

	v, _ := util.NewUint128FromInt(1)
	data := `{"Function":"donation","Args":"[\"d\"]"}", "type":"call"}`
	tx1, _ := NewTransaction(bc.chainID, a, b, v, uint64(1), TxPayloadDeployType, []byte(data), TransactionGasPrice, TransactionMaxGas)
	expectedOut := fmt.Sprintf(`{"chainID":100,"data":"{\"Function\":\"donation\",\"Args\":\"[\\\"d\\\"]\"}\", \"type\":\"call\"}","from":"%s","gaslimit":"50000000000","gasprice":"1000000","hash":"","nonce":1,"timestamp":%d,"to":"%s","type":"deploy","value":"1"}`, a, tx1.timestamp, b)

	if tx1.String() == tx1.JSONString() {
		t.Errorf("tx String() != tx.JsonString")
	}

	if tx1.JSONString() != expectedOut {
		fmt.Println(tx1.JSONString())
		fmt.Println(expectedOut)
		t.Errorf("tx JsonString() is not working as xpected")
	}
}
