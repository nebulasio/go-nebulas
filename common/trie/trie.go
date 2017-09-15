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

package trie

// Flag to identify the type of node
type flag int

const (
	ext flag = iota
	leaf
	branch
)

// Node in trie
type node interface {
	key() []byte
	val() [][]byte
	flag() flag
}

// Branch Node [hash_0, hash_1, ..., hash_f, hash]
type brNode struct {
	Val [17][]byte
}

// Extension Node [flag, encodedPath, next hash]
type extNode struct {
	Val [3][]byte
}

// Leaf Node [flag, encodedPath, value]
type leafNode struct {
	Val [3][]byte
}

// Trie is a Merkle Patricia Trie.
type Trie struct {
	root    node
	storage *Storage
}

// New a trie
func New(root []byte) (*Trie, error) {
	return nil, nil
}

// Get the node's value in trie
// Branch Node: 17-elements array, value is [hash_0, hash_1, ..., hash_f, hash]
// Extension Node: 3-elements array, value is [ext flag, encodedPath, next hash]
// Leaf Node: 3-elements array, value is [leaf flag, encodedPath, value]
func (t *Trie) Get(key []byte) ([][]byte, error) {
	return nil, nil
}

// Update or add the node's value in trie
func (t *Trie) Update(key []byte, val [][]byte) error {
	return nil
}

// Del the node's value in trie
func (t *Trie) Del(key []byte) error {
	return nil
}

// Clone the trie to create a new trie sharing the same storage
func (t *Trie) Clone() (*Trie, error) {
	return nil, nil
}
