'use strict';

var Node = require('../../../node');
var FS = require("fs");
var expect = require('chai').expect;
var BigNumber = require('bignumber.js');

var nodes = new Node(6);
nodes.Start();

function checkTransaction(hash, done, count) {
    if (count > 6) {
        console.log("tx receipt timeout:"+hash);
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
            checkTransaction(hash, done, count+1);
        }, 2000);
    });
}

describe('contract transaction', function () {
    before(function (done) {
        this.timeout(10000);
        setTimeout(done, 5000);
    });

    it('erc20 contract', function (done) {

        var state;
        var node = nodes.Node(0);
        node.RPC().api.getAccountState(node.Coinbase()).then(function (resp) {

            state = resp;
            return node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase());
        }).then(function (resp) {

            var erc20 = FS.readFileSync("../nf/nvm/test/ERC20.js","utf-8");
            // console.log("erc20:"+erc20);
            var contract = {
                "source": erc20,
                "sourceType": "js",
                "args": '["NebulasToken", "NAS", 1000000000]'
            }

            // var price = node.RPC().api.gasPrice();
            // var gas = node.RPC().api.estimateGas(node.Coinbase(), node.Coinbase(), "0", parseInt(state.nonce)+1, "0", "0", contract);
            // console.log("gas:"+gas.estimate_gas);
            return node.RPC().api.sendTransaction(node.Coinbase(), nodes.Coinbase(1), "0", parseInt(state.nonce)+1, "0", "2000000", contract);
        }).then(function (resp) {

            console.log("send resp:"+JSON.stringify(resp));
            expect(resp).to.be.have.property('contract_address');

            var call = {
                "function": "totalSupply"
            }
            // gas = node.RPC().api.estimateGas(node.Coinbase(), node.Coinbase(), "0", parseInt(state.nonce)+1, "0", "0", call);
            // console.log("gas:"+gas.estimate_gas);
            return node.RPC().api.call(node.Coinbase(), resp.contract_address, "0", parseInt(state.nonce)+2, "0", "2000000", call);
        }).then(function (resp) {

            console.log("call resp:"+JSON.stringify(resp));
            // expect(resp).to.be.have.property('txhash');
            // done();
            checkTransaction(resp.txhash, done, 1);
        }).catch(function (err) {
            console.log("send err:"+JSON.stringify(err.error))
            done();
        });
    });

    it('bank vault js', function (done) {

        var state;
        var node = nodes.Node(0);
        node.RPC().api.getAccountState(node.Coinbase()).then(function (resp) {

            state = resp;
            return node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase());
        }).then( function (resp) {

            var bank = FS.readFileSync("../nf/nvm/test/bank_vault_contract.js","utf-8");
            // console.log("erc20:"+erc20);
            var contract = {
                "source": bank,
                "sourceType": "js",
            }

            // var price = node.RPC().api.gasPrice();
            // var gas = node.RPC().api.estimateGas(node.Coinbase(), node.Coinbase(), "0", parseInt(state.nonce)+1, "0", "0", contract);
            // // console.log("gas:"+gas.estimate_gas);
            return node.RPC().api.sendTransaction(node.Coinbase(), nodes.Coinbase(1), "0", parseInt(state.nonce)+1, "0", "200000", contract);
        }).then( function (resp) {

            console.log("resp:"+JSON.stringify(resp));
            expect(resp).to.be.have.property('contract_address');

            var call = {
                "function": "save",
                "args":"[1]"
            }
            // gas = node.RPC().api.estimateGas(node.Coinbase(), node.Coinbase(), "0", parseInt(state.nonce)+1, "0", "0", call);
            // console.log("gas:"+gas.estimate_gas);
            return node.RPC().api.call(node.Coinbase(), resp.contract_address, state.balance, parseInt(state.nonce)+2, "0", "2000000", call);
        }).then( function (resp) {

            console.log("resp:"+JSON.stringify(resp));
            // expect(resp).to.be.have.property('txhash');
            // done();
            checkTransaction(resp.txhash, done, 1);
        }).catch(function (err) {
            console.log("send err:"+JSON.stringify(err.error))
            done();
        });
    });

    it('bank vault ts', function (done) {

        var state;
        var node = nodes.Node(0);
        node.RPC().api.getAccountState(node.Coinbase()).then(function (resp) {

            state = resp;
            return node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase());
        }).then( function (resp) {

            var bank = FS.readFileSync("../nf/nvm/test/bank_vault_contract.ts","utf-8");
            // console.log("erc20:"+erc20);
            var contract = {
                "source": bank,
                "sourceType": "ts",
            }

            // var price = node.RPC().api.gasPrice();
            // var gas = node.RPC().api.estimateGas(node.Coinbase(), node.Coinbase(), "0", parseInt(state.nonce)+1, "0", "0", contract);
            // console.log("gas:"+gas.estimate_gas);
            return node.RPC().api.sendTransaction(node.Coinbase(), nodes.Coinbase(1), "0", parseInt(state.nonce)+1, "0", "2000000", contract);
        }).then( function (resp) {

            console.log("resp:"+JSON.stringify(resp));
            expect(resp).to.be.have.property('contract_address');

            var call = {
                "function": "save",
                "args":"[1]"
            }
            // gas = node.RPC().api.estimateGas(node.Coinbase(), node.Coinbase(), "0", parseInt(state.nonce)+1, "0", "0", call);
            // console.log("gas:"+gas.estimate_gas);
            return node.RPC().api.call(node.Coinbase(), resp.contract_address, state.balance, parseInt(state.nonce)+2, "0", "200000", call);
        }).then( function (resp) {

            console.log("resp:"+JSON.stringify(resp));
            // expect(resp).to.be.have.property('txhash');
            // done();
            checkTransaction(resp.txhash, done, 1);
        }).catch(function (err) {
            console.log("send err:"+JSON.stringify(err.error))
            done();
        });
    });

    it('quit', function () {
        nodes.Stop();
    });
});
