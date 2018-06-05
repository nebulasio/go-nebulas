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

function eq(a, b) {
    if (a !== b) {
        throw new Error("Not equal: " + a + " <--> " + b);
    }
}

var crypto = require('crypto.js');

var input = "Nebulas is a next generation public blockchain, aiming for a continuously improving ecosystem."

// 
eq(crypto.sha256(input), "a32d6d686968192663b9c9e21e6a3ba1ba9b2e288470c2f98b790256530933e0");
eq(crypto.sha3256(input), "564733f9f3e139b925cfb1e7e50ba8581e9107b13e4213f2e4708d9c284be75b");
eq(crypto.ripemd160(input), "4236aa9974eb7b9ddb0f7a7ed06d4bf3d9c0e386");
eq(crypto.recoverAddress(1, "564733f9f3e139b925cfb1e7e50ba8581e9107b13e4213f2e4708d9c284be75b", "d80e282d165f8c05d8581133df7af3c7c41d51ec7cd8470c18b84a31b9af6a9d1da876ab28a88b0226707744679d4e180691aca6bdef5827622396751a0670c101"), "n1F8QbdnhqpPXDPFT2c9a581tpia8iuF7o2")