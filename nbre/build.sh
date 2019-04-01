# Copyright (C) 2019 go-nebulas authors
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

CUR_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}"  )" >/dev/null && pwd  )"
OS="$(uname -s)"

if [ "$OS" = "Darwin" ]; then
  LOGICAL_CPU=$(sysctl -n hw.ncpu)
else
  LOGICAL_CPU=$(cat /proc/cpuinfo |grep "processor"|wc -l)
fi
PARALLEL=$LOGICAL_CPU

build_with_cmake() {
	source $CUR_DIR/env.set.sh
	mkdir -p $CUR_DIR/build
	pushd $CUR_DIR/build
	if [ "$1" = "debug" ]; then 
		cmake ..
	else
		cmake -DRelease=1 ..
	fi
	make -j$PARALLEL && make install
	popd
}

clean() {
	rm -rf $CUR_DIR/build
}


if [ "$1" = "debug" ]; then
	build_with_cmake debug
elif [ "$1" = "clear" ]; then
	clean
	build_with_cmake
else
	build_with_cmake
fi
