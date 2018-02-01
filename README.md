# go-nebulas

Official Go implementation of the Nebulas protocol.

[![Build Status](https://travis-ci.org/nebulasio/go-nebulas.svg?branch=master)](https://travis-ci.org/nebulasio/go-nebulas)

For the roadmap of Nebulas, please visit the [Roadmap](https://github.com/nebulasio/wiki/blob/master/roadmap.md) page.

For more information of Nebulas protocol, design documents, please refer to our [wiki](https://github.com/nebulasio/wiki).

TestNet is released, please check [here](https://github.com/nebulasio/wiki/blob/master/testnet.md) for more details

## Building from source

### Prerequisites

| Components | Version | Description |
|----------|-------------|-------------|
|[Golang](https://golang.org) | >= 1.8| The Go Programming Language |
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

2. Install dependencies packages.

```bash
make dep
```

3. Install dependent v8 libraries.

```bash
make deploy-v8
```

4. Build the neb binary.

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
INFO[2018-02-01T20:43:04+08:00] Setuped Neblet.                               file=neblet.go func="neblet.(*Neblet).Setup" line=151
INFO[2018-02-01T20:43:04+08:00] Starting Neblet...                            file=neblet.go func="neblet.(*Neblet).Start" line=177
INFO[2018-02-01T20:43:04+08:00] Starting NetService...                        file=net_service.go func="net.(*NetService).Start" line=54
INFO[2018-02-01T20:43:04+08:00] Starting NetService Dispatcher...             file=dispatcher.go func="net.(*Dispatcher).Start" line=88
INFO[2018-02-01T20:43:04+08:00] Starting NetService Node...                   file=node.go func="net.(*Node).Start" line=96
INFO[2018-02-01T20:43:04+08:00] Started NewService Dispatcher.                file=dispatcher.go func="net.(*Dispatcher).loop" line=96
INFO[2018-02-01T20:43:04+08:00] Starting NetService StreamManager...          file=stream_manager.go func="net.(*StreamManager).Start" line=57
INFO[2018-02-01T20:43:04+08:00] Starting NetService RouteTable Sync...        file=route_table.go func="net.(*RouteTable).Start" line=91
INFO[2018-02-01T20:43:04+08:00] Started NetService StreamManager.             file=stream_manager.go func="net.(*StreamManager).loop" line=117
INFO[2018-02-01T20:43:04+08:00] Started NetService Node.                      file=net_service.go func="net.(*NetService).Start" id=QmP7HDFcYmJL12Ez4ZNVCKjKedfE7f48f1LAkUc3Whz4jP line=61 listening address="[/ip4/127.0.0.1/tcp/8680 /ip4/127.94.0.1/tcp/8680 /ip4/127.94.0.2/tcp/8680 /ip4/192.168.1.13/tcp/8680]"
INFO[2018-02-01T20:43:04+08:00] Started NetService RouteTable Sync.           file=route_table.go func="net.(*RouteTable).syncLoop" line=123
INFO[2018-02-01T20:43:04+08:00] Started NetService.                           file=net_service.go func="net.(*NetService).Start" line=70
INFO[2018-02-01T20:43:04+08:00] Starting RPC GRPCServer...                    file=server.go func="rpc.(*Server).Start" line=81
INFO[2018-02-01T20:43:04+08:00] Started RPC GRPCServer.                       address="127.0.0.1:8684" file=server.go func="rpc.(*Server).Start" line=89
INFO[2018-02-01T20:43:04+08:00] Starting RPC Gateway GRPCServer...            file=neblet.go func="neblet.(*Neblet).Start" http-server="[127.0.0.1:8685]" line=202 rpc-server="127.0.0.1:8684"
INFO[2018-02-01T20:43:04+08:00] Starting BlockChain...                        file=blockchain.go func="core.(*BlockChain).Start" line=168
INFO[2018-02-01T20:43:04+08:00] Starting BlockPool...                         file=neblet.go func="neblet.(*Neblet).Start" line=209 size=1024
INFO[2018-02-01T20:43:04+08:00] Starting TransactionPool...                   file=neblet.go func="neblet.(*Neblet).Start" line=210 size=40960
INFO[2018-02-01T20:43:04+08:00] Started BlockChain.                           file=blockchain.go func="core.(*BlockChain).loop" line=181
INFO[2018-02-01T20:43:04+08:00] Started BlockPool.                            file=block_pool.go func="core.(*BlockPool).loop" line=252
INFO[2018-02-01T20:43:04+08:00] Starting EventEmitter...                      file=neblet.go func="neblet.(*Neblet).Start" line=211 size=1024
INFO[2018-02-01T20:43:04+08:00] Started TransactionPool.                      file=asm_amd64.s func=runtime.goexit line=2338 size=40960
INFO[2018-02-01T20:43:04+08:00] Started EventEmitter.                         file=event.go func="core.(*EventEmitter).loop" line=139
INFO[2018-02-01T20:43:04+08:00] Starting Dpos Mining...                       file=dpos.go func="dpos.(*Dpos).Start" line=123
INFO[2018-02-01T20:43:04+08:00] Started Sync Service.                         file=sync_service.go func="sync.(*Service).startLoop" line=151
INFO[2018-02-01T20:43:04+08:00] Started Dpos Mining.                          file=dpos.go func="dpos.(*Dpos).blockLoop" line=505
INFO[2018-02-01T20:43:04+08:00] Enabled Dpos Mining...                        file=dpos.go func="dpos.(*Dpos).EnableMining" line=142
INFO[2018-02-01T20:43:04+08:00] This is a seed node.                          file=neblet.go func="neblet.(*Neblet).Start" line=238
INFO[2018-02-01T20:43:04+08:00] Resumed Dpos Mining.                          file=dpos.go func="dpos.(*Dpos).ResumeMining" line=219
INFO[2018-02-01T20:43:04+08:00] Started Neblet.                               file=neblet.go func="neblet.(*Neblet).Start" line=245
```

From the log, we can see the binary execution starts neblet, starts network service, starts RPC API server, and starts consensus state machine.

## Docker

### Build

`docker build -t nebulas`

### Run

`docker run -it -v $(pwd)/data.db:/nebulas/data.db nebulas`

### Configurations
Neb uses [Protocol Buffer](https://github.com/google/protobuf) to load configurations. The default config file is named as config.conf and looks like following:

```protobuf
network {
  listen: ["0.0.0.0:8680"]
  private_key: "conf/network/ed25519key"
  network_id: 1
}

chain {
  chain_id: 100
  datadir: "data.db"
  keydir: "keydir"
  genesis: "conf/default/genesis.conf"
  start_mine: true
  coinbase: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8"
  miner: "75e4e5a71d647298b88928d8cb5da43d90ab1a6c52d0905f"
  passphrase: "passphrase"
  signature_ciphers: ["ECC_SECP256K1"]
}

rpc {
    rpc_listen: ["127.0.0.1:8684"]
    http_listen: ["127.0.0.1:8685"]
    http_module: ["api","admin"]
}

app {
    log_level: "debug"
    log_file: "logs"
    enable_crash_report: true
    crash_report_url: "https://crashreport.nebulas.io"
    pprof: {
        http_listen: "127.0.0.1:7777"
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
{"version":3,"id":"272a46f1-5141-4234-b948-1b45c6708962","address":"555fcb1b7051d3aea5cf2c0167b4e19ed6a4f98d","Crypto":{"ciphertext":"ecd4b817fa9ebed736235476c91dec43e73e0ca3e8d2f13c004725349882fb49","cipherparams":{"iv":"1ab4ed89c95f66e994f183fed23df9f9"},"cipher":"aes-128-ctr","kdf":"scrypt","kdfparams":{"dklen":32,"salt":"baef3f92cdde9fd97a00879ce060763101530e9e66e4c75ec74352a41419bde0","n":1024,"r":8,"p":1},"mac":"d8ea471cea8184fb7b19c1563804b85a31a2b3d792dc59ecccdb15dbfb3cebc0"}}

```

### Run node
Now we can get the seed address from the seed node log output above. Get pretty address and id from log starts with **"node start"**. The seed address from log above is

```
INFO[2018-02-01T20:43:04+08:00] Started NetService Node.                      file=net_service.go func="net.(*NetService).Start" id=QmP7HDFcYmJL12Ez4ZNVCKjKedfE7f48f1LAkUc3Whz4jP line=61 listening address="[/ip4/127.0.0.1/tcp/8680 /ip4/127.94.0.1/tcp/8680 /ip4/127.94.0.2/tcp/8680 /ip4/192.168.1.13/tcp/8680]"
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
./neb -c conf/example/config.2fe3f9.conf
```

Then it will connect to the seed node started earlier to join that network. The log output will look like:

```

INFO[2018-02-01T20:50:14+08:00] Setuping Neblet...                            file=neblet.go func="neblet.(*Neblet).Setup" line=98
INFO[2018-02-01T20:50:14+08:00] Genesis Configuration.                        consensus.dpos.dynasty="[1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c 2fe3f9f51f9a05dd5f7c5329127f7c917917149b4e16b0b8 333cb3ed8c417971845382ede3cf67a0a96270c05fe2f700 48f981ed38910f1232c1bab124f650c482a57271632db9e3 59fc526072b09af8a8ca9732dae17132c4e9127e43cf2232 75e4e5a71d647298b88928d8cb5da43d90ab1a6c52d0905f 7da9dabedb4c6e121146fb4250a9883d6180570e63d6b080 98a3eed687640b75ec55bf5c9e284371bdcaeab943524d51]" file=neblet.go func="neblet.(*Neblet).Setup" line=122 meta.chainid=100 token.distribution="[address:\"1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c\" value:\"10000000000000000000000\"  address:\"2fe3f9f51f9a05dd5f7c5329127f7c917917149b4e16b0b8\" value:\"10000000000000000000000\" ]"
INFO[2018-02-01T20:50:14+08:00] Tail Block.                                   file=neblet.go func="neblet.(*Neblet).Setup" line=122 tail="{\"height\": 1, \"hash\": \"0000000000000000000000000000000000000000000000000000000000000000\", \"parent_hash\": \"0000000000000000000000000000000000000000000000000000000000000000\", \"state\": \"df69d8eac19d1c6829007a284bf5cbeede8e529a002235a48c362d37626bb3e0\", \"txs\": \"\", \"events\": \"\", \"nonce\": 0, \"timestamp\": 0, \"dynasty\": \"3948716dc08db1e0f5f797daad39adb25b375874ca2552ac46f88427bced9314\", \"tx\": 0}"
INFO[2018-02-01T20:50:14+08:00] Latest Irreversible Block.                    block="{\"height\": 1, \"hash\": \"0000000000000000000000000000000000000000000000000000000000000000\", \"parent_hash\": \"0000000000000000000000000000000000000000000000000000000000000000\", \"state\": \"df69d8eac19d1c6829007a284bf5cbeede8e529a002235a48c362d37626bb3e0\", \"txs\": \"\", \"events\": \"\", \"nonce\": 0, \"timestamp\": 0, \"dynasty\": \"3948716dc08db1e0f5f797daad39adb25b375874ca2552ac46f88427bced9314\", \"tx\": 0}" file=neblet.go func="neblet.(*Neblet).Setup" line=122
INFO[2018-02-01T20:50:14+08:00] Setuped Neblet.                               file=neblet.go func="neblet.(*Neblet).Setup" line=151
INFO[2018-02-01T20:50:14+08:00] Starting Neblet...                            file=neblet.go func="neblet.(*Neblet).Start" line=177
INFO[2018-02-01T20:50:14+08:00] Starting NetService...                        file=net_service.go func="net.(*NetService).Start" line=54
INFO[2018-02-01T20:50:14+08:00] Starting NetService Dispatcher...             file=dispatcher.go func="net.(*Dispatcher).Start" line=88
INFO[2018-02-01T20:50:14+08:00] Starting NetService Node...                   file=node.go func="net.(*Node).Start" line=96
INFO[2018-02-01T20:50:14+08:00] Starting NetService StreamManager...          file=stream_manager.go func="net.(*StreamManager).Start" line=57
INFO[2018-02-01T20:50:14+08:00] Started NewService Dispatcher.                file=dispatcher.go func="net.(*Dispatcher).loop" line=96
INFO[2018-02-01T20:50:14+08:00] Starting NetService RouteTable Sync...        file=route_table.go func="net.(*RouteTable).Start" line=91
INFO[2018-02-01T20:50:14+08:00] Started NetService StreamManager.             file=stream_manager.go func="net.(*StreamManager).loop" line=117
INFO[2018-02-01T20:50:14+08:00] Started NetService Node.                      file=net_service.go func="net.(*NetService).Start" id=Qmd5fVC3i428aEgzLXBhhQ4YwXTD2FPq5HpQrhbzeGhfod line=61 listening address="[/ip4/127.0.0.1/tcp/10001 /ip4/127.94.0.1/tcp/10001 /ip4/127.94.0.2/tcp/10001 /ip4/192.168.1.13/tcp/10001]"
INFO[2018-02-01T20:50:14+08:00] Started NetService.                           file=net_service.go func="net.(*NetService).Start" line=70
INFO[2018-02-01T20:50:14+08:00] Starting RPC GRPCServer...                    file=server.go func="rpc.(*Server).Start" line=81
INFO[2018-02-01T20:50:14+08:00] Started RPC GRPCServer.                       address="127.0.0.1:51511" file=server.go func="rpc.(*Server).Start" line=89
INFO[2018-02-01T20:50:14+08:00] Starting RPC Gateway GRPCServer...            file=neblet.go func="neblet.(*Neblet).Start" http-server="[127.0.0.1:8091]" line=202 rpc-server="127.0.0.1:51511"
INFO[2018-02-01T20:50:14+08:00] Starting BlockChain...                        file=blockchain.go func="core.(*BlockChain).Start" line=168
INFO[2018-02-01T20:50:14+08:00] Starting BlockPool...                         file=neblet.go func="neblet.(*Neblet).Start" line=209 size=1024
INFO[2018-02-01T20:50:14+08:00] Started BlockChain.                           file=blockchain.go func="core.(*BlockChain).loop" line=181
INFO[2018-02-01T20:50:14+08:00] Started NetService RouteTable Sync.           file=route_table.go func="net.(*RouteTable).syncLoop" line=123
INFO[2018-02-01T20:50:14+08:00] Started BlockPool.                            file=block_pool.go func="core.(*BlockPool).loop" line=252
INFO[2018-02-01T20:50:14+08:00] Starting TransactionPool...                   file=neblet.go func="neblet.(*Neblet).Start" line=210 size=40960
INFO[2018-02-01T20:50:14+08:00] Started TransactionPool.                      file=asm_amd64.s func=runtime.goexit line=2338 size=40960
INFO[2018-02-01T20:50:14+08:00] Starting EventEmitter...                      file=neblet.go func="neblet.(*Neblet).Start" line=211 size=1024
INFO[2018-02-01T20:50:14+08:00] Started EventEmitter.                         file=event.go func="core.(*EventEmitter).loop" line=139
INFO[2018-02-01T20:50:14+08:00] Starting Dpos Mining...                       file=dpos.go func="dpos.(*Dpos).Start" line=123
INFO[2018-02-01T20:50:14+08:00] Started Sync Service.                         file=sync_service.go func="sync.(*Service).startLoop" line=151
INFO[2018-02-01T20:50:14+08:00] Started Dpos Mining.                          file=dpos.go func="dpos.(*Dpos).blockLoop" line=505
INFO[2018-02-01T20:50:14+08:00] Enabled Dpos Mining...                        file=dpos.go func="dpos.(*Dpos).EnableMining" line=142
INFO[2018-02-01T20:50:14+08:00] Started Active Sync Task.                     file=blockchain.go func="core.(*BlockChain).StartActiveSync" line=570 syncpoint="{\"height\": 1, \"hash\": \"0000000000000000000000000000000000000000000000000000000000000000\", \"parent_hash\": \"0000000000000000000000000000000000000000000000000000000000000000\", \"state\": \"df69d8eac19d1c6829007a284bf5cbeede8e529a002235a48c362d37626bb3e0\", \"txs\": \"\", \"events\": \"\", \"nonce\": 0, \"timestamp\": 0, \"dynasty\": \"3948716dc08db1e0f5f797daad39adb25b375874ca2552ac46f88427bced9314\", \"tx\": 0}"
INFO[2018-02-01T20:50:14+08:00] Suspended Dpos Mining.                        file=dpos.go func="dpos.(*Dpos).SuspendMining" line=213
INFO[2018-02-01T20:50:14+08:00] Started Neblet.                               file=neblet.go func="neblet.(*Neblet).Start" line=245
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
api.accounts              api.getAccountState       api.getNebState           api.setRequest
api.blockDump             api.getBlockByHash        api.getTransactionReceipt api.subscribe
api.call                  api.getBlockByHeight      api.nodeInfo
api.estimateGas           api.getEventsByHash       api.sendRawTransaction
api.gasPrice              api.getGasUsed            api.sendTransaction

> admin.
admin.changeNetworkID               admin.sendTransactionWithPassphrase admin.startPprof
admin.getDelegateVoters             admin.setHost                       admin.stopMining
admin.getDynasty                    admin.setRequest                    admin.unlockAccount
admin.lockAccount                   admin.signTransaction
admin.newAccount                    admin.startMining

```

For example, if we want to unlock account 1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c:

```javascript
> admin.unlockAccount('1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c')
Unlock account 1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c
Passphrase:
{
    "result": true
}
```

The command parameters of the command line are consistent with the parameters of the RPC interface. [NEB RPC](https://github.com/nebulasio/wiki/blob/master/rpc.md).


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
GetAccountState 8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf nonce 1 value 78
SendTransaction 8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf -> 22ac3a9a2b1c31b7a9084e46eae16e761f83f02324092b09 value 2 hash:"d9258c06899412169f969807629e1c152b54a3c4033e43727f3a74855849ffa6"
GetAccountState 22ac3a9a2b1c31b7a9084e46eae16e761f83f02324092b09 nonce 0 value 2
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
// call BlockDump api
curl -i -H 'Accept: application/json' -X POST http://localhost:8685/v1/user/blockdump -H 'Content-Type: application/json' -d '{"count":1}'
```

#### API list


For more details, please refer to [NEB RPC](https://github.com/nebulasio/wiki/blob/master/rpc.md).


## NVM
Nebulas implemented an nvm to run smart contracts like ethereum. NVM provides a javascript runtime environment through v8-engine. Users can write smart contracts by javascript, which is the most popular language in the world.

We can deploy and run smart contracts by two RPC methods:

```
SendTransaction()
Call()
```

Now you can create & deploy & call smart contracts directly over HTTP/console just by 'SendTransaction()'.
If you want to create & deploy smart contracts:

```javascript
'use strict';

var DepositeContent = function (text) {
	if (text) {
		let o = JSON.parse(text);
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
			value = value.plus(balance);
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
			throw new Error("Can't takeout before expiryHeight.");
		}

		if (amount.gt(deposit.balance)) {
			throw new Error("Insufficient balance.");
		}

		var result = Blockchain.transfer(from, amount);
		if (result != 0) {
			throw new Error("transfer failed.");
		}
        Event.Trigger("BankVault", {
            Transfer: {
                from: Blockchain.transaction.to,
                to: from,
                value: amount.toString(),
            }
        });

		deposit.balance = deposit.balance.sub(amount);
		this.bankVault.put(from, deposit);
	}
};

module.exports = BankVaultContract;

```

1. create your smart contracts source.
2. call 'SendTransaction()', the params 'from' and 'to' must be the same.

```bash

curl -i -H 'Accept: application/json' -X POST http://localhost:8685/v1/user/transaction -H 'Content-Type: application/json' -d '{"from":"1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c","to":"1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c", "value":"0","nonce":2,"gasPrice":"1000000","gasLimit":"2000000","contract":{
"source":"\"use strict\";var BankVaultContract=function(){LocalContractStorage.defineMapProperty(this,\"bankVault\")};BankVaultContract.prototype={init:function(){},save:function(height){var deposit=this.bankVault.get(Blockchain.transaction.from);var value=new BigNumber(Blockchain.transaction.value);if(deposit!=null&&deposit.balance.length>0){var balance=new BigNumber(deposit.balance);value=value.plus(balance)}var content={balance:value.toString(),height:Blockchain.block.height+height};this.bankVault.put(Blockchain.transaction.from,content)},takeout:function(amount){var deposit=this.bankVault.get(Blockchain.transaction.from);if(deposit==null){return 0}if(Blockchain.block.height<deposit.height){return 0}var balance=new BigNumber(deposit.balance);var value=new BigNumber(amount);if(balance.lessThan(value)){return 0}var result=Blockchain.transfer(Blockchain.transaction.from,value);if(result>0){deposit.balance=balance.dividedBy(value).toString();this.bankVault.put(Blockchain.transaction.from,deposit)}return result}};module.exports=BankVaultContract;","sourceType":"js", "args":""}}'
```

If you succeed in deploying a smart contract, you will get the contract address & transaction hash as response.
Then you can call this smart contract:

1. get the smart contract address.
2. give the 'function' you want to call.

```bash

curl -i -H 'Accept: application/json' -X POST http://localhost:8685/v1/user/call -H 'Content-Type: application/json' -d '{"from":"1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c","to":"333cb3ed8c417971845382ede3cf67a0a96270c05fe2f700","value":"0","nonce":3,"gasPrice":"1000000","gasLimit":"2000000","contract":{"function":"save","args":"[0]"}}'
```

## TestNet

We are glad to release Nebulas Testnet here. You can use and join our [TestNet](https://github.com/nebulasio/wiki/blob/master/testnet.md) right now.

## Contribution

We are very glad that you are considering to help Nebulas Team or go-nebulas project, including but not limited to source code, documents or others.

If you'd like to contribute, please fork, fix, commit and send a pull request for the maintainers to review and merge into the main code base. If you wish to submit more complex changes though, please check up with the core devs first on our [slack channel](http://nebulasio.herokuapp.com) to ensure those changes are in line with the general philosophy of the project and/or get some early feedback which can make both your efforts much lighter as well as our review and merge procedures quick and simple.

Please refer to our [contribution guideline](https://github.com/nebulasio/wiki/blob/master/contribute.md) for more information.

Thanks.

## License

The go-nebulas project is licensed under the [GNU Lesser General Public License Version 3.0 (“LGPL v3”)](https://www.gnu.org/licenses/lgpl-3.0.en.html).

For the more information about licensing, please refer to [Licensing](https://github.com/nebulasio/wiki/blob/master/licensing.md) page.
