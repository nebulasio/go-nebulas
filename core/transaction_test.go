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

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/crypto"
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
	payload, _ := NewDeployPayload(source, sourceType, args).ToBytes()
	return mockTransaction(chainID, nonce, TxPayloadDeployType, payload)
}

func mockCallTransaction(chainID uint32, nonce uint64, function, args string) *Transaction {
	payload, _ := NewCallPayload(function, args).ToBytes()
	return mockTransaction(chainID, nonce, TxPayloadCallType, payload)
}

func mockDelegateTransaction(chainID uint32, nonce uint64, action, addr string) *Transaction {
	payload, _ := NewDelegatePayload(action, addr).ToBytes()
	return mockTransaction(chainID, nonce, TxPayloadDelegateType, payload)
}

func mockCandidateTransaction(chainID uint32, nonce uint64, action string) *Transaction {
	payload, _ := NewCandidatePayload(action).ToBytes()
	return mockTransaction(chainID, nonce, TxPayloadCandidateType, payload)
}

func mockTransaction(chainID uint32, nonce uint64, payloadType string, payload []byte) *Transaction {
	from := mockAddress()
	to := mockAddress()
	tx := NewTransaction(chainID, from, to, util.NewUint128(), nonce, payloadType, payload, TransactionGasPrice, TransactionMaxGas)
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
				12,
				&corepb.Data{Type: TxPayloadBinaryType, Payload: []byte("hwllo")},
				util.NewUint128(),
				util.NewUint128(),
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
				alg:       tt.fields.alg,
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

		tx := NewTransaction(1, from, to, util.NewUint128(), 10, TxPayloadBinaryType, []byte("datadata"), TransactionGasPrice, util.NewUint128FromInt(200000))

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
		eventTopic      []string
	}
	tests := []testTx{}

	bc, _ := NewBlockChain(testNeb())
	var c MockConsensus
	bc.SetConsensusHandler(c)

	// 1NAS = 10^18
	balance := util.NewUint128FromBigInt(util.NewUint128().Exp(util.NewUint128FromInt(10).Int, util.NewUint128FromInt(18).Int, nil))
	// normal tx
	normalTx := mockNormalTransaction(bc.chainID, 0)
	normalTx.value = util.NewUint128FromInt(1000000)
	afterBalance := util.NewUint128FromBigInt(util.NewUint128().Sub(balance.Int, util.NewUint128().Mul(normalTx.gasPrice.Int, MinGasCountPerTransaction.Int)))
	coinbaseBalance := util.NewUint128FromBigInt(util.NewUint128().Mul(normalTx.gasPrice.Int, MinGasCountPerTransaction.Int))
	tests = append(tests, testTx{
		name:            "normal tx",
		tx:              normalTx,
		fromBalance:     balance,
		gasUsed:         MinGasCountPerTransaction,
		afterBalance:    util.NewUint128FromBigInt(util.NewUint128().Sub(afterBalance.Int, normalTx.value.Int)),
		toBalance:       normalTx.value,
		coinbaseBalance: coinbaseBalance,
		wanted:          nil,
		eventTopic:      []string{TopicExecuteTxSuccess},
	})

	// contract deploy tx
	deployTx := mockDeployTransaction(bc.chainID, 0)
	deployTx.value = util.NewUint128()
	gasUsed := util.NewUint128FromInt(21232)
	coinbaseBalance = util.NewUint128FromBigInt(util.NewUint128().Mul(deployTx.gasPrice.Int, gasUsed.Int))
	tests = append(tests, testTx{
		name:            "contract deploy tx",
		tx:              deployTx,
		fromBalance:     balance,
		gasUsed:         gasUsed,
		afterBalance:    util.NewUint128FromBigInt(util.NewUint128().Sub(balance.Int, util.NewUint128().Mul(deployTx.gasPrice.Int, gasUsed.Int))),
		toBalance:       deployTx.value,
		coinbaseBalance: coinbaseBalance,
		wanted:          nil,
		eventTopic:      []string{TopicExecuteTxSuccess},
	})

	// contract call tx
	callTx := mockCallTransaction(bc.chainID, 1, "totalSupply", "")
	callTx.value = util.NewUint128()
	gasUsed = util.NewUint128FromInt(20036)
	coinbaseBalance = util.NewUint128FromBigInt(util.NewUint128().Mul(callTx.gasPrice.Int, gasUsed.Int))
	tests = append(tests, testTx{
		name:            "contract call tx",
		tx:              callTx,
		fromBalance:     balance,
		gasUsed:         gasUsed,
		afterBalance:    util.NewUint128FromBigInt(util.NewUint128().Sub(balance.Int, util.NewUint128().Mul(callTx.gasPrice.Int, gasUsed.Int))),
		toBalance:       callTx.value,
		coinbaseBalance: coinbaseBalance,
		wanted:          nil,
		eventTopic:      []string{TopicExecuteTxFailed},
	})

	// candidate tx
	candidateTx := mockCandidateTransaction(bc.chainID, 0, LoginAction)
	candidateTx.value = util.NewUint128()
	gasUsed = util.NewUint128FromInt(40018)
	coinbaseBalance = util.NewUint128FromBigInt(util.NewUint128().Mul(candidateTx.gasPrice.Int, gasUsed.Int))
	tests = append(tests, testTx{
		name:            "candidate tx",
		tx:              candidateTx,
		fromBalance:     balance,
		gasUsed:         gasUsed,
		afterBalance:    util.NewUint128FromBigInt(util.NewUint128().Sub(balance.Int, util.NewUint128().Mul(candidateTx.gasPrice.Int, gasUsed.Int))),
		toBalance:       candidateTx.value,
		coinbaseBalance: coinbaseBalance,
		wanted:          nil,
		eventTopic:      []string{TopicExecuteTxSuccess},
	})

	// delegate tx
	delegateTx := mockDelegateTransaction(bc.chainID, 0, DelegateAction, mockAddress().String())
	delegateTx.value = util.NewUint128()
	gasUsed = util.NewUint128FromInt(40078)
	coinbaseBalance = util.NewUint128FromBigInt(util.NewUint128().Mul(delegateTx.gasPrice.Int, gasUsed.Int))
	tests = append(tests, testTx{
		name:            "delegate tx",
		tx:              delegateTx,
		fromBalance:     balance,
		gasUsed:         gasUsed,
		afterBalance:    util.NewUint128FromBigInt(util.NewUint128().Sub(balance.Int, util.NewUint128().Mul(delegateTx.gasPrice.Int, gasUsed.Int))),
		toBalance:       delegateTx.value,
		coinbaseBalance: coinbaseBalance,
		wanted:          nil,
		eventTopic:      []string{TopicExecuteTxFailed},
	})

	// normal tx insufficient fromBalance before execution
	insufficientBlanceTx := mockNormalTransaction(bc.chainID, 0)
	insufficientBlanceTx.value = util.NewUint128()
	tests = append(tests, testTx{
		name:         "normal tx insufficient fromBalance",
		tx:           insufficientBlanceTx,
		fromBalance:  util.NewUint128(),
		gasUsed:      util.NewUint128(),
		afterBalance: util.NewUint128(),
		toBalance:    insufficientBlanceTx.value,
		wanted:       ErrInsufficientBalance,
		eventTopic:   []string{TopicExecuteTxFailed},
	})

	// normal tx out of  gasLimit
	outOfGasLimitTx := mockNormalTransaction(bc.chainID, 0)
	outOfGasLimitTx.value = util.NewUint128()
	outOfGasLimitTx.gasLimit = util.NewUint128FromInt(1)
	tests = append(tests, testTx{
		name:         "normal tx out of gasLimit",
		tx:           outOfGasLimitTx,
		fromBalance:  balance,
		gasUsed:      util.NewUint128(),
		afterBalance: balance,
		toBalance:    util.NewUint128(),
		wanted:       ErrOutOfGasLimit,
		eventTopic:   []string{TopicExecuteTxFailed},
	})

	// tx payload load err
	payloadErrTx := mockDeployTransaction(bc.chainID, 0)
	payloadErrTx.value = util.NewUint128()
	payloadErrTx.data.Payload = []byte("0x00")
	coinbaseBalance = util.NewUint128FromBigInt(util.NewUint128().Mul(delegateTx.gasPrice.Int, payloadErrTx.GasCountOfTxBase().Int))
	tests = append(tests, testTx{
		name:            "payload error tx",
		tx:              payloadErrTx,
		fromBalance:     balance,
		gasUsed:         payloadErrTx.GasCountOfTxBase(),
		afterBalance:    util.NewUint128FromBigInt(util.NewUint128().Sub(balance.Int, util.NewUint128().Mul(payloadErrTx.gasPrice.Int, payloadErrTx.GasCountOfTxBase().Int))),
		toBalance:       util.NewUint128(),
		coinbaseBalance: coinbaseBalance,
		wanted:          nil,
		eventTopic:      []string{TopicExecuteTxFailed},
	})

	// tx execution err
	executionErrTx := mockCallTransaction(bc.chainID, 0, "test", "")
	executionErrTx.value = util.NewUint128()
	gasUsed = util.NewUint128FromInt(20029)
	coinbaseBalance = util.NewUint128FromBigInt(util.NewUint128().Mul(executionErrTx.gasPrice.Int, gasUsed.Int))
	tests = append(tests, testTx{
		name:            "execution err tx",
		tx:              executionErrTx,
		fromBalance:     balance,
		gasUsed:         gasUsed,
		afterBalance:    util.NewUint128FromBigInt(util.NewUint128().Sub(balance.Int, util.NewUint128().Mul(executionErrTx.gasPrice.Int, gasUsed.Int))),
		toBalance:       util.NewUint128(),
		coinbaseBalance: coinbaseBalance,
		wanted:          nil,
		eventTopic:      []string{TopicExecuteTxFailed},
	})

	// tx execution insufficient fromBalance after execution
	executionInsufficientBalanceTx := mockDeployTransaction(bc.chainID, 0)
	executionInsufficientBalanceTx.value = balance
	gasUsed = util.NewUint128FromInt(21232)
	coinbaseBalance = util.NewUint128FromBigInt(util.NewUint128().Mul(executionInsufficientBalanceTx.gasPrice.Int, gasUsed.Int))
	tests = append(tests, testTx{
		name:            "execution insufficient fromBalance after execution tx",
		tx:              executionInsufficientBalanceTx,
		fromBalance:     balance,
		gasUsed:         gasUsed,
		afterBalance:    util.NewUint128FromBigInt(util.NewUint128().Sub(balance.Int, util.NewUint128().Mul(normalTx.gasPrice.Int, gasUsed.Int))),
		toBalance:       util.NewUint128(),
		coinbaseBalance: coinbaseBalance,
		wanted:          nil,
		eventTopic:      []string{TopicExecuteTxFailed},
	})

	// tx execution equal fromBalance after execution
	executionEqualBalanceTx := mockDeployTransaction(bc.chainID, 0)
	gasUsed = util.NewUint128FromInt(21232)
	coinbaseBalance = util.NewUint128FromBigInt(util.NewUint128().Mul(executionInsufficientBalanceTx.gasPrice.Int, gasUsed.Int))
	executionEqualBalanceTx.value = balance
	gasCost := util.NewUint128FromBigInt(util.NewUint128().Mul(executionEqualBalanceTx.gasPrice.Int, gasUsed.Int))
	tests = append(tests, testTx{
		name:            "execution equal fromBalance after execution tx",
		tx:              executionEqualBalanceTx,
		fromBalance:     util.NewUint128FromBigInt(util.NewUint128().Add(gasCost.Int, balance.Int)),
		gasUsed:         gasUsed,
		afterBalance:    util.NewUint128(),
		toBalance:       balance,
		coinbaseBalance: coinbaseBalance,
		wanted:          nil,
		eventTopic:      []string{TopicExecuteTxSuccess},
	})

	ks := keystore.DefaultKS
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, _ := ks.GetUnlocked(tt.tx.from.String())
			signature, _ := crypto.NewSignature(keystore.SECP256K1)
			signature.InitSign(key.(keystore.PrivateKey))

			err := tt.tx.Sign(signature)
			assert.Nil(t, err)

			block := bc.tailBlock
			block.begin()
			fromAcc, err := block.accState.GetOrCreateUserAccount(tt.tx.from.address)
			assert.Nil(t, err)
			fromAcc.AddBalance(tt.fromBalance)
			gasUsed, executionErr := tt.tx.VerifyExecution(block)
			fromAcc, err = block.accState.GetOrCreateUserAccount(tt.tx.from.address)
			assert.Nil(t, err)
			toAcc, err := block.accState.GetOrCreateUserAccount(tt.tx.to.address)
			assert.Nil(t, err)
			coinbaseAcc, err := block.accState.GetOrCreateUserAccount(block.header.coinbase.address)
			assert.Nil(t, err)
			if tt.gasUsed != nil {
				assert.Equal(t, tt.gasUsed, gasUsed)
			}
			if tt.afterBalance != nil {
				assert.Equal(t, tt.afterBalance.String(), fromAcc.Balance().String())
			}
			if tt.toBalance != nil {
				assert.Equal(t, tt.toBalance, toAcc.Balance())
			}
			if tt.coinbaseBalance != nil {
				assert.Equal(t, tt.coinbaseBalance, coinbaseAcc.Balance())
			}

			assert.Equal(t, tt.wanted, executionErr)

			events, _ := block.FetchEvents(tt.tx.hash)

			for index, event := range events {
				assert.Equal(t, tt.eventTopic[index], event.Topic)
			}

			block.rollback()
		})
	}

}
