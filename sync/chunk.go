// Copyright (C) 2018 go-nebulas authors
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

package sync

import (
	"bytes"
	"errors"

	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/core"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/sync/pb"

	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

var (
	ErrTooSmallGapToSync        = errors.New("the gap between syncpoint and current tail is smaller than a dynasty interval, ignore the sync task")
	ErrCannotFindBlockByHeight  = errors.New("cannot find the block at given height")
	ErrCannotFindBlockByHash    = errors.New("cannot find the block with the given hash")
	ErrWrongChunkHeaderRootHash = errors.New("wrong chunk header root hash")
	ErrWrongChunkDataRootHash   = errors.New("wrong chunk data root hash")
)

type Chunk struct {
	blockChain *core.BlockChain
	chunksTrie *trie.BatchTrie
}

func NewChunk(blockChain *core.BlockChain) *Chunk {
	return &Chunk{
		blockChain: blockChain,
		chunksTrie: nil,
	}
}

func (c *Chunk) generateChunkHeaders(syncpoint *core.Block) (*syncpb.ChunkHeaders, error) {
	if err := c.blockChain.CheckBlockOnCanonicalChain(syncpoint); err != nil {
		return nil, err
	}
	tail := c.blockChain.TailBlock()
	if tail.Timestamp()-syncpoint.Timestamp() < core.DynastyInterval {
		logging.VLog().WithFields(logrus.Fields{
			"err": ErrTooSmallGapToSync,
		}).Warn("Failed to generate sync blocks meta info")
		return nil, ErrTooSmallGapToSync
	}

	chunkHeaders := []*syncpb.ChunkHeader{}
	stor, err := storage.NewMemoryStorage()
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to create memory storage")
		return nil, err
	}
	chunksTrie, err := trie.NewBatchTrie(nil, stor)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to create merkle tree")
		return nil, err
	}

	startChunk := (syncpoint.Height() - 1) / core.ChunkSize
	endChunk := (tail.Height() - 1) / core.ChunkSize
	curChunk := startChunk
	for curChunk < endChunk && curChunk-startChunk < MaxChunkPerSyncRequest {
		headers := [][]byte{}
		blocksTrie, err := trie.NewBatchTrie(nil, stor)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err": err,
			}).Error("Failed to create merkle tree")
			return nil, err
		}

		startHeight := curChunk*core.ChunkSize + 2
		endHeight := (curChunk+1)*core.ChunkSize + 2
		curHeight := startHeight
		for curHeight < endHeight {
			block := c.blockChain.GetBlockByHeight(curHeight)
			if block == nil {
				logging.VLog().WithFields(logrus.Fields{
					"height": curHeight + 1,
				}).Error("Failed to find the block on canonical chain.")
				return nil, ErrCannotFindBlockByHeight
			}
			headers = append(headers, block.Hash())
			blocksTrie.Put(block.Hash(), block.Hash())
			curHeight++
		}
		chunkHeaders = append(chunkHeaders, &syncpb.ChunkHeader{Headers: headers, Root: blocksTrie.RootHash()})
		chunksTrie.Put(blocksTrie.RootHash(), blocksTrie.RootHash())

		curChunk++
	}

	logging.VLog().WithFields(logrus.Fields{
		"syncpoint": syncpoint,
		"start":     startChunk,
		"end":       endChunk,
		"limit":     MaxChunkPerSyncRequest,
		"synced":    len(chunkHeaders),
	}).Debug("Succeed to generate chunks meta info.")
	return &syncpb.ChunkHeaders{ChunkHeaders: chunkHeaders, Root: chunksTrie.RootHash()}, nil
}

func VerifyChunkHeaders(chunkHeaders *syncpb.ChunkHeaders) (bool, error) {
	stor, err := storage.NewMemoryStorage()
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to create memory storage")
		return false, err
	}

	chunksTrie, err := trie.NewBatchTrie(nil, stor)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to create merkle tree")
		return false, err
	}

	for _, chunkHeader := range chunkHeaders.ChunkHeaders {
		blocksTrie, err := trie.NewBatchTrie(nil, stor)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err": err,
			}).Error("Failed to create merkle tree")
			return false, err
		}

		for _, blockHash := range chunkHeader.Headers {
			blocksTrie.Put(blockHash, blockHash)
		}

		chunksTrie.Put(blocksTrie.RootHash(), blocksTrie.RootHash())
	}

	return bytes.Compare(chunksTrie.RootHash(), chunkHeaders.Root) == 0, nil
}

func (c *Chunk) generateChunkData(chunkHeader *syncpb.ChunkHeader) (*syncpb.ChunkData, error) {
	stor, err := storage.NewMemoryStorage()
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to create memory storage")
		return nil, err
	}

	blocksTrie, err := trie.NewBatchTrie(nil, stor)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to create merkle tree")
		return nil, err
	}

	blocks := []*corepb.Block{}
	for k, v := range chunkHeader.Headers {
		block := c.blockChain.GetBlock(v)
		if block == nil {
			logging.VLog().WithFields(logrus.Fields{
				"index": k,
				"hash":  byteutils.Hex(v),
				"err":   ErrCannotFindBlockByHash,
			}).Error("Failed to find the block on canonical chain.")
			return nil, ErrCannotFindBlockByHash
		}
		pbBlock, err := block.ToProto()
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"block": block,
				"err":   err,
			}).Error("Failed to serialize block.")
			return nil, err
		}
		blocks = append(blocks, pbBlock.(*corepb.Block))
		blocksTrie.Put(block.Hash(), block.Hash())
	}

	logging.VLog().WithFields(logrus.Fields{
		"size": len(blocks),
	}).Debug("Succeed to generate chunk.")

	if bytes.Compare(blocksTrie.RootHash(), chunkHeader.Root) != 0 {
		logging.VLog().WithFields(logrus.Fields{
			"size":                len(blocks),
			"localChunkRootHash":  byteutils.Hex(blocksTrie.RootHash()),
			"chunkHeader":         chunkHeader,
			"chunkHeaderRootHash": byteutils.Hex(chunkHeader.Root),
		}).Debug("Wrong chunk header root hash.")
		return nil, ErrWrongChunkHeaderRootHash
	}

	return &syncpb.ChunkData{Blocks: blocks, Root: blocksTrie.RootHash()}, nil
}

func VerifyChunkData(chunkHeader *syncpb.ChunkHeader, chunkData *syncpb.ChunkData) (bool, error) {
	stor, err := storage.NewMemoryStorage()
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to create memory storage")
		return false, err
	}

	blocksTrie, err := trie.NewBatchTrie(nil, stor)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to create merkle tree")
		return false, err
	}

	for _, block := range chunkData.Blocks {
		blocksTrie.Put(block.Header.Hash, block.Header.Hash)
	}

	if bytes.Compare(blocksTrie.RootHash(), chunkHeader.Root) != 0 {
		logging.VLog().WithFields(logrus.Fields{
			"size":                len(chunkData.Blocks),
			"localChunkRootHash":  byteutils.Hex(blocksTrie.RootHash()),
			"chunkHeader":         chunkHeader,
			"chunkHeaderRootHash": byteutils.Hex(chunkHeader.Root),
		}).Debug("Wrong chunk header root hash.")
		return false, ErrWrongChunkDataRootHash
	}

	return true, nil
}

func (c *Chunk) processChunkData(chunk *syncpb.ChunkData) error {
	for k, v := range chunk.Blocks {
		block := new(core.Block)
		if err := block.FromProto(v); err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"index": k,
				"hash":  v.Header.Hash,
				"err":   err,
			}).Error("Failed to recover a block from proto data.")
			return err
		}
		if err := c.blockChain.BlockPool().Push(block); err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"index": k,
				"hash":  v.Header.Hash,
				"err":   err,
			}).Error("Failed to recover a block from proto data.")
			return err
		}
	}

	logging.VLog().WithFields(logrus.Fields{
		"size": len(chunk.Blocks),
	}).Debug("Succeed to process chunk.")
	return nil
}
