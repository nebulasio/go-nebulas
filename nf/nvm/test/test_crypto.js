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
eq(crypto.recoverAddress(1, "564733f9f3e139b925cfb1e7e50ba8581e9107b13e4213f2e4708d9c284be75b", "d80e282d165f8c05d8581133df7af3c7c41d51ec7cd8470c18b84a31b9af6a9d1da876ab28a88b0226707744679d4e180691aca6bdef5827622396751a0670c101"), "n1F8QbdnhqpPXDPFT2c9a581tpia8iuF7o2");
eq(crypto.md5(input), "9954125a33a380c3117269cff93f76a7");
eq(crypto.base64(input), "TmVidWxhcyBpcyBhIG5leHQgZ2VuZXJhdGlvbiBwdWJsaWMgYmxvY2tjaGFpbiwgYWltaW5nIGZvciBhIGNvbnRpbnVvdXNseSBpbXByb3ZpbmcgZWNvc3lzdGVtLg==");

// alg is not a safe integer
try {
    crypto.recoverAddress(1000000000000000000010000000000000000000, "564733f9f3e139b925cfb1e7e50ba8581e9107b13e4213f2e4708d9c284be75b",
     "d80e282d165f8c05d8581133df7af3c7c41d51ec7cd8470c18b84a31b9af6a9d1da876ab28a88b0226707744679d4e180691aca6bdef5827622396751a0670c101");
} catch (err) {
    if (err.message !== "alg must be non-negative integer") {
        throw err;
    }
}

// negative alg
try {
    crypto.recoverAddress(-1000, "564733f9f3e139b925cfb1e7e50ba8581e9107b13e4213f2e4708d9c284be75b",
     "d80e282d165f8c05d8581133df7af3c7c41d51ec7cd8470c18b84a31b9af6a9d1da876ab28a88b0226707744679d4e180691aca6bdef5827622396751a0670c101");
} catch (err) {
    if (err.message !== "alg must be non-negative integer") {
        throw err;
    }
}

// odd hash
try {
    crypto.recoverAddress(1, "564733f9f3e139b925cfb1e7e50ba8581e9107b13e4213f2e4708d9c284be75",
     "d80e282d165f8c05d8581133df7af3c7c41d51ec7cd8470c18b84a31b9af6a9d1da876ab28a88b0226707744679d4e180691aca6bdef5827622396751a0670c101");
} catch (err) {
    if (err.message !== "hash & sign must be hex string") {
        throw err;
    }
}

// not hex hash
try {
    crypto.recoverAddress(1, "TT564733f9f3e139b925cfb1e7e50ba8581e9107b13e4213f2e4708d9c284be75b",
     "d80e282d165f8c05d8581133df7af3c7c41d51ec7cd8470c18b84a31b9af6a9d1da876ab28a88b0226707744679d4e180691aca6bdef5827622396751a0670c101");
} catch (err) {
    if (err.message !== "hash & sign must be hex string") {
        throw err;
    }
}

// not hex sign
try {
    crypto.recoverAddress(1, "564733f9f3e139b925cfb1e7e50ba8581e9107b13e4213f2e4708d9c284be75b",
     "TTd80e282d165f8c05d8581133df7af3c7c41d51ec7cd8470c18b84a31b9af6a9d1da876ab28a88b0226707744679d4e180691aca6bdef5827622396751a0670c101");
} catch (err) {
    if (err.message !== "hash & sign must be hex string") {
        throw err;
    }
}

// odd sign
try {
    crypto.recoverAddress(1, "564733f9f3e139b925cfb1e7e50ba8581e9107b13e4213f2e4708d9c284be75b",
     "d80e282d165f8c05d8581133df7af3c7c41d51ec7cd8470c18b84a31b9af6a9d1d876ab28a88b0226707744679d4e180691aca6bdef5827622396751a0670c101");
} catch (err) {
    if (err.message !== "hash & sign must be hex string") {
        throw err;
    }
}

try {
    crypto.nativeCrypto.sha256(1231432);
} catch (err) {
    if (err !== "sha256() requires a string argument") {
        throw err;
    }
}

try {
    crypto.nativeCrypto.sha256();
} catch (err) {
    if (err !== "sha256() requires only 1 argument") {
        throw err;
    }
}

try {
    crypto.nativeCrypto.sha3256(null);
} catch (err) {
    if (err !== "sha3256() requires a string argument") {
        throw err;
    }
}

try {
    crypto.nativeCrypto.sha3256();
} catch (err) {
    if (err !== "sha3256() requires only 1 argument") {
        throw err;
    }
}

try {
    crypto.nativeCrypto.ripemd160();
} catch (err) {
    if (err !== "ripemd160() requires only 1 argument") {
        throw err;
    }
}

try {
    var ret = crypto.nativeCrypto.ripemd(-121);
} catch (err) {
    if (err.message !== "crypto.nativeCrypto.ripemd is not a function") {
        throw err;
    }
}

// negative alg
var ret = crypto.nativeCrypto.recoverAddress(-10, "564733f9f3e139b925cfb1e7e50ba8581e9107b13e4213f2e4708d9c284be75b", "d80e282d165f8c05d8581133df7af3c7c41d51ec7cd8470c18b84a31b9af6a9d1da876ab28a88b0226707744679d4e180691aca6bdef5827622396751a0670c101");
eq(ret, null);

// invalid alg
ret = crypto.nativeCrypto.recoverAddress(10, "564733f9f3e139b925cfb1e7e50ba8581e9107b13e4213f2e4708d9c284be75b", "d80e282d165f8c05d8581133df7af3c7c41d51ec7cd8470c18b84a31b9af6a9d1da876ab28a88b0226707744679d4e180691aca6bdef5827622396751a0670c101");
eq(ret, null);

// odd/invalid sign
ret = crypto.nativeCrypto.recoverAddress(1, "564733f9f3e139b925cfb1e7e50ba8581e9107b13e4213f2e4708d9c284be75b", "d80e282d165f8c05d8581133df7af3c7c41d51ec7cd");
eq(ret, null);

// empty sign
ret = crypto.nativeCrypto.recoverAddress(1, "564733f9f3e139b925cfb1e7e50ba8581e9107b13e4213f2e4708d9c284be75b", "");
eq(ret, null);

// empty hash
ret = crypto.nativeCrypto.recoverAddress(1, "", "d80e282d165f8c05d8581133df7af3c7c41d51ec7cd8470c18b84a31b9af6a9d1da876ab28a88b0226707744679d4e180691aca6bdef5827622396751a0670c101");
eq(ret, null);

// odd/invalid hash
ret = crypto.nativeCrypto.recoverAddress(1, "564733f9f3e139b925cfb1e7e50ba8581e9107b13e421", "d80e282d165f8c05d8581133df7af3c7c41d51ec7cd8470c18b84a31b9af6a9d1da876ab28a88b0226707744679d4e180691aca6bdef5827622396751a0670c101");
eq(ret, null);

try {
    crypto.nativeCrypto.recoverAddress("", "564733f9f3e139b925cfb1e7e50ba8581e9107b13e4213f2e4708d9c284be75b", "d80e282d165f8c05d8581133df7af3c7c41d51ec7cd8470c18b84a31b9af6a9d1da876ab28a88b0226707744679d4e180691aca6bdef5827622396751a0670c101");
} catch (err) {
    if (err !== "recoverAddress(): 1st arg should be integer") {
        throw err;
    }
}
try {
    crypto.nativeCrypto.recoverAddress(null, "564733f9f3e139b925cfb1e7e50ba8581e9107b13e4213f2e4708d9c284be75b", "d80e282d165f8c05d8581133df7af3c7c41d51ec7cd8470c18b84a31b9af6a9d1da876ab28a88b0226707744679d4e180691aca6bdef5827622396751a0670c101");
} catch (err) {
    if (err !== "recoverAddress(): 1st arg should be integer") {
        throw err;
    }
}

try {
    crypto.nativeCrypto.recoverAddress(1, 123, "d80e282d165f8c05d8581133df7af3c7c41d51ec7cd8470c18b84a31b9af6a9d1da876ab28a88b0226707744679d4e180691aca6bdef5827622396751a0670c101");
} catch (err) {
    if (err !== "recoverAddress(): 2nd arg should be string") {
        throw err;
    }
}

try {
    crypto.nativeCrypto.recoverAddress(1, "564733f9f3e139b925cfb1e7e50ba8581e9107b13e4213f2e4708d9c284be75b", null);
} catch (err) {
    if (err !== "recoverAddress(): 3rd arg should be string") {
        throw err;
    }
}

try {
    crypto.nativeCrypto.recoverAddress(1, "564733f9f3e139b925cfb1e7e50ba8581e9107b13e4213f2e4708d9c284be75b");
} catch (err) {
    if (err !== "recoverAddress() requires 3 arguments") {
        throw err;
    }
}
