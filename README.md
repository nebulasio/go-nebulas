# go-nebulas

Official Go implementation of the Nebulas protocol.

[![Build Status](https://travis-ci.org/nebulasio/go-nebulas.svg?branch=develop)](https://travis-ci.org/nebulasio/go-nebulas)

For the roadmap of Nebulas, please visit the [Roadmap](https://github.com/nebulasio/wiki/blob/master/roadmap.md) page.

For more information of Nebulas protocol, design documents, please refer to our [wiki](https://github.com/nebulasio/wiki).

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

time="2017-12-29T23:25:17+08:00" level=info msg="Starting neblet..." file=neblet.go func="neblet.(*Neblet).Start" line=121
time="2017-12-29T23:25:17+08:00" level=info msg="node start" addrs="[/ip4/127.0.0.1/tcp/8680 /ip4/127.94.0.1/tcp/8680 /ip4/127.94.0.2/tcp/8680 /ip4/192.168.1.13/tcp/8680 /ip4/169.254.66.0/tcp/8680]" file=net_service.go func="p2p.(*NetService).Start" id=QmP7HDFcYmJL12Ez4ZNVCKjKedfE7f48f1LAkUc3Whz4jP line=691
time="2017-12-29T23:25:17+08:00" level=info msg="net.start: node start and join to p2p network success and listening for connections on [0.0.0.0:8680]... " file=net_service.go func="p2p.(*NetService).start" line=758
time="2017-12-29T23:25:17+08:00" level=info msg="Sync.Start: i am a seed node." file=sync_manager.go func="sync.(*Manager).Start" line=109
time="2017-12-29T23:25:17+08:00" level=info msg="Starting RPC server at: 127.0.0.1:8684" file=api_server.go func="rpc.(*APIServer).start" line=59
time="2017-12-29T23:25:17+08:00" level=info msg="control mining." file=sync_manager.go func="sync.(*Manager).Start" line=111 start=true
time="2017-12-29T23:25:18+08:00" level=warning msg=mintBlock. elapsed=17118 err="now is not time to forg block" file=dpos.go func="dpos.(*Dpos).blockLoop" line=357 tail="{\"height\":3, \"hash\":\"07052bac3c3f7efa673144710f98d02a9fda7369a096834939f81e42debe2fa6\", \"parentHash\":\"4343ed4818b5242cca6e4b350a082029279deb8730e496c789916e57536469d8\", \"accState\":\"3b607836efc991d59a9c1c11bba3f8be7c6632c55e08990a18c9df18c6a9915c\", \"nonce\":0, \"timestamp\": 1514544000, \"coinbase\": \"1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c\"}"
time="2017-12-29T23:25:19+08:00" level=warning msg=mintBlock. elapsed=17119 err="now is not time to forg block" file=dpos.go func="dpos.(*Dpos).blockLoop" line=357 tail="{\"height\":3, \"hash\":\"07052bac3c3f7efa673144710f98d02a9fda7369a096834939f81e42debe2fa6\", \"parentHash\":\"4343ed4818b5242cca6e4b350a082029279deb8730e496c789916e57536469d8\", \"accState\":\"3b607836efc991d59a9c1c11bba3f8be7c6632c55e08990a18c9df18c6a9915c\", \"nonce\":0, \"timestamp\": 1514544000, \"coinbase\": \"1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c\"}"
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
  coinbase: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8"
  signature_ciphers: ["ECC_SECP256K1"]
  miner: "9341709022928b38dae1f9e1cfbad25611e81f736fd192c5"
  passphrase: "passphrase"
}

rpc {
    rpc_listen: ["127.0.0.1:8684"]
    http_listen: ["127.0.0.1:8685"]
    http_module: ["api","admin"]
}

app {
    log_level: "info"
    log_file: "logs"
    enable_crash_report: false
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
time="2017-12-29T23:25:17+08:00" level=info msg="node start" addrs="[/ip4/127.0.0.1/tcp/8680 /ip4/127.94.0.1/tcp/8680 /ip4/127.94.0.2/tcp/8680 /ip4/192.168.1.13/tcp/8680 /ip4/169.254.66.0/tcp/8680]" file=net_service.go func="p2p.(*NetService).Start" id=QmP7HDFcYmJL12Ez4ZNVCKjKedfE7f48f1LAkUc3Whz4jP line=691
```
To start a node on another machine, we need to update p2p seed configuration in _config.conf_:

```protobuf
p2p {
  seed: "/ip4/127.0.0.1/tcp/8680/ipfs/QmP7HDFcYmJL12Ez4ZNVCKjKedfE7f48f1LAkUc3Whz4jP"
  port: 8681
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

INFO[2017-12-29T23:28:28+08:00] Starting neblet...                            file=neblet.go func="neblet.(*Neblet).Start" line=121
INFO[2017-12-29T23:28:28+08:00] node start                                    addrs="[/ip4/127.0.0.1/tcp/10002 /ip4/127.94.0.1/tcp/10002 /ip4/127.94.0.2/tcp/10002 /ip4/192.168.1.13/tcp/10002 /ip4/169.254.66.0/tcp/10002]" file=net_service.go func="p2p.(*NetService).Start" id=QmaQeoEG1XKztL7RnracUiWTVdxQX2QpybuXpnApCtYuwQ line=691
INFO[2017-12-29T23:28:28+08:00] net.start: node start and join to p2p network success and listening for connections on [0.0.0.0:10002]...   file=net_service.go func="p2p.(*NetService).start" line=758
INFO[2017-12-29T23:28:28+08:00] Starting RPC server at: 127.0.0.1:51512       file=api_server.go func="rpc.(*APIServer).start" line=59
INFO[2017-12-29T23:28:28+08:00] syncWithPeers: got tail                       block="{\"height\":3, \"hash\":\"07052bac3c3f7efa673144710f98d02a9fda7369a096834939f81e42debe2fa6\", \"parentHash\":\"4343ed4818b5242cca6e4b350a082029279deb8730e496c789916e57536469d8\", \"accState\":\"3b607836efc991d59a9c1c11bba3f8be7c6632c55e08990a18c9df18c6a9915c\", \"nonce\":0, \"timestamp\": 1514544000, \"coinbase\": \"1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c\"}" file=sync_manager.go func="sync.(*Manager).startSync" line=119 tail="&{QmaQeoEG1XKztL7RnracUiWTVdxQX2QpybuXpnApCtYuwQ 1 0xc4201322d0}"
INFO[2017-12-29T23:28:28+08:00] Sync: allNode -> [<peer.ID P7HDFc> <peer.ID aQeoEG>]  file=sync.go func="p2p.(*NetService).Sync" line=55
INFO[2017-12-29T23:28:28+08:00] dispatcher.loop: recvMsgCount=%d1             file=dispatcher.go func="net.(*Dispatcher).Start.func1" line=84
INFO[2017-12-29T23:28:28+08:00] StartMsgHandle.receiveSyncReplyCh: receive receiveSyncReplyCh message.  blocks="[{\"height\":3, \"hash\":\"07052bac3c3f7efa673144710f98d02a9fda7369a096834939f81e42debe2fa6\", \"parentHash\":\"4343ed4818b5242cca6e4b350a082029279deb8730e496c789916e57536469d8\", \"accState\":\"3b607836efc991d59a9c1c11bba3f8be7c6632c55e08990a18c9df18c6a9915c\", \"nonce\":0, \"timestamp\": 1514544000, \"coinbase\": \"1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c\"}]" file=asm_amd64.s from=QmP7HDFcYmJL12Ez4ZNVCKjKedfE7f48f1LAkUc3Whz4jP func=runtime.goexit line=2338
INFO[2017-12-29T23:28:28+08:00] control mining.                               file=sync_manager.go func="sync.(*Manager).loop" line=128 start=true
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
api.accounts              api.gasPrice              api.getNebState           api.sendTransaction
api.blockDump             api.getAccountState       api.getTransactionReceipt api.setRequest
api.call                  api.getBlockByHash        api.nodeInfo              api.subscribe
api.estimateGas           api.getEventsByHash       api.sendRawTransaction

> admin.
admin.getDynasty                    admin.sendTransactionWithPassphrase admin.signTransaction
admin.lockAccount                   admin.setHost                       admin.unlockAccount
admin.newAccount                    admin.setRequest
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

We are glad to release Nebulas Testnet. You can use and join our [TestNet](https://github.com/nebulasio/wiki/blob/master/testnet.md) right now. 

## Contribution

We are very glad that you are considering to help Nebulas Team or go-nebulas project, including but not limited to source code, documents or others.

If you'd like to contribute, please fork, fix, commit and send a pull request for the maintainers to review and merge into the main code base. If you wish to submit more complex changes though, please check up with the core devs first on our [slack channel](http://nebulasio.herokuapp.com) to ensure those changes are in line with the general philosophy of the project and/or get some early feedback which can make both your efforts much lighter as well as our review and merge procedures quick and simple.

Please refer to our [contribution guideline](https://github.com/nebulasio/wiki/blob/master/contribute.md) for more information.

Thanks.

## License

The go-nebulas project is licensed under the [GNU Lesser General Public License Version 3.0 (“LGPL v3”)](https://www.gnu.org/licenses/lgpl-3.0.en.html).

For the more information about licensing, please refer to [Licensing](https://github.com/nebulasio/wiki/blob/master/licensing.md) page.
