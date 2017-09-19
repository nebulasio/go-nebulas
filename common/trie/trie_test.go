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
	"fmt"
	"reflect"
	"testing"

	"github.com/nebulasio/go-nebulas/crypto/hash"
	"github.com/nebulasio/go-nebulas/utils/byteutils"
)

func TestNewTrie(t *testing.T) {
	type args struct {
		rootHash []byte
	}
	rootHash, _ := byteutils.FromHex("2bf57a4ca277d533a4af0f591bfb589e022c2e36546f5c9f1506f0817d391191")
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			"nil",
			args{nil},
			rootHash,
			false,
		},
		{
			"root",
			args{[]byte("hello")},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTrie(tt.args.rootHash)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				if tt.want == nil {
					return
				} else {

					t.Errorf("New() = %v, want %v", got, tt.want)
					return
				}
			}
			if !reflect.DeepEqual(got.rootHash, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrie_Clone(t *testing.T) {
	type fields struct {
		rootHash   []byte
		serializer byteutils.Serializable
		storage    *Storage
	}
	storage, _ := NewStorage()
	serializer := &byteutils.JSONSerializer{}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			"case 1",
			fields{[]byte("hello"), serializer, storage},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Trie{
				rootHash:   tt.fields.rootHash,
				serializer: tt.fields.serializer,
				storage:    tt.fields.storage,
			}
			got, err := tr.Clone()
			if (err != nil) != tt.wantErr {
				t.Errorf("Trie.Clone() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.serializer, tt.fields.serializer) {
				t.Errorf("Trie.Clone() = %v, want %v", got.serializer, tt.fields.serializer)
			}
			if &got.rootHash == &tt.fields.rootHash {
				t.Errorf("Trie.Clone() = %p, not want %p", &got.rootHash, &tt.fields.rootHash)
			}
			if !reflect.DeepEqual(got.storage, tt.fields.storage) {
				t.Errorf("Trie.Clone() = %v, want %v", got.storage, tt.fields.storage)
			}
		})
	}
}

func TestTrie_Update(t *testing.T) {
	tr, _ := NewTrie(nil)
	// add a new leaf node
	addr1, _ := byteutils.FromHex("1f345678e9")
	key1 := []byte{0x1, 0xf, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0xe, 0x9}
	tr.Update(addr1, []byte("leaf 1"))
	leaf1 := [][]byte{[]byte{byte(leaf)}, key1[1:], []byte("leaf 1")}
	leaf1IR, _ := tr.serializer.Serialize(leaf1)
	leaf1H := hash.Sha3256(leaf1IR)
	branch1 := [][]byte{nil, leaf1H, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}
	branch1IR, _ := tr.serializer.Serialize(branch1)
	branch1H := hash.Sha3256(branch1IR)
	if !reflect.DeepEqual(branch1H, tr.rootHash) {
		t.Errorf("1 Trie.Update() = %v, want %v", branch1H, tr.rootHash)
	}
	// add a new leaf node with 3-length common prefix
	addr2, _ := byteutils.FromHex("1f355678e9")
	key2 := []byte{0x1, 0xf, 0x3, 0x5, 0x5, 0x6, 0x7, 0x8, 0xe, 0x9}
	tr.Update(addr2, []byte("leaf 2"))
	leaf2 := [][]byte{[]byte{byte(leaf)}, key2[4:], []byte("leaf 2")}
	leaf2IR, _ := tr.serializer.Serialize(leaf2)
	leaf2H := hash.Sha3256(leaf2IR)
	fmt.Println("leaf2")
	fmt.Println(leaf2H)
	leaf3 := [][]byte{[]byte{byte(leaf)}, key1[4:], []byte("leaf 1")}
	leaf3IR, _ := tr.serializer.Serialize(leaf3)
	leaf3H := hash.Sha3256(leaf3IR)
	fmt.Println("leaf3")
	fmt.Println(leaf3H)
	branch2 := [][]byte{nil, nil, nil, nil, leaf3H, leaf2H, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}
	branch2IR, _ := tr.serializer.Serialize(branch2)
	branch2H := hash.Sha3256(branch2IR)
	fmt.Println("branch2")
	fmt.Println(branch2H)
	ext1 := [][]byte{[]byte{(byte(ext))}, key2[1:3], branch2H}
	ext1IR, _ := tr.serializer.Serialize(ext1)
	ext1H := hash.Sha3256(ext1IR)
	fmt.Println("ext1")
	fmt.Println(ext1H)
	branch3 := [][]byte{nil, ext1H, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}
	branch3IR, _ := tr.serializer.Serialize(branch3)
	branch3H := hash.Sha3256(branch3IR)
	fmt.Println("branch3")
	fmt.Println(branch3H)
	if !reflect.DeepEqual(branch3H, tr.rootHash) {
		t.Errorf("2 Trie.Update() = %v, want %v", branch3H, tr.rootHash)
	}
	// add a new node with 2-length common prefix
	addr3, _ := byteutils.FromHex("1f555678e9")
	key3 := []byte{0x1, 0xf, 0x5, 0x5, 0x5, 0x6, 0x7, 0x8, 0xe, 0x9}
	tr.Update(addr3, []byte("leaf 3"))
	leaf4 := [][]byte{[]byte{byte(leaf)}, key3[3:], []byte("leaf 3")}
	leaf4IR, _ := tr.serializer.Serialize(leaf4)
	leaf4H := hash.Sha3256(leaf4IR)
	fmt.Println("leaf4")
	fmt.Println(leaf4H)
	branch4 := [][]byte{nil, nil, nil, branch2H, nil, leaf4H, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}
	branch4IR, _ := tr.serializer.Serialize(branch4)
	branch4H := hash.Sha3256(branch4IR)
	fmt.Println("branch4")
	fmt.Println(branch4H)
	ext2 := [][]byte{[]byte{(byte(ext))}, key3[1:2], branch4H}
	ext2IR, _ := tr.serializer.Serialize(ext2)
	ext2H := hash.Sha3256(ext2IR)
	fmt.Println("ext2")
	fmt.Println(ext2H)
	branch5 := [][]byte{nil, ext2H, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}
	branch5IR, _ := tr.serializer.Serialize(branch5)
	branch5H := hash.Sha3256(branch5IR)
	fmt.Println("branch5")
	fmt.Println(branch5H)
	if !reflect.DeepEqual(branch5H, tr.rootHash) {
		t.Errorf("3 Trie.Update() = %v, want %v", branch5H, tr.rootHash)
	}
	// update node leaf1
	tr.Update(addr1, []byte("leaf 11"))
	leaf5 := [][]byte{[]byte{byte(leaf)}, key1[4:], []byte("leaf 11")}
	leaf5IR, _ := tr.serializer.Serialize(leaf5)
	leaf5H := hash.Sha3256(leaf5IR)
	fmt.Println("leaf5")
	fmt.Println(leaf5H)
	branch6 := [][]byte{nil, nil, nil, nil, leaf5H, leaf2H, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}
	branch6IR, _ := tr.serializer.Serialize(branch6)
	branch6H := hash.Sha3256(branch6IR)
	fmt.Println("branch6")
	fmt.Println(branch6H)
	branch7 := [][]byte{nil, nil, nil, branch6H, nil, leaf4H, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}
	branch7IR, _ := tr.serializer.Serialize(branch7)
	branch7H := hash.Sha3256(branch7IR)
	fmt.Println("branch7")
	fmt.Println(branch7H)
	ext3 := [][]byte{[]byte{(byte(ext))}, key3[1:2], branch7H}
	ext3IR, _ := tr.serializer.Serialize(ext3)
	ext3H := hash.Sha3256(ext3IR)
	fmt.Println("ext3")
	fmt.Println(ext3H)
	branch8 := [][]byte{nil, ext3H, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}
	branch8IR, _ := tr.serializer.Serialize(branch8)
	branch8H := hash.Sha3256(branch8IR)
	fmt.Println("branch8")
	fmt.Println(branch8H)
	if !reflect.DeepEqual(branch8H, tr.rootHash) {
		t.Errorf("4 Trie.Update() = %v, want %v", branch8H, tr.rootHash)
	}
}
