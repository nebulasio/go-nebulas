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
if (typeof Storage === "undefined") {
    throw new Error("Storage is undefined.");
}
if (typeof Storage !== "function") {
    throw new Error("Storage is not a function.");
}

try {
    new Storage({});
    throw new Error("Storage should only accept _storage_handlers member.");
} catch (e) {
    // pass.
}

[_storage_handlers.lcs, _storage_handlers.gcs].forEach(function (handler) {
    var stor = new Storage(handler);
    if (stor.get("non-exist-key") !== null) {
        throw new Error("get non-exist-key should return null.");
    }

    stor.put('k1', 'welcome');
    if (stor.get('k1') !== 'welcome') {
        throw new Error("key k1 should return string [welcome].")
    }

    stor.del('k1');
    if (stor.get('k1') !== null) {
        throw new Error("key k1 should not exist, return null when get.")
    }

    stor.del('k2');

    [123, {}, function () {}, undefined].forEach(function (v) {
        try {
            stor.put('k3', v);
            throw new Error('value of put should be string, not support ' + typeof v);
        } catch (e) {}
    });
});
