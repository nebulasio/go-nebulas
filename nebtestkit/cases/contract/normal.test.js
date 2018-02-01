'use strict';

var TestnetNodes = require('../../testnet-nodes');
var expect = require('chai').expect;
var BigNumber = require('bignumber.js');

var nodes = new TestnetNodes();
nodes.Start();

var from = "1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c";
var to = "2fe3f9f51f9a05dd5f7c5329127f7c917917149b4e16b0b8";

function checkTransaction(hash, done, count) {
    if (count > 4) {
        console.log("tx receipt timeout:" + hash);
        done();
        return;
    }

    nodes.RPC(0).api.getTransactionReceipt(hash).then(function (resp) {

        console.log("tx receipt:" + JSON.stringify(resp));
        return nodes.RPC(0).api.getAccountState(from);
    }).then(function (resp) {
        console.log("after state:" + JSON.stringify(resp));
        done();
    }).catch(function (err) {
        setTimeout(function () {
            checkTransaction(hash, done, count + 1);
        }, 2000);
    });
}

describe('normal transaction', function () {
    before(function (done) {
        this.timeout(1000000);
        setTimeout(done, 5000);
    });

    it('1.normal transfer', function (done) {

        nodes.RPC(0).api.getAccountState(from).then(function (resp) {

            console.log("resp:" + JSON.stringify(resp));
            return nodes.SendTransaction(from, to, "1", parseInt(resp.nonce) + 1);
        }).then(function (resp) {

            console.log("send:" + JSON.stringify(resp));
            expect(resp).to.be.have.property('txhash');
            checkTransaction(resp.txhash, done, 1);
            // done();
        }).catch(function (err) {
            console.log(JSON.stringify(err.error));
            done();
        });
    });

    it('2.from & to are same', function (done) {

        nodes.RPC(0).api.getAccountState(from).then(function (resp) {

            console.log("resp:" + JSON.stringify(resp));
            return nodes.SendTransaction(from, from, "1", parseInt(resp.nonce) + 1)
        }).then(function (resp) {

            console.log("send:" + JSON.stringify(resp));
            expect(resp).to.be.have.property('txhash');
            checkTransaction(resp.txhash, done, 1);
            // done();
        }).catch(function (err) {
            console.log(JSON.stringify(err.error));
            done();
        });
    });

    it('3.from balance is insufficient', function (done) {

        nodes.RPC(0).api.getAccountState(from).then(function (resp) {

            console.log("resp:" + JSON.stringify(resp));
            var value = new BigNumber(resp.balance).add("1");
            return nodes.SendTransaction(from, to, value.toString(), parseInt(resp.nonce) + 1);
        }).then(function (resp) {

            console.log("send:" + JSON.stringify(resp));
            expect(resp).to.be.have.property('txhash');
            checkTransaction(resp.txhash, done, 1);
            // done();
        }).catch(function (err) {
            console.log(JSON.stringify(err.error));
            done();
        });
    });

    it('4.gas is insufficient', function (done) {

        nodes.RPC(0).api.getAccountState(from).then(function (resp) {

            console.log("resp:" + JSON.stringify(resp));
            return nodes.SendTransaction(from, to, resp.balance, parseInt(resp.nonce) + 1, "0", "1");
        }).then(function (resp) {

            console.log("send:" + JSON.stringify(resp));
            expect(resp).to.be.have.property('txhash');
            checkTransaction(resp.txhash, done, 1);
            // done();
        }).catch(function (err) {
            console.log(JSON.stringify(err.error));
            done();
        });
    });

    it('5.from is invalid address', function (done) {

        nodes.RPC(0).api.getAccountState(from).then(function (resp) {

            console.log("resp:" + JSON.stringify(resp));
            return nodes.SendTransaction("0x00", to, resp.balance, parseInt(resp.nonce) + 1)
        }).then(function (resp) {

            console.log("send:" + JSON.stringify(resp));
            expect(resp).to.be.have.property('txhash');
            checkTransaction(resp.txhash, done, 1);
            // done();
        }).catch(function (err) {
            console.log(JSON.stringify(err.error));
            done();
        });
    });

    it('6.to is invalid address', function (done) {

        nodes.RPC(0).api.getAccountState(from).then(function (resp) {

            console.log("resp:" + JSON.stringify(resp));
            return nodes.SendTransaction(from, "0x00", resp.balance, parseInt(resp.nonce) + 1);
        }).then(function (resp) {

            console.log("send:" + JSON.stringify(resp));
            expect(resp).to.be.have.property('txhash');
            checkTransaction(resp.txhash, done, 1);
            // done();
        }).catch(function (err) {
            console.log(JSON.stringify(err.error));
            done();
        });
    });

    it('7.nonce is below', function (done) {

        nodes.RPC(0).api.getAccountState(from).then(function (resp) {

            console.log("resp:" + JSON.stringify(resp));
            return nodes.SendTransaction(from, to, resp.balance, parseInt(resp.nonce));
        }).then(function (resp) {

            console.log("send:" + JSON.stringify(resp));
            expect(resp).to.be.have.property('txhash');
            checkTransaction(resp.txhash, done, 1);
            // done();
        }).catch(function (err) {
            console.log(JSON.stringify(err.error));
            done();
        });
    });

    it('8.nonce is heigher', function (done) {

        nodes.RPC(0).api.getAccountState(from).then(function (resp) {

            console.log("resp:" + JSON.stringify(resp));
            var nonce = new BigNumber(resp.nonce).add(2);
            return nodes.SendTransaction(from, to, "1", nonce.toNumber());
        }).then(function (resp) {

            console.log("send:" + JSON.stringify(resp));
            expect(resp).to.be.have.property('txhash');
            checkTransaction(resp.txhash, done, 1);
            // done();
        }).catch(function (err) {
            console.log(JSON.stringify(err.error));
            done();
        });
    });

    it('9.gasPrice is below', function (done) {

        nodes.RPC(0).api.getAccountState(from).then(function (resp) {

            console.log("resp:" + JSON.stringify(resp));
            return nodes.SendTransaction(from, to, resp.balance, parseInt(resp.nonce) + 1, "1");
        }).then(function (resp) {

            console.log("send:" + JSON.stringify(resp));
            expect(resp).to.be.have.property('txhash');
            checkTransaction(resp.txhash, done, 1);
            // done();
        }).catch(function (err) {
            console.log(JSON.stringify(err.error));
            done();
        });
    });

    it('10.gas is higher than max', function (done) {

        nodes.RPC(0).api.getAccountState(from).then(function (resp) {

            console.log("resp:" + JSON.stringify(resp));
            var maxGas = new BigNumber(10).pow(9).mul(60);
            return nodes.SendTransaction(from, to, resp.balance, parseInt(resp.nonce) + 1, "", maxGas.toString());
        }).then(function (resp) {

            console.log("send:" + JSON.stringify(resp));
            expect(resp).to.be.have.property('txhash');
            checkTransaction(resp.txhash, done, 1);
            // done();
        }).catch(function (err) {
            console.log(JSON.stringify(err.error));
            done();
        });
    });

    it('quit', function () {
        nodes.Stop();
    });
});
