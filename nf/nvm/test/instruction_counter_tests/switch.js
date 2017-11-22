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
var gen1X = function (a) {
    var o = {};
    o.get = function () {
        return a;
    }
    return o;
}
var test1 = function (x) {
    var ret = 0;
    switch (x.get()) {
        case 1:
            ret = 1 + 2;
            break;
        case 2:
            ret = 2 * 2;
            break;
        case 3:
            ret = 3 * 3;
        case 4:
            ret += 4 * 4;
            break;
        case 100:
            return 100;
        case 101:
            return 3 * x.get();
        default:
            ret = x.get() * 2 + 13;
    }
    return ret;
};
assertEqual(test1, [gen1X(1)], 3, 4);
assertEqual(test1, [gen1X(2)], 4, 4);
assertEqual(test1, [gen1X(3)], 25, 6);
assertEqual(test1, [gen1X(4)], 16, 4);
assertEqual(test1, [gen1X(100)], 100, 2);
assertEqual(test1, [gen1X(101)], 303, 5);
assertEqual(test1, [gen1X(50)], 113, 7);

// test2.
var gen2X = function (a) {
    return [a];
}
var test2 = function (x) {
    var ret = 0;
    switch (x[0]) {
        case 1:
            ret = 1 + 2;
            break;
        case 2:
            ret = 2 * 2;
            break;
        case 3:
            ret = 3 * 3;
        case 4:
            ret += 4 * 4;
            break;
        case 100:
            return 100;
        case 101:
            return 3 * x[0];
        default:
            ret = x[0] * 2 + 13;
    }
    return ret;
};
assertEqual(test2, [gen2X(1)], 3, 3);
assertEqual(test2, [gen2X(2)], 4, 3);
assertEqual(test2, [gen2X(3)], 25, 5);
assertEqual(test2, [gen2X(4)], 16, 3);
assertEqual(test2, [gen2X(100)], 100, 1);
assertEqual(test2, [gen2X(101)], 303, 3);
assertEqual(test2, [gen2X(50)], 113, 5);
