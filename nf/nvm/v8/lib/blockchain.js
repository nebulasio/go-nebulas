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
    this.nativeBlockchain = _native_blockchain;
};

Blockchain.prototype = {
    getTransactionByHash: function (hash) {
        var tx = this.nativeBlockchain.getTransactionByHash(hash);
        if (tx === null) {
            return null
        }
        tx = JSON.parse(tx);
        if (tx.value != undefined) {
            tx.value = new BigNumber(tx.value)
        }
        if (tx.gasPrice != undefined) {
            tx.gasPrice = new BigNumber(tx.gasPrice)
        }
        if (tx.gasLimit != undefined) {
            tx.gasLimit = new BigNumber(tx.gasLimit)
        }
        return tx
    },
    getAccountState: function (address) {
        var acc = this.nativeBlockchain.getAccountState(address);
        if (acc === null) {
            return null
        }
        acc = JSON.parse(acc);
        if (acc.balance != undefined) {
            acc.balance = new BigNumber(acc.balance)
        }
        return acc
    },
    transfer: function (address, value) {
        return this.nativeBlockchain.transfer(address, value.toString());
    },
    verifyAddress: function (address) {
        return this.nativeBlockchain.verifyAddress(address);
    }
};

module.exports = new Blockchain();
module.exports.Blockchain = Blockchain;
