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

var BN = BigNumber;
if (!BN) {
    BN = require('bignumber.js');
}

const MAX_UINT64 = new BN('18446744073709551615', 16);
const MAX_UINT128 = new BN('340282366920938463463374607431768211455', 16);
const MAX_UINT256 = new BN('115792089237316195423570985008687907853269984665640564039457584007913129639935', 16);
const MAX_UINT512 = new BN('13407807929942597099574024998205846127479365820592393377723561443721764030073546976801874298166903427690031858186486050853753882811946569946433649006084095', 16);

const MAX_UINTS = {
    64: MAX_UINT64,
    128: MAX_UINT128,
    256: MAX_UINT256,
    512: MAX_UINT512
}

class Uint {
    constructor(n, b, s) {
        this.inner = new BN(n, b);
        this.size = s;

        this._validate();
    }

    _validate() {
        // check integer
        if (!this.inner.isInteger()) {
            throw new Error('[uint Error] not a integer');
        }

        // check negative
        if (this.inner.isNegative()) {
            throw new Error('[uint' + this.size + ' Error] underflow');
        }

        // check overflow
        if (this.inner.gt(MAX_UINTS[this.size])) {
            throw new Error('[uint' + this.size + ' Error] overflow');
        }
    }

    _checkOperands(left, right) {
        if (typeof left !== typeof right) {
            throw new Error('[uint Error] mismatched operand type');
        }

        right._validate();
    }

    div(o) {
        _checkOperands(this, o);
        var r = this.inner.idiv(o.inner);
        return new this.constructor(r, null, this.size);
    }

    pow(o) {
        _checkOperands(this, o);
        var r = this.inner.pow(o.inner);
        return new this.constructor(r, null, this.size);
    }

    minus(o) {
        _checkOperands(this, o);
        var r = this.inner.minus(o.inner);
        return new this.constructor(r, null, this.size);
    }

    mod(o) {
        _checkOperands(this, o);
        var r = this.inner.mod(o.inner);
        return new this.constructor(r, null, this.size);
    }

    mul(o) {
        _checkOperands(this, o);
        var r = this.inner.times(o.inner);
        return new this.constructor(r, null, this.size);
    }

    plus(o) {
        _checkOperands(this, o);
        var r = this.inner.plus(o.inner);
        return new this.constructor(r, null, this.size);
    }

    cmp(o) {
        _checkOperands(this, o);
        return this.inner.comparedTo(o.inner);
    }

    toString() {
        return this.inner.toString.call(this.inner, arguments);
    }
}

class Uint64 extends Uint {
    constructor(n, b) {
        super(n, b, 64);
    }
}

class Uint128 extends Uint {
    constructor(n, b) {
        super(n, b, 128);
    }
}

class Uint256 extends Uint {
    constructor(n, b) {
        super(n, b, 256);
    }
}
class Uint512 extends Uint {
    constructor(n, b) {
        super(n, b, 512);
    }
}

module.exports = {
    Uint64: Uint64,
    Uint128: Uint128,
    Uint256: Uint256,
    Uint512: Uint512
};