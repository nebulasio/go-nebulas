'use strict';

var TestnetNodes = require('../../testnet-nodes');
var FS = require("fs");
var expect = require('chai').expect;
var BigNumber = require('bignumber.js');

var from = "1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c";

var nodes = new TestnetNodes();
nodes.Start();

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

describe('contract transaction', function () {
    before(function (done) {
        this.timeout(10000);
        setTimeout(done, 5000);
    });

    it('erc20 contract', function (done) {

        var state;
        nodes.RPC(0).api.getAccountState(from).then(function (resp) {

            state = resp;
            var erc20 = FS.readFileSync("../nf/nvm/test/ERC20.js", "utf-8");
            // console.log("erc20:"+erc20);
            var contract = {
                "source": erc20,
                "sourceType": "js",
                "args": '["NebulasToken", "NAS", 1000000000]'
            }
            return nodes.SendTransaction(from, from, "0", parseInt(resp.nonce) + 1, "0", "2000000", contract);
        }).then(function (resp) {

            console.log("send resp:" + JSON.stringify(resp));
            expect(resp).to.be.have.property('contract_address');
            var call = {
                "function": "totalSupply",
                "args":""
            }
            return nodes.SendTransaction(from, resp.contract_address, "0", parseInt(state.nonce) + 2, "0", "200000000", call);
        }).then(function (resp) {
            console.log("call resp:" + JSON.stringify(resp));
            checkTransaction(resp.txhash, done, 1);
        }).catch(function (err) {
            console.log(JSON.stringify(err.error))
            done();
        });
    });

    it('bank vault js', function (done) {

        var state;
        nodes.RPC(0).api.getAccountState(from).then(function (resp) {

            state = resp;
            var bank = FS.readFileSync("../nf/nvm/test/bank_vault_contract.js", "utf-8");
            // console.log("erc20:"+erc20);
            var contract = {
                "source": bank,
                "sourceType": "js",
            }

            // var price = node.RPC().api.gasPrice();
            // var gas = node.RPC().api.estimateGas(node.Coinbase(), node.Coinbase(), "0", parseInt(state.nonce)+1, "0", "0", contract);
            // // console.log("gas:"+gas.estimate_gas);
            return nodes.SendTransaction(from, from, "0", parseInt(state.nonce) + 1, "0", "200000", contract);
        }).then(function (resp) {

            console.log("resp:" + JSON.stringify(resp));
            expect(resp).to.be.have.property('contract_address');

            var call = {
                "function": "save",
                "args": "[1]"
            }
            // gas = node.RPC().api.estimateGas(node.Coinbase(), node.Coinbase(), "0", parseInt(state.nonce)+1, "0", "0", call);
            // console.log("gas:"+gas.estimate_gas);
            return nodes.SendTransaction(from, resp.contract_address, "1", parseInt(state.nonce) + 2, "0", "2000000", call);
        }).then(function (resp) {

            console.log("resp:" + JSON.stringify(resp));
            // expect(resp).to.be.have.property('txhash');
            // done();
            checkTransaction(resp.txhash, done, 1);
        }).catch(function (err) {
            console.log(JSON.stringify(err.error))
            done();
        });
    });

    it('bank vault ts', function (done) {

        var state;
        nodes.RPC(0).api.getAccountState(from).then(function (resp) {

            state = resp;
            var bank = FS.readFileSync("../nf/nvm/test/bank_vault_contract.ts", "utf-8");
            // console.log("erc20:"+erc20);
            var contract = {
                "source": bank,
                "sourceType": "ts",
            }

            // var price = node.RPC().api.gasPrice();
            // var gas = node.RPC().api.estimateGas(node.Coinbase(), node.Coinbase(), "0", parseInt(state.nonce)+1, "0", "0", contract);
            // console.log("gas:"+gas.estimate_gas);
            return nodes.SendTransaction(from, from, "0", parseInt(state.nonce) + 1, "0", "2000000", contract);
        }).then(function (resp) {

            console.log("resp:" + JSON.stringify(resp));
            expect(resp).to.be.have.property('contract_address');

            var call = {
                "function": "save",
                "args": "[1]"
            }
            // gas = node.RPC().api.estimateGas(node.Coinbase(), node.Coinbase(), "0", parseInt(state.nonce)+1, "0", "0", call);
            // console.log("gas:"+gas.estimate_gas);
            return nodes.SendTransaction(from, resp.contract_address, "1", parseInt(state.nonce) + 2, "0", "200000", call);
        }).then(function (resp) {

            console.log("resp:" + JSON.stringify(resp));
            // expect(resp).to.be.have.property('txhash');
            // done();
            checkTransaction(resp.txhash, done, 1);
        }).catch(function (err) {
            console.log(JSON.stringify(err.error))
            done();
        });
    });

    it('quit', function () {
        nodes.Stop();
    });
});
