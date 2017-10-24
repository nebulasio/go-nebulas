package batchtrie

import (
	"testing"

	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/stretchr/testify/assert"
)

func TestBatchTrie_Clone(t *testing.T) {
	tr1, _ := NewBatchTrie(nil)
	tr2, _ := tr1.Clone()
	assert.Equal(t, tr1.trie.RootHash(), tr2.trie.RootHash())
	assert.Equal(t, tr1 == tr2, false)
	assert.Equal(t, tr1.trie == tr2.trie, false)
}

func TestBatchTrie(t *testing.T) {
	tr, _ := NewBatchTrie(nil)
	assert.Equal(t, []byte(nil), tr.trie.RootHash())

	err1 := tr.BeginBatch()
	assert.Nil(t, err1)

	// add a new leaf node
	addr1, _ := byteutils.FromHex("1f345678e9")
	val1 := []byte("leaf 1")
	tr.Put(addr1, val1)

	err2 := tr.BeginBatch()
	assert.NotNil(t, err2)

	// add a new leaf node with 3-length common prefix
	addr2, _ := byteutils.FromHex("1f355678e9")
	val2 := []byte("leaf 2")
	tr.Put(addr2, val2)

	tr.Commit()
	err3 := tr.BeginBatch()
	assert.Nil(t, err3)

	// add a new node with 2-length common prefix
	addr3, _ := byteutils.FromHex("1f555678e9")
	val3 := []byte{}
	tr.Put(addr3, val3)
	// update node leaf1
	val11 := []byte("leaf 11")
	tr.Put(addr1, val11)
	tr.Del(addr2)
	tr.Del(addr1)

	tr.RollBack()

	// get node "1f345678e9"
	checkVal1, _ := tr.Get(addr1)
	assert.Equal(t, checkVal1, val1)
	// get node "1f355678e9"
	checkVal2, _ := tr.Get(addr2)
	assert.Equal(t, checkVal2, val2)
	// get node "1f555678e9"
	_, err4 := tr.Get(addr3)
	assert.NotNil(t, err4)
}
