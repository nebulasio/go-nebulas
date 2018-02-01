'use strict';

var TestnetNodes = require('../../testnet-nodes');
var Wallet = require('../../../cmd/console/neb.js/lib/wallet.js');
var sleep = require("system-sleep");
const https = require("https");

const AddressNumber = 500;
const SendTimes = 50;
var lastnonce = 0;


var nodes = new TestnetNodes();
nodes.Start();

// var master = Wallet.Account.NewAccount();
var from = new Wallet.Account("43181d58178263837a9a6b08f06379a348a5b362bfab3631ac78d2ac771c5df3");


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
    sendTransaction(0, nonce, accountArray[j]);
    nonce = nonce + SendTimes;
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

function sendTransaction(sendtimes, nonce, address) {
    if (sendtimes < SendTimes) {
        var transaction = new Wallet.Transaction(1002, from, address, "1", ++nonce);
        transaction.signTransaction();
        var rawTx = transaction.toProtoString();
        nodes.RPC(0).api.sendRawTransaction(rawTx).then(function (resp) {
            console.log("send raw transaction resp:" + JSON.stringify(resp));
            sendtimes++;
            if (resp.txhash) {
                sendTransaction(sendtimes, nonce, address);
            }
        });
    }

}