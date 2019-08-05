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

# details for rocksdb install(offical):https://github.com/facebook/rocksdb/blob/master/INSTALL.md
OS="$(uname -s)"

DEST=$1
if [ $# -eq 0 ]; then 
  DEST="rocksdb"
fi

if [ "$OS" = "Darwin" ]; then
  LOGICAL_CPU=$(sysctl -n hw.ncpu)
  DYLIB="dylib"
else
  LOGICAL_CPU=$(cat /proc/cpuinfo |grep "processor"|wc -l)
  DYLIB="so"
fi

PARALLEL=$LOGICAL_CPU


prepare() {
  if [ "$OS" = "Darwin" ]; then
    if ! hash brew 2>/dev/null; then
      echo "install brew for macOS"
      /usr/bin/ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"
    fi
  fi

  if ! hash git 2>/dev/null; then
    case $OS in
      'Linux')
        sudo apt-get update
        sudo apt install -y git
        ;;
      'Darwin')
        brew install git
        ;;
      *) ;;
    esac
  fi
}

check_install() {
  if [ "$OS" = "Linux" ]; then
    result=`ldconfig -p | grep -c $1`
  else
    result=`brew list | grep -c $1`
  fi
  return `test $result -ne 0`
}

install_rocksdb() {
  echo "install rocksdb..."
  case $OS in
    'Linux')
      sudo apt-get install -y libsnappy-dev
      sudo apt-get install -y zlib1g-dev
      sudo apt-get install -y libbz2-dev
      sudo apt-get install -y liblz4-dev
      sudo apt-get install -y libzstd-dev
      
      if [ ! -d $DEST ]; then
        git clone https://github.com/facebook/rocksdb.git $DEST
      fi
      pushd $DEST
        git checkout v5.18.3
        sudo make install-shared -j$PARALLEL
      popd
      sudo ldconfig
      ;;
    'Darwin')
      xcode-select --install 2>/dev/null
      brew install rocksdb
      ;;
    *) ;;
  esac
}

uninstall_rocksdb() {
  echo "uninstall rocksdb..."
  case $OS in
    'Linux')
      rm -rf $DEST
      sudo rm -rf /usr/local/include/rocksdb
      sudo rm -rf /usr/local/lib/librocksdb*
      sudo ldconfig
      ;;
    'Darwin')
      brew uninstall rocksdb
      ;;
    *) ;;
  esac
}

check_install rocksdb
if [ $? -ne 0 ]; then
  prepare
  install_rocksdb
  if [ $? -eq 0 ]; then
    echo "rocksdb install success"
  fi
fi
