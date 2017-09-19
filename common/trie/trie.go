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
	"github.com/nebulasio/go-nebulas/crypto/hash"
	"github.com/nebulasio/go-nebulas/utils/byteutils"
)

// Flag to identify the type of node
type Flag int

const (
	unknown Flag = iota
	ext
	leaf
	branch
)

// Node in trie, three kinds,
// Branch Node [hash_0, hash_1, ..., hash_f]
// Extension Node [flag, encodedPath, next hash]
// Leaf Node [flag, encodedPath, value]
type Node struct {
	Key []byte
	Val [][]byte
}

// Flag identify Branch, Extension, Leaf Node
func (n *Node) Flag() (Flag, error) {
	if n.Val == nil {
		return unknown, errors.New("nil Node")
	}
	switch len(n.Val) {
	case 16: // Branch Node
		return branch, nil
	case 3: // Extension Node or Leaf Node
		if n.Val[0] == nil {
			return unknown, errors.New("nil flag in Node")
		}
		return Flag(n.Val[0][0]), nil
	default:
		return unknown, errors.New("wrong Node, expect [16][]byte or [3][]byte")
	}
}

// Trie is a Merkle Patricia Triee.
type Trie struct {
	rootHash   []byte
	serializer byteutils.Serializable
	storage    *Storage
}

// CreateNode in trie
func (t *Trie) CreateNode(val [][]byte) (*Node, error) {
	ir, err := t.serializer.Serialize(val)
	if err != nil {
		return nil, err
	}
	key := hash.Sha3256(ir)
	err = t.storage.Put(key, ir)
	if err != nil {
		return nil, err
	}
	return &Node{key, val}, nil
}

// FetchNode in trie
func (t *Trie) FetchNode(key []byte) (*Node, error) {
	ir, err := t.storage.Get(key)
	if err != nil {
		return nil, err
	}
	var val [][]byte
	err = t.serializer.Deserialize(ir, &val)
	if err != nil {
		return nil, err
	}
	return &Node{key, val}, nil
}

// CommitNode node in trie into storage
func (t *Trie) CommitNode(n *Node) error {
	ir, err := t.serializer.Serialize(n.Val)
	if err != nil {
		return err
	}
	n.Key = hash.Sha3256(ir)
	return t.storage.Put(n.Key, ir)
}

// NewTrie if rootHash is nil, create a new Trie, otherwise, build an existed trie
func NewTrie(rootHash []byte) (*Trie, error) {
	var serializer byteutils.Serializable = &byteutils.JSONSerializer{}
	storage, _ := NewStorage()
	if rootHash == nil {
		empty := [][]byte{nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}
		ir, err := serializer.Serialize(empty)
		if err != nil {
			return nil, err
		}
		rootHash = hash.Sha3256(ir)
		storage.Put(rootHash, ir)
	} else if _, err := storage.Get(rootHash); err != nil {
		return nil, err
	}
	return &Trie{rootHash, serializer, storage}, nil
}

// Get the node's value in trie, all node's key is hash(value)
// Branch Node: 16-elements array, value is [hash_0, hash_1, ..., hash_f, hash]
// Extension Node: 3-elements array, value is [ext flag, prefix path, next hash]
// Leaf Node: 3-elements array, value is [leaf flag, suffix path, value]
func (t *Trie) Get(key []byte) (*Node, error) {
	return nil, nil
}

// Update or add the node's value in trie
func (t *Trie) Update(key []byte, val []byte) error {
	newHash, err := t.update(t.rootHash, keybytesToHex(key), val)
	t.rootHash = newHash
	return err
}

// update or add node on root node, return new node hash
func (t *Trie) update(root []byte, route []byte, val []byte) ([]byte, error) {
	// directly add leaf node
	if root == nil {
		// create leaf node
		value := [][]byte{[]byte{byte(leaf)}, route, val}
		node, err := t.CreateNode(value)
		if err != nil {
			return nil, err
		}
		return node.Key, nil
	}
	// fetch sub-trie root node
	rootNode, err := t.FetchNode(root)
	if err != nil {
		return nil, err
	}
	flag, err := rootNode.Flag()
	if err != nil {
		return nil, err
	}
	// add new node to the sub-trie
	switch flag {
	case branch:
		return t.updateWhenMeetBranch(rootNode, route, val)
	case ext:
		return t.updateWhenMeetExt(rootNode, route, val)
	case leaf:
		return t.updateWhenMeetLeaf(rootNode, route, val)
	default:
		return nil, errors.New("unknown node type")
	}
}

// add new node to one branch of branch node's 16 branches according to route
func (t *Trie) updateWhenMeetBranch(rootNode *Node, route []byte, val []byte) ([]byte, error) {
	// update sub-trie
	newHash, err := t.update(rootNode.Val[route[0]], route[1:], val)
	if err != nil {
		return nil, err
	}
	// update the branch hash
	rootNode.Val[route[0]] = newHash
	// save updated node to storage
	t.CommitNode(rootNode)
	return rootNode.Key, nil
}

// split ext node's into an ext node and a branch node based on
// the longest common prefix between route and ext node's path
// add new node to the branch node
func (t *Trie) updateWhenMeetExt(rootNode *Node, route []byte, val []byte) ([]byte, error) {
	var err error
	path := rootNode.Val[1]
	next := rootNode.Val[2]
	if len(path) > len(route) {
		return nil, errors.New("route should be longer than ext node's path")
	}
	matchLen := prefixLen(path, route)
	// add new node to the ext node's sub-trie
	if matchLen == len(path) {
		newHash, err := t.update(next, route[matchLen:], val)
		if err != nil {
			return nil, err
		}
		// update the new hash
		rootNode.Val[2] = newHash
		// save updated node to storage
		t.CommitNode(rootNode)
		return rootNode.Key, nil
	}
	// create a new branch for the new node
	brNode := emptyBrNode()
	// a branch to hold the ext node's sub-trie
	brNode.Val[path[matchLen]] = next
	// a branch to hold the new node
	if brNode.Val[route[matchLen]], err = t.update(nil, route[matchLen+1:], val); err != nil {
		return nil, err
	}
	// save branch to the storage
	if err := t.CommitNode(brNode); err != nil {
		return nil, err
	}
	// if no common prefix, replace the ext node with the new branch node
	if matchLen == 0 {
		return brNode.Key, nil
	}
	// use the new branch node as the ext node's sub-trie
	rootNode.Val[1] = path[0:matchLen]
	rootNode.Val[2] = brNode.Key
	if err := t.CommitNode(rootNode); err != nil {
		return nil, err
	}
	return rootNode.Key, nil
}

// split leaf node's into an ext node and a branch node based on
// the longest common prefix between route and leaf node's path
// add new node to the branch node
func (t *Trie) updateWhenMeetLeaf(rootNode *Node, route []byte, val []byte) ([]byte, error) {
	var err error
	path := rootNode.Val[1]
	leafVal := rootNode.Val[2]
	if len(path) > len(route) {
		return nil, errors.New("route should be longer than ext node's path")
	}
	matchLen := prefixLen(path, route)
	// node exists, update its value
	if matchLen == len(path) {
		rootNode.Val[2] = val
		// save updated node to storage
		t.CommitNode(rootNode)
		return rootNode.Key, nil
	}
	// create a new branch for the new node
	brNode := emptyBrNode()
	// a branch to hold the leaf node
	if brNode.Val[path[matchLen]], err = t.update(nil, path[matchLen+1:], leafVal); err != nil {
		return nil, err
	}
	// a branch to hold the new node
	if brNode.Val[route[matchLen]], err = t.update(nil, route[matchLen+1:], val); err != nil {
		return nil, err
	}
	// save the new branch node to storage
	if err := t.CommitNode(brNode); err != nil {
		return nil, err
	}
	// if no common prefix, replace the leaf node with the new branch node
	if matchLen == 0 {
		return brNode.Key, nil
	}
	// create a new ext node, and use the new branch node as the new ext node's sub-trie
	rootNode.Val[0] = []byte{byte(ext)}
	rootNode.Val[1] = path[0:matchLen]
	rootNode.Val[2] = brNode.Key
	if err := t.CommitNode(rootNode); err != nil {
		return nil, err
	}
	return rootNode.Key, nil
}

// Del the node's value in trie
func (t *Trie) Del(key []byte) error {
	return nil
}

// Clone the trie to create a new trie sharing the same storage
func (t *Trie) Clone() (*Trie, error) {
	return &Trie{t.rootHash, t.serializer, t.storage}, nil
}

// prefixLen returns the length of the common prefix of a and b.
func prefixLen(a, b []byte) int {
	var i, length = 0, len(a)
	if len(b) < length {
		length = len(b)
	}
	for ; i < length; i++ {
		if a[i] != b[i] {
			break
		}
	}
	return i
}

func keybytesToHex(str []byte) []byte {
	l := len(str) * 2
	var nibbles = make([]byte, l)
	for i, b := range str {
		nibbles[i*2] = b / 16
		nibbles[i*2+1] = b % 16
	}
	return nibbles
}

func emptyBrNode() *Node {
	empty := [][]byte{nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}
	return &Node{[]byte{}, empty}
}
