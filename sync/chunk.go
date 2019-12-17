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

	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/core"
	corepb "github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/storage"
	syncpb "github.com/nebulasio/go-nebulas/sync/pb"

	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/sirupsen/logrus"
)

// Chunk packs some blocks
type Chunk struct {
	blockChain *core.BlockChain
	chunksTrie *trie.Trie
}

// NewChunk return a new chunk
func NewChunk(blockChain *core.BlockChain) *Chunk {
	return &Chunk{
		blockChain: blockChain,
		chunksTrie: nil,
	}
}

func (c *Chunk) generateChunkHeaders(syncpointHash byteutils.Hash) (*syncpb.ChunkHeaders, error) {
	syncpoint := c.blockChain.GetBlockOnCanonicalChainByHash(syncpointHash)
	if syncpoint == nil {
		logging.VLog().WithFields(logrus.Fields{
			"syncpointHash": syncpointHash.Hex(),
		}).Debug("Failed to find the block on canonical chain")
		return nil, ErrCannotFindBlockByHash
	}
	tail := c.blockChain.TailBlock()
	if int(tail.Height())-int(syncpoint.Height()) <= core.ChunkSize {
		logging.VLog().WithFields(logrus.Fields{
			"err": ErrTooSmallGapToSync,
		}).Debug("Failed to generate sync blocks meta info")
		return &syncpb.ChunkHeaders{}, ErrTooSmallGapToSync
	}

	chunkHeaders := []*syncpb.ChunkHeader{}
	stor, err := storage.NewMemoryStorage()
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Debug("Failed to create memory storage")
		return nil, err
	}
	chunksTrie, err := trie.NewTrie(nil, stor, false)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Debug("Failed to create merkle tree")
		return nil, err
	}

	startChunk := (syncpoint.Height() - 1) / core.ChunkSize
	endChunk := (tail.Height() - 1) / core.ChunkSize
	curChunk := startChunk
	for curChunk < endChunk && curChunk-startChunk < MaxChunkPerSyncRequest {
		headers := [][]byte{}
		blocksTrie, err := trie.NewTrie(nil, stor, false)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err": err,
			}).Debug("Failed to create merkle tree")
			return nil, err
		}

		startHeight := curChunk*core.ChunkSize + 2
		endHeight := (curChunk+1)*core.ChunkSize + 2
		curHeight := startHeight
		for curHeight < endHeight {
			block := c.blockChain.GetBlockOnCanonicalChainByHeight(curHeight)
			if block == nil {
				logging.VLog().WithFields(logrus.Fields{
					"height": curHeight + 1,
				}).Debug("Failed to find the block on canonical chain.")
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

func verifyChunkHeaders(chunkHeaders *syncpb.ChunkHeaders) (bool, error) {
	if len(chunkHeaders.ChunkHeaders) == 0 && len(chunkHeaders.Root) == 0 {
		// fast quit.
		return true, nil
	}

	stor, err := storage.NewMemoryStorage()
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Debug("Failed to create memory storage")
		return false, err
	}

	chunksTrie, err := trie.NewTrie(nil, stor, false)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Debug("Failed to create merkle tree")
		return false, err
	}

	for _, chunkHeader := range chunkHeaders.ChunkHeaders {
		blocksTrie, err := trie.NewTrie(nil, stor, false)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err": err,
			}).Debug("Failed to create merkle tree")
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
		}).Debug("Failed to create memory storage")
		return nil, err
	}

	blocksTrie, err := trie.NewTrie(nil, stor, false)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Debug("Failed to create merkle tree")
		return nil, err
	}

	blocks := []*corepb.Block{}
	for k, v := range chunkHeader.Headers {
		block := c.blockChain.GetBlockOnCanonicalChainByHash(v)
		if block == nil {
			logging.VLog().WithFields(logrus.Fields{
				"index": k,
				"hash":  byteutils.Hex(v),
				"err":   ErrCannotFindBlockByHash,
			}).Debug("Failed to find the block on canonical chain.")
			return nil, ErrCannotFindBlockByHash
		}
		pbBlock, err := block.ToProto()
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"block": block,
				"err":   err,
			}).Debug("Failed to serialize block.")
			return nil, err
		}
		blocks = append(blocks, pbBlock.(*corepb.Block))
		blocksTrie.Put(block.Hash(), block.Hash())
	}

	if bytes.Compare(blocksTrie.RootHash(), chunkHeader.Root) != 0 {
		logging.VLog().WithFields(logrus.Fields{
			"size":                len(blocks),
			"localChunkRootHash":  byteutils.Hex(blocksTrie.RootHash()),
			"chunkHeader":         chunkHeader,
			"chunkHeaderRootHash": byteutils.Hex(chunkHeader.Root),
		}).Debug("Wrong chunk header root hash.")
		return nil, ErrWrongChunkHeaderRootHash
	}

	logging.VLog().WithFields(logrus.Fields{
		"size": len(blocks),
	}).Debug("Succeed to generate chunk.")

	return &syncpb.ChunkData{Blocks: blocks, Root: blocksTrie.RootHash()}, nil
}

func verifyChunkData(chunkHeader *syncpb.ChunkHeader, chunkData *syncpb.ChunkData) (bool, error) {
	stor, err := storage.NewMemoryStorage()
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Debug("Failed to create memory storage")
		return false, err
	}

	blocksTrie, err := trie.NewTrie(nil, stor, false)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Debug("Failed to create merkle tree")
		return false, err
	}

	if len(chunkHeader.Headers) != len(chunkData.Blocks) {
		logging.VLog().WithFields(logrus.Fields{
			"chunkData.size":   len(chunkData.Blocks),
			"chunkHeader.size": len(chunkHeader.Headers),
			"err":              ErrWrongChunkDataSize,
		}).Debug("Wrong chunk data size.")
		return false, ErrWrongChunkDataSize
	}

	for k, block := range chunkData.Blocks {
		hash := chunkHeader.Headers[k]
		calculated, err := core.HashPbBlock(block)
		if err != nil {
			return false, err
		}
		if bytes.Compare(calculated, block.Header.Hash) != 0 {
			logging.VLog().WithFields(logrus.Fields{
				"index":                k,
				"chunkData.size":       len(chunkData.Blocks),
				"chunkHeader.size":     len(chunkHeader.Headers),
				"data.header.hash":     byteutils.Hex(block.Header.Hash),
				"data.calculated.hash": byteutils.Hex(calculated),
				"err":                  ErrInvalidBlockHashInChunk,
			}).Debug("Invalid block hash.")
			return false, ErrInvalidBlockHashInChunk
		}
		if bytes.Compare(hash, block.Header.Hash) != 0 {
			logging.VLog().WithFields(logrus.Fields{
				"index":            k,
				"chunkData.size":   len(chunkData.Blocks),
				"chunkHeader.size": len(chunkHeader.Headers),
				"data.hash":        byteutils.Hex(block.Header.Hash),
				"header.hash":      byteutils.Hex(hash),
				"err":              ErrWrongBlockHashInChunk,
			}).Debug("Wrong block hash.")
			return false, ErrWrongBlockHashInChunk
		}
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

func (c *Chunk) processChunkData(chunk *syncpb.ChunkData) (*core.Block, error) {
	var last *core.Block
	for k, v := range chunk.Blocks {
		block := new(core.Block)
		if err := block.FromProto(v); err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"index": k,
				"hash":  byteutils.Hex(v.Header.Hash),
				"err":   err,
			}).Debug("Failed to recover a block from proto data.")
			return nil, err
		}
		if err := c.blockChain.BlockPool().Push(block); err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"index": k,
				"hash":  byteutils.Hex(v.Header.Hash),
				"err":   err,
			}).Debug("Failed to push a block into block pool.")
			return nil, err
		}
		last = block
	}

	logging.VLog().WithFields(logrus.Fields{
		"size": len(chunk.Blocks),
	}).Debug("Succeed to process chunk.")
	return last, nil
}
