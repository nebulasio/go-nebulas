# go-nebulas

Official Go implementation of the Nebulas protocol.

For more information of Nebulas protocol, design documents, please refer to our [wiki](https://github.com/nebulasio/wiki).

## Building from source

### Prerequisites

| Components | Version | Description |
|----------|-------------|-------------|
|[Golang](https://golang.org) | >= 1.8| The Go Programming Language |
[Dep](https://github.com/golang/dep) | >= 0.3.1 | Dep is a dependency management tool for Go. |

### Building

1. install dependencies packages, run
```
make dep
```

2. build the neb application, run
```
make build
```

## Running

For the version 0.1.0, *neb* supports two running mode, Network and Dummy mode.

### Dummy mode

In Dummy mode, *neb* creates 4 client thread (goroutine) to simulate multi-clients, run
```
./neb --dummy
```

### Network mode

Running in Network mode, we first need to run a seed node by run
```
$ ./neb
INFO[2017-09-27T17:35:58+08:00] NewNode: node make Host success               file=node.go func=p2p.NewNode line=60
INFO[2017-09-27T17:35:58+08:00] makeHost: boot node pretty id is QmcvYFjy2d5Kt1DcmTJJEUKsKCXg8WJsCGRDD2bgyRYk6p  file=node.go func="p2p.(*Node).makeHost" line=165
INFO[2017-09-27T17:35:58+08:00] RegisterBlockMsgService: node register block message service success...  file=block.go func="p2p.(*Manager).RegisterBlockMsgService" line=46
chainID is 1
INFO[2017-09-27T17:35:58+08:00] Start: node create success...                 file=node.go func="p2p.(*Node).Start" line=71
INFO[2017-09-27T17:35:58+08:00] Start: node info {id -> <peer.ID cvYFjy>, address -> [/ip4/192.168.1.159/tcp/9999]}  file=node.go func="p2p.(*Node).Start" line=74
INFO[2017-09-27T17:35:58+08:00] Start: node start join p2p network...         file=node.go func="p2p.(*Node).Start" line=79
INFO[2017-09-27T17:35:58+08:00] RegisterPingService: node register ping service success...  file=ping.go func="p2p.(*Node).RegisterPingService" line=52
INFO[2017-09-27T17:35:58+08:00] RegisterLookupService: node register lookup service success...  file=lookup.go func="p2p.(*Node).RegisterLookupService" line=47
INFO[2017-09-27T17:35:58+08:00] Start: node start and join to p2p network success and listening for connections on port 9999...   file=node.go func="p2p.(*Node).Start" line=104
DEBU[2017-09-27T17:35:58+08:00] running.                                      file=asm_amd64.s func=runtime.goexit line=2338
DEBU[2017-09-27T17:35:58+08:00] PrepareState enter.                           file=prepare.go func="pow.(*PrepareState).Enter" line=79
```

Now we can get the seed address from log, get pretty id from log starts with **"makeHost: boot node pretty id is "**, get address from log starts with **"Start: node info {id -> <peer.ID ....>, address -> [/ip4...]"**. The seed address from log above is
```
/ip4/192.168.1.159/tcp/9999/ipfs/QmcvYFjy2d5Kt1DcmTJJEUKsKCXg8WJsCGRDD2bgyRYk6p
```

Then we start other nodes, run
```
./neb -s [seed address] -p [your custom port]
```
Example,
```
./neb -s "/ip4/192.168.1.159/tcp/9999/ipfs/QmcvYFjy2d5Kt1DcmTJJEUKsKCXg8WJsCGRDD2bgyRYk6p" -p 10001
./neb -s "/ip4/192.168.1.159/tcp/9999/ipfs/QmcvYFjy2d5Kt1DcmTJJEUKsKCXg8WJsCGRDD2bgyRYk6p" -p 10002
./neb -s "/ip4/192.168.1.159/tcp/9999/ipfs/QmcvYFjy2d5Kt1DcmTJJEUKsKCXg8WJsCGRDD2bgyRYk6p" -p 10003
```

And, in v0.1.0, the **Sync Protocol** is not implemented. So the non-seed nodes should be start after seed node start, before seed node finds the first nonce, minted the first block. Otherwise the Chain will be forked at the very beginning.


### Reading the Log

The Chain will be dump when new block is minted or received and put to tail, you can easily find in log, starting with **"Dump"**:

```
...
DEBU[2017-09-27T17:36:01+08:00] Dump:  --> {2, hash: 0000002f3a7f6887c1186cb351174b9089818798895c4c43e00c2c58b7980005, parent: 0000000000000000000000000000000000000000000000000000000000000000, stateRoot: 899d499ff2077abe53c88ca3289dfbbc8f8fdfb9d34c47a3830a13dba739fb69} --> {1, hash: 0000000000000000000000000000000000000000000000000000000000000000, parent: 0000000000000000000000000000000000000000000000000000000000000000, stateRoot: }  file=fork_choice.go func="pow.(*Pow).ForkChoice" line=63
...
```

## Contribution

We are very glad that you are considering to help Nebulas Team or go-nebulas project, including but not limited to source code, documents or others.

If you'd like to contribute, please fork, fix, commit and send a pull request for the maintainers to review and merge into the main code base. If you wish to submit more complex changes though, please check up with the core devs first on our [slack channel](http://nebulasio.herokuapp.com) to ensure those changes are in line with the general philosophy of the project and/or get some early feedback which can make both your efforts much lighter as well as our review and merge procedures quick and simple.

Please refer to our [contribution guideline](https://github.com/nebulasio/wiki/blob/master/contribute.md) for more information.

Thanks.

## License

The go-nebulas project is licensed under the [GNU Lesser General Public License Version 3.0 (“LGPL v3”)](https://www.gnu.org/licenses/lgpl-3.0.en.html).

For the more information about licensing, please refer to [Licensing Page](https://github.com/nebulasio/wiki/blob/master/licensing.md).
