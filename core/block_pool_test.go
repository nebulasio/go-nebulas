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

func (c MockConsensus) ForkChoice() error {
	return nil
}

func (c MockConsensus) SuspendMining() {}

func (c MockConsensus) ResumeMining() {}

var (
	received = []byte{}
)

type MockNetService struct{}

func (n MockNetService) Start() error { return nil }
func (n MockNetService) Stop()        {}

func (n MockNetService) Node() *net.Node { return nil }

func (n MockNetService) Sync(net.Serializable) error { return nil }

func (n MockNetService) Register(...*net.Subscriber)   {}
func (n MockNetService) Deregister(...*net.Subscriber) {}

func (n MockNetService) Broadcast(name string, msg net.Serializable, priority int) {}
func (n MockNetService) Relay(name string, msg net.Serializable, priority int)     {}
func (n MockNetService) SendMsg(name string, msg []byte, target string, priority int) error {
	received = msg
	return nil
}

func (n MockNetService) SendMessageToPeers(messageName string, data []byte, priority int, filter net.PeerFilterAlgorithm) []string {
	return make([]string, 0)
}
func (n MockNetService) SendMessageToPeer(messageName string, data []byte, priority int, peerID string) error {
	return nil
}

func (n MockNetService) ClosePeer(peerID string, reason error) {}

func (n MockNetService) BroadcastNetworkID([]byte) {}

func (n MockNetService) BuildRawMessageData([]byte, string) []byte { return nil }

func TestBlockPool(t *testing.T) {
	received = []byte{}

	neb := testNeb()
	bc, err := NewBlockChain(neb)
	var n MockNetService
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
	key, _ := ks.GetUnlocked(from.String())
	signature, _ := crypto.NewSignature(keystore.SECP256K1)
	signature.InitSign(key.(keystore.PrivateKey))
	bc.tailBlock.begin()
	balance := util.NewUint128FromBigInt(util.NewUint128().Mul(TransactionGasPrice.Int, util.NewUint128FromInt(2000000).Int))
	acc, err := bc.tailBlock.accState.GetOrCreateUserAccount(from.Bytes())
	assert.Nil(t, err)
	acc.AddBalance(balance)
	bc.tailBlock.header.stateRoot, err = bc.tailBlock.accState.RootHash()
	assert.Nil(t, err)
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

	addr = &Address{validators[2]}
	block1, err := NewBlock(bc.ChainID(), addr, block0)
	assert.Nil(t, err)
	block1.header.timestamp = block0.header.timestamp + BlockInterval
	block1.SetMiner(addr)
	block1.Seal()

	addr = &Address{validators[3]}
	block2, err := NewBlock(bc.ChainID(), addr, block1)
	assert.Nil(t, err)
	block2.header.timestamp = block1.header.timestamp + BlockInterval
	block2.SetMiner(addr)
	block2.Seal()

	addr = &Address{validators[4]}
	block3, err := NewBlock(bc.ChainID(), addr, block2)
	assert.Nil(t, err)
	block3.header.timestamp = block2.header.timestamp + BlockInterval
	block3.SetMiner(addr)
	block3.Seal()

	addr = &Address{validators[5]}
	block4, err := NewBlock(bc.ChainID(), addr, block3)
	assert.Nil(t, err)
	block4.header.timestamp = block3.header.timestamp + BlockInterval
	block4.SetMiner(addr)
	block4.Seal()

	err = pool.Push(block0)
	assert.NoError(t, err)
	assert.Equal(t, pool.cache.Len(), 0)
	err = pool.PushAndBroadcast(block0)
	assert.Equal(t, err, ErrDuplicatedBlock)

	err = pool.Push(block3)
	assert.Equal(t, pool.cache.Len(), 1)
	assert.Error(t, err, ErrMissingParentBlock)
	err = pool.Push(block4)
	assert.Equal(t, pool.cache.Len(), 2)
	assert.Error(t, err, ErrMissingParentBlock)
	err = pool.Push(block2)
	assert.Equal(t, pool.cache.Len(), 3)
	assert.Error(t, err, ErrMissingParentBlock)

	err = pool.Push(block1)
	assert.NoError(t, err)
	assert.Equal(t, pool.cache.Len(), 0)

	bc.SetTailBlock(block4)
	assert.Equal(t, bc.tailBlock.Hash(), block4.Hash())

	addr = &Address{validators[0]}
	block5, err := NewBlock(bc.ChainID(), addr, block4)
	assert.Nil(t, err)
	block5.header.timestamp = block4.header.timestamp + BlockInterval
	block5.SetMiner(addr)
	block5.Seal()
	block5.header.hash[0]++
	assert.Equal(t, pool.Push(block5), ErrInvalidBlockHash)

	addr = &Address{validators[1]}
	block41, err := NewBlock(bc.ChainID(), addr, block3)
	assert.Nil(t, err)
	block41.header.timestamp = block3.header.timestamp + BlockInterval
	block41.SetMiner(addr)
	block41.Seal()
	assert.Equal(t, pool.Push(block41), ErrDoubleBlockMinted)
}

func TestHandleBlock(t *testing.T) {
	neb := testNeb()
	bc, err := NewBlockChain(neb)
	var n MockNetService
	bc.bkPool.RegisterInNetwork(n)
	assert.Nil(t, err)
	cons := &MockConsensus{neb.storage}
	bc.SetConsensusHandler(cons)
	from := mockAddress()
	ks := keystore.DefaultKS
	key, err := ks.GetUnlocked(from.String())
	signature, err := crypto.NewSignature(keystore.SECP256K1)
	assert.Nil(t, err)
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
	msg := net.NewBaseMessage(MessageTypeNewTx, "from", data)
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
	msg = net.NewBaseMessage(MessageTypeNewBlock, "from", data)
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
	msg = net.NewBaseMessage(MessageTypeNewBlock, "from", data)
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
	msg = net.NewBaseMessage(MessageTypeDownloadedBlockReply, "from", data)
	bc.bkPool.handleBlock(msg)
	assert.NotNil(t, bc.GetBlock(block.Hash()))
}

func TestHandleDownloadedBlock(t *testing.T) {
	received = []byte{}

	neb := testNeb()
	bc, err := NewBlockChain(neb)
	assert.Nil(t, err)
	var n MockNetService
	bc.bkPool.RegisterInNetwork(n)
	assert.Equal(t, n, bc.bkPool.ns)
	cons := &MockConsensus{neb.storage}
	bc.SetConsensusHandler(cons)
	from := mockAddress()
	ks := keystore.DefaultKS
	key, err := ks.GetUnlocked(from.String())
	signature, err := crypto.NewSignature(keystore.SECP256K1)
	assert.Nil(t, err)
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
	msg := net.NewBaseMessage(MessageTypeNewBlock, "from", data)
	bc.bkPool.handleDownloadedBlock(msg)
	assert.Equal(t, received, []byte{})

	downloadBlock = new(corepb.DownloadBlock)
	downloadBlock.Hash = bc.genesisBlock.Hash()
	downloadBlock.Sign = bc.genesisBlock.Signature()
	data, err = proto.Marshal(downloadBlock)
	assert.Nil(t, err)
	msg = net.NewBaseMessage(MessageTypeDownloadedBlock, "from", data)
	bc.bkPool.handleDownloadedBlock(msg)
	assert.Equal(t, received, []byte{})

	downloadBlock = new(corepb.DownloadBlock)
	downloadBlock.Hash = block1.Hash()
	downloadBlock.Sign = block1.Signature()
	data, err = proto.Marshal(downloadBlock)
	assert.Nil(t, err)
	msg = net.NewBaseMessage(MessageTypeDownloadedBlock, "from", data)
	bc.bkPool.handleDownloadedBlock(msg)
	assert.Equal(t, received, []byte{})
	bc.storeBlockToStorage(block1)

	downloadBlock = new(corepb.DownloadBlock)
	downloadBlock.Hash = block1.Hash()
	downloadBlock.Sign = block2.Signature()
	data, err = proto.Marshal(downloadBlock)
	assert.Nil(t, err)
	msg = net.NewBaseMessage(MessageTypeDownloadedBlock, "from", data)
	bc.bkPool.handleDownloadedBlock(msg)
	assert.Equal(t, received, []byte{})

	downloadBlock = new(corepb.DownloadBlock)
	downloadBlock.Hash = block1.Hash()
	downloadBlock.Sign = block1.Signature()
	data, err = proto.Marshal(downloadBlock)
	assert.Nil(t, err)
	msg = net.NewBaseMessage(MessageTypeDownloadedBlock, "from", data)
	bc.bkPool.handleDownloadedBlock(msg)
	pbGenesis, err := bc.genesisBlock.ToProto()
	assert.Nil(t, err)
	data, err = proto.Marshal(pbGenesis)
	assert.Nil(t, err)
	assert.Equal(t, received, data)
}
