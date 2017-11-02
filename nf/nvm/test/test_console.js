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
if (typeof console === 'undefined') {
    throw new Error("console is undefined.");
}

if (typeof console !== 'object') {
    throw new Error("console is not an object.");
}

['debug', 'warn', 'info', 'log', 'error'].forEach(function (val) {
    if (typeof console[val] !== 'function') {
        throw new Error("console." + val + " is not a function.");
    }

    var f = console[val];
    f("output:", "string", 123, {
        x: 4,
        y: 5
    }, [6, 7, 8, {
        z: 9
    }], console, f);
});
