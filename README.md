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
The project is under active development. Its default branch is _develop_. New users may want to checkout and use the stable _master_ release.
```
cd github.com/nebulasio/go-nebulas
git checkout master
```

2. Install dependencies packages.
```
make dep
```

3. Build the neb binary.
```
make build
```

## Run

### Run seed node
Starting a Nebulas seed node is simple. After the build step above, run command:
```
./neb
```

You will see log message output like:
```
INFO[2017-10-18T03:16:09+08:00] Loading Neb config from file config.pb.txt    file=config.go func=neblet.LoadConfig line=30
  file=config.go func=neblet.LoadConfig line=37
INFO[2017-10-18T03:16:09+08:00] Loaded Neb config proto p2p:<port:51413 > rpc:<port:51510 > pow:<coinbase:"8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf" > account:<signature:1 encrypt:16 key_dir:"testKey" test_passphrase:"passphrase" >   file=config.go func=neblet.LoadConfig line=44
INFO[2017-10-18T03:16:09+08:00] Starting neblet...                            file=neblet.go func="neblet.(*Neblet).Start" line=57
INFO[2017-10-18T03:16:09+08:00] load test keys form:/Users/duranliu/go/src/github.com/nebulasio/go-nebulas/testKey  file=manager.go func="account.(*Manager).loadTestKey" line=86
INFO[2017-10-18T03:16:09+08:00] load test addr:8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf  file=manager.go func="account.(*Manager).loadTestKey" line=88
INFO[2017-10-18T03:16:09+08:00] load test addr:22ac3a9a2b1c31b7a9084e46eae16e761f83f02324092b09  file=manager.go func="account.(*Manager).loadTestKey" line=88
INFO[2017-10-18T03:16:09+08:00] load test addr:5cdadc1cfe3da0a3d067e9f1b195b90c5aebfb5afc8d43b4  file=manager.go func="account.(*Manager).loadTestKey" line=88
INFO[2017-10-18T03:16:09+08:00] NewNode: node make Host success               file=node.go func=p2p.NewNode line=64
INFO[2017-10-18T03:16:09+08:00] makeHost: boot node pretty id is QmVMamgHNoR8GBUbq4r9BTFzod5F2BvhSgaGLVxPpXcWNm  file=node.go func="p2p.(*Node).makeHost" line=153
chainID is 1
INFO[2017-10-18T03:16:09+08:00] Launch: node info {id -> <peer.ID VMamgH>, address -> [/ip4/192.168.1.18/tcp/51413]}  file=net_service.go func="p2p.(*NetService).Launch" line=536
INFO[2017-10-18T03:16:09+08:00] Launch: node start to join p2p network...     file=net_service.go func="p2p.(*NetService).Launch" line=541
INFO[2017-10-18T03:16:09+08:00] RegisterNetService: register netservice success  file=net_service.go func="p2p.(*NetService).RegisterNetService" line=85
INFO[2017-10-18T03:16:09+08:00] Launch: node start and join to p2p network success and listening for connections on port 51413...   file=net_service.go func="p2p.(*NetService).Launch" line=565
INFO[2017-10-18T03:16:09+08:00] Sync.Start: i am a seed node.                 file=sync_manager.go func="sync.(*Manager).Start" line=76
DEBU[2017-10-18T03:16:09+08:00] running.                                      file=asm_amd64.s func=runtime.goexit line=2338
INFO[2017-10-18T03:16:09+08:00] Starting RPC server at: 127.0.0.1:51510       file=server.go func="rpc.(*Server).Start" line=56
DEBU[2017-10-18T03:16:09+08:00] running.                                      file=asm_amd64.s func=runtime.goexit line=2338
DEBU[2017-10-18T03:16:09+08:00] StartState enter.                             file=start.go func="pow.(*StartState).Enter" line=56
```

From the log, we can see the binary execution loads configuration, loads keystore, starts network service, starts RPC API server, and starts consensus state machine.


### Configurations
Neb uses Protocol Buffer to load configurations. The default config file is named as config.pb.txt and looks like following:
```
p2p {
  # seed: "UNCOMMENT_AND_SET_SEED_NODE_ADDRESS"
  port: 51413
}

rpc {
  port: 51510
}

pow {
  coinbase: "8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf"
}

account {
  # keystore.SECP256K1 = 1
  signature: 1

  # keystore.SCRYPT = 1 << 4
  encrypt: 16

  key_dir: "testKey"

  test_passphrase: "passphrase"
}

```

The configuration schema is defined in proto _neblet/pb/config.proto:Config_. To load a different config file when starting the neb binary, use flag -c. For example,
```
./neb -c config1.pb.txt
```

Neb supports loading KeyStore file of ethereum format. KeyStore files from config _key_dir_ are loaded during neb bootstrap. Example testing KeyStore looks like
```
{"version":3,"id":"272a46f1-5141-4234-b948-1b45c6708962","address":"555fcb1b7051d3aea5cf2c0167b4e19ed6a4f98d","Crypto":{"ciphertext":"ecd4b817fa9ebed736235476c91dec43e73e0ca3e8d2f13c004725349882fb49","cipherparams":{"iv":"1ab4ed89c95f66e994f183fed23df9f9"},"cipher":"aes-128-ctr","kdf":"scrypt","kdfparams":{"dklen":32,"salt":"baef3f92cdde9fd97a00879ce060763101530e9e66e4c75ec74352a41419bde0","n":1024,"r":8,"p":1},"mac":"d8ea471cea8184fb7b19c1563804b85a31a2b3d792dc59ecccdb15dbfb3cebc0"}}

```

### Run non-seed node
Now we can get the seed address from the seed node log output above. Get pretty id from log starts with **"makeHost: boot node pretty id is "**, get address from log starts with **"Start: node info {id -> <peer.ID ....>, address -> [/ip4...]"**. The seed address from log above is
```
/ip4/192.168.1.18/tcp/51413/ipfs/QmVMamgHNoR8GBUbq4r9BTFzod5F2BvhSgaGLVxPpXcWNm
```
To start a non-seed node on another machine, we need to update p2p seed configuration in _config.pb.txt_:
```
p2p {
  seed: "/ip4/192.168.1.18/tcp/51413/ipfs/QmVMamgHNoR8GBUbq4r9BTFzod5F2BvhSgaGLVxPpXcWNm"
  port: 51413
}
...
```

Note, if the non-seed node is running on the same machine, we need to use a different config file, and also assign different P2P and RPC server ports.

Now we can start the non-seed nodes by simply run command
```
./neb
```

The binary will join Nebulas network and connects to the seed node started earlier. The log output will look like:
```
...
INFO[2017-10-18T03:59:23+08:00] SayHello: bootNode addr -> /ip4/192.168.1.18/tcp/51413  file=node.go func="p2p.(*NetService).SayHello" line=162
INFO[2017-10-18T03:59:23+08:00] SayHello: nnode.host.Addrs -> /ip4/1192.168.1.18/tcp/10001, bootAddr -> /ip4/172.31.8.110/tcp/51413  file=node.go func="p2p.(*NetService).SayHello" line=172
INFO[2017-10-18T03:59:23+08:00] Hello: say hello addrs -> [/ip4/192.168.1.18/tcp/51413]  file=net_service.go func="p2p.(*NetService).Hello" line=427
INFO[2017-10-18T03:59:23+08:00] SayHello: node say hello to boot node success...   file=node.go func="p2p.(*NetService).SayHello" line=186
...
```


### RPC API
We can use API _GetAccountState_ and _SendTransaction_ to check and modify account state. We can play the RPC example testing client code:
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

Now we have provided HTTP to access the RPC API.You first need to generate a mapping from GRPC to HTTP:
```
cd rpc/pb
make all
```
The file that ends with gw.go is the mapping file.
Now we can access the rpc API directly from our browser.
For http GET request:
```
GetNebState:
http://localhost:8080/v1/neb/state

Accounts:
http://localhost:8080/v1/accounts

GetAccountState:
http://localhost:8080/v1/account/state
```
For http POST request:
```
BlockDump:
curl -i -H Accept:application/json -X POST http://localhost:8080/v1/block/dump -H Content-Type: application/json -d '{"count":10}'
```
Now you can even create & deploy & call smart contracts directly over HTTP just by 'SendTransaction()'.
If you want to create & deploy a smart contracts:
1. create your smart contracts source.
2. call 'SendTransaction()', the params from and to must be the same.
3. params 'function' must be nil.
```
curl -i -H Accept:application/json -X POST http://localhost:8080/v1/transaction -H Content-Type: application/json -d '{"from":"0x8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf","to":"0x8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf","nonce":1,"source":"'use strict';var SampleContract = function () {LocalContractStorage.defineProperties(this, {name: null,count: null});LocalContractStorage.defineMapProperty(this, \"allocation\");};SampleContract.prototype = {init: function (name, count, allocation) {this.name = name;this.count = count;allocation.forEach(function (item) {this.allocation.put(item.name, item.count);}, this);},dump: function () {console.log('dump: this.name = ' + this.name);console.log('dump: this.count = ' + this.count);},verify: function (expectedName, expectedCount, expectedAllocation) {if (!Object.is(this.name, expectedName)) {throw new Error(\"name is not the same, expecting \" + expectedName + \", actual is \" + this.name + \".\");}if (!Object.is(this.count, expectedCount)) {throw new Error(\"count is not the same, expecting \" + expectedCount + \", actual is \" + this.count + \".\");}expectedAllocation.forEach(function (expectedItem) {var count = this.allocation.get(expectedItem.name);if (!Object.is(count, expectedItem.count)) {throw new Error(\"count of \" + expectedItem.name + \" is not the same, expecting \" + expectedItem.count + \", actual is \" + count + \".\");}}, this);}};module.exports = SampleContract;", "args":"[\"TEST001\", 123,[{\"name\":\"robin\",\"count\":2},{\"name\":\"roy\",\"count\":3},{\"name\":\"leon\",\"count\":4}]]"}'
```
If you succeed in deploying a smart contract, you will get the contract address & transaction hash as response.
Then you can call this samrt contract:
1. get the smart contract address.
2. give the method you want to call.
```
curl -i -H Accept:application/json -X POST http://localhost:8080/v1/transaction -H Content-Type: application/json -d '{"from":"0x8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf","to":"8f5aad7e7ad59c9d9eaa351b3f41f887e49d13f37974a02c", "nonce":2,"function":"dump"}'
```

### Debug chain status from log

The chain will be dump when a new block is minted, or received and added to tail, we can easily find it in log, starting with **"Dump"**:

```
...
Dump:  --> {2, hash: 0000009a658e094ec9bb642ba4d0d0ada63fe70aee5796b0631f9c844da4f5f0, parent: 0000000000000000000000000000000000000000000000000000000000000000, stateRoot: 3b55bd09bff8718070a341455faf21d1d32a1fde5cc3ddec990a553859efb002, coinbase: 8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf} --> {1, hash: 0000000000000000000000000000000000000000000000000000000000000000, parent: 0000000000000000000000000000000000000000000000000000000000000000, stateRoot: , coinbase: 000000000000000000000000000000000000000000000000}
...
```

## Contribution

We are very glad that you are considering to help Nebulas Team or go-nebulas project, including but not limited to source code, documents or others.

If you'd like to contribute, please fork, fix, commit and send a pull request for the maintainers to review and merge into the main code base. If you wish to submit more complex changes though, please check up with the core devs first on our [slack channel](http://nebulasio.herokuapp.com) to ensure those changes are in line with the general philosophy of the project and/or get some early feedback which can make both your efforts much lighter as well as our review and merge procedures quick and simple.

Please refer to our [contribution guideline](https://github.com/nebulasio/wiki/blob/master/contribute.md) for more information.

Thanks.

## License

The go-nebulas project is licensed under the [GNU Lesser General Public License Version 3.0 (“LGPL v3”)](https://www.gnu.org/licenses/lgpl-3.0.en.html).

For the more information about licensing, please refer to [Licensing](https://github.com/nebulasio/wiki/blob/master/licensing.md) page.
