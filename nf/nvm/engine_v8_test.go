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
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

type mockBlock struct {
}

func (m *mockBlock) CoinbaseHash() byteutils.Hash {
	return []byte("8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf")
}

func (m *mockBlock) Nonce() uint64 {
	return 1
}

func (m *mockBlock) Hash() byteutils.Hash {
	return []byte("c7174759e86c59dcb7df87def82f61eb")
}

func (m *mockBlock) Height() uint64 {
	return 2
}

func (m *mockBlock) VerifyAddress(str string) bool {
	return true
}

func (m *mockBlock) RecordEvent(txHash byteutils.Hash, topic, data string) error {
	return nil
}

func (m *mockBlock) SerializeTxByHash(hash byteutils.Hash) (proto.Message, error) {
	from, _ := byteutils.FromHex("8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf")
	to, _ := byteutils.FromHex("22ac3a9a2b1c31b7a9084e46eae16e761f83f02324092b09")
	value, _ := util.NewUint128FromString("10").ToFixedSizeByteSlice()
	gasPrice, _ := util.NewUint128FromString("1").ToFixedSizeByteSlice()
	gasLimit, _ := util.NewUint128FromString("100").ToFixedSizeByteSlice()
	block := &corepb.Transaction{
		From:     from,
		To:       to,
		Value:    value,
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		Hash:     hash,
	}
	return proto.Message(block), nil
}

func testContextBlock() Block {
	return new(mockBlock)
}

func testContextTransaction() *ContextTransaction {
	return &ContextTransaction{
		From:     "8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf",
		To:       "22ac3a9a2b1c31b7a9084e46eae16e761f83f02324092b09",
		Value:    "5",
		Nonce:    3,
		Hash:     "c7174759e86c59dcb7df87def82f61eb",
		GasPrice: util.NewUint128FromInt(1).String(),
		GasLimit: util.NewUint128FromInt(10).String(),
	}
}

func TestRunScriptSource(t *testing.T) {
	tests := []struct {
		filepath    string
		expectedErr error
	}{
		{"test/test_require.js", nil},
		{"test/test_console.js", nil},
		{"test/test_storage_handlers.js", nil},
		{"test/test_storage_class.js", nil},
		{"test/test_storage.js", nil},
		{"test/test_eval.js", ErrExecutionFailed},
	}

	for _, tt := range tests {
		t.Run(tt.filepath, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.filepath)
			assert.Nil(t, err, "filepath read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewAccountState(nil, mem)
			owner := context.GetOrCreateUserAccount([]byte("account1"))
			owner.AddBalance(util.NewUint128FromInt(1000000000))
			contract, _ := context.CreateContractAccount([]byte("account2"), nil)
			ctx := NewContext(testContextBlock(), testContextTransaction(), owner, contract, context)

			engine := NewV8Engine(ctx)
			engine.SetExecutionLimits(900000, 10000000)
			err = engine.RunScriptSource(string(data), 0)
			assert.Equal(t, tt.expectedErr, err)
			engine.Dispose()
		})
	}
}

func TestRunScriptSourceInModule(t *testing.T) {
	tests := []struct {
		filepath    string
		expectedErr error
	}{
		{"./test/test_require.js", nil},
		{"./test/test_console.js", nil},
		{"./test/test_storage_handlers.js", nil},
		{"./test/test_storage_class.js", nil},
		{"./test/test_storage.js", nil},
		{"./test/test_ERC20.js", nil},
		{"./test/test_eval.js", ErrExecutionFailed},
	}

	for _, tt := range tests {
		t.Run(tt.filepath, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.filepath)
			assert.Nil(t, err, "filepath read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewAccountState(nil, mem)
			owner := context.GetOrCreateUserAccount([]byte("account1"))
			owner.AddBalance(util.NewUint128FromInt(1000000000))
			contract, _ := context.CreateContractAccount([]byte("account2"), nil)
			ctx := NewContext(testContextBlock(), testContextTransaction(), owner, contract, context)

			engine := NewV8Engine(ctx)
			engine.SetExecutionLimits(100000, 10000000)
			engine.AddModule(tt.filepath, string(data), 0)
			runnableSource := fmt.Sprintf("require(\"%s\");", tt.filepath)
			err = engine.RunScriptSource(runnableSource, 0)
			assert.Equal(t, tt.expectedErr, err)
			engine.Dispose()
		})
	}
}

func TestRunScriptSourceWithLimits(t *testing.T) {
	tests := []struct {
		filepath                      string
		limitsOfExecutionInstructions uint64
		limitsOfTotalMemorySize       uint64
		expectedErr                   error
	}{
		{"test/test_oom_1.js", 100000, 0, ErrInsufficientGas},
		{"test/test_oom_1.js", 0, 50000000, ErrExceedMemoryLimits},
		{"test/test_oom_1.js", 1000000, 50000000, ErrInsufficientGas},
		{"test/test_oom_1.js", 5000000, 7000000, ErrExceedMemoryLimits},

		{"test/test_oom_2.js", 100000, 0, ErrInsufficientGas},
		{"test/test_oom_2.js", 0, 8000000, ErrExceedMemoryLimits},
		{"test/test_oom_2.js", 10000000, 10000000, ErrInsufficientGas},
		{"test/test_oom_2.js", 10000000, 7000000, ErrExceedMemoryLimits},
	}

	for _, tt := range tests {
		t.Run(tt.filepath, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.filepath)
			assert.Nil(t, err, "filepath read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewAccountState(nil, mem)
			owner := context.GetOrCreateUserAccount([]byte("account1"))
			owner.AddBalance(util.NewUint128FromInt(100000))
			contract, _ := context.CreateContractAccount([]byte("account2"), nil)
			ctx := NewContext(testContextBlock(), testContextTransaction(), owner, contract, context)

			// direct run.
			(func() {
				engine := NewV8Engine(ctx)
				engine.SetExecutionLimits(tt.limitsOfExecutionInstructions, tt.limitsOfTotalMemorySize)
				err = engine.RunScriptSource(string(data), 0)
				assert.Equal(t, tt.expectedErr, err)
				engine.Dispose()
			})()

			// modularized run.
			(func() {
				moduleID := fmt.Sprintf("./%s", tt.filepath)
				runnableSource := fmt.Sprintf("require(\"%s\");", moduleID)

				engine := NewV8Engine(ctx)
				engine.SetExecutionLimits(tt.limitsOfExecutionInstructions, tt.limitsOfTotalMemorySize)
				engine.AddModule(moduleID, string(data), 0)
				err = engine.RunScriptSource(runnableSource, 0)
				assert.Equal(t, tt.expectedErr, err)
				engine.Dispose()
			})()
		})
	}
}

func TestRunScriptSourceTimeout(t *testing.T) {
	tests := []struct {
		filepath string
	}{
		{"test/test_infinite_loop.js"},
	}

	for _, tt := range tests {
		t.Run(tt.filepath, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.filepath)
			assert.Nil(t, err, "filepath read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewAccountState(nil, mem)
			owner := context.GetOrCreateUserAccount([]byte("account1"))
			contract, _ := context.CreateContractAccount([]byte("account2"), nil)
			ctx := NewContext(testContextBlock(), testContextTransaction(), owner, contract, context)

			// direct run.
			(func() {
				engine := NewV8Engine(ctx)
				err = engine.RunScriptSource(string(data), 0)
				assert.Equal(t, ErrExecutionTimeout, err)
				engine.Dispose()
			})()

			// modularized run.
			(func() {
				moduleID := fmt.Sprintf("./%s", tt.filepath)
				runnableSource := fmt.Sprintf("require(\"%s\");", moduleID)

				engine := NewV8Engine(ctx)
				engine.AddModule(moduleID, string(data), 0)
				err = engine.RunScriptSource(runnableSource, 0)
				assert.Equal(t, ErrExecutionTimeout, err)
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
			context, _ := state.NewAccountState(nil, mem)
			owner := context.GetOrCreateUserAccount([]byte("account1"))
			owner.AddBalance(util.NewUint128FromInt(10000000))
			contract, _ := context.CreateContractAccount([]byte("account2"), nil)

			ctx := NewContext(testContextBlock(), testContextTransaction(), owner, contract, context)
			engine := NewV8Engine(ctx)
			engine.SetExecutionLimits(10000, 10000000)
			err = engine.DeployAndInit(string(data), tt.sourceType, tt.initArgs)
			assert.Nil(t, err)
			engine.Dispose()

			engine = NewV8Engine(ctx)
			engine.SetExecutionLimits(10000, 10000000)
			err = engine.Call(string(data), tt.sourceType, "dump", "")
			assert.Nil(t, err)
			engine.Dispose()

			engine = NewV8Engine(ctx)
			engine.SetExecutionLimits(10000, 10000000)
			err = engine.Call(string(data), tt.sourceType, "verify", tt.verifyArgs)
			assert.Nil(t, err)
			engine.Dispose()

			// force error.
			mem, _ = storage.NewMemoryStorage()
			context, _ = state.NewAccountState(nil, mem)
			owner = context.GetOrCreateUserAccount([]byte("account1"))
			contract, _ = context.CreateContractAccount([]byte("account2"), nil)

			ctx = NewContext(testContextBlock(), testContextTransaction(), owner, contract, context)
			engine = NewV8Engine(ctx)
			engine.SetExecutionLimits(10000, 10000000)
			err = engine.Call(string(data), tt.sourceType, "verify", tt.verifyArgs)
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
			context, _ := state.NewAccountState(nil, mem)
			owner := context.GetOrCreateUserAccount([]byte("account1"))
			owner.AddBalance(util.NewUint128FromInt(10000000))
			contract, _ := context.CreateContractAccount([]byte("account2"), nil)
			ctx := NewContext(testContextBlock(), testContextTransaction(), owner, contract, context)

			// deploy and init.
			engine := NewV8Engine(ctx)
			engine.SetExecutionLimits(1000, 10000000)
			err = engine.DeployAndInit(string(data), tt.sourceType, tt.initArgs)
			assert.Nil(t, err)
			engine.Dispose()

			// call.
			for _, fields := range tt.calls {
				engine = NewV8Engine(ctx)
				engine.SetExecutionLimits(1000, 10000000)
				err = engine.Call(string(data), tt.sourceType, fields.function, fields.args)
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
		{"9dump", ErrDisallowCallPrivateFunction, ""},
		{"_dump", ErrDisallowCallPrivateFunction, ""},
	}

	for _, tt := range tests {
		t.Run(tt.function, func(t *testing.T) {
			data, err := ioutil.ReadFile("test/sample_contract.js")
			sourceType := "js"
			assert.Nil(t, err, "contract path read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewAccountState(nil, mem)
			owner := context.GetOrCreateUserAccount([]byte("account1"))
			owner.AddBalance(util.NewUint128FromInt(1000000))
			contract, _ := context.CreateContractAccount([]byte("account2"), nil)
			ctx := NewContext(testContextBlock(), testContextTransaction(), owner, contract, context)

			engine := NewV8Engine(ctx)
			engine.SetExecutionLimits(1000, 10000000)
			err = engine.Call(string(data), sourceType, tt.function, tt.args)
			assert.Equal(t, tt.expectedErr, err)
			engine.Dispose()
		})
	}
}

func TestMultiEngine(t *testing.T) {
	mem, _ := storage.NewMemoryStorage()
	context, _ := state.NewAccountState(nil, mem)
	owner := context.GetOrCreateUserAccount([]byte("account1"))
	owner.AddBalance(util.NewUint128FromInt(1000000))
	contract, _ := context.CreateContractAccount([]byte("account2"), nil)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			ctx := NewContext(testContextBlock(), testContextTransaction(), owner, contract, context)

			engine := NewV8Engine(ctx)
			engine.SetExecutionLimits(1000, 10000000)
			defer engine.Dispose()

			err := engine.RunScriptSource("console.log('running.');", 0)
			assert.Nil(t, err)
		}()
	}
	wg.Wait()
}

func TestInstructionCounterTestSuite(t *testing.T) {
	tests := []struct {
		filepath    string
		expectedErr error
	}{
		{"./test/instruction_counter_tests/redefine1.js", ErrInjectTracingInstructionFailed},
		{"./test/instruction_counter_tests/redefine2.js", ErrInjectTracingInstructionFailed},
		{"./test/instruction_counter_tests/redefine3.js", ErrInjectTracingInstructionFailed},
		{"./test/instruction_counter_tests/redefine4.js", ErrExecutionFailed},
		{"./test/instruction_counter_tests/function.js", nil},
		{"./test/instruction_counter_tests/if.js", nil},
		{"./test/instruction_counter_tests/switch.js", nil},
		{"./test/instruction_counter_tests/for.js", nil},
		{"./test/instruction_counter_tests/with.js", nil},
		{"./test/instruction_counter_tests/while.js", nil},
		{"./test/instruction_counter_tests/throw.js", nil},
		{"./test/instruction_counter_tests/switch.js", nil},
		{"./test/instruction_counter_tests/condition_operator.js", nil},
		{"./test/instruction_counter_tests/storage_usage.js", nil},
		{"./test/instruction_counter_tests/event_usage.js", nil},
	}

	for _, tt := range tests {
		t.Run(tt.filepath, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.filepath)
			assert.Nil(t, err, "filepath read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewAccountState(nil, mem)
			owner := context.GetOrCreateUserAccount([]byte("account1"))
			owner.AddBalance(util.NewUint128FromInt(1000000000))
			contract, _ := context.CreateContractAccount([]byte("account2"), nil)
			ctx := NewContext(testContextBlock(), testContextTransaction(), owner, contract, context)

			moduleID := tt.filepath
			runnableSource := fmt.Sprintf("require(\"%s\");", moduleID)

			engine := NewV8Engine(ctx)
			engine.enableLimits = true
			err = engine.AddModule(moduleID, string(data), 0)
			if err != nil {
				assert.Equal(t, tt.expectedErr, err)
			} else {
				err = engine.RunScriptSource(runnableSource, 0)
				assert.Equal(t, tt.expectedErr, err)
			}
			engine.Dispose()
		})
	}
}

func TestTypeScriptExecution(t *testing.T) {
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
			context, _ := state.NewAccountState(nil, mem)
			owner := context.GetOrCreateUserAccount([]byte("account1"))
			owner.AddBalance(util.NewUint128FromInt(1000000000))
			contract, _ := context.CreateContractAccount([]byte("account2"), nil)
			ctx := NewContext(testContextBlock(), testContextTransaction(), owner, contract, context)

			moduleID := tt.filepath
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
				err = engine.RunScriptSource(runnableSource, 0)
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestRunMozillaJSTestSuite(t *testing.T) {
	mem, _ := storage.NewMemoryStorage()
	context, _ := state.NewAccountState(nil, mem)
	owner := context.GetOrCreateUserAccount([]byte("account1"))
	owner.AddBalance(util.NewUint128FromInt(1000000000))

	contract, _ := context.CreateContractAccount([]byte("account2"), nil)
	ctx := NewContext(testContextBlock(), testContextTransaction(), owner, contract, context)

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
			err = engine.RunScriptSource(buf.String(), 0)
			assert.Nil(t, err)
		}
	}

	runTest("test/mozilla_js_tests", "")
}

func TestBlockChain(t *testing.T) {
	tests := []struct {
		filepath    string
		expectedErr error
	}{
		{"test/test_blockchain.js", nil},
	}

	for _, tt := range tests {
		t.Run(tt.filepath, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.filepath)
			assert.Nil(t, err, "filepath read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewAccountState(nil, mem)
			owner := context.GetOrCreateUserAccount([]byte("8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf"))
			owner.AddBalance(util.NewUint128FromInt(1000000000))
			contract, _ := context.CreateContractAccount([]byte("16464b93292d7c99099d4d982a05140f12779f5e299d6eb4"), nil)

			ctx := NewContext(testContextBlock(), testContextTransaction(), owner, contract, context)
			engine := NewV8Engine(ctx)
			engine.SetExecutionLimits(100000, 10000000)
			err = engine.RunScriptSource(string(data), 0)
			assert.Equal(t, tt.expectedErr, err)
			engine.Dispose()
		})
	}
}

func TestBankVaultContract(t *testing.T) {
	type TakeoutTest struct {
		args        string
		expectedErr error
	}

	tests := []struct {
		name         string
		contractPath string
		sourceType   string
		saveArgs     string
		takeoutTests []TakeoutTest
	}{
		{"deploy bank_vault_contract.js", "./test/bank_vault_contract.js", "js", "[0]",
			[]TakeoutTest{
				{"[1]", nil},
				{"[5]", ErrExecutionFailed},
				{"[4]", nil},
				{"[1]", ErrExecutionFailed},
			},
		},
		{"deploy bank_vault_contract.ts", "./test/bank_vault_contract.ts", "ts", "[0]",
			[]TakeoutTest{
				{"[1]", nil},
				{"[5]", ErrExecutionFailed},
				{"[4]", nil},
				{"[1]", ErrExecutionFailed},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.contractPath)
			assert.Nil(t, err, "contract path read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewAccountState(nil, mem)
			owner := context.GetOrCreateUserAccount([]byte("account1"))
			owner.AddBalance(util.NewUint128FromInt(10000000))

			// prepare the contract.
			contract, _ := context.CreateContractAccount([]byte("account2"), nil)
			contract.AddBalance(util.NewUint128FromInt(5))

			// parepare env, block & transactions.
			ctx := NewContext(testContextBlock(), testContextTransaction(), owner, contract, context)

			// execute.
			engine := NewV8Engine(ctx)
			engine.SetExecutionLimits(10000, 100000000)
			err = engine.DeployAndInit(string(data), tt.sourceType, "")
			assert.Nil(t, err)
			engine.Dispose()

			engine = NewV8Engine(ctx)
			engine.SetExecutionLimits(10000, 100000000)
			err = engine.Call(string(data), tt.sourceType, "save", tt.saveArgs)
			assert.Nil(t, err)
			engine.Dispose()

			for _, tot := range tt.takeoutTests {
				engine = NewV8Engine(ctx)
				engine.SetExecutionLimits(10000, 100000000)
				err = engine.Call(string(data), tt.sourceType, "takeout", tot.args)
				assert.Equal(t, err, tot.expectedErr)
				engine.Dispose()
			}
		})
	}
}

func TestEvent(t *testing.T) {
	tests := []struct {
		filepath string
	}{
		{"test/test_event.js"},
	}

	for _, tt := range tests {
		t.Run(tt.filepath, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.filepath)
			assert.Nil(t, err, "filepath read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewAccountState(nil, mem)
			owner := context.GetOrCreateUserAccount([]byte("8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf"))
			owner.AddBalance(util.NewUint128FromInt(1000000000))
			contract, _ := context.CreateContractAccount([]byte("16464b93292d7c99099d4d982a05140f12779f5e299d6eb4"), nil)

			ctx := NewContext(testContextBlock(), testContextTransaction(), owner, contract, context)
			engine := NewV8Engine(ctx)
			engine.SetExecutionLimits(100000, 10000000)
			err = engine.RunScriptSource(string(data), 0)
			engine.Dispose()
		})
	}
}
