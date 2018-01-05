'use strict';

var Neb = require('../cmd/console/neb.js/lib/neb.js');
var HttpRequest = require("./node-request");
var LocalNodes = require("./local-nodes");
var sleep = require("system-sleep");

var testnet = ["https://testnet.nebulas.io"];

var port = 8680;

var Node = function () {
    this.local = new LocalNodes(1);
    this.local.Start();
};

Node.prototype = {
    Start: function () {
        var nodes = new Array();
        for (var i = 0; i < testnet.length; i++) {
            var httpRequest = new HttpRequest(testnet[i]);
            nodes.push(new Neb(httpRequest));
        }
        this.nodes = nodes;
    },
    Stop: function () {
        this.local.Stop();
    },
    RPC: function (index) {
        return this.nodes[index];
    },
    SendTransaction: function (from, to, value, nonce, gasprice, gaslimit, contract, candidate, delegate) {
        var node = this.local.RPC(0);
        var nodes = this.nodes;
        let data = "";
        node.admin.unlockAccount(from, "passphrase");
        sleep(1000);
        return node.admin.signTransaction(from, to, value, nonce, gasprice, gaslimit, contract, candidate, delegate)
            .then(function (resp) {
                return nodes[0].api.sendRawTransaction(resp.data);
            });
    },
};

module.exports = Node;