#! /usr/bin/env bash

declare NEBULAS_SRC=${GOPATH}/src/github.com/nebulasio/go-nebulas

[ -z ${NEBULAS_BRANCH} ] && NEBULAS_BRANCH=master

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
        wget http://ory7cn4fx.bkt.clouddn.com/vendor.tar.gz
        tar -vxzf vendor.tar.gz
    else
        echo './nodep not found. Run dep...'
        make dep
    fi
fi
make deploy-v8
if [[ -x ${NEBULAS_SRC}/neb ]]; then
    echo './neb exists.'
else
    echo './neb not found. Building ./neb...'
    make clean && make build
fi

echo 'Run ./neb '$@
./neb $@
