#! /usr/bin/env bash

echo "config file path:$config"

source $NEBULAS_SRC/setup.sh

export GO111MODULE=on

make clean && make build

command="./neb -c $config"
echo "Run $command"
eval $command
