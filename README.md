# go-nebulas

Official Go implementation of the Nebulas protocol.

[![Build Status](https://travis-ci.org/nebulasio/go-nebulas.svg?branch=master)](https://travis-ci.org/nebulasio/go-nebulas)

For the roadmap of Nebulas, please visit the [Roadmap](https://github.com/nebulasio/wiki/blob/master/roadmap.md) page.

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

New users may want to checkout and use the stable _master_ release.

```bash
git checkout master
```

2. Install rocksdb dependencies.

* **OS X**:
    * Install latest C++ compiler that supports C++ 11:
        * Update XCode:  run `xcode-select --install` (or install it from XCode App's settting).
        * Install via [homebrew](http://brew.sh/).
            * If you're first time developer in MacOS, you still need to run: `xcode-select --install` in your command line.
            * run `brew tap homebrew/versions; brew install gcc48 --use-llvm` to install gcc 4.8 (or higher).
    * run `brew install rocksdb`

* **Linux - Ubuntu**
    * Upgrade your gcc to version at least 4.8 to get C++11 support.
    * Install gflags. First, try: `sudo apt-get install libgflags-dev`
      If this doesn't work and you're using Ubuntu, here's a nice tutorial:
      (http://askubuntu.com/questions/312173/installing-gflags-12-04)
    * Install snappy. This is usually as easy as:
      `sudo apt-get install libsnappy-dev`.
    * Install zlib. Try: `sudo apt-get install zlib1g-dev`.
    * Install bzip2: `sudo apt-get install libbz2-dev`.
    * Install lz4: `sudo apt-get install liblz4-dev`.
    * Install zstandard: `sudo apt-get install libzstd-dev`.
3. Install dependencies packages.

```bash
make dep
```

4. Install dependent v8 libraries.

```bash
make deploy-v8
```

5. Build the neb binary.

```bash
make build
```

## Run

### Run seed node
Starting a Nebulas seed node is simple. After the build step above, run a command:

```bash
./neb
```

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

`docker build -t nebulas .`

### Run

`docker run -it -v $(pwd)/data.db:/nebulas/data.db nebulas`

### Configurations
Neb uses [Protocol Buffer](https://github.com/google/protobuf) to load configurations. The default config file is named as config.conf and looks like following:

```protobuf
# Neb configuration text file. Scheme is defined in neblet/pb/config.proto:Config.
#

network {
  listen: ["0.0.0.0:8680"]
  private_key: "conf/network/ed25519key"
}

chain {
  chain_id: 100
  datadir: "data.db"
  keydir: "keydir"
  genesis: "conf/default/genesis.conf"

  start_mine: true
  miner: "n1SAQy3ix1pZj8MPzNeVqpAmu1nCVqb5w8c"
  passphrase: "passphrase"
  coinbase: "n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE"
  
  signature_ciphers: ["ECC_SECP256K1"]
}

rpc {
    rpc_listen: ["0.0.0.0:8684"]
    http_listen: ["0.0.0.0:8685"]
    http_module: ["api", "admin"]
    
    # HTTP CORS allowed origins
    # http_cors: []
}

app {
    log_level: "debug"
    log_file: "logs"
    enable_crash_report: false
    crash_report_url: "https://crashreport.nebulas.io"
    pprof:{
        http_listen: "0.0.0.0:8888"
    }    
}

stats {
    enable_metrics: false
    influxdb: {
        host: "http://localhost:8086"
        db: "nebulas"
        user: "admin"
        password: "admin"
    }
}
```

The configuration schema is defined in proto _neblet/pb/config.proto:Config_. To load a different config file when starting the neb binary, use flag -c. For example,

```bash
./neb -c <path>/config.conf
```

Neb supports loading KeyStore file in Ethereum format. KeyStore files from config _key_dir_ are loaded during neb bootstrap. Example testing KeyStore looks like

```json
{"address":"n1SAQy3ix1pZj8MPzNeVqpAmu1nCVqb5w8c","crypto":{"cipher":"aes-128-ctr","ciphertext":"40701b061f1f6d3935dc43c2c06c7ed619c3b85f5ad4934fc440e1d61e878333","cipherparams":{"iv":"5e2ec4e7a241f2a086754df398373605"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":4096,"p":1,"r":8,"salt":"af935393fdde22073b99cd23898ba3f681b53272fe662f71d1220736ef517a1b"},"mac":"be6ba64359a617fbbd55558dd2dd412b1f5205fe1f130186f0155d1645313ee6","machash":"sha3256"},"id":"fcadcf90-858c-46f0-88dd-0fa4b5d98f51","version":3}

```

### Run node
Now we can get the seed address from the seed node log output above. Get pretty address and id from log starts with **"Started NetService Node"**. The seed address from log above is

```
INFO[2018-03-30T01:39:16+08:00] Started NebService Node.                      file=net_service.go func="net.(*NebService).Start" id=QmP7HDFcYmJL12Ez4ZNVCKjKedfE7f48f1LAkUc3Whz4jP line=65 listening address="[/ip4/127.0.0.1/tcp/8680 /ip4/127.94.0.1/tcp/8680 /ip4/127.94.0.2/tcp/8680 /ip4/192.168.1.13/tcp/8680]"
```
To start a node on another machine, we need to update p2p seed configuration in _config.conf_:

```protobuf
p2p {
  seed: "/ip4/127.0.0.1/tcp/8680/ipfs/QmP7HDFcYmJL12Ez4ZNVCKjKedfE7f48f1LAkUc3Whz4jP"
}
...
```

Note, if the node is running on the same machine, we need to use a different config file, and also assign different P2P and RPC server ports.

Now we can start the nodes by simply run command

```bash
./neb -c conf/example/miner.conf
```

Then it will connect to the seed node started earlier to join that network. The log output will look like:

```

INFO[2018-03-30T01:46:17+08:00] Setuped Neblet.                               file=neblet.go func="neblet.(*Neblet).Setup" line=161
INFO[2018-03-30T01:46:17+08:00] Starting Neblet...                            file=neblet.go func="neblet.(*Neblet).Start" line=183
INFO[2018-03-30T01:46:17+08:00] Starting NebService...                        file=net_service.go func="net.(*NebService).Start" line=58
INFO[2018-03-30T01:46:17+08:00] Starting NebService Dispatcher...             file=dispatcher.go func="net.(*Dispatcher).Start" line=85
INFO[2018-03-30T01:46:17+08:00] Starting NebService Node...                   file=node.go func="net.(*Node).Start" line=96
INFO[2018-03-30T01:46:17+08:00] Starting NebService StreamManager...          file=stream_manager.go func="net.(*StreamManager).Start" line=74
INFO[2018-03-30T01:46:17+08:00] Started NewService Dispatcher.                file=dispatcher.go func="net.(*Dispatcher).loop" line=93
INFO[2018-03-30T01:46:17+08:00] Starting NebService RouteTable Sync...        file=route_table.go func="net.(*RouteTable).Start" line=91
INFO[2018-03-30T01:46:17+08:00] Started NebService StreamManager.             file=stream_manager.go func="net.(*StreamManager).loop" line=146
INFO[2018-03-30T01:46:17+08:00] Started NebService Node.                      file=net_service.go func="net.(*NebService).Start" id=QmeEftJEYZwkUogRqCg2pbNQLkyqGEcgNcJEiArtcxkNyL line=65 listening address="[/ip4/127.0.0.1/tcp/8780 /ip4/127.94.0.1/tcp/8780 /ip4/127.94.0.2/tcp/8780 /ip4/192.168.1.13/tcp/8780]"
INFO[2018-03-30T01:46:17+08:00] Started NebService.                           file=net_service.go func="net.(*NebService).Start" line=74
INFO[2018-03-30T01:46:17+08:00] Started NebService RouteTable Sync.           file=route_table.go func="net.(*RouteTable).syncLoop" line=123
INFO[2018-03-30T01:46:17+08:00] Starting RPC GRPCServer...                    file=server.go func="rpc.(*Server).Start" line=87
INFO[2018-03-30T01:46:17+08:00] Started RPC GRPCServer.                       address="127.0.0.1:8784" file=server.go func="rpc.(*Server).Start" line=95
INFO[2018-03-30T01:46:17+08:00] Starting RPC Gateway GRPCServer...            file=neblet.go func="neblet.(*Neblet).Start" http-cors="[]" http-server="[127.0.0.1:8785]" line=212 rpc-server="127.0.0.1:8784"
INFO[2018-03-30T01:46:17+08:00] Starting BlockChain...                        file=blockchain.go func="core.(*BlockChain).Start" line=194
INFO[2018-03-30T01:46:17+08:00] Starting BlockPool...                         file=neblet.go func="neblet.(*Neblet).Start" line=219 size=128
INFO[2018-03-30T01:46:17+08:00] Started BlockChain.                           file=blockchain.go func="core.(*BlockChain).loop" line=208
INFO[2018-03-30T01:46:17+08:00] Started BlockPool.                            file=block_pool.go func="core.(*BlockPool).loop" line=232
INFO[2018-03-30T01:46:17+08:00] Starting TransactionPool...                   file=neblet.go func="neblet.(*Neblet).Start" line=220 size=327680
INFO[2018-03-30T01:46:17+08:00] Starting EventEmitter...                      file=neblet.go func="neblet.(*Neblet).Start" line=221 size=40960
INFO[2018-03-30T01:46:17+08:00] Started TransactionPool.                      file=asm_amd64.s func=runtime.goexit line=2362 size=327680
INFO[2018-03-30T01:46:17+08:00] Started EventEmitter.                         file=event.go func="core.(*EventEmitter).loop" line=156
INFO[2018-03-30T01:46:17+08:00] Starting Dpos Mining...                       file=dpos.go func="dpos.(*Dpos).Start" line=136
INFO[2018-03-30T01:46:17+08:00] Started Sync Service.                         file=sync_service.go func="sync.(*Service).startLoop" line=150
INFO[2018-03-30T01:46:17+08:00] Started Dpos Mining.                          file=dpos.go func="dpos.(*Dpos).blockLoop" line=619
INFO[2018-03-30T01:46:17+08:00] Enabled Dpos Mining...                        file=dpos.go func="dpos.(*Dpos).EnableMining" line=155
INFO[2018-03-30T01:46:17+08:00] Started Active Sync Task.                     file=blockchain.go func="core.(*BlockChain).StartActiveSync" line=524 syncpoint="{\"height\": 1, \"hash\": \"0000000000000000000000000000000000000000000000000000000000000000\", \"parent_hash\": \"0000000000000000000000000000000000000000000000000000000000000000\", \"acc_root\": \"db2a692aa8e21ba3a65fb952f441c5b346db29b3d4d10a7530b024e0ffc27050\", \"timestamp\": 0, \"tx\": 0, \"miner\": \"\"}"
INFO[2018-03-30T01:46:17+08:00] Suspended Dpos Mining.                        file=dpos.go func="dpos.(*Dpos).SuspendMining" line=290
INFO[2018-03-30T01:46:17+08:00] Started Neblet.                               file=neblet.go func="neblet.(*Neblet).Start" line=259
INFO[2018-03-30T01:46:17+08:00] Active Sync Task Finished.                    file=blockchain.go func="core.(*BlockChain).StartActiveSync.func1" line=527 tail="{\"height\": 1, \"hash\": \"0000000000000000000000000000000000000000000000000000000000000000\", \"parent_hash\": \"0000000000000000000000000000000000000000000000000000000000000000\", \"acc_root\": \"db2a692aa8e21ba3a65fb952f441c5b346db29b3d4d10a7530b024e0ffc27050\", \"timestamp\": 0, \"tx\": 0, \"miner\": \"\"}"
INFO[2018-03-30T01:46:17+08:00] Resumed Dpos Mining.                          file=dpos.go func="dpos.(*Dpos).ResumeMining" line=296
...
```

## REPL console
Nebulas provide an interactive javascript console, which can invoke all API and management RPC methods. Some management methods may require passphrase. Start console using the command:

```bash
./neb console
```

We have API and admin two schemes to access the console cmds. Users can quickly enter instructions using the TAB key.

```javascript
> api.
api.call                    api.getBlockByHash          api.getNebState             api.subscribe
api.estimateGas             api.getBlockByHeight        api.getTransactionReceipt
api.gasPrice                api.getDynasty              api.latestIrreversibleBlock
api.getAccountState         api.getEventsByHash         api.sendRawTransaction

> admin.
admin.accounts                      admin.nodeInfo                      admin.signHash
admin.getConfig                     admin.sendTransaction               admin.signTransactionWithPassphrase
admin.lockAccount                   admin.sendTransactionWithPassphrase admin.startPprof
admin.newAccount                    admin.setHost                       admin.unlockAccount

```

For example, if we want to unlock account n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE:

```javascript
> admin.unlockAccount("n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE")
Unlock account n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE
Passphrase:
{
    "result": {
        "result": true
    }
}
```
*ps: all test key files default in keydir passphrase is:passphrase*

The command parameters of the command line are consistent with the parameters of the RPC interface. [NEB RPC](https://github.com/nebulasio/wiki/blob/master/rpc.md) and [NEB RPC_Admin](https://github.com/nebulasio/wiki/blob/master/rpc_admin.md).


## RPC
Nebulas provide both [gRPC](https://grpc.io) and RESTful API, let users interact with Nebulas.

#### Endpoint

Default endpoints:

| API | URL | Protocol |
|-------|:------------:|:------------:|
| gRPC |  http://localhost:8684 | Protobuf
| RESTful |http://localhost:8685 | HTTP |

##### gRPC API
We can play the gRPC example testing client code:

```bash
cd rpc/testing/client/
go run main.go
```

The testing client gets account state from sender address, makes a transaction from sender to receiver, and also checks the account state of receiver address.

We can see client log output like:

```
GetAccountState n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE nonce 1 value 78
SendTransaction n1FF1nz6tarkDVwWQkMnnwFPuPKUaQTdptE -> n1GmkKH6nBMw4rrjt16RrJ9WcgvKUtAZP1s value 2 hash:"d9258c06899412169f969807629e1c152b54a3c4033e43727f3a74855849ffa6"
GetAccountState n1GmkKH6nBMw4rrjt16RrJ9WcgvKUtAZP1s nonce 0 value 2
```

##### HTTP
Now we also provided HTTP to access the RPC API.You first need to generate a mapping from GRPC to HTTP:

```bash
cd rpc/pb
make all
```

The file that ends with gw.go is the mapping file.
Now we can access the rpc API directly from our browser, you can update the *http_listen* in _config.conf_ to change HTTP default port.

###### Example:
```bash
// call nebstate api
curl -i -H 'Content-Type: application/json' -X GET http://localhost:8685/v1/user/nebstate

```

#### API list


For more details, please refer to [NEB RPC](https://github.com/nebulasio/wiki/blob/master/rpc.md) and [NEB RPC_Admin](https://github.com/nebulasio/wiki/blob/master/rpc_admin.md).


## NVM
Nebulas implemented an nvm to run smart contracts like ethereum. NVM provides a javascript runtime environment through v8-engine. Users can write smart contracts by javascript, which is the most popular language in the world.

We can deploy and run smart contracts by two RPC methods:

```
SendTransaction()/SendrawTransaction
Call()
```

Now you can create & deploy & call smart contracts directly over HTTP/console just by 'SendTransaction()'.
If you want to create & deploy smart contracts:

```javascript
"use strict";

var DepositeContent = function (text) {
	if (text) {
		var o = JSON.parse(text);
		this.balance = new BigNumber(o.balance);
		this.expiryHeight = new BigNumber(o.expiryHeight);
	} else {
		this.balance = new BigNumber(0);
		this.expiryHeight = new BigNumber(0);
	}
};

DepositeContent.prototype = {
	toString: function () {
		return JSON.stringify(this);
	}
};

var BankVaultContract = function () {
	LocalContractStorage.defineMapProperty(this, "bankVault", {
		parse: function (text) {
			return new DepositeContent(text);
		},
		stringify: function (o) {
			return o.toString();
		}
	});
};

// save value to contract, only after height of block, users can takeout
BankVaultContract.prototype = {
	init: function () {
		//TODO:
	},

	save: function (height) {
		var from = Blockchain.transaction.from;
		var value = Blockchain.transaction.value;
		var bk_height = new BigNumber(Blockchain.block.height);

		var orig_deposit = this.bankVault.get(from);
		if (orig_deposit) {
			value = value.plus(orig_deposit.balance);
		}

		var deposit = new DepositeContent();
		deposit.balance = value;
		deposit.expiryHeight = bk_height.plus(height);

		this.bankVault.put(from, deposit);
	},

	takeout: function (value) {
		var from = Blockchain.transaction.from;
		var bk_height = new BigNumber(Blockchain.block.height);
		var amount = new BigNumber(value);

		var deposit = this.bankVault.get(from);
		if (!deposit) {
			throw new Error("No deposit before.");
		}

		if (bk_height.lt(deposit.expiryHeight)) {
			throw new Error("Can not takeout before expiryHeight.");
		}

		if (amount.gt(deposit.balance)) {
			throw new Error("Insufficient balance.");
		}

		var result = Blockchain.transfer(from, amount);
		if (!result) {
			throw new Error("transfer failed.");
		}
		Event.Trigger("BankVault", {
			Transfer: {
				from: Blockchain.transaction.to,
				to: from,
				value: amount.toString()
			}
		});

		deposit.balance = deposit.balance.sub(amount);
		this.bankVault.put(from, deposit);
	},

	balanceOf: function () {
		var from = Blockchain.transaction.from;
		return this.bankVault.get(from);
	},

	verifyAddress: function (address) {
		// 1-valid, 0-invalid
		var result = Blockchain.verifyAddress(address);
		return {
			valid: result == 0 ? false : true
		};
	}
};

module.exports = BankVaultContract;

```

1. create your smart contracts source.
2. call 'SendTransaction()', the params 'from' and 'to' must be the same. Detailed Interface Documentation [RPC_Admin](https://github.com/nebulasio/wiki/blob/master/rpc_admin.md#sendtransaction).

```bash
// Request
curl -i -H 'Accept: application/json' -X POST http://localhost:8685/v1/admin/transaction -H 'Content-Type: application/json' -d '{"from":"n1NZttPdrJCwHgFN3V6YnSDaD5g8UbVppoC","to":"n1NZttPdrJCwHgFN3V6YnSDaD5g8UbVppoC", "value":"0","nonce":7,"gasPrice":"1000000","gasLimit":"2000000","contract":{"source":"\"use strict\";var DepositeContent=function(text){if(text){var o=JSON.parse(text);this.balance=new BigNumber(o.balance);this.expiryHeight=new BigNumber(o.expiryHeight);}else{this.balance=new BigNumber(0);this.expiryHeight=new BigNumber(0);}};DepositeContent.prototype={toString:function(){return JSON.stringify(this);}};var BankVaultContract=function(){LocalContractStorage.defineMapProperty(this,\"bankVault\",{parse:function(text){return new DepositeContent(text);},stringify:function(o){return o.toString();}});};BankVaultContract.prototype={init:function(){},save:function(height){var from=Blockchain.transaction.from;var value=Blockchain.transaction.value;var bk_height=new BigNumber(Blockchain.block.height);var orig_deposit=this.bankVault.get(from);if(orig_deposit){value=value.plus(orig_deposit.balance);} var deposit=new DepositeContent();deposit.balance=value;deposit.expiryHeight=bk_height.plus(height);this.bankVault.put(from,deposit);},takeout:function(value){var from=Blockchain.transaction.from;var bk_height=new BigNumber(Blockchain.block.height);var amount=new BigNumber(value);var deposit=this.bankVault.get(from);if(!deposit){throw new Error(\"No deposit before.\");} if(bk_height.lt(deposit.expiryHeight)){throw new Error(\"Can not takeout before expiryHeight.\");} if(amount.gt(deposit.balance)){throw new Error(\"Insufficient balance.\");} var result=Blockchain.transfer(from,amount);if(!result){throw new Error(\"transfer failed.\");} Event.Trigger(\"BankVault\",{Transfer:{from:Blockchain.transaction.to,to:from,value:amount.toString()}});deposit.balance=deposit.balance.sub(amount);this.bankVault.put(from,deposit);},balanceOf:function(){var from=Blockchain.transaction.from;return this.bankVault.get(from);},verifyAddress:function(address){var result=Blockchain.verifyAddress(address);return{valid:result==0?false:true};}};module.exports=BankVaultContract;","sourceType":"js", "args":""}}'

// Result
{
	"result":
	{
		"txhash":"2dd7186d266c2139fcc92446b364ef1a1037bc96d571f7c8a1716bec44fe25d8","contract_address":"n1qsgj2C5zmYzS9TSkPTnp15bhCCocRPwno"
	}
}

```

The return value for deploying a smart contract is the transaction's hash address `txhash` and the contract's deployment address `contract_address`. We can easily check the contract's address information using the console to verify whether the contract has been deployed successfully.

```js
> api.getTransactionReceipt("2dd7186d266c2139fcc92446b364ef1a1037bc96d571f7c8a1716bec44fe25d8")

```
The transaction `status ` in receipt: 0 failed, 1 success, 2 pending.

The way to call a smart contract in Nebulas is also straightforward, using the sendTransaction() method to invoke the smart contract directly.

```bash

// Request
curl -i -H 'Accept: application/json' -X POST http://localhost:8685/v1/admin/transaction -H 'Content-Type: application/json' -d '{"from":"n1NZttPdrJCwHgFN3V6YnSDaD5g8UbVppoC","to":"n1qsgj2C5zmYzS9TSkPTnp15bhCCocRPwno", "value":"100","nonce":8,"gasPrice":"1000000","gasLimit":"2000000","contract":{"function":"save","args":"[0]"}}'

// Result
{
	"result":{"txhash":"b55358c2e12c1d48d4e6beaee7002a59138294fb2896ea8059ff5277553af59f","contract_address":""}
}

```

The smart contracts and execution methods that have been submitted in Nebulas are submitted to the chain. It is also easy to find out how smart contracts have generated data. Smart contracts can be invoked via the rpc interface call() method. Calling a contract method via the `call()` method is not posted to the chain.

```js
call(from, to, value, nonce, gasPrice, gasLimit, contract)
```
Call the smart contract balanceOf() method:

```js
// Request
curl -i -H 'Accept: application/json' -X POST http://localhost:8685/v1/user/call -H 'Content-Type: application/json' -d '{"from":"n1NZttPdrJCwHgFN3V6YnSDaD5g8UbVppoC","to":"n1qsgj2C5zmYzS9TSkPTnp15bhCCocRPwno","value":"0","nonce":10,"gasPrice":"1000000","gasLimit":"2000000","contract":{"function":"balanceOf","args":""}}'

// Result
{
	"result":{"result":"{\"balance\":\"50\",\"expiryHeight\":\"556\"}","execute_err":"","estimate_gas":"20209"}
}

```
The essence of smart contract query is to submit a transaction, transactions are submitted only in the local implementation or local network, so the smart contract inquiries immediately take effect. With the query method it returns the results and you can see the results.

## TestNet

We are glad to release Nebulas Testnet here. You can use and join our [TestNet](https://github.com/nebulasio/wiki/blob/master/testnet.md) right now.

## MaintNet

We are glad to release Nebulas Mainnet here. You can use and join our [MainNet](https://github.com/nebulasio/wiki/blob/master/mainnet.md) right now.

## Explorer

Nebulas provides a block explorer to view block/transaction information.[explorer](https://explorer.nebulas.io/#/)

## Wallet

Nebulas provides a web wallet to send transaction and deploy/call contract.[wallet](https://github.com/nebulasio/web-wallet)

## Contribution

We are very glad that you are considering to help Nebulas Team or go-nebulas project, including but not limited to source code, documents or others.

If you'd like to contribute, please fork, fix, commit and send a pull request for the maintainers to review and merge into the main code base. If you wish to submit more complex changes though, please check up with the core devs first on our [slack channel](http://nebulasio.herokuapp.com) to ensure those changes are in line with the general philosophy of the project and/or get some early feedback which can make both your efforts much lighter as well as our review and merge procedures quick and simple.

Please refer to our [contribution guideline](https://github.com/nebulasio/wiki/blob/master/contribute.md) for more information.

Thanks.

## License

The go-nebulas project is licensed under the [GNU Lesser General Public License Version 3.0 (“LGPL v3”)](https://www.gnu.org/licenses/lgpl-3.0.en.html).

For the more information about licensing, please refer to [Licensing](https://github.com/nebulasio/wiki/blob/master/licensing.md) page.
