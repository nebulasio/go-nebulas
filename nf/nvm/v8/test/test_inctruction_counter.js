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

_instruction_counter.incr(1);
console.log('count = ' + _instruction_counter.count);
if (_instruction_counter.count != 1) throw new Error('_instruction_counter.count error, expected ' + 1 + ', actual is ' + _instruction_counter.count);

_instruction_counter.incr(2);
console.log('count = ' + _instruction_counter.count);
if (_instruction_counter.count != 3) throw new Error('_instruction_counter.count error, expected ' + 3 + ', actual is ' + _instruction_counter.count);

_instruction_counter.incr(3);
console.log('count = ' + _instruction_counter.count);
if (_instruction_counter.count != 6) throw new Error('_instruction_counter.count error, expected ' + 6 + ', actual is ' + _instruction_counter.count);


_instruction_counter.count = 0123;
if (_instruction_counter.count != 6) throw new Error('_instruction_counter.count error, expected ' + 6 + ', actual is ' + _instruction_counter.count);

_instruction_counter.incr(4);
console.log('count = ' + _instruction_counter.count);
if (_instruction_counter.count != 10) throw new Error('_instruction_counter.count error, expected ' + 10 + ', actual is ' + _instruction_counter.count);

delete _instruction_counter.count;
if (_instruction_counter.count != 10) throw new Error('_instruction_counter.count error, expected ' + 10 + ', actual is ' + _instruction_counter.count);

_instruction_counter.incr(5);
console.log('count = ' + _instruction_counter.count);
if (_instruction_counter.count != 15) throw new Error('_instruction_counter.count error, expected ' + 15 + ', actual is ' + _instruction_counter.count);

_instruction_counter.incr(-1);
console.log('count = ' + _instruction_counter.count);
if (_instruction_counter.count != 15) throw new Error('_instruction_counter.count error, expected ' + 15 + ', actual is ' + _instruction_counter.count);
