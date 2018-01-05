## Neb test kit

` nebtestkit ` is a testing framework based on [mocha](https://github.com/mochajs/mocha). we can use it to build a private block chain or join an exist network. we also can use it to send a transation or deploy & call a smart contract.

## Usage

### Install dependencies
```
npm install -g mocha
npm install
```
### Write test suite
Start a seed server:
```javascript
var seed = new Neblet("192.168.1.25", 51413, 8191);
var seedJsAgent = seed.RPC();
seed.Init();
var nebSeed = seed.Start();
```

Start a common server and connect to seed server:
```javascript
var server = new Neblet("192.168.1.25", 51414, 8192);
var jsAgent = server.RPC();
server.Init(seed);
var neb = server.Start();
```
Because RPC server is 3 seconds later than neblet server. so we should wait for several seconds before run our test suite.
```javascript
    before(function(done) {
        this.timeout(10000);
        setTimeout(done, 8000);
    });
```


### Run test suite
Fist fo all, we shold copy the binary file 'neb' to nebtestkit. And then run:
```sh
$ mocha neblet.test.js

seed server A test suite
    ✓ start server A (104ms)
    ✓ get accounts info from server A (100ms)
    ✓ get account B balance from server A (101ms)
    ✓ unlock account A from server A (138ms)
    ✓ transfer 10 from account A to B (2187ms)
    ✓ verify transaction from server A (95ms)
    ✓ verify account B balance from server A (96ms)

  server B test suite
    ✓ start server B & connect to server A (92ms)
    ✓ verify transaction from server B (95ms)
    ✓ verify account balance from server B (97ms)

  quit
    ✓ quit

```
