'use strict';

var Node = require('../../Node');
var expect = require('chai').expect;
var sleep = require("system-sleep");
var Neblet = require('../../neblet');
var os = require('os');
var count = 6;

var nodes = new Node(count);
nodes.Start();

describe('start nodes', function () {
    before(function (done) {
        this.timeout(10000);
        setTimeout(done, 4000);
    });

    it('check status', function (done) {
        this.timeout(20000);
        while (true) {
            var nodeInfo = nodes.RPC(0).api.nodeInfo();
            if (nodeInfo.route_table.length == count - 1) {
                console.log("√ all nodes have started");
                break;
            }
            sleep(3000);
        }
        while (true) {
            var block = JSON.parse(nodes.RPC(0).api.blockDump(1).data)[0];
            // console.log(block);
            if (block.height > 3) {
                console.log("√ the height of current tail block is higher than 3");
                done();
                break;
            }
            sleep(3000);
        }
    });

    it('test downloader', function(done){
        this.timeout(150000);
        nodes.RPC(5).admin.changeNetworkID(10);
        sleep(15000);
        var blockA;
        nodes.RPC(0).api.blockDump(1, function (err, resp) {
            blockA = JSON.parse(resp.data)[0];
        });

        var blockB;
        nodes.RPC(5).api.blockDump(1, function (err, resp) {
            blockB = JSON.parse(resp.data)[0];
        });
        sleep(3000);
        expect(blockA.hash).not.to.be.equal(blockB.hash);
        console.log("√ changed a node networkID to 10 and the node go to forked");


        nodes.RPC(5).admin.changeNetworkID(1);
        sleep(20000);
        nodes.RPC(0).api.blockDump(1, function (err, resp) {
            blockA = JSON.parse(resp.data)[0];
        });

        // var blockSeed = JSON.parse(nodes.RPC(0).api.blockDump(1).data)[0];
        nodes.RPC(5).api.blockDump(1, function (err, resp) {
            blockB = JSON.parse(resp.data)[0];
        });
        sleep(3000);
        expect(blockA.hash).to.be.equal(blockB.hash);
        console.log("√ recover the node networkID to 1 and the node is synchronized");
        sleep(10000);
        done();
    });
});

describe('quit', function () {
    it('quit', function () {
        setTimeout(function() {
            nodes.Stop();
        }, 20000);  
    });
});