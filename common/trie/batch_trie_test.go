package trie

import (
	"fmt"
	"testing"
	"time"

	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/stretchr/testify/assert"
)

func TestBatchTrie_Clone(t *testing.T) {
	storage, _ := storage.NewMemoryStorage()
	tr1, _ := NewBatchTrie(nil, storage)
	tr2, _ := tr1.Clone()
	assert.Equal(t, tr1.trie.RootHash(), tr2.trie.RootHash())
	assert.Equal(t, tr1 == tr2, false)
	assert.Equal(t, tr1.trie == tr2.trie, false)
}

func TestBatchTrie_Batch(t *testing.T) {
	storage, _ := storage.NewMemoryStorage()
	tr, _ := NewBatchTrie(nil, storage)
	assert.Equal(t, []byte(nil), tr.trie.RootHash())

	tr.BeginBatch()

	// add a new leaf node
	addr1, _ := byteutils.FromHex("1f345678e9")
	val1 := []byte("leaf 1")
	tr.Put(addr1, val1)

	// add a new leaf node with 3-length common prefix
	addr2, _ := byteutils.FromHex("1f355678e9")
	val2 := []byte("leaf 2")
	tr.Put(addr2, val2)

	tr.Commit()
	tr.BeginBatch()

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

func TestBatchTrie_Iterator(t *testing.T) {
	storage, _ := storage.NewMemoryStorage()
	tr, _ := NewBatchTrie(nil, storage)
	assert.Equal(t, []byte(nil), tr.trie.RootHash())

	domain1 := []string{"a", "b", "c", "d"}
	value1 := []byte("abcd")
	domain2 := []string{"a", "b", "c", "e"}
	value2 := []byte("abce")
	domain3 := []string{"a", "b", "d"}
	value3 := []byte("abd")
	domain4 := []string{"a", "b", "e"}
	value4 := []byte("abe")
	domain5 := []string{"a", "c"}
	value5 := []byte("ac")
	domain6 := []string{"a", "e"}
	value6 := []byte("ae")

	tr.Put(HashDomains(domain1...), value1)
	tr.Put(HashDomains(domain2...), value2)
	tr.Put(HashDomains(domain3...), value3)
	tr.Put(HashDomains(domain4...), value4)
	tr.Put(HashDomains(domain5...), value5)
	tr.Put(HashDomains(domain6...), value6)

	// traverse abcd
	it, err := tr.Iterator(HashDomainsPrefix("a", "b", "c", "d"))
	assert.Nil(t, err)
	next, err := it.Next()
	assert.Nil(t, err)
	assert.Equal(t, next, true)
	assert.Equal(t, it.Value(), value1)
	next, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, next, false)

	// traverse abcd, abce
	it, err = tr.Iterator(HashDomainsPrefix("a", "b", "c"))
	assert.Nil(t, err)
	next, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, next, true)
	assert.Equal(t, it.Value(), value2)
	next, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, next, true)
	assert.Equal(t, it.Value(), value1)
	next, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, next, false)

	// traverse abd, abe, abcd, abce
	it, err = tr.Iterator(HashDomainsPrefix("a", "b"))
	assert.Nil(t, err)
	next, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, next, true)
	assert.Equal(t, it.Value(), value2)
	next, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, next, true)
	assert.Equal(t, it.Value(), value1)
	next, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, next, true)
	assert.Equal(t, it.Value(), value4)
	next, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, next, true)
	assert.Equal(t, it.Value(), value3)
	next, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, next, false)

	// traverse ae, ad, abd, abe, abcd, abce
	it, err = tr.Iterator(HashDomainsPrefix("a"))
	assert.Nil(t, err)
	next, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, next, true)
	assert.Equal(t, it.Value(), value5)
	next, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, next, true)
	assert.Equal(t, it.Value(), value6)
	next, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, next, true)
	assert.Equal(t, it.Value(), value2)
	next, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, next, true)
	assert.Equal(t, it.Value(), value1)
	next, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, next, true)
	assert.Equal(t, it.Value(), value4)
	next, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, next, true)
	assert.Equal(t, it.Value(), value3)
	next, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, next, false)

	// nothing
	it, err = tr.Iterator(HashDomainsPrefix("b"))
	assert.NotNil(t, err)
}

func TestBatchTrie_Stress_1(t *testing.T) {
	COUNT := int64(10000)
	storage, _ := storage.NewMemoryStorage()
	tr, _ := NewBatchTrie(nil, storage)

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
	// 10000 Put, cost 762181000

	startAt = time.Now().UnixNano()
	for i := int64(0); i < COUNT; i++ {
		tr.Get(keys[i])
	}
	endAt = time.Now().UnixNano()
	fmt.Printf("%d Get, cost %d\n", COUNT, endAt-startAt)
	// 10000 Get, cost 385232000
}

func TestBatchTrie_Stress_2(t *testing.T) {
	COUNT := int64(10000)
	storage, _ := storage.NewMemoryStorage()
	tr, _ := NewBatchTrie(nil, storage)

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
	tr.BeginBatch()
	for i := int64(0); i < COUNT; i++ {
		tr.Put(keys[i], keys[i])
	}
	tr.Commit()
	endAt := time.Now().UnixNano()
	fmt.Printf("%d Put, cost %d\n", COUNT, endAt-startAt)
	// 10000 Put, cost 760037000

	startAt = time.Now().UnixNano()
	tr.BeginBatch()
	for i := int64(0); i < COUNT; i++ {
		tr.Get(keys[i])
	}
	tr.Commit()
	endAt = time.Now().UnixNano()
	fmt.Printf("%d Get, cost %d\n", COUNT, endAt-startAt)
	// 10000 Get, cost 423994000
}

func TestBatchTrie_Test1(t *testing.T) {
	storage, _ := storage.NewMemoryStorage()
	tr, _ := NewTrie(nil, storage)
	assert.Equal(t, []byte(nil), tr.RootHash())

	var err error

	// put key1.
	_, err = tr.Put([]byte("key1"), []byte("value1"))
	assert.Nil(t, err)

	// get key1, should pass.
	val, err := tr.Get([]byte("key1"))
	assert.Nil(t, err)
	assert.Equal(t, []byte("value1"), val)

	// put key2, change the t.RootHash.
	_, err = tr.Put([]byte("key2"), []byte("value2"))
	assert.Nil(t, err)

	// get key1, should pass.
	val, err = tr.Get([]byte("key1"))
	assert.Nil(t, err)
	assert.Equal(t, []byte("value1"), val)

	// del key1.
	_, err = tr.Del([]byte("key1"))
	assert.Nil(t, err)

	_, err = tr.Get([]byte("key1"))
	assert.NotNil(t, err)
}
