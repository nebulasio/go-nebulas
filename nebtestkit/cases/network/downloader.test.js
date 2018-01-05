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

describe('check downloader', function () {
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

        let height = 0;
        while (true) {
            nodes.RPC(0).api.blockDump(1).then(function (resp) {
                var block = JSON.parse(resp.data)[0];
                console.log(block);
                if (height < block.height) {
                    height = block.height;
                    console.log("H", height);
                }
            });
            console.log(height);
            if (height > 3) {
                console.log("√ the height of current tail block is higher than 3");
                break;
            }
            sleep(3000);
        }

        nodes.RPC(5).admin.changeNetworkID(10);
        sleep(15000);
        var blockA;
        nodes.RPC(0).api.blockDump(1).then(function (resp) {
            blockA = JSON.parse(resp.data)[0];
        });

        var blockB;
        nodes.RPC(5).api.blockDump(1).then(function (resp) {
            blockB = JSON.parse(resp.data)[0];
        });
        sleep(3000);
        console.log(blockA, " vs ", blockB);
        expect(blockA.hash).not.to.be.equal(blockB.hash);
        console.log("√ changed a node networkID to 10 and the node go to forked");


        nodes.RPC(5).admin.changeNetworkID(1);
        sleep(20000);
        nodes.RPC(0).api.blockDump(1).then(function (resp) {
            blockA = JSON.parse(resp.data)[0];
        });

        // var blockSeed = JSON.parse(nodes.RPC(0).api.blockDump(1).data)[0];
        nodes.RPC(5).api.blockDump(1).then(function (resp) {
            blockB = JSON.parse(resp.data)[0];
        });
        sleep(3000);
        console.log(blockA, " vs ", blockB);
        expect(blockA.hash).to.be.equal(blockB.hash);
        console.log("√ recover the node networkID to 1 and the node is synchronized");
        sleep(10000);

        nodes.Stop();
        done();
    });
});