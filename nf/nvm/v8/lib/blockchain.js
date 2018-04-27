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

var Contract = function(address, contract_interface) {
    //check args
    if (typeof(address) != "string") {
        throw("contract address should be a string");
    }
    if (typeof(contract_interface) != 'object') {
        throw("wrong interface");
    }

    this.v = 0;
    this.address = address;

    //TODO:
    //load callee contract
    var src = _native_blockchain.getContractSource(address);
    if (src == null) {
        throw("no contract at this address " + address);
    }
    // var arguments_length = contract_interface[func].length;  
    for(var func in contract_interface) {
        //check propertys in interface are function
        if (typeof(contract_interface[func]) !== 'function') {
            throw("wrong interface define")
        }
        //TODO: check if this works 
        var expression = new RegExp(func + " *: *function", "m");
        if (src.search(expression) == -1) {
            throw("contract have no function called : " + func);
        } 
    }
        
}

Contract.prototype = {
    value: function (value) {
        if (value == null) {
            this.v = 0;
        } else {
            this.v = new BigNumber(v);
        }    
        return this
    },
    call: function (func, args) {
        if (typeof(func) != "string") {
            throw("function name should be a string")
        }

        if (typeof(args) != "args") {
            throw("function args should be a string")
        }

        var value = this.v;
        this.v = 0;
        //TODO: check how to handle err?
        return _native_blockchain.runContractSource(this.address, func, value.toString(), args);
    }
}


var Blockchain = function () {
    this.nativeBlockchain = _native_blockchain;
    this.Contract = Contract;
};

Blockchain.prototype = {
    AccountAddress: 0x57,
    ContractAddress: 0x58,

    blockParse: function (str) {
        var block = JSON.parse(str);
        if (block != null) {
            this.block = Object.freeze(block);
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
            this.transaction = Object.freeze(tx);
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
    },
    
};

module.exports = new Blockchain();
