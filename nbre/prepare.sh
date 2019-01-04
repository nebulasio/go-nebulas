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

CUR_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}"  )" >/dev/null && pwd  )"
#CUR_DIR="$( pwd )"
OS="$(uname -s)"

if [ "$OS" = "Darwin" ]; then
  LOGICAL_CPU=$(sysctl -n hw.ncpu)
else
  LOGICAL_CPU=$(cat /proc/cpuinfo |grep "processor"|wc -l)
fi

PARALLEL=$LOGICAL_CPU

if [ ! -d $CUR_DIR/3rd_party/cmake-3.12.2 ]; then
  cd $CUR_DIR/3rd_party/
  tar -xf cmake-3.12.2.tar.gz
fi

if [ ! -f $CUR_DIR/lib/bin/cmake ]; then
  cd $CUR_DIR/3rd_party/cmake-3.12.2/
  ./bootstrap --prefix=$CUR_DIR/lib --parallel=$PARALLEL && make -j$PARALLEL && make install
fi
export PATH=$CUR_DIR/lib/bin:$PATH

git submodule update --init

if ! hash autoreconf 2>/dev/null; then
  case $OS in
    'Linux')
      sudo apt-get install autoconf
      ;;
    'Darwin')
      brew install autoconf
      ;;
    *) ;;
  esac
fi

if ! hash libtool 2>/dev/null; then
  case $OS in
    'Linux')
      sudo apt-get install libtool-bin
      ;;
    'Darwin')
      brew install libtool
      ;;
    *) ;;
  esac
fi

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

if [ ! -d $CUR_DIR/lib_llvm/include/llvm ]; then
  ln -s $CUR_DIR/3rd_party/cfe-$LLVM_VERSION.src $CUR_DIR/3rd_party/llvm-$LLVM_VERSION.src/tools/clang
  ln -s $CUR_DIR/3rd_party/lld-$LLVM_VERSION.src $CUR_DIR/3rd_party/llvm-$LLVM_VERSION.src/tools/lld
  ln -s $CUR_DIR/3rd_party/clang-tools-extra-$LLVM_VERSION.src $CUR_DIR/3rd_party/llvm-$LLVM_VERSION.src/tools/clang/tools/extra
  ln -s $CUR_DIR/3rd_party/compiler-rt-$LLVM_VERSION.src $CUR_DIR/3rd_party/llvm-$LLVM_VERSION.src/projects/compiler-rt
  ln -s $CUR_DIR/3rd_party/libcxx-$LLVM_VERSION.src $CUR_DIR/3rd_party/llvm-$LLVM_VERSION.src/projects/libcxx
  ln -s $CUR_DIR/3rd_party/libcxxabi-$LLVM_VERSION.src $CUR_DIR/3rd_party/llvm-$LLVM_VERSION.src/projects/libcxxabi

  cd $CUR_DIR/3rd_party
  mkdir llvm-build
  cd llvm-build
  cmake -DLLVM_ENABLE_RTTI=ON -DLLVM_ENABLE_EH=ON -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX=$CUR_DIR/lib_llvm/ ../llvm-$LLVM_VERSION.src
  make CC=clang -j$PARALLEL && make install
fi

export CXX=$CUR_DIR/lib_llvm/bin/clang++
export CC=$CUR_DIR/lib_llvm/bin/clang

cd $CUR_DIR/3rd_party
if [ ! -d "boost_1_67_0"  ]; then
  tar -zxvf boost_1_67_0.tar.gz
fi
if [ ! -d $CUR_DIR/lib/include/boost ]; then
  cd boost_1_67_0
  ./bootstrap.sh --prefix=$CUR_DIR/lib/
  ./b2 --toolset=clang -j$PARALLEL
  ./b2 install
fi

#if [ -f $CUR_DIR/lib/include/boost/property_tree/detail/ptree_implementation.hpp ]; then
  #if [ ! -f $CUR_DIR/lib/include/boost/property_tree/detail/boost_ptree_rtti.patch ]; then
    #cp $CUR_DIR/3rd_party/boost_ptree_rtti.patch $CUR_DIR/lib/include/boost/property_tree/detail/.
    #cd $CUR_DIR/lib/include/boost/property_tree/detail/
    #patch -t -p1 < boost_ptree_rtti.patch
  #fi
#fi

build_with_cmake(){
  cd $CUR_DIR/3rd_party/$1
  build="build.tmp"
  if [ -d $build ]; then
    rm -rf $build
  fi
  mkdir $build
  cd $build

  params=("$@")
  flags=${params[@]:1}
  flagsStr=`echo $flags`
  cmake -DCMAKE_MODULE_PATH=$CUR_DIR/lib/lib/cmake -DCMAKE_LIBRARY_PATH=$CUR_DIR/lib/lib -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX=$CUR_DIR/lib/ -DRelease=1 ../ $flagsStr
  make -j$PARALLEL && make install && make clean
  cd ../ && rm -rf $build
}

check_install() {
  if [ "$OS" = "Linux" ]; then
    echo `ldconfig -p | grep -c $1`
    return
  fi
  echo 0
}

build_with_configure(){
  cd $CUR_DIR/3rd_party/$1
  ./configure --prefix=$CUR_DIR/lib/
  make -j$PARALLEL && make install && make clean
}

build_with_make(){
  cd $CUR_DIR/3rd_party/$1
  make -j$PARALLEL && make install PREFIX=$CUR_DIR/lib/
}


if [ "$OS" = "Darwin" ]; then
  if [ ! -d $CUR_DIR/3rd_party/gflags ]; then
    cd $CUR_DIR/3rd_party
    git clone -b v2.2.1 https://github.com/gflags/gflags.git
  fi
  if [ ! -d $CUR_DIR/lib/include/gflags/ ]; then
    build_with_cmake gflags
  fi
fi

if [ ! -d $CUR_DIR/lib/include/glog/ ]; then
  build_with_cmake glog
fi
if [ ! -d $CUR_DIR/lib/include/gtest/ ]; then
  build_with_cmake googletest
fi

if [ ! -d $CUR_DIR/lib/include/ff/ ]; then
  build_with_cmake functionflow
fi

if [ ! -f $CUR_DIR/lib/include/snappy.h ]; then
  #cd $CUR_DIR/3rd_party/snappy && cp ../snappy.patch ./ && git apply snappy.patch
  # cd $CUR_DIR/3rd_party/snappy
  # turn off unittest
  build_with_cmake snappy -DBUILD_SHARED_LIBS=true -DSNAPPY_BUILD_TESTS=false
fi

if [ ! -f $CUR_DIR/lib/include/zlib.h ]; then
  build_with_configure zlib
fi

if [ ! -f $CUR_DIR/lib/include/zstd.h ]; then
  if [ `check_install zstd` -eq 0 ]; then
    build_with_make zstd
  fi
fi

if [ ! -f $CUR_DIR/lib/include/bzlib.h ]; then
  cd $CUR_DIR/3rd_party/bzip2-1.0.6
  case $OS in
    'Linux')
      BZLib="so"
      ;;
    'Darwin')
      BZLib="dylib"
      ;;
    *) ;;
  esac
  cp -f ../Makefile-libbz2_$BZLib ./
  make -j$PARALLEL -f Makefile-libbz2_$BZLib && make -f Makefile-libbz2_$BZLib install PREFIX=$CUR_DIR/lib/ && make -f Makefile-libbz2_$BZLib clean
  rm -rf Makefile-libbz2_$BZLib
  git checkout .
fi

if [ ! -f $CUR_DIR/lib/include/lz4.h ]; then
  build_with_make lz4
fi

if [ ! -d $CUR_DIR/lib/include/rocksdb ]; then
  cd $CUR_DIR/3rd_party/rocksdb
  LIBRARY_PATH=$CUR_DIR/lib/lib CPATH=$CUR_DIR/lib/include LDFLAGS=-stdlib=libc++ make install-shared INSTALL_PATH=$CUR_DIR/lib -j$PARALLEL
fi

#if [ ! -d $CUR_DIR/lib/include/grpc ]; then
  #cd $CUR_DIR/3rd_party/grpc
  #git submodule update --init
  #make -j$PARALLEL && make install prefix=$CUR_DIR/lib/
#fi

if [ ! -f $CUR_DIR/lib/bin/protoc ]; then
  cd $CUR_DIR/3rd_party/protobuf
  ./autogen.sh
  ./configure --prefix=$CUR_DIR/lib/
  make -j$PARALLEL && make install && make clean
fi

if [ ! -e $CUR_DIR/lib/lib/libc++.so ]; then
  cp -f $CUR_DIR/lib_llvm/lib/libc++* $CUR_DIR/lib/lib/
fi

if [ ! -f $CUR_DIR/lib/include/softfloat.h ]; then
  cd $CUR_DIR/3rd_party/SoftFloat-3e/build/Linux-x86_64-GCC/
  make -j$PARALLEL
  cp libsoftfloat.so $CUR_DIR/lib/lib/
  cp ../../source/include/softfloat.h $CUR_DIR/lib/include/
  cp ../../source/include/softfloat_types.h $CUR_DIR/lib/include/
  make clean
fi

cd $CUR_DIR
