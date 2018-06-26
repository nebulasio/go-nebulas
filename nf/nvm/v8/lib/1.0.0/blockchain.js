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
    Object.defineProperty(this, "nativeBlockchain", {
        configurable: false,
        enumerable: false,
        get: function(){
            return _native_blockchain;
        }
    });
};

Blockchain.prototype = {
    AccountAddress: 0x57,
    ContractAddress: 0x58,

    blockParse: function (str) {
        var block = JSON.parse(str);
        if (block != null) {
            var fb = Object.freeze(block);
            Object.defineProperty(this, "block", {
                configurable: false,
                enumerable: false,
                get: function(){
                    return fb;
                }
            });
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
            
            var ft = Object.freeze(tx);
            Object.defineProperty(this, "transaction", {
                configurable: false,
                enumerable: false,
                get: function(){
                    return ft;
                }
            });
        }
    },
    transfer: function (address, value) {
        if (!(value instanceof BigNumber)) {
            value = new BigNumber(value);
        }
        var ret = this.nativeBlockchain.transfer(address, value.toString(10));
        return ret == 0;
    },

    verifyAddress: function (address) {
        return this.nativeBlockchain.verifyAddress(address);
    }
};
module.exports = new Blockchain();