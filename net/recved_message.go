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

package net

import (
	"fmt"
	"sync"

	"github.com/willf/bloom"
)

const (
	// according to https://krisives.github.io/bloom-calculator/
	// Count (n) = 100000, Error (p) = 0.001
	maxCountOfRecvMessageInBloomFiler = 100000
	bloomFilterOfRecvMessageArgM      = 1437759
	bloomFilterOfRecvMessageArgK      = 10
)

var (
	bloomFilterOfRecvMessage        = bloom.New(bloomFilterOfRecvMessageArgM, bloomFilterOfRecvMessageArgK)
	bloomFilterMutex                sync.Mutex
	countOfRecvMessageInBloomFilter = 0
)

// RecordKey add key to bloom filter.
func RecordKey(key string) {
	bloomFilterMutex.Lock()
	defer bloomFilterMutex.Unlock()

	countOfRecvMessageInBloomFilter++
	if countOfRecvMessageInBloomFilter > maxCountOfRecvMessageInBloomFiler {
		// reset.
		bloomFilterOfRecvMessage = bloom.New(bloomFilterOfRecvMessageArgM, bloomFilterOfRecvMessageArgK)
	}

	bloomFilterOfRecvMessage.AddString(key)
}

// HasKey use bloom filter to check if the key exists quickly
func HasKey(key string) bool {
	bloomFilterMutex.Lock()
	defer bloomFilterMutex.Unlock()

	return bloomFilterOfRecvMessage.TestString(key)
}

// RecordRecvMessage records received message
func RecordRecvMessage(s *Stream, hash uint32) {
	RecordKey(fmt.Sprintf("%s-%d", s.pid, hash))
}

// HasRecvMessage check if the received message exists before
func HasRecvMessage(s *Stream, hash uint32) bool {
	return HasKey(fmt.Sprintf("%s-%d", s.pid, hash))
}
