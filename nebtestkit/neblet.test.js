'use strict';

var Neblet = require('./neblet');
var os = require('os');
var expect = require('chai').expect;
var ip = "127.0.0.1";

// start a seed server.
var seed = new Neblet(ip, 51413, 8191);
var seedJsAgent = seed.NebJs();
seed.Init();
var nebSeed = seed.Start();

// start a non-seed server.
var serverA = new Neblet(ip, 10000, 9000);
var jsAgentA = serverA.NebJs();
serverA.Init(seed);
var nebA = serverA.Start();

var serverB = new Neblet(ip, 10001, 9001);
var jsAgentB = serverB.NebJs();
serverB.Init(seed);
var nebB = serverB.Start();

var serverC = new Neblet(ip, 10002, 9002);
var jsAgentC = serverC.NebJs();
serverC.Init(seed);
var nebC = serverC.Start();

var coinbase = 'eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8';
var addresses = ['5cdadc1cfe3da0a3d067e9f1b195b90c5aebfb5afc8d43b4', '22ac3a9a2b1c31b7a9084e46eae16e761f83f02324092b09', '88c1573761de0b48503536a0d60f056a08ea2e3cdc947f3f', 'a8565ee007ebbdfabdc9c0c32f931a7f5416eff7b2fbd5cd'];
var nebArray = [nebSeed, nebA, nebB, nebC];

var txhash;
var txhashArr = new Array();
var loop = 0;
describe('seed server A test suite', function() {
    before(function(done) {
        this.timeout(10000);
        setTimeout(done, 8000);
      });
    it('start server A', function() {
        var nodeinfo = seedJsAgent.api.nodeInfo();
        expect(nodeinfo.id).to.be.equal('QmPyr4ZbDmwF1nWxymTktdzspcBFPL6X1v3Q5nT7PGNtUN');
        expect(nodeinfo.chain_id).to.be.equal(100);
    });
    it('get accounts info from seed server', function() {
        var accounts = seedJsAgent.api.accounts();
        expect(accounts.addresses).to.be.have.length(5);
        expect(accounts.addresses).to.be.have.contains(coinbase);
    });

    it('get accounts balance from seed server', function() {
        for (var i = 0;i<addresses.length;i++) {
            var accountState = seedJsAgent.api.getAccountState(addresses[i]);
            expect(accountState).to.be.have.property('balance').eq('0');
        }
    });

    it('unlock account A from seed server', function() {
        var result = seedJsAgent.admin.unlockAccount(coinbase, 'zaq12wsx');
        expect(result).to.be.have.property('result').eq(true);
    });

    // A transfer to B 10.
    for (var i = 0;i < addresses.length;i++) {  
        it('transfer 10 from account A to ' + addresses[i], function(done) {
            this.timeout(10000);
            var tx;
            var timeout;
            var txhash = seedJsAgent.api.sendTransaction(coinbase, addresses[loop], 10, loop+1);
            txhashArr[loop] = txhash;
            loop++;
            timeout = setInterval(function() {
                tx = seedJsAgent.api.getTransactionReceipt(txhash.txhash);
                if (tx.error == undefined) {
                    expect(txhash).to.be.have.property('txhash');
                    clearInterval(timeout);
                    done();
                }
            }, 2000);
        });
    }
    

    // query transaction by txhash.
    it('verify transaction from seed server', function() {
        for (var i=0;i<txhashArr.length;i++) {
            var tx = seedJsAgent.api.getTransactionReceipt(txhashArr[i].txhash);
            expect(tx).to.be.have.property('from').equals(coinbase);
            expect(tx).to.be.have.property('hash').equals(txhashArr[i].txhash);
            expect(tx).to.be.have.property('nonce').equals((i+1).toString());
        }
    });

    it('verify all account balance from seed server', function() {
        for (var i=0;i<addresses.length;i++) {
            var accountState = seedJsAgent.api.getAccountState(addresses[i]);
            expect(accountState).to.be.have.property('balance').eq('10');
        }
    });
});

describe('Server A test suite', function(){
    it('start Server A & connect to seed Server', function() {
        var nodeinfo = jsAgentA.api.nodeInfo();
        expect(nodeinfo.chain_id).to.be.equal(100);
        expect(nodeinfo.route_table[0]).to.be.have.property('id').equals('QmPyr4ZbDmwF1nWxymTktdzspcBFPL6X1v3Q5nT7PGNtUN');
    });

    it('verify transaction from seed Server', function() {
        for (var i=0;i<txhashArr.length;i++) {
            var tx = jsAgentA.api.getTransactionReceipt(txhashArr[i].txhash);
            expect(tx).to.be.have.property('from').equals(coinbase);
            expect(tx).to.be.have.property('hash').equals(txhashArr[i].txhash);
            expect(tx).to.be.have.property('nonce').equals((i+1).toString());
        }
    });

    it('verify account balance from seed Server', function() {
        var accountState = jsAgentA.api.getAccountState('5cdadc1cfe3da0a3d067e9f1b195b90c5aebfb5afc8d43b4');
        expect(accountState).to.be.have.property('balance').eq('10');
    });
});

describe('Server B test suite', function(){
    it('start Server B & connect to seed Server', function() {
        var nodeinfo = jsAgentB.api.nodeInfo();
        expect(nodeinfo.chain_id).to.be.equal(100);
        expect(nodeinfo.route_table[0]).to.be.have.property('id').equals('QmPyr4ZbDmwF1nWxymTktdzspcBFPL6X1v3Q5nT7PGNtUN');
    });

    it('verify transaction from seed Server', function() {
        for (var i=0;i<txhashArr.length;i++) {
            var tx = jsAgentB.api.getTransactionReceipt(txhashArr[i].txhash);
            expect(tx).to.be.have.property('from').equals(coinbase);
            expect(tx).to.be.have.property('hash').equals(txhashArr[i].txhash);
            expect(tx).to.be.have.property('nonce').equals((i+1).toString());
        }
    });

    it('verify account balance from seed Server', function() {
        var accountState = jsAgentB.api.getAccountState('5cdadc1cfe3da0a3d067e9f1b195b90c5aebfb5afc8d43b4');
        expect(accountState).to.be.have.property('balance').eq('10');
    });
});

describe('Server C test suite', function(){
    it('start Server C & connect to seed Server', function() {
        var nodeinfo = jsAgentC.api.nodeInfo();
        expect(nodeinfo.chain_id).to.be.equal(100);
        expect(nodeinfo.route_table[0]).to.be.have.property('id').equals('QmPyr4ZbDmwF1nWxymTktdzspcBFPL6X1v3Q5nT7PGNtUN');
    });

    it('verify transaction from seed Server', function() {
        for (var i=0;i<txhashArr.length;i++) {
            var tx = jsAgentC.api.getTransactionReceipt(txhashArr[i].txhash);
            expect(tx).to.be.have.property('from').equals(coinbase);
            expect(tx).to.be.have.property('hash').equals(txhashArr[i].txhash);
            expect(tx).to.be.have.property('nonce').equals((i+1).toString());
        }
    });

    it('verify account balance from seed Server', function() {
        var accountState = jsAgentC.api.getAccountState('5cdadc1cfe3da0a3d067e9f1b195b90c5aebfb5afc8d43b4');
        expect(accountState).to.be.have.property('balance').eq('10');
    });
});


describe('quit', function(){
    it('quit', function() {
        for (var i = 0; i< nebArray.length; i++) {
            nebArray[i].kill('SIGINT');
        }
    });
});