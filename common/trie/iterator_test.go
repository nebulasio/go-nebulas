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
	"testing"

	"github.com/gogo/protobuf/proto"
	triepb "github.com/nebulasio/go-nebulas/common/trie/pb"
	"github.com/nebulasio/go-nebulas/crypto/hash"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/stretchr/testify/assert"
)

func TestIterator1(t *testing.T) {
	storage, _ := storage.NewMemoryStorage()
	tr, err := NewTrie(nil, storage, false)
	assert.Nil(t, err)
	names := []string{"123450", "123350", "122450", "223350", "133350"}
	keys := [][]byte{}
	for _, v := range names {
		key, err := byteutils.FromHex(v)
		assert.Nil(t, err)
		keys = append(keys, key)
	}
	tr.Put(keys[0], []byte(names[0]))
	leaf1 := [][]byte{[]byte{byte(leaf)}, keyToRoute(keys[0]), []byte(names[0])}
	leaf1IR, _ := proto.Marshal(&triepb.Node{Val: leaf1})
	leaf1H := hash.Sha3256(leaf1IR)
	assert.Equal(t, leaf1H, tr.rootHash)

	tr.Put(keys[1], []byte(names[1]))
	leaf2 := [][]byte{[]byte{byte(leaf)}, keyToRoute(keys[0])[4:], []byte(names[0])}
	leaf2IR, _ := proto.Marshal(&triepb.Node{Val: leaf2})
	leaf2H := hash.Sha3256(leaf2IR)
	leaf3 := [][]byte{[]byte{byte(leaf)}, keyToRoute(keys[1])[4:], []byte(names[1])}
	leaf3IR, _ := proto.Marshal(&triepb.Node{Val: leaf3})
	leaf3H := hash.Sha3256(leaf3IR)
	branch2 := [][]byte{nil, nil, nil, leaf3H, leaf2H, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}
	branch2IR, _ := proto.Marshal(&triepb.Node{Val: branch2})
	branch2H := hash.Sha3256(branch2IR)
	ext1 := [][]byte{[]byte{(byte(ext))}, keyToRoute(keys[0])[:3], branch2H}
	ext1IR, _ := proto.Marshal(&triepb.Node{Val: ext1})
	ext1H := hash.Sha3256(ext1IR)
	assert.Equal(t, ext1H, tr.RootHash())

	tr.Put(keys[2], []byte(names[2]))
	leaf4 := [][]byte{[]byte{byte(leaf)}, keyToRoute(keys[2])[3:], []byte(names[2])}
	leaf4IR, _ := proto.Marshal(&triepb.Node{Val: leaf4})
	leaf4H := hash.Sha3256(leaf4IR)
	branch4 := [][]byte{nil, nil, leaf4H, branch2H, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}
	branch4IR, _ := proto.Marshal(&triepb.Node{Val: branch4})
	branch4H := hash.Sha3256(branch4IR)
	ext2 := [][]byte{[]byte{(byte(ext))}, keyToRoute(keys[2])[:2], branch4H}
	ext2IR, _ := proto.Marshal(&triepb.Node{Val: ext2})
	ext2H := hash.Sha3256(ext2IR)
	assert.Equal(t, ext2H, tr.rootHash)

	tr.Put(keys[3], []byte(names[3]))
	leaf5 := [][]byte{[]byte{byte(leaf)}, keyToRoute(keys[3])[1:], []byte(names[3])}
	leaf5IR, _ := proto.Marshal(&triepb.Node{Val: leaf5})
	leaf5H := hash.Sha3256(leaf5IR)
	ext3 := [][]byte{[]byte{(byte(ext))}, keyToRoute(keys[0])[1:2], branch4H}
	ext3IR, _ := proto.Marshal(&triepb.Node{Val: ext3})
	ext3H := hash.Sha3256(ext3IR)
	branch5 := [][]byte{nil, ext3H, leaf5H, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}
	branch5IR, _ := proto.Marshal(&triepb.Node{Val: branch5})
	branch5H := hash.Sha3256(branch5IR)
	assert.Equal(t, branch5H, tr.rootHash)

	tr.Put(keys[4], []byte(names[4]))
	leaf6 := [][]byte{[]byte{byte(leaf)}, keyToRoute(keys[4])[2:], []byte(names[4])}
	leaf6IR, _ := proto.Marshal(&triepb.Node{Val: leaf6})
	leaf6H := hash.Sha3256(leaf6IR)
	branch6 := [][]byte{nil, nil, branch4H, leaf6H, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}
	branch6IR, _ := proto.Marshal(&triepb.Node{Val: branch6})
	branch6H := hash.Sha3256(branch6IR)
	branch5[1] = branch6H
	branch5IR, _ = proto.Marshal(&triepb.Node{Val: branch5})
	branch5H = hash.Sha3256(branch5IR)
	assert.Equal(t, branch5H, tr.rootHash)

	it, err := tr.Iterator([]byte{0x12})
	assert.Nil(t, err)
	next, err := it.Next()
	assert.Nil(t, err)
	assert.Equal(t, next, true)
	assert.Equal(t, it.Value(), []byte(names[2]))
	assert.Equal(t, it.Key(), []byte(keys[2]))

	next, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, next, true)
	assert.Equal(t, it.Value(), []byte(names[1]))
	assert.Equal(t, it.Key(), []byte(keys[1]))

	next, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, next, true)
	assert.Equal(t, it.Value(), []byte(names[0]))
	assert.Equal(t, it.Key(), []byte(keys[0]))

	next, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, next, false)
}

func TestIterator2(t *testing.T) {
	storage, _ := storage.NewMemoryStorage()
	tr, err := NewTrie(nil, storage, false)
	assert.Nil(t, err)
	names := []string{"123450", "123350", "122450", "223350", "133350"}
	keys := [][]byte{}
	for _, v := range names {
		key, err := byteutils.FromHex(v)
		assert.Nil(t, err)
		keys = append(keys, key)
	}
	tr.Put(keys[0], []byte(names[0]))

	_, err1 := tr.Iterator([]byte{0x12, 0x34, 0x50, 0x12})
	assert.NotNil(t, err1)

	it, err := tr.Iterator([]byte{0x12})
	assert.Nil(t, err)
	fmt.Println(err)

	next, err := it.Next()
	assert.Nil(t, err)
	assert.Equal(t, next, true)
	assert.Equal(t, it.Value(), []byte(names[0]))
	assert.Equal(t, it.Key(), []byte(keys[0]))

	tr.Put(keys[1], []byte(names[1]))
	it, err = tr.Iterator([]byte{0x12})
	assert.Nil(t, err)
	next, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, next, true)
	assert.Equal(t, it.Value(), []byte(names[1]))
	assert.Equal(t, it.Key(), []byte(keys[1]))
	next, err = it.Next()
	assert.Equal(t, next, true)
	assert.Equal(t, it.Value(), []byte(names[0]))
	assert.Equal(t, it.Key(), []byte(keys[0]))

	tr.Put(keys[2], []byte(names[2]))

	it, err = tr.Iterator(nil)
	assert.Nil(t, err)
	next, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, next, true)
	assert.Equal(t, it.Value(), []byte(names[2]))
	assert.Equal(t, it.Key(), []byte(keys[2]))

	next, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, next, true)
	assert.Equal(t, it.Value(), []byte(names[1]))
	assert.Equal(t, it.Key(), []byte(keys[1]))

	next, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, next, true)
	assert.Equal(t, it.Value(), []byte(names[0]))
	assert.Equal(t, it.Key(), []byte(keys[0]))

	next, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, next, false)
}

func TestIteratorEmpty(t *testing.T) {
	stor, _ := storage.NewMemoryStorage()
	tr, _ := NewTrie(nil, stor, false)
	iter, err := tr.Iterator([]byte("he"))
	assert.Nil(t, iter)
	assert.Equal(t, err, storage.ErrKeyNotFound)
	iter, err = tr.Iterator(nil)
	assert.Nil(t, iter)
	assert.Equal(t, err, storage.ErrKeyNotFound)
}
