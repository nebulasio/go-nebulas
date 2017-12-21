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

var coinbase = '1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c';
var addresses = ['333cb3ed8c417971845382ede3cf67a0a96270c05fe2f700'];
var nebArray = [nebSeed, nebA, nebB];

var txhash;
var txhashArr = new Array();
var loop = 0;
describe('seed server test suite', function () {
    before(function (done) {
        this.timeout(6000);
        setTimeout(done, 5000);
    });
    it('start seed server', function () {
        var nodeinfo = seedJsAgent.api.nodeInfo();
        expect(nodeinfo.id).to.be.equal('QmPyr4ZbDmwF1nWxymTktdzspcBFPL6X1v3Q5nT7PGNtUN');
        expect(nodeinfo.chain_id).to.be.equal(100);
    });
    it('get accounts info from seed server', function () {
        var accounts = seedJsAgent.api.accounts();
        expect(accounts.addresses).to.be.have.contains(coinbase);
    });
});

describe('Server A test suite', function () {
    before(function (done) {
        this.timeout(10000);
        setTimeout(done, 8000);
    });
    it('start Server A & connect to seed Server', function () {
        var nodeinfo = jsAgentA.api.nodeInfo();
        expect(nodeinfo.chain_id).to.be.equal(100);
        expect(nodeinfo.route_table[0]).to.be.have.property('id').equals('QmPyr4ZbDmwF1nWxymTktdzspcBFPL6X1v3Q5nT7PGNtUN');
    });

    it('get accounts balance from server A', function () {
        for (var i = 0; i < addresses.length; i++) {
            var accountState = jsAgentA.api.getAccountState(addresses[i]);
            expect(accountState).to.be.have.property('balance').eq('0');
        }
    });

    it('unlock account A from server A', function () {
        var result = jsAgentA.admin.unlockAccount(coinbase, 'passphrase');
        expect(result).to.be.have.property('result').eq(true);
    });

    // A transfer to B 10.
    for (var i = 0; i < addresses.length; i++) {
        it('transfer 10 from account A to ' + addresses[i], function (done) {
            this.timeout(20000);
            var tx;
            var timeout;
            var txhash = jsAgentA.api.sendTransaction(coinbase, addresses[loop], 10, loop + 1);
            txhashArr[loop] = txhash;
            loop++;
            timeout = setInterval(function () {
                tx = jsAgentA.api.getTransactionReceipt(txhash.txhash);
                if (tx.error == undefined) {
                    expect(txhash).to.be.have.property('txhash');
                    clearInterval(timeout);
                    done();
                }
            }, 2000);
        });
    }


    // query transaction by txhash.
    it('verify transaction from server A', function () {
        for (var i = 0; i < txhashArr.length; i++) {
            var tx = jsAgentA.api.getTransactionReceipt(txhashArr[i].txhash);
            expect(tx).to.be.have.property('from').equals(coinbase);
            expect(tx).to.be.have.property('hash').equals(txhashArr[i].txhash);
            expect(tx).to.be.have.property('nonce').equals((i + 1).toString());
        }
    });

    it('verify all account balance from server A', function () {
        for (var i = 0; i < addresses.length; i++) {
            var accountState = jsAgentA.api.getAccountState(addresses[i]);
            expect(accountState).to.be.have.property('balance').eq('10');
        }
    });
});

describe('Server B test suite', function () {
    it('start Server B & connect to seed Server', function () {
        var nodeinfo = jsAgentB.api.nodeInfo();
        expect(nodeinfo.chain_id).to.be.equal(100);
        expect(nodeinfo.route_table[0]).to.be.have.property('id').equals('QmPyr4ZbDmwF1nWxymTktdzspcBFPL6X1v3Q5nT7PGNtUN');
    });

    it('verify transaction from seed Server', function () {
        for (var i = 0; i < txhashArr.length; i++) {
            var tx = jsAgentB.api.getTransactionReceipt(txhashArr[i].txhash);
            expect(tx).to.be.have.property('from').equals(coinbase);
            expect(tx).to.be.have.property('hash').equals(txhashArr[i].txhash);
            expect(tx).to.be.have.property('nonce').equals((i + 1).toString());
        }
    });

    it('verify account balance from seed Server', function () {
        var accountState = jsAgentB.api.getAccountState('333cb3ed8c417971845382ede3cf67a0a96270c05fe2f700');
        expect(accountState).to.be.have.property('balance').eq('10');
    });
});


describe('quit', function () {
    it('quit', function () {
        for (var i = 0; i < nebArray.length; i++) {
            nebArray[i].kill('SIGINT');
        }
    });
});