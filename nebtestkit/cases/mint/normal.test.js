'use strict';

var Node = require('../../Node');
var expect = require('chai').expect;
var sleep = require("system-sleep")
var os = require('os');

var nodes = new Node(6);
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
            if (nodeInfo.route_table.length == 5) {
                break;
            }
            sleep(3000);
        }
    });
});

describe('check dump blocks', function () {
    it('check no empty slots', function () {
        var size = nodes.count;
        var block = null;
        for (var i = 0; i < size; i++) {
            var now = new Date();
            var block = JSON.parse(nodes.RPC(0).api.blockDump(1).data)[0];
            console.log(block)
            expect(block.timestamp % 5).to.be.equal(0)
            if (block.height > 1) {
                expect(nodes.Coinbase((block.timestamp % 60) % (5 * 6) / 5)).to.be.equal(block.coinbase)
            }
            sleep(5000);
        }
    });
})

describe('quit', function () {
    it('quit', function () {
        nodes.Stop();
    });
});
