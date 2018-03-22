'use strict';

var Wallet = require('../../../cmd/console/neb.js/lib/wallet.js');
var HttpRequest = require("../../node-request");
var schedule = require('node-schedule');
var sleep = require("system-sleep");

var env; // local testneb1 testneb2
var AddressNumber = 1200;
var EachAccountSendTimes = 1000;

var args = process.argv.splice(2);

if (args.length != 3) {
    // give default config
    env = "testneb3";
} else {
    env = args[0]; // local testneb1 testneb2

    AddressNumber = parseInt(args[1]);
    EachAccountSendTimes = parseInt(args[2]);
}

if (AddressNumber <= 0 || EachAccountSendTimes <= 0) {

    console.log("please input correct AddressNumber and SendTimes");
    return;
}

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

//local
if (env == 'local') {
    neb.setRequest(new HttpRequest("http://127.0.0.1:8685")); //https://testnet.nebulas.io
    ChainID = 100;
    from = new Wallet.Account("a6e5eb290e1438fce79f5cb8774a72621637c2c9654c8b2525ed1d7e4e73653f");
    nodes.push("http://127.0.0.1:8685");
} else if (env == 'testneb1') {
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
} else if (env == "testneb2") {
    neb.setRequest(new HttpRequest("http://34.205.26.12:8685"));
    ChainID = 1002;
    from = new Wallet.Account("43181d58178263837a9a6b08f06379a348a5b362bfab3631ac78d2ac771c5df3");
    nodes.push("http://34.205.26.12:8685");
} else if (env == "testneb3") {
    neb.setRequest(new HttpRequest("http://35.177.214.138:8685"));
    ChainID = 1003;
    from = new Wallet.Account("43181d58178263837a9a6b08f06379a348a5b362bfab3631ac78d2ac771c5df3");
    nodes.push("http://35.177.214.138:8685");
    nodes.push("http://13.57.19.76:8685");
    nodes.push("http://18.218.165.90:8685");
    nodes.push("http://35.176.94.224:8685");
    nodes.push("http://35.182.205.40:8685");
    nodes.push("http://52.47.199.42:8685");

} else {
    console.log("please input correct env local testneb1 testneb2");
    return;
}

var j = schedule.scheduleJob('*/30 * * * *', function () {
    neb.api.getAccountState(from.getAddressString()).then(function (resp) {
        console.log("master accountState resp:" + JSON.stringify(resp));
        lastnonce = parseInt(resp.nonce);
        console.log("lastnonce:", lastnonce);

        claimTokens(lastnonce);
    });
});

function claimTokens(nonce) {
    console.log("initializing " + AddressNumber + " accounts with coins !!!")
    accountArray = new Array();
    for (var i = 0; i < AddressNumber; i++) {
        var account = Wallet.Account.NewAccount();
        accountArray.push(account);
        sendTransaction(0, 1, from, account, "1000000000000000", ++nonce);
        sleep(10);
    }
    checkClaimTokens();
}

function sendTransaction(index, totalTimes, from, to, value, nonce, randomToAddr) {
    if (index < totalTimes) {

        if (randomToAddr !== null && randomToAddr === true) {
            var randomTo = Math.floor((Math.random() * AddressNumber));
            to = accountArray[randomTo];
        }

        var transaction = new Wallet.Transaction(ChainID, from, to, value, nonce);
        transaction.signTransaction();
        var rawTx = transaction.toProtoString();

        var i = Math.floor((Math.random() * nodes.length));
        var node = nodes[i];
        neb.setRequest(new HttpRequest(node));
        neb.api.sendRawTransaction(rawTx).then(function (resp) {
            console.log("send raw transaction resp:" + JSON.stringify(resp));
            if (resp.txhash) {
                if (nonce % 10 === 0) {
                    sleep(2);
                }
                sendTransaction(++index, totalTimes, from, to, value, ++nonce, randomToAddr);
            }
        }).catch(function (err) {
            console.log("send tx error, retry: " + "from:" + from.getAddressString() + " tx_index: (" + index + "/" + totalTimes + ")" + " node:" + node);
            sleep(20);
            sendTransaction(index, totalTimes, from, to, value, nonce, randomToAddr);
        });
    }
}

function checkClaimTokens() {
    var interval = setInterval(function () {
        neb.api.getAccountState(from.getAddressString()).then(function (resp) {
            console.log("master accountState resp:" + JSON.stringify(resp));
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
        var randomValue = Math.floor((Math.random() * 10));
        sendTransaction(0, EachAccountSendTimes, accountArray[i], null, randomValue, 1, true /*random to addr*/ );
        sleep(20);
    }
}