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

if (typeof _native_blockchain === "undefined") {
    throw new Error("_native_blockchain is undefined.");
}

var tx = _native_blockchain.getTransactionByHash("5e6d587f26121f96a07cf4b8b569aac1");
console.log("tx:" + tx);
try {
    JSON.parse(tx);
} catch (e) {
    throw error("tx parse err");
}

var accState = _native_blockchain.getAccountState("5e6d587f26121f96a07cf4b8b569aac1");
console.log("accState:" + accState);
try {
    JSON.parse(accState);
} catch (e) {
    throw error("accState parse err");
}

var result = _native_blockchain.transfer("5e6d587f26121f96a07cf4b8b569aac1", "1");
console.log("transfer:" + result)

var result = _native_blockchain.verifyAddress("70e30fcae5e7f4b2460faaa9e5b1bd912332ebb5");
console.log("verifyAddress:" + result)

var tx = Blockchain.getTransactionByHash("5e6d587f26121f96a07cf4b8b569aac1");
console.log("tx:" + tx.value);
