#!/bin/bash
# Copyright (C) 2017-2020 go-nebulas authors
#
# This file is part of the go-nebulas library.
#
# the go-nebulas library is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# the go-nebulas library is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with the go-nebulas library.  If not, see <http://www.gnu.org/licenses/>.
#

CUR_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}"  )" >/dev/null && pwd  )"
#CUR_DIR="$( pwd )"
OS="$(uname -s)"

if [ "$OS" = "Darwin" ]; then
  DYLIB="dylib"
else
  DYLIB="so"
fi

help() {
    echo "USAGE: $0 mainnet|testnet|[-c path/to/config]"
    echo " e.g.: $0 mainnet"
    echo " e.g.: $0 testnet"
    echo " e.g.: $0 -c path/config.conf"
    exit 1
}

if [ $# = 0 ] ; then
    help
elif [ $# = 1 ] ; then
    case $1 in
        mainnet)
            conf=mainnet/conf/config.conf
            ;;
        testnet)
            conf=testnet/conf/config.conf
            ;;
        *)
            help
            ;;
    esac
elif [ $# = 2 ] ; then
    case $1 in
        -c)
            conf=$2
            ;;
        *)
            help
            ;;
    esac
else
    help      
fi

check_install() {
  if [ "$OS" = "Linux" ]; then
    result=`ldconfig -p | grep -c $1`
  else
    result=`brew list | grep -c $1`
  fi
  return `test $result -ne 0`
}

check() {
    check_install rocksdb
    if [ $? -ne 0 ]; then
        echo "rocksdb not installed. run `source setup.sh` first!"
        exit 1
    fi
    nvm_lib=$CUR_DIR/nf/nvm/native-lib
    if [ ! -f $nvm_lib/libnebulasv8.$DYLIB ]; then
        echo "nvm not installed. run `source setup.sh` first!"
        exit 1
    fi
}

start() {
    chainid=`cat $conf|grep "chain_id"|awk -F":" '{print $2}'`
    case `echo $chainid` in
        1)
            env=mainnet
            ;;
        1001)
            env=testnet
            ;;
        *)
            env=private
            ;;
    esac

    echo "================$env=================="
    echo "env: $env"
    echo "config: $conf"
    keydir=`cat $conf|grep "keydir:"|awk -F":" '{print $2}'`
    echo "keydir:$keydir"
    datadir=`cat $conf|grep "datadir:"|awk -F":" '{print $2}'`
    echo "datadir:$datadir"

    mint=`cat $conf|grep "start_mine"|awk -F":" '{print $2}'`
    mint=`echo $mint | tr "[A-Z]" "[a-z]"`
    echo "start_mine: $mint"
    if [ "$mint" = "true" ] ; then
        coinbase=`cat $conf|grep "coinbase"|awk -F":" '{print $2}'`
        echo "coinbase:$coinbase"
        miner=`cat $conf|grep "miner:"|awk -F":" '{print $2}'`
        echo "miner:$miner"
        echo -e '***Notice***:\nMake sure the coinbase and miner address is correct, \nthe miner keystore file is in `keydir` folder, \nand the passphrase is set correct in the configuration file!'
    else
        echo 'Join the PoD mint need register at https://node.nebulas.io, and set the `start_mine` to `true` in config file.'
    fi
    echo "======================================"

    if [ "$env" != "private" ]; then
        datadir=`echo $datadir | sed 's/\"//g'`
        if [ ! -d `echo $datadir` ]; then
            echo "    "
            echo "***Snapshot Tips***:"
            echo "Since Nebulas is running there for certain period of time, it will take quite some time to sync all the data from scratch."
            echo "you can download the package directly by following either link below:"
            echo "    "
            echo "http://develop-center.oss-cn-zhangjiakou.aliyuncs.com/data/$env/data.db.tar"
            echo "https://develop-center.s3-us-west-1.amazonaws.com/data/$env/data.db.tar"
            echo "    "
            echo 'After downloading, the package needs to be decompressed and placed in the location specified by `datadir` in the configuration file.'
        fi
    fi

    # echo "Check the configuration. Enter if you want to continue[y/n]:"
    read -p "Check the configuration. Enter if you want to continue[y/n]:" command
    command=`echo $command | tr "[A-Z]" "[a-z]"`
    case $command in
        y|yes)
            ./neb -c $conf
            ;;
        *)
            exit 0
            ;;
    esac
}

main() {
    check
    start
}

main
