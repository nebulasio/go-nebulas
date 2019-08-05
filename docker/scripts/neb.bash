#! /usr/bin/env bash

declare NEBULAS_SRC=${GOPATH}/src/github.com/nebulasio/go-nebulas

echo "config file path:$config"

source $NEBULAS_SRC/setup.sh

make clean && make build

command="./neb -c $config"
echo "Run $command"
eval $command
