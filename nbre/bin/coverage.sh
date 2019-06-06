#!/bin/bash

# Copyright (C) 2017-2019 go-nebulas authors
#
# This file is part of the go-nebulas library.
#
# the go-nebulas library is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.

# the go-nebulas library is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with the go-nebulas library.  If not, see <http://www.gnu.org/licenses/>.

CUR_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}"  )" >/dev/null && pwd  )"

UNITTEST_BIN=$1
COVERAGE_DIR=$2
CMD=$3

rm -rf default.profraw default.profdata
$CUR_DIR/$UNITTEST_BIN
llvm-profdata merge -sparse default.profraw -o default.profdata
llvm-cov $CMD $CUR_DIR/$UNITTEST_BIN -instr-profile=default.profdata $COVERAGE_DIR -use-color | grep -v lib | grep -v proto
