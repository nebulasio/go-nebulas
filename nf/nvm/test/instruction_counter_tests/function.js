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
const assert = require('assert.js');
const count_of_helper_statement = 42;

var count = 0;
var x = 1,
    y = 2,
    z = 3;

function test1(a, b) {
    return a + b;
};
count = _instruction_counter.count;
assert.equal(test1(x, y), z);
assert.equal(_instruction_counter.count - count - count_of_helper_statement, 3);

var test2 = function (a, b) {
    return a + b;
};
count = _instruction_counter.count;
assert.equal(test2(x, y), z);
assert.equal(_instruction_counter.count - count - count_of_helper_statement, 3);


var test3 = (a, b) => a + b;
count = _instruction_counter.count;
assert.equal(test3(x, y), z);
assert.equal(_instruction_counter.count - count - count_of_helper_statement, 3);

var test4 = (a, b) => {
    return a + b;
};
count = _instruction_counter.count;
assert.equal(test4(x, y), z);
assert.equal(_instruction_counter.count - count - count_of_helper_statement, 3);
