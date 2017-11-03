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
	"io/ioutil"
	"testing"

	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/stretchr/testify/assert"
)

func TestRunScriptSource(t *testing.T) {
	tests := []struct {
		filepath string
	}{
		{"test/test_console.js"},
		{"test/test_storage_handlers.js"},
		{"test/test_storage_class.js"},
		{"test/test_storage.js"},
		{"test/test_ERC20.js"},
	}

	for _, tt := range tests {
		t.Run(tt.filepath, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.filepath)
			assert.Nil(t, err, "filepath read error")

			mem, _ := storage.NewMemoryStorage()
			balanceTrie, _ := trie.NewBatchTrie(nil, mem)
			lcsTrie, _ := trie.NewBatchTrie(nil, mem)
			gcsTrie, _ := trie.NewBatchTrie(nil, mem)

			engine := NewV8Engine(balanceTrie, lcsTrie, gcsTrie)
			err = engine.RunScriptSource(string(data))
			engine.Dispose()

			assert.Nil(t, err)
		})
	}
}

func TestDeployAndInitAndCall(t *testing.T) {
	tests := []struct {
		name         string
		contractPath string
		init_args    string
		verify_args  string
	}{
		{"deploy sample_contract.js", "test/sample_contract.js", "[\"TEST001\", 123,[{\"name\":\"robin\",\"count\":2},{\"name\":\"roy\",\"count\":3},{\"name\":\"leon\",\"count\":4}]]", "[\"TEST001\", 123,[{\"name\":\"robin\",\"count\":2},{\"name\":\"roy\",\"count\":3},{\"name\":\"leon\",\"count\":4}]]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.contractPath)
			assert.Nil(t, err, "contract path read error")

			mem, _ := storage.NewMemoryStorage()
			balanceTrie, _ := trie.NewBatchTrie(nil, mem)
			lcsTrie, _ := trie.NewBatchTrie(nil, mem)
			gcsTrie, _ := trie.NewBatchTrie(nil, mem)

			engine := NewV8Engine(balanceTrie, lcsTrie, gcsTrie)
			err = engine.DeployAndInit(string(data), tt.init_args)
			assert.Nil(t, err)
			engine.Dispose()

			engine = NewV8Engine(balanceTrie, lcsTrie, gcsTrie)
			err = engine.Call(string(data), "dump", "")
			assert.Nil(t, err)
			engine.Dispose()

			engine = NewV8Engine(balanceTrie, lcsTrie, gcsTrie)
			err = engine.Call(string(data), "verify", tt.verify_args)
			assert.Nil(t, err)
			engine.Dispose()

			// force error.
			mem, _ = storage.NewMemoryStorage()
			balanceTrie, _ = trie.NewBatchTrie(nil, mem)
			lcsTrie, _ = trie.NewBatchTrie(nil, mem)
			gcsTrie, _ = trie.NewBatchTrie(nil, mem)

			engine = NewV8Engine(balanceTrie, lcsTrie, gcsTrie)
			err = engine.Call(string(data), "verify", tt.verify_args)
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
			balanceTrie, _ := trie.NewBatchTrie(nil, mem)
			lcsTrie, _ := trie.NewBatchTrie(nil, mem)
			gcsTrie, _ := trie.NewBatchTrie(nil, mem)

			engine := NewV8Engine(balanceTrie, lcsTrie, gcsTrie)
			err = engine.Call(string(data), tt.function, tt.args)
			assert.Equal(t, tt.expectedErr, err)
			engine.Dispose()
		})
	}
}
