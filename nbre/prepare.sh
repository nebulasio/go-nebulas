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

#CUR_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}"  )" >/dev/null && pwd  )"
CUR_DIR="$( pwd )"
cd $CUR_DIR/3rd_party
if [ ! -d "boost_1_67_0"  ]; then
  tar -zxvf boost_1_67_0.tar.gz
fi
cd boost_1_67_0
./bootstrap.sh --prefix=$CUR_DIR/lib/ --toolset=clang
./b2 install

build_with_cmake(){
  cd $CUR_DIR/3rd_party/$1
  if [ -d "build" ]; then
    rm -rf build
  fi
  mkdir build
  cd build
  cmake -DCMAKE_LIBRARY_PATH=$CUR_DIR/lib/lib -DCPATH=$CUR_DIR/lib/include -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX=$CUR_DIR/lib/ ../
  make && make install && make clean
  cd ../ && rm -rf build
}

build_with_configure(){
  cd $CUR_DIR/3rd_party/$1
  ./configure --prefix=$CUR_DIR/lib/
  make && make install && make clean
}

build_with_make(){
  cd $CUR_DIR/3rd_party/$1
  make && make install PREFIX=$CUR_DIR/lib/
}

build_with_cmake glog
build_with_cmake gtest
cd $CUR_DIR/3rd_patch/snappy && cp ../snappy.patch ./ && git apply snappy.patch
build_with_cmake snappy

build_with_configure gflags
build_with_configure zlib

cd $CUR_DIR/3rd_party
if [ ! -d "zstd-1.1.3"  ]; then
  tar -zxvf zstd-1.1.3.tar.gz
fi
build_with_make zstd-1.1.3
build_with_make bzip2-1.0.6
build_with_make lz4

cd $CUR_DIR/3rd_party/rocksdb
LIBRARY_PATH=$CUR_DIR/lib/lib CPATH=$CUR_DIR/lib/include make static_lib shared_lib -j4
make install INSTALL_PATH=$CUR_DIR/lib
