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
    blockParse: function (str) {
        var block = JSON.parse(str);
        if (block != null) {
            this.block = block;
        }
    },
    transactionParse: function (str) {
        var tx = JSON.parse(str);
        if (tx != null) {
            var value = tx.value === undefined || tx.value.length === 0 ? "0" : tx.value;
            tx.value = new BigNumber(value);
            var gasPrice = tx.gasPrice === undefined || tx.gasPrice.length === 0 ? "0" : tx.gasPrice;
            tx.gasPrice = new BigNumber(gasPrice);
            var gasLimit = tx.gasLimit === undefined || tx.gasLimit.length === 0 ? "0" : tx.gasLimit;
            tx.gasLimit = new BigNumber(gasLimit);
            this.transaction = tx;
        }
    },
    // The current NVM does not support.
    // getTransactionByHash: function (hash) {
    //     var tx = this.nativeBlockchain.getTransactionByHash(hash);
    //     if (tx === null) {
    //         return null
    //     }
    //     tx = JSON.parse(tx);
    //     var value = tx.value === undefined || tx.value.length === 0 ? "0" : tx.value;
    //     tx.value = new BigNumber(value);
    //     var gasPrice = tx.gasPrice === undefined || tx.gasPrice.length === 0 ? "0" : tx.gasPrice;
    //     tx.gasPrice = new BigNumber(gasPrice);
    //     var gasLimit = tx.gasLimit === undefined || tx.gasLimit.length === 0 ? "0" : tx.gasLimit;
    //     tx.gasLimit = new BigNumber(gasLimit);
    //     return tx
    // },
    // getAccountState: function (address) {
    //     var acc = this.nativeBlockchain.getAccountState(address);
    //     if (acc === null) {
    //         return null
    //     }
    //     acc = JSON.parse(acc);
    //     var balance = acc.balance === undefined || acc.balance.length === 0 ? "0" : acc.balance;
    //     acc.balance = new BigNumber(balance);
    //     if (acc.nonce === undefined) {
    //         acc.nonce = 0;
    //     }
    //     return acc
    // },
    transfer: function (address, value) {
        if (!(value instanceof BigNumber)) {
            value = new BigNumber(value);
        }
        return this.nativeBlockchain.transfer(address, value.toString(10));
    },
    verifyAddress: function (address) {
        return this.nativeBlockchain.verifyAddress(address);
    }
};

module.exports = new Blockchain();
module.exports.Blockchain = Blockchain;
