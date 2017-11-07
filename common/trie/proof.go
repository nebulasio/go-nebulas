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
	"bytes"
	"errors"
)

// MerkleProof is a path from root to the proved node
// every element in path is the value of a node
type MerkleProof [][][]byte

// Prove the associated node to the key exists in trie
// if exists, MerkleProof is a complete path from root to the node
// otherwise, MerkleProof is nil
func (t *Trie) Prove(key []byte) (MerkleProof, error) {
	curRoute := keyToRoute(key)
	curRootHash := t.rootHash
	var proof MerkleProof
	for len(curRoute) > 0 {
		// fetch sub-trie root node
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
			proof = append(proof, rootNode.Val)
			curRootHash = rootNode.Val[curRoute[0]]
			curRoute = curRoute[1:]
		case ext:
			path := rootNode.Val[1]
			next := rootNode.Val[2]
			matchLen := prefixLen(path, curRoute)
			if matchLen != len(path) {
				return nil, ErrNotFound
			}
			proof = append(proof, rootNode.Val)
			curRootHash = next
			curRoute = curRoute[matchLen:]
		case leaf:
			path := rootNode.Val[1]
			matchLen := prefixLen(path, curRoute)
			if matchLen != len(path) {
				return nil, ErrNotFound
			}
			proof = append(proof, rootNode.Val)
			return proof, nil
		default:
			return nil, ErrNotFound
		}
	}
	return nil, ErrNotFound
}

// Verify whether the merkle proof from root to the associated node is right
func (t *Trie) Verify(rootHash []byte, key []byte, proof MerkleProof) error {
	curRoute := keyToRoute(key)
	length := len(proof)
	wantHash := rootHash
	for i := 0; i < length; i++ {
		val := proof[i]
		n, err := t.createNode(val)
		if err != nil {
			return err
		}
		proofHash := n.Hash
		if !bytes.Equal(wantHash, proofHash) {
			return errors.New("wrong hash")
		}
		switch len(val) {
		case 16: // Branch Node
			wantHash = val[curRoute[0]]
			curRoute = curRoute[1:]
			break
		case 3: // Extension Node or Leaf Node
			if val[0] == nil || len(val) == 0 {
				return errors.New("unknown node type")
			}
			if val[0][0] == byte(ext) {
				extLen := len(val[1])
				if !bytes.Equal(val[1], curRoute[:extLen]) {
					return errors.New("wrong hash")
				}
				wantHash = val[2]
				curRoute = curRoute[extLen:]
				break
			} else if val[0][0] == byte(leaf) {
				if !bytes.Equal(val[1], curRoute) {
					return errors.New("wrong hash")
				}
				return nil
			}
			return errors.New("unknown node type")
		default:
			return errors.New("wrong node value, expect [16][]byte or [3][]byte, get [" + string(len(proofHash)) + "][]byte")
		}
	}
	return nil
}
