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

if [ ! -d $CUR_DIR/3rd_party/cmake-3.12.2 ]; then
  tar -xf cmake-3.12.2.tar.gz
fi

if [ ! -f $CUR_DIR/lib/bin/cmake ]; then
  cd $CUR_DIR/3rd_party/cmake-3.12.2/
  ./bootstrap --prefix=$CUR_DIR/lib --parallel=4 && make && make install
fi
export PATH=$CUR_DIR/lib/bin:$PATH

cd $CUR_DIR/3rd_party
LLVM_VERSION=6.0.1
unzip_llvm_tar(){
  if [ ! -d $1-$LLVM_VERSION.src ]; then
    tar -xf $1-$LLVM_VERSION.src.tar.xz
  fi
}
unzip_llvm_tar llvm
unzip_llvm_tar cfe
unzip_llvm_tar clang-tools-extra
unzip_llvm_tar compiler-rt
unzip_llvm_tar libcxx
unzip_llvm_tar libcxxabi
unzip_llvm_tar libunwind
unzip_llvm_tar lld

if [ ! -d $CUR_DIR/lib/include/llvm ]; then
  ln -s $CUR_DIR/3rd_party/cfe-$LLVM_VERSION.src $CUR_DIR/3rd_party/llvm-$LLVM_VERSION.src/tools/clang
  ln -s $CUR_DIR/3rd_party/lld-$LLVM_VERSION.src $CUR_DIR/3rd_party/llvm-$LLVM_VERSION.src/tools/lld
  ln -s $CUR_DIR/3rd_party/clang-tools-extra-$LLVM_VERSION.src $CUR_DIR/3rd_party/llvm-$LLVM_VERSION.src/tools/clang/tools/extra
  ln -s $CUR_DIR/3rd_party/compiler-rt-$LLVM_VERSION.src $CUR_DIR/3rd_party/llvm-$LLVM_VERSION.src/projects/compiler-rt
  ln -s $CUR_DIR/3rd_party/libcxx-$LLVM_VERSION.src $CUR_DIR/3rd_party/llvm-$LLVM_VERSION.src/projects/libcxx
  ln -s $CUR_DIR/3rd_party/libcxxabi-$LLVM_VERSION.src $CUR_DIR/3rd_party/llvm-$LLVM_VERSION.src/projects/libcxxabi

  cd $CUR_DIR/3rd_party
  mkdir llvm-build
  cd llvm-build
  cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX=$CUR_DIR/lib/ ../llvm-$LLVM_VERSION.src
  make -j4 && make install
fi

export CXX=$CUR_DIR/lib/bin/clang++
export CC=$CUR_DIR/lib/bin/clang

cd $CUR_DIR/3rd_party
if [ ! -d "boost_1_67_0"  ]; then
  tar -zxvf boost_1_67_0.tar.gz
fi
if [ ! -d $CUR_DIR/lib/include/boost ]; then
  cd boost_1_67_0
  ./bootstrap.sh --prefix=$CUR_DIR/lib/
  ./b2 --toolset=clang
  ./b2 install
fi

build_with_cmake(){
  cd $CUR_DIR/3rd_party/$1
  if [ -d "build" ]; then
    rm -rf build
  fi
  mkdir build
  cd build
  cmake -DCMAKE_MODULE_PATH=$CUR_DIR/lib/lib/cmake -DCMAKE_LIBRARY_PATH=$CUR_DIR/lib/lib -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX=$CUR_DIR/lib/ ../
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

if [ ! -d $CUR_DIR/lib/include/glog/ ]; then
  build_with_cmake glog
fi
if [ ! -d $CUR_DIR/lib/include/gtest/ ]; then
  build_with_cmake googletest
fi

if [ ! -f $CUR_DIR/lib/include/snappy.h ]; then
  cd $CUR_DIR/3rd_patch/snappy && cp ../snappy.patch ./ && git apply snappy.patch
  build_with_cmake snappy
fi

if [ ! -d $CUR_DIR/lib/include/gflags ]; then
  build_with_configure gflags
fi

if [ ! -f $CUR_DIR/lib/include/zlib.h ]; then
  build_with_configure zlib
fi

cd $CUR_DIR/3rd_party
if [ ! -d "zstd-1.1.3"  ]; then
  tar -zxvf zstd-1.1.3.tar.gz
fi
if [ ! -f $CUR_DIR/lib/include/zstd.h ]; then
  build_with_make zstd-1.1.3
fi
if [ ! -f $CUR_DIR/lib/include/bzlib.h ]; then
build_with_make bzip2-1.0.6
fi
if [ ! -f $CUR_DIR/lib/include/lz4.h ]; then
  build_with_make lz4
fi

if [ ! -d $CUR_DIR/lib/include/rocksdb ]; then
  cd $CUR_DIR/3rd_party/rocksdb
  LIBRARY_PATH=$CUR_DIR/lib/lib CPATH=$CUR_DIR/lib/include make install-static INSTALL_PATH=$CUR_DIR/lib -j4
fi

if [ ! -d $CUR_DIR/3rd_party/grpc ]; then
  cd $CUR_DIR/3rd_party
  git clone -b $(curl -L https://grpc.io/release) https://github.com/grpc/grpc
  cd grpc
  git submodule update --init
fi

if [ ! -d $CUR_DIR/lib/include/grpc ]; then
  cd $CUR_DIR/3rd_party/grpc
  make && make install prefix=$CUR_DIR/lib/
fi

if [ ! -d $CUR_DIR/test/data/data.db ]; then
  cd $CUR_DIR/test/data
  tar -xf data.db.tar.gz
fi
