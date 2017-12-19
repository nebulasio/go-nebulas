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
var test1 = function (v) {
    Event.Trigger("ERC20", v);
};

assertEqual(test1, ["123"], undefined, 42);
assertEqual(test1, ["1234"], undefined, 42 + 1);
assertEqual(test1, ["12345"], undefined, 42 + 1 * 2);
