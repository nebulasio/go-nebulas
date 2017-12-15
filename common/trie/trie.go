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

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/common/trie/pb"
	"github.com/nebulasio/go-nebulas/crypto/hash"
	"github.com/nebulasio/go-nebulas/storage"
)

// Flag to identify the type of node
type ty int

const (
	unknown ty = iota
	ext
	leaf
	branch
)

// Errors
var (
	ErrNotFound = storage.ErrKeyNotFound
)

// Node in trie, three kinds,
// Branch Node [hash_0, hash_1, ..., hash_f]
// Extension Node [flag, encodedPath, next hash]
// Leaf Node [flag, encodedPath, value]
type node struct {
	Hash  []byte
	Bytes []byte
	Val   [][]byte
}

func (n *node) ToProto() (proto.Message, error) {
	return &triepb.Node{
		Val: n.Val,
	}, nil
}

func (n *node) FromProto(msg proto.Message) error {
	if msg, ok := msg.(*triepb.Node); ok {
		n.Bytes, _ = proto.Marshal(msg)
		n.Hash = hash.Sha3256(n.Bytes)
		n.Val = msg.Val
		return nil
	}
	return errors.New("Pb Message cannot be converted into Node")
}

// Type of node, e.g. Branch, Extension, Leaf Node
func (n *node) Type() (ty, error) {
	if n.Val == nil {
		return unknown, errors.New("nil node")
	}
	switch len(n.Val) {
	case 16: // Branch Node
		return branch, nil
	case 3: // Extension Node or Leaf Node
		if n.Val[0] == nil {
			return unknown, errors.New("unknown node type")
		}
		return ty(n.Val[0][0]), nil
	default:
		return unknown, errors.New("wrong node value, expect [16][]byte or [3][]byte, get [" + string(len(n.Val)) + "][]byte")
	}
}

// Trie is a Merkle Patricia Triee, consists of three kinds of nodes,
// Branch Node: 16-elements array, value is [hash_0, hash_1, ..., hash_f, hash]
// Extension Node: 3-elements array, value is [ext flag, prefix path, next hash]
// Leaf Node: 3-elements array, value is [leaf flag, suffix path, value]
type Trie struct {
	rootHash []byte
	storage  storage.Storage
}

// CreateNode in trie
func (t *Trie) createNode(val [][]byte) (*node, error) {
	n := &node{Val: val}
	if err := t.commitNode(n); err != nil {
		return nil, err
	}
	return n, nil
}

// FetchNode in trie
func (t *Trie) fetchNode(hash []byte) (*node, error) {
	ir, err := t.storage.Get(hash)
	if err != nil {
		return nil, err
	}
	pb := new(triepb.Node)
	if err := proto.Unmarshal(ir, pb); err != nil {
		return nil, err
	}
	n := new(node)
	if err := n.FromProto(pb); err != nil {
		return nil, err
	}
	return n, nil
}

// CommitNode node in trie into storage
func (t *Trie) commitNode(n *node) error {
	pb, err := n.ToProto()
	if err != nil {
		return err
	}
	n.Bytes, err = proto.Marshal(pb)
	if err != nil {
		return err
	}
	n.Hash = hash.Sha3256(n.Bytes)
	return t.storage.Put(n.Hash, n.Bytes)
}

// NewTrie if rootHash is nil, create a new Trie, otherwise, build an existed trie
func NewTrie(rootHash []byte, storage storage.Storage) (*Trie, error) {
	t := &Trie{rootHash, storage}
	if t.rootHash == nil {
		return t, nil
	} else if _, err := t.storage.Get(rootHash); err != nil {
		return nil, err
	}
	return t, nil
}

// RootHash return trie's rootHash
func (t *Trie) RootHash() []byte {
	return t.rootHash
}

// Empty return if the trie is empty
func (t *Trie) Empty() bool {
	return t.rootHash == nil
}

// Get the value to the key in trie
func (t *Trie) Get(key []byte) ([]byte, error) {
	return t.get(t.rootHash, keyToRoute(key))
}

func (t *Trie) get(rootHash []byte, route []byte) ([]byte, error) {
	curRootHash := rootHash
	curRoute := route
	for i := 0; i < len(curRoute); i++ {
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
			break
		case ext:
			path := rootNode.Val[1]
			next := rootNode.Val[2]
			matchLen := prefixLen(path, curRoute)
			if matchLen != len(path) {
				return nil, ErrNotFound
			}
			curRootHash = next
			curRoute = curRoute[matchLen:]
			break
		case leaf:
			path := rootNode.Val[1]
			matchLen := prefixLen(path, curRoute)
			if matchLen != len(path) || matchLen != len(curRoute) {
				return nil, ErrNotFound
			}
			return rootNode.Val[2], nil
		default:
			return nil, errors.New("unknown node type")
		}
	}
	return nil, ErrNotFound
}

// Put the key-value pair in trie
func (t *Trie) Put(key []byte, val []byte) ([]byte, error) {
	newHash, err := t.update(t.rootHash, keyToRoute(key), val)
	if err != nil {
		return nil, err
	}
	t.rootHash = newHash
	return newHash, nil
}

func (t *Trie) update(root []byte, route []byte, val []byte) ([]byte, error) {
	if root == nil || len(root) == 0 {
		// directly add leaf node
		value := [][]byte{[]byte{byte(leaf)}, route, val}
		node, err := t.createNode(value)
		if err != nil {
			return nil, err
		}
		return node.Hash, nil
	}
	rootNode, err := t.fetchNode(root)
	if err != nil {
		return nil, err
	}
	flag, err := rootNode.Type()
	if err != nil {
		return nil, err
	}
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
func (t *Trie) updateWhenMeetBranch(rootNode *node, route []byte, val []byte) ([]byte, error) {
	// update sub-trie
	newHash, err := t.update(rootNode.Val[route[0]], route[1:], val)
	if err != nil {
		return nil, err
	}
	// update the branch hash
	rootNode.Val[route[0]] = newHash
	// save updated node to storage
	t.commitNode(rootNode)
	return rootNode.Hash, nil
}

// split ext node's into an ext node and a branch node based on
// the longest common prefix between route and ext node's path
// add ext node's child and new node to the branch node
func (t *Trie) updateWhenMeetExt(rootNode *node, route []byte, val []byte) ([]byte, error) {
	var err error
	path := rootNode.Val[1]
	next := rootNode.Val[2]
	if len(path) > len(route) {
		return nil, errors.New("wrong key, too short")
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
		t.commitNode(rootNode)
		return rootNode.Hash, nil
	}
	// create a new branch for the new node
	brNode := emptyBranchNode()
	// 4 situations
	// 1. matchLen > 0 && matchLen == len(path), 24 meets 24... => ext - branch - ...
	// 2. matchLen > 0 && matchLen + 1 < len(path), 23... meets 24... => ext - branch - ext ...
	// 3. matchLen = 0 && len(path) = 1, 1 meets 2 => branch - ...
	if matchLen > 0 || len(path) == 1 {
		// a branch to hold the ext node's sub-trie
		brNode.Val[path[matchLen]] = next
		if matchLen > 0 && matchLen+1 < len(path) {
			value := [][]byte{[]byte{byte(ext)}, path[matchLen+1:], next}
			extNode, err := t.createNode(value)
			if err != nil {
				return nil, err
			}
			brNode.Val[path[matchLen]] = extNode.Hash
		}
		// a branch to hold the new node
		if brNode.Val[route[matchLen]], err = t.update(nil, route[matchLen+1:], val); err != nil {
			return nil, err
		}
		// save branch to the storage
		if err := t.commitNode(brNode); err != nil {
			return nil, err
		}
		// if no common prefix, replace the ext node with the new branch node
		if matchLen == 0 {
			return brNode.Hash, nil
		}
		// use the new branch node as the ext node's sub-trie
		rootNode.Val[1] = path[0:matchLen]
		rootNode.Val[2] = brNode.Hash
		if err := t.commitNode(rootNode); err != nil {
			return nil, err
		}
		return rootNode.Hash, nil
	}
	// 4. matchLen = 0 && len(path) > 1, 12... meets 23... => branch - ext - ...
	rootNode.Val[1] = path[1:]
	// save ext node to storage
	if err := t.commitNode(rootNode); err != nil {
		return nil, err
	}
	brNode.Val[path[matchLen]] = rootNode.Hash
	// a branch to hold the new node
	if brNode.Val[route[matchLen]], err = t.update(nil, route[matchLen+1:], val); err != nil {
		return nil, err
	}
	// save branch to the storage
	if err := t.commitNode(brNode); err != nil {
		return nil, err
	}
	return brNode.Hash, nil
}

// split leaf node's into an ext node and a branch node based on
// the longest common prefix between route and leaf node's path
// add new node to the branch node
func (t *Trie) updateWhenMeetLeaf(rootNode *node, route []byte, val []byte) ([]byte, error) {
	var err error
	path := rootNode.Val[1]
	leafVal := rootNode.Val[2]
	if len(path) > len(route) {
		return nil, errors.New("wrong key, too short")
	}
	matchLen := prefixLen(path, route)
	// node exists, update its value
	if matchLen == len(path) {
		rootNode.Val[2] = val
		// save updated node to storage
		t.commitNode(rootNode)
		return rootNode.Hash, nil
	}
	// create a new branch for the new node
	brNode := emptyBranchNode()
	// a branch to hold the leaf node
	if brNode.Val[path[matchLen]], err = t.update(nil, path[matchLen+1:], leafVal); err != nil {
		return nil, err
	}
	// a branch to hold the new node
	if brNode.Val[route[matchLen]], err = t.update(nil, route[matchLen+1:], val); err != nil {
		return nil, err
	}
	// save the new branch node to storage
	if err := t.commitNode(brNode); err != nil {
		return nil, err
	}
	// if no common prefix, replace the leaf node with the new branch node
	if matchLen == 0 {
		return brNode.Hash, nil
	}
	// create a new ext node, and use the new branch node as the new ext node's sub-trie
	rootNode.Val[0] = []byte{byte(ext)}
	rootNode.Val[1] = path[0:matchLen]
	rootNode.Val[2] = brNode.Hash
	if err := t.commitNode(rootNode); err != nil {
		return nil, err
	}
	return rootNode.Hash, nil
}

// Del the node's value in trie
func (t *Trie) Del(key []byte) ([]byte, error) {
	newHash, err := t.del(t.rootHash, keyToRoute(key))
	if err != nil {
		return nil, err
	}
	t.rootHash = newHash
	return newHash, nil
}

func (t *Trie) del(root []byte, route []byte) ([]byte, error) {
	if root == nil || len(root) == 0 {
		return nil, ErrNotFound
	}
	// fetch sub-trie root node
	rootNode, err := t.fetchNode(root)
	if err != nil {
		return nil, err
	}
	flag, err := rootNode.Type()
	if err != nil {
		return nil, err
	}
	switch flag {
	case branch:
		newHash, err := t.del(rootNode.Val[route[0]], route[1:])
		if err != nil {
			return nil, err
		}
		rootNode.Val[route[0]] = newHash
		// remove empty branch node
		if isEmptyBranch(rootNode) {
			return nil, nil
		}
		if err := t.commitNode(rootNode); err != nil {
			return nil, err
		}
		return rootNode.Hash, nil
	case ext:
		path := rootNode.Val[1]
		next := rootNode.Val[2]
		matchLen := prefixLen(path, route)
		if matchLen != len(path) {
			return nil, ErrNotFound
		}
		newHash, err := t.del(next, route[matchLen:])
		if err != nil {
			return nil, err
		}
		// remove empty ext node
		if newHash == nil {
			return nil, nil
		}
		rootNode.Val[2] = newHash
		if err := t.commitNode(rootNode); err != nil {
			return nil, err
		}
		return rootNode.Hash, nil
	case leaf:
		path := rootNode.Val[1]
		matchLen := prefixLen(path, route)
		if matchLen != len(path) {
			return nil, ErrNotFound
		}
		return nil, nil
	default:
		return nil, errors.New("unknown node type")
	}
}

// Clone the trie to create a new trie sharing the same storage
func (t *Trie) Clone() (*Trie, error) {
	return &Trie{t.rootHash, t.storage}, nil
}

// prefixLen returns the length of the common prefix between a and b.
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

// keyToRoute returns hex bytes
// e.g {0xa1, 0xf2} -> {0xa, 0x1, 0xf, 0x2}
func keyToRoute(key []byte) []byte {
	l := len(key) * 2
	var route = make([]byte, l)
	for i, b := range key {
		route[i*2] = b / 16
		route[i*2+1] = b % 16
	}
	return route
}

func emptyBranchNode() *node {
	empty := &node{Val: [][]byte{nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}}
	pb, _ := empty.ToProto()
	empty.Bytes, _ = proto.Marshal(pb)
	empty.Hash = hash.Sha3256(empty.Bytes)
	return empty
}

func isEmptyBranch(n *node) bool {
	for idx := range n.Val {
		if len(n.Val[idx]) != 0 {
			return false
		}
	}
	return true
}
