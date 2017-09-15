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

// Transaction type is used to handle all transaction data.
type Transaction struct {
	hash  string
	from  Address
	to    Address
	value int64
	nonce int64
	data  []byte
}

// Transactions is an alias of Transaction array.
type Transactions []*Transaction

// NewTransaction create @Transaction instance.
func NewTransaction(from, to Address, value int64, nonce int64) *Transaction {
	tx := &Transaction{
		from:  from,
		to:    to,
		value: value,
		nonce: nonce,
	}
	return tx
}
