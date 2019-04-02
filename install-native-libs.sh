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

#!/bin/bash

# usage: source native-libs.sh

CUR_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}"  )" >/dev/null && pwd  )"
#CUR_DIR="$( pwd )"
OS="$(uname -s)"
#default region is China
REGION="China"

if [ "$OS" = "Darwin" ]; then
  LOGICAL_CPU=$(sysctl -n hw.ncpu)
  DYLIB="dylib"
else
  LOGICAL_CPU=$(cat /proc/cpuinfo |grep "processor"|wc -l)
  DYLIB="so"
fi
PARALLEL=$LOGICAL_CPU

if [ "$REGION" = "China" ]; then
  SOURCE_URL="http://develop-center.oss-cn-zhangjiakou.aliyuncs.com"
else
  SOURCE_URL="https://s3-us-west-1.amazonaws.com/develop-center"
fi

rm -rf $CUR_DIR/native-lib/*
mkdir -p $CUR_DIR/native-lib

install_nvm() {
  nvm_lib=$CUR_DIR/nf/nvm/native-lib
  if [ ! -d $nvm_lib ]; then
    mkdir -p $nvm_lib
    pushd $nvm_lib
    wget $SOURCE_URL/setup/nvm/lib_nvm_$OS.tar.gz -O lib_nvm_$OS.tar.gz
    tar -zxvf lib_nvm_$OS.tar.gz
    cp -R lib_nvm_$OS/* $nvm_lib/
    rm -rf lib_nvm_$OS
    rm -rf lib_nvm_$OS.tar.gz
    popd
  fi
  libs=`ls $nvm_lib|grep .$DYLIB`
  for lib in $libs; do
    ln -s $nvm_lib/$lib  $CUR_DIR/native-lib/$lib
  done
}

install_nbre() {
  nbre_lib=$CUR_DIR/nbre/lib
  mkdir -p $nbre_lib
  if [ ! -d $nbre_lib/lib ]; then
    pushd $nbre_lib
    wget $SOURCE_URL/setup/nbre/lib_nbre_$OS.tar.gz -O lib_nbre_$OS.tar.gz
    tar -zxvf lib_nbre_$OS.tar.gz
    cp -R lib_nbre_$OS/* $nbre_lib/
    mv -f $nbre_lib/bin $CUR_DIR/nbre/
    rm -rf lib_nbre_$OS
    rm -rf lib_nbre_$OS.tar.gz
    popd
  fi
  libs=`ls $nbre_lib/lib|grep .$DYLIB`
  for lib in $libs; do
    ln -s $nbre_lib/lib/$lib  $CUR_DIR/native-lib/$lib
  done
}

export_libs() {
  case $OS in
    'Linux')
      export LD_LIBRARY_PATH=$CUR_DIR/native-lib:$LD_LIBRARY_PATH
      ;;
    'Darwin')
      ln -fs $CUR_DIR/native-lib ~/lib
      ;;
    *) ;;
  esac
}

install_nvm
install_nbre
export_libs
    