package nvm

/*
#include <stdlib.h>
#include "v8/lib/nvm_error.h"

*/
import "C"

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/crypto"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/keystore/secp256k1"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/stretchr/testify/assert"
)

func newUint128FromIntWrapper2(a int64) *util.Uint128 {
	b, _ := util.NewUint128FromInt(a)
	return b
}

func mockNormalTransaction2(from, to, value string) *core.Transaction {

	fromAddr, _ := core.AddressParse(from)
	toAddr, _ := core.AddressParse(to)
	payload, _ := core.NewBinaryPayload(nil).ToBytes()
	gasPrice, _ := util.NewUint128FromString("1000000")
	gasLimit, _ := util.NewUint128FromString("2000000")
	v, _ := util.NewUint128FromString(value)
	tx, _ := core.NewTransaction(1, fromAddr, toAddr, v, 1, core.TxPayloadBinaryType, payload, gasPrice, gasLimit)

	priv1 := secp256k1.GeneratePrivateKey()
	signature, _ := crypto.NewSignature(keystore.SECP256K1)
	signature.InitSign(priv1)
	tx.Sign(signature)
	return tx
}

type testBlock2 struct {
}

// Coinbase mock
func (block *testBlock2) Coinbase() *core.Address {
	addr, _ := core.AddressParse("n1FkntVUMPAsESuCAAPK711omQk19JotBjM")
	return addr
}

// Hash mock
func (block *testBlock2) Hash() byteutils.Hash {
	return []byte("59fc526072b09af8a8ca9732dae17132c4e9127e43cf2232")
}

// Height mock
func (block *testBlock2) Height() uint64 {
	return core.NebCompatibility.NvmMemoryLimitWithoutInjectHeight()
}

// RandomSeed mock
func (block *testBlock2) RandomSeed() string {
	return "59fc526072b09af8a8ca9732dae17132c4e9127e43cf2232"
}

// RandomAvailable mock
func (block *testBlock2) RandomAvailable() bool {
	return true
}

// DateAvailable
func (block *testBlock2) DateAvailable() bool {
	return true
}

// GetTransaction mock
func (block *testBlock2) GetTransaction(hash byteutils.Hash) (*core.Transaction, error) {
	return nil, nil
}

// RecordEvent mock
func (block *testBlock2) RecordEvent(txHash byteutils.Hash, topic, data string) error {
	return nil
}

func (block *testBlock2) Timestamp() int64 {
	return int64(0)
}

func (block *testBlock2) NR() core.NR {
	return nil
}

func mockBlock2() Block {
	block := &testBlock2{}
	return block
}

func testRandomFunc(t *testing.T, context WorldState) {
	contractAddr, err := core.AddressParse("n1FkntVUMPAsESuCAAPK711omQk19JotBjM")
	assert.Nil(t, err)
	contract, _ := context.CreateContractAccount(contractAddr.Bytes(), nil, nil)
	contract.AddBalance(newUint128FromIntWrapper2(5))

	tx := mockNormalTransaction2("n1FkntVUMPAsESuCAAPK711omQk19JotBjM", "n1TV3sU6jyzR4rJ1D7jCAmtVGSntJagXZHC", "0")
	ctx, err := NewContext(mockBlock2(), tx, contract, context)
	assert.Nil(t, err)

	// execute.
	engine := NewV8Engine(ctx)
	assert.Nil(t, engine.ctx.contextRand.rand)

	var cnt C.size_t
	var result *C.char
	var exception *C.char
	r1 := GetTxRandomFunc(unsafe.Pointer(uintptr(engine.lcsHandler)), &cnt, &result, &exception)
	fmt.Printf("r1:%v\n", r1)
	assert.Equal(t, r1, C.NVM_SUCCESS)
	assert.NotNil(t, result)
	assert.NotNil(t, engine.ctx.contextRand.rand)
	rs1 := C.GoString(result)
	if result != nil {
		C.free(unsafe.Pointer(result))
	}
	if exception != nil {
		C.free(unsafe.Pointer(exception))
	}
	result = nil
	exception = nil
	r2 := GetTxRandomFunc(unsafe.Pointer(uintptr(engine.lcsHandler)), &cnt, &result, &exception)
	assert.Equal(t, r2, C.NVM_SUCCESS)
	assert.NotNil(t, result)
	rs2 := C.GoString(result)

	assert.NotEqual(t, rs1, rs2)

	fmt.Println(rs1, rs2)
	if result != nil {
		C.free(unsafe.Pointer(result))
	}
	if exception != nil {
		C.free(unsafe.Pointer(exception))
	}
	engine.Dispose()
}
