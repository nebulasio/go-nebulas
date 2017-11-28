'use strict';

var Neblet = require('./neblet');
var os = require('os');
var expect = require('chai').expect;
var ip = getLocalIP();

// start a seed server.
var seed = new Neblet(ip, 51413, 8191);
var seedJsAgent = seed.NebJs();
seed.Init();
var nebSeed = seed.Start();

// start a non-seed server.
var server = new Neblet(ip, 10000, 9000);
var jsAgent = server.NebJs();
server.Init(seed);
var neb = server.Start();


 

var txhash;
describe('seed server A test suite', function() {
    before(function(done) {
        this.timeout(10000);
        setTimeout(done, 8000);
      });
    it('start server A', function() {
        var nodeinfo = seedJsAgent.api.nodeInfo();
        expect(nodeinfo.id).to.be.equal('QmPyr4ZbDmwF1nWxymTktdzspcBFPL6X1v3Q5nT7PGNtUN');
        expect(nodeinfo.chain_id).to.be.equal(100);
        expect(nodeinfo.version).to.be.equal(1);
    });
    it('get accounts info from server A', function() {
        var accounts = seedJsAgent.api.accounts();
        expect(accounts.addresses).to.be.have.length(3);
        expect(accounts.addresses).to.be.have.contains('8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf');
    });

    it('get account B balance from server A', function() {
        var accountState = seedJsAgent.api.getAccountState('22ac3a9a2b1c31b7a9084e46eae16e761f83f02324092b09');
        expect(accountState).to.be.have.property('balance').eq('0');
    });

    it('unlock account A from server A', function() {
        var result = seedJsAgent.admin.unlockAccount('8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf', 'passphrase');
        expect(result).to.be.have.property('result').eq(true);
    });

    // A transfer to B 10.
    it('transfer 10 from account A to B', function(done) {
        this.timeout(8000);
        txhash = seedJsAgent.api.sendTransaction('8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf', '22ac3a9a2b1c31b7a9084e46eae16e761f83f02324092b09', 10, 1, '', '', '', '');
        var tx;
        var timeout;
        timeout = setInterval(function() {
            tx = seedJsAgent.api.getTransactionReceipt(txhash.hash);
            if (tx.error == undefined) {
                expect(txhash).to.be.have.property('hash');
                clearInterval(timeout);
                done();
            }
        }, 2000);
    });

    // query transaction by txhash.
    it('verify transaction from server A', function() {
        var tx = seedJsAgent.api.getTransactionReceipt(txhash.hash);
        expect(tx).to.be.have.property('from').equals('8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf');
        expect(tx).to.be.have.property('to').equals('22ac3a9a2b1c31b7a9084e46eae16e761f83f02324092b09');
    });

    it('verify account B balance from server A', function() {
        var accountState = seedJsAgent.api.getAccountState('22ac3a9a2b1c31b7a9084e46eae16e761f83f02324092b09');
        expect(accountState).to.be.have.property('balance').eq('10');
    });
});

describe('server B test suite', function(){
    it('start server B & connect to server A', function() {
        var nodeinfo = jsAgent.api.nodeInfo();
        expect(nodeinfo.chain_id).to.be.equal(100);
        expect(nodeinfo.version).to.be.equal(1);
        expect(nodeinfo.peer_count).to.be.eq(1);
        expect(nodeinfo.route_table[0]).to.be.have.property('id').equals('QmPyr4ZbDmwF1nWxymTktdzspcBFPL6X1v3Q5nT7PGNtUN');
    });

    it('verify transaction from server B', function() {
        var tx = jsAgent.api.getTransactionReceipt(txhash.hash);
        expect(tx).to.be.have.property('from').equals('8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf');
        expect(tx).to.be.have.property('to').equals('22ac3a9a2b1c31b7a9084e46eae16e761f83f02324092b09');
    });

    it('verify account balance from server B', function() {
        var accountState = jsAgent.api.getAccountState('22ac3a9a2b1c31b7a9084e46eae16e761f83f02324092b09');
        expect(accountState).to.be.have.property('balance').eq('10');
    });
})

describe('quit', function(){
    it('quit', function() {
        nebSeed.kill('SIGINT');
        neb.kill('SIGINT');
    });
})

function getLocalIP() {
    var interfaces = os.networkInterfaces();
    for (var k in interfaces) {
        for (var k2 in interfaces[k]) {
            var address = interfaces[k][k2];
            if (address.family === 'IPv4' && !address.internal) {
                return address.address;
            }
        }
    }
    return null;
}