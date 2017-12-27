'use strict';

var Node = require('../../../node');
var expect = require('chai').expect;
var BigNumber = require('bignumber.js');

var nodes = new Node(6);
nodes.Start();

function checkTransaction(hash, done, count) {
    if (count > 5) {
        console.log("tx receipt timeout:"+hash);
        done();
        return;
    }
    var node = nodes.Node(0);
    node.RPC().api.getTransactionReceipt(hash).then(function (resp) {
        console.log(JSON.stringify(resp));
        done();
    }).catch(function (err) {
        setTimeout(function () {
            checkTransaction(hash, done, count+1);
        }, 3000);
    });
}

describe('normal transaction', function () {
    before(function (done) {
        this.timeout(1000000);
        setTimeout(done, 5000);
    });

    it('normal transfer', function (done) {
        var node = nodes.Node(0);
        node.RPC().api.getAccountState(node.Coinbase()).then(function (state) {
            console.log("resp:"+JSON.stringify(state));
            node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase()).then(function (resp) {
                node.RPC().api.sendTransaction(node.Coinbase(), nodes.Coinbase(1), state.balance, parseInt(state.nonce) + 1).then(function (resp) {
                    console.log("send:"+JSON.stringify(resp))
                    expect(resp).to.be.have.property('txhash');
                    // checkTransaction(resp.txhash, done, 1);
                    done();
                })
            });
        });
    });

    it('from & to are same', function (done) {
        var node = nodes.Node(0);
        node.RPC().api.getAccountState(node.Coinbase()).then(function (state) {
            console.log("resp:"+JSON.stringify(state));
            node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase()).then(function (resp) {
                node.RPC().api.sendTransaction(node.Coinbase(), nodes.Coinbase(0), state.balance, parseInt(state.nonce) + 1).then(function (resp) {
                    console.log("send:"+JSON.stringify(resp))
                    expect(resp).to.be.have.property('txhash');
                    // checkTransaction(resp.txhash, done, 1);
                    done();
                })
            });
        });
    });

    it('from balance is insufficient', function (done) {

        var node = nodes.Node(0);
        node.RPC().api.getAccountState(node.Coinbase()).then(function (state) {
            console.log("resp:"+JSON.stringify(state));
            node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase()).then(function (resp) {
                var value = new BigNumber(state.balance).add("1");
                node.RPC().api.sendTransaction(node.Coinbase(), nodes.Coinbase(0), value.toString(), parseInt(state.nonce) + 1).then(function (resp) {
                    console.log("send:"+JSON.stringify(resp))
                    expect(resp).to.be.have.property('txhash');
                    // checkTransaction(resp.txhash, done, 1);
                    done();
                })
            });
        });
    });

    it('gas is insufficient', function (done) {

        var node = nodes.Node(0);
        node.RPC().api.getAccountState(node.Coinbase()).then(function (state) {
            console.log("resp:"+JSON.stringify(state));
            node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase()).then(function (resp) {
                node.RPC().api.sendTransaction(node.Coinbase(), nodes.Coinbase(0), state.balance, parseInt(state.nonce) + 1, "0", "1").then(function (resp) {
                    console.log("send:"+JSON.stringify(resp))
                    expect(resp).to.be.have.property('txhash');
                    // checkTransaction(resp.txhash, done, 1);
                    done();
                })
            });
        });
    });

    it('from is invalid address', function (done) {

        var node = nodes.Node(0);
        node.RPC().api.getAccountState(node.Coinbase()).then(function (state) {
            console.log("resp:"+JSON.stringify(state));
            node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase()).then(function (resp) {
                node.RPC().api.sendTransaction("0x00", nodes.Coinbase(0), state.balance, parseInt(state.nonce) + 1).then(function (resp) {
                    console.log("send:"+JSON.stringify(resp))
                    expect(resp).to.be.have.property('txhash');
                    // checkTransaction(resp.txhash, done, 1);
                    done();
                }).catch(function (err) {
                    console.log("send err:"+JSON.stringify(err.error))
                    done();
                });
            });
        });
    });

    it('to is invalid address', function (done) {

        var node = nodes.Node(0);
        node.RPC().api.getAccountState(node.Coinbase()).then(function (state) {
            console.log("resp:"+JSON.stringify(state));
            node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase()).then(function (resp) {
                node.RPC().api.sendTransaction(node.Coinbase(), "0x00", state.balance, parseInt(state.nonce) + 1).then(function (resp) {
                    console.log("send:"+JSON.stringify(resp))
                    expect(resp).to.be.have.property('txhash');
                    // checkTransaction(resp.txhash, done, 1);
                    done();
                }).catch(function (err) {
                    console.log("send err:"+JSON.stringify(err.error))
                    done();
                });
            });
        });
    });

    it('nonce is below', function (done) {

        var node = nodes.Node(0);
        node.RPC().api.getAccountState(node.Coinbase()).then(function (state) {
            console.log("resp:"+JSON.stringify(state));
            node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase()).then(function (resp) {
                node.RPC().api.sendTransaction(node.Coinbase(), nodes.Coinbase(1), state.balance, parseInt(state.nonce)).then(function (resp) {
                    console.log("send:"+JSON.stringify(resp))
                    expect(resp).to.be.have.property('txhash');
                    // checkTransaction(resp.txhash, done, 1);
                    done();
                }).catch(function (err) {
                    console.log("send err:"+JSON.stringify(err.error))
                    done();
                });
            });
        });
    });

    it('nonce is heigher', function (done) {

        var node = nodes.Node(0);
        node.RPC().api.getAccountState(node.Coinbase()).then(function (state) {
            console.log("resp:"+JSON.stringify(state));
            node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase()).then(function (resp) {
                var nonce = new BigNumber(state.nonce).add(2);
                node.RPC().api.sendTransaction(node.Coinbase(), nodes.Coinbase(1), state.balance, nonce.toNumber()).then(function (resp) {
                    console.log("send:"+JSON.stringify(resp))
                    expect(resp).to.be.have.property('txhash');
                    // checkTransaction(resp.txhash, done, 1);
                    done();
                }).catch(function (err) {
                    console.log("send err:"+JSON.stringify(err.error))
                    done();
                });
            });
        });
    });

    it('gasPrice is below', function (done) {

        var node = nodes.Node(0);
        node.RPC().api.getAccountState(node.Coinbase()).then(function (state) {
            console.log("resp:"+JSON.stringify(state));
            node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase()).then(function (resp) {
                node.RPC().api.sendTransaction(node.Coinbase(), nodes.Coinbase(1), state.balance, parseInt(state.nonce) + 1, "1").then(function (resp) {
                    console.log("send:"+JSON.stringify(resp))
                    expect(resp).to.be.have.property('txhash');
                    // checkTransaction(resp.txhash, done, 1);
                    done();
                }).catch(function (err) {
                    console.log("send err:"+JSON.stringify(err.error))
                    done();
                });
            });
        });
    });

    it('gas is higher than max', function (done) {

        var node = nodes.Node(0);
        node.RPC().api.getAccountState(node.Coinbase()).then(function (state) {
            console.log("resp:"+JSON.stringify(state));
            node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase()).then(function (resp) {
                var maxGas = new BigNumber(10).pow(9).mul(60);
                node.RPC().api.sendTransaction(node.Coinbase(), nodes.Coinbase(1), state.balance, parseInt(state.nonce) + 1, "", maxGas.toString()).then(function (resp) {
                    console.log("send:"+JSON.stringify(resp))
                    expect(resp).to.be.have.property('txhash');
                    // checkTransaction(resp.txhash, done, 1);
                    done();
                }).catch(function (err) {
                    console.log("send err:"+JSON.stringify(err.error))
                    done();
                });
            });
        });
    });
});

describe('quit', function () {
    it('quit', function () {
        nodes.Stop();
    });
});