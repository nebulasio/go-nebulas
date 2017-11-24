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
```
cd $GOPATH/src
go get -u -v github.com/nebulasio/go-nebulas
```
The project is under active development. Its default branch is _develop_.
```
cd github.com/nebulasio/go-nebulas
```

New users may want to checkout and use the stable _master_ release.
```
git checkout master
```

2. Install dependencies packages.
```
make dep
```

3. Install dependent v8 libraries.
```
make deploy-v8
```

4. Build the neb binary.
```
make build
```

## Run

### Run seed node
Starting a Nebulas seed node is simple. After the build step above, run a command:

```
./neb
```

You will see log message output like:

```
time="2017-11-22T14:43:10+08:00" level=info msg="Starting neblet..." file=neblet.go func="neblet.(*Neblet).Start" line=68
time="2017-11-22T14:43:11+08:00" level=debug msg="node init success" file=net_service.go func=p2p.NewNetService line=129 node.id=QmPyr4ZbDmwF1nWxymTktdzspcBFPL6X1v3Q5nT7PGNtUN node.port=51413
time="2017-11-22T14:43:11+08:00" level=info msg="node start" addrs="[/ip4/192.168.1.13/tcp/51413]" file=net_service.go func="p2p.(*NetService).Start" id=QmPyr4ZbDmwF1nWxymTktdzspcBFPL6X1v3Q5nT7PGNtUN line=665
time="2017-11-22T14:43:11+08:00" level=info msg="RegisterNetService: register netservice success" file=net_service.go func="p2p.(*NetService).registerNetService" line=142
time="2017-11-22T14:43:11+08:00" level=info msg="net.start: node start and join to p2p network success and listening for connections on port 51413... " file=net_service.go func="p2p.(*NetService).start" line=731
time="2017-11-22T14:43:11+08:00" level=info msg="Sync.Start: i am a seed node." file=sync_manager.go func="sync.(*Manager).Start" line=105
time="2017-11-22T14:43:11+08:00" level=debug msg=running. file=asm_amd64.s func=runtime.goexit line=2338
time="2017-11-22T14:43:11+08:00" level=debug msg=running. file=asm_amd64.s func=runtime.goexit line=2338
time="2017-11-22T14:43:11+08:00" level=info msg="sync over, start mining" file=pow.go func="pow.(*Pow).SetCanMining" line=134
time="2017-11-22T14:43:11+08:00" level=debug msg="StartState enter." file=start.go func="pow.(*StartState).Enter" line=57
time="2017-11-22T14:43:11+08:00" level=info msg="Starting RPC server at: 127.0.0.1:52520" file=management_server.go func="rpc.(*ManagementServer).Start" line=51
time="2017-11-22T14:43:11+08:00" level=debug msg="State Transition." current="StartState 0xc4205d2320" file=pow.go from="StartState 0xc4205d2320" func="pow.(*Pow).stateLoop" line=218 success=true to="PrepareState 0xc4205d0010"
time="2017-11-22T14:43:11+08:00" level=debug msg="StartState leave." file=start.go func="pow.(*StartState).Leave" line=67
time="2017-11-22T14:43:11+08:00" level=info msg="Starting RPC server at: 127.0.0.1:51510" file=api_server.go func="rpc.(*APIServer).Start" line=44
time="2017-11-22T14:43:11+08:00" level=debug msg="PrepareState enter." file=prepare.go func="pow.(*PrepareState).Enter" line=55
```

From the log, we can see the binary execution starts neblet, starts network service, starts RPC API server, and starts consensus state machine.


### Configurations
Neb uses [Protocol Buffer](https://github.com/google/protobuf) to load configurations. The default config file is named as config.pb.txt and looks like following:

```
p2p {
  # seed: "UNCOMMENT_AND_SET_SEED_NODE_ADDRESS"
  port: 51413
  chain_id: 100
  version: 1
}

rpc {
  api_port: 51510
  management_port: 52520
  api_http_port: 8090
  management_http_port: 8191
}

pow {
  coinbase: "8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf"
}

storage {
  location: "seed.db"
}

account {
  # keystore.SECP256K1 = 1
  signature: 1

  # keystore.SCRYPT = 1 << 4
  encrypt: 16

  key_dir: "keydir"

  test_passphrase: "passphrase"
}

influxdb {
  host: "http://localhost:8086"
  db: "nebulas"
  username: "test"
  password: "test"
}

metrics {
  enable: false
}

```
The configuration schema is defined in proto _neblet/pb/config.proto:Config_. To load a different config file when starting the neb binary, use flag -c. For example,
```
./neb -c <path>/config-seed.pb.txt
```

Neb supports loading KeyStore file in Ethereum format. KeyStore files from config _key_dir_ are loaded during neb bootstrap. Example testing KeyStore looks like

```
{"version":3,"id":"272a46f1-5141-4234-b948-1b45c6708962","address":"555fcb1b7051d3aea5cf2c0167b4e19ed6a4f98d","Crypto":{"ciphertext":"ecd4b817fa9ebed736235476c91dec43e73e0ca3e8d2f13c004725349882fb49","cipherparams":{"iv":"1ab4ed89c95f66e994f183fed23df9f9"},"cipher":"aes-128-ctr","kdf":"scrypt","kdfparams":{"dklen":32,"salt":"baef3f92cdde9fd97a00879ce060763101530e9e66e4c75ec74352a41419bde0","n":1024,"r":8,"p":1},"mac":"d8ea471cea8184fb7b19c1563804b85a31a2b3d792dc59ecccdb15dbfb3cebc0"}}

```

### Run non-seed node
Now we can get the seed address from the seed node log output above. Get pretty address and id from log starts with **"node start"**. The seed address from log above is

```
time="2017-11-22T15:01:43+08:00" level=info msg="node start" addrs="[/ip4/192.168.1.13/tcp/51413]" file=net_service.go func="p2p.(*NetService).Start" id=QmPyr4ZbDmwF1nWxymTktdzspcBFPL6X1v3Q5nT7PGNtUN line=665
```
To start a non-seed node on another machine, we need to update p2p seed configuration in _config-normal.pb.txt_:

```
p2p {
  seed: "/ip4/192.168.1.13/tcp/51413/ipfs/QmPyr4ZbDmwF1nWxymTktdzspcBFPL6X1v3Q5nT7PGNtUN"
  port: 51415
}
...
```

Note, if the non-seed node is running on the same machine, we need to use a different config file, and also assign different P2P and RPC server ports.

Now we can start the non-seed nodes by simply run command
```
./neb -c config-normal.pb.txt
```

Then it will connect to the seed node started earlier to join that network. The log output will look like:

```
time="2017-11-22T15:14:50+08:00" level=info msg="Starting neblet..." file=neblet.go func="neblet.(*Neblet).Start" line=68
time="2017-11-22T15:14:50+08:00" level=debug msg="node init success" file=net_service.go func=p2p.NewNetService line=129 node.id=QmZTq1fopzbPU4dY3Bu12R89yHtjZb9GSM2FhcKYbA67fC node.port=10000
time="2017-11-22T15:14:50+08:00" level=info msg="node start" addrs="[/ip4/192.168.1.13/tcp/10000]" file=net_service.go func="p2p.(*NetService).Start" id=QmZTq1fopzbPU4dY3Bu12R89yHtjZb9GSM2FhcKYbA67fC line=665
time="2017-11-22T15:14:50+08:00" level=info msg="RegisterNetService: register netservice success" file=net_service.go func="p2p.(*NetService).registerNetService" line=142
time="2017-11-22T15:14:50+08:00" level=debug msg="say hello to a node success" bootNode=/ip4/192.168.1.13/tcp/51413/ipfs/QmPyr4ZbDmwF1nWxymTktdzspcBFPL6X1v3Q5nT7PGNtUN file=net_service.go func="p2p.(*NetService).start.func1" line=712
time="2017-11-22T15:14:50+08:00" level=info msg="net.start: node start and join to p2p network success and listening for connections on port 10000... " file=net_service.go func="p2p.(*NetService).start" line=731
...
```

## REPL console
Nebulas provide an interactive javascript console, which can invoke all API and management RPC methods. Some management methods may require passphrase. Start console using the command:

```
./neb console
```

We have API and admin two schemes to access the console cmds. Users can quickly enter instructions using the TAB key.

```
> api.
api.accounts              api.getBlockByHash        api.sendRawTransaction
api.blockDump             api.getNebState           api.sendTransaction
api.call                  api.getTransactionReceipt
api.getAccountState       api.nodeInfo
```
```
> admin.
admin.lockAccount                   admin.setHost
admin.newAccount                    admin.signTransaction
admin.sendTransactionWithPassphrase admin.unlockAccount
```

For example, if we want to unlock account 8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf:

```
> admin.unlockAccount('8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf')
Unlock account 8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf
Passphrase:
{
    "result": true
}
```

The command parameters of the command line are consistent with the parameters of the RPC interface. [NEB RPC](https://github.com/nebulasio/wiki/blob/master/json-rpc.md).


## RPC
Nebulas provide both [gRPC](https://grpc.io) and RESTful API, let users interact with Nebulas.

#### Endpoint

Default endpoints:

| API | URL | Protocol |
|-------|:------------:|:------------:|
| gRPC |  http://localhost:52520 | Protobuf
| RESTful |http://localhost:8191 | HTTP |

##### gRPC API
We can play the gRPC example testing client code:

```
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

```
cd rpc/pb
make all
```
The file that ends with gw.go is the mapping file.
Now we can access the rpc API directly from our browser, you can update the *api-http-port* and *management-http-port* in _config.pb.txt_ to change HTTP default port.

###### Example:
```
BlockDump:
curl -i -H 'Accept: application/json' -X POST http://localhost:8191/v1/block/dump -H 'Content-Type: application/json' -d '{"count":10}'
```

#### API list


For more details, please refer to [NEB RPC](https://github.com/nebulasio/wiki/blob/master/json-rpc.md).


## NVM
Nebulas implemented an nvm to run smart contracts like ethereum.NVM provides a javascript runtime environment through v8-engine.Users can write smart contracts by javascript, which is the most popular language in the world.

We can deploy and run smart contracts by two RPC methods:

```
SendTransaction()
Call()
```
Now you can create & deploy & call smart contracts directly over HTTP/console just by 'SendTransaction()'.
If you want to create & deploy smart contracts:

```
"use strict";

var BankVaultContract = function() {
    LocalContractStorage.defineMapProperty(this, "bankVault");
};

// save value to contract, only after height of block, users can takeout
BankVaultContract.prototype = {
    init:function() {},
    save:function(height) {
        var deposit = this.bankVault.get(Blockchain.transaction.from);
        var value = new BigNumber(Blockchain.transaction.value);
        if (deposit != null && deposit.balance.length > 0) {
            var balance = new BigNumber(deposit.balance);
            value = value.plus(balance);
        }
        var content = {
            balance:value.toString(),
            height:Blockchain.block.height + height
        };
        this.bankVault.put(Blockchain.transaction.from, content);
    },
    takeout:function(amount) {
        var deposit = this.bankVault.get(Blockchain.transaction.from);
        if (deposit == null) {
            return 0;
        }
        if (Blockchain.block.height < deposit.height) {
            return 0;
        }
        var balance = new BigNumber(deposit.balance);
        var value = new BigNumber(amount);
        if (balance.lessThan(value)) {
            return 0;
        }
        var result = Blockchain.transfer(Blockchain.transaction.from, value);
        if (result > 0) {
            deposit.balance = balance.dividedBy(value).toString();
            this.bankVault.put(Blockchain.transaction.from, deposit);
        }
        return result;
    }
};

module.exports = BankVaultContract;
```

1. create your smart contracts source.
2. call 'SendTransaction()', the params 'from' and 'to' must be the same.

```
curl -i -H 'Accept: application/json' -X POST http://localhost:8191/v1/transaction -H 'Content-Type: application/json' -d '{"from":"8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf","to":"8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf","nonce":1,"source":"\"use strict\";var BankVaultContract=function(){LocalContractStorage.defineMapProperty(this,\"bankVault\")};BankVaultContract.prototype={init:function(){},save:function(height){var deposit=this.bankVault.get(Blockchain.transaction.from);var value=new BigNumber(Blockchain.transaction.value);if(deposit!=null&&deposit.balance.length>0){var balance=new BigNumber(deposit.balance);value=value.plus(balance)}var content={balance:value.toString(),height:Blockchain.block.height+height};this.bankVault.put(Blockchain.transaction.from,content)},takeout:function(amount){var deposit=this.bankVault.get(Blockchain.transaction.from);if(deposit==null){return 0}if(Blockchain.block.height<deposit.height){return 0}var balance=new BigNumber(deposit.balance);var value=new BigNumber(amount);if(balance.lessThan(value)){return 0}var result=Blockchain.transfer(Blockchain.transaction.from,value);if(result>0){deposit.balance=balance.dividedBy(value).toString();this.bankVault.put(Blockchain.transaction.from,deposit)}return result}};module.exports=BankVaultContract;", "args":""}'
```

If you succeed in deploying a smart contract, you will get the contract address & transaction hash as response.
Then you can call this smart contract:

1. get the smart contract address.
2. give the 'function' you want to call.

```
curl -i -H 'Accept: application/json' -X POST http://localhost:8191/v1/call -H 'Content-Type: application/json' -d '{"from":"8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf","to":"4df690cad7727510f386cddb9416f601de69e48ac662c44c","nonce":2,"function":"save","args":"[0]"}'
```

## Contribution

We are very glad that you are considering to help Nebulas Team or go-nebulas project, including but not limited to source code, documents or others.

If you'd like to contribute, please fork, fix, commit and send a pull request for the maintainers to review and merge into the main code base. If you wish to submit more complex changes though, please check up with the core devs first on our [slack channel](http://nebulasio.herokuapp.com) to ensure those changes are in line with the general philosophy of the project and/or get some early feedback which can make both your efforts much lighter as well as our review and merge procedures quick and simple.

Please refer to our [contribution guideline](https://github.com/nebulasio/wiki/blob/master/contribute.md) for more information.

Thanks.

## License

The go-nebulas project is licensed under the [GNU Lesser General Public License Version 3.0 (“LGPL v3”)](https://www.gnu.org/licenses/lgpl-3.0.en.html).

For the more information about licensing, please refer to [Licensing](https://github.com/nebulasio/wiki/blob/master/licensing.md) page.
