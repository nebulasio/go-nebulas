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
        if [[ "$REGION" = "China" ]]; then
            echo "downloading vendor from remote..."
            wget $SOURCE_URL/setup/vendor/vendor.tar.gz
            tar -vxzf vendor.tar.gz
        else
            echo 'Run dep...'
            go get -u github.com/golang/dep/cmd/dep
            make dep
        fi
    fi    
}

setup_with_vendor
source $NEBULAS_SRC/install-native-libs.sh

make clean && make build

command="./neb -c $config"
echo "Run $command"
eval $command
