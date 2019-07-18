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

package nvm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/nebulasio/go-nebulas/nr"

	"github.com/nebulasio/go-nebulas/account"
	"github.com/nebulasio/go-nebulas/consensus/dpos"

	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/crypto"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/keystore/secp256k1"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const contractStr = "n218MQSwc7hcXvM7rUkr6smMoiEf2VbGuYr"

func newUint128FromIntWrapper(a int64) *util.Uint128 {
	b, _ := util.NewUint128FromInt(a)
	return b
}

type testNR struct {
}

func (r *testNR) GetNRListByHeight(height uint64) (core.Data, error) {
	data := &nr.NRData{
		Nrs: []*nr.NRItem{&nr.NRItem{
			Address: "n1FkntVUMPAsESuCAAPK711omQk19JotBjM",
			Score:   "10.09",
		}},
	}
	return data, nil
}

func (r *testNR) GetNRSummary(height uint64) (core.Data, error) {
	if height > 1000 {
		sum := &nr.NRSummary{
			StartHeight: 1,
			EndHeight:   500,
			Sum: &nr.NRSummaryData{
				InOuts: "123123",
				Score:  "100",
			},
		}
		return sum, nil
	}
	return nil, nr.ErrNRSummaryNotFound
}

func (r *testNR) GetNRHandle(start, end, version uint64) (string, error) {
	return "", nil
}

func (r *testNR) GetNRListByHandle(handle []byte) (core.Data, error) {
	return nil, nil
}

type testBlock struct {
	height uint64

	nr core.NR
}

// Coinbase mock
func (block *testBlock) Coinbase() *core.Address {
	addr, _ := core.AddressParse("n1FkntVUMPAsESuCAAPK711omQk19JotBjM")
	return addr
}

// Hash mock
func (block *testBlock) Hash() byteutils.Hash {
	return []byte("59fc526072b09af8a8ca9732dae17132c4e9127e43cf2232")
}

// Height mock
func (block *testBlock) Height() uint64 {
	return block.height
}

// RandomSeed mock
func (block *testBlock) RandomSeed() string {
	return "59fc526072b09af8a8ca9732dae17132c4e9127e43cf2232"
}

// RandomAvailable mock
func (block *testBlock) RandomAvailable() bool {
	return true
}

// DateAvailable
func (block *testBlock) DateAvailable() bool {
	return true
}

// GetTransaction mock
func (block *testBlock) GetTransaction(hash byteutils.Hash) (*core.Transaction, error) {
	return nil, nil
}

// RecordEvent mock
func (block *testBlock) RecordEvent(txHash byteutils.Hash, topic, data string) error {
	return nil
}

func (block *testBlock) Timestamp() int64 {
	return int64(0)
}

func (block *testBlock) NR() core.NR {
	return block.nr
}

func mockBlock() Block {
	block := &testBlock{core.NebCompatibility.NvmMemoryLimitWithoutInjectHeight(), &testNR{}}
	return block
}

func mockBlockForLib(height uint64) Block {
	block := &testBlock{height, &testNR{}}
	return block
}

func mockTransaction() *core.Transaction {
	return mockNormalTransaction("n1FkntVUMPAsESuCAAPK711omQk19JotBjM", "n1JNHZJEUvfBYfjDRD14Q73FX62nJAzXkMR", "0")
}

const ContractName = "contract.js"

func mockNormalTransaction(from, to, value string) *core.Transaction {

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

func TestRunScriptSource(t *testing.T) {
	tests := []struct {
		filepath       string
		expectedErr    error
		expectedResult string
	}{
		{"test/test_require.js", nil, "\"\""},
		{"test/test_console.js", nil, "\"\""},
		{"test/test_storage_handlers.js", nil, "\"\""},
		{"test/test_storage_class.js", nil, "\"\""},
		{"test/test_storage.js", nil, "\"\""},
		{"test/test_eval.js", core.ErrExecutionFailed, "EvalError: Code generation from strings disallowed for this context"},
		{"test/test_date.js", nil, "\"\""},
		{"test/test_bignumber_random.js", core.ErrExecutionFailed, "Error: BigNumber.random is not allowed in nvm."},
		{"test/test_random_enable.js", nil, "\"\""},
		{"test/test_random_disable.js", core.ErrExecutionFailed, "Error: Math.random func is not allowed in nvm."},
		{"test/test_random_seed.js", core.ErrExecutionFailed, "Error: input seed must be a string"},
	}

	for _, tt := range tests {
		t.Run(tt.filepath, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.filepath)
			assert.Nil(t, err, "filepath read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewWorldState(dpos.NewDpos(), mem)
			owner, err := context.GetOrCreateUserAccount([]byte("account1"))
			assert.Nil(t, err)
			owner.AddBalance(newUint128FromIntWrapper(1000000000))
			contract, _ := context.CreateContractAccount([]byte("account2"), nil, nil)
			ctx, err := NewContext(mockBlock(), mockTransaction(), contract, context)

			engine := NewV8Engine(ctx)
			engine.SetExecutionLimits(900000, 10000000)
			result, err := engine.RunScriptSource(string(data), 0)
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedResult, result)
			engine.Dispose()
		})
	}
}

func TestRunScriptSourceInModule(t *testing.T) {
	tests := []struct {
		filepath    string
		sourceType  string
		expectedErr error
	}{
		{"./test/test_require.js", "js", nil},
		{"./test/test_setTimeout.js", "js", core.ErrExecutionFailed},
		{"./test/test_console.js", "js", nil},
		{"./test/test_storage_handlers.js", "js", nil},
		{"./test/test_storage_class.js", "js", nil},
		{"./test/test_storage.js", "js", nil},
		{"./test/test_ERC20.js", "js", nil},
		{"./test/test_eval.js", "js", core.ErrExecutionFailed},
	}

	for _, tt := range tests {
		t.Run(tt.filepath, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.filepath)
			assert.Nil(t, err, "filepath read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewWorldState(dpos.NewDpos(), mem)
			owner, err := context.GetOrCreateUserAccount([]byte("account1"))
			assert.Nil(t, err)
			owner.AddBalance(newUint128FromIntWrapper(1000000000))
			contract, _ := context.CreateContractAccount([]byte("account2"), nil, nil)
			ctx, err := NewContext(mockBlock(), mockTransaction(), contract, context)

			engine := NewV8Engine(ctx)
			engine.SetExecutionLimits(100000, 10000000)
			engine.AddModule(ContractName, string(data), 0)
			runnableSource := fmt.Sprintf("require(\"%s\");", ContractName)
			_, err = engine.RunScriptSource(runnableSource, 0)

			assert.Equal(t, tt.expectedErr, err)
			engine.Dispose()
		})
	}
}

func TestRunScriptSourceWithLimits(t *testing.T) {
	tests := []struct {
		name                          string
		filepath                      string
		limitsOfExecutionInstructions uint64
		limitsOfTotalMemorySize       uint64
		expectedErr                   error
	}{
		{"1", "test/test_oom_1.js", 100000, 0, ErrInsufficientGas},
		{"2", "test/test_oom_1.js", 0, 500000, ErrExceedMemoryLimits},
		{"3", "test/test_oom_1.js", 1000000, 50000000, ErrInsufficientGas},
		{"4", "test/test_oom_1.js", 5000000, 70000, ErrExceedMemoryLimits},

		{"5", "test/test_oom_2.js", 100000, 0, ErrInsufficientGas},
		{"6", "test/test_oom_2.js", 0, 80000, ErrExceedMemoryLimits},
		{"7", "test/test_oom_2.js", 10000000, 10000000, ErrInsufficientGas},
		{"8", "test/test_oom_2.js", 10000000, 70000, ErrExceedMemoryLimits},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.filepath)
			assert.Nil(t, err, "filepath read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewWorldState(dpos.NewDpos(), mem)
			owner, err := context.GetOrCreateUserAccount([]byte("account1"))
			assert.Nil(t, err)
			owner.AddBalance(newUint128FromIntWrapper(100000))
			contract, _ := context.CreateContractAccount([]byte("account2"), nil, nil)
			ctx, err := NewContext(mockBlock(), mockTransaction(), contract, context)

			// direct run.
			(func() {
				engine := NewV8Engine(ctx)
				engine.SetExecutionLimits(tt.limitsOfExecutionInstructions, tt.limitsOfTotalMemorySize)
				source, _, _ := engine.InjectTracingInstructions(string(data))
				_, err = engine.RunScriptSource(source, 0)
				fmt.Printf("err:%v\n", err)
				assert.Equal(t, tt.expectedErr, err)
				engine.Dispose()
			})()

			// modularized run.
			(func() {
				moduleID := fmt.Sprintf("%s", ContractName)
				runnableSource := fmt.Sprintf("require(\"%s\");", moduleID)

				engine := NewV8Engine(ctx)
				engine.SetExecutionLimits(tt.limitsOfExecutionInstructions, tt.limitsOfTotalMemorySize)
				engine.AddModule(ContractName, string(data), 0)
				_, err = engine.RunScriptSource(runnableSource, 0)
				assert.Equal(t, tt.expectedErr, err)
				engine.Dispose()
			})()
		})
	}
}

func TestRunScriptSourceMemConsistency(t *testing.T) {
	tests := []struct {
		name                          string
		filepath                      string
		limitsOfExecutionInstructions uint64
		limitsOfTotalMemorySize       uint64
		expectedMem                   uint64
	}{
		{"3", "test/test_oom_3.js", 1000000000, 5000000000, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.filepath)
			assert.Nil(t, err, "filepath read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewWorldState(dpos.NewDpos(), mem)
			owner, err := context.GetOrCreateUserAccount([]byte("account1"))
			assert.Nil(t, err)
			owner.AddBalance(newUint128FromIntWrapper(100000))
			contract, _ := context.CreateContractAccount([]byte("account2"), nil, nil)
			ctx, err := NewContext(mockBlock(), mockTransaction(), contract, context)

			// direct run.
			(func() {
				engine := NewV8Engine(ctx)
				engine.SetExecutionLimits(tt.limitsOfExecutionInstructions, tt.limitsOfTotalMemorySize)
				source, _, _ := engine.InjectTracingInstructions(string(data))
				_, err = engine.RunScriptSource(source, 0)
				// assert.Equal(t, tt.expectedErr, err)
				assert.Nil(t, err)
				engine.Dispose()
			})()

			// modularized run.
			(func() {
				moduleID := fmt.Sprintf("%s", ContractName)
				runnableSource := fmt.Sprintf("require(\"%s\");", moduleID)

				engine := NewV8Engine(ctx)
				engine.SetExecutionLimits(tt.limitsOfExecutionInstructions, tt.limitsOfTotalMemorySize)
				engine.AddModule(ContractName, string(data), 0)
				_, err = engine.RunScriptSource(runnableSource, 0)
				// assert.Equal(t, tt.expectedErr, err)
				assert.Nil(t, err)
				engine.CollectTracingStats()
				// fmt.Printf("total:%v", engine.actualTotalMemorySize)
				assert.Equal(t, uint64(6703104), engine.actualTotalMemorySize)
				engine.Dispose()
			})()
		})
	}
}

func TestV8ResourceLimit(t *testing.T) {
	tests := []struct {
		name          string
		contractPath  string
		sourceType    string
		initArgs      string
		callArgs      string
		initExceptErr string
		callExceptErr string
	}{
		{"deploy test_oom_4.js", "./test/test_oom_4.js", "js", "[31457280]", "[31457280]", "", ""},
		{"deploy test_oom_4.js", "./test/test_oom_4.js", "js", "[37748736]", "[37748736]", "", ""},
		{"deploy test_oom_4.js", "./test/test_oom_4.js", "js", "[41943039]", "[41943039]", "exceed memory limits", "exceed memory limits"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.contractPath)
			assert.Nil(t, err, "contract path read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewWorldState(dpos.NewDpos(), mem)
			owner, err := context.GetOrCreateUserAccount([]byte("account1"))
			assert.Nil(t, err)
			owner.AddBalance(newUint128FromIntWrapper(10000000))
			contract, _ := context.CreateContractAccount([]byte("account2"), nil, nil)

			ctx, err := NewContext(mockBlock(), mockTransaction(), contract, context)
			engine := NewV8Engine(ctx)
			engine.SetExecutionLimits(10000, 10000000)
			_, err = engine.DeployAndInit(string(data), tt.sourceType, tt.initArgs)
			engine.CollectTracingStats()
			fmt.Printf("total:%v", engine.actualTotalMemorySize)
			// assert.Nil(t, err)
			if err != nil {
				fmt.Printf("err:%v", err.Error())
				assert.Equal(t, tt.initExceptErr, err.Error())
			} else {
				assert.Equal(t, tt.initExceptErr, "")
			}
			// assert.Equal(t, tt.initExceptErr, err.Error)

			engine.Dispose()

			engine = NewV8Engine(ctx)
			engine.SetExecutionLimits(10000, 10000000)
			_, err = engine.Call(string(data), tt.sourceType, "newMem", tt.callArgs)
			// assert.Nil(t, err)
			// assert.Equal(t, tt.initExceptErr, err.Error)
			if err != nil {
				assert.Equal(t, tt.initExceptErr, err.Error())
			} else {
				assert.Equal(t, tt.initExceptErr, "")
			}
			engine.Dispose()

		})
	}
}
func TestRunScriptSourceTimeout(t *testing.T) {
	core.NebCompatibility = core.NewCompatibilityLocal()
	tests := []struct {
		filepath    string
		height      uint64
		expectedErr error
	}{
		{"test/test_infinite_loop.js", 1, ErrExecutionTimeout},
		{"test/test_infinite_loop.js", 2, core.ErrUnexpected},
	}

	for _, tt := range tests {
		t.Run(tt.filepath, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.filepath)
			assert.Nil(t, err, "filepath read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewWorldState(dpos.NewDpos(), mem)

			// owner, err := context.GetOrCreateUserAccount([]byte("account1"))
			// assert.Nil(t, err)

			contract, _ := context.CreateContractAccount([]byte("account2"), nil, nil)
			ctx, err := NewContext(mockBlockForLib(tt.height), mockTransaction(), contract, context)

			// direct run.
			(func() {
				engine := NewV8Engine(ctx)
				_, err = engine.RunScriptSource(string(data), 0)
				assert.Equal(t, tt.expectedErr, err)
				engine.Dispose()
			})()

			// modularized run.
			(func() {
				moduleID := fmt.Sprintf("%s", ContractName)
				runnableSource := fmt.Sprintf("require(\"%s\");", moduleID)

				engine := NewV8Engine(ctx)
				engine.AddModule(moduleID, string(data), 0)
				_, err = engine.RunScriptSource(runnableSource, 0)
				assert.Equal(t, tt.expectedErr, err)
				engine.Dispose()
			})()
		})
	}
}

func TestDeployAndInitAndCall(t *testing.T) {
	tests := []struct {
		name         string
		contractPath string
		sourceType   string
		initArgs     string
		verifyArgs   string
	}{
		{"deploy sample_contract.js", "./test/sample_contract.js", "js", "[\"TEST001\", 123,[{\"name\":\"robin\",\"count\":2},{\"name\":\"roy\",\"count\":3},{\"name\":\"leon\",\"count\":4}]]", "[\"TEST001\", 123,[{\"name\":\"robin\",\"count\":2},{\"name\":\"roy\",\"count\":3},{\"name\":\"leon\",\"count\":4}]]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.contractPath)
			assert.Nil(t, err, "contract path read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewWorldState(dpos.NewDpos(), mem)
			owner, err := context.GetOrCreateUserAccount([]byte("account1"))
			assert.Nil(t, err)
			owner.AddBalance(newUint128FromIntWrapper(10000000))
			contract, _ := context.CreateContractAccount([]byte("account2"), nil, nil)

			ctx, err := NewContext(mockBlock(), mockTransaction(), contract, context)
			engine := NewV8Engine(ctx)
			engine.SetExecutionLimits(10000, 10000000)
			_, err = engine.DeployAndInit(string(data), tt.sourceType, tt.initArgs)
			assert.Nil(t, err)
			engine.Dispose()

			engine = NewV8Engine(ctx)
			engine.SetExecutionLimits(10000, 10000000)
			_, err = engine.Call(string(data), tt.sourceType, "dump", "")
			assert.Nil(t, err)
			engine.Dispose()

			engine = NewV8Engine(ctx)
			engine.SetExecutionLimits(10000, 10000000)
			_, err = engine.Call(string(data), tt.sourceType, "verify", tt.verifyArgs)
			assert.Nil(t, err)
			engine.Dispose()

			// force error.
			mem, _ = storage.NewMemoryStorage()
			context, _ = state.NewWorldState(dpos.NewDpos(), mem)
			owner, err = context.GetOrCreateUserAccount([]byte("account1"))
			assert.Nil(t, err)
			contract, err = context.CreateContractAccount([]byte("account2"), nil, nil)
			assert.Nil(t, err)

			ctx, err = NewContext(mockBlock(), mockTransaction(), contract, context)
			engine = NewV8Engine(ctx)
			engine.SetExecutionLimits(10000, 10000000)
			_, err = engine.Call(string(data), tt.sourceType, "verify", tt.verifyArgs)
			assert.NotNil(t, err)
			engine.Dispose()
		})
	}
}

func TestContracts(t *testing.T) {
	type fields struct {
		function string
		args     string
	}
	tests := []struct {
		contract   string
		sourceType string
		initArgs   string
		calls      []fields
	}{
		{
			"./test/contract_rectangle.js",
			"js",
			"[\"1024\", \"768\"]",
			[]fields{
				{"calcArea", "[]"},
				{"verify", "[\"786432\"]"},
			},
		},
		{
			"./test/contract_rectangle.js",
			"js",
			"[\"999\", \"123\"]",
			[]fields{
				{"calcArea", "[]"},
				{"verify", "[\"122877\"]"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.contract, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.contract)
			assert.Nil(t, err, "contract path read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewWorldState(dpos.NewDpos(), mem)
			owner, err := context.GetOrCreateUserAccount([]byte("account1"))
			assert.Nil(t, err)
			owner.AddBalance(newUint128FromIntWrapper(10000000))
			contract, err := context.CreateContractAccount([]byte("account2"), nil, nil)
			assert.Nil(t, err)
			ctx, err := NewContext(mockBlock(), mockTransaction(), contract, context)

			// deploy and init.
			engine := NewV8Engine(ctx)
			engine.SetExecutionLimits(1000, 10000000)
			_, err = engine.DeployAndInit(string(data), tt.sourceType, tt.initArgs)
			assert.Nil(t, err)
			engine.Dispose()

			// call.
			for _, fields := range tt.calls {
				engine = NewV8Engine(ctx)
				engine.SetExecutionLimits(1000, 10000000)
				_, err = engine.Call(string(data), tt.sourceType, fields.function, fields.args)
				assert.Nil(t, err)
				engine.Dispose()
			}
		})
	}
}

func TestFunctionNameCheck(t *testing.T) {
	tests := []struct {
		function    string
		expectedErr error
		args        string
	}{
		{"$dump", nil, ""},
		{"dump", nil, ""},
		{"dump_1", nil, ""},
		{"init", ErrDisallowCallPrivateFunction, ""},
		{"Init", ErrDisallowCallPrivateFunction, ""},
		{"9dump", ErrDisallowCallNotStandardFunction, ""},
		{"_dump", ErrDisallowCallNotStandardFunction, ""},
	}

	for _, tt := range tests {
		t.Run(tt.function, func(t *testing.T) {
			data, err := ioutil.ReadFile("test/sample_contract.js")
			sourceType := "js"
			assert.Nil(t, err, "contract path read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewWorldState(dpos.NewDpos(), mem)
			owner, err := context.GetOrCreateUserAccount([]byte("account1"))
			assert.Nil(t, err)
			owner.AddBalance(newUint128FromIntWrapper(1000000))
			contract, _ := context.CreateContractAccount([]byte("account2"), nil, nil)
			ctx, err := NewContext(mockBlock(), mockTransaction(), contract, context)

			engine := NewV8Engine(ctx)
			engine.SetExecutionLimits(1000, 10000000)
			_, err = engine.Call(string(data), sourceType, tt.function, tt.args)
			assert.Equal(t, tt.expectedErr, err)
			engine.Dispose()
		})
	}
}
func TestMultiEngine(t *testing.T) {
	mem, _ := storage.NewMemoryStorage()
	context, _ := state.NewWorldState(dpos.NewDpos(), mem)
	owner, err := context.GetOrCreateUserAccount([]byte("account1"))
	assert.Nil(t, err)
	owner.AddBalance(newUint128FromIntWrapper(1000000))
	contract, _ := context.CreateContractAccount([]byte("account2"), nil, nil)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			ctx, err := NewContext(mockBlock(), mockTransaction(), contract, context)

			engine := NewV8Engine(ctx)
			engine.SetExecutionLimits(1000, 10000000)
			defer engine.Dispose()

			_, err = engine.RunScriptSource("console.log('running.');", 0)
			assert.Nil(t, err)
		}()
	}
	wg.Wait()
}

func TestInstructionCounter1_1_0TestSuite(t *testing.T) {
	core.NebCompatibility = core.NewCompatibilityLocal()
	tests := []struct {
		filepath                                string
		strictDisallowUsageOfInstructionCounter int
		expectedErr                             error
		expectedResult                          string
	}{
		{"./test/instruction_counter_tests/redefine1.js", 1, ErrInjectTracingInstructionFailed, ""},
		{"./test/instruction_counter_tests/redefine2.js", 1, ErrInjectTracingInstructionFailed, ""},
		{"./test/instruction_counter_tests/redefine3.js", 1, ErrInjectTracingInstructionFailed, ""},
		{"./test/instruction_counter_tests/redefine4.js", 1, ErrInjectTracingInstructionFailed, ""},
		{"./test/instruction_counter_tests/redefine5.js", 1, ErrInjectTracingInstructionFailed, ""},
		{"./test/instruction_counter_tests/redefine6.js", 1, ErrInjectTracingInstructionFailed, ""},
		{"./test/instruction_counter_tests/redefine7.js", 1, ErrInjectTracingInstructionFailed, ""},
		{"./test/instruction_counter_tests/function.js", 1, ErrInjectTracingInstructionFailed, ""},
		{"./test/instruction_counter_tests/redefine1.js", 0, ErrInjectTracingInstructionFailed, ""},
		{"./test/instruction_counter_tests/redefine2.js", 0, ErrInjectTracingInstructionFailed, ""},
		{"./test/instruction_counter_tests/redefine3.js", 0, ErrInjectTracingInstructionFailed, ""},
		{"./test/instruction_counter_tests/redefine4.js", 0, core.ErrExecutionFailed, "Error: still not break the jail of _instruction_counter."},
		{"./test/instruction_counter_tests/redefine5.js", 0, ErrInjectTracingInstructionFailed, ""},
		{"./test/instruction_counter_tests/redefine6.js", 0, nil, "\"\""},
		{"./test/instruction_counter_tests/redefine7.js", 0, nil, "\"\""},
		{"./test/instruction_counter_tests/function.js", 0, nil, "\"\""},
		{"./test/instruction_counter_tests/if.js", 0, nil, "\"\""},
		{"./test/instruction_counter_tests/switch.js", 0, nil, "\"\""},
		{"./test/instruction_counter_tests/for.js", 0, nil, "\"\""},
		{"./test/instruction_counter_tests/with.js", 0, nil, "\"\""},
		{"./test/instruction_counter_tests/while.js", 0, nil, "\"\""},
		{"./test/instruction_counter_tests/throw.js", 0, nil, "\"\""},
		{"./test/instruction_counter_tests/switch.js", 0, nil, "\"\""},
		{"./test/instruction_counter_tests/condition_operator.js", 0, nil, "\"\""},
		{"./test/instruction_counter_tests/storage_usage.js", 0, nil, "\"\""},
		{"./test/instruction_counter_tests/event_usage.js", 0, nil, "\"\""},
		{"./test/instruction_counter_tests/blockchain_usage.js", 0, nil, "\"\""},
	}

	for _, tt := range tests {
		t.Run(tt.filepath, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.filepath)
			assert.Nil(t, err, "filepath read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewWorldState(dpos.NewDpos(), mem)
			owner, err := context.GetOrCreateUserAccount([]byte("account1"))
			assert.Nil(t, err)
			owner.AddBalance(newUint128FromIntWrapper(1000000000))
			addr, err := core.NewContractAddressFromData([]byte("n1FkntVUMPAsESuCAAPK711omQk19JotBjM"), byteutils.FromUint64(1))
			assert.Nil(t, err)
			contract, err := context.CreateContractAccount(addr.Bytes(), nil, &corepb.ContractMeta{Version: "1.1.0"})
			assert.Nil(t, err)
			ctx, err := NewContext(mockBlockForLib(3), mockTransaction(), contract, context)

			moduleID := ContractName
			runnableSource := fmt.Sprintf("var x = require(\"%s\");", moduleID)

			engine := NewV8Engine(ctx)
			engine.strictDisallowUsageOfInstructionCounter = tt.strictDisallowUsageOfInstructionCounter
			engine.enableLimits = true
			err = engine.AddModule(moduleID, string(data), 0)
			if err != nil {
				assert.Equal(t, tt.expectedErr, err)
			} else {
				result, err := engine.RunScriptSource(runnableSource, 0)
				assert.Equal(t, tt.expectedErr, err)
				assert.Equal(t, tt.expectedResult, result)
			}
			engine.Dispose()
		})
	}
}
func TestCounter1_0_0TestSuite(t *testing.T) {
	core.NebCompatibility = core.NewCompatibilityTestNet()
	ClearSourceModuleCache()
	tests := []struct {
		filepath                                string
		strictDisallowUsageOfInstructionCounter int
		expectedErr                             error
		expectedResult                          string
	}{
		{"./test/instruction_counter_tests/redefine1.js", 1, ErrInjectTracingInstructionFailed, ""},
		{"./test/instruction_counter_tests/redefine2.js", 1, ErrInjectTracingInstructionFailed, ""},
		{"./test/instruction_counter_tests/redefine3.js", 1, ErrInjectTracingInstructionFailed, ""},
		{"./test/instruction_counter_tests/redefine4.js", 1, ErrInjectTracingInstructionFailed, ""},
		{"./test/instruction_counter_tests/redefine5.js", 1, ErrInjectTracingInstructionFailed, ""},
		{"./test/instruction_counter_tests/redefine6.js", 1, ErrInjectTracingInstructionFailed, ""},
		{"./test/instruction_counter_tests/redefine7.js", 1, ErrInjectTracingInstructionFailed, ""},
		{"./test/instruction_counter_tests/function.js", 1, ErrInjectTracingInstructionFailed, ""},
		{"./test/instruction_counter_tests/redefine1.js", 0, ErrInjectTracingInstructionFailed, ""},
		{"./test/instruction_counter_tests/redefine2.js", 0, ErrInjectTracingInstructionFailed, ""},
		{"./test/instruction_counter_tests/redefine3.js", 0, ErrInjectTracingInstructionFailed, ""},
		{"./test/instruction_counter_tests/redefine4.js", 0, core.ErrExecutionFailed, "Error: still not break the jail of _instruction_counter."},
		{"./test/instruction_counter_tests/redefine5.js", 0, ErrInjectTracingInstructionFailed, ""},
		{"./test/instruction_counter_tests/redefine6.js", 0, nil, "\"\""},
		{"./test/instruction_counter_tests/redefine7.js", 0, nil, "\"\""},
		{"./test/instruction_counter_tests/function.js", 0, nil, "\"\""},
		{"./test/instruction_counter_tests/if.js", 0, nil, "\"\""},
		{"./test/instruction_counter_tests/switch.js", 0, nil, "\"\""},
		{"./test/instruction_counter_tests/for_1_0_0.js", 0, nil, "\"\""},
		{"./test/instruction_counter_tests/with.js", 0, nil, "\"\""},
		{"./test/instruction_counter_tests/while_1_0_0.js", 0, nil, "\"\""},
		{"./test/instruction_counter_tests/throw.js", 0, nil, "\"\""},
		{"./test/instruction_counter_tests/switch.js", 0, nil, "\"\""},
		{"./test/instruction_counter_tests/condition_operator.js", 0, nil, "\"\""},
		{"./test/instruction_counter_tests/storage_usage.js", 0, nil, "\"\""},
		{"./test/instruction_counter_tests/event_usage.js", 0, nil, "\"\""},
		{"./test/instruction_counter_tests/blockchain_usage.js", 0, nil, "\"\""},
	}

	for _, tt := range tests {
		t.Run(tt.filepath, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.filepath)
			assert.Nil(t, err, "filepath read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewWorldState(dpos.NewDpos(), mem)
			owner, err := context.GetOrCreateUserAccount([]byte("account1"))
			assert.Nil(t, err)
			owner.AddBalance(newUint128FromIntWrapper(1000000000))
			addr, err := core.NewContractAddressFromData([]byte("n1FkntVUMPAsESuCAAPK711omQk19JotBjM"), byteutils.FromUint64(1))
			assert.Nil(t, err)
			contract, err := context.CreateContractAccount(addr.Bytes(), nil, nil)
			assert.Nil(t, err)
			ctx, err := NewContext(mockBlockForLib(1), mockTransaction(), contract, context)

			moduleID := ContractName
			runnableSource := fmt.Sprintf("var x = require(\"%s\");", moduleID)

			engine := NewV8Engine(ctx)
			engine.strictDisallowUsageOfInstructionCounter = tt.strictDisallowUsageOfInstructionCounter
			engine.enableLimits = true
			err = engine.AddModule(moduleID, string(data), 0)
			if err != nil {
				assert.Equal(t, tt.expectedErr, err)
			} else {
				result, err := engine.RunScriptSource(runnableSource, 0)
				assert.Equal(t, tt.expectedErr, err)
				assert.Equal(t, tt.expectedResult, result)
			}
			engine.Dispose()
		})
	}
}

func TestTypeScriptExecution(t *testing.T) {
	core.NebCompatibility = core.NewCompatibilityLocal()
	tests := []struct {
		filepath    string
		expectedErr error
	}{
		{"./test/test_greeter.ts", nil},
	}

	for _, tt := range tests {
		t.Run(tt.filepath, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.filepath)
			assert.Nil(t, err, "filepath read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewWorldState(dpos.NewDpos(), mem)
			owner, err := context.GetOrCreateUserAccount([]byte("account1"))
			assert.Nil(t, err)
			owner.AddBalance(newUint128FromIntWrapper(1000000000))
			contract, err := context.CreateContractAccount([]byte("account2"), nil, &corepb.ContractMeta{Version: "1.1.0"})
			assert.Nil(t, err)
			ctx, err := NewContext(mockBlock(), mockTransaction(), contract, context)

			moduleID := ContractName
			runnableSource := fmt.Sprintf("require(\"%s\");", moduleID)

			engine := NewV8Engine(ctx)
			defer engine.Dispose()

			engine.enableLimits = true
			jsSource, _, err := engine.TranspileTypeScript(string(data))
			if err != nil {
				assert.Equal(t, tt.expectedErr, err)
				return
			}

			err = engine.AddModule(moduleID, string(jsSource), 0)
			if err != nil {
				assert.Equal(t, tt.expectedErr, err)
			} else {
				_, err := engine.RunScriptSource(runnableSource, 0)
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func DeprecatedTestRunMozillaJSTestSuite(t *testing.T) {
	mem, _ := storage.NewMemoryStorage()
	context, _ := state.NewWorldState(dpos.NewDpos(), mem)
	owner, err := context.GetOrCreateUserAccount([]byte("account1"))
	assert.Nil(t, err)
	owner.AddBalance(newUint128FromIntWrapper(1000000000))

	contract, err := context.CreateContractAccount([]byte("account2"), nil, nil)
	assert.Nil(t, err)
	ctx, err := NewContext(mockBlock(), mockTransaction(), contract, context)

	var runTest func(dir string, shelljs string)
	runTest = func(dir string, shelljs string) {
		files, err := ioutil.ReadDir(dir)
		require.Nil(t, err)

		cwdShelljs := fmt.Sprintf("%s/shell.js", dir)
		if _, err := os.Stat(cwdShelljs); !os.IsNotExist(err) {
			shelljs = fmt.Sprintf("%s;%s", shelljs, cwdShelljs)
		}

		for _, file := range files {
			filepath := fmt.Sprintf("%s/%s", dir, file.Name())
			fi, err := os.Stat(filepath)
			require.Nil(t, err)

			if fi.IsDir() {
				runTest(filepath, shelljs)
				continue
			}

			if !strings.HasSuffix(file.Name(), ".js") {
				continue
			}
			if strings.Compare(file.Name(), "browser.js") == 0 || strings.Compare(file.Name(), "shell.js") == 0 || strings.HasPrefix(file.Name(), "toLocale") {
				continue
			}

			buf := bytes.NewBufferString("this.print = console.log;var native_eval = eval;eval = function (s) { try {  return native_eval(s); } catch (e) { return \"error\"; }};")

			jsfiles := fmt.Sprintf("%s;%s;%s", shelljs, "test/mozilla_js_tests_loader.js", filepath)

			for _, v := range strings.Split(jsfiles, ";") {
				if len(v) == 0 {
					continue
				}

				fi, err := os.Stat(v)
				require.Nil(t, err)
				f, err := os.Open(v)
				require.Nil(t, err)
				reader := bufio.NewReader(f)
				buf.Grow(int(fi.Size()))
				buf.ReadFrom(reader)
			}
			// execute.
			engine := NewV8Engine(ctx)
			engine.SetTestingFlag(true)
			engine.enableLimits = true
			_, err = engine.RunScriptSource(buf.String(), 0)
			//t.Logf("ret:%v, err:%v", ret, err)
			assert.Nil(t, err)
		}
	}

	runTest("test/mozilla_js_tests", "")
}

func TestNebulasContract(t *testing.T) {
	core.NebCompatibility = core.NewCompatibilityLocal()
	tests := []struct {
		name     string
		value    string
		function string
		args     string
		err      error
	}{
		{"1", "0", "unpayable", "", nil},
		{"2", "0", "unpayable", "[1]", nil},
		{"3", "1", "unpayable", "", nil},
		{"4", "0", "payable", "", core.ErrExecutionFailed},
		{"5", "1", "payable", "", nil},
		{"6", "1", "payable", "[1]", nil},
		{"7", "0", "contract1", "[1]", nil},
		{"8", "1", "contract1", "[1]", nil},
		{"9", "0", "contract2", "[1]", core.ErrExecutionFailed},
		{"10", "1", "contract2", "[1]", core.ErrExecutionFailed},
		{"11", "0", "contract3", "[1]", core.ErrExecutionFailed},
		{"12", "1", "contract3", "[1]", nil},
		{"13", "0", "contract4", "[1]", core.ErrExecutionFailed},
		{"14", "1", "contract4", "[1]", core.ErrExecutionFailed},
	}

	mem, _ := storage.NewMemoryStorage()
	context, _ := state.NewWorldState(dpos.NewDpos(), mem)

	addr, _ := core.NewAddressFromPublicKey([]byte{
		2, 3, 5, 7, 2, 3, 5, 7, 2, 3, 5, 7, 2, 3, 5, 7, 2, 3, 5, 7,
		2, 3, 5, 7, 2, 3, 5, 7, 2, 3, 5, 7, 2, 3, 5, 7, 2, 3, 5, 7,
		2, 3, 5, 7, 2, 3, 5, 7, 2, 3, 5, 7, 2, 3, 5, 7, 2, 3, 5, 7, 1, 2, 4, 5, 3})
	owner, err := context.GetOrCreateUserAccount(addr.Bytes())
	assert.Nil(t, err)
	owner.AddBalance(newUint128FromIntWrapper(1000000000))

	addr, _ = core.NewContractAddressFromData([]byte{1, 2, 3, 5, 7}, []byte{1, 2, 3, 5, 7})
	contract, _ := context.CreateContractAccount(addr.Bytes(), nil, &corepb.ContractMeta{Version: "1.1.0"})

	ctx, err := NewContext(mockBlock(), mockTransaction(), contract, context)

	data, err := ioutil.ReadFile("test/mixin.js")
	assert.Nil(t, err, "filepath read error")
	sourceType := "js"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx.tx = mockNormalTransaction("n1FkntVUMPAsESuCAAPK711omQk19JotBjM", "n1JNHZJEUvfBYfjDRD14Q73FX62nJAzXkMR", tt.value)
			engine := NewV8Engine(ctx)
			engine.SetExecutionLimits(10000, 100000000)
			_, err := engine.Call(string(data), sourceType, tt.function, tt.args)
			assert.Equal(t, tt.err, err)
			engine.Dispose()
		})
	}
}

func TestThreadStackOverflow(t *testing.T) {
	core.NebCompatibility = core.NewCompatibilityLocal()
	tests := []struct {
		filepath    string
		expectedErr error
	}{
		{"test/test_stack_overflow.js", core.ErrExecutionFailed},
	}
	// lockx := sync.RWMutex{}

	for _, tt := range tests {
		t.Run(tt.filepath, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.filepath)
			assert.Nil(t, err, "filepath read error")
			for j := 0; j < 10; j++ {

				var wg sync.WaitGroup
				for i := 0; i < 5; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()

						mem, _ := storage.NewMemoryStorage()
						context, _ := state.NewWorldState(dpos.NewDpos(), mem)
						owner, err := context.GetOrCreateUserAccount([]byte("n1FkntVUMPAsESuCAAPK711omQk19JotBjM"))
						assert.Nil(t, err)
						owner.AddBalance(newUint128FromIntWrapper(1000000000))
						contract, err := context.CreateContractAccount([]byte("n1JNHZJEUvfBYfjDRD14Q73FX62nJAzXkMR"), nil, &corepb.ContractMeta{Version: "1.1.0"})
						assert.Nil(t, err)

						ctx, err := NewContext(mockBlock(), mockTransaction(), contract, context)
						engine := NewV8Engine(ctx)
						engine.SetExecutionLimits(100000000, 10000000)
						_, err = engine.DeployAndInit(string(data), "js", "")
						fmt.Printf("err:%v", err)
						// _, err = engine.RunScriptSource("", 0)
						assert.Equal(t, tt.expectedErr, err)
						engine.Dispose()

					}()
					// }
				}
				wg.Wait()
			}

		})
	}
}
func TestGetRandomBySingle(t *testing.T) {
	core.NebCompatibility = core.NewCompatibilityLocal()
	type TransferTest struct {
		to     string
		result bool
		value  string
	}

	tests := []struct {
		test         string
		contractPath string
		sourceType   string
		name         string
		from         string
	}{
		{"getRandomBySingle", "./test/inner_call_tests/test_inner_transaction.js", "js", "getRandomSingle",
			"n1FkntVUMPAsESuCAAPK711omQk19JotBjM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.contractPath)
			assert.Nil(t, err, "contract path read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewWorldState(dpos.NewDpos(), mem)
			owner, err := context.GetOrCreateUserAccount([]byte(tt.from))
			assert.Nil(t, err)
			owner.AddBalance(newUint128FromIntWrapper(10000000))

			// prepare the contract.
			contractAddr, err := core.AddressParse(contractStr)
			contract, _ := context.CreateContractAccount(contractAddr.Bytes(), nil, &corepb.ContractMeta{Version: "1.1.0"})
			contract.AddBalance(newUint128FromIntWrapper(5))

			// parepare env, block & transactions.
			tx := mockNormalTransaction(tt.from, "n1TV3sU6jyzR4rJ1D7jCAmtVGSntJagXZHC", "0")
			ctx, err := NewContext(mockBlock(), tx, contract, context)

			// execute.
			engine := NewV8Engine(ctx)
			engine.SetExecutionLimits(10000, 100000000)
			args := fmt.Sprintf("[]")
			_, err = engine.DeployAndInit(string(data), tt.sourceType, args)
			assert.Nil(t, err)
			engine.Dispose()

			// call name.
			engine = NewV8Engine(ctx)
			engine.SetExecutionLimits(10000, 100000000)
			rand, err := engine.Call(string(data), tt.sourceType, "getRandomSingle", "")
			fmt.Printf("rand:%v\n", rand)
			assert.Nil(t, err)
			// var nameStr string
			// err = json.Unmarshal([]byte(name), &nameStr)
			// assert.Nil(t, err)
			// assert.Equal(t, tt.name, nameStr)
			engine.Dispose()

		})
	}
}

func TestParallelBug(t *testing.T) {
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
					"test/test_parallel_bug.js",
					"js",
					"[1]",
				},
				contract{
					"test/test_parallel_bug.js",
					"js",
					"[2]",
				},
				contract{
					"test/test_parallel_bug.js",
					"js",
					"[3]",
				},
				contract{
					"test/test_parallel_bug.js",
					"js",
					"[4]",
				},
			},
			[]call{
				call{
					"test",
					"[]",
					[]string{"1"},
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
		var nonce uint64 = 0
		for i := 0; i < 8; i++ {
			for _, v := range tt.contracts {
				nonce = nonce + 1
				data, err := ioutil.ReadFile(v.contractPath)
				assert.Nil(t, err, "contract path read error")
				source := string(data)
				sourceType := "js"
				argsDeploy := v.initArgs
				deploy, _ := core.NewDeployPayload(source, sourceType, argsDeploy)
				payloadDeploy, _ := deploy.ToBytes()

				value, _ := util.NewUint128FromInt(0)
				gasLimit, _ := util.NewUint128FromInt(200000)
				txDeploy, err := core.NewTransaction(neb.BlockChain().ChainID(), from, from, value, nonce, core.TxPayloadDeployType, payloadDeploy, core.TransactionGasPrice, gasLimit)
				assert.Nil(t, err)
				assert.Nil(t, manager.SignTransaction(from, txDeploy))
				txs = append(txs, txDeploy)

				contractAddr, err := txDeploy.GenerateContractAddress()
				assert.Nil(t, err)
				contractsAddr = append(contractsAddr, contractAddr.String())
			}
		}
		// })

		// mint for contract deploy
		mintBlockWithDuration(t, neb, txs, 2)

		for _, v := range contractsAddr {
			contract, err := core.AddressParse(v)
			assert.Nil(t, err)
			_, err = neb.BlockChain().TailBlock().CheckContract(contract)
			assert.Nil(t, err)
		}

		// Contract1 := contractsAddr[1]
		// Contract2 := contractsAddr[2]
		callPayload, _ := core.NewCallPayload(call.function, "[]")
		payloadCall, _ := callPayload.ToBytes()

		value, _ := util.NewUint128FromInt(6)
		gasLimit, _ := util.NewUint128FromInt(200000)

		callTxs := []*core.Transaction{}
		for i := 0; i < 4; i++ {
			ContractAddr1, err := core.AddressParse(contractsAddr[0+i*8])
			txCall1, err := core.NewTransaction(neb.BlockChain().ChainID(), addrs[1], ContractAddr1, value,
				uint64(1+i), core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
			callTxs = append(callTxs, txCall1)
			assert.Nil(t, err)
			assert.Nil(t, manager.SignTransaction(txCall1.From(), txCall1))

			ContractAddr2, err := core.AddressParse(contractsAddr[1+i*8])
			from = addrs[2]
			txCall2, err := core.NewTransaction(neb.BlockChain().ChainID(), addrs[2], ContractAddr2, value,
				uint64(1+i), core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
			callTxs = append(callTxs, txCall2)
			assert.Nil(t, err)
			assert.Nil(t, manager.SignTransaction(txCall2.From(), txCall2))

			ContractAddr3, err := core.AddressParse(contractsAddr[2+i*8])
			txCall3, err := core.NewTransaction(neb.BlockChain().ChainID(), addrs[3], ContractAddr3, value,
				uint64(1+i), core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
			callTxs = append(callTxs, txCall3)
			assert.Nil(t, err)
			assert.Nil(t, manager.SignTransaction(txCall3.From(), txCall3))

			ContractAddr4, err := core.AddressParse(contractsAddr[3+i*8])
			txCall4, err := core.NewTransaction(neb.BlockChain().ChainID(), addrs[4], ContractAddr4, value,
				uint64(1+i), core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
			callTxs = append(callTxs, txCall4)
			assert.Nil(t, err)
			assert.Nil(t, manager.SignTransaction(txCall4.From(), txCall4))

			ContractAddr5, err := core.AddressParse(contractsAddr[4+i*8])
			txCall5, err := core.NewTransaction(neb.BlockChain().ChainID(), addrs[5], ContractAddr5, value,
				uint64(1+i), core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
			callTxs = append(callTxs, txCall5)
			assert.Nil(t, err)
			assert.Nil(t, manager.SignTransaction(txCall5.From(), txCall5))

			ContractAddr6, err := core.AddressParse(contractsAddr[5+i*8])
			txCall6, err := core.NewTransaction(neb.BlockChain().ChainID(), addrs[6], ContractAddr6, value,
				uint64(1+i), core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
			callTxs = append(callTxs, txCall6)
			assert.Nil(t, err)
			assert.Nil(t, manager.SignTransaction(txCall6.From(), txCall6))

			ContractAddr7, err := core.AddressParse(contractsAddr[6+i*8])
			txCall7, err := core.NewTransaction(neb.BlockChain().ChainID(), addrs[7], ContractAddr7, value,
				uint64(1+i), core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
			callTxs = append(callTxs, txCall7)
			assert.Nil(t, err)
			assert.Nil(t, manager.SignTransaction(txCall7.From(), txCall7))

			ContractAddr8, err := core.AddressParse(contractsAddr[7+i*8])
			txCall8, err := core.NewTransaction(neb.BlockChain().ChainID(), addrs[8], ContractAddr8, value,
				uint64(1+i), core.TxPayloadCallType, payloadCall, core.TransactionGasPrice, gasLimit)
			callTxs = append(callTxs, txCall8)
			assert.Nil(t, err)
			assert.Nil(t, manager.SignTransaction(txCall8.From(), txCall8))
		}
		core.PackedParallelNum = 8

		// mint for contract call
		mintBlockWithDuration(t, neb, callTxs, 2)

		// check
		tail := neb.BlockChain().TailBlock()
		for _, v := range callTxs {
			events, err := tail.FetchEvents(v.Hash())
			assert.Nil(t, err)
			for _, event := range events {
				txEvent := core.TransactionEvent{}
				err = json.Unmarshal([]byte(event.Data), &txEvent)
				assert.Nil(t, err)
				assert.Equal(t, txEvent.GasUsed, "20142")
				// fmt.Println("==============", event.Data)
			}
		}
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

func TestNebulasRank(t *testing.T) {
	core.NebCompatibility = core.NewCompatibilityLocal()
	tests := []struct {
		name         string
		contractPath string
		function     string
		args         string
		expectedErr  error
		result       string
	}{
		{"case1", "./test/test_nebulas_rank.js", "rank", "[\"addr\"]", core.ErrExecutionFailed, "Address is invalid"},
		{"case2", "./test/test_nebulas_rank.js", "rank", "[\"n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE\"]", core.ErrExecutionFailed, "Failed to find nr value"},
		{"case3", "./test/test_nebulas_rank.js", "rank", "[\"n1FkntVUMPAsESuCAAPK711omQk19JotBjM\"]", nil, "\"10.09\""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.contractPath)
			assert.Nil(t, err, "filepath read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewWorldState(dpos.NewDpos(), mem)
			owner, err := context.GetOrCreateUserAccount([]byte("account1"))
			assert.Nil(t, err)
			owner.AddBalance(newUint128FromIntWrapper(1000000000))
			contract, _ := context.CreateContractAccount([]byte("account2"), nil, &corepb.ContractMeta{Version: "1.1.0"})
			ctx, err := NewContext(mockBlockForLib(2), mockTransaction(), contract, context)

			engine := NewV8Engine(ctx)
			engine.SetExecutionLimits(900000, 10000000)
			result, err := engine.Call(string(data), "js", tt.function, tt.args)
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.result, result)
			engine.Dispose()
		})
	}
}

func TestNebulasRankSummary(t *testing.T) {
	core.NebCompatibility = core.NewCompatibilityLocal()
	tests := []struct {
		name         string
		contractPath string
		height       uint64
		function     string
		expectedErr  error
		result       string
	}{
		{"case1", "./test/test_nebulas_rank.js", 100, "summary", core.ErrExecutionFailed, "Failed to get nr summary"},
		{"case2", "./test/test_nebulas_rank.js", 1200, "summary", nil, "{\"start_height\":\"1\",\"end_height\":\"500\",\"version\":\"0\",\"sum\":{\"in_outs\":\"123123\",\"score\":\"100\"},\"err\":\"\"}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			data, err := ioutil.ReadFile(tt.contractPath)
			assert.Nil(t, err, "filepath read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewWorldState(dpos.NewDpos(), mem)
			owner, err := context.GetOrCreateUserAccount([]byte("account1"))
			assert.Nil(t, err)
			owner.AddBalance(newUint128FromIntWrapper(1000000000))
			contract, _ := context.CreateContractAccount([]byte("account2"), nil, &corepb.ContractMeta{Version: "1.1.0"})
			ctx, err := NewContext(mockBlockForLib(tt.height), mockTransaction(), contract, context)

			engine := NewV8Engine(ctx)
			engine.SetExecutionLimits(900000, 10000000)
			result, err := engine.Call(string(data), "js", tt.function, "")
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.result, result)
			engine.Dispose()
		})
	}
}
