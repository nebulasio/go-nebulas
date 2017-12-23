'use strict';

var Node = require('../../Node');
var expect = require('chai').expect;
var sleep = require("system-sleep")
var os = require('os');

var count = 21;
var validators = 6;
var blockInterval = 5;
var dynastyInterval = 60;
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

describe('check dump blocks', function () {
    it('check no empty slots', function () {
        var block = null;
        for (var i = 0; i < count; i++) {
            var block = JSON.parse(nodes.RPC(0).api.blockDump(1).data)[0];
            console.log(block);
            expect(block.timestamp % 5).to.be.equal(0)
            if (block.height > 1) {
                var idx = (block.timestamp % dynastyInterval) % (blockInterval * validators) / blockInterval;
                console.log(idx);
                if (idx != validators - 1) {
                    expect(nodes.Coinbase(idx)).to.be.equal(block.coinbase)
                }
            }
            sleep(blockInterval * 1000);
        }
    });
})

describe('quit', function () {
    it('quit', function () {
        nodes.Stop();
    });
});
