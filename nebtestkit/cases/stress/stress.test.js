'use strict';

var TestnetNodes = require('../../testnet-nodes');
var Wallet = require('../../../cmd/console/neb.js/lib/wallet.js');
var sleep = require("system-sleep");
const https = require("https");

const AddressNumber = 5;
const SendTimes = 1;


var nodes = new TestnetNodes();
nodes.Start();

var master = Wallet.Account.NewAccount();
var from = master.getAddressString();

var email = Math.random() + "test@demo.io";
var url = "https://testnet.nebulas.io/claim/api/claim/" + email + "/" + from + "/";
https.get(url, res => {});

sleep(10000);
nodes.RPC(0).api.getAccountState(from).then(function (resp) {
    console.log("master accountState resp:" + JSON.stringify(resp));
});

var accountArray = new Array();
for (var i = 0; i < AddressNumber; i++) {
    var account = Wallet.Account.NewAccount();
    var hash = account.getAddressString();
    accountArray.push(hash);
}

var nonce = 1;
var t1 = new Date().getTime();
for (var j = 0; j < AddressNumber; j++) {
    for (var k = 0; k < SendTimes; k++) {
        var transaction = new Wallet.Transaction(1001, master, accountArray[j], "1", nonce++);
        transaction.signTransaction();
        var rawTx = transaction.toProtoString();
        nodes.RPC(0).api.sendRawTransaction(rawTx).then(function (resp) {
            console.log("send raw transaction resp:" + JSON.stringify(resp));
        });
        sleep(10);
    }
}

sleep(5000);

var interval = setInterval(function () {
    for (var i = 0; i < AddressNumber; i++) {
        nodes.RPC(0).api.getAccountState(accountArray[i]).then(function (resp) {
            console.log("accountState resp:" + JSON.stringify(resp));
        });
    }
    nodes.RPC(0).api.getAccountState(from).then(function (resp) {
        console.log("master accountState resp:" + JSON.stringify(resp));
        if (resp.nonce == AddressNumber * SendTimes) {
            var t2 = new Date().getTime();
            console.log("Time consumption：" + (t2 - t1) / 1000);
            console.log("Tps：" + AddressNumber * SendTimes * 1000 / (t2 - t1));
            clearInterval(interval);
        }
    });

}, 2000);