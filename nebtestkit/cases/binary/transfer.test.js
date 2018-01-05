'use strict';

var TestnetNodes = require('../../testnet-nodes');
var FS = require("fs");
var expect = require('chai').expect;
var BigNumber = require('bignumber.js');
var sleep = require("system-sleep");

var nodes = new TestnetNodes();
var coinbase = "c5bcfcb3fa8250be4f2bf2b1e70e1da500c668377ba8cd4a"
var to = "git "
var passphrase = "passphrase"
nodes.Start();

describe('binary transaction', function () {
    before(function (done) {
        this.timeout(10000);
        setTimeout(done, 5000);
    });

    it('erc20 contract', function (done) {
        var node = nodes.RPC(0);

        node.api.getAccountState(coinbase).then(function (resp) {
            console.log(resp);
        });
        sleep(1000);

        node.api.getAccountState(to).then(function (resp) {
            console.log(resp);
        });
        sleep(1000);

        node.api.getAccountState(coinbase).then(function (resp) {
            console.log(resp);
            var nonce = parseInt(resp.nonce);
            for (var i = 0; i < 100; i++) {
                nodes.SendTransaction(coinbase, to, "1000000000000000000", nonce + i + 1, "0", "2000000").then(function (resp) {
                    console.log(resp);
                });
            }
        })
        sleep(60000);

        nodes.Stop();
        done();
    });
});
