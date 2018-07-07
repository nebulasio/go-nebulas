"use strict";

var BigNumber = require('bignumber.js');
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
env = 'local';
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


function doTest(testInput, testExpect, done) {
    try {
        nonce = nonce + 1;
        var gasLimit = 2000000;
        if (testInput.gasLimit) {
            gasLimit = testInput.gasLimit;
        }
        var tx = new Transaction(ChainID, sourceAccount, callerContractAddress, Unit.nasToBasic(testInput.value), nonce, 1000000, gasLimit, testInput.contract);
        // tx.to = contractAddress;
        tx.signTransaction();
        // console.log("silent_debug");
        var coinState;
        neb.api.getAccountState(coinbase).then(function(state){
            coinState = state;
            return neb.api.sendRawTransaction(tx.toProtoString());
        }).then(function(resp) {
            console.log("----step1. call callerTx ", resp);
            checkTransaction(resp.txhash, function(resp) {
                try {
                    expect(resp).to.not.be.a('undefined');
                    console.log("----step2. have been on chain， To check balances");
                    console.log("================result", JSON.stringify(resp));
                    expect(resp.status).to.be.equal(testExpect.txStatus);
                    neb.api.getAccountState(callerContractAddress).then(function(state){
                        expect(state.balance).to.be.equal(testExpect.callerBalance);
                        return neb.api.getAccountState(calleeContractAddress);
                    }).then(function(state) {
                        expect(state.balance).to.be.equal(testExpect.calleeBalance);  
                        console.log("----step3, to check the result");
                        return neb.api.call(sourceAccount.getAddressString(), callerContractAddress, 
                            Unit.nasToBasic(0), nonce, 1000000, 2000000, testInput.resultCheckContract);
                    }).then(function(result){
                        console.log(JSON.stringify(result));
                        //result = {"result":"{\"key\":\"msg1\",\"value\":\"湖人总冠军\"}","execute_err":"","estimate_gas":"20511"}
                        if (testExpect.txStatus == 1) {
                            expect(result.result).equal(testExpect.result);
                        }
                        console.log("----step4, to check the err info by get event");
                        return neb.api.getEventsByHash(resp.hash);
                    }).then(function(result){
                        console.log(JSON.stringify(result));
                        if (testExpect.txStatus == 0) {
                            expect(JSON.parse(result.events[0].data).error).equal(testExpect.errInfo);
                        }
                        return neb.api.getAccountState(coinbase);
                    }).then(function (state) {
                        console.log("get coinbase account state after tx:" + JSON.stringify(state));
                        var reward = new BigNumber(state.balance).sub(coinState.balance);
                        reward = reward.mod(new BigNumber(1.42694).mul(new BigNumber(10).pow(18)));
                        // The transaction should be only
                        console.log("=========================",reward.toString());
                        done();
                        expect(reward.toString()).to.equal(testExpect.reward);
                    }).catch(function(err) {
                        console.log(err);
                        console.log("unexpected err in level 2" );
                        done(err);
                    });
                    
                } catch(err) {
                    console.log("check tx err :" + err);
                    done(err);
                    return;
                }
            });
        }).catch(function(err) {
            console.log("unexpected err in level 1" );
            done(err);
        });
    } catch(err) {
        console.log("unexpected err in level 0");
        done(err);
    }
}



var calleeContractSrc = FS.readFileSync("nf/nvm/test/kvStore.js", "utf-8");
var callerContractSrc = FS.readFileSync("nf/nvm/test/kvStoreProxy.js", "utf-8"); 
var calleeContractAddress;
var callerContractAddress;
var notExistAddress = "n1i8P4uhhgmHQmagmFRsk9cRzfJSkfnv2cp";
var calleeBalance = 0;
var callerBalance = 0;

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

    it ('1# test normal call', function(done) {
        var testInput = {
            contract: {
                "function": "save",
                "args": "[\"" + calleeContractAddress + "\",\"msg1\", \"湖人总冠军\"]"
            },
            resultCheckContract: {
                "function": "get",
                "args": "[\"" + calleeContractAddress + "\",\"msg1\"]"
            },
            value: 2
        };

        calleeBalance += 2000000000000000000;
        
        var testExpect = {
            txStatus: 1,
            reward: "57549000000",
            callerBalance: callerBalance.toString(),
            calleeBalance: calleeBalance.toString(),
            result: "{\"key\":\"msg1\",\"value\":\"湖人总冠军\"}",
        }

        doTest(testInput, testExpect, done);
    });
    it ('2# test normal call by use "call"', function(done) {
        var testInput = {
            contract: {
                "function": "saveByCall",
                "args": "[\"" + calleeContractAddress + "\",\"msg2\", \"湖人总冠军\"]"
            },
            resultCheckContract: {
                "function": "get",
                "args": "[\"" + calleeContractAddress + "\",\"msg2\"]"
            },
            value: 2
        };
        calleeBalance += 2000000000000000000;
        var testExpect = {
            txStatus: 1,
            reward: "57555000000",
            callerBalance: callerBalance.toString(),
            calleeBalance: calleeBalance.toString(),
            result: "{\"key\":\"msg2\",\"value\":\"湖人总冠军\"}",
        }

        doTest(testInput, testExpect, done);
    });
    
    it ('3# not exsit callee contract', function(done) {
        var testInput = {
            contract: {
                "function": "save",
                "args": "[\"" + notExistAddress + "\",\"msg1.5\", \"湖人总冠军\"]"
            },
            value: 2
        };
        
        var testExpect = {
            txStatus: 0,
            reward: "25188000000",
            callerBalance: callerBalance.toString(),
            calleeBalance: calleeBalance.toString(),
            errInfo: "Call: Inner Call: no contract at this address",
        }
        
        doTest(testInput, testExpect, done);
    });
    
    it ('4# function not exsit in callee contract', function(done) {
        var testInput = {
            contract: {
                "function": "testFuncNotExist",
                "args": "[\"" + calleeContractAddress + "\",\"msg1.5\", \"湖人总冠军\"]"
            },
            value: 2
        };
        
        var testExpect = {
            txStatus: 0,
            reward: "25212000000",
            callerBalance: callerBalance.toString(),
            calleeBalance: calleeBalance.toString(),
            errInfo: "execution failed",
        }

        doTest(testInput, testExpect, done);
    });

    it ('5# test usage of value', function(done) {
        var testInput = {
            contract: {
                "function": "testUsageOfValue",
                "args": "[\"" + calleeContractAddress + "\",\"msg3\", \"湖人总冠军\"]"
            },
            resultCheckContract: {
                "function": "get",
                "args": "[\"" + calleeContractAddress + "\",\"msg3\"]"
            },
            value: 0
        };
        
        var testExpect = {
            txStatus: 1,
            reward: "57538000000",
            callerBalance: callerBalance.toString(),
            calleeBalance: calleeBalance.toString(),
            result: "{\"key\":\"msg3\",\"value\":\"湖人总冠军\"}",
        }

        doTest(testInput, testExpect, done);
    });

    it ('6# caller contract has not enough balance', function(done) {
        var testInput = {
            contract: {
                "function": "save",
                "args": "[\"" + calleeContractAddress + "\",\"msg4\", \"湖人总冠军\"]"
            },
            value: 1
        };
        
        var testExpect = {
            txStatus: 0,
            reward: "25210000000",
            callerBalance: callerBalance.toString(),
            calleeBalance: calleeBalance.toString(),
            errInfo: "Call: inner transfer failed",
        }

        doTest(testInput, testExpect, done);
    });

    it ('7# gasLimit is not enough', function(done) {
        var testInput = {
            contract: {
                "function": "save",
                "args": "[\"" + calleeContractAddress + "\",\"msg4\", \"湖人总冠军\"]"
            },
            value: 2,
            gasLimit: 32300
        };
        
        var testExpect = {
            txStatus: 0,
            reward: "25210000000", //TODO: to check
            callerBalance: callerBalance.toString(),
            calleeBalance: calleeBalance.toString(),
            errInfo: "insufficient gas",
        }

        doTest(testInput, testExpect, done);
    });

    it ('8# nas is not enough and but catch the error', function(done) {
        var testInput = {
            contract: {
                "function": "safeSave",
                "args": "[\"" + calleeContractAddress + "\",\"msg4\", \"湖人总冠军\"]"
            },
            value: 1,
        };
        
        var testExpect = {
            txStatus: 0,
            reward: "25214000000",
            callerBalance: callerBalance.toString(),
            calleeBalance: calleeBalance.toString(),
            errInfo: "Call: inner transfer failed",
        }

        doTest(testInput, testExpect, done);
    });

    it ('9# trigger the err in callee contract and but catch the error', function(done) {
        var testInput = {
            contract: {
                "function": "testTryCatch",
                "args": "[\"" + calleeContractAddress + "\",\"msg4\", \"湖人总冠军\"]"
            },
            value: 2,
        };
        
        var testExpect = {
            txStatus: 0,
            reward: "25206000000",
            callerBalance: callerBalance.toString(),
            calleeBalance: calleeBalance.toString(),
            errInfo: "execution failed", //TODO: ["execution failed", ...]
        }

        doTest(testInput, testExpect, done);
    });

    it ('10# callee contract timeout ', function(done) {
        var testInput = {
            contract: {
                "function": "testTimeOut",
                "args": "[\"" + calleeContractAddress + "\",\"msg4\", \"湖人总冠军\"]"
            },
            value: 2,
        };
        
        var testExpect = {
            txStatus: 0,
            reward: "2000000000000",
            callerBalance: callerBalance.toString(),
            calleeBalance: calleeBalance.toString(),
            errInfo: "insufficient gas", 
        }

        doTest(testInput, testExpect, done);
    });
});
