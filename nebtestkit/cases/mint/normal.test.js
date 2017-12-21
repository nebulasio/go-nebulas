'use strict';

var Neblet = require('../../neblet');
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

var nodes = new Array();
for (var i = 0; i < miners.length; i++) {
    var server = new Neblet(
        local, port + i, http_port + i, rpc_port + i,
        miners[i], miners[i], 'passphrase'
    );
    var node = server.NebJs();
    if (i == 0) {
        server.Init()
    } else {
        server.Init(nodes[0])
    }
    nodes.push(node);
}