'use strict';

var Wallet = require('../../../cmd/console/neb.js/lib/wallet.js');
var HttpRequest = require("../../node-request");
var schedule = require('node-schedule');
var sleep = require("system-sleep");

var env; // local testneb1 testneb2
var AddressNumber = 100;
var EachAccountSendTimes = 100;

var args = process.argv.splice(2);

if (args.length != 3) {
    // give default config
    env = "local";
    AddressNumber = 100;
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
    nodes.push("http://13.57.120.136:8685");
    nodes.push("http://18.218.165.90:8685");
    nodes.push("http://35.177.214.138:8685");
    nodes.push("http://35.176.94.224:8685");
    nodes.push("http://35.182.205.40:8685");
    nodes.push("http://52.47.199.42:8685");

}else {
    console.log("please input correct env local testneb1 testneb2")
    return;
}

var j = schedule.scheduleJob('10,40 */1 * * *', function() {
    neb.api.getAccountState(from.getAddressString()).then(function (resp) {
        console.log("master accountState resp:" + JSON.stringify(resp));
        lastnonce = parseInt(resp.nonce);
        console.log("lastnonce:", lastnonce);

        //claimTokens(lastnonce);
        claimSubMaster(lastnonce);
    });
});

var maxCliamTime = 2
var cliamTimes = 0
var subMaster
//claim a new account to distribute money instead of master account
function claimSubMaster(nonce){
    console.log("initializing the subMaster account to distribute coins...");
    subMaster = Wallet.Account.NewAccount();
    var value = (maxCliamTime * AddressNumber + 1) * 1000000000000000;

    var transaction = new Wallet.Transaction(ChainID, from, subMaster, value.toString(), ++nonce);
    transaction.signTransaction();
    var rawTx = transaction.toProtoString();
    cliamTimes ++;
    neb.api.sendRawTransaction(rawTx)
        .then(function (rawTxResp){
            console.log("\tresp of send raw Tx for claiming subMaster:" + JSON.stringify(rawTxResp));
            checkTransaction(rawTxResp.txhash, function (resp) {
                //console.log("resp" + JSON.stringify(resp));
                try {
                    if (resp && resp.status === 1) {
                        cliamTimes = 0;
                        console.log("send TX to sumMaster account success.");
                        claimTokens(0);  //thr nonce of a new account is 0
                    } else if (cliamTimes < maxCliamTime) {
                        claimSubMaster(nonce);
                    } else {
                        cliamTimes = 0;
                        console.log("claim sumMaster failed!!!!");
                    }
                }catch (err) {
                    console.log(JSON.stringify(err));
                    console.log(err);
                }
            });
        }).catch(function (err) {
            console.log(err);
            claimSubMaster(nonce);
        });
}

var maxCheckTime = 20;
var checkTimes = 0;
//check tx result to make sure the tx is completed
function checkTransaction(hash, callback) {
    checkTimes += 1;
    if (checkTimes > maxCheckTime) {
        console.log("\tcheck tx receipt timeout:" + hash);
        checkTimes = 0;
        callback();
        return;
    }

    neb.api.getTransactionReceipt(hash).then(function (resp) {
        console.log("\ttx receipt status:" + resp.status);
        if (resp.status === 2) {
            setTimeout(function () {
                checkTransaction(hash, callback);
            }, 2000);
        } else {
            checkTimes = 0;
            callback(resp);
        }
    }).catch(function (err) {
        console.log("\tfail to get tx receipt hash: " + hash);
        console.log("\tit may because the tx is being packing, we are going on to check it!");
        console.log("\t" + err.error);
        setTimeout(function () {
            checkTransaction(hash, callback);
        }, 2000);
    });
}

function claimTokens(nonce) {
    console.log("initializing " + AddressNumber + " accounts with coins !!!")
    accountArray = new Array();
    for (var i = 0; i < AddressNumber; i++) {
        var account = Wallet.Account.NewAccount();
        accountArray.push(account);

        sendTransaction(0, 1, subMaster, account, "1000000000000000", ++nonce);

        sleep(10);
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
        neb.api.sendRawTransaction(rawTx).then(function (resp) {
            console.log("\tsend raw transaction resp:" + JSON.stringify(resp));
            if (resp.txhash) {
                if (nonce % 10 === 0){
                    sleep(10);
                }
                sendTransaction(++index, totalTimes, from, to, value, ++nonce, randomToAddr);
            }
        });
    }
}

function checkClaimTokens() {
    var interval = setInterval(function () {
        neb.api.getAccountState(subMaster.getAddressString()).then(function (resp) {
            console.log("\tsubMaster accountState resp:" + JSON.stringify(resp));
            if (resp.nonce >=  AddressNumber) {
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

        sendTransaction(0, EachAccountSendTimes, accountArray[i], null, "0.001", 1, true /*random to addr*/);
        sleep(10);
    }
}