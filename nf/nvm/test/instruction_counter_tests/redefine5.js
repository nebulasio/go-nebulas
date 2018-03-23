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

var [_instruction_counter] = [{
    incr: function () {},
    storIncr: function () {},
    eventIncr: function () {},
    count: 0,
}];

(function () {
    _instruction_counter.incr(1);
    if (_instruction_counter.count == 0) {
        console.log('WARNING: succeed bypass the instruction counter.');
    } else {
        throw new Error("still not break the jail of _instruction_counter.");
    }
})();
