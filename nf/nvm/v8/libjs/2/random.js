// Copyright (C) 2018 go-nebulas
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

// A port of an algorithm by Johannes Baagøe <baagoe@baagoe.com>, 2010
// http://baagoe.com/en/RandomMusings/javascript/
// https://github.com/nquinlan/better-random-numbers-for-javascript-mirror
// Original work is under MIT license -

// Other seeded random number generators for JavaScript, see https://github.com/davidbau/seedrandom.


'use strict';

function Alea(seed) {
    var me = this, mash = Mash();

    me.next = function () {
        var t = 2091639 * me.s0 + me.c * 2.3283064365386963e-10; // 2^-32
        me.s0 = me.s1;
        me.s1 = me.s2;
        return me.s2 = t - (me.c = t | 0);
    };

    // Apply the seeding algorithm from Baagoe.
    me.c = 1;
    me.s0 = mash(' ');
    me.s1 = mash(' ');
    me.s2 = mash(' ');
    me.s0 -= mash(seed);
    if (me.s0 < 0) { me.s0 += 1; }
    me.s1 -= mash(seed);
    if (me.s1 < 0) { me.s1 += 1; }
    me.s2 -= mash(seed);
    if (me.s2 < 0) { me.s2 += 1; }
    mash = null;
}

function copy(f, t) {
    t.c = f.c;
    t.s0 = f.s0;
    t.s1 = f.s1;
    t.s2 = f.s2;
    return t;
}

function impl(seed, opts) {
    var xg = new Alea(seed),
        state = opts && opts.state,
        prng = xg.next;
    prng.int32 = function () { return (xg.next() * 0x100000000) | 0; }
    prng.double = function () {
        return prng() + (prng() * 0x200000 | 0) * 1.1102230246251565e-16; // 2^-53
    };
    prng.quick = prng;
    if (state) {
        if (typeof (state) == 'object') copy(state, xg);
        prng.state = function () { return copy(xg, {}); }
    }
    return prng;
}

function Mash() {
    var n = 0xefc8249d;

    var mash = function (data) {
        data = data.toString();
        for (var i = 0; i < data.length; i++) {
            n += data.charCodeAt(i);
            var h = 0.02519603282416938 * n;
            n = h >>> 0;
            h -= n;
            h *= n;
            n = h >>> 0;
            h -= n;
            n += h * 0x100000000; // 2^32
        }
        return (n >>> 0) * 2.3283064365386963e-10; // 2^-32
    };

    return mash;
}

module.exports = (function(){

    var arng = null;

    function checkCtx() {
        if (!Blockchain) {
            throw new Error("'Blockchain' is undefined.");
        }
        if (!Blockchain.block) {
            throw new Error("'Blockchain.block' is undefined.");
        }
    
        if (Blockchain.block.seed == null || typeof(Blockchain.block.seed) === 'undefined') {
            throw new Error("Math.random func is not allowed in nvm.");
        }
    }

    function rand() {
        if (arng == null) {
            checkCtx();
            arng = new impl(Blockchain.block.seed);
        }
        return arng();
    }
    rand.seed = function(userseed) {
        if (typeof(userseed) !== 'string') {
            throw new Error("input seed must be a string")
        }
        checkCtx();
        arng = new impl(Blockchain.block.seed + userseed);
    }

    return rand;
})();