'use strict';

var Wallet = require('nebulas');
var HttpRequest = require("../../node-request");
var sleep = require("system-sleep");
var FS = require("fs");
var TestNetConfig = require("../testnet_config.js");
var expect = require('chai').expect;

var env; // local testneb1 testneb2
var AddressNumber;
var SendTimes;
var ContractNumber;
var testType //1:binary 2:合约调用 3:多级合约调用

var args = process.argv.splice(2);

if (args.length != 5) {
    // give default config
    env = "testneb2";
    AddressNumber = 4000; 
    SendTimes = 10;
    ContractNumber = 20;
    testType = 1;
} else {
    env = args[0]; // local testneb1 testneb2

    AddressNumber = parseInt(args[1]);
    SendTimes = parseInt(args[2]);
    ContractNumber = parseInt(args[3]);
    testType = parseInt(args[4]);
}
var testNetConfig = new TestNetConfig(env);

if (AddressNumber <= 0 || SendTimes <= 0) {

    console.log("please input correct AddressNumber and SendTimes");
    return;
}
var Neb = Wallet.Neb;
var Transaction = Wallet.Transaction;
var from;
var accountArray;
var to = Wallet.Account.NewAccount();
var toAddresses;
var nonce = 0;

// statics for tps check start time.
var startTime;

var nodes = new Array();

var neb = new Neb();
var ChainID = testNetConfig.ChainId;
var sourceAccount = testNetConfig.sourceAccount;
var coinbase = testNetConfig.coinbase;
var apiEndPoint = testNetConfig.apiEndPoint;
neb.setRequest(new HttpRequest(apiEndPoint));

var calleeContractSrc = FS.readFileSync("nf/nvm/test/kvStore.js", "utf-8");
var callerContractSrc = FS.readFileSync("nf/nvm/test/kvStoreProxy.js", "utf-8");
var deployContractSrc = FS.readFileSync("nebtestkit/cases/stress/game.ts", "utf-8");
var calleeContractAddresses;
var callerContractAddresses;
var to = Wallet.Account.NewAccount();

var maxCheckTime = 60;
var checkTimes = 0;
var beginCheckTime;

function checkTransaction(hash, callback) {
    if (checkTimes === 0) {
        beginCheckTime = new Date().getTime();
    }
    checkTimes += 1;
    if (checkTimes > maxCheckTime) {
        console.log("check tx receipt timeout:" + hash);
        checkTimes = 0;
        callback();
        return;
    }

    neb.api.getTransactionReceipt(hash).then(function (resp) {

        console.log("tx receipt status:" + resp.status);
        if (resp.status === 2) {
            setTimeout(function () {
                checkTransaction(hash, callback);
            }, 2000);
        } else {
            checkTimes = 0;
            var endCheckTime = new Date().getTime();
            console.log("check tx time: : " + (endCheckTime - beginCheckTime) / 1000);
            callback(resp);
        }
    }).catch(function (err) {
        console.log("fail to get tx receipt hash: " + hash);
        console.log("it may becuase the tx is being packing, we are going on to check it!");
        // console.log(err);
        setTimeout(function () {
            checkTransaction(hash, callback);
        }, 2000);
    });
}

if (testType == 4) {
    console.log("test from deploy contract");
    testBinary();
} else if (testType == 3) {
    console.log("test for muti-nvm");
    deployContracts();
} else if (testType == 2) {
    console.log("test for normal call contracts");
    deployContracts();
} else if (testType == 1) {
    console.log("test for noraml binary");
    testBinary();
}

function testBinary() {
    toAddresses = new Array();
    for (var i = 0; i < ContractNumber; i++) {
        toAddresses.push(Wallet.Account.NewAccount());
    }
    neb.api.getAccountState(sourceAccount.getAddressString()).then(function(resp){
        nonce = parseInt(resp.nonce);
        claimTokens();
    })
}

function deployContracts() {
    calleeContractAddresses = new Array();
    callerContractAddresses = new Array();
    try {
        var calleeContract = {
            "source": calleeContractSrc,
            "sourceType": "js",
            "args": ''
        };
        var callerContract = {
            "source": callerContractSrc,
            "sourceType": "js",
            "args": ''
        };
        neb.api.getAccountState(sourceAccount.getAddressString()).then(function (resp) {
            console.log("----step0. get source account state: " + JSON.stringify(resp));
            nonce = parseInt(resp.nonce);

            for (var i = 0; i < ContractNumber - 1; i++) {
                var tx = new Transaction(ChainID, sourceAccount, sourceAccount, 0, ++nonce, 1000000, 20000000, calleeContract);
                tx.signTransaction();
                neb.api.sendRawTransaction(tx.toProtoString()).then(function (resp) {
                    calleeContractAddresses.push(resp.contract_address);
                });

                tx = new Transaction(ChainID, sourceAccount, sourceAccount, 0, ++nonce, 1000000, 20000000, callerContract);
                tx.signTransaction();
                neb.api.sendRawTransaction(tx.toProtoString()).then(function (resp) {
                    callerContractAddresses.push(resp.contract_address);
                })
            }
            var tx = new Transaction(ChainID, sourceAccount, sourceAccount, 0, ++nonce, 1000000, 20000000, calleeContract);
            tx.signTransaction();
            return neb.api.sendRawTransaction(tx.toProtoString());
        }).then(function (resp) {
            console.log("----step1. deploy last callee contract: " + JSON.stringify(resp));
            calleeContractAddresses.push(resp.contract_address);

            var tx = new Transaction(ChainID, sourceAccount, sourceAccount, 0, ++nonce, 1000000, 20000000, callerContract);
            tx.signTransaction();
            console.log(tx.contract);
            return neb.api.sendRawTransaction(tx.toProtoString());
        }).then(function (resp) {
            console.log("----step2. deploy last caller contract: " + JSON.stringify(resp));
            callerContractAddresses.push(resp.contract_address);
            checkTransaction(resp.txhash, function (resp) {
                try {
                    expect(resp).to.not.be.a('undefined');
                    expect(resp.status).to.be.equal(1);
                    console.log("----step3. have been on chain, to claim tokens");
                    claimTokens();
                } catch (err) {
                    console.log("check tx err :" + err);
                    return;
                }
            });
        }).catch(function (err) {
            console.log("unexpected err: " + err);
        });
    } catch (err) {
        console.log("unexpected err: " + err);
    }
}


function claimTokens() {
    accountArray = new Array();
    for (var i = 0; i < AddressNumber; i++) {
        var account = Wallet.Account.NewAccount();
        accountArray.push(account);

        sendTransaction(0, 1, sourceAccount, account, "1000000000000000", ++nonce);
        sleep(10);
    }
    checkClaimTokens();
}

function sendTransaction(index, totalTimes, from, to, value, nonce) {
    if (index < totalTimes) {
        var transaction = new Wallet.Transaction(ChainID, from, to, value, nonce);
        transaction.signTransaction();
        var rawTx = transaction.toProtoString();
        neb.api.sendRawTransaction(rawTx).then(function (resp) {
            console.log("send raw transaction (claim token)resp:" + JSON.stringify(resp));
            if (resp.txhash) {
                sendTransaction(++index, totalTimes, from, to, value, ++nonce);
            }
        });
    }
}

function checkClaimTokens() {
    console.log("=========", nonce);
    var interval = setInterval(function () {
        neb.api.getAccountState(sourceAccount.getAddressString()).then(function (resp) {
            console.log("check claim token")
            console.log("master accountState resp:" + JSON.stringify(resp));
            if (resp.nonce >= nonce) {
                clearInterval(interval);
                sendTransactionsForTps();
            }
        });

    }, 2000);
}

function sendTransactionsForTps() {

    console.log("start tps transaction sending...");

    startTime = new Date().getTime();

    console.log(SendTimes);
    console.log(AddressNumber);

    for (var i = 0; i < AddressNumber; i++) {
        for (var j = 0; j < SendTimes; j++) {
            var contract_index = i % ContractNumber;
            callMutiLevelNvm(accountArray[i], j + 1, contract_index);
            // neb.setRequest(new HttpRequest(node));

            // sendTransaction(0, SendTimes, accountArray[i], to, "1", 1);
            //to call()
            //sleep(10);
        };
        console.log("sending transaction... address number: ", i);
        sleep(10);
    }

    checkTps();
}

var callTimes = 0
var receiveReciptTimes = 0;

function callMutiLevelNvm(from, nonce, contract_index) {
    // console.log(callTimes);
    callTimes += 1;
    var tx;
    var contract;

    if (testType == 3) {
        contract = {
            "function": "testTpsForMutiNvm",
            "args": "[\"" + calleeContractAddresses[contract_index] + "\"]"
        };
    } else if (testType == 2) {

        contract = {
            "function": "testTpsForNormalCall",
            "args": "[]"
        };
    } else if (testType == 4) {
        contract = {
            "source": deployContractSrc,
            "sourceType": "ts",
            "args": "[]"
        };
    }


    //send 1 wei by the way
    if (testType == 3 || testType == 2) {
        tx = new Transaction(ChainID, from, callerContractAddresses[contract_index], 1, nonce, 1000000, 2000000, contract);
    } else if (testType == 1) {
        tx = new Transaction(ChainID, from, toAddresses[contract_index], 1, nonce, 1000000, 2000000);
    } else if (testType == 4) {
        tx = new Transaction(ChainID, from, from, 1, nonce, 1000000, 2000000, contract);
    } else {
        throw "no test type";
    }
    // var tx = new Transaction(ChainID, from, to, 1, nonce, 1000000, 2000000);

    // tx.to = contractAddress;
    tx.signTransaction();
    // console.log("silent_debug");
    try {
     neb.api.sendRawTransaction(tx.toProtoString());
    // .then(function(resp) {
    //     ++receiveReciptTimes;
    //     // console.log(JSON.stringify(resp));
    // }).catch(function(err) {
    //     ++receiveReciptTimes;
    // });
    } catch (err) {
        console.log(err);
    }
}

function checkTps() {

    // console.log("send times: ", callTimes)
    // console.log("receive times: ", receiveReciptTimes);

    // var times = 1;
    // var interval = setInterval(function () {
    //     neb.api.getAccountState(to.getAddressString()).then(function (resp) {
    //         console.log("to address state:" + JSON.stringify(resp), "times :", times);
    //         times += 1;
    //         if (resp.balance >= receiveReciptTimes) {
    //             clearInterval(interval);

    //             var endTime = new Date().getTime();

    //             console.log("====================");
    //             console.log("env is ", env);
    //             console.log("concurrency number is ", AddressNumber);
    //             console.log("total number is ", AddressNumber * SendTimes);
    //             console.log("tps is: ", parseInt(resp.balance) / ((endTime - startTime) / 1000));
    //             console.log("====================")
    //         }
    //     });

    // }, 1000);
}