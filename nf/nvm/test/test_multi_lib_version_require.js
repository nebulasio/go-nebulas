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

'use strict';

var console2 = require('console.js');
if (!Object.is(console, console2)) {
    throw new Error("require should caches libs.");
}

var err = new Error("require should throw error when file does not exist.");
try {
    require("./not-exist-file");
    throw err;
} catch (e) {
    if (e === err) {
        throw e;
    }
}

try {
    require("/ab/c/console.js");
    throw err;
} catch (e) {
    if (e === err) {
        throw e;
    }
}
