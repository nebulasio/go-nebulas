'use strict';

var Node = require('../../../node');
var FS = require("fs");
var expect = require('chai').expect;
var BigNumber = require('bignumber.js');

var nodes = new Node(3);
nodes.Start();

describe('contract transaction', function () {
    before(function (done) {
        this.timeout(10000);
        setTimeout(done, 5000);
    });

    it('erc20 contract', function () {
        var node = nodes.Node(0);
        var state = node.RPC().api.getAccountState(node.Coinbase());
        node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase(), function (err, resp) {
            if (err != null) {
                console.log(err);
                return;
            }

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
            node.RPC().api.sendTransaction(node.Coinbase(), nodes.Coinbase(1), "0", parseInt(state.nonce)+1, "0", "2000000", contract, null, null, function (err, resp) {
                // console.log("send resp:"+JSON.stringify(resp));
                expect(resp).to.be.have.property('contract_address');

                var call = {
                    "function": "totalSupply"
                }
                // gas = node.RPC().api.estimateGas(node.Coinbase(), node.Coinbase(), "0", parseInt(state.nonce)+1, "0", "0", call);
                // console.log("gas:"+gas.estimate_gas);
                node.RPC().api.call(node.Coinbase(), resp.contract_address, "0", parseInt(state.nonce)+2, "0", "2000000", call, function (err, resp) {
                    // console.log("call resp:"+JSON.stringify(resp));
                    expect(resp).to.be.have.property('txhash');
                });
            });

        });

    });

    it('bank vault js', function () {
        var node = nodes.Node(0);
        var state = node.RPC().api.getAccountState(node.Coinbase());
        node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase(), function (err, resp) {
            if (err != null) {
                console.log(err);
                return;
            }

            var bank = FS.readFileSync("../nf/nvm/test/bank_vault_contract.js","utf-8");
            // console.log("erc20:"+erc20);
            var contract = {
                "source": bank,
                "sourceType": "js",
            }

            // var price = node.RPC().api.gasPrice();
            // var gas = node.RPC().api.estimateGas(node.Coinbase(), node.Coinbase(), "0", parseInt(state.nonce)+1, "0", "0", contract);
            // // console.log("gas:"+gas.estimate_gas);
            var resp = node.RPC().api.sendTransaction(node.Coinbase(), nodes.Coinbase(1), "0", parseInt(state.nonce)+1, "0", "200000", contract, null, null, function (err, resp) {
                // console.log("resp:"+JSON.stringify(resp));
                expect(resp).to.be.have.property('contract_address');

                var call = {
                    "function": "save",
                    "args":"[1]"
                }
                // gas = node.RPC().api.estimateGas(node.Coinbase(), node.Coinbase(), "0", parseInt(state.nonce)+1, "0", "0", call);
                // console.log("gas:"+gas.estimate_gas);
                node.RPC().api.call(node.Coinbase(), resp.contract_address, state.balance, parseInt(state.nonce)+2, "0", "2000000", call, function (err, resp) {
                    // console.log("resp:"+JSON.stringify(resp));
                    expect(resp).to.be.have.property('txhash');
                });
            });
        });
    });
    it('bank vault ts', function () {
        var node = nodes.Node(0);
        var state = node.RPC().api.getAccountState(node.Coinbase());
        node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase(), function (err, resp) {
            if (err != null) {
                console.log(err);
                return;
            }

            var bank = FS.readFileSync("../nf/nvm/test/bank_vault_contract.ts","utf-8");
            // console.log("erc20:"+erc20);
            var contract = {
                "source": bank,
                "sourceType": "ts",
            }

            // var price = node.RPC().api.gasPrice();
            // var gas = node.RPC().api.estimateGas(node.Coinbase(), node.Coinbase(), "0", parseInt(state.nonce)+1, "0", "0", contract);
            // console.log("gas:"+gas.estimate_gas);
            node.RPC().api.sendTransaction(node.Coinbase(), nodes.Coinbase(1), "0", parseInt(state.nonce)+1, "0", "2000000", contract, null, null, function (err, resp) {
                // console.log("resp:"+JSON.stringify(resp));
                expect(resp).to.be.have.property('contract_address');

                var call = {
                    "function": "save",
                    "args":"[1]"
                }
                // gas = node.RPC().api.estimateGas(node.Coinbase(), node.Coinbase(), "0", parseInt(state.nonce)+1, "0", "0", call);
                // console.log("gas:"+gas.estimate_gas);
                node.RPC().api.call(node.Coinbase(), resp.contract_address, state.balance, parseInt(state.nonce)+2, "0", "200000", call, function (err, resp) {
                    // console.log("resp:"+JSON.stringify(resp));
                    expect(resp).to.be.have.property('txhash');
                });
            });
        });
    });
});

describe('quit', function () {
    it('quit', function () {
        setTimeout(function () {
            nodes.Stop();
        }, 2000);
    });
});