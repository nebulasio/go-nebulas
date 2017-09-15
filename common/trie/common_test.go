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

package trie

import (
	"fmt"
	"testing"
)

func TestHex(t *testing.T) {
	str := "afb1"
	bytes := []byte{0xaf, 0xb1}
	hexStr := hex(bytes)
	if str != hexStr {
		t.Errorf("hex %v failed, got %v, expect %v", bytes, str, hexStr)
	}
	unhexBytes, err := unhex(hexStr)
	if err != nil {
		t.Errorf("unhex %v corrupt", hexStr)
	}
	for k, v := range unhexBytes {
		if v != bytes[k] {
			t.Errorf("unhex %v failed, at %d got %v, expect %v", hexStr, k, v, bytes[k])
		}
	}
}

func TestHash(t *testing.T) {
	bytes := []byte{0xaf, 0xb1}
	digest := hash(bytes)
	if len(digest) != 32 {
		t.Errorf("hash %v failed, length got %v, expect 64", bytes, len(digest))
	}
	fmt.Println(hex(digest))
}
