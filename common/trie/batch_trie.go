package trie

import (
	"errors"

	"github.com/nebulasio/go-nebulas/crypto/hash"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

// Errors
var (
	ErrCloneInBatch      = errors.New("cannot clone AccountState with a batch task unfinished")
	ErrBeginAgainInBatch = errors.New("cannot begin AccountState with a batch task unfinished")
)

// Action represents operation types in BatchTrie
type Action int

// Action constants
const (
	Insert Action = iota
	Update
	Delete
)

// Entry in changelog, [key, old value, new value]
type Entry struct {
	action Action
	key    []byte
	old    []byte
	update []byte
}

// BatchTrie is a trie that supports batch task
type BatchTrie struct {
	trie      *Trie
	changelog []*Entry
	batching  bool
}

// NewBatchTrie if rootHash is nil, create a new BatchTrie, otherwise, build an existed BatchTrie
func NewBatchTrie(rootHash []byte, storage storage.Storage) (*BatchTrie, error) {
	t, err := NewTrie(rootHash, storage)
	if err != nil {
		return nil, err
	}
	return &BatchTrie{trie: t, batching: false}, nil
}

// RootHash of the BatchTrie
func (bt *BatchTrie) RootHash() []byte {
	return bt.trie.RootHash()
}

// Clone a the BatchTrie
func (bt *BatchTrie) Clone() (*BatchTrie, error) {
	if bt.batching {
		return nil, ErrCloneInBatch
	}
	tr, err := bt.trie.Clone()
	if err != nil {
		return nil, err
	}
	return &BatchTrie{trie: tr, batching: false}, nil
}

// Get the value to the key in BatchTrie
// return value to the key
func (bt *BatchTrie) Get(key []byte) ([]byte, error) {
	return bt.trie.Get(key)
}

// Put the key-value pair in BatchTrie
// return new rootHash
func (bt *BatchTrie) Put(key []byte, val []byte) ([]byte, error) {
	entry := &Entry{Update, key, nil, val}
	old, getErr := bt.trie.Get(key)
	if getErr != nil {
		entry.action = Insert
	} else {
		entry.old = old
	}
	rootHash, putErr := bt.trie.Put(key, val)
	if putErr != nil {
		return nil, putErr
	}
	if bt.batching {
		bt.changelog = append(bt.changelog, entry)
	}
	return rootHash, nil
}

// Del the key-value pair in BatchTrie
// return new rootHash
func (bt *BatchTrie) Del(key []byte) ([]byte, error) {
	entry := &Entry{Delete, key, nil, nil}
	old, getErr := bt.trie.Get(key)
	if getErr == nil {
		entry.old = old
	}
	rootHash, err := bt.trie.Del(key)
	if err != nil {
		return nil, err
	}
	if bt.batching {
		bt.changelog = append(bt.changelog, entry)
	}
	return rootHash, nil
}

// SyncTrie data from other servers
// Sync whole trie to build snapshot
func (bt *BatchTrie) SyncTrie(rootHash []byte) error {
	return bt.trie.SyncTrie(rootHash)
}

// SyncPath from rootHash to key node from other servers
// Useful for verification quickly
func (bt *BatchTrie) SyncPath(rootHash []byte, key []byte) error {
	return bt.trie.SyncPath(rootHash, key)
}

// Prove the associated node to the key exists in trie
// if exists, MerkleProof is a complete path from root to the node
// otherwise, MerkleProof is nil
func (bt *BatchTrie) Prove(key []byte) (MerkleProof, error) {
	return bt.trie.Prove(key)
}

// Verify whether the merkle proof from root to the associated node is right
func (bt *BatchTrie) Verify(rootHash []byte, key []byte, proof MerkleProof) error {
	return bt.trie.Verify(rootHash, key, proof)
}

// Empty return if the trie is empty
func (bt *BatchTrie) Empty() bool {
	return bt.trie.Empty()
}

// Iterator return an trie Iterator to traverse leaf node's value in this trie
func (bt *BatchTrie) Iterator(prefix []byte) (*Iterator, error) {
	return bt.trie.Iterator(prefix)
}

// BeginBatch to process a batch task
func (bt *BatchTrie) BeginBatch() error {
	if bt.batching {
		return ErrBeginAgainInBatch
	}
	bt.batching = true
	return nil
}

// Commit a batch task
func (bt *BatchTrie) Commit() {
	// clear changelog
	bt.changelog = bt.changelog[:0]
	bt.batching = false
}

// RollBack a batch task
func (bt *BatchTrie) RollBack() {
	// compress changelog
	changelog := make(map[string]*Entry)
	for _, entry := range bt.changelog {
		if _, ok := changelog[byteutils.Hex(entry.key)]; !ok {
			changelog[byteutils.Hex(entry.key)] = entry
		}
	}
	// clear changelog
	bt.changelog = bt.changelog[:0]
	// rollback
	for _, entry := range changelog {
		switch entry.action {
		case Insert:
			bt.trie.Del(entry.key)
		case Update, Delete:
			bt.trie.Put(entry.key, entry.old)
		}
	}
	bt.batching = false
}

// HashDomains for each variable in contract
// each domain will represented as 6 bytes, support 4 level domain at most
// such as,
// 4a56b7 000000 000000 000000,
// 4a56b8 1c9812 000000 000000,
// 4a56b8 3a1289 000000 000000,
// support iterator with same prefix
func HashDomains(domains ...string) []byte {
	if len(domains) > 24/6 {
		panic("only support 4 level domain at most")
	}
	key := [24]byte{0}
	for k, v := range domains {
		domain := hash.Sha3256([]byte(v))[0:6]
		for i := 0; i < len(domain); i++ {
			key[k*6+i] = domain[i]
		}
	}
	return key[:]
}

// HashDomainsPrefix is same as HashDomains, but without tail zeros
func HashDomainsPrefix(domains ...string) []byte {
	if len(domains) > 24/6 {
		panic("only support 4 level domain at most")
	}
	key := []byte{}
	for _, v := range domains {
		domain := hash.Sha3256([]byte(v))[0:6]
		key = append(key, domain...)
	}
	return key[:]
}
