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
	"flag"
	"io/ioutil"
	"os"
	"sync"
	"testing"

	"github.com/nebulasio/go-nebulas/state"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util/logging"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	logging.EnableFuncNameLogger()

	flag.Parse()
	os.Exit(m.Run())
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
		{"test/test_ERC20.js", nil},
		{"test/test_eval.js", ErrExecutionFailed},
	}

	for _, tt := range tests {
		t.Run(tt.filepath, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.filepath)
			assert.Nil(t, err, "filepath read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewAccountState(nil, mem)
			owner := context.GetOrCreateUserAccount([]byte("account1"))
			contract, _ := context.CreateContractAccount([]byte("account2"), nil)
			engine := NewV8Engine(owner, contract, context)
			err = engine.RunScriptSource(string(data))
			assert.Equal(t, tt.expectedErr, err)
			engine.Dispose()
		})
	}
}

func TestDeployAndInitAndCall(t *testing.T) {
	tests := []struct {
		name         string
		contractPath string
		initArgs     string
		verifyArgs   string
	}{
		{"deploy sample_contract.js", "test/sample_contract.js", "[\"TEST001\", 123,[{\"name\":\"robin\",\"count\":2},{\"name\":\"roy\",\"count\":3},{\"name\":\"leon\",\"count\":4}]]", "[\"TEST001\", 123,[{\"name\":\"robin\",\"count\":2},{\"name\":\"roy\",\"count\":3},{\"name\":\"leon\",\"count\":4}]]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.contractPath)
			assert.Nil(t, err, "contract path read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewAccountState(nil, mem)
			owner := context.GetOrCreateUserAccount([]byte("account1"))
			contract, _ := context.CreateContractAccount([]byte("account2"), nil)

			engine := NewV8Engine(owner, contract, context)
			err = engine.DeployAndInit(string(data), tt.initArgs)
			assert.Nil(t, err)
			engine.Dispose()

			engine = NewV8Engine(owner, contract, context)
			err = engine.Call(string(data), "dump", "")
			assert.Nil(t, err)
			engine.Dispose()

			engine = NewV8Engine(owner, contract, context)
			err = engine.Call(string(data), "verify", tt.verifyArgs)
			assert.Nil(t, err)
			engine.Dispose()

			// force error.
			mem, _ = storage.NewMemoryStorage()
			context, _ = state.NewAccountState(nil, mem)
			owner = context.GetOrCreateUserAccount([]byte("account1"))
			contract, _ = context.CreateContractAccount([]byte("account2"), nil)

			engine = NewV8Engine(owner, contract, context)
			err = engine.Call(string(data), "verify", tt.verifyArgs)
			assert.NotNil(t, err)
			engine.Dispose()

		})
	}
}

func TestFunctionNameCheck(t *testing.T) {
	tests := []struct {
		function    string
		expectedErr error
		args        string
	}{
		{"init", ErrInvalidFunctionName, ""},
		{"9dump", ErrInvalidFunctionName, ""},
		{"$dump", ErrInvalidFunctionName, ""},
		{"dump", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.function, func(t *testing.T) {
			data, err := ioutil.ReadFile("test/sample_contract.js")
			assert.Nil(t, err, "contract path read error")

			mem, _ := storage.NewMemoryStorage()
			context, _ := state.NewAccountState(nil, mem)
			owner := context.GetOrCreateUserAccount([]byte("account1"))
			contract, _ := context.CreateContractAccount([]byte("account2"), nil)

			engine := NewV8Engine(owner, contract, context)
			err = engine.Call(string(data), tt.function, tt.args)
			assert.Equal(t, tt.expectedErr, err)
			engine.Dispose()
		})
	}
}
func TestMultiEngine(t *testing.T) {
	mem, _ := storage.NewMemoryStorage()
	context, _ := state.NewAccountState(nil, mem)
	owner := context.GetOrCreateUserAccount([]byte("account1"))
	contract, _ := context.CreateContractAccount([]byte("account2"), nil)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		idx := i
		go func() {
			defer wg.Done()
			engine := NewV8Engine(owner, contract, context)
			defer engine.Dispose()

			err := engine.RunScriptSource("console.log('running.');")
			log.Infof("run script %d; err %v", idx, err)
			assert.Nil(t, err)
		}()
	}
	wg.Wait()
}
