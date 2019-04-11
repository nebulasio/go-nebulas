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
    const count_of_helper_statement = 46;
    var count = _instruction_counter.count;
    assert.equal(func.apply(null, args), expected);
    assert.equal(_instruction_counter.count - count - count_of_helper_statement, expected_count, msg);
};

// test1.
var test1 = function (x) {
    var ret = 1;
    var i = x;
    while (ret < 1024 && i > 0) {
        ret *= 2;
        i--;
    }
    return ret;
};
assertEqual(test1, [0], 1, 9);
assertEqual(test1, [2], 4, 16 * 2 + 9);
assertEqual(test1, [10], 1024, 16 * 10 + 9);
assertEqual(test1, [11], 1024, 16 * 10 + 9);

// test2.
var test2 = function (x) {
    var ret = 1;
    var i = x;
    do {
        ret *= 2;
        i--;
    } while (ret < 1024 && i > 0);
    return ret;
}
assertEqual(test2, [0], 2, 16);
assertEqual(test2, [2], 4, 16 * 2);
assertEqual(test2, [10], 1024, 16 * 10);
assertEqual(test2, [11], 1024, 16 * 10);

// test3.
var test3 = function (x) {
    var ret = 1;
    while (ret < 1024) ret *= 2;
    return ret;
};
assertEqual(test3, [1], 1024, 73);
assertEqual(test3, [2], 1024, 73);
assertEqual(test3, [10], 1024, 73);
