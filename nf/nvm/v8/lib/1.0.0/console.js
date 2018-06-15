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

function Console() {}

function log(...args) {
    var level = args.shift();
    if (typeof (level) != 'number') {
        throw 'level must be number.';
    }

    var msg = '';
    for (var i = 0; i < args.length - 1; i++) {
        msg += format(args[i]) + ' ';
    }
    msg += format(args[args.length - 1]);

    _native_log(level, msg);
}

function format(obj) {
    if (typeof (obj) == 'object') {
        return JSON.stringify(obj);
    }
    return obj;
}

[
    ['debug', 1],
    ['warn', 2],
    ['info', 3],
    ['log', 3],
    ['error', 4]
].forEach(function (val) {
    Console.prototype[val[0]] = log.bind(null, val[1]);
});

module.exports = new Console();
module.exports.Console = Console;
