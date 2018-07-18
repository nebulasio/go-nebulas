# go-nebulas

Official Go implementation of the Nebulas protocol.

[![Build Status](https://travis-ci.org/nebulasio/go-nebulas.svg?branch=master)](https://travis-ci.org/nebulasio/go-nebulas)

For the roadmap of Nebulas, please visit the [roadmap](https://github.com/nebulasio/wiki/blob/master/roadmap.md) page.

For more information of Nebulas protocol, design documents, please refer to our [wiki](https://github.com/nebulasio/wiki).

TestNet is released, please check [here](https://github.com/nebulasio/wiki/blob/master/testnet.md) for more details.

Mainnet is released, please check [here](https://github.com/nebulasio/wiki/blob/master/mainnet.md) for more details.

## Building from source

### Prerequisites

| Components | Version | Description |
|----------|-------------|-------------|
|[Golang](https://golang.org) | >= 1.9.2| The Go Programming Language |
[Dep](https://github.com/golang/dep) | >= 0.3.1 | Dep is a dependency management tool for Go. |

### Build

1. Checkout repo.

```bash
cd $GOPATH/src
go get -u -v github.com/nebulasio/go-nebulas
```

The project is under active development. Its default branch is _develop_.

```bash
cd github.com/nebulasio/go-nebulas
```

New users may want to checkout and use the stable mainnet release in __master__.

```bash
git checkout master
```

Or use the stable testnet release in __testnet__.

```bash
git checkout testnet
```

2. Install rocksdb.

* **OS X**:
    * Install via [homebrew](http://brew.sh/).
    * run `brew install rocksdb`

* **Linux - Ubuntu**
    * Install Dependencies
        ```bash
        apt-get update
        apt-get -y install build-essential libgflags-dev libsnappy-dev zlib1g-dev libbz2-dev liblz4-dev libzstd-dev
        ```
    * Install rocksdb by source code:
        ```bash
        git clone https://github.com/facebook/rocksdb.git
        cd rocksdb & make shared_lib && make install-shared
        ```
* **Linux - Centos**
    * Install Dependencies
        ```bash
        yum -y install epel-release && yum -y update
        yum -y install gflags-devel snappy-devel zlib-devel bzip2-devel gcc-c++  libstdc++-devel
        ```
    * Install rocksdb by source code:
        ```bash
        git clone https://github.com/facebook/rocksdb.git
        cd rocksdb && make shared_lib && make install-shared
        ```

3. Install dependencies packages.
    * all golang dependencies will be stored in ./vendor.
    * run `make dep` to install dependencies.
    * If you failed to run this, please download our [vendor.tar.gz](http://ory7cn4fx.bkt.clouddn.com/vendor.tar.gz) directly.

4. Install v8 libraries.
    * run `make deploy-v8`

5. Build the neb binary.
    * run `make build`

## Run

### Run seed node
Starting a Nebulas seed node is simple. After the build step above, run a command:

```bash
./neb [-c /path/to/config.conf]
```

> tips: more details about configuration, please refer to [`template.conf`](https://github.com/nebulasio/wiki/blob/master/resources/conf/template.conf)

You will see log message output like:

```
INFO[2018-03-30T01:39:16+08:00] Setuped Neblet.                               file=neblet.go func="neblet.(*Neblet).Setup" line=161
INFO[2018-03-30T01:39:16+08:00] Starting Neblet...                            file=neblet.go func="neblet.(*Neblet).Start" line=183
INFO[2018-03-30T01:39:16+08:00] Starting NebService...                        file=net_service.go func="net.(*NebService).Start" line=58
INFO[2018-03-30T01:39:16+08:00] Starting NebService Dispatcher...             file=dispatcher.go func="net.(*Dispatcher).Start" line=85
INFO[2018-03-30T01:39:16+08:00] Starting NebService Node...                   file=node.go func="net.(*Node).Start" line=96
INFO[2018-03-30T01:39:16+08:00] Starting NebService StreamManager...          file=stream_manager.go func="net.(*StreamManager).Start" line=74
INFO[2018-03-30T01:39:16+08:00] Started NewService Dispatcher.                file=dispatcher.go func="net.(*Dispatcher).loop" line=93
INFO[2018-03-30T01:39:16+08:00] Starting NebService RouteTable Sync...        file=route_table.go func="net.(*RouteTable).Start" line=91
INFO[2018-03-30T01:39:16+08:00] Started NebService StreamManager.             file=stream_manager.go func="net.(*StreamManager).loop" line=146
INFO[2018-03-30T01:39:16+08:00] Started NebService Node.                      file=net_service.go func="net.(*NebService).Start" id=QmP7HDFcYmJL12Ez4ZNVCKjKedfE7f48f1LAkUc3Whz4jP line=65 listening address="[/ip4/127.0.0.1/tcp/8680 /ip4/127.94.0.1/tcp/8680 /ip4/127.94.0.2/tcp/8680 /ip4/192.168.1.13/tcp/8680]"
INFO[2018-03-30T01:39:16+08:00] Started NebService.                           file=net_service.go func="net.(*NebService).Start" line=74
INFO[2018-03-30T01:39:16+08:00] Starting RPC GRPCServer...                    file=server.go func="rpc.(*Server).Start" line=87
INFO[2018-03-30T01:39:16+08:00] Started RPC GRPCServer.                       address="0.0.0.0:8684" file=server.go func="rpc.(*Server).Start" line=95
INFO[2018-03-30T01:39:16+08:00] Started NebService RouteTable Sync.           file=route_table.go func="net.(*RouteTable).syncLoop" line=123
INFO[2018-03-30T01:39:16+08:00] Starting RPC Gateway GRPCServer...            file=neblet.go func="neblet.(*Neblet).Start" http-cors="[]" http-server="[0.0.0.0:8685]" line=212 rpc-server="0.0.0.0:8684"
INFO[2018-03-30T01:39:16+08:00] Starting BlockChain...                        file=blockchain.go func="core.(*BlockChain).Start" line=194
INFO[2018-03-30T01:39:16+08:00] Starting BlockPool...                         file=neblet.go func="neblet.(*Neblet).Start" line=219 size=128
INFO[2018-03-30T01:39:16+08:00] Starting TransactionPool...                   file=neblet.go func="neblet.(*Neblet).Start" line=220 size=327680
INFO[2018-03-30T01:39:16+08:00] Started BlockChain.                           file=blockchain.go func="core.(*BlockChain).loop" line=208
INFO[2018-03-30T01:39:16+08:00] Starting EventEmitter...                      file=neblet.go func="neblet.(*Neblet).Start" line=221 size=40960
INFO[2018-03-30T01:39:16+08:00] Started BlockPool.                            file=block_pool.go func="core.(*BlockPool).loop" line=232
INFO[2018-03-30T01:39:16+08:00] Started TransactionPool.                      file=asm_amd64.s func=runtime.goexit line=2362 size=327680
INFO[2018-03-30T01:39:16+08:00] Started EventEmitter.                         file=event.go func="core.(*EventEmitter).loop" line=156
INFO[2018-03-30T01:39:16+08:00] Starting Dpos Mining...                       file=dpos.go func="dpos.(*Dpos).Start" line=136
INFO[2018-03-30T01:39:16+08:00] Started Sync Service.                         file=sync_service.go func="sync.(*Service).startLoop" line=150
INFO[2018-03-30T01:39:16+08:00] Started Dpos Mining.                          file=dpos.go func="dpos.(*Dpos).blockLoop" line=619
INFO[2018-03-30T01:39:16+08:00] Enabled Dpos Mining...                        file=dpos.go func="dpos.(*Dpos).EnableMining" line=155
INFO[2018-03-30T01:39:16+08:00] This is a seed node.                          file=neblet.go func="neblet.(*Neblet).Start" line=247
INFO[2018-03-30T01:39:16+08:00] Resumed Dpos Mining.                          file=dpos.go func="dpos.(*Dpos).ResumeMining" line=296
INFO[2018-03-30T01:39:16+08:00] Started Neblet.                               file=neblet.go func="neblet.(*Neblet).Start" line=259
```

From the log, we can see the binary execution starts neblet, starts network service, starts RPC API server, and starts consensus state machine.

## Docker

### Build
* pull from dockerhub directly
    ```bash
    docker pull bkbabydp/go-nebulas
    ```
* build locally
    ```bash
    docker-compose up
    ```

### Run
* edit [your conf path] in docker-compose.yml
```yml
  # node:
  #   image: bkbabydp/go-nebulas
  #   build:
  #     context: ./docker
  #   ports:
  #     - '8680'
  #     - '8684'
  #     - '8685'
  #     - '8888'
  #     - '8086'
  #   volumes:
  #     - .:/go/src/github.com/nebulasio/go-nebulas
  #   environment:
  #     - TZ=Asia/Shanghai
  #     - NEBULAS_BRANCH=master
  #   command: bash docker/scripts/neb.bash -c [your conf path]
```
* start the node
```bash
    cd /path/to/go-nebulas
    docker-compose up node
```

## TestNet

We are glad to release Nebulas Testnet here. 
You can use and join our [TestNet](https://github.com/nebulasio/wiki/blob/master/testnet.md) right now.

## MaintNet

We are glad to release Nebulas Mainnet here. 
You can use and join our [MainNet](https://github.com/nebulasio/wiki/blob/master/mainnet.md) right now.

## Explorer

Nebulas provides a block explorer to view block/transaction information.
Please check [Explorer](https://explorer.nebulas.io/#/).

## Wallet

Nebulas provides a web wallet to send transaction and deploy/call contract.
Please check [Web-Wallet](https://github.com/nebulasio/web-wallet)

## Wiki

Please check our [Wiki](https://github.com/nebulasio/wiki) to learn more about Nebulas.

## Contribution

We are very glad that you are considering to help Nebulas Team or go-nebulas project, including but not limited to source code, documents or others.

If you'd like to contribute, please fork, fix, commit and send a pull request for the maintainers to review and merge into the main code base. If you wish to submit more complex changes though, please check up with the core devs first on our [slack channel](http://nebulasio.herokuapp.com) to ensure those changes are in line with the general philosophy of the project and/or get some early feedback which can make both your efforts much lighter as well as our review and merge procedures quick and simple.

Please refer to our [contribution guideline](https://github.com/nebulasio/wiki/blob/master/contribute.md) for more information.

Thanks.

## License

The go-nebulas project is licensed under the [GNU Lesser General Public License Version 3.0 (“LGPL v3”)](https://www.gnu.org/licenses/lgpl-3.0.en.html).

For the more information about licensing, please refer to [Licensing](https://github.com/nebulasio/wiki/blob/master/licensing.md) page.

