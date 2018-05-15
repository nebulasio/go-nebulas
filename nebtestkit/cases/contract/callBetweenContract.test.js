"use strict";

var Wallet = require("nebulas");
var HttpRequest = require("../../node-request.js");
var TestNetConfig = require("../testnet_config.js");
var Neb = Wallet.Neb;
var Transaction = Wallet.Transaction;
var FS = require("fs");
var expect = require('chai').expect;
var Unit = Wallet.Unit;

// mocha cases/contract/xxx testneb2 -t 2000000
var args = process.argv.splice(2);
var env = args[1];
env = 'testneb3';
var testNetConfig = new TestNetConfig(env);

var neb = new Neb();
var ChainID = testNetConfig.ChainId;
var sourceAccount = testNetConfig.sourceAccount;
var coinbase = testNetConfig.coinbase;
var apiEndPoint = testNetConfig.apiEndPoint;
neb.setRequest(new HttpRequest(apiEndPoint));

var toAddress = Wallet.Account.NewAccount();
var nonce;
var contractNonce = 0;

/*
 * set this value according to the status of your testnet.
 * the smaller the value, the faster the test, with the risk of causing error
 */

var maxCheckTime = 30;
var checkTimes = 0;
var beginCheckTime;

console.log("env:", env);

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


var calleeContractSrc = FS.readFileSync("nf/nvm/test/kvStore.js", "utf-8");
var callerContractSrc = FS.readFileSync("nf/nvm/test/kvStoreProxy.js", "utf-8"); 
var calleeContractAddress;
var callerContractAddress;
var notExistAddress = "n1hW4dGjRs8pTN1aH6TH6gAhmWYnuRAwuzE";
var calleebalance = 0;

describe('test transfer from contract', function () {
    before('0. deploy contracts', function (done) {
        try {
            neb.api.getAccountState(sourceAccount.getAddressString()).then(function(resp) {
                console.log("----step0. get source account state: " + JSON.stringify(resp));
                var calleeContract = {
                    "source": calleeContractSrc,
                    "sourceType": "js",
                    "args": ''
                };
                nonce = parseInt(resp.nonce);
                nonce = nonce + 1;
                var tx = new Transaction(ChainID, sourceAccount, sourceAccount, 0, nonce, 1000000, 20000000, calleeContract);
                tx.signTransaction();
                return neb.api.sendRawTransaction(tx.toProtoString());
            }).then(function(resp) {
                console.log("----step1. deploy callee contract: " + JSON.stringify(resp));
                calleeContractAddress = resp.contract_address;
                // checkTransaction(resp.txhash, function(resp) {
                //     expect(resp).to.not.be.a('undefined');
                //     console.log("----step2. have been on chain");
                //     done();
                // });
                
                var callerContract = {
                    "source": callerContractSrc,
                    "sourceType": "js",
                    "args": ''
                };
        
                nonce = nonce + 1;
                var tx = new Transaction(ChainID, sourceAccount, sourceAccount, 0, nonce, 1000000, 20000000, callerContract);
                tx.signTransaction();
                console.log(tx.contract);
                return neb.api.sendRawTransaction(tx.toProtoString());
            }).then(function(resp) {
                console.log("----step2. deploy caller contract: " + JSON.stringify(resp));
                callerContractAddress = resp.contract_address;
                checkTransaction(resp.txhash, function(resp) {
                    try {
                        expect(resp).to.not.be.a('undefined');
                        expect(resp.status).to.be.equal(1);
                        console.log("----step3. have been on chain");
                        done();
                    } catch(err) {
                        console.log("check tx err :" + err);
                        done(err);
                        return;
                    }
                });
            }).catch(function(err) {
                console.log("unexpected err: " + err);
                done(err);
            });
        } catch (err) {
            console.log("unexpected err: " + err);
            done(err);
        }
    });

    // it ('1. test normal call', function (done) {
    //     nonce = nonce + 1;
    //     console.log(callerContractAddress);
    //     var contract = {
    //         "function": "save",
    //         "args": "[\"" + calleeContractAddress + "\",\"msg1\", \"湖人总冠军\"]"
    //     };
    //     var tx = new Transaction(ChainID, sourceAccount, callerContractAddress, Unit.nasToBasic(2), nonce, 1000000, 2000000, contract);
    //     // tx.to = contractAddress;
    //     tx.signTransaction();
    //     // console.log("silent_debug");
    //     neb.api.sendRawTransaction(tx.toProtoString()).then(function(resp) {
    //         console.log("----step1. call callerTx ", resp);
    //         checkTransaction(resp.txhash, function(resp) {
    //             try {
    //                 expect(resp).to.not.be.a('undefined');
    //                 expect(resp.status).to.be.equal(1);
    //                 console.log("----step2. have been on chain， To check balances");
    //                 neb.api.getAccountState(callerContractAddress).then(function(state){
    //                     expect(state.balance).to.be.equal("0");
    //                     return neb.api.getAccountState(calleeContractAddress);
    //                 }).then(function(state){
                        //     calleebalance += 2000000000000000000;
                        // expect(state.balance).to.be.equal(calleebalance.toString());
    //                     console.log("----step3, to check the result");
    //                     var contract = {
    //                         "function": "get",
    //                         "args": "[\"" + calleeContractAddress + "\",\"msg1\"]"
    //                     };
    //                     return neb.api.call(sourceAccount.getAddressString(), callerContractAddress, Unit.nasToBasic(0), nonce, 1000000, 2000000, contract);
    //                 }).then(function(result){
    //                     console.log("to be deleted");
    //                     //result = {"result":"{\"key\":\"msg1\",\"value\":\"湖人总冠军\"}","execute_err":"","estimate_gas":"20511"}
    //                     expect(JSON.parse(result.result).value).equal("湖人总冠军");
    //                     done();
    //                 }).catch(function(err) {
    //                     done(err);
    //                 });
    //             } catch(err) {
    //                 console.log("check tx err :" + err);
    //                 done(err);
    //                 return;
    //             }
    //         });
    //     }).catch(function(err) {
    //         console.log("unexpected err: " + err);
    //         done(err);
    //     });
    // });


    // it ('2. not exist callee contract', function (done) {
    //     nonce = nonce + 1;
    //     console.log(callerContractAddress);
    //     var contract = {
    //         "function": "save",
    //         "args": "[\"" + notExistAddress + "\",\"msg1\", \"湖人总冠军\"]"
    //     };
    //     var tx = new Transaction(ChainID, sourceAccount, callerContractAddress, Unit.nasToBasic(10), nonce, 1000000, 2000000, contract);
    //     // tx.to = contractAddress;
    //     tx.signTransaction();
    //     // console.log("silent_debug");
    //     neb.api.sendRawTransaction(tx.toProtoString()).then(function(resp) {
    //         console.log("----step1. call callerTx ", resp);
    //         checkTransaction(resp.txhash, function(resp) {
    //             try {
    //                 expect(resp).to.not.be.a('undefined');
    //                 expect(resp.status).to.be.equal(0);
    //                 console.log("----step2. have been on chain");
    //                 neb.api.getAccountState(callerContractAddress).then(function(state){
    //                     expect(state.balance).to.be.equal("0");
    //                     return neb.api.getAccountState(calleeContractAddress);
    //                 }).then(function(state){
    //                     expect(state.balance).to.be.equal(calleebalance.toString());
    //                     return neb.api.getEventsByHash(resp.hash);
    //                 }).then(function(result){
    //                     expect(JSON.parse(result.events[0].data).error).equal("Call: no contract at this address " + notExistAddress);
    //                     done();
    //                 }).catch(function(err){
    //                     done(err);
    //                 })
    //             } catch(err) {
    //                 console.log("check tx err :" + err);
    //                 done(err);
    //                 return;
    //             }
    //         });
    //     }).catch(function(err) {
    //         console.log("unexpected err: " + err);
    //         done(err);
    //     });
    // });

    // it ('3. caller contract has not enough balance', function (done) {
    //     nonce = nonce + 1;
    //     console.log(callerContractAddress);
    //     var contract = {
    //         "function": "save",
    //         "args": "[\"" + calleeContractAddress + "\",\"msg1\", \"湖人总冠军\"]"
    //     };
    //     var tx = new Transaction(ChainID, sourceAccount, callerContractAddress, Unit.nasToBasic(1), nonce, 1000000, 2000000, contract);
    //     // tx.to = contractAddress;
    //     tx.signTransaction();
    //     // console.log("silent_debug");
    //     neb.api.sendRawTransaction(tx.toProtoString()).then(function(resp) {
    //         console.log("----step1. call callerTx ", resp);
    //         checkTransaction(resp.txhash, function(resp) {
    //             try {
    //                 expect(resp).to.not.be.a('undefined');
    //                 expect(resp.status).to.be.equal(0);
    //                 console.log("----step2. have been on chain");
    //                 neb.api.getAccountState(callerContractAddress).then(function(state){
    //                     expect(state.balance).to.be.equal("0");
    //                     return neb.api.getAccountState(calleeContractAddress);
    //                 }).then(function(state){
    //                     expect(state.balance).to.be.equal(calleebalance.toString());
    //                     return neb.api.getEventsByHash(resp.hash);
    //                 }).then(function(result){
    //                     expect(JSON.parse(result.events[0].data).error).equal("Mult Call: inner transation err [inner transfer failed] engine index:0");
    //                     done();
    //                 }).catch(function(err){
    //                     done(err);
    //                 });
    //             } catch(err) {
    //                 console.log("check tx err :" + err);
    //                 done(err);
    //                 return;
    //             }
    //         });
    //     }).catch(function(err) {
    //         console.log("unexpected err: " + err);
    //         done(err);
    //     });
    // });

    // it ('4. gasLimit is not enough', function (done) {
    //     nonce = nonce + 1;
    //     console.log(callerContractAddress);
    //     var contract = {
    //         "function": "safeSave",
    //         "args": "[\"" + calleeContractAddress + "\",\"msg1\", \"湖人总冠军\"]"
    //     };
    //     var tx = new Transaction(ChainID, sourceAccount, callerContractAddress, Unit.nasToBasic(10), nonce, 1000000, 20400, contract);
    //     // tx.to = contractAddress;
    //     tx.signTransaction();
    //     // console.log("silent_debug");
    //     neb.api.sendRawTransaction(tx.toProtoString()).then(function(resp) {
    //         console.log("----step1. call callerTx ", resp);
    //         checkTransaction(resp.txhash, function(resp) {
    //             try {
    //                 expect(resp).to.not.be.a('undefined');
    //                 expect(resp.status).to.be.equal(0);
    //                 console.log("----step2. have been on chain");
    //                 neb.api.getAccountState(callerContractAddress).then(function(state){
    //                     expect(state.balance).to.be.equal("0");
    //                     return neb.api.getAccountState(calleeContractAddress);
    //                 }).then(function(state){
    //                     expect(state.balance).to.be.equal(calleebalance.toString());
    //                     return neb.api.getEventsByHash(resp.hash);
    //                 }).then(function(result){
    //                     expect(JSON.parse(result.events[0].data).error).equal("Mult Call: inner transation err [preparation inner nvm insufficient gas] engine index:0");
    //                     done();
    //                 }).catch(function(err){
    //                     done(err);
    //                 });
    //             } catch(err) {
    //                 console.log("check tx err :" + err);
    //                 done(err);
    //                 return;
    //             }
    //         });
    //     }).catch(function(err) {
    //         console.log("unexpected err: " + err);
    //         done(err);
    //     });
    // });

    // it ('5. nas is not enough and but catch the error', function (done) {
    //     nonce = nonce + 1;
    //     console.log(callerContractAddress);
    //     var contract = {
    //         "function": "safeSave",
    //         "args": "[\"" + calleeContractAddress + "\",\"msg1\", \"湖人总冠军\"]"
    //     };
    //     var tx = new Transaction(ChainID, sourceAccount, callerContractAddress, Unit.nasToBasic(1), nonce, 1000000, 203320, contract);
    //     // tx.to = contractAddress;
    //     tx.signTransaction();
    //     // console.log("silent_debug");
    //     neb.api.sendRawTransaction(tx.toProtoString()).then(function(resp) {
    //         console.log("----step1. call callerTx ", resp);
    //         checkTransaction(resp.txhash, function(resp) {
    //             try {
    //                 expect(resp).to.not.be.a('undefined');
    //                 expect(resp.status).to.be.equal(0);
    //                 console.log("----step2. have been on chain");
    //                 neb.api.getAccountState(callerContractAddress).then(function(state){
    //                     expect(state.balance).to.be.equal("0");
    //                     return neb.api.getAccountState(calleeContractAddress);
    //                 }).then(function(state){
    //                     expect(state.balance).to.be.equal(calleebalance.toString());
    //                     return neb.api.getEventsByHash(resp.hash);
    //                 }).then(function(result){
    //                     expect(JSON.parse(result.events[0].data).error).equal("Mult Call: inner transation err [inner transfer failed] engine index:0");
    //                     done();
    //                 }).catch(function(err){
    //                     done(err);
    //                 });
    //             } catch(err) {
    //                 console.log("check tx err :" + err);
    //                 done(err);
    //                 return;
    //             }
    //         });
    //     }).catch(function(err) {
    //         console.log("unexpected err: " + err);
    //         done(err);
    //     });
    // });

    // it ('6. trigger the err in callee contract and but catch the error', function (done) {
    //     nonce = nonce + 1;
    //     console.log(callerContractAddress);
    //     var contract = {
    //         "function": "testTryCatch",
    //         "args": "[\"" + calleeContractAddress + "\"]"

    //     };
    //     var tx = new Transaction(ChainID, sourceAccount, callerContractAddress, Unit.nasToBasic(1), nonce, 1000000, 2000000, contract);
    //     // tx.to = contractAddress;
    //     tx.signTransaction();
    //     // console.log("silent_debug");
    //     neb.api.sendRawTransaction(tx.toProtoString()).then(function(resp) {
    //         console.log("----step1. call callerTx ", resp);
    //         checkTransaction(resp.txhash, function(resp) {
    //             try {
    //                 expect(resp).to.not.be.a('undefined');
    //                 expect(resp.status).to.be.equal(0);
    //                 console.log("----step2. have been on chain");
    //                 neb.api.getAccountState(callerContractAddress).then(function(state){
    //                     expect(state.balance).to.be.equal("0");
    //                     return neb.api.getAccountState(calleeContractAddress);
    //                 }).then(function(state){
    //                     expect(state.balance).to.be.equal(calleebalance.toString());
    //                     return neb.api.getEventsByHash(resp.hash);
    //                 }).then(function(result){
    //                     console.log("hello");
    //                     //expect(JSON.parse(result.events[0].data).error).equal("multi execution failed");
    //                     //{"result":{"events":[{"topic":"chain.innerTransferContract","data":"{\"from\":\"n21G2RuMsnKAhQL3B6iLQ5oejpmt6oHDEsT\",\"to\":\"n1ioFjZJqfMW8Aa8LjvGjJjMpntQdg4z9HB\",\"value\":\"0\",\"err\":\"execution failed\"}"},{"topic":"chain.transactionResult","data":"{\"hash\":\"da3b29b31d4270ad108748b86f5122616830c0a388fc0280515e81e521e7fdcc\",\"status\":0,\"gas_used\":\"20276\",\"error\":\"multi execution failed\"}"}]}}
    //                     done();
    //                 }).catch(function(err){
    //                     done(err);
    //                 });
    //             } catch(err) {
    //                 console.log("check tx err :" + err);
    //                 done(err);
    //                 return;
    //             }
    //         });
    //     }).catch(function(err) {
    //         console.log("unexpected err: " + err);
    //         done(err);
    //     });
    // });

        it ('7. test wrong interface not func', function (done) {
        nonce = nonce + 1;
        console.log(callerContractAddress);
        var contract = {
            "function": "testWrongInterfaceNotFunc",
            "args": "[\"" + calleeContractAddress + "\"]"
        };
        var tx = new Transaction(ChainID, sourceAccount, callerContractAddress, Unit.nasToBasic(2), nonce, 1000000, 2000000, contract);
        // tx.to = contractAddress;
        tx.signTransaction();
        // console.log("silent_debug");
        neb.api.sendRawTransaction(tx.toProtoString()).then(function(resp) {
            console.log("----step1. call callerTx ", resp);
            checkTransaction(resp.txhash, function(resp) {
                try {
                    expect(resp).to.not.be.a('undefined');
                    expect(resp.status).to.be.equal(0);
                    console.log("----step2. have been on chain， To check balances");
                    neb.api.getAccountState(callerContractAddress).then(function(state){
                        expect(state.balance).to.be.equal("0");
                        return neb.api.getAccountState(calleeContractAddress);
                    }).then(function(state){
                        expect(state.balance).to.be.equal(calleebalance.toString());
                        console.log("----step3, to check the result");
                        var contract = {
                            "function": "get",
                            "args": "[\"" + calleeContractAddress + "\",\"msg1\"]"
                        };
                        return neb.api.getEventsByHash(resp.hash);
                    }).then(function(result) {
                        expect(JSON.parse(result.events[0].data).error).equal("Call: wrong interface define");
                        done();
                    }).catch(function(err) {
                        done(err);
                    });
                } catch(err) {
                    console.log("check tx err :" + err);
                    done(err);
                    return;
                }
            });
        }).catch(function(err) {
            console.log("unexpected err: " + err);
            done(err);
        });
    });
    

    it ('8. test wrong inteface miss func', function (done) {
        nonce = nonce + 1;
        console.log(callerContractAddress);
        var contract = {
            "function": "testWrongInterfaceMissFunc",
            "args":  "[\"" + calleeContractAddress + "\"]"
        };
        var tx = new Transaction(ChainID, sourceAccount, callerContractAddress, Unit.nasToBasic(2), nonce, 1000000, 2000000, contract);
        // tx.to = contractAddress;
        tx.signTransaction();
        // console.log("silent_debug");
        neb.api.sendRawTransaction(tx.toProtoString()).then(function(resp) {
            console.log("----step1. call callerTx ", resp);
            checkTransaction(resp.txhash, function(resp) {
                try {
                    expect(resp).to.not.be.a('undefined');
                    expect(resp.status).to.be.equal(0);
                    console.log("----step2. have been on chain， To check balances");
                    neb.api.getAccountState(callerContractAddress).then(function(state){
                        expect(state.balance).to.be.equal("0");
                        return neb.api.getAccountState(calleeContractAddress);
                    }).then(function(state){
                        expect(state.balance).to.be.equal(calleebalance.toString());
                        console.log("----step3, to check the result");
                        var contract = {
                            "function": "get",
                            "args": "[\"" + calleeContractAddress + "\",\"msg1\"]"
                        };
                        return neb.api.getEventsByHash(resp.hash);
                    }).then(function(result) {
                        expect(JSON.parse(result.events[0].data).error).equal("Call: contract have no function called : testTryCatch");
                        done();
                    }).catch(function(err) {
                        done(err);
                    });
                } catch(err) {
                    console.log("check tx err :" + err);
                    done(err);
                    return;
                }
            });
        }).catch(function(err) {
            console.log("unexpected err: " + err);
            done(err);
        });
    });
});
