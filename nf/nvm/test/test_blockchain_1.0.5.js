// Copyright (C) 2018 go-nebulas authors
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

var ok = Blockchain.transfer("n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE", "1");
console.log("transfer:" + ok)

var result = Blockchain.verifyAddress("n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE");
console.log("verifyAddress:" + result)

try {
    Blockchain.transfer("n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE", -1);
} catch (err) {
    if (err.message !== "invalid value" ) {
        throw err;
    }
}
try {
    Blockchain.transfer("n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE", NaN);
} catch (err) {
    if (err.message !== "invalid value" ) {
        throw err;
    }
}
try {
    Blockchain.transfer("n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE", Infinity);
} catch (err) {
    if (err.message !== "invalid value" ) {
        throw err;
    }
}