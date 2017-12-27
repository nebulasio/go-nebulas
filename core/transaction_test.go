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
	payload, _ := NewBinaryPayload([]byte("data")).ToBytes()
	return mockTransaction(chainID, nonce, TxPayloadBinaryType, payload)
}

func mockDeployTransaction(chainID uint32, nonce uint64) *Transaction {
	source := `
	"use strict";var StandardToken=function(){LocalContractStorage.defineProperties(this,{name:null,symbol:null,totalSupply:null,totalIssued:null});LocalContractStorage.defineMapProperty(this,"balances")};StandardToken.prototype={init:function(name,symbol,totalSupply){this.name=name;this.symbol=symbol;this.totalSupply=totalSupply;this.totalIssued=0},totalSupply:function(){return this.totalSupply},balanceOf:function(owner){return this.balances.get(owner)||0},transfer:function(to,value){var balance=this.balanceOf(msg.sender);if(balance<value){return false}var finalBalance=balance-value;this.balances.set(msg.sender,finalBalance);this.balances.set(to,this.balanceOf(to)+value);return true},pay:function(msg,amount){if(this.totalIssued+amount>this.totalSupply){throw new Error("too much amount, exceed totalSupply")}this.balances.set(msg.sender,this.balanceOf(msg.sender)+amount);this.totalIssued+=amount}};module.exports=StandardToken;
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
		name    string
		tx      *Transaction
		balance *util.Uint128
		gas     *util.Uint128
		wanted  error
	}
	tests := []testTx{}

	bc, _ := NewBlockChain(testNeb())
	var c MockConsensus
	bc.SetConsensusHandler(c)

	balance := util.NewUint128FromBigInt(util.NewUint128().Mul(TransactionMaxGas.Int, TransactionGasPrice.Int))

	// normal tx
	normalTx := mockNormalTransaction(bc.chainID, 0)
	tests = append(tests, testTx{
		name:    "normal tx",
		tx:      normalTx,
		balance: balance,
		gas:     normalTx.GasCountOfTxBase(),
		wanted:  nil,
	})

	// contract deploy tx
	deployTx := mockDeployTransaction(bc.chainID, 0)
	tests = append(tests, testTx{
		name:    "contract deploy tx",
		tx:      deployTx,
		balance: balance,
		wanted:  nil,
	})

	// contract call tx
	callTx := mockCallTransaction(bc.chainID, 1, "totalSupply", "")
	tests = append(tests, testTx{
		name:    "contract call tx",
		tx:      callTx,
		balance: balance,
		wanted:  nil,
	})

	// candidate tx
	candidateTx := mockCandidateTransaction(bc.chainID, 0, LoginAction)
	tests = append(tests, testTx{
		name:    "candidate tx",
		tx:      candidateTx,
		balance: balance,
		wanted:  nil,
	})

	// delegate tx
	delegateTx := mockDelegateTransaction(bc.chainID, 0, DelegateAction, mockAddress().String())
	tests = append(tests, testTx{
		name:    "delegate tx",
		tx:      delegateTx,
		balance: balance,
		wanted:  nil,
	})

	// normal tx insufficient balance before execution
	insufficientBlanceTx := mockNormalTransaction(bc.chainID, 0)
	tests = append(tests, testTx{
		name:    "normal tx insufficient balance",
		tx:      insufficientBlanceTx,
		balance: util.NewUint128(),
		wanted:  ErrInsufficientBalance,
	})

	// normal tx out of  gasLimit
	outOfGasLimitTx := mockNormalTransaction(bc.chainID, 0)
	outOfGasLimitTx.gasLimit = util.NewUint128FromInt(1)
	tests = append(tests, testTx{
		name:    "normal tx out of gasLimit",
		tx:      outOfGasLimitTx,
		balance: balance,
		wanted:  ErrOutOfGasLimit,
	})

	// tx payload load err
	payloadErrTx := mockDeployTransaction(bc.chainID, 0)
	payloadErrTx.data.Payload = []byte("0x00")
	tests = append(tests, testTx{
		name:    "payload error tx",
		tx:      payloadErrTx,
		balance: balance,
		gas:     payloadErrTx.GasCountOfTxBase(),
		wanted:  nil,
	})

	// tx execution err
	executionErrTx := mockCallTransaction(bc.chainID, 0, "test", "")
	tests = append(tests, testTx{
		name:    "execution err tx",
		tx:      executionErrTx,
		balance: balance,
		wanted:  nil,
	})

	// tx execution insufficient balance after execution
	executionInsufficientTx := mockDeployTransaction(bc.chainID, 0)
	executionInsufficientTx.value = balance

	tests = append(tests, testTx{
		name:    "execution insufficient balance after execution tx",
		tx:      executionInsufficientTx,
		balance: balance,
		wanted:  nil,
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
			fromAcc := block.accState.GetOrCreateUserAccount(tt.tx.from.address)
			fromAcc.AddBalance(tt.balance)
			gasUsed, err := tt.tx.VerifyExecution(block)
			block.rollback()
			if tt.gas != nil {
				assert.Equal(t, tt.gas, gasUsed)
			}
			assert.Equal(t, tt.wanted, err)
		})
	}

}
