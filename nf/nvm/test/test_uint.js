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

var Uint64 = Uint.Uint64;
var Uint128 = Uint.Uint128;
var Uint256 = Uint.Uint256;
var Uint512 = Uint.Uint512;

var err1 = "[Uint64 Error] overflow";
var err2 = '[Uint128 Error] incompatible type';
var err3 = "[Uint64 Error] underflow";
var err4 = '[Uint512 Error] incompatible type';
var err5 = '[Uint256 Error] overflow';
var err6 = '[Uint64 Error] NaN';
var err7 = '[Uint64 Error] not an integer';

var a  = new Uint64(100000000000000);
var a1 = new Uint64(100000000000001);
var b = new Uint128(100000000000000);
var c = new Uint256(100000000000000);
var d = new Uint512(100000000000000);

var f = new Uint64(0);
if (f.toString(10) !== "0") {
    throw new Error("not eq");
}

if (!Uint.isUint(a)) {
    throw new Error("uint64 should be uint");
}

if (Uint.isUint(new BigNumber("123"))) {
    throw new Error("bignumber should not be uint");
}

if (Uint.isUint(123)) {
    throw new Error("number should not be uint");
}

if (Uint.isUint("123")) {
    throw new Error("string should not be uint");
}

// overflow
try {
    a.pow(new Uint64(2));
} catch (e) {
    if (e.message !== err1) {
        throw e;
    }
}
a.pow(new Uint64(1));

try {
    c.mul(c).mul(c).mul(c).mul(c).mul(c).mul(c);
} catch (e) {
    if (e.message !== err5) {
        throw e;
    }
}

var bpow2 = b.pow(new Uint128(2));
if (bpow2.toString(10) !== "10000000000000000000000000000") {
    throw new Error("b.pow(2) not equal");
}

// incompatible
try {
    b.plus(c);
} catch (e) {
    if (e.message !== err2) {
        throw e;
    }
}
b.plus(b);

try {
    d.minus(1);
} catch (e) {
    if (e.message !== err4) {
        throw e;
    }
}

// underflow
try {
    a.minus(a1);
} catch (e) {
    if (e.message !== err3) {
        throw e;
    }
}
if (a.minus(a).toString(10) !== "0") {
    throw new Error("a.minus(a) not 0");
}

// NaN
try {
    a.div(null);
} catch (e) {
    if (e.message !== err6) {
        throw e;
    }
}

if (a.div(a).toString(10) !== "1") {
    throw new Error("a.div(a) not 1");
}

if (a.mod(a).toString(10) !== "0") {
    throw new Error("a.mod(a) not 0");
}

// not an integer
try {
    new Uint64(1.2);
} catch (e) {
    if (e.message !== err7) {
        throw e;
    }
}
try {
    new Uint64("1.2");
} catch (e) {
    if (e.message !== err7) {
        throw e;
    }
}
try {
    a.div(new Uint64(0));
} catch (e) {
    if (e.message !== err7) {
        throw e;
    }
}