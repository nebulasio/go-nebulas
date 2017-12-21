'use strict';

var Neblet = require('../../neblet');
var expect = require('chai').expect;
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

describe('seed server start correctly', function () {
    before(function (done) {
        this.timeout(6000);
        setTimeout(done, 5000);
    });

    it('check status', function () {
        var nodeInfo = agents[0].api.nodeInfo()
        expect(nodeInfo.chain_id).to.be.equal(100);
    });

    it('start normal servers', function () {
        for (var i = 1; i < servers.length; i++) {
            console.log(i);
            var node = servers[i].Start()
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
            var nodeInfo = agents[i].api.nodeInfo()
            expect(nodeInfo.chain_id).to.be.equal(100);
            // expect(nodeInfo.route_table[0]).to.be.have.property('id').equals('QmPyr4ZbDmwF1nWxymTktdzspcBFPL6X1v3Q5nT7PGNtUN');
            expect(nodeInfo.route_table).to.be.have.contains({ "id": "QmPyr4ZbDmwF1nWxymTktdzspcBFPL6X1v3Q5nT7PGNtUN", "address": ["/ip4/127.0.0.1/tcp/10000"] });
        }
    });
});

describe('quit', function () {
    it('quit', function () {
        for (var i = 0; i < nodes.length; i++) {
            nodes[i].kill('SIGINT');
        }
    });
});