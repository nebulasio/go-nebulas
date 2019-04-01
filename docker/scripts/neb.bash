#! /usr/bin/env bash

declare NEBULAS_SRC=${GOPATH}/src/github.com/nebulasio/go-nebulas

OS="$(uname -s)"
case $OS in
'Linux')
  DYLIB="so"
  ;;
'Darwin')
  DYLIB="dylib"
  ;;
*) ;;
esac

if [ "$REGION" = "China" ]; then
  SOURCE_URL="http://develop-center.oss-cn-zhangjiakou.aliyuncs.com"
else
  SOURCE_URL="https://s3-us-west-1.amazonaws.com/develop-center"
fi

echo "REGION is:$REGION"
echo "source url is:$SOURCE_URL"
echo "config file path:$config"

setup_with_vendor() {
    echo "check vendor..."
    if [[ -d ${NEBULAS_SRC}/vendor ]]; then
        echo './vendor exists.'
    else
        echo './vendor not found. Createing ./vendor...'
        if [[ "$TZ" = "Asia/Shanghai" ]]; then
            echo "downloading vendor from remote..."
            wget $SOURCE_URL/setup/vendor.tar.gz
            tar -vxzf vendor.tar.gz
        else
            echo 'Run dep...'
            make dep
        fi
    fi    
}

setup_with_nbre() {
    echo "check nbre..."
    mkdir -p ${NEBULAS_SRC}/nbre/lib/lib
    count=`ls ${NEBULAS_SRC}/nbre/lib/lib|grep -c $DYLIB`
    if [[ $count -gt 1 ]]; then
        echo './nbre/lib exists.'
    else
        echo './nbre/lib not found. Downlading ./nbre/lib...'
        pushd ${NEBULAS_SRC}/nbre/lib
        wget $SOURCE_URL/nbre/lib.tar.gz.$OS -O lib.tar.gz.$OS
        tar -vxzf lib.tar.gz.$OS
        popd
    fi 

    if [[ -f ${NEBULAS_SRC}/nbre/bin/nbre ]]; then
        echo './nbre/bin/nbre exists.'
    else
        echo './nbre/bin/nbre not found. Downlading ./nbre/bin/nbre...'
        mkdir -p ${NEBULAS_SRC}/nbre/bin
        pushd ${NEBULAS_SRC}/nbre/bin
        wget $SOURCE_URL/nbre/nbre.$OS -O nbre
        popd
    fi 
}

setup_with_vendor
setup_with_nbre

make clean && make build

command="./neb -c $config"
echo "Run $command"
eval $command
