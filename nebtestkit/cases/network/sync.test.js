'use strict';

var Node = require('../../Node');
var Neblet = require('../../neblet');
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
var nodes = new Node(nodeCnt - 1);
var connected = 0;
nodes.Start();

describe('check sync', function () {
    before(function (done) {
        this.timeout(300000);
        setTimeout(done, 5000);
    });

    it('check', function (done) {
        // start servers
        console.log("start servers");
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

        // start another node to test sync
        var i = nodeCnt - 1;
        var miner = "75e4e5a71d647298b88928d8cb5da43d90ab1a6c52d0905f";
        var server = new Neblet(
            "127.0.0.1", 10000 + i, 8090 + i, 9090 + i,
            miner, miner, 'passphrase'
        );
        server.Init(nodes.Node(0));
        server.Start();
        nodes.Nodes().push(server);
        sleep(15000);
        var blockSeed;
        nodes.RPC(0).api.blockDump(1).then(function (resp) {
            blockSeed = JSON.parse(resp.data)[0];
        });

        var block;
        server.NebJs().api.blockDump(1).then(function (resp) {
            block = JSON.parse(resp.data)[0];
        });
        sleep(3000);
        console.log(blockSeed, " vs ", block);
        expect(blockSeed.height).to.be.equal(block.height);
        console.log("√ start a new node and the new node has synced all the blocks");

        // verify block sync
        nodes.RPC(5).admin.changeNetworkID(10);

        sleep(10000);
        var blockA;
        nodes.RPC(0).api.blockDump(1).then(function (resp) {
            blockA = JSON.parse(resp.data)[0];
        });

        var blockB;
        nodes.RPC(5).api.blockDump(1).then(function (resp) {
            blockB = JSON.parse(resp.data)[0];
        });
        sleep(10000);
        console.log(blockA, " vs ", blockB);
        expect(blockA.hash).not.to.be.equal(blockB.hash);
        console.log("√ changed the new node networkID to 10 and the new node go to forked");

        nodes.RPC(5).admin.changeNetworkID(1);
        sleep(10000);
        nodes.RPC(0).api.blockDump(1).then(function (resp) {
            blockA = JSON.parse(resp.data)[0];
        });

        nodes.RPC(5).api.blockDump(1).then(function (resp) {
            blockB = JSON.parse(resp.data)[0];
        });

        sleep(10000);

        console.log(blockA, " vs ", blockB);
        expect(blockA.hash).to.be.equal(blockB.hash);
        console.log("√ recover the new node networkID to 1 and the new node is synchronized");
        sleep(10000);

        // quit
        nodes.Stop();
        done();
    });
});