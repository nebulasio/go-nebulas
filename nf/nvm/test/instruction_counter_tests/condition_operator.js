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

function assertEqual(func, args, expected, expected_count, msg) {
    const count_of_helper_statement = 9;
    var count = _instruction_counter.count;
    assert.equal(func.apply(null, args), expected);
    assert.equal(_instruction_counter.count - count - count_of_helper_statement, expected_count, msg);
};

// test1.
var test1 = function (x) {
    return x > 10 ? x - 1 : x + 1;
};
assertEqual(test1, [1], 2, 2);
assertEqual(test1, [11], 10, 2);

// test2.
var test2 = function (x) {
    return x > 10 && x < 20 ? x * 8 - 1 : x - 3;
};
assertEqual(test2, [15], 119, 5);
assertEqual(test2, [10], 7, 4);

// test3.
var test3 = function (x) {
    return x > 10 ? x < 20 ? x * 8 - 1 : x - 2 : x * 2 + 3;
};
assertEqual(test3, [10], 23, 3);
assertEqual(test3, [15], 119, 4);
assertEqual(test3, [20], 18, 3);
