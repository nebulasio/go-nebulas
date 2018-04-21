'use strict';

var Wallet = require("nebulas");
//var HttpRequest = Wallet.HttpRequest
var HttpRequest = require("../../node-request.js");
var utils = Wallet.Utils;
var schedule = require('node-schedule');
var sleep = require("system-sleep");

var env; // local testneb1 testneb2
var AddressNumber = 300;
var EachAccountSendTimes = 50;

var args = process.argv.splice(2);
env = args[0];

var Neb = Wallet.Neb;
var neb = new Neb();

var ChainID;
var from;
var accountArray;
// var to = Wallet.Account.NewAccount();
var lastnonce = 0;

// statics for tps check start time.
var startTime;

var nodes = new Array();

console.log(args);

//local
if (env === 'local') {
    neb.setRequest(new HttpRequest("http://127.0.0.1:8685")); //https://testnet.nebulas.io
    ChainID = 100;
    from = new Wallet.Account("a6e5eb290e1438fce79f5cb8774a72621637c2c9654c8b2525ed1d7e4e73653f");
    nodes.push("http://127.0.0.1:8685");
} else if (env === 'testneb1') {
    neb.setRequest(new HttpRequest("http://35.182.48.19:8685"));
    ChainID = 1001;
    from = new Wallet.Account("43181d58178263837a9a6b08f06379a348a5b362bfab3631ac78d2ac771c5df3");
    nodes.push("http://35.182.48.19:8685");
    nodes.push("http://13.57.245.249:8685");
    nodes.push("http://54.219.151.126:8685");
    nodes.push("http://18.218.165.90:8685");
    nodes.push("http://18.219.28.97:8685");
    nodes.push("http://13.58.44.3:8685");
    nodes.push("http://35.177.214.138:8685");
    nodes.push("http://35.176.94.224:8685");
} else if (env === "testneb2") {
    neb.setRequest(new HttpRequest("http://34.205.26.12:8685"));
    ChainID = 1002;
    from = new Wallet.Account("25a3a441a34658e7a595a0eda222fa43ac51bd223017d17b420674fb6d0a4d52");
    nodes.push("http://34.205.26.12:8685");
    // nodes.push("http://54.206.9.246:8685");
    // nodes.push("http://54.252.158.117:8685");
    // nodes.push("http://34.206.53.244:8685");
    // nodes.push("http://34.205.53.3:8685");
    // nodes.push("http://52.3.226.40:8685");

} else if (env === "testneb3") {

    //neb.setRequest(new HttpRequest("http://52.47.199.42:8685"));
    neb.setRequest(new HttpRequest("http://35.177.214.138:8685"));
    //neb.setRequest(new HttpRequest("http://13.127.227.177:8685"));
    ChainID = 1003;
    from = new Wallet.Account("25a3a441a34658e7a595a0eda222fa43ac51bd223017d17b420674fb6d0a4d52");
    //nodes.push("http://13.127.227.177:8685");
    nodes.push("http://35.177.214.138:8685");
    //nodes.push("http://52.47.199.42:8685");

} else if (env === "testneb4") {
    neb.setRequest(new HttpRequest("http://34.208.233.164:8685"));
    ChainID = 1004;
    from = new Wallet.Account("25a3a441a34658e7a595a0eda222fa43ac51bd223017d17b420674fb6d0a4d52");
    nodes.push("http://34.208.233.164:8685");
    nodes.push("http://54.245.29.152:8685");
    nodes.push("http://52.34.73.0:8685");
    nodes.push("http://54.71.175.99:8685");
    nodes.push("http://34.213.130.120:8685");
    nodes.push("http://18.197.107.228:8685");
    nodes.push("http://18.197.106.150:8685");
    nodes.push("http://54.93.121.146:8685");
    nodes.push("http://18.195.159.210:8685");
    nodes.push("http://18.197.157.46:8685");
    nodes.push("http://18.228.3.118:8685");
    nodes.push("http://18.231.173.99:8685");
    nodes.push("http://18.231.124.140:8685");
    nodes.push("http://18.231.183.193:8685");
    nodes.push("http://18.231.162.23:8685");
    nodes.push("http://34.253.237.122:8685");
    nodes.push("http://34.244.129.30:8685");
    nodes.push("http://54.229.241.235:8685");
    nodes.push("http://54.229.177.109:8685");
    nodes.push("http://34.250.18.201:8685");
    nodes.push("http://13.127.227.177:8685");

} else if (env === "liuliang") {
    neb.setRequest(new HttpRequest("http://35.154.108.11:8685"));
    ChainID = 1001;
    from = new Wallet.Account("c75402f6ffe6edcc2c062134b5932151cb39b6486a7beb984792bb9da3f38b9f");
    nodes.push("http://35.154.108.11:8685");

} else if (env === "maintest"){
    ChainID = 2;
    from = new Wallet.Account("d2319a8a63b1abcb0cc6d4183198e5d7b264d271f97edf0c76cfdb1f2631848c");
    neb.setRequest(new HttpRequest("http://54.149.15.132:8685"));
    nodes.push("http://54.149.15.132:8685");
    nodes.push("http://18.188.27.35:8685");
    nodes.push("http://34.201.23.199:8685");
    nodes.push("http://13.251.33.39:8685");
    nodes.push("http://52.56.55.238:8685");

} else if (env === "testnet_cal_super") {
    neb.setRequest(new HttpRequest("http://13.57.96.40:8685"));
    ChainID = 1001;
    from = new Wallet.Account("25a3a441a34658e7a595a0eda222fa43ac51bd223017d17b420674fb6d0a4d52");
    // from = new Wallet.Account("d2319a8a63b1abcb0cc6d4183198e5d7b264d271f97edf0c76cfdb1f2631848c");

    nodes.push("http://52.60.150.236:8685");
    nodes.push("http://13.57.96.40:8685");

} else {
    console.log("please input correct env local testneb1 testneb2");
    return;
}


var count = AddressNumber * EachAccountSendTimes;

var interval = setInterval(function () {

    if (count >= AddressNumber * EachAccountSendTimes) {
        count = 0;

        console.log("sending request to ", env);
        neb.api.getAccountState(from.getAddressString()).then(function (resp) {
            console.log("master accountState resp:" + JSON.stringify(resp));
            //lastnonce = parseInt(resp.result.nonce);
            lastnonce = parseInt(resp.nonce);
            console.log("lastnonce:", lastnonce);

            claimTokens(lastnonce);
        });
    }

}, 20000);






function claimTokens(nonce) {
    console.log("initializing " + AddressNumber + " accounts with coins !!!")
    console.log(from.getAddressString());

    accountArray = new Array();
    for (var i = 0; i < AddressNumber; i++) {
        var account = Wallet.Account.NewAccount();
        accountArray.push(account);
        sendTransaction(0, 1, from, account, "1000000000000000", ++nonce);
        sleep(5);
    }
    checkClaimTokens();
}

function sendTransaction(index, totalTimes, from, to, value, nonce, randomToAddr) {
    if (index < totalTimes) {

        if (randomToAddr !== null && randomToAddr === true){
            var randomTo = Math.floor((Math.random() * AddressNumber));
            to = accountArray[randomTo];
        }

        var transaction = new Wallet.Transaction(ChainID, from, to, value, nonce);
        transaction.signTransaction();
        var rawTx = transaction.toProtoString();

        var i = Math.floor((Math.random() * nodes.length));
        var node = nodes[i];
        neb.setRequest(new HttpRequest(node));
        count ++;
        neb.api.sendRawTransaction(rawTx).then(function (resp) {
            console.log("send raw tx resp:" + JSON.stringify(resp) + " (" + index + "/" + totalTimes + ")");
            //if (resp.result.txhash) {
            if (resp.txhash) {
                if (nonce % 10 === 0){
                    sleep(20);
                }
                sendTransaction(++index, totalTimes, from, to, value, ++nonce, randomToAddr);
            }
        });

        //     .catch(function (err) {
        //     console.log(err);
        //     console.log(JSON.stringify(err));
        //     console.log("send tx error, retry: " + "from:" + from.getAddressString() +  " tx_index: (" + index + "/" + totalTimes + ")" + " node:" + node);
        //     sleep(10);
        //     sendTransaction(index, totalTimes, from, to, value, nonce, randomToAddr);
        // }
        // );
    }
}

function checkClaimTokens() {
    var interval = setInterval(function () {
        neb.api.getAccountState(from.getAddressString()).then(function (resp) {
            console.log("master accountState resp:" + JSON.stringify(resp));
            //if (resp.result.nonce >= lastnonce + AddressNumber) {
            if (resp.nonce >= lastnonce + AddressNumber) {
                clearInterval(interval);
                sendTransactionsForTps();
            }
        });
    }, 2000);
}

function sendTransactionsForTps() {

    console.log("start tps transaction sending...");

    startTime = new Date().getTime();

    for (var i = 0; i < AddressNumber; i++) {
        var node = nodes[i % nodes.length];
        neb.setRequest(new HttpRequest(node));
        var randomValue = Math.floor((Math.random() * 100));
        sendTransaction(0, EachAccountSendTimes, accountArray[i], null, randomValue, 1, true /*random to addr*/);
        sleep(5);
    }
}
