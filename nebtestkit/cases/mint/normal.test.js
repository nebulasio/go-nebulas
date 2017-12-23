'use strict';

var Neblet = require('../../neblet');
var expect = require('chai').expect;
var sleep = require("system-sleep")
var os = require('os');

var local = "127.0.0.1";
var port = 10000;
var http_port = 8090;
var rpc_port = 9090;
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
]

var agents = new Array();
var servers = new Array();
var nodes = new Array();
for (var i = 0; i < miners.length; i++) {
    var server = new Neblet(
        local, port + i, http_port + i, rpc_port + i,
        miners[i], miners[i], 'passphrase'
    );
    var agent = server.NebJs();
    if (i == 0) {
        server.Init()
    } else {
        server.Init(servers[0])
    }
    servers.push(server);
    agents.push(agent);
}

var node = servers[0].Start()
console.log(0);
nodes.push(node)
sleep(5000)

/* describe('seed server start correctly', function () {
    before(function (done) {
        this.timeout(10000);
        setTimeout(done, 4000);
    });

    it('check status', function () {
        var nodeInfo = agents[0].api.nodeInfo()
        expect(nodeInfo.chain_id).to.be.equal(100);
    });

    it('start servers', function () {
        for (var i = 1; i < servers.length; i++) {
            var node = servers[i].Start()
            console.log(i);
            nodes.push(node)
        }
    });
});

describe('normal servers start correctly', function () {
    before(function (done) {
        this.timeout(10000);
        setTimeout(done, 8000);
    });

    it('check status', function () {
        for (var i = 1; i < agents.length; i++) {
            console.log(i);
            var nodeInfo = agents[i].api.nodeInfo()
            expect(nodeInfo.chain_id).to.be.equal(100);
            expect(nodeInfo.route_table).to.be.have.contains({ "id": "QmPyr4ZbDmwF1nWxymTktdzspcBFPL6X1v3Q5nT7PGNtUN", "address": ["/ip4/127.0.0.1/tcp/10000"] });
        }
    });
}); */

/* describe('check dump blocks', function () {
    it('check no empty slots', function () {
        var size = miners.length;
        var block = null;
        for (var i = 0; i < size; i++) {
            var now = new Date();
            var blocks = JSON.parse(agents[0].api.blockDump(1).data);
            expect(blocks.length).to.be.equal(1);
            console.log(now);
            console.log(new Date(blocks[0].timestamp * 1000))
            console.log(blocks[0]);
            if (block == null) {
                block = blocks[0]
            } else if (block.height > 1) {
                expect(blocks[0].height).to.be.gt(block.height)
                expect(blocks[0].timestamp - block.timestamp).to.be.equal((blocks[0].height - block.height) * 5)
                expect(now.getTime() / 1000).to.be.lte(blocks[0].timestamp + 7)
                block = blocks[0]
            }
            sleep(5000);
        }
    });
}) */

/* describe('quit', function () {
    it('quit', function () {
        for (var i = 0; i < nodes.length; i++) {
            nodes[i].kill('SIGINT');
        }
    });
}); */

for (var i = 1; i < servers.length; i++) {
    var node = servers[i].Start()
    console.log(i);
    nodes.push(node)
}
sleep(10000)

for (var i = 0; i < nodes.length; i++) {
    nodes[i].kill('SIGINT');
}