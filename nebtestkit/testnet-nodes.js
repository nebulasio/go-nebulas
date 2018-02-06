'use strict';

var Neb = require('../cmd/console/neb.js/lib/neb.js');
var HttpRequest = require("./node-request");
var LocalNodes = require("./local-nodes");
var sleep = require("system-sleep");

// var testnet = ["https://testnet.nebulas.io"];
var testnet = ["http://35.182.48.19:8685"];

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
        var local = this.local.RPC(0);
        var nodes = this.nodes;
        return local.admin.unlockAccount(from, "passphrase").then(function (resp) {
            console.log("unlock:" + JSON.stringify(resp));
            return local.admin.signTransaction(from, to, value, nonce, gasprice, gaslimit, contract, candidate, delegate);
        }).then(function (resp) {
            // console.log(resp);
            return nodes[0].api.sendRawTransaction(resp.data);
        });
    },
};

module.exports = Node;