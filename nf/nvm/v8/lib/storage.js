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

var fieldNameRe = /^[a-zA-Z_].*/;

var ContractStorage = function (handler) {
    this.storage = new Storage(handler);
};

var StorageMap = function (storage, fieldName) {
    if (!storage instanceof ContractStorage) {
        throw new Error("StorageMap only accept instance of ContractStorage");
    }

    if (typeof fieldName !== "string" || fieldNameRe.exec(fieldName) == null) {
        throw new Error("StorageMap fieldName must starts with [a-zA-Z_]");
    }

    Object.defineProperty(this, "storage", {
        configurable: false,
        enumerable: false,
        get: function () {
            return storage;
        }
    });
    Object.defineProperty(this, "fieldName", {
        configurable: false,
        enumerable: false,
        get: function () {
            return fieldName;
        }
    });
};

var combineStorageMapKey = function (fieldName, key) {
    return "@" + fieldName + "[" + key + "]";
};

StorageMap.prototype = {
    delete: function (key) {
        return this.storage.del(combineStorageMapKey(this.fieldName, key));
    },
    get: function (key) {
        var val = this.storage.get(combineStorageMapKey(this.fieldName, key));
        // console.log('cs.get: key=' + key + '; val=' + val + ' (' + typeof val + ')');

        if (val != null) {
            val = JSON.parse(val);
        }
        return val;
    },
    set: function (key, value) {
        var val = JSON.stringify(value);
        // console.log("cs.set: key=" + key + "; val=" + value + " (" + typeof value + ") to " + val + " (" + typeof val + ")");
        return this.storage.put(combineStorageMapKey(this.fieldName, key), val);
    }
};
StorageMap.prototype.put = StorageMap.prototype.set;
StorageMap.prototype.del = StorageMap.prototype.delete;


ContractStorage.prototype = {
    delete: function (key) {
        return this.storage.del(key)
    },
    get: function (key) {
        var val = this.storage.get(key);
        // console.log('cs.get: key=' + key + '; val=' + val + ' (' + typeof val + ')');

        if (val != null) {
            val = JSON.parse(val);
        }
        return val;
    },
    set: function (key, value) {
        var val = JSON.stringify(value);
        // console.log("cs.set: key=" + key + "; val=" + value + " (" + typeof value + ") to " + val + " (" + typeof val + ")");
        return this.storage.put(key, val);
    },
    defineProperty: function (obj, fieldName) {
        if (!obj || !fieldName) {
            throw new Error("defineProperty requires two parameters.");
        }
        var $this = this;
        Object.defineProperty(obj, fieldName, {
            configurable: false,
            enumerable: true,
            get: function () {
                return $this.get(fieldName);
            },
            set: function (val) {
                return $this.set(fieldName, val);
            }
        });
    },
    defineProperties: function () {
        if (arguments.length < 2) {
            throw new Error("defineProperties requires more or equal to two parameters.");
        }

        var obj = arguments[0];
        for (var i = 1; i < arguments.length; i++) {
            this.defineProperty(obj, arguments[i]);
        }
    },
    defineMapProperty: function (obj, fieldName) {
        if (!obj || !fieldName) {
            throw new Error("defineMapProperty requires two parameters.");
        }

        var mapObj = new StorageMap(this, fieldName);
        Object.defineProperty(obj, fieldName, {
            configurable: false,
            enumerable: true,
            get: function () {
                return mapObj;
            }
        });
    },
    defineMapProperties: function () {
        if (arguments.length < 2) {
            throw new Error("defineMapProperties requires more or equal to two parameters.");
        }

        var obj = arguments[0];
        for (var i = 1; i < arguments.length; i++) {
            this.defineMapProperty(obj, arguments[i]);
        }
    }
};

ContractStorage.prototype.put = ContractStorage.prototype.set;
ContractStorage.prototype.del = ContractStorage.prototype.delete;

module.exports = {
    ContractStorage: ContractStorage,
    lcs: new ContractStorage(_storage_handlers.lcs),
    gcs: new ContractStorage(_storage_handlers.gcs)
};
