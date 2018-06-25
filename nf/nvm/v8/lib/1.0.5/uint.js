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

/*
 * this module must be required after bignumber.js
 */
'use strict';

/*
 * ffffffffffffffff 18446744073709551615
 * ffffffffffffffffffffffffffffffff 340282366920938463463374607431768211455
 * ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff 115792089237316195423570985008687907853269984665640564039457584007913129639935
 * ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff 13407807929942597099574024998205846127479365820592393377723561443721764030073546976801874298166903427690031858186486050853753882811946569946433649006084095
 */
const MAX_UINT64 = new BigNumber('ffffffffffffffff', 16);
const MAX_UINT128 = new BigNumber('ffffffffffffffffffffffffffffffff', 16);
const MAX_UINT256 = new BigNumber('ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff', 16);
const MAX_UINT512 = new BigNumber('ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff', 16);

const MAX_UINTS = {
    64: MAX_UINT64,
    128: MAX_UINT128,
    256: MAX_UINT256,
    512: MAX_UINT512
}

class Uint {
    constructor(n, b, s) {

        Object.defineProperties(this, {
            _inner: {
                value: new BigNumber(n, b)
            },
            _size: {
                value: s
            }
        });

        this._validate();
    }

    _validate() {
        // check integer
        if (!this._inner.isInteger()) {
            throw new Error('[Uint' + this._size + ' Error] not an integer');
        }

        // check negative
        if (this._inner.isNegative()) {
            throw new Error('[Uint' + this._size + ' Error] underflow');
        }

        // check overflow
        if (this._inner.gt(MAX_UINTS[this._size])) {
            throw new Error('[Uint' + this._size + ' Error] overflow');
        }
    }

    _checkRightOperand(right) {
        if (typeof right === 'undefined' || right == null) {
            throw new Error('[Uint' + this._size + ' Error] NaN');
        }

        if (!right instanceof Uint || this.constructor !== right.constructor) {
            throw new Error('[Uint' + this._size + ' Error] incompatible type');
        }
        right._validate();
    }

    div(o) {
        this._checkRightOperand(o);
        var r = this._inner.divToInt(o._inner);
        return new this.constructor(r, null, this._size);
    }

    pow(o) {
        this._checkRightOperand(o);
        var r = this._inner.pow(o._inner);
        return new this.constructor(r, null, this._size);
    }

    minus(o) {
        this._checkRightOperand(o);
        var r = this._inner.minus(o._inner);
        return new this.constructor(r, null, this._size);
    }

    mod(o) {
        this._checkRightOperand(o);
        var r = this._inner.mod(o._inner);
        return new this.constructor(r, null, this._size);
    }

    mul(o) {
        this._checkRightOperand(o);
        var r = this._inner.times(o._inner);
        return new this.constructor(r, null, this._size);
    }

    plus(o) {
        this._checkRightOperand(o);
        var r = this._inner.plus(o._inner);
        return new this.constructor(r, null, this._size);
    }

    cmp(o) {
        this._checkRightOperand(o);
        return this._inner.comparedTo(o._inner);
    }

    isZero() {
        return this._inner.isZero();
    }

    toString() {
        return this._inner.toString.apply(this._inner, Array.prototype.slice.call(arguments));
    }
}

class Uint64 extends Uint {
    constructor(n, b) {
        super(n, b, 64);
    }

    static get MaxValue () {
        return new Uint64(MAX_UINTS[64], null, 64);
    }
}

class Uint128 extends Uint {
    constructor(n, b) {
        super(n, b, 128);
    }

    static get MaxValue () {
        return new Uint128(MAX_UINTS[128], null, 128);
    }
}

class Uint256 extends Uint {
    constructor(n, b) {
        super(n, b, 256);
    }

    static get MaxValue () {
        return new Uint256(MAX_UINTS[256], null, 256);
    }
}
class Uint512 extends Uint {
    constructor(n, b) {
        super(n, b, 512);
    }

    static get MaxValue () {
        return new Uint512(MAX_UINTS[512], null, 512);
    }
}

module.exports = {
    Uint64: Uint64,
    Uint128: Uint128,
    Uint256: Uint256,
    Uint512: Uint512,
    isUint: function(o) {
        return o instanceof Uint;
    }
};