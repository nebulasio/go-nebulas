'use strict';

var TestnetNodes = require('../../testnet-nodes');
var Wallet = require('../../../cmd/console/neb.js/lib/wallet.js');
var sleep = require("system-sleep");
const https = require("https");

const AddressNumber = 50;
const SendTimes = 100;
var lastnonce = 0;


var nodes = new TestnetNodes();
nodes.Start();

// var master = Wallet.Account.NewAccount();
var from = new Wallet.Account("43181d58178263837a9a6b08f06379a348a5b362bfab3631ac78d2ac771c5df3");

// var email = Math.random() + "test@demo.io";
// var url = "https://testnet.nebulas.io/claim/api/claim/" + email + "/" + from + "/";
// https.get(url, res => {});


nodes.RPC(0).api.getAccountState(from.getAddressString()).then(function (resp) {
    console.log("master accountState resp:" + JSON.stringify(resp));
    lastnonce = parseInt(resp.nonce);
    console.log("lastnonce:", lastnonce);
});

sleep(10000);

var accountArray = new Array();
for (var i = 0; i < AddressNumber; i++) {
    var account = Wallet.Account.NewAccount();
    var hash = account.getAddressString();
    accountArray.push(hash);
}

var nonce = lastnonce;
var t1 = new Date().getTime();
for (var j = 0; j < AddressNumber; j++) {
    for (var k = 0; k < SendTimes; k++) {
        var transaction = new Wallet.Transaction(1001, from, accountArray[j], "1", ++nonce);
        transaction.signTransaction();
        var rawTx = transaction.toProtoString();
        nodes.RPC(0).api.sendRawTransaction(rawTx).then(function (resp) {
            console.log("send raw transaction resp:" + JSON.stringify(resp));
        });
    }
}

sleep(2000);

var interval = setInterval(function () {
    // for (var i = 0; i < AddressNumber; i++) {
    //     nodes.RPC(0).api.getAccountState(accountArray[i]).then(function (resp) {
    //         console.log("accountState resp:" + JSON.stringify(resp));
    //     });
    // }
    nodes.RPC(0).api.getAccountState(from.getAddressString()).then(function (resp) {
        console.log("master accountState resp:" + JSON.stringify(resp));
        if (resp.nonce == lastnonce + AddressNumber * SendTimes) {
            var t2 = new Date().getTime();
            console.log("Time consumption：" + (t2 - t1) / 1000);
            console.log("Tps：" + AddressNumber * SendTimes * 1000 / (t2 - t1));
            clearInterval(interval);
        }
    });

}, 2000);