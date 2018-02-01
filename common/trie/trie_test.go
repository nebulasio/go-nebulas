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
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/common/trie/pb"
	"github.com/nebulasio/go-nebulas/crypto/hash"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

func TestNewTrie(t *testing.T) {
	type args struct {
		rootHash []byte
	}
	storage, _ := storage.NewMemoryStorage()
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			"nil",
			args{nil},
			nil,
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
			got, err := NewTrie(tt.args.rootHash, storage)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				if tt.want == nil {
					return
				}
				t.Errorf("New() = %v, want %v", got, tt.want)
				return
			}
			if !reflect.DeepEqual(got.rootHash, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrie_Clone(t *testing.T) {
	type fields struct {
		rootHash []byte
		storage  storage.Storage
	}
	storage, _ := storage.NewMemoryStorage()
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			"case 1",
			fields{[]byte("hello"), storage},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Trie{
				rootHash: tt.fields.rootHash,
				storage:  tt.fields.storage,
			}
			got, err := tr.Clone()
			if (err != nil) != tt.wantErr {
				t.Errorf("Trie.Clone() error = %v, wantErr %v", err, tt.wantErr)
				return
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

func TestTrie_Operation(t *testing.T) {
	storage, _ := storage.NewMemoryStorage()
	tr, _ := NewTrie(nil, storage)
	if !reflect.DeepEqual([]byte(nil), tr.rootHash) {
		t.Errorf("3 Trie.Del() = %v, want %v", nil, tr.rootHash)
	}
	// add a new leaf node
	addr1, _ := byteutils.FromHex("1f345678e9")
	key1 := []byte{0x1, 0xf, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0xe, 0x9}
	val1 := []byte("leaf 11")
	tr.Put(addr1, val1)
	leaf1 := [][]byte{[]byte{byte(leaf)}, key1, val1}
	leaf1IR, _ := proto.Marshal(&triepb.Node{Val: leaf1})
	leaf1H := hash.Sha3256(leaf1IR)
	if !reflect.DeepEqual(leaf1H, tr.rootHash) {
		t.Errorf("1 Trie.Update() = %v, want %v", leaf1H, tr.rootHash)
	}
	// add a new leaf node with 3-length common prefix
	addr2, _ := byteutils.FromHex("1f355678e9")
	key2 := []byte{0x1, 0xf, 0x3, 0x5, 0x5, 0x6, 0x7, 0x8, 0xe, 0x9}
	val2 := []byte("leaf 2")
	tr.Put(addr2, val2)
	leaf2 := [][]byte{[]byte{byte(leaf)}, key2[4:], val2}
	leaf2IR, _ := proto.Marshal(&triepb.Node{Val: leaf2})
	leaf2H := hash.Sha3256(leaf2IR)
	leaf3 := [][]byte{[]byte{byte(leaf)}, key1[4:], val1}
	leaf3IR, _ := proto.Marshal(&triepb.Node{Val: leaf3})
	leaf3H := hash.Sha3256(leaf3IR)
	branch2 := [][]byte{nil, nil, nil, nil, leaf3H, leaf2H, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}
	branch2IR, _ := proto.Marshal(&triepb.Node{Val: branch2})
	branch2H := hash.Sha3256(branch2IR)
	ext1 := [][]byte{[]byte{(byte(ext))}, key2[:3], branch2H}
	ext1IR, _ := proto.Marshal(&triepb.Node{Val: ext1})
	ext1H := hash.Sha3256(ext1IR)
	if !reflect.DeepEqual(ext1H, tr.RootHash()) {
		t.Errorf("2 Trie.Update() = %v, want %v", ext1H, tr.rootHash)
	}
	// add a new node with 2-length common prefix
	addr3, _ := byteutils.FromHex("1f555678e9")
	key3 := []byte{0x1, 0xf, 0x5, 0x5, 0x5, 0x6, 0x7, 0x8, 0xe, 0x9}
	val3 := []byte{}
	tr.Put(addr3, val3)
	leaf4 := [][]byte{[]byte{byte(leaf)}, key3[3:], val3}
	leaf4IR, _ := proto.Marshal(&triepb.Node{Val: leaf4})
	leaf4H := hash.Sha3256(leaf4IR)
	branch4 := [][]byte{nil, nil, nil, branch2H, nil, leaf4H, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}
	branch4IR, _ := proto.Marshal(&triepb.Node{Val: branch4})
	branch4H := hash.Sha3256(branch4IR)
	ext2 := [][]byte{[]byte{(byte(ext))}, key3[:2], branch4H}
	ext2IR, _ := proto.Marshal(&triepb.Node{Val: ext2})
	ext2H := hash.Sha3256(ext2IR)
	if !reflect.DeepEqual(ext2H, tr.rootHash) {
		t.Errorf("3 Trie.Update() = %v, want %v", ext2H, tr.rootHash)
	}
	// update node leaf1
	hash4, _ := tr.Put(addr1, []byte("leaf 11"))
	val11 := []byte("leaf 11")
	leaf5 := [][]byte{[]byte{byte(leaf)}, key1[4:], val11}
	leaf5IR, _ := proto.Marshal(&triepb.Node{Val: leaf5})
	leaf5H := hash.Sha3256(leaf5IR)
	branch6 := [][]byte{nil, nil, nil, nil, leaf5H, leaf2H, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}
	branch6IR, _ := proto.Marshal(&triepb.Node{Val: branch6})
	branch6H := hash.Sha3256(branch6IR)
	branch7 := [][]byte{nil, nil, nil, branch6H, nil, leaf4H, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}
	branch7IR, _ := proto.Marshal(&triepb.Node{Val: branch7})
	branch7H := hash.Sha3256(branch7IR)
	ext3 := [][]byte{[]byte{(byte(ext))}, key3[:2], branch7H}
	ext3IR, _ := proto.Marshal(&triepb.Node{Val: ext3})
	ext3H := hash.Sha3256(ext3IR)
	if !reflect.DeepEqual(ext3H, hash4) {
		t.Errorf("4 Trie.Update() = %v, want %v", ext3H, tr.rootHash)
	}
	// get merkle proof of "1f345678e9"
	proof, err := tr.Prove(addr1)
	if err != nil {
		t.Errorf("1 Trie.Prove() %v", err.Error())
	}
	if err := tr.Verify(tr.rootHash, addr1, proof); err != nil {
		t.Errorf("1 Trie.Verify() %v", err.Error())
	}
	// get node "1f345678e9"
	checkVal1, _ := tr.Get(addr1)
	if !reflect.DeepEqual(checkVal1, val11) {
		t.Errorf("1 Trie.Get() val = %v, want %v", checkVal1, val11)
	}
	// get node "1f355678e9"
	checkVal2, _ := tr.Get(addr2)
	if !reflect.DeepEqual(checkVal2, val2) {
		t.Errorf("2 Trie.Get() val = %v, want %v", checkVal2, val2)
	}
	// get node "1f555678e9"
	checkVal3, _ := tr.Get(addr3)
	if !reflect.DeepEqual(checkVal3, val3) {
		t.Errorf("2 Trie.Get() val = %v, want %v", checkVal3, val3)
	}
	// del node "1f345678e9"
	hash5, _ := tr.Del(addr1)
	branch9 := [][]byte{nil, nil, nil, nil, nil, leaf2H, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}
	branch9IR, _ := proto.Marshal(&triepb.Node{Val: branch9})
	branch9H := hash.Sha3256(branch9IR)
	branch10 := [][]byte{nil, nil, nil, branch9H, nil, leaf4H, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}
	branch10IR, _ := proto.Marshal(&triepb.Node{Val: branch10})
	branch10H := hash.Sha3256(branch10IR)
	ext4 := [][]byte{[]byte{(byte(ext))}, key3[:2], branch10H}
	ext4IR, _ := proto.Marshal(&triepb.Node{Val: ext4})
	ext4H := hash.Sha3256(ext4IR)
	if !reflect.DeepEqual(ext4H, hash5) {
		t.Errorf("1 Trie.Del() = %v, want %v", ext4H, tr.rootHash)
	}
	// del node "1f355678e9"
	hash6, _ := tr.Del(addr2)
	branch12 := [][]byte{nil, nil, nil, nil, nil, leaf4H, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}
	branch12IR, _ := proto.Marshal(&triepb.Node{Val: branch12})
	branch12H := hash.Sha3256(branch12IR)
	ext5 := [][]byte{[]byte{(byte(ext))}, key3[0:2], branch12H}
	ext5IR, _ := proto.Marshal(&triepb.Node{Val: ext5})
	ext5H := hash.Sha3256(ext5IR)
	if !reflect.DeepEqual(ext5H, hash6) {
		t.Errorf("2 Trie.Del() = %v, want %v", ext5H, tr.rootHash)
	}
	// del node "1f555678e9"
	tr.Del(addr3)
	if !reflect.DeepEqual([]byte(nil), tr.rootHash) {
		t.Errorf("3 Trie.Del() = %v, want %v", nil, tr.rootHash)
	}
}

func TestTrie_Stress(t *testing.T) {
	COUNT := int64(10000)
	storage, _ := storage.NewMemoryStorage()
	tr, _ := NewTrie(nil, storage)

	keys := make([][]byte, COUNT)
	for i := int64(0); i < COUNT; i++ {
		bytes := byteutils.FromInt64(i)
		for j := 0; j < 4; j++ {
			bytes = append(bytes, bytes...)
		}
		keys[i] = bytes
	}
	fmt.Println(len(keys[0]))
	// 128

	startAt := time.Now().UnixNano()
	for i := int64(0); i < COUNT; i++ {
		tr.Put(keys[i], keys[i])
	}
	endAt := time.Now().UnixNano()
	fmt.Printf("%d Put, cost %d\n", COUNT, endAt-startAt)
	// 10000 Put, cost 509313000

	startAt = time.Now().UnixNano()
	for i := int64(0); i < COUNT; i++ {
		tr.Get(keys[i])
	}
	endAt = time.Now().UnixNano()
	fmt.Printf("%d Get, cost %d\n", COUNT, endAt-startAt)
	// 10000 Get, cost 396201000
}
