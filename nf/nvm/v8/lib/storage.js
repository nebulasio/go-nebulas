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

//TODO: @robin using blockchain api instead the javascript map.
'use strict';

var localContractStorage = function () {
    this._map = new Map();
};


var globalContractStorage = function () {
    this._map = new Map();
};

localContractStorage.prototype = {
    clear: function () {
        this._map.clear();
    },
    delete: function (key) {
        return this._map.delete(key);
    },
    forEach: function (callbackFn, thisArg) {
        return this._map.forEach(callbackFn, thisArg);
    },
    get: function (key) {
        return this._map.get(key);
    },
    has: function (key) {
        return this._map.has(key);
    },
    set: function (key, value) {
        return this._map.set(key, value);
    },
    size: function () {
        return this._map.size();
    }
};

globalContractStorage.prototype = {
    clear: function () {
        this._map.clear();
    },
    delete: function (key) {
        return this._map.delete(key);
    },
    forEach: function (callbackFn, thisArg) {
        return this._map.forEach(callbackFn, thisArg);
    },
    get: function (key) {
        return this._map.get(key);
    },
    has: function (key) {
        return this._map.has(key);
    },
    set: function (key, value) {
        return this._map.set(key, value);
    },
    size: function () {
        return this._map.size();
    }
};

module.exports = {
    LocalContractStorage: localContractStorage,
    GlobalContractStorage: globalContractStorage
};
