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

// MerkleProof is a path from root to the proved node
type MerkleProof [][]byte

// Prove the associated node to the key exists in trie
// if exists, MerkleProof is a complete path from root to the node
// otherwise, MerkleProof is a longest existing prefix of the node
func (t *Trie) Prove(key []byte) MerkleProof {
	return nil
}

// Verify whether the merkle proof from root to the associated node is right
func Verify(root []byte, key []byte, proof MerkleProof) error {
	return nil
}
