'use strict';

var LocalNodes = require('../../local-nodes');
var BigNumber = require('bignumber.js');
var expect = require('chai').expect;
var sleep = require("system-sleep")
var os = require('os');

var nodeCnt = 6;
var validators = 6;
var blockInterval = 5;
var dynastyInterval = 60;
var reward = new BigNumber("48e16");
var initial = new BigNumber("1e18");
var nodes = new LocalNodes(nodeCnt);
var now = new Date().getTime() / 1000;
var connected = 0;
nodes.Start();

function unlockAccount() {
    var node = nodes.Node(1);
    return node.RPC().admin.unlockAccount(node.Coinbase(), node.Passphrase());
}

function changeCandidate() {
    var node = nodes.Node(1);
    return node.RPC().api.sendTransaction(
        node.Coinbase(), node.Coinbase(),
        0, 1, 0, 20000000,
        null, { action: "logout" }, null)
}

function changeVote() {
    var node = nodes.Node(1);
    return node.RPC().api.sendTransaction(
        node.Coinbase(), node.Coinbase(),
        0, 2, 0, 20000000,
        null, null, { action: "do", delegatee: "fc751b484bd5296f8d267a8537d33f25a848f7f7af8cfcf6" });
}


function check(resp) {
    return new Promise(function (res, rej) {
        (resp != undefined && dynasty > 2) ? res() : rej()
    })
}

function Stop(done) {
    nodes.Stop();
    done();
}

describe('right miner', function () {
    before(function (done) {
        this.timeout(300000);
        setTimeout(done, 5000);
    });

    it('check', function (done) {
        // start servers
        console.log("start servers");
        while (true) {
            if (connected == nodeCnt - 1) break;
            var nodeInfo = nodes.RPC(0).api.nodeInfo().then(function (resp) {
                console.log(resp)
                connected = resp.route_table.length
            });
            sleep(3000);
        }

        Promise.resolve()
            .then(unlockAccount)
            .then(function (resp) {
                console.log(resp);
                expect(resp).to.not.have.property("error");
            })
            .then(changeCandidate)
            .then(function (resp) {
                console.log(resp);
                expect(resp).to.not.have.property("error");
            })
            .then(changeVote)
            .then(function (resp) {
                console.log(resp);
                expect(resp).to.be.have.property("txhash");
            });

        for (var i = 0; i < dynastyInterval * 2 + blockInterval; i++) {
            console.log(i, dynastyInterval * 2 + blockInterval);
            sleep(1000);
        }

        nodes.RPC(0).admin.getDynasty()
            .then(function (resp) {
                console.log(resp);
                expect(resp.delegatees[1]).to.be.equal("333cb3ed8c417971845382ede3cf67a0a96270c05fe2f700");
                expect(resp.delegatees[validators - 1]).to.be.equal("fc751b484bd5296f8d267a8537d33f25a848f7f7af8cfcf6");
            })
            .then(function () {
                Stop(done);
            })
            .catch(function (err) {
                console.log(err);
                Stop(done);
            });
    });
});
