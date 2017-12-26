'use strict';

var Node = require('../../Node');
var expect = require('chai').expect;
var sleep = require("system-sleep");
var Neblet = require('../../neblet');
var os = require('os');

var count = 6;
var validators = 6;
var blockInterval = 5;
var dynastyInterval = 60;

var nodes = new Node(count);
nodes.Start();

describe('start five nodes normally', function () {
    before(function (done) {
        this.timeout(10000);
        setTimeout(done, 4000);
    });

    it('check status', function () {
        while (true) {
            var nodeInfo = nodes.RPC(0).api.nodeInfo();
            console.log(nodeInfo);
            if (nodeInfo.route_table.length == count - 1) {
                break;
            }
            sleep(3000);
        }
    });

    it('change a node network ID', function(){
        sleep(10000);
        nodes.RPC(5).api.ChangeNetworkID(10, function (err, resp) {
            if (resp.result) {
                sleep(10000);
                var blockSeed;
                nodes.RPC(0).api.blockDump(1, function (err, resp) {
                    blockSeed = JSON.parse(resp.data)[0];
                });
        
                // var blockSeed = JSON.parse(nodes.RPC(0).api.blockDump(1).data)[0];
                var block;
                nodes.RPC(5).api.blockDump(1, function (err, resp) {
                    block = JSON.parse(resp.data)[0];
                });
                sleep(10000);
                expect(blockSeed.hash).not.to.be.equal(bolck.hash);
            }
        });
    });

    it('resume the node network ID', function(){
        sleep(10000);
        nodes.RPC(5).api.ChangeNetworkID(1, function (err, resp) {
            if (resp.result) {
                sleep(10000);
                var blockSeed;
                nodes.RPC(0).api.blockDump(1, function (err, resp) {
                    blockSeed = JSON.parse(resp.data)[0];
                });
        
                // var blockSeed = JSON.parse(nodes.RPC(0).api.blockDump(1).data)[0];
                var block;
                nodes.RPC(5).api.blockDump(1, function (err, resp) {
                    block = JSON.parse(resp.data)[0];
                });
                sleep(10000);
                expect(blockSeed.hash).to.be.equal(bolck.hash);
            }
        });
    });
});