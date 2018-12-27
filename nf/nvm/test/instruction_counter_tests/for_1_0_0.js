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
    var ret = 0;
    for (var i = 0; i < x; i++)
        ret += i;
    return ret;
};
assertEqual(test1, [0], 0, 3);
assertEqual(test1, [10], 45, 10 * 9 + 3);
assertEqual(test1, [11], 55, 11 * 9 + 3);
assertEqual(test1, [23], 253, 23 * 9 + 3);

// test2.
var test2 = function (x) {
    var ret = 0;
    for (var i = 0; i < x && i < 20; ret++, i++) {
        ret += i;
    }
    return ret;
}
assertEqual(test2, [10], 55, 10 * (18) + 9);
assertEqual(test2, [11], 66, 11 * (18) + 9);
assertEqual(test2, [20], 210, 20 * (18) + 9);
assertEqual(test2, [23], 210, 20 * (18) + 9);

// test3.
var test3 = function (x) {
    var ret = 0;
    for (var i = 0; i < x; ret++, i++) {
        ret += i;
        if (i % 2 == 0) {
            continue;
        }
        ret += 1;
    }
    return ret;
}
assertEqual(test3, [1], 1, 1 * (18) + 3);
assertEqual(test3, [10], 60, 10 * (18) + 3 + 10 / 2 * 3);
assertEqual(test3, [11], 71, 11 * (18) + 3 + Math.floor(11 / 2) * 3);
assertEqual(test3, [12], 84, 12 * (18) + 3 + Math.floor(12 / 2) * 3);

// test4.
var gen4X = function (a) {
    var x = new Array();
    for (var i = 0; i < a; i++) {
        x.push(i + 1);
    }
    return x;
};
var test4 = function (x) {
    var ret = 0;
    for (var i in x) {
        ret += x[i];
    }
    return ret;
};
assertEqual(test4, [gen4X(0)], 0, 0);
assertEqual(test4, [gen4X(1)], 1, 8 * 1);
assertEqual(test4, [gen4X(2)], 3, 8 * 2);
assertEqual(test4, [gen4X(3)], 6, 8 * 3);

// test5.
var gen5X = function (a) {
    var x = new Object();
    for (var i = 0; i < a; i++) {
        x[i] = i + 1;
    }
    return x;
}
var test5 = function (x) {
    var ret = 0;
    for (var i in x)
        ret += x[i];
    return ret;
};
assertEqual(test5, [gen5X(0)], 0, 0);
assertEqual(test5, [gen5X(1)], 1, 8 * 1);
assertEqual(test5, [gen5X(2)], 3, 8 * 2);
assertEqual(test5, [gen5X(10)], 55, 8 * 10);

// test6.
var gen6X = function (a) {
    var x = new Map();
    for (var i = 0; i < a; i++) {
        x.set(i, i + 1);
    }
    return x;
};
var test6 = function (x) {
    var ret = 0;
    for (var i of x.values())
        ret += i;
    return ret;
};
assertEqual(test6, [gen6X(0)], 0, 12);
assertEqual(test6, [gen6X(1)], 1, 4 * 1 + 12);
assertEqual(test6, [gen6X(2)], 3, 4 * 2 + 12);
assertEqual(test6, [gen6X(10)], 55, 4 * 10 + 12);
