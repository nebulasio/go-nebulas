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
	"github.com/nebulasio/go-nebulas/util"
	"github.com/stretchr/testify/assert"
)

func TestBlockPool(t *testing.T) {
	received = []byte{}

	ks := keystore.DefaultKS
	priv := secp256k1.GeneratePrivateKey()
	pubdata, _ := priv.PublicKey().Encoded()
	from, _ := NewAddressFromPublicKey(pubdata)
	ks.SetKey(from.String(), priv, []byte("passphrase"))
	ks.Unlock(from.String(), []byte("passphrase"), time.Second*60*60*24*365)
	key, _ := ks.GetUnlocked(from.String())
	signature, _ := crypto.NewSignature(keystore.SECP256K1)
	signature.InitSign(key.(keystore.PrivateKey))

	neb := testNeb(t)
	bc := neb.chain
	pool := bc.bkPool
	assert.Equal(t, pool.cache.Len(), 0)

	bc.tailBlock.begin()
	baseGas, _ := util.NewUint128FromInt(2000000)
	balance, err := TransactionGasPrice.Mul(baseGas)
	assert.Nil(t, err)
	acc, err := bc.tailBlock.accState.GetOrCreateUserAccount(from.Bytes())
	assert.Nil(t, err)
	acc.AddBalance(balance)
	bc.tailBlock.header.stateRoot, err = bc.tailBlock.accState.RootHash()
	assert.Nil(t, err)
	bc.tailBlock.commit()
	bc.StoreBlockToStorage(bc.tailBlock)

	addr, err := AddressParse(MockDynasty[1])
	assert.Nil(t, err)
	block0, err := NewBlock(bc.ChainID(), addr, bc.tailBlock)
	assert.Nil(t, err)
	consensusState, err := bc.tailBlock.NextConsensusState(BlockInterval)
	assert.Nil(t, err)
	block0.LoadConsensusState(consensusState)
	assert.Nil(t, block0.Seal())
	assert.Nil(t, pool.Push(block0))

	addr, err = AddressParse(MockDynasty[2])
	assert.Nil(t, err)
	block1, err := NewBlock(bc.ChainID(), addr, block0)
	assert.Nil(t, err)
	consensusState, err = block0.NextConsensusState(BlockInterval)
	assert.Nil(t, err)
	block1.LoadConsensusState(consensusState)
	block1.Seal()
	assert.Nil(t, pool.Push(block1))

	addr, err = AddressParse(MockDynasty[3])
	assert.Nil(t, err)
	block2, err := NewBlock(bc.ChainID(), addr, block1)
	assert.Nil(t, err)
	consensusState, err = block1.NextConsensusState(BlockInterval)
	assert.Nil(t, err)
	block2.LoadConsensusState(consensusState)
	block2.Seal()
	assert.Nil(t, pool.Push(block2))

	addr, err = AddressParse(MockDynasty[4])
	assert.Nil(t, err)
	block3, err := NewBlock(bc.ChainID(), addr, block2)
	assert.Nil(t, err)
	consensusState, err = block2.NextConsensusState(BlockInterval)
	assert.Nil(t, err)
	block3.LoadConsensusState(consensusState)
	block3.Seal()
	assert.Nil(t, pool.Push(block3))

	addr, err = AddressParse(MockDynasty[5])
	assert.Nil(t, err)
	block4, err := NewBlock(bc.ChainID(), addr, block3)
	assert.Nil(t, err)
	consensusState, err = block3.NextConsensusState(BlockInterval)
	assert.Nil(t, err)
	block4.LoadConsensusState(consensusState)
	block4.Seal()
	assert.Nil(t, pool.Push(block4))

	neb = testNeb(t)
	bc = neb.chain
	pool = bc.bkPool

	bc.tailBlock.begin()
	baseGas, _ = util.NewUint128FromInt(2000000)
	balance, err = TransactionGasPrice.Mul(baseGas)
	assert.Nil(t, err)
	acc, err = bc.tailBlock.accState.GetOrCreateUserAccount(from.Bytes())
	assert.Nil(t, err)
	acc.AddBalance(balance)
	bc.tailBlock.header.stateRoot, err = bc.tailBlock.accState.RootHash()
	assert.Nil(t, err)
	bc.tailBlock.commit()
	bc.StoreBlockToStorage(bc.tailBlock)

	err = pool.Push(block0)
	assert.Nil(t, err)
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

	addr, err = AddressParse(MockDynasty[0])
	assert.Nil(t, err)
	block5, err := NewBlock(bc.ChainID(), addr, block4)
	assert.Nil(t, err)
	consensusState, err = block4.NextConsensusState(BlockInterval)
	assert.Nil(t, err)
	block5.LoadConsensusState(consensusState)
	assert.Nil(t, block5.Seal())
	block5.header.hash[0]++
	assert.Equal(t, pool.Push(block5), ErrInvalidBlockHash)

	addr, err = AddressParse(MockDynasty[3])
	assert.Nil(t, err)
	block41, err := NewBlock(bc.ChainID(), addr, block3)
	assert.Nil(t, err)
	consensusState, err = block3.NextConsensusState(BlockInterval)
	assert.Nil(t, err)
	block41.LoadConsensusState(consensusState)
	assert.Nil(t, block41.Seal())
	assert.Equal(t, pool.Push(block41), ErrDoubleBlockMinted)
}

func TestHandleBlock(t *testing.T) {
	neb := testNeb(t)
	bc := neb.chain
	from := mockAddress()
	ks := keystore.DefaultKS
	key, err := ks.GetUnlocked(from.String())
	signature, err := crypto.NewSignature(keystore.SECP256K1)
	assert.Nil(t, err)
	signature.InitSign(key.(keystore.PrivateKey))

	block, err := bc.NewBlock(from)
	assert.Nil(t, err)
	consensusState, err := bc.tailBlock.NextConsensusState(BlockInterval)
	assert.Nil(t, err)
	block.LoadConsensusState(consensusState)
	block.Seal()
	block.Sign(signature)
	pbMsg, err := block.ToProto()
	assert.Nil(t, err)
	data, err := proto.Marshal(pbMsg)
	assert.Nil(t, err)
	msg := net.NewBaseMessage(MessageTypeNewTx, "from", data)
	bc.bkPool.handleReceivedBlock(msg)
	assert.Nil(t, bc.GetBlock(block.Hash()))

	block, err = bc.NewBlock(from)
	assert.Nil(t, err)
	consensusState, err = block.NextConsensusState(time.Now().Unix() - AcceptedNetWorkDelay - 1)
	assert.Nil(t, err)
	block.LoadConsensusState(consensusState)
	block.Seal()
	block.Sign(signature)
	pbMsg, err = block.ToProto()
	assert.Nil(t, err)
	data, err = proto.Marshal(pbMsg)
	msg = net.NewBaseMessage(MessageTypeNewBlock, "from", data)
	bc.bkPool.handleReceivedBlock(msg)
	assert.Nil(t, bc.GetBlock(block.Hash()))

	block, err = bc.NewBlock(from)
	assert.Nil(t, err)
	consensusState, err = block.NextConsensusState(0)
	assert.Nil(t, err)
	block.LoadConsensusState(consensusState)
	block.Seal()
	block.Sign(signature)
	pbMsg, err = block.ToProto()
	assert.Nil(t, err)
	data, err = proto.Marshal(pbMsg)
	msg = net.NewBaseMessage(MessageTypeNewBlock, "from", data)
	bc.bkPool.handleReceivedBlock(msg)
	assert.Nil(t, bc.GetBlock(block.Hash()))

	block, err = bc.NewBlock(from)
	assert.Nil(t, err)
	consensusState, err = block.NextConsensusState(0)
	assert.Nil(t, err)
	block.LoadConsensusState(consensusState)
	block.Seal()
	block.Sign(signature)
	pbMsg, err = block.ToProto()
	assert.Nil(t, err)
	data, err = proto.Marshal(pbMsg)
	msg = net.NewBaseMessage(MessageTypeBlockDownloadResponse, "from", data)
	bc.bkPool.handleReceivedBlock(msg)
	assert.NotNil(t, bc.GetBlock(block.Hash()))
}

func TestHandleDownloadedBlock(t *testing.T) {
	received = []byte{}

	neb := testNeb(t)
	bc := neb.chain
	from := mockAddress()
	ks := keystore.DefaultKS
	key, err := ks.GetUnlocked(from.String())
	signature, err := crypto.NewSignature(keystore.SECP256K1)
	assert.Nil(t, err)
	signature.InitSign(key.(keystore.PrivateKey))

	block1, err := bc.NewBlock(from)
	assert.Nil(t, err)
	block1.Seal()
	block1.Sign(signature)
	bc.SetTailBlock(block1)

	block2, err := bc.NewBlock(from)
	assert.Nil(t, err)
	block2.Seal()
	block2.Sign(signature)
	bc.SetTailBlock(block2)
	bc.StoreBlockToStorage(block2)

	downloadBlock := new(corepb.DownloadBlock)
	downloadBlock.Hash = block1.Hash()
	downloadBlock.Sign = block1.Signature()
	data, err := proto.Marshal(downloadBlock)
	assert.Nil(t, err)
	msg := net.NewBaseMessage(MessageTypeNewBlock, "from", data)
	bc.bkPool.handleParentDownloadRequest(msg)
	assert.Equal(t, received, []byte{})

	downloadBlock = new(corepb.DownloadBlock)
	downloadBlock.Hash = bc.genesisBlock.Hash()
	downloadBlock.Sign = bc.genesisBlock.Signature()
	data, err = proto.Marshal(downloadBlock)
	assert.Nil(t, err)
	msg = net.NewBaseMessage(MessageTypeParentBlockDownloadRequest, "from", data)
	bc.bkPool.handleParentDownloadRequest(msg)
	assert.Equal(t, received, []byte{})

	downloadBlock = new(corepb.DownloadBlock)
	downloadBlock.Hash = block1.Hash()
	downloadBlock.Sign = block1.Signature()
	data, err = proto.Marshal(downloadBlock)
	assert.Nil(t, err)
	msg = net.NewBaseMessage(MessageTypeParentBlockDownloadRequest, "from", data)
	bc.bkPool.handleParentDownloadRequest(msg)
	assert.Equal(t, received, []byte{})
	bc.StoreBlockToStorage(block1)

	downloadBlock = new(corepb.DownloadBlock)
	downloadBlock.Hash = block1.Hash()
	downloadBlock.Sign = block2.Signature()
	data, err = proto.Marshal(downloadBlock)
	assert.Nil(t, err)
	msg = net.NewBaseMessage(MessageTypeParentBlockDownloadRequest, "from", data)
	bc.bkPool.handleParentDownloadRequest(msg)
	assert.Equal(t, received, []byte{})

	downloadBlock = new(corepb.DownloadBlock)
	downloadBlock.Hash = block1.Hash()
	downloadBlock.Sign = block1.Signature()
	data, err = proto.Marshal(downloadBlock)
	assert.Nil(t, err)
	msg = net.NewBaseMessage(MessageTypeParentBlockDownloadRequest, "from", data)
	bc.bkPool.handleParentDownloadRequest(msg)
	pbGenesis, err := bc.genesisBlock.ToProto()
	assert.Nil(t, err)
	data, err = proto.Marshal(pbGenesis)
	assert.Nil(t, err)
	assert.Equal(t, received, data)
}
