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

var callFunc = function (func, args) {//TODO：检查是否会被覆盖。
    if (typeof(func) != "string") {
        throw("Inner Call: function name should be a string");
    }

    if (typeof(args) != "string") {
        throw("Inner Call: function args should be a string");
    }

    if (!this.methods[func]) {
        throw("Inner Call: matches no function in the interface");
    }

    var result =  _native_blockchain.runContractSource(this.address, func, this.v.toString(10), args);
    if (result) {
        return JSON.parse(result);
    } else {
        throw "Inner Call: TODO:"; // will be executed if  runContractSource fails
    }
}
 
var funcGenerator = function (func) {
    return function() {
        var args = new Array();
        for (var i = 0; i < arguments.length; i++) {
            args.push(arguments[i]);
        }

        var result =  _native_blockchain.runContractSource(this.address, func, this.v.toString(10), JSON.stringify(args));
        if (result) {
            //if no return, result === "\"\"",  JSON.parse(result) === "", !!JSON.parse(result) === false,
            //if return is a number,2, result === "1", typeof(JSON.parse(result)) === 'number'; 
            return JSON.parse(result);
        } else {
            throw "Inner Call: TODO:";
        }
    }
}

var dumpContract = function (address, v, methods) {
    this.address = address;
    this.v = v;
    this.methods = methods;

    for (var func in this.methods) {
        this[func] = funcGenerator(func);
    }
}

dumpContract.prototype = {
    call: callFunc,
}

var Contract = function(address, contract_interface) {
    //check args
    if (typeof(address) != "string") {
        throw("Inner Call: contract address should be a string");
    }
    if (typeof(contract_interface) != 'object') {
        throw("Inner Call: wrong interface");
    }

    this.v = new BigNumber(0);
    this.address = address;
    this.methods = {};

    var src = _native_blockchain.getContractSource(address);
    if (src == null) {
        throw("Inner Call: no contract at this address");
    }

    // var arguments_length = contract_interface[func].length;  
    for (var func in contract_interface) {
        if (func === "call" || func === "value") {
            continue;
        }

        //check propertys in interface are function
        if (typeof(contract_interface[func]) !== 'function') {
            throw("Inner Call: wrong interface define");
        }
      
        var expression = new RegExp(func, "m");
        if (src.search(expression) == -1) {
            throw("Inner Call: function not implenmented");
        }
        
        this.methods[func] = 1;
        this[func] = funcGenerator(func);
    }
        
}

Contract.prototype = {
    value: function (value) {
        var v = value || 0;
        v = new BigNumber(v);
        
        if (!v.isInteger() || v.lessThan(0)) {
            throw("Inner Call: invalid value");
        }

        return new dumpContract(this.address, v, this.methods);
    },

    call: callFunc,
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
        //console.log("-----ret:err", ret, err);
        return ret == 0;
    },
    verifyAddress: function (address) {
        return this.nativeBlockchain.verifyAddress(address);
    },
    
};
module.exports = new Blockchain();
