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
const lcs = LocalContractStorage;

function assertEqual(func, args, expected, expected_count, msg) {
    const count_of_helper_statement = 46;
    var count = _instruction_counter.count;
    assert.equal(func.apply(null, args), expected);
    assert.equal(_instruction_counter.count - count - count_of_helper_statement, expected_count, msg);
};

// test1.
var test1 = function (k, v) {
    lcs.set(k, v);
    return lcs.get(k);
};

assertEqual(test1, ["k", "1"], "1", 28);
assertEqual(test1, ["k", "12"], "12", 28 + 1);
assertEqual(test1, ["k", "123"], "123", 28 + 2);
assertEqual(test1, ["k1", "1"], "1", 28 + 1);
assertEqual(test1, ["k12", "1"], "1", 28 + 2);

// test2.
var test2 = function (k) {
    lcs.del(k);
};

assertEqual(test2, ["k"], undefined, 12);
assertEqual(test2, ["k1"], undefined, 12);
assertEqual(test2, ["k12"], undefined, 12);
