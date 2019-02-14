#! /usr/bin/env bash

declare NEBULAS_SRC=${GOPATH}/src/github.com/nebulasio/go-nebulas
declare NEBULAS_BRANCH=master

# git clone https://github.com/nebulasio/go-nebulas ${NEBULAS_SRC}
echo NEBULAS_BRANCH=${NEBULAS_BRANCH}
cd ${NEBULAS_SRC} && \
    git checkout ${NEBULAS_BRANCH}

if [[ -d ${NEBULAS_SRC}/vendor ]]; then
    echo './vendor exists.'
else
    echo './vendor not found. Createing ./vendor...'
    if [[ -f ${NEBULAS_SRC}/nodep ]]; then
        echo './nodep exists. Downloading vendor...'
        wget http://develop-center.oss-cn-zhangjiakou.aliyuncs.com/setup/vendor.tar.gz
        tar -vxzf vendor.tar.gz
    else
        echo './nodep not found. Run dep...'
        make dep
    fi
fi

make deploy-v8
make clean && make build

echo 'Run ./neb '$@
./neb $@
