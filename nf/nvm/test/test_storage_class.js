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
if (typeof NativeStorage === "undefined") {
    throw new Error("NativeStorage is undefined.");
}
if (typeof NativeStorage !== "function") {
    throw new Error("NativeStorage is not a function.");
}

try {
    new NativeStorage({});
    throw new Error("NativeStorage should accept _storage_handlers member only.");
} catch (e) {
    // pass.
}

// disable gcs according to https://github.com/nebulasio/go-nebulas/issues/23
var _e = new Error('_native_storage_handlers.gcs should be disabled.');
try {
    _native_storage_handlers.gcs.put('k1', 'v1');
    throw _e;
} catch (e) {
    if (e == _e) {
        throw e;
    } else {
        // pass.
    }
}

// [_native_storage_handlers.lcs, _native_storage_handlers.gcs].forEach(function (handler) {
[_native_storage_handlers.lcs].forEach(function (handler) {
    var stor = new NativeStorage(handler);
    if (stor.get("non-exist-key") !== null) {
        throw new Error("get non-exist-key should return null.");
    }

    var v1 = 'this is v1';
    if (stor.put('k1', v1) != 0) {
        throw new Error("put k1 failed.");
    }

    if (stor.get('k1') !== v1) {
        throw new Error("key k1 should return string [" + v1 + "].")
    }

    if (stor.del('k1') != 0) {
        throw new Error("del k1 failed.");
    }

    if (stor.get('k1') !== null) {
        throw new Error("key k1 should not exist, return null when get.")
    }

    if (stor.del('k2') != 0) {
        throw new Error("del k2 failed.");
    }

    [123, {}, function () {}, undefined].forEach(function (v) {
        var err = new Error('value of put should be string, not support ' + typeof v);
        var err1 = new Error("put k3 failed.");
        try {
            if (stor.put('k3', v) != 0) {
                throw err1;
            }
            throw err;
        } catch (e) {
            if (e == err || e == err1) {
                throw e;
            }
        }
    });
});
