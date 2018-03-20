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
if (typeof ContractStorage === 'undefined') {
    throw new Error("ContractStorage is undefined.");
}

var err = new Error("ContractStorage should accept _storage_handlers member only.");
try {
    new ContractStorage();
    throw err;
} catch (e) {
    if (e == err) {
        throw e;
    }
}

// disable gcs according to https://github.com/nebulasio/go-nebulas/issues/23
var _e = new Error('GlobalContractStorage should be disabled.');
try {
    GlobalContractStorage.put('k1', 'v1');
    throw _e;
} catch (e) {
    if (e == _e) {
        throw e;
    } else {
        // pass.
    }
}

[
    [LocalContractStorage, 'LocalContractStorage']
    // disable gcs according to https://github.com/nebulasio/go-nebulas/issues/23
    // [GlobalContractStorage, 'GlobalContractStorage']
].forEach(function (item) {
    var stor = item[0];
    var name = item[1];
    if (typeof stor === 'undefined') {
        throw new Error(name + " is undefined.");
    }

    var v1 = 'now is v1';
    stor.put('k1', v1);
    if (stor.get('k1') !== v1) {
        throw new Error("key k1 should return string [" + v1 + "].")
    }

    stor.del('k1');
    if (stor.get('k1') !== null) {
        throw new Error("key k1 should not exist, return null when get.")
    }

    stor.del('k2');

    ["haha", 123, false, true, null, {
        x: 1,
        y: {
            a: 2
        },
        z: ["zzz", 3, 4]
    }].forEach(function (v, idx) {
        var key = 'key-' + idx;
        stor.put(key, v);
        var val = stor.get(key);
        if (!Object.is(val, v)) {
            if (typeof v === 'object') {
                if (JSON.stringify(v) == JSON.stringify(val)) {
                    // pass.
                    return;
                }
            }
            throw new Error("ContractStorage should support value type " + typeof v + "; expected is " + v + ", actual is " + val);
        }
    });
});
