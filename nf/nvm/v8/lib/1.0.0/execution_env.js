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

Function.prototype.toString = function(){return "";};

const require = (function (global) {
    var PathRegexForNotLibFile = /^\.{0,2}\//;
    var modules = new Map();

    var Module = function (id, parent) {
        this.exports = {};
        Object.defineProperty(this, "id", {
            enumerable: false,
            configurable: false,
            writable: false,
            value: id
        });

        if (parent && !(parent instanceof Module)) {
            throw new Error("parent parameter of Module construction must be instance of Module or null.");
        }
    };

    Module.prototype = {
        _load: function () {
            var $this = this,
                native_req_func = _native_require(this.id),
                temp_global = Object.create(global);
            native_req_func.call(temp_global, this.exports, this, curry(require_func, $this));
        },
        _resolve: function (id) {
            var paths = this.id.split("/");
            paths.pop();

            if (!PathRegexForNotLibFile.test(id)) {
                id = "lib/" + id;
                paths = [];
            }

            for (const p of id.split("/")) {
                if (p == "" || p == ".") {
                    continue;
                } else if (p == ".." && paths.length > 0) {
                    paths.pop();
                } else {
                    paths.push(p);
                }
            }

            if (paths.length > 0 && paths[0] == "") {
                paths.shift();
            }

            return paths.join("/");
        },
    };

    var globalModule = new Module("main.js");
    modules.set(globalModule.id, globalModule);

    function require_func(parent, id) {
        id = parent._resolve(id);
        var module = modules.get(id);
        if (!module || !(module instanceof Module)) {
            module = new Module(id, parent);
            module._load();
            modules.set(id, module);
        }
        return module.exports;
    };

    function curry(uncurried) {
        var parameters = Array.prototype.slice.call(arguments, 1);
        var f = function () {
            return uncurried.apply(this, parameters.concat(
                Array.prototype.slice.call(arguments, 0)
            ));
        };
        Object.defineProperty(f, "main", {
            enumerable: true,
            configurable: false,
            writable: false,
            value: globalModule,
        });
        return f;
    };

    return curry(require_func, globalModule);
})(this);

const GlobalVars = {};

const console = require('console.js');
const ContractStorage = require('storage.js');
const LocalContractStorage = ContractStorage.lcs;
const GlobalContractStorage = ContractStorage.gcs;
const BigNumber = require('bignumber.js');
const Blockchain = require('blockchain.js');
GlobalVars.Blockchain = Blockchain;
const Event = require('event.js');

var Date = require('date.js');
Math.random = require('random.js');