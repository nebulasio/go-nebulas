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

'use strict';

var Blockchain = function () {
    this.nativeBlockchain = new NativeBlockchain();
}

Blockchain.prototype = {
    getBlockByHash: function (hash) {
        var block = this.nativeBlockchain.getBlockByHash(hash);
        if (block != null) {
            block = JSON.parse(block)
        }
        return block
    },
    getTransactionByHash: function (hash) {
        var tx = this.nativeBlockchain.getTransactionByHash(hash);
        if (tx != null) {
            tx = JSON.parse(tx)
        }
        return tx
    },
    getAccountState: function (address) {
        var acc = this.nativeBlockchain.getAccountState(address);
        if (acc != null) {
            acc = JSON.parse(acc)
        }
        return acc
    },
    send: function (address, value) {
        return this.nativeBlockchain.send(address, value.toString())
    }
}

module.exports.Blockchain = Blockchain;

