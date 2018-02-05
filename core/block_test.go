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
	"reflect"
	"testing"
	"time"

	pb "github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/crypto"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/keystore/secp256k1"
	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/stretchr/testify/assert"
)

type mockNeb struct {
	genesis *corepb.Genesis
	config  *nebletpb.Config
	storage storage.Storage
	emitter *EventEmitter
}

func (n *mockNeb) Genesis() *corepb.Genesis {
	return n.genesis
}

func (n *mockNeb) Config() *nebletpb.Config {
	return n.config
}

func (n *mockNeb) Storage() storage.Storage {
	return n.storage
}

func (n *mockNeb) EventEmitter() *EventEmitter {
	return n.emitter
}

func (n *mockNeb) StartActiveSync() {}

func testNeb() *mockNeb {
	storage, _ := storage.NewMemoryStorage()
	eventEmitter := NewEventEmitter(1024)
	neb := &mockNeb{
		genesis: MockGenesisConf(),
		config:  &nebletpb.Config{Chain: &nebletpb.ChainConfig{ChainId: MockGenesisConf().Meta.ChainId}},
		storage: storage,
		emitter: eventEmitter,
	}
	return neb
}

func TestBlock(t *testing.T) {
	type fields struct {
		header       *BlockHeader
		miner        *Address
		height       uint64
		transactions Transactions
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			"full struct",
			fields{
				&BlockHeader{
					hash:       []byte("124546"),
					parentHash: []byte("344543"),
					stateRoot:  []byte("43656"),
					txsRoot:    []byte("43656"),
					eventsRoot: []byte("43656"),
					dposContext: &corepb.DposContext{
						DynastyRoot:     []byte("43656"),
						NextDynastyRoot: []byte("43656"),
						DelegateRoot:    []byte("43656"),
					},
					nonce:     3546456,
					coinbase:  &Address{[]byte("hello")},
					timestamp: time.Now().Unix(),
					chainID:   100,
				},
				&Address{[]byte("hello")},
				1,
				Transactions{
					&Transaction{
						[]byte("123452"),
						&Address{[]byte("1335")},
						&Address{[]byte("1245")},
						util.NewUint128(),
						456,
						1516464510,
						&corepb.Data{Type: TxPayloadBinaryType, Payload: []byte("hello")},
						1,
						util.NewUint128(),
						util.NewUint128(),
						uint8(keystore.SECP256K1),
						nil,
					},
					&Transaction{
						[]byte("123455"),
						&Address{[]byte("1235")},
						&Address{[]byte("1425")},
						util.NewUint128(),
						446,
						1516464511,
						&corepb.Data{Type: TxPayloadBinaryType, Payload: []byte("hllo")},
						2,
						util.NewUint128(),
						util.NewUint128(),
						uint8(keystore.SECP256K1),
						nil,
					},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Block{
				header:       tt.fields.header,
				miner:        tt.fields.miner,
				height:       tt.fields.height,
				transactions: tt.fields.transactions,
			}
			proto, _ := b.ToProto()
			ir, _ := pb.Marshal(proto)
			nb := new(Block)
			pb.Unmarshal(ir, proto)
			nb.FromProto(proto)
			b.header.timestamp = nb.header.timestamp
			if !reflect.DeepEqual(*b.header, *nb.header) {
				t.Errorf("Transaction.Serialize() = %v, want %v", *b.header, *nb.header)
			}
			if !reflect.DeepEqual(*b.transactions[0], *nb.transactions[0]) {
				t.Errorf("Transaction.Serialize() = %v, want %v", *b.transactions[0], *nb.transactions[0])
			}
			if !reflect.DeepEqual(*b.transactions[1], *nb.transactions[1]) {
				t.Errorf("Transaction.Serialize() = %v, want %v", *b.transactions[1], *nb.transactions[1])
			}
		})
	}
}

func TestBlock_LinkParentBlock(t *testing.T) {
	bc, _ := NewBlockChain(testNeb())
	genesis := bc.genesisBlock
	assert.Equal(t, genesis.Height(), uint64(1))
	block1 := &Block{
		header: &BlockHeader{
			hash:       []byte("124546"),
			parentHash: GenesisHash,
			stateRoot:  []byte("43656"),
			txsRoot:    []byte("43656"),
			eventsRoot: []byte("43656"),
			dposContext: &corepb.DposContext{
				DynastyRoot:     []byte("43656"),
				NextDynastyRoot: []byte("43656"),
				DelegateRoot:    []byte("43656"),
			},
			nonce:     3546456,
			coinbase:  &Address{[]byte("hello")},
			timestamp: BlockInterval,
			chainID:   100,
		},
		transactions: []*Transaction{},
	}
	assert.Equal(t, block1.Height(), uint64(0))
	assert.Equal(t, block1.LinkParentBlock(bc, genesis), nil)
	assert.Equal(t, block1.Height(), uint64(2))
	assert.Equal(t, block1.ParentHash(), genesis.Hash())
	block2 := &Block{
		header: &BlockHeader{
			hash:       []byte("124546"),
			parentHash: []byte("344543"),
			stateRoot:  []byte("43656"),
			txsRoot:    []byte("43656"),
			eventsRoot: []byte("43656"),
			dposContext: &corepb.DposContext{
				DynastyRoot:     []byte("43656"),
				NextDynastyRoot: []byte("43656"),
				DelegateRoot:    []byte("43656"),
			},
			nonce:     3546456,
			coinbase:  &Address{[]byte("hello")},
			timestamp: BlockInterval * 2,
			chainID:   100,
		},
		transactions: []*Transaction{},
	}
	assert.Equal(t, block2.LinkParentBlock(bc, genesis), ErrLinkToWrongParentBlock)
	assert.Equal(t, block2.Height(), uint64(0))
}

func TestBlock_CollectTransactions(t *testing.T) {
	bc, _ := NewBlockChain(testNeb())
	var c MockConsensus
	bc.SetConsensusHandler(c)

	tail := bc.tailBlock

	ks := keystore.DefaultKS
	priv := secp256k1.GeneratePrivateKey()
	pubdata, _ := priv.PublicKey().Encoded()
	from, _ := NewAddressFromPublicKey(pubdata)
	ks.SetKey(from.String(), priv, []byte("passphrase"))
	ks.Unlock(from.String(), []byte("passphrase"), time.Second*60*60*24*365)

	key, _ := ks.GetUnlocked(from.String())
	signature, _ := crypto.NewSignature(keystore.SECP256K1)
	signature.InitSign(key.(keystore.PrivateKey))

	priv1 := secp256k1.GeneratePrivateKey()
	pubdata1, _ := priv1.PublicKey().Encoded()
	to, _ := NewAddressFromPublicKey(pubdata1)
	priv2 := secp256k1.GeneratePrivateKey()
	pubdata2, _ := priv2.PublicKey().Encoded()
	coinbase, _ := NewAddressFromPublicKey(pubdata2)

	block0, _ := NewBlock(bc.ChainID(), from, tail)
	block0.header.timestamp = BlockInterval
	block0.SetMiner(from)
	block0.Seal()
	//bc.BlockPool().push(block0)
	bc.SetTailBlock(block0)

	block, _ := NewBlock(bc.ChainID(), coinbase, block0)
	block.header.timestamp = BlockInterval * 2

	tx1 := NewTransaction(bc.ChainID(), from, to, util.NewUint128FromInt(1), 1, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx1.Sign(signature)
	tx2 := NewTransaction(bc.ChainID(), from, to, util.NewUint128FromInt(1), 2, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx2.Sign(signature)
	tx3 := NewTransaction(bc.ChainID(), from, to, util.NewUint128FromInt(1), 0, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx3.Sign(signature)
	tx4 := NewTransaction(bc.ChainID(), from, to, util.NewUint128FromInt(1), 4, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx4.Sign(signature)
	tx5 := NewTransaction(bc.ChainID(), from, to, util.NewUint128FromInt(1), 3, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx5.Sign(signature)
	tx6 := NewTransaction(bc.ChainID()+1, from, to, util.NewUint128FromInt(1), 1, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx6.Sign(signature)

	assert.Nil(t, bc.txPool.Push(tx1))
	assert.Nil(t, bc.txPool.Push(tx2))
	assert.Nil(t, bc.txPool.Push(tx3))
	assert.Nil(t, bc.txPool.Push(tx4))
	assert.Nil(t, bc.txPool.Push(tx5))
	assert.NotNil(t, bc.txPool.Push(tx6), ErrInvalidChainID)

	assert.Equal(t, len(block.transactions), 0)
	assert.Equal(t, bc.txPool.cache.Len(), 5)
	block.CollectTransactions(time.Now().Unix() + 2)
	assert.Equal(t, len(block.transactions), 4)
	assert.Equal(t, block.txPool.cache.Len(), 0)

	assert.Equal(t, block.Sealed(), false)
	balance, err := block.GetBalance(block.header.coinbase.address)
	assert.Nil(t, err)
	assert.Equal(t, balance.Cmp(util.NewUint128().Int), 1)
	block.SetMiner(coinbase)
	block.Seal()
	assert.Equal(t, block.Sealed(), true)
	assert.Equal(t, block.transactions[0], tx1)
	assert.Equal(t, block.transactions[1], tx2)
	stateRoot, err := block.accState.RootHash()
	assert.Nil(t, err)
	assert.Equal(t, block.StateRoot().Equals(stateRoot), true)
	assert.Equal(t, block.TxsRoot().Equals(block.txsTrie.RootHash()), true)
	balance, err = block.GetBalance(block.header.coinbase.address)
	assert.Nil(t, err)
	// balance > BlockReward (BlockReward + gas)
	//gas, _ := bc.EstimateGas(tx1)
	logging.CLog().Info(balance.String())
	logging.CLog().Info(BlockReward.String())
	assert.NotEqual(t, balance.Cmp(BlockReward.Int), 0)
	// mock net message
	block, _ = mockBlockFromNetwork(block)
	assert.Equal(t, block.LinkParentBlock(bc, bc.tailBlock), nil)
	block.SetMiner(coinbase)
	assert.Nil(t, block.VerifyExecution(bc.tailBlock, bc.ConsensusHandler()))
}

func TestBlock_DposCandidates(t *testing.T) {
	bc, _ := NewBlockChain(testNeb())
	var c MockConsensus
	bc.SetConsensusHandler(c)

	tail := bc.tailBlock

	ks := keystore.DefaultKS
	priv := secp256k1.GeneratePrivateKey()
	pubdata, _ := priv.PublicKey().Encoded()
	from, _ := NewAddressFromPublicKey(pubdata)
	ks.SetKey(from.String(), priv, []byte("passphrase"))
	ks.Unlock(from.String(), []byte("passphrase"), time.Second*60*60*24*365)

	key, _ := ks.GetUnlocked(from.String())
	signature, _ := crypto.NewSignature(keystore.SECP256K1)
	signature.InitSign(key.(keystore.PrivateKey))

	priv1 := secp256k1.GeneratePrivateKey()
	pubdata1, _ := priv1.PublicKey().Encoded()
	to, _ := NewAddressFromPublicKey(pubdata1)
	priv2 := secp256k1.GeneratePrivateKey()
	pubdata2, _ := priv2.PublicKey().Encoded()
	coinbase, _ := NewAddressFromPublicKey(pubdata2)

	block0, _ := NewBlock(bc.ChainID(), from, tail)
	block0.header.timestamp = BlockInterval
	block0.SetMiner(from)
	block0.Seal()
	assert.Nil(t, bc.storeBlockToStorage(block0))
	assert.Nil(t, bc.SetTailBlock(block0))

	block, _ := NewBlock(bc.ChainID(), coinbase, block0)
	block.header.timestamp = BlockInterval * 2
	bytes, _ := NewCandidatePayload(LoginAction).ToBytes()
	tx := NewTransaction(bc.ChainID(), from, to, util.NewUint128FromInt(1), 1, TxPayloadCandidateType, bytes, TransactionGasPrice, util.NewUint128FromInt(200000))
	tx.Sign(signature)
	bc.txPool.Push(tx)
	payload := NewDelegatePayload(DelegateAction, from.String())
	bytes, _ = payload.ToBytes()
	tx = NewTransaction(bc.ChainID(), from, to, util.NewUint128FromInt(1), 2, TxPayloadDelegateType, bytes, TransactionGasPrice, util.NewUint128FromInt(200000))
	tx.Sign(signature)
	bc.txPool.Push(tx)
	assert.Equal(t, len(block.transactions), 0)
	assert.Equal(t, bc.txPool.cache.Len(), 2)
	block.CollectTransactions(time.Now().Unix() + 2)
	assert.Equal(t, len(block.transactions), 2)
	assert.Equal(t, block.txPool.cache.Len(), 0)
	block.SetMiner(coinbase)
	assert.Equal(t, block.Seal(), nil)
	block, _ = mockBlockFromNetwork(block)
	assert.Equal(t, block.LinkParentBlock(bc, bc.tailBlock), nil)
	block.SetMiner(coinbase)
	assert.Nil(t, block.VerifyExecution(bc.tailBlock, bc.ConsensusHandler()))
	bytes, _ = block.dposContext.candidateTrie.Get(from.Bytes())
	assert.Equal(t, bytes, from.Bytes())
	bytes, _ = block.dposContext.voteTrie.Get(from.Bytes())
	assert.Equal(t, bytes, from.Bytes())
	bytes, _ = block.dposContext.delegateTrie.Get(append(from.Bytes(), from.Bytes()...))
	assert.Equal(t, bytes, from.Bytes())
	assert.Nil(t, bc.storeBlockToStorage(block))
	assert.Nil(t, bc.SetTailBlock(block))

	block, _ = NewBlock(bc.ChainID(), coinbase, block)
	block.header.timestamp = BlockInterval * 3
	payload = NewDelegatePayload(UnDelegateAction, from.String())
	bytes, _ = payload.ToBytes()
	tx = NewTransaction(bc.ChainID(), from, to, util.NewUint128FromInt(1), 3, TxPayloadDelegateType, bytes, TransactionGasPrice, util.NewUint128FromInt(200000))
	tx.Sign(signature)
	bc.txPool.Push(tx)
	assert.Equal(t, len(block.transactions), 0)
	assert.Equal(t, bc.txPool.cache.Len(), 1)
	block.CollectTransactions(time.Now().Unix() + 2)
	assert.Equal(t, len(block.transactions), 1)
	assert.Equal(t, block.txPool.cache.Len(), 0)
	block.SetMiner(coinbase)
	assert.Equal(t, block.Seal(), nil)
	block, _ = mockBlockFromNetwork(block)
	assert.Equal(t, block.LinkParentBlock(bc, bc.tailBlock), nil)
	block.SetMiner(coinbase)
	assert.Nil(t, block.VerifyExecution(bc.tailBlock, bc.ConsensusHandler()))
	_, err := block.dposContext.candidateTrie.Get(from.Bytes())
	assert.Equal(t, err, nil)
	_, err = block.dposContext.voteTrie.Get(from.Bytes())
	assert.Equal(t, err, storage.ErrKeyNotFound)
	_, err = block.dposContext.delegateTrie.Iterator(from.Bytes())
	assert.Equal(t, err, storage.ErrKeyNotFound)
	assert.Nil(t, bc.storeBlockToStorage(block))
	assert.Nil(t, bc.SetTailBlock(block))

	block, _ = NewBlock(bc.ChainID(), coinbase, block)
	block.header.timestamp = BlockInterval * 4
	payload = NewDelegatePayload(DelegateAction, from.String())
	bytes, _ = payload.ToBytes()
	tx = NewTransaction(bc.ChainID(), from, to, util.NewUint128FromInt(1), 4, TxPayloadDelegateType, bytes, TransactionGasPrice, util.NewUint128FromInt(200000))
	tx.Sign(signature)
	bc.txPool.Push(tx)
	bytes, _ = NewCandidatePayload(LogoutAction).ToBytes()
	tx = NewTransaction(bc.ChainID(), from, to, util.NewUint128FromInt(1), 5, TxPayloadCandidateType, bytes, TransactionGasPrice, util.NewUint128FromInt(200000))
	tx.Sign(signature)
	bc.txPool.Push(tx)
	assert.Equal(t, len(block.transactions), 0)
	assert.Equal(t, bc.txPool.cache.Len(), 2)
	block.CollectTransactions(time.Now().Unix() + 2)
	assert.Equal(t, len(block.transactions), 2)
	assert.Equal(t, block.txPool.cache.Len(), 0)
	block.SetMiner(coinbase)
	assert.Equal(t, block.Seal(), nil)
	block, _ = mockBlockFromNetwork(block)
	assert.Equal(t, block.LinkParentBlock(bc, bc.tailBlock), nil)
	block.SetMiner(coinbase)
	assert.Nil(t, block.VerifyExecution(bc.tailBlock, bc.ConsensusHandler()))
	_, err = block.dposContext.candidateTrie.Get(from.Bytes())
	assert.Equal(t, err, storage.ErrKeyNotFound)
	_, err = block.dposContext.voteTrie.Get(from.Bytes())
	assert.Equal(t, err, storage.ErrKeyNotFound)
	_, err = block.dposContext.delegateTrie.Iterator(from.Bytes())
	assert.Equal(t, err, storage.ErrKeyNotFound)
	assert.Nil(t, bc.storeBlockToStorage(block))
	assert.Nil(t, bc.SetTailBlock(block))
}

func TestBlock_fetchEvents(t *testing.T) {
	bc, _ := NewBlockChain(testNeb())
	tail := bc.tailBlock
	events := []*Event{
		&Event{Topic: "chain.block", Data: "hello"},
		&Event{Topic: "chain.tx", Data: "hello"},
		&Event{Topic: "chain.block", Data: "hello"},
		&Event{Topic: "chain.block", Data: "hello"},
	}
	tx := &Transaction{hash: []byte("tx")}
	for _, event := range events {
		assert.Nil(t, tail.recordEvent(tx.Hash(), event))
	}
	es, err := tail.FetchEvents(tx.Hash())
	assert.Nil(t, err)
	for idx, event := range es {
		assert.Equal(t, events[idx], event)
	}
}

func TestSerializeTxByHash(t *testing.T) {
	bc, err := NewBlockChain(testNeb())
	assert.Nil(t, err)
	block := bc.tailBlock
	tx := NewTransaction(bc.ChainID(), mockAddress(), mockAddress(), util.NewUint128(), 1, TxPayloadBinaryType, []byte(""), TransactionGasPrice, TransactionMaxGas)
	hash, err := HashTransaction(tx)
	assert.Nil(t, err)
	tx.hash = hash
	block.acceptTransaction(tx)
	msg, err := block.SerializeTxByHash(hash)
	assert.Nil(t, err)
	bytes, err := pb.Marshal(msg)
	assert.Nil(t, err)
	msg2, err := tx.ToProto()
	assert.Nil(t, err)
	bytes2, err := pb.Marshal(msg2)
	assert.Equal(t, bytes, bytes2)
}

func TestBlockSign(t *testing.T) {
	bc, err := NewBlockChain(testNeb())
	assert.Nil(t, err)
	block := bc.tailBlock
	ks := keystore.DefaultKS
	signature, _ := crypto.NewSignature(keystore.SECP256K1)
	signer := mockAddress()
	key, _ := ks.GetUnlocked(signer.String())
	signature.InitSign(key.(keystore.PrivateKey))
	assert.Nil(t, block.Sign(signature))
	assert.Equal(t, block.Alg(), uint8(keystore.SECP256K1))
	assert.Equal(t, block.Signature(), block.header.sign)
}

func TestGivebackInvalidTx(t *testing.T) {
	bc, err := NewBlockChain(testNeb())
	assert.Nil(t, err)
	from := mockAddress()
	ks := keystore.DefaultKS
	tx := NewTransaction(bc.ChainID(), from, from, util.NewUint128(), 2, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	key, err := ks.GetUnlocked(from.String())
	assert.Nil(t, err)
	signature, err := crypto.NewSignature(keystore.SECP256K1)
	assert.Nil(t, err)
	signature.InitSign(key.(keystore.PrivateKey))
	tx.Sign(signature)
	assert.Nil(t, bc.txPool.Push(tx))
	assert.Equal(t, len(bc.txPool.all), 1)
	block, err := bc.NewBlock(from)
	assert.Nil(t, err)
	block.CollectTransactions(time.Now().Unix() + 2)
	timer := time.NewTimer(time.Second).C
	<-timer
	assert.Equal(t, len(bc.txPool.all), 1)
}

func TestRecordEvent(t *testing.T) {
	bc, err := NewBlockChain(testNeb())
	assert.Nil(t, err)
	txHash := []byte("hello")
	assert.Nil(t, bc.tailBlock.RecordEvent(txHash, TopicSendTransaction, "world"))
	events, err := bc.tailBlock.FetchEvents(txHash)
	assert.Nil(t, err)
	assert.Equal(t, len(events), 1)
	assert.Equal(t, events[0].Topic, TopicSendTransaction)
	assert.Equal(t, events[0].Data, "world")
}

func TestBlockVerifyIntegrity(t *testing.T) {
	var cons MockConsensus
	bc, err := NewBlockChain(testNeb())
	bc.SetConsensusHandler(cons)
	assert.Nil(t, err)
	assert.Equal(t, bc.tailBlock.VerifyIntegrity(0, nil), ErrInvalidChainID)
	bc.tailBlock.header.hash[0] = 1
	assert.Equal(t, bc.tailBlock.VerifyIntegrity(bc.ChainID(), nil), ErrInvalidBlockHash)
	ks := keystore.DefaultKS
	from := mockAddress()
	key, err := ks.GetUnlocked(from.String())
	assert.Nil(t, err)
	signature, err := crypto.NewSignature(keystore.SECP256K1)
	assert.Nil(t, err)
	signature.InitSign(key.(keystore.PrivateKey))
	block, err := bc.NewBlock(from)
	assert.Nil(t, err)
	tx1 := NewTransaction(bc.ChainID(), from, from, util.NewUint128(), 1, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx1.Sign(signature)
	tx2 := NewTransaction(bc.ChainID(), from, from, util.NewUint128(), 2, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx2.Sign(signature)
	tx2.hash[0]++
	block.transactions = append(block.transactions, tx1)
	block.transactions = append(block.transactions, tx2)
	block.miner = from
	block.Seal()
	block.Sign(signature)
	assert.NotNil(t, block.VerifyIntegrity(bc.ChainID(), bc.ConsensusHandler()))
}

func TestBlockVerifyIntegrityDup(t *testing.T) {
	var cons MockConsensus
	bc, err := NewBlockChain(testNeb())
	bc.SetConsensusHandler(cons)
	assert.Nil(t, err)
	assert.Equal(t, bc.tailBlock.VerifyIntegrity(0, nil), ErrInvalidChainID)
	bc.tailBlock.header.hash[0] = 1
	assert.Equal(t, bc.tailBlock.VerifyIntegrity(bc.ChainID(), nil), ErrInvalidBlockHash)
	ks := keystore.DefaultKS
	from := mockAddress()
	key, err := ks.GetUnlocked(from.String())
	assert.Nil(t, err)
	signature, err := crypto.NewSignature(keystore.SECP256K1)
	assert.Nil(t, err)
	signature.InitSign(key.(keystore.PrivateKey))
	block, err := bc.NewBlock(from)
	assert.Nil(t, err)
	tx1 := NewTransaction(bc.ChainID(), from, from, util.NewUint128(), 1, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx1.Sign(signature)
	block.transactions = append(block.transactions, tx1)
	block.transactions = append(block.transactions, tx1)
	block.miner = from
	block.Seal()
	block.Sign(signature)
	assert.Equal(t, block.VerifyExecution(bc.tailBlock, bc.ConsensusHandler()), ErrSmallTransactionNonce)
}

func TestBlockVerifyExecution(t *testing.T) {
	var cons MockConsensus
	bc, err := NewBlockChain(testNeb())
	bc.SetConsensusHandler(cons)
	assert.Nil(t, err)
	assert.Equal(t, bc.tailBlock.VerifyIntegrity(0, nil), ErrInvalidChainID)
	bc.tailBlock.header.hash[0] = 1
	assert.Equal(t, bc.tailBlock.VerifyIntegrity(bc.ChainID(), nil), ErrInvalidBlockHash)
	ks := keystore.DefaultKS
	from := mockAddress()
	key, err := ks.GetUnlocked(from.String())
	assert.Nil(t, err)
	signature, err := crypto.NewSignature(keystore.SECP256K1)
	assert.Nil(t, err)
	signature.InitSign(key.(keystore.PrivateKey))
	block, err := bc.NewBlock(from)
	assert.Nil(t, err)
	tx1 := NewTransaction(bc.ChainID(), from, from, util.NewUint128(), 1, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx1.Sign(signature)
	tx2 := NewTransaction(bc.ChainID(), from, from, util.NewUint128(), 3, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx2.Sign(signature)
	block.transactions = append(block.transactions, tx1)
	block.transactions = append(block.transactions, tx2)
	block.miner = from
	block.Seal()
	block.Sign(signature)
	assert.Nil(t, block.VerifyIntegrity(bc.ChainID(), bc.ConsensusHandler()))
	root1, err := block.accState.RootHash()
	assert.Nil(t, err)
	assert.Equal(t, block.VerifyExecution(bc.tailBlock, bc.ConsensusHandler()), ErrLargeTransactionNonce)
	root2, err := block.accState.RootHash()
	assert.Nil(t, err)
	assert.Equal(t, root1, root2)
}

func TestBlockVerifyState(t *testing.T) {
	var cons MockConsensus
	bc, err := NewBlockChain(testNeb())
	bc.SetConsensusHandler(cons)
	assert.Nil(t, err)
	assert.Equal(t, bc.tailBlock.VerifyIntegrity(0, nil), ErrInvalidChainID)
	bc.tailBlock.header.hash[0] = 1
	assert.Equal(t, bc.tailBlock.VerifyIntegrity(bc.ChainID(), nil), ErrInvalidBlockHash)
	ks := keystore.DefaultKS
	from := mockAddress()
	key, err := ks.GetUnlocked(from.String())
	assert.Nil(t, err)
	signature, err := crypto.NewSignature(keystore.SECP256K1)
	assert.Nil(t, err)
	signature.InitSign(key.(keystore.PrivateKey))
	block, err := bc.NewBlock(from)
	assert.Nil(t, err)
	tx1 := NewTransaction(bc.ChainID(), from, from, util.NewUint128(), 1, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx1.Sign(signature)
	tx2 := NewTransaction(bc.ChainID(), from, from, util.NewUint128(), 2, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx2.Sign(signature)
	block.transactions = append(block.transactions, tx1)
	block.transactions = append(block.transactions, tx2)
	block.miner = from
	block.Seal()
	block.Sign(signature)
	assert.Nil(t, block.VerifyIntegrity(bc.ChainID(), bc.ConsensusHandler()))
	block.header.stateRoot[0]++
	assert.NotNil(t, block.VerifyExecution(bc.tailBlock, bc.ConsensusHandler()))
}

func TestBlock_String(t *testing.T) {
	bc, err := NewBlockChain(testNeb())
	assert.Nil(t, err)
	bc.genesisBlock.miner = nil
	logging.CLog().Info(bc.genesisBlock)
	assert.NotNil(t, bc.genesisBlock.String())
}
