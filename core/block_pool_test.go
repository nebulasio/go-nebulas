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

package core

import (
	"testing"

	"github.com/nebulasio/go-nebulas/core/pb"

	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/crypto"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/keystore/secp256k1"
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/net/messages"
	"github.com/nebulasio/go-nebulas/net/p2p"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/stretchr/testify/assert"
)

type MockConsensus struct {
	storage storage.Storage
}

func (c MockConsensus) FastVerifyBlock(block *Block) error {
	block.miner = block.Coinbase()
	return nil
}
func (c MockConsensus) VerifyBlock(block *Block, parent *Block) error {
	block.miner = block.Coinbase()
	return nil
}

var (
	received = []byte{}
)

type MockNetManager struct{}

func (n MockNetManager) Start() error { return nil }
func (n MockNetManager) Stop()        {}

func (n MockNetManager) Node() *p2p.Node { return nil }

func (n MockNetManager) Sync(net.Serializable) error            { return nil }
func (n MockNetManager) SendSyncReply(string, net.Serializable) {}

func (n MockNetManager) Register(...*net.Subscriber)   {}
func (n MockNetManager) Deregister(...*net.Subscriber) {}

func (n MockNetManager) Broadcast(name string, msg net.Serializable) {}
func (n MockNetManager) Relay(name string, msg net.Serializable)     {}
func (n MockNetManager) SendMsg(name string, msg []byte, target string) error {
	received = msg
	return nil
}

func (n MockNetManager) BroadcastNetworkID([]byte) {}

func (n MockNetManager) BuildData([]byte, string) []byte { return nil }

func TestBlockPool(t *testing.T) {
	received = []byte{}

	neb := testNeb()
	bc, err := NewBlockChain(neb)
	var n MockNetManager
	bc.bkPool.RegisterInNetwork(n)
	cons := &MockConsensus{neb.storage}
	bc.SetConsensusHandler(cons)
	pool := bc.bkPool
	assert.Equal(t, pool.cache.Len(), 0)

	ks := keystore.DefaultKS
	priv := secp256k1.GeneratePrivateKey()
	pubdata, _ := priv.PublicKey().Encoded()
	from, _ := NewAddressFromPublicKey(pubdata)
	ks.SetKey(from.String(), priv, []byte("passphrase"))
	ks.Unlock(from.String(), []byte("passphrase"), time.Second*60*60*24*365)
	to := &Address{from.address}
	key, _ := ks.GetUnlocked(from.String())
	signature, _ := crypto.NewSignature(keystore.SECP256K1)
	signature.InitSign(key.(keystore.PrivateKey))
	bc.tailBlock.begin()
	balance := util.NewUint128FromBigInt(util.NewUint128().Mul(TransactionGasPrice.Int, util.NewUint128FromInt(2000000).Int))
	bc.tailBlock.accState.GetOrCreateUserAccount(from.Bytes()).AddBalance(balance)
	bc.tailBlock.header.stateRoot = bc.tailBlock.accState.RootHash()
	bc.tailBlock.commit()
	bc.storeBlockToStorage(bc.tailBlock)

	validators, err := TraverseDynasty(bc.tailBlock.dposContext.dynastyTrie)
	assert.Nil(t, err)

	addr := &Address{validators[1]}
	block0, err := NewBlock(bc.ChainID(), addr, bc.tailBlock)
	assert.Nil(t, err)
	block0.header.timestamp = bc.tailBlock.header.timestamp + BlockInterval
	block0.SetMiner(addr)
	block0.Seal()

	tx1 := NewTransaction(bc.ChainID(), from, to, util.NewUint128FromInt(1), 1, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx1.Sign(signature)
	tx2 := NewTransaction(bc.ChainID(), from, to, util.NewUint128FromInt(2), 2, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx2.Sign(signature)
	tx3 := NewTransaction(bc.ChainID(), from, to, util.NewUint128FromInt(3), 3, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx3.Sign(signature)
	err = bc.txPool.Push(tx1)
	assert.NoError(t, err)
	err = bc.txPool.Push(tx2)
	assert.NoError(t, err)
	err = bc.txPool.Push(tx3)
	assert.NoError(t, err)

	addr = &Address{validators[2]}
	block1, _ := NewBlock(bc.ChainID(), addr, block0)
	block1.header.timestamp = block0.header.timestamp + BlockInterval
	block1.CollectTransactions(1)
	block1.SetMiner(addr)
	block1.Seal()

	addr = &Address{validators[3]}
	block2, _ := NewBlock(bc.ChainID(), addr, block1)
	block2.header.timestamp = block1.header.timestamp + BlockInterval
	block2.CollectTransactions(1)
	block2.SetMiner(addr)
	block2.Seal()

	addr = &Address{validators[4]}
	block3, _ := NewBlock(bc.ChainID(), addr, block2)
	block3.header.timestamp = block2.header.timestamp + BlockInterval
	block3.CollectTransactions(1)
	block3.SetMiner(addr)
	block3.Seal()

	addr = &Address{validators[5]}
	block4, _ := NewBlock(bc.ChainID(), addr, block3)
	block4.header.timestamp = block3.header.timestamp + BlockInterval
	block4.CollectTransactions(1)
	block4.SetMiner(addr)
	block4.Seal()

	err = pool.Push(block0)
	assert.NoError(t, err)
	assert.Equal(t, pool.cache.Len(), 0)
	err = pool.PushAndBroadcast(block0)
	assert.Equal(t, err, ErrDuplicatedBlock)

	err = pool.Push(block3)
	assert.Equal(t, pool.cache.Len(), 1)
	assert.Error(t, ErrMissingParentBlock)
	err = pool.Push(block4)
	assert.Equal(t, pool.cache.Len(), 2)
	assert.NoError(t, err)
	err = pool.Push(block2)
	assert.Equal(t, pool.cache.Len(), 3)
	assert.Error(t, ErrMissingParentBlock)

	err = pool.Push(block1)
	assert.NoError(t, err)
	assert.Equal(t, pool.cache.Len(), 0)

	bc.SetTailBlock(block0)
	nonce := bc.tailBlock.GetNonce(from.Bytes())
	assert.Equal(t, nonce, uint64(0))

	bc.SetTailBlock(block1)
	nonce = bc.tailBlock.GetNonce(from.Bytes())
	assert.Equal(t, nonce, uint64(1))
	bc.SetTailBlock(block2)
	nonce = bc.tailBlock.GetNonce(from.Bytes())
	assert.Equal(t, nonce, uint64(2))
	bc.SetTailBlock(block3)
	nonce = bc.tailBlock.GetNonce(from.Bytes())
	assert.Equal(t, nonce, uint64(3))

	addr = &Address{validators[0]}
	block5, _ := NewBlock(bc.ChainID(), addr, block4)
	block5.header.timestamp = block4.header.timestamp + BlockInterval
	block5.CollectTransactions(1)
	block5.SetMiner(addr)
	block5.Seal()
	block5.header.hash[0]++
	assert.Equal(t, pool.Push(block5), ErrInvalidBlockHash)

	addr = &Address{validators[1]}
	block41, _ := NewBlock(bc.ChainID(), addr, block3)
	block41.header.timestamp = block3.header.timestamp + BlockInterval
	block41.CollectTransactions(1)
	block41.SetMiner(addr)
	block41.Seal()
	assert.Equal(t, pool.Push(block41), ErrDoubleBlockMinted)

	addr = &Address{validators[0]}
	block6, _ := NewBlock(bc.ChainID(), addr, block5)
	block6.header.timestamp = block3.header.timestamp + BlockInterval*DynastySize - 1
	block6.CollectTransactions(1)
	block6.SetMiner(addr)
	block6.Seal()
	assert.Equal(t, pool.push("fake", block6), ErrInvalidBlockCannotFindParentInLocal)
	downloadMsg := &corepb.DownloadBlock{
		Hash: block6.Hash(),
		Sign: block6.Signature(),
	}
	bytes, _ := proto.Marshal(downloadMsg)
	assert.Equal(t, received, bytes)

	received = []byte{}
	addr = &Address{validators[0]}
	block7, _ := NewBlock(bc.ChainID(), addr, block5)
	block7.header.timestamp = block3.header.timestamp + BlockInterval*DynastySize + 1
	block7.CollectTransactions(1)
	block7.SetMiner(addr)
	block7.Seal()
	assert.Equal(t, pool.push("fake", block7), nil)
	assert.Equal(t, received, []byte{})
}

func TestHandleBlock(t *testing.T) {
	neb := testNeb()
	bc, err := NewBlockChain(neb)
	var n MockNetManager
	bc.bkPool.RegisterInNetwork(n)
	assert.Nil(t, err)
	cons := &MockConsensus{neb.storage}
	bc.SetConsensusHandler(cons)
	from := mockAddress()
	ks := keystore.DefaultKS
	key, err := ks.GetUnlocked(from.String())
	signature, _ := crypto.NewSignature(keystore.SECP256K1)
	signature.InitSign(key.(keystore.PrivateKey))

	block, err := bc.NewBlock(from)
	assert.Nil(t, err)
	block.SetMiner(from)
	block.Seal()
	block.Sign(signature)
	pbMsg, err := block.ToProto()
	assert.Nil(t, err)
	data, err := proto.Marshal(pbMsg)
	assert.Nil(t, err)
	msg := messages.NewBaseMessage(MessageTypeNewTx, "from", data)
	bc.bkPool.handleBlock(msg)
	assert.Nil(t, bc.GetBlock(block.Hash()))

	block, err = bc.NewBlock(from)
	assert.Nil(t, err)
	block.header.timestamp = time.Now().Unix() - AcceptedNetWorkDelay - 1
	block.SetMiner(from)
	block.Seal()
	block.Sign(signature)
	pbMsg, err = block.ToProto()
	assert.Nil(t, err)
	data, err = proto.Marshal(pbMsg)
	msg = messages.NewBaseMessage(MessageTypeNewBlock, "from", data)
	bc.bkPool.handleBlock(msg)
	assert.Nil(t, bc.GetBlock(block.Hash()))

	block, err = bc.NewBlock(from)
	assert.Nil(t, err)
	block.header.timestamp = 0
	assert.Nil(t, err)
	block.SetMiner(from)
	block.Seal()
	block.Sign(signature)
	pbMsg, err = block.ToProto()
	assert.Nil(t, err)
	data, err = proto.Marshal(pbMsg)
	msg = messages.NewBaseMessage(MessageTypeNewBlock, "from", data)
	bc.bkPool.handleBlock(msg)
	assert.Nil(t, bc.GetBlock(block.Hash()))

	block, err = bc.NewBlock(from)
	assert.Nil(t, err)
	block.header.timestamp = 0
	assert.Nil(t, err)
	block.SetMiner(from)
	block.Seal()
	block.Sign(signature)
	pbMsg, err = block.ToProto()
	assert.Nil(t, err)
	data, err = proto.Marshal(pbMsg)
	msg = messages.NewBaseMessage(MessageTypeDownloadedBlockReply, "from", data)
	bc.bkPool.handleBlock(msg)
	assert.NotNil(t, bc.GetBlock(block.Hash()))
}

func TestHandleDownloadedBlock(t *testing.T) {
	received = []byte{}

	neb := testNeb()
	bc, err := NewBlockChain(neb)
	assert.Nil(t, err)
	var n MockNetManager
	bc.bkPool.RegisterInNetwork(n)
	assert.Equal(t, n, bc.bkPool.nm)
	cons := &MockConsensus{neb.storage}
	bc.SetConsensusHandler(cons)
	from := mockAddress()
	ks := keystore.DefaultKS
	key, err := ks.GetUnlocked(from.String())
	signature, _ := crypto.NewSignature(keystore.SECP256K1)
	signature.InitSign(key.(keystore.PrivateKey))

	block1, err := bc.NewBlock(from)
	assert.Nil(t, err)
	block1.SetMiner(from)
	block1.Seal()
	block1.Sign(signature)
	bc.SetTailBlock(block1)

	block2, err := bc.NewBlock(from)
	assert.Nil(t, err)
	block2.SetMiner(from)
	block2.Seal()
	block2.Sign(signature)
	bc.SetTailBlock(block2)
	bc.storeBlockToStorage(block2)

	downloadBlock := new(corepb.DownloadBlock)
	downloadBlock.Hash = block1.Hash()
	downloadBlock.Sign = block1.Signature()
	data, err := proto.Marshal(downloadBlock)
	assert.Nil(t, err)
	msg := messages.NewBaseMessage(MessageTypeNewBlock, "from", data)
	bc.bkPool.handleDownloadedBlock(msg)
	assert.Equal(t, received, []byte{})

	downloadBlock = new(corepb.DownloadBlock)
	downloadBlock.Hash = bc.genesisBlock.Hash()
	downloadBlock.Sign = bc.genesisBlock.Signature()
	data, err = proto.Marshal(downloadBlock)
	assert.Nil(t, err)
	msg = messages.NewBaseMessage(MessageTypeDownloadedBlock, "from", data)
	bc.bkPool.handleDownloadedBlock(msg)
	assert.Equal(t, received, []byte{})

	downloadBlock = new(corepb.DownloadBlock)
	downloadBlock.Hash = block1.Hash()
	downloadBlock.Sign = block1.Signature()
	data, err = proto.Marshal(downloadBlock)
	assert.Nil(t, err)
	msg = messages.NewBaseMessage(MessageTypeDownloadedBlock, "from", data)
	bc.bkPool.handleDownloadedBlock(msg)
	assert.Equal(t, received, []byte{})
	bc.storeBlockToStorage(block1)

	downloadBlock = new(corepb.DownloadBlock)
	downloadBlock.Hash = block1.Hash()
	downloadBlock.Sign = block2.Signature()
	data, err = proto.Marshal(downloadBlock)
	assert.Nil(t, err)
	msg = messages.NewBaseMessage(MessageTypeDownloadedBlock, "from", data)
	bc.bkPool.handleDownloadedBlock(msg)
	assert.Equal(t, received, []byte{})

	downloadBlock = new(corepb.DownloadBlock)
	downloadBlock.Hash = block1.Hash()
	downloadBlock.Sign = block1.Signature()
	data, err = proto.Marshal(downloadBlock)
	assert.Nil(t, err)
	msg = messages.NewBaseMessage(MessageTypeDownloadedBlock, "from", data)
	bc.bkPool.handleDownloadedBlock(msg)
	pbGenesis, err := bc.genesisBlock.ToProto()
	assert.Nil(t, err)
	data, err = proto.Marshal(pbGenesis)
	assert.Nil(t, err)
	assert.Equal(t, received, data)
}
