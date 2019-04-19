#!/bin/bash

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

CUR_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}"  )" >/dev/null && pwd  )"
#CUR_DIR="$( pwd )"
OS="$(uname -s)"


if [ "$OS" = "Darwin" ]; then
  LOGICAL_CPU=$(sysctl -n hw.ncpu)
  DYLIB="dylib"
else
  LOGICAL_CPU=$(cat /proc/cpuinfo |grep "processor"|wc -l)
  DYLIB="so"
fi

PARALLEL=$LOGICAL_CPU
NEED_CHECK_INSTALL=false

mkdir -p $CUR_DIR/lib
git submodule update --init

install_system_tools() {
  if [ "$OS" = "Darwin" ]; then
    if ! hash brew 2>/dev/null; then
      echo "install brew for macOS"
      /usr/bin/ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"
    fi
  fi

  if ! hash unzip 2>/dev/null; then
    case $OS in
      'Linux')
        sudo apt-get install -y unzip
        ;;
      'Darwin')
        brew install unzip
        ;;
      *) ;;
    esac
  fi
  check_script_run unzip

  if ! hash autoreconf 2>/dev/null; then
    case $OS in
      'Linux')
        sudo apt-get install -y autoconf
        ;;
      'Darwin')
        brew install autoconf
        ;;
      *) ;;
    esac
  fi
  check_script_run autoconf

  if ! hash libtool 2>/dev/null; then
    case $OS in
      'Linux')
        sudo apt-get install -y libtool-bin
        ;;
      'Darwin')
        brew install libtool
        ;;
      *) ;;
    esac
  fi
  check_script_run libtool
}

check_script_run() {
  if [ $? -ne 0 ]; then
    echo "$1 install failed. Please check environment!"
    exit 1
  fi
}

check_install() {
  if $NEED_CHECK_INSTALL; then
    if [ "$OS" = "Linux" ]; then
      result=`ldconfig -p | grep -c $1`
    else
      result=`brew list | grep -c $1`
    fi
    return `test $result -ne 0`
  else
    return 1
  fi
}

build_with_cmake(){
  check_install $1
  if [ $? -eq 0 ]; then
    # has been installed in system, skip make and install
    echo "$1 has installed"
    return
  fi

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

build_with_configure(){
  check_install $1
  if [ $? -eq 0 ]; then
    # has been installed in system, skip make and install
    return
  fi

  cd $CUR_DIR/3rd_party/$1
  ./configure --prefix=$CUR_DIR/lib/
  make -j$PARALLEL && make install && make clean
}

build_with_make(){
  check_install $1
  if [ $? -eq 0 ]; then
    # has been installed in system, skip make and install
    return
  fi

  cd $CUR_DIR/3rd_party/$1
  make -j$PARALLEL && make install PREFIX=$CUR_DIR/lib/
  make clean
}

# V1 >= V2
function version_ge() { test "$(echo "$@" | tr " " "\n" | sort -rV | head -n 1)" == "$1"; }

check_install_cmake() {
  #check if cmake has been installed
  if hash cmake 2>/dev/null; then
    version=$(cmake --version|grep version|awk '{print $3}')
    #echo "check cmake installed $version"
    if version_ge $version "3.12.2"; then
      return
    fi
  fi

  if [ ! -d $CUR_DIR/3rd_party/cmake-3.12.2 ]; then
    cd $CUR_DIR/3rd_party/
    tar -xf cmake-3.12.2.tar.gz
  fi

  if [ ! -f $CUR_DIR/lib/bin/cmake ]; then
    cd $CUR_DIR/3rd_party/cmake-3.12.2/
    ./bootstrap --prefix=$CUR_DIR/lib --parallel=$PARALLEL && make -j$PARALLEL && make install
  fi

  check_script_run cmake
  export PATH=$CUR_DIR/lib/bin:$PATH
}

check_install_llvm() {
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

  if [ ! -d $CUR_DIR/3rd_party/llvm-$LLVM_VERSION.src/tools/clang ]; then
    ln -s $CUR_DIR/3rd_party/cfe-$LLVM_VERSION.src $CUR_DIR/3rd_party/llvm-$LLVM_VERSION.src/tools/clang
    ln -s $CUR_DIR/3rd_party/lld-$LLVM_VERSION.src $CUR_DIR/3rd_party/llvm-$LLVM_VERSION.src/tools/lld
    ln -s $CUR_DIR/3rd_party/clang-tools-extra-$LLVM_VERSION.src $CUR_DIR/3rd_party/llvm-$LLVM_VERSION.src/tools/clang/tools/extra
    ln -s $CUR_DIR/3rd_party/compiler-rt-$LLVM_VERSION.src $CUR_DIR/3rd_party/llvm-$LLVM_VERSION.src/projects/compiler-rt
    ln -s $CUR_DIR/3rd_party/libcxx-$LLVM_VERSION.src $CUR_DIR/3rd_party/llvm-$LLVM_VERSION.src/projects/libcxx
    ln -s $CUR_DIR/3rd_party/libcxxabi-$LLVM_VERSION.src $CUR_DIR/3rd_party/llvm-$LLVM_VERSION.src/projects/libcxxabi
  fi

  cd $CUR_DIR/3rd_party
  if [ ! -f $CUR_DIR/lib_llvm/bin/clang ]; then
    mkdir llvm-build
    cd llvm-build
    cmake -DCLANG_ENABLE_STATIC_ANALYZER=OFF -DCLANG_ENABLE_ARCMT=OFF -DLLVM_TARGETS_TO_BUILD="X86;XCore" -DLLVM_DISTRIBUTION_COMPONENTS="LLVMDemangle;LLVMBinaryFormat;LLVMAsmParser;LLVMBitReader;LLVMAnalysis;LLVMBitWriter;LLVMProfileData;LLVMTarget;LLVMTransformUtils;LLVMScalarOpts;LLVMBinaryFormat;LLVMDebugInfoCodeView;LLVMDebugInfoMSF;LLVMObject;LLVMInstCombine;LLVMExecutionEngine;LLVMRuntimeDyld;LLVMCore;clang;LLVMMC;LLVMMCJIT;LLVMSupport;LLVMInterpreter;LLVMCodeGen;LLVMIRReader;LLVMOrcJIT"  -DCMAKE_CXX_COMPILER=g++ -DLLVM_ENABLE_RTTI=ON -DLLVM_ENABLE_EH=ON -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX=$CUR_DIR/lib_llvm/ ../llvm-$LLVM_VERSION.src
    make -j$PARALLEL && make install
    cd ..
  fi

  if [ ! -e $CUR_DIR/lib/lib/libc++.$DYLIB ]; then
    mkdir -p $CUR_DIR/lib/lib
    cp -Rf $CUR_DIR/lib_llvm/lib/libc++* $CUR_DIR/lib/lib/
  fi

  check_script_run llvm

  export LD_LIBRARY_PATH=$CUR_DIR/lib/lib:$CUR_DIR/lib_llvm/lib:$LD_LIBRARY_PATH
  export PATH=$CUR_DIR/lib_llvm/bin:$PATH
  export CXX=$CUR_DIR/lib_llvm/bin/clang++
  export CC=$CUR_DIR/lib_llvm/bin/clang
}

check_install_boost() {
  cd $CUR_DIR/3rd_party
  if [ ! -d "boost_1_67_0"  ]; then
    tar -zxvf boost_1_67_0.tar.gz
  fi
  if [ ! -e $CUR_DIR/lib/lib/libboost_system.$DYLIB ]; then
    cd boost_1_67_0
    ./bootstrap.sh --with-toolset=clang --prefix=$CUR_DIR/lib/
    ./b2 clean
    ./b2 toolset=clang --with-date_time --with-graph --with-program_options --with-filesystem --with-system --with-thread -j$PARALLEL
    ./b2 install toolset=clang --with-date_time --with-graph --with-program_options --with-filesystem --with-system --with-thread --prefix=$CUR_DIR/lib/
  fi
  check_script_run boost
}

check_install_gflags() {
  if [ ! -d $CUR_DIR/3rd_party/gflags ]; then
    cd $CUR_DIR/3rd_party
    git clone -b v2.2.1 https://github.com/gflags/gflags.git
  fi

  if [ ! -e $CUR_DIR/lib/lib/libgflags.$DYLIB ]; then
    #cp $CUR_DIR/3rd_party/build_option_bak/CMakeLists.txt-gflags $CUR_DIR/3rd_party/gflags/CMakeLists.txt
    build_with_cmake gflags -DGFLAGS_NAMESPACE=google -DCMAKE_CXX_FLAGS=-fPIC -DBUILD_SHARED_LIBS=true
  fi
  check_script_run gflags
}

check_install_glog() {
  if [ ! -e $CUR_DIR/lib/lib/libglog.$DYLIB ]; then
    #cp $CUR_DIR/3rd_party/build_option_bak/CMakeLists.txt-glog $CUR_DIR/3rd_party/glog/CMakeLists.txt
    build_with_cmake glog -DGFLAGS_NAMESPACE=google -DCMAKE_CXX_FLAGS=-fPIC -DBUILD_SHARED_LIBS=true
  fi
  check_script_run glog
}

check_install_gtest() {
  if [ ! -e $CUR_DIR/lib/lib/libgtest.$DYLIB ]; then
    #cp $CUR_DIR/3rd_party/build_option_bak/CMakeList.txt-googletest $CUR_DIR/3rd_party/googletest/CMakeLists.txt
    build_with_cmake googletest -DGFLAGS_NAMESPACE=google -DCMAKE_CXX_FLAGS=-fPIC -DBUILD_SHARED_LIBS=true
  fi
  check_script_run gtest
}

check_install_ff() {
  if [ ! -e $CUR_DIR/lib/lib/libff_functionflow.$DYLIB ]; then
    build_with_cmake fflib
  fi
  check_script_run ff
}

check_install_snappy() {
  if [ ! -e $CUR_DIR/lib/lib/libsnappy.$DYLIB ]; then
    #cd $CUR_DIR/3rd_party/snappy && cp ../snappy.patch ./ && git apply snappy.patch
    # cd $CUR_DIR/3rd_party/snappy
    # turn off unittest
    build_with_cmake snappy -DBUILD_SHARED_LIBS=true -DSNAPPY_BUILD_TESTS=false
  fi
  check_script_run snappy
}

check_install_zlib() {
  if [ ! -e $CUR_DIR/lib/lib/libz.$DYLIB ]; then
    build_with_configure zlib
    cd $CUR_DIR/3rd_party/zlib
    git checkout .
  fi
  check_script_run zlib
}

check_install_zstd() {
  if [ ! -e $CUR_DIR/lib/lib/libzstd.$DYLIB ]; then
    build_with_make zstd
  fi
  check_script_run zstd
}

check_install_bzlib() {
  if [ ! -e $CUR_DIR/lib/lib/libbz2.$DYLIB ]; then
    cd $CUR_DIR/3rd_party/bzip2-1.0.6
    cp -f ../Makefile-libbz2_$DYLIB ./
    make -j$PARALLEL -f Makefile-libbz2_$DYLIB && make -f Makefile-libbz2_$DYLIB install PREFIX=$CUR_DIR/lib/ && make -f Makefile-libbz2_$DYLIB clean
    rm -rf Makefile-libbz2_$DYLIB
  fi
  check_script_run bzlib
}

check_install_lz4() {
  if [ ! -e $CUR_DIR/lib/lib/liblz4.$DYLIB ]; then
    build_with_make lz4
  fi
  check_script_run lz4
}

check_install_rocksdb() {
  #check if rocksdb has been installed in system
  check_install rocksdb
  if [ $? -eq 0 ]; then
    return
  fi

  check_install_snappy
  check_install_zlib
  # check_install_zstd
  # check_install_bzlib
  check_install_lz4

  if [ ! -e $CUR_DIR/lib/lib/librocksdb.$DYLIB ]; then
    # cp $CUR_DIR/3rd_party/build_option_bak/CMakeLists.txt-rocksdb $CUR_DIR/3rd_party/rocksdb/CMakeLists.txt
    # cp $CUR_DIR/3rd_party/build_option_bak/Makefile-rocksdb $CUR_DIR/3rd_party/rocksdb/Makefile

    cd $CUR_DIR/3rd_party/rocksdb
    #export CXX=$CUR_DIR/lib_llvm/bin/clang++
    ROCKSDB_DISABLE_GFLAGS=On LIBRARY_PATH=$CUR_DIR/lib/lib CPATH=$CUR_DIR/lib/include make install-shared INSTALL_PATH=$CUR_DIR/lib -j$PARALLEL
    #ROCKSDB_DISABLE_GFLAGS=On LIBRARY_PATH=$CUR_DIR/lib/lib CPATH=$CUR_DIR/lib/include CXXFLAGS=-stdlib=libc++ LDFLAGS=-lc++ make install-shared INSTALL_PATH=$CUR_DIR/lib -j$PARALLEL
    make clean
  fi
  check_script_run rocksdb
}

check_install_protoc() {
  if [ ! -f $CUR_DIR/lib/bin/protoc ]; then
    cd $CUR_DIR/3rd_party/protobuf
    ./autogen.sh
    ./configure --prefix=$CUR_DIR/lib/
    make -j$PARALLEL && make install && make clean
  fi
  check_script_run protoc
}

check_install_softfloat() {
  if [ ! -e $CUR_DIR/lib/lib/libsoftfloat.$DYLIB ]; then
    cd $CUR_DIR/3rd_party/SoftFloat-3e/build/Linux-x86_64-GCC/
    make -j$PARALLEL
    cp libsoftfloat.so $CUR_DIR/lib/lib/libsoftfloat.$DYLIB
    cp ../../source/include/softfloat.h $CUR_DIR/lib/include/
    cp ../../source/include/softfloat_types.h $CUR_DIR/lib/include/
    make clean
  fi
  check_script_run softfloat
}

check_install_sha() {
  if [ ! -e $CUR_DIR/3rd_party/cryptopp/sha3.h ]; then
    cd $CUR_DIR/3rd_party/cryptopp
    unzip cryptopp810.zip
  fi
  check_script_run sha
}

check_install_crypto() {
  if [ ! -e $CUR_DIR/lib/lib/libcryptopp.$DYLIB ]; then
    cd $CUR_DIR/3rd_party/cryptopp
    export CXX=$CUR_DIR/lib_llvm/bin/clang++
    make dynamic -j$PARALLEL && make install PREFIX=$CUR_DIR/lib/
    make clean
  fi
  check_script_run cryptopp
}

check_install_gperftools() {
  if [ ! -e $CUR_DIR/lib/lib/libprofiler.$DYLIB ]; then
    cd $CUR_DIR/3rd_party/gperftools
    ./autogen.sh
    build_with_configure gperftools
  fi
  check_script_run gperftools
}

install_system_tools
check_install_cmake
check_install_llvm
check_install_boost
check_install_gflags
check_install_glog
check_install_gtest
check_install_ff
check_install_bzlib
check_install_zstd
check_install_rocksdb
check_install_protoc
check_install_softfloat
check_install_sha
check_install_crypto
check_install_gperftools

cd $CUR_DIR
