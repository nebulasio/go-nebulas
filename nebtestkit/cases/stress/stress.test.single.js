'use strict';

var HttpRequest = require("../../node-request");
var Wallet = require('../../../cmd/console/neb.js/lib/wallet.js');
var neb = new Wallet.Neb(new HttpRequest("http://127.0.0.1:8685"));


var sleep = require("system-sleep");

const AddressNumber = 200;
const SendTimes = 40;
var lastnonce = 0;

var chainID = 100;

// var master = Wallet.Account.NewAccount();
var from = new Wallet.Account("a6e5eb290e1438fce79f5cb8774a72621637c2c9654c8b2525ed1d7e4e73653f");


neb.api.getAccountState(from.getAddressString()).then(function (resp) {
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
    //     neb.api.getAccountState(accountArray[i]).then(function (resp) {
    //         console.log("accountState resp:" + JSON.stringify(resp));
    //     });
    // }
    neb.api.getAccountState(from.getAddressString()).then(function (resp) {
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
        var transaction = new Wallet.Transaction(chainID, from, address, "1", ++nonce);
        transaction.signTransaction();
        var rawTx = transaction.toProtoString();
        neb.api.sendRawTransaction(rawTx).then(function (resp) {
            console.log("send raw transaction resp:" + JSON.stringify(resp));
            sendtimes++;
            if (resp.txhash) {
                sendTransaction(sendtimes, nonce, address);
            }
        });
    }

}