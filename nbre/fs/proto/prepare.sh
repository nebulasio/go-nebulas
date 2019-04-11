# Copyright (C) 2017 go-nebulas authors
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

#!/bin/bash

CUR_DIR="$( pwd )"
#CUR_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}"  )" >/dev/null && pwd  )"
#cd $CUR_DIR/fs/proto
cp ../../../core/pb/block.proto ./
cp ../../../core/pb/dynasty.proto ./
cp ../../../core/pb/genesis.proto ./
cp ../../../consensus/pb/state.proto ./
cp ../../../common/dag/pb/dag.proto ./
cp ../../../common/trie/pb/trie.proto ./

patch < block.proto.patch

#unzip protoc-3.2.0-linux-x86_64.zip -d protoc3

protoc --cpp_out=./ block.proto dag.proto dynasty.proto genesis.proto state.proto ir.proto trie.proto
