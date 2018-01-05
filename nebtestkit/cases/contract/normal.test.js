'use strict';

var TestnetNodes = require('../../testnet-nodes');
var expect = require('chai').expect;
var BigNumber = require('bignumber.js');

var nodes = new TestnetNodes(6);
nodes.Start();

function checkTransaction(hash, done, count) {
    if (count > 6) {
        console.log("tx receipt timeout:" + hash);
        done();
        return;
    }

    var node = nodes.Node(0);
    node.RPC().api.getTransactionReceipt(hash).then(function (resp) {

        console.log("tx receipt:" + JSON.stringify(resp));
        return node.RPC().api.getAccountState(node.Coinbase());
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

        var state;
        var node = nodes.Node(0);
        node.RPC().api.getAccountState(node.Coinbase()).then(function (resp) {

            console.log("resp:" + JSON.stringify(resp));
            state = resp;
            return node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase());
        }).then(function (resp) {

            return node.RPC().api.sendTransaction(node.Coinbase(), nodes.Coinbase(1), "1", parseInt(state.nonce) + 1);
        }).then(function (resp) {

            console.log("send:" + JSON.stringify(resp))
            expect(resp).to.be.have.property('txhash');
            checkTransaction(resp.txhash, done, 1);
            // done();
        }).catch(function (err) {
            console.log("send err:" + JSON.stringify(err.error))
            done();
        });
    });

    it('2.from & to are same', function (done) {

        var state;
        var node = nodes.Node(0);
        node.RPC().api.getAccountState(node.Coinbase()).then(function (resp) {

            console.log("resp:" + JSON.stringify(resp));
            state = resp;
            return node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase());
        }).then(function (resp) {

            return node.RPC().api.sendTransaction(node.Coinbase(), nodes.Coinbase(0), "1", parseInt(state.nonce) + 1)
        }).then(function (resp) {

            console.log("send:" + JSON.stringify(resp))
            expect(resp).to.be.have.property('txhash');
            checkTransaction(resp.txhash, done, 1);
            // done();
        }).catch(function (err) {
            console.log("send err:" + JSON.stringify(err.error))
            done();
        });
    });

    it('3.from balance is insufficient', function (done) {

        var state;
        var node = nodes.Node(0);
        node.RPC().api.getAccountState(node.Coinbase()).then(function (resp) {

            console.log("resp:" + JSON.stringify(resp));
            state = resp;
            return node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase());
        }).then(function (resp) {

            var value = new BigNumber(state.balance).add("1");
            return node.RPC().api.sendTransaction(node.Coinbase(), nodes.Coinbase(0), value.toString(), parseInt(state.nonce) + 1);
        }).then(function (resp) {

            console.log("send:" + JSON.stringify(resp))
            expect(resp).to.be.have.property('txhash');
            checkTransaction(resp.txhash, done, 1);
            // done();
        }).catch(function (err) {
            console.log("send err:" + JSON.stringify(err.error))
            done();
        });
    });

    it('4.gas is insufficient', function (done) {

        var state;
        var node = nodes.Node(0);
        node.RPC().api.getAccountState(node.Coinbase()).then(function (resp) {

            console.log("resp:" + JSON.stringify(resp));
            state = resp;
            return node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase())
        }).then(function (resp) {

            return node.RPC().api.sendTransaction(node.Coinbase(), nodes.Coinbase(0), state.balance, parseInt(state.nonce) + 1, "0", "1");
        }).then(function (resp) {

            console.log("send:" + JSON.stringify(resp))
            expect(resp).to.be.have.property('txhash');
            checkTransaction(resp.txhash, done, 1);
            // done();
        }).catch(function (err) {
            console.log("send err:" + JSON.stringify(err.error))
            done();
        });
    });

    it('5.from is invalid address', function (done) {

        var state;
        var node = nodes.Node(0);
        node.RPC().api.getAccountState(node.Coinbase()).then(function (resp) {

            console.log("resp:" + JSON.stringify(resp));
            state = resp;
            return node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase());
        }).then(function (resp) {

            return node.RPC().api.sendTransaction("0x00", nodes.Coinbase(0), state.balance, parseInt(state.nonce) + 1)
        }).then(function (resp) {

            console.log("send:" + JSON.stringify(resp))
            expect(resp).to.be.have.property('txhash');
            checkTransaction(resp.txhash, done, 1);
            // done();
        }).catch(function (err) {
            console.log("send err:" + JSON.stringify(err.error))
            done();
        });
    });

    it('6.to is invalid address', function (done) {

        var state;
        var node = nodes.Node(0);
        node.RPC().api.getAccountState(node.Coinbase()).then(function (resp) {

            console.log("resp:" + JSON.stringify(resp));
            state = resp;
            return node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase());
        }).then(function (resp) {

            return node.RPC().api.sendTransaction(node.Coinbase(), "0x00", state.balance, parseInt(state.nonce) + 1);
        }).then(function (resp) {

            console.log("send:" + JSON.stringify(resp))
            expect(resp).to.be.have.property('txhash');
            checkTransaction(resp.txhash, done, 1);
            // done();
        }).catch(function (err) {
            console.log("send err:" + JSON.stringify(err.error))
            done();
        });
    });

    it('7.nonce is below', function (done) {

        var state;
        var node = nodes.Node(0);
        node.RPC().api.getAccountState(node.Coinbase()).then(function (resp) {

            console.log("resp:" + JSON.stringify(resp));
            state = resp;
            return node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase());
        }).then(function (resp) {

            return node.RPC().api.sendTransaction(node.Coinbase(), nodes.Coinbase(1), state.balance, parseInt(state.nonce));
        }).then(function (resp) {

            console.log("send:" + JSON.stringify(resp))
            expect(resp).to.be.have.property('txhash');
            checkTransaction(resp.txhash, done, 1);
            // done();
        }).catch(function (err) {
            console.log("send err:" + JSON.stringify(err.error))
            done();
        });
    });

    it('8.nonce is heigher', function (done) {

        var state;
        var node = nodes.Node(0);
        node.RPC().api.getAccountState(node.Coinbase()).then(function (resp) {

            console.log("resp:" + JSON.stringify(resp));
            state = resp;
            return node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase());
        }).then(function (resp) {

            var nonce = new BigNumber(state.nonce).add(2);
            return node.RPC().api.sendTransaction(node.Coinbase(), nodes.Coinbase(1), state.balance, nonce.toNumber());
        }).then(function (resp) {

            console.log("send:" + JSON.stringify(resp))
            expect(resp).to.be.have.property('txhash');
            checkTransaction(resp.txhash, done, 1);
            // done();
        }).catch(function (err) {
            console.log("send err:" + JSON.stringify(err.error))
            done();
        });
    });

    it('9.gasPrice is below', function (done) {

        var state;
        var node = nodes.Node(0);
        node.RPC().api.getAccountState(node.Coinbase()).then(function (resp) {

            console.log("resp:" + JSON.stringify(resp));
            state = resp;
            return node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase());
        }).then(function (resp) {

            return node.RPC().api.sendTransaction(node.Coinbase(), nodes.Coinbase(1), state.balance, parseInt(state.nonce) + 1, "1");
        }).then(function (resp) {

            console.log("send:" + JSON.stringify(resp))
            expect(resp).to.be.have.property('txhash');
            checkTransaction(resp.txhash, done, 1);
            // done();
        }).catch(function (err) {
            console.log("send err:" + JSON.stringify(err.error))
            done();
        });
    });

    it('10.gas is higher than max', function (done) {

        var state;
        var node = nodes.Node(0);
        node.RPC().api.getAccountState(node.Coinbase()).then(function (resp) {

            console.log("resp:" + JSON.stringify(resp));
            state = resp;
            return node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase());
        }).then(function (resp) {

            var maxGas = new BigNumber(10).pow(9).mul(60);
            return node.RPC().api.sendTransaction(node.Coinbase(), nodes.Coinbase(1), state.balance, parseInt(state.nonce) + 1, "", maxGas.toString());
        }).then(function (resp) {

            console.log("send:" + JSON.stringify(resp))
            expect(resp).to.be.have.property('txhash');
            checkTransaction(resp.txhash, done, 1);
            // done();
        }).catch(function (err) {
            console.log("send err:" + JSON.stringify(err.error))
            done();
        });
    });

    it('quit', function () {
        nodes.Stop();
    });
});
