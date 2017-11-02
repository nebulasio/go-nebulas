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
	"testing"

	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/stretchr/testify/assert"
)

func TestRunScriptSource(t *testing.T) {
	mem, _ := storage.NewMemoryStorage()
	balanceTrie, _ := trie.NewBatchTrie(nil, mem)
	lcsTrie, _ := trie.NewBatchTrie(nil, mem)
	gcsTrie, _ := trie.NewBatchTrie(nil, mem)

	engine := NewV8Engine(balanceTrie, lcsTrie, gcsTrie)
	defer engine.Dispose()

	err := engine.RunScriptSource("console.log('wow.');")
	assert.Nil(t, err)
}

func TestRunScriptSource1(t *testing.T) {
	mem, _ := storage.NewMemoryStorage()
	balanceTrie, _ := trie.NewBatchTrie(nil, mem)
	lcsTrie, _ := trie.NewBatchTrie(nil, mem)
	gcsTrie, _ := trie.NewBatchTrie(nil, mem)

	engine := NewV8Engine(balanceTrie, lcsTrie, gcsTrie)
	defer engine.Dispose()

	err := engine.RunScriptSource("console.log('wow.');")
	assert.Nil(t, err)
}
