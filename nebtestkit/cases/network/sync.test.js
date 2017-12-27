'use strict';

var Node = require('../../Node');
var expect = require('chai').expect;
var sleep = require("system-sleep");
var Neblet = require('../../neblet');
var os = require('os');
var local = "127.0.0.1";
var port = 10000;
var http_port = 8091;
var rpc_port = 9091;
var miners = [
    "1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c",
    "2fe3f9f51f9a05dd5f7c5329127f7c917917149b4e16b0b8",
    "333cb3ed8c417971845382ede3cf67a0a96270c05fe2f700",
    "48f981ed38910f1232c1bab124f650c482a57271632db9e3",
    "59fc526072b09af8a8ca9732dae17132c4e9127e43cf2232",
    "75e4e5a71d647298b88928d8cb5da43d90ab1a6c52d0905f",
    "7da9dabedb4c6e121146fb4250a9883d6180570e63d6b080",
    "98a3eed687640b75ec55bf5c9e284371bdcaeab943524d51",
    "a8f1f53952c535c6600c77cf92b65e0c9b64496a8a328569",
    "b040353ec0f2c113d5639444f7253681aecda1f8b91f179f",
    "b414432e15f21237013017fa6ee90fc99433dec82c1c8370",
    "b49f30d0e5c9c88cade54cd1adecf6bc2c7e0e5af646d903",
    "b7d83b44a3719720ec54cdb9f54c0202de68f1ebcb927b4f",
    "ba56cc452e450551b7b9cffe25084a069e8c1e94412aad22",
    "c5bcfcb3fa8250be4f2bf2b1e70e1da500c668377ba8cd4a",
    "c79d9667c71bb09d6ca7c3ed12bfe5e7be24e2ffe13a833d",
    "d1abde197e97398864ba74511f02832726edad596775420a",
    "d86f99d97a394fa7a623fdf84fdc7446b99c3cb335fca4bf",
    "e0f78b011e639ce6d8b76f97712118f3fe4a12dd954eba49",
    "f38db3b6c801dddd624d6ddc2088aa64b5a24936619e4848",
    "fc751b484bd5296f8d267a8537d33f25a848f7f7af8cfcf6"
];

var count = 6;

var nodes = new Node(count-1);
nodes.Start();

describe('start five nodes normally', function () {
    before(function (done) {
        this.timeout(10000);
        setTimeout(done, 4000);
    });

    it('check status', function (done) {
        this.timeout(40000);
        while (true) {
            var nodeInfo = nodes.RPC(0).api.nodeInfo();
            console.log(nodeInfo);
            if (nodeInfo.route_table.length == count - 2) {
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

    it('start another node to test sync', function (done) {
        this.timeout(40000);
        var i = count-1;
        var server = new Neblet(
            local, port + i, http_port + i, rpc_port + i,
            miners[i], miners[i], 'passphrase'
        );
        server.Init(nodes.Node(0));
        server.Start();
        nodes.Nodes().push(server);
        sleep(10000);
        var blockSeed;
        nodes.RPC(0).api.blockDump(1, function (err, resp) {
            blockSeed = JSON.parse(resp.data)[0];
        });

        var block;
        server.NebJs().api.blockDump(1, function (err, resp) {
            block = JSON.parse(resp.data)[0];
        });
        sleep(3000);
        expect(blockSeed.height).to.be.equal(block.height);
        console.log("√ start a new node and the new node has synced all the blocks");
        done();
    });

    it('verify block sync', function (done) {
        this.timeout(150000);
        nodes.RPC(5).admin.changeNetworkID(10);

        sleep(10000);
        var blockA;
        nodes.RPC(0).api.blockDump(1, function (err, resp) {
            blockA = JSON.parse(resp.data)[0];
        });

        var blockB;
        nodes.RPC(5).api.blockDump(1, function (err, resp) {
            blockB = JSON.parse(resp.data)[0];
        });
        sleep(10000);
        expect(blockA.hash).not.to.be.equal(blockB.hash);
        console.log("√ changed the new node networkID to 10 and the new node go to forked");

        nodes.RPC(5).admin.changeNetworkID(1);
        sleep(60000);
        nodes.RPC(0).api.blockDump(1, function (err, resp) {
            blockA = JSON.parse(resp.data)[0];
        });

        nodes.RPC(5).api.blockDump(1, function (err, resp) {
            blockB = JSON.parse(resp.data)[0];
        });

        sleep(10000);

        expect(blockA.hash).to.be.equal(blockB.hash);
        console.log("√ recover the new node networkID to 1 and the new node is synchronized");
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