## Neb test kit

` nebtestkit ` is a testing framework based on [mocha](https://github.com/mochajs/mocha). we can use it to build a private block chain or join an exist network. we also can use it to send a transation or deploy & call a smart contract.

## Usage

### Install dependencies
```
npm install -g mocha
npm install
```

### Run test suite
Fist fo all, we shold move the binary file 'neb' to nebtestkit. And then run:
```
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
