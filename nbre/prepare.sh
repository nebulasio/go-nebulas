#!/bin/bash

CUR_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}"  )" >/dev/null && pwd  )"
cd $CUR_DIR/3rd_party
if [ ! -d "boost_1_67_0"  ]; then
  tar -zxvf boost_1_67_0.tar.gz
fi
cd boost_1_67_0
./bootstrap.sh --prefix=$CUR_DIR/lib/ --toolset=clang
./b2 install


cd $CUR_DIR/3rd_party/glog
if [ -d "build" ]; then
  rm -rf build
fi
mkdir build
cd build
cmake -DCMAKE_INSTALL_PREFIX=$CUR_DIR/lib/ ../
make && make install && make clean
cd ../ && rm -rf build

cd $CUR_DIR/3rd_party/gtest
if [ -d "build" ]; then
  rm -rf build
fi
mkdir build
cd build
cmake -DCMAKE_INSTALL_PREFIX=$CUR_DIR/lib/ ../
make && make install && make clean
cd ../ && rm -rf build

