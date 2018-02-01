'use strict';

var TestnetNodes = require('../../testnet-nodes');
var FS = require("fs");
var expect = require('chai').expect;
var BigNumber = require('bignumber.js');

var nodes = new TestnetNodes();
var coinbase = "1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c"
var passphrase = "passphrase"
nodes.Start();

function checkTransaction(hash, done, count) {
    if (count > 6) {
        console.log("tx receipt timeout:" + hash);
        nodes.Stop();
        done();
        return;
    }

    var node = nodes.RPC(0);
    node.api.getTransactionReceipt(hash).then(function (resp) {
        console.log("tx receipt:" + JSON.stringify(resp));
        return node.api.getAccountState(coinbase);
    }).then(function (resp) {
        console.log("after state:" + JSON.stringify(resp));
        nodes.Stop();
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
        var node = nodes.RPC(0);
        node.api.getAccountState(coinbase).then(function (resp) {
            console.log(resp);
            state = resp;
            var erc20 = FS.readFileSync("../nf/nvm/test/ERC20.js", "utf-8");
            var contract = {
                "source": erc20,
                "sourceType": "js",
                "args": '["TestToken", "NAS", 1000000000]'
            }
            return nodes.SendTransaction(coinbase, coinbase, "0", parseInt(state.nonce) + 1, "0", "2000000", contract);
        }).then(function (resp) {
            console.log(resp);
            expect(resp).to.be.have.property('contract_address');
            var call = {
                "function": "totalSupply"
            }
            return nodes.SendTransaction(coinbase, resp.contract_address, "0", parseInt(state.nonce) + 2, "0", "2000000", call);
        }).then(function (resp) {
            console.log(resp);
            checkTransaction(resp.txhash, done, 1);
        }).catch(function (err) {
            console.log("send err:" + err)
            nodes.Stop();
            done();
        });
    });
});
