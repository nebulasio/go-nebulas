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

package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util/byteutils"
)

func main() {
	file := os.Args[1]
	cnt, err := strconv.ParseInt(os.Args[2], 10, 32)
	if err != nil {
		fmt.Println("Parse Int Error ", err)
		return
	}
	rootHash := os.Args[3]
	stor, err := storage.NewDiskStorage(file)
	if err != nil {
		fmt.Println("OpenDB Error ", err)
		return
	}
	fmt.Println("cnt:", cnt, " file:", file, " root:", rootHash)

	startAt := time.Now().Unix()
	root, err := byteutils.FromHex(rootHash)
	if err != nil {
		fmt.Println("Parse Hex Error ", err)
		return
	}
	txsState, err := trie.NewTrie(root, stor, false)
	if err != nil {
		fmt.Println("NewTrie Error ", err)
		return
	}
	iter, err := txsState.Iterator(nil)
	if err != nil {
		fmt.Println("Iterator Error ", err)
		return
	}
	exist, err := iter.Next()
	if err != nil {
		fmt.Println("Next Error1 ", err)
		return
	}
	i := cnt
	for exist {
		exist, err = iter.Next()
		i--
		if i == 0 {
			fmt.Println("Read Over")
			break
		}
		if err != nil {
			fmt.Println("Next Error2 ", err)
			return
		}
	}
	endAt := time.Now().Unix()
	diff := endAt - startAt
	if diff == 0 {
		fmt.Println("Diff is zero")
		return
	}
	op := cnt - i
	fmt.Println("TPS ", op, "/", diff, "s = ", op/diff)
	fmt.Println("Done")
}
