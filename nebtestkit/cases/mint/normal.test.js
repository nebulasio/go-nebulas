'use strict';

var Node = require('../../Node');
var BigNumber = require('bignumber.js');
var expect = require('chai').expect;
var sleep = require("system-sleep")
var os = require('os');

var count = 6;
var validators = 6;
var blockInterval = 5;
var dynastyInterval = 60;
var reward = new BigNumber("48e16");
var initial = new BigNumber("1e18");
var nodes = new Node(count);
nodes.Start();

describe('seed server start correctly', function () {
    before(function (done) {
        this.timeout(10000);
        setTimeout(done, 4000);
    });

    it('check status', function () {
        while (true) {
            var nodeInfo = nodes.RPC(0).api.nodeInfo()
            console.log(nodeInfo)
            if (nodeInfo.route_table.length == count - 1) {
                break;
            }
            sleep(3000);
        }
    });
});

describe('check blocks', function () {
    it('check right miner', function () {
        var block = null;
        for (var i = 0; i < count; i++) {
            var block = JSON.parse(nodes.RPC(0).api.blockDump(1).data)[0];
            console.log(block);
            expect(block.timestamp % 5).to.be.equal(0)
            if (block.height > 1) {
                var idx = (block.timestamp % dynastyInterval) % (blockInterval * validators) / blockInterval;
                if (idx != validators - 1) {
                    expect(nodes.Coinbase(idx)).to.be.equal(block.coinbase)
                }
            }
            sleep(blockInterval * 1000);
        }
    });
})

describe('check state', function () {
    it('check right balances', function () {
        var blocks = JSON.parse(nodes.RPC(0).api.blockDump(count * 10).data);
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
        for (var i = 0; i < keys.length; i++) {
            var address = keys[i];
            // coinbase in genesis, skip it. it's not a valid address
            if (address == "000000000000000000000000000000000000000000000000") {
                continue
            }
            var state = nodes.RPC(0).api.getAccountState(keys[i], tail);
            console.log(address, state);
            var balance = new BigNumber(state.balance);
            if (address == "1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c" ||
                address == "2fe3f9f51f9a05dd5f7c5329127f7c917917149b4e16b0b8") {
                balance = balance.minus(initial);
            }
            expect(balance.toString()).to.be.equal(balances[address].toString());
        }
    });
})

describe('quit', function () {
    it('quit', function () {
        nodes.Stop();
    });
});
