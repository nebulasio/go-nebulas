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

import (
	"errors"
)

// errors constants
var (
	ErrNotIterable = errors.New("leaf node is not iterable")
)

// IteratorState represents the intermediate statue in iterator
type IteratorState struct {
	node *node
	pos  int
}

// Iterator to traverse leaf node in a trie
type Iterator struct {
	stack []*IteratorState
	value []byte
	root  *Trie
}

func validElementsInBranchNode(offset int, node *node) []int {
	var valid []int
	ty, _ := node.Type()
	if ty != branch {
		return valid
	}
	for i := offset; i < 16; i++ {
		if node.Val[i] != nil && len(node.Val[i]) > 0 {
			valid = append(valid, i)
		}
	}
	return valid
}

// Iterator return an iterator
func (t *Trie) Iterator(prefix []byte) (*Iterator, error) {
	rootHash, err := t.getSubTrieWithMaxCommonPrefix(prefix)
	if err != nil {
		return nil, err
	}
	node, err := t.fetchNode(rootHash)
	if err != nil {
		return nil, err
	}
	pos := -1
	valid := validElementsInBranchNode(0, node)
	if len(valid) > 0 {
		pos = valid[0]
	}
	return &Iterator{
		root:  t,
		stack: []*IteratorState{&IteratorState{node, pos}},
		value: nil,
	}, nil
}

func (t *Trie) getSubTrieWithMaxCommonPrefix(prefix []byte) ([]byte, error) {
	curRootHash := t.rootHash
	curRoute := keyToRoute(prefix)
	for len(curRoute) > 0 {
		rootNode, err := t.fetchNode(curRootHash)
		if err != nil {
			return nil, err
		}
		flag, err := rootNode.Type()
		if err != nil {
			return nil, err
		}
		switch flag {
		case branch:
			curRootHash = rootNode.Val[curRoute[0]]
			curRoute = curRoute[1:]
		case ext:
			path := rootNode.Val[1]
			next := rootNode.Val[2]
			matchLen := prefixLen(path, curRoute)
			if matchLen != len(path) && matchLen != len(curRoute) {
				return nil, ErrNotFound
			}
			curRootHash = next
			curRoute = curRoute[matchLen:]
		case leaf:
			path := rootNode.Val[1]
			matchLen := prefixLen(path, curRoute)
			if matchLen != len(path) && matchLen != len(curRoute) {
				return nil, ErrNotFound
			}
			curRootHash = rootNode.Hash
			curRoute = curRoute[matchLen:]
		default:
			return nil, errors.New("unknown node type")
		}
	}
	return curRootHash, nil
}

func (it *Iterator) push(node *node, pos int) {
	it.stack = append(it.stack, &IteratorState{node, pos})
}

func (it *Iterator) pop() (*IteratorState, error) {
	size := len(it.stack)
	if size == 0 {
		return nil, errors.New("empty stack")
	}
	state := it.stack[size-1]
	it.stack = it.stack[0 : size-1]
	return state, nil
}

// Next return if there is next leaf node
func (it *Iterator) Next() (bool, error) {
	state, err := it.pop()
	if err != nil {
		return false, nil
	}
	node := state.node
	pos := state.pos
	ty, err := node.Type()
	for {
		switch ty {
		case branch:
			valid := validElementsInBranchNode(pos, node)
			if len(valid) == 0 {
				return false, errors.New("empty branch node")
			}
			if len(valid) > 1 {
				it.push(node, valid[1])
			}
			node, err = it.root.fetchNode(node.Val[valid[0]])
			if err != nil {
				return false, err
			}
			ty, err = node.Type()
		case ext:
			node, err = it.root.fetchNode(node.Val[2])
			if err != nil {
				return false, err
			}
			ty, err = node.Type()
		case leaf:
			it.value = node.Val[2]
			return true, nil
		default:
			return false, err
		}
		pos = 0
	}
}

// Value return current leaf node's value
func (it *Iterator) Value() []byte {
	return it.value
}
