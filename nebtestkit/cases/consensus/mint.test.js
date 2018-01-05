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
var connected = 0;
nodes.Start();

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

        // check right miner
        console.log("check right miner");
        var block = null;
        for (var i = 0; i < nodeCnt; i++) {
            nodes.RPC(0).api.blockDump(1).then(function (resp) {
                var block = JSON.parse(resp.data)[0];
                console.log(block);
                expect(block.timestamp % 5).to.be.equal(0)
                if (block.height > 1) {
                    var idx = (block.timestamp % dynastyInterval) % (blockInterval * validators) / blockInterval;
                    if (idx != validators - 1) {
                        expect(nodes.Coinbase(idx)).to.be.equal(block.coinbase)
                    }
                }
            });
            sleep(blockInterval * 1000);
        }

        // check balances correct
        console.log("check balances correct")
        nodes.RPC(0).api.blockDump(nodeCnt * 10).then(function (resp) {
            var blocks = JSON.parse(resp.data);
            console.log(blocks);
            var balances = {}
            for (var i = 0; i < blocks.length; i++) {
                var block = blocks[i];
                var coinbase = block.coinbase;
                if (balances[coinbase] == undefined) {
                    balances[coinbase] = new BigNumber(0)
                }
                balances[blocks[i].coinbase] = balances[blocks[i].coinbase].plus(reward)
            }

            var keys = Object.keys(balances);
            var tail = blocks[0].hash;
            var finished = 0;
            for (let i = 0; i < keys.length; i++) {
                // coinbase in genesis, skip it. it's not a valid address
                if (keys[i] == "000000000000000000000000000000000000000000000000") {
                    finished += 1;
                    continue
                }
                nodes.RPC(0).api.getAccountState(keys[i], tail).then(function (state) {
                    var address = keys[i];
                    console.log(address, state);
                    var balance = new BigNumber(state.balance);
                    if (address == "1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c" ||
                        address == "2fe3f9f51f9a05dd5f7c5329127f7c917917149b4e16b0b8") {
                        balance = balance.minus(initial);
                    }
                    expect(balance.toString()).to.be.equal(balances[address].toString());
                    finished += 1;
                    if (finished == keys.length) {
                        console.log("over")
                        nodes.Stop();
                        done()
                    }
                });
            }
        });
    });
});