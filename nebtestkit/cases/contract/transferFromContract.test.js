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
if (env == null){
    env = "local";
}
var testNetConfig = new TestNetConfig(env);

var neb = new Neb();
var ChainID = testNetConfig.ChainId;
var sourceAccount = testNetConfig.sourceAccount;
var coinbase = testNetConfig.coinbase;
var apiEndPoint = testNetConfig.apiEndPoint;
neb.setRequest(new HttpRequest(apiEndPoint));

var contractAddress;
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

describe('test transfer from contract', function () {
    before('0. deploy contract', function (done) {
        try {
            neb.api.getAccountState(sourceAccount.getAddressString()).then(function(resp) {
                console.log("----step0. get source account state: " + JSON.stringify(resp));
                var contractSource = FS.readFileSync("../nf/nvm/test/transfer_value_from_contract.js", "UTF-8");
                var contract = {
                    'source': contractSource,
                    "sourceType": "js",
                    "arges": ''
                };
                nonce = parseInt(resp.nonce);
                nonce = nonce + 1;
                var tx = new Transaction(ChainID, sourceAccount, sourceAccount, 0, nonce, 1000000, 20000000, contract);
                tx.signTransaction();
                return neb.api.sendRawTransaction(tx.toProtoString());
            }).then(function(resp) {
                console.log("----step1. deploy contract: " + JSON.stringify(resp));
                contractAddress = resp.contract_address;
                checkTransaction(resp.txhash, function(resp) {
                    expect(resp).to.not.be.a('undefined');
                    console.log("----step2. have been on chain");
                    done();
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

    it ('1. send 10 nas to contract address', function (done) {
        nonce = nonce + 1;
        console.log(contractAddress);
        var tx = new Transaction(ChainID, sourceAccount, contractAddress, Unit.nasToBasic(10), nonce, 1000000, 2000000);
        // tx.to = contractAddress;
        tx.signTransaction();
        console.log(tx.toString());
        // console.log("silent_debug");
        neb.api.sendRawTransaction(tx.toProtoString()).then(function(resp) {
            console.log("----step3. send nas to contract address: ", resp);
            checkTransaction(resp.txhash, function(resp) {
                expect(resp).to.not.be.a('undefined');
                console.log("----step4. have been on chain");
                done();
            });
        }).catch(function(err) {
            console.log("unexpected err: " + err);
            done(err);
        });
    });

    it ('2. send 5 nas from contract address to toAddress', function (done) {
        nonce = nonce + 1;
        var contract = {
            "function": "transferSpecialValue",
            "args": "[\"" + toAddress.getAddressString() + "\", \"5000000000000000000\"]"
        };
        var tx = new Transaction(ChainID, sourceAccount, contractAddress, 0, nonce, 1000000, 2000000, contract);
        tx.signTransaction();
        console.log(tx.toString);
        neb.api.sendRawTransaction(tx.toProtoString()).then(function (resp) {
            console.log("----step5. call transferSpecialValue function", JSON.stringify(resp));
            var hash = resp.txhash;
            checkTransaction(resp.txhash, function(resp) {
                try {
                    expect(resp).to.not.be.a('undefined');
                    console.log("----step6. have be on chain");
                    neb.api.getAccountState(toAddress.getAddressString()).then(function(resp) {
                        expect(resp.balance).equal("5000000000000000000");
                        return neb.api.getAccountState(contractAddress);
                    }).then(function(resp){
                        expect(resp.balance).equal("5000000000000000000");
                        return neb.api.getEventsByHash(hash);
                    }).then(function(resp){
                        console.log(JSON.stringify(resp));
                        expect(resp.events[0].topic).equal("chain.transferFromContract");
                        done();
                    }).catch(function(err){
                        console.log("unexpected err :", err);
                        done(err);
                    });
                } catch(err) {
                    console.log("unexpected err :", err);
                    done(err);
                }
            });
        }).catch(function(err){
            console.log("unexpected err :", err);
            done(err);
        });
    });

    it ('3. send 5 nas to contract and try to send 11 nas from contract toAddress',
     function (done) {
        nonce = nonce + 1;
        var contract = {
            "function": "transferSpecialValue",
            "args": "[\"" + toAddress.getAddressString() + "\", \"11000000000000000000\"]"
        };
        var tx = new Transaction(ChainID, sourceAccount, contractAddress, Unit.nasToBasic(5), nonce, 1000000, 2000000, contract);
        tx.signTransaction();
        neb.api.sendRawTransaction(tx.toProtoString()).then(function (resp) {
            console.log("----step7. call transferSpecialValue function", JSON.stringify(resp));
            var hash = resp.txhash;
            checkTransaction(resp.txhash, function(resp) {
                try {
                    expect(resp).to.not.be.a('undefined');
                    expect(resp.status).to.be.equal(0);
                    console.log("----step8. have be on chain");
                    neb.api.getAccountState(toAddress.getAddressString()).then(function(resp) {
                        expect(resp.balance).equal("5000000000000000000");
                        return neb.api.getAccountState(contractAddress);
                    }).then(function(resp){
                        expect(resp.balance).equal("5000000000000000000");
                        return neb.api.getEventsByHash(hash);
                    }).then(function(resp){
                        console.log("======", JSON.stringify(resp))
                        expect(resp.events[0].topic).equal("chain.transferFromContract");
                        expect(JSON.parse(resp.events[0].data).error).equal("failed to sub balace from contract address");
                        done();
                    }).catch(function(err){
                        console.log("unexpected err :", err);
                        done(err);
                    });
                } catch (err) {
                    console.log("unexpected error: ", err);
                    done(err);
                }

            });
        }).catch(function(err){
            console.log("unexpected err :", err);
            done(err);
        });
    });

    it ('4. send 5 nas to contract and try to send 10 nas from contract toAddress',
    function (done) {
       nonce = nonce + 1;
       var contract = {
           "function": "transferSpecialValue",
           "args": "[\"" + toAddress.getAddressString() + "\", \"10000000000000000000\"]"
       };
       var tx = new Transaction(ChainID, sourceAccount, contractAddress, Unit.nasToBasic(5), nonce, 1000000, 2000000, contract);
       tx.signTransaction();
       neb.api.sendRawTransaction(tx.toProtoString()).then(function (resp) {
           console.log("----step9. call transferSpecialValue function", JSON.stringify(resp));
           checkTransaction(resp.txhash, function(resp) {
               try {
                   expect(resp).to.not.be.a('undefined');
                   expect(resp.status).to.be.equal(1);
                   console.log("----step10. have be on chain");
                   neb.api.getAccountState(toAddress.getAddressString()).then(function(resp) {
                       expect(resp.balance).equal("15000000000000000000");
                       return neb.api.getAccountState(contractAddress);
                   }).then(function(resp){
                       expect(resp.balance).equal("0");
                       done();
                   }).catch(function(err){
                       console.log("unexpected err :", err);
                       done(err);
                   });
               } catch (err) {
                   console.log("unexpected error: ", err);
                   done(err);
               }

           });
       }).catch(function(err){
           console.log("unexpected err :", err);
           done(err);
       });
   });
});
