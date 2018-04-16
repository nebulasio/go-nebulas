'use strict';
var async = require('async');
var sleep = require("system-sleep");
var HttpRequest = require("../../node-request");
var TestNetConfig = require("../testnet_config.js");

var FS = require("fs");

var expect = require('chai').expect;
var BigNumber = require('bignumber.js');

var Nebulas = require('nebulas');
var Neb = Nebulas.Neb;
var Account = Nebulas.Account;
var Transaction = Nebulas.Transaction;
var CryptoUtils = Nebulas.CryptoUtils;
var Utils = Nebulas.Utils;


var coinbase, coinState;
var testCases = new Array();
var caseIndex = 0;

// mocha cases/contract/contract.nrc20.mult.event.test.js testneb2 -t 200000
var args = process.argv.splice(2);
var env = args[1];
var testNetConfig = new TestNetConfig(env);


var neb = new Neb();
var source, deploy, from, fromState, contractAddr;
var ChainID = testNetConfig.ChainId;
var originSource = testNetConfig.sourceAccount;
var coinbase = testNetConfig.coinbase;
var apiEndPoint = testNetConfig.apiEndPoint;
neb.setRequest(new HttpRequest(apiEndPoint));


var lastnonce = 0;

function prepareSource(done) {
    neb.api.getAccountState(originSource.getAddressString()).then(function (resp) {
        console.log("prepare source account state:" + JSON.stringify(resp));
        var nonce = parseInt(resp.nonce);

        source = Account.NewAccount();

        var tx = new Transaction(ChainID, originSource, source, neb.nasToBasic(1000), nonce + 1, "1000000", "200000");
        tx.signTransaction();

        console.log("cliam source tx:", tx.toString());

        return neb.api.sendRawTransaction(tx.toProtoString());
    }).then(function (resp) {
        console.log("send Raw Tx:" + JSON.stringify(resp));
        expect(resp).to.be.have.property('txhash');
        checkTransaction(resp.txhash, function (receipt) {
            console.log("tx receipt : " + JSON.stringify(receipt));
            expect(receipt).to.be.have.property('status').equal(1);

            done();
        });
    }).catch(function (err) {
        done(err);
    });
}

function prepareContractCall(testCase, done) {
    neb.api.getAccountState(source.getAddressString()).then(function (resp) {
        console.log("source account state:" + JSON.stringify(resp));
        lastnonce = parseInt(resp.nonce);

        var accounts = new Array();
        var values = new Array();
        if (Utils.isNull(contractAddr)) {
            deploy = Account.NewAccount();
            accounts.push(deploy);
            values.push(neb.nasToBasic(1));
        }

        if (typeof testCase.testInput.from !== "undefined") {
            accounts.push(testCase.testInput.from);
            values.push(neb.nasToBasic(1));
        }

        if (typeof testCase.testInput.to !== "undefined") {
            accounts.push(testCase.testInput.to);
            values.push(neb.nasToBasic(1));
        }

        if (accounts.length > 0) {
            cliamTokens(accounts, values, function () {
                if (Utils.isNull(contractAddr)) {
                    deployContract(done);
                } else {
                    done();
                }
            });
        } else {
            done();
        }

    });
}

function cliamTokens(accounts, values, done) {
    for (var i = 0; i < accounts.length; i++) {
        // console.log("acc:"+accounts[i].getAddressString()+"value:"+values[i]);
        sendTransaction(source, accounts[i], values[i], ++lastnonce);
        sleep(30);
    }
    checkCliamTokens(done);
}

function sendTransaction(from, address, value, nonce) {
    var transaction = new Transaction(ChainID, from, address, value, nonce, "1000000", "2000000");
    transaction.signTransaction();
    var rawTx = transaction.toProtoString();
    // console.log("send transaction:", transaction.toString());
    neb.api.sendRawTransaction(rawTx).then(function (resp) {
        console.log("send raw transaction resp:" + JSON.stringify(resp));
    });
}

function checkCliamTokens(done) {
    var intervalAccount = setInterval(function () {
        neb.api.getAccountState(source.getAddressString()).then(function (resp) {
            // console.log("master accountState resp:" + JSON.stringify(resp));
            var nonce = parseInt(resp.nonce);
            console.log("check cliam tokens nonce:", lastnonce);

            if (lastnonce <= nonce){
                console.log("cliam tokens success");
                clearInterval(intervalAccount);

                done();
            }
        });
    }, 2000);
}

function deployContract(done){

    // create contract
    var source = FS.readFileSync("../nf/nvm/test/NRC20MultEvent.js", "utf-8");
    var contract = {
        "source": source,
        "sourceType": "js",
        "args": "[\"StandardToken\", \"NRC\", 18, \"1000000000\"]"
    };

    var transaction = new Transaction(ChainID, deploy, deploy, "0", 1, "10000000", "2000000", contract);
    transaction.signTransaction();
    var rawTx = transaction.toProtoString();

    // console.log("contract:" + rawTx);

    neb.api.sendRawTransaction(rawTx).then(function (resp) {
        console.log("deploy contract:" + JSON.stringify(resp));

        checkTransaction(resp.txhash, done);
    });
}

function checkTransaction(txhash, done){

    var retry = 0;
    var maxRetry = 20;

    // contract status and get contract_address
    var interval = setInterval(function () {
        // console.log("getTransactionReceipt hash:"+txhash);
        neb.api.getTransactionReceipt(txhash).then(function (resp) {
            retry++;

            console.log("check transaction status:" + resp.status);

            if(resp.status && resp.status === 1) {
                clearInterval(interval);

                if (resp.contract_address) {
                    console.log("deploy private key:" + deploy.getPrivateKeyString());
                    console.log("deploy address:" + deploy.getAddressString());
                    console.log("deploy contract address:" + resp.contract_address);
                    // console.log("deploy receipt:" + JSON.stringify(resp));

                    contractAddr = resp.contract_address;

                    // checkNRCBalance(resp.from, resp.contract_address);
                }

                done(resp);
            } else if (resp.status && resp.status === 2) {
                if (retry > maxRetry) {
                    console.log("check transaction time out");
                    clearInterval(interval);
                    done(resp);
                }
            } else {
                clearInterval(interval);
                console.log("transaction execution failed");
                done(resp);
            }
        }).catch(function (err) {
            retry++;
            console.log("check transaction not found retry");
            if (retry > maxRetry) {
                console.log(JSON.stringify(err.error));
                clearInterval(interval);
                done(err);
            }
        });

    }, 2000);
}


function testTransferByAsync(testInput, testExpect, done) {
    var from = (Utils.isNull(testInput.from)) ? deploy : testInput.from;
    var to = Account.NewAccount();
    var fromBalance, toBalance;
    console.log("testTransferByAsync, fromAddress:%s, toAddress:%s", 
        from.getAddressString(), to.getAddressString());
    async.auto({
        getFromBalance: function(callback) {
            var RR = balanceOfNRC20(from.getAddressString());
            RR.then(function(resp) {
                fromBalance = JSON.parse(resp.result);
                console.log("from balance:", fromBalance);
                callback(null, fromBalance);
            }).catch(function(err){
                console.log("getFromBalance err:", err);
                callback(err, null);
            })
        },
        getToBalance: function(callback) {
            var RR = balanceOfNRC20(to.getAddressString());
            RR.then(function(resp) {
                toBalance = JSON.parse(resp.result);
                console.log("to balance:", toBalance);
                callback(null, toBalance);
            }).catch(function(err){
                console.log("getToBalance err:", err);
                callback(err, null);
            })
            
        },
        getAccountState: function(callback) {
            var RR = neb.api.getAccountState(from.getAddressString());
            //callback(null, resp);
            RR.then(function(resp) {
                console.log("state:", resp);
                callback(null, resp);
            }).catch(function(err){
                console.log("getAccountState err:", err);
                callback(err, null);
            })
        },
        executeContract: ['getFromBalance', 'getToBalance', 'getAccountState', function(callback, results){
            console.log("from state:", JSON.stringify(results.getAccountState));

            var args = testInput.args;
            if (!Utils.isNull(testInput.transferValue)) {
                if (testInput.transferValue === "from.balance") {
                    testInput.transferValue = fromBalance;
                }
                args = "[\""+ to.getAddressString() +"\", \""+ testInput.transferValue +"\"]";
            }
            
            var contract = {
                "function": testInput.function,
                "args": args
            };
            var tx = new Transaction(ChainID, from, contractAddr, "0", parseInt(results.getAccountState.nonce) + 1, "1000000", "2000000", contract);
            tx.signTransaction();

            console.log("raw tx:", tx.toString());
            var RR = neb.api.sendRawTransaction(tx.toProtoString());
            RR.then(function(resp) {
                callback(null, resp);
            }).catch(function(err){
                console.log("executeContract err:", err);
                callback(err, null);
            })
        }],
        checkContract: ['executeContract', function(callback, newtx){
            checkTransaction(newtx.executeContract.txhash, function(resp) {
                console.log("checkTransaction:", resp);
                if (resp.status == 0) {
                    //callback("checkTransaction execut contract failed!", null);
                    callback(null, resp);
                } else {
                    
                    expect(resp).to.be.have.property('status').equal(testExpect.status);
                    callback(null, resp);
                }
            });
        }],
        getAfterFromBalance: ['checkContract', function(callback, receipt){
            var RR = balanceOfNRC20(from.getAddressString());
            RR.then(function(resp) {
                var balance = JSON.parse(resp.result);
                console.log("after from balance:", balance);
                if (testExpect.status === 1) {
                    var balanceNumber = new BigNumber(fromBalance).sub(testInput.transferValue);
                    expect(balanceNumber.toString(10)).to.equal(balance);
                } else {
                    expect(balance).to.equal(fromBalance);
                }
                fromBalance = balance;
                callback(null, balance);
            }).catch(function(err){
                console.log("after getFromBalance err:", err);
                callback(err, null);
            })
        }],
        getAfterToBalance: ['checkContract', function(callback, receipt){
            var RR = balanceOfNRC20(to.getAddressString());
            RR.then(function(resp) {
                var balance = JSON.parse(resp.result);
                console.log("after to balance:", balance);
                if (testExpect.status === 1) {
                    var balanceNumber = new BigNumber(toBalance).plus(testInput.transferValue);
                    expect(balanceNumber.toString(10)).to.equal(balance);
                } else {
                    expect(toBalance).to.equal(balance);
                }
                toBalance = balance;
                callback(null, balance);
            }).catch(function(err){
                console.log("after getToBalance err:", err);
                callback(err, null);
            })
        }],
        getEventsByHash: ['checkContract', function(callback, receipt){
            var RR = neb.api.getEventsByHash(receipt.checkContract.hash);
            RR.then(function(events) {
                //console.log("events:", events);
                for (var i = 0; i < events.events.length; i++) {
                    var event = events.events[i];
                    console.log("tx event:", event);
                    if (event.topic == "chain.transactionResult") {
                        var result = JSON.parse(event.data);
                        expect(result.status).to.equal(testExpect.status);
                    } else {
                        var data = JSON.parse(event.data);
                        if (testExpect.events[i] == null) {
                            expect(typeof(data.Transfer)).equal('undefined');
                            continue;
                        }
                        //check event value
                        if (testExpect.events[i].from) {
                            expect(data.Transfer.from).to.equal(from.getAddressString());
                        } else {
                            expect(typeof(data.Transfer.from)).equal('undefined');
                        }
                        if (testExpect.events[i].to) {
                            expect(data.Transfer.to).to.equal(to.getAddressString());
                        } else {
                            expect(typeof(data.Transfer.to)).equal('undefined');
                        }
                        if (testExpect.events[i].value) {
                            expect(data.Transfer.value).to.equal(testInput.transferValue);
                        } else {
                            expect(typeof(data.Transfer.value)).equal('undefined');
                        }
                    }
                    
                }
                expect(events.events.length).to.equal(testExpect.events.length + 1);
                callback(null, events);
            }).catch(function(err) {
                console.log("getEventsByHash err");
                callback(err, null);
            })
            
        }],
     }, function(err, results) {
        if (err) {
            console.log("async.auto hava break:", err);
            if (err == "checkTransaction execut contract failed!") {
                done();
            } else {
                done(err);
            }
        } else {
            console.log("end async.auto");
            done();
        }
     });
}

function balanceOfNRC20(address) {
    var contract = {
        "function": "balanceOf",
        "args": "[\"" + address + "\"]"
    };
    return neb.api.call(address, contractAddr, "0", 1, "1000000", "200000", contract)
}

function allowanceOfNRC20(owner, spender) {
    var contract = {
        "function": "allowance",
        "args": "[\"" + owner + "\", \""+ spender +"\"]"
    };
    return neb.api.call(owner, contractAddr, "0", 1, "1000000", "2000000", contract)
}
var EventExpect = function(from, to, value) {
    this.from = from;
    this.to = to;
    this.value = value;
};

var testCase = {
    "name": "1. transfer mult event",
    "testInput": {
        isTransfer: true,
        transferValue: "1",
        function: "transferforMultEvent",
        args: ""
    },
    "testExpect": {
        status: 1,
        events: [new EventExpect(1, 1, 1), new EventExpect(1,0,1), new EventExpect(0, 1, 1), new EventExpect(0, 0, 1),],
    }
};
    
testCases.push(testCase);

testCase = {
    "name": "2. transfer mult event Status is err",
    "testInput": {
        isTransfer: true,
        function: "transferforMultEventStatus",
        args: ""
    },
    "testExpect": {
        status: 0,
        events: []
    }
};
testCases.push(testCase);

testCase = {
    "name": "3. transfer mult event args has err",
    "testInput": {
        isTransfer: true,
        transferValue: "1",
        function: "transferforMultEventTransfer",
        args: ""
    },
    "testExpect": {
        status: 1,
        events: [null, new EventExpect(1,1,1),],
    }
};
testCases.push(testCase);

describe('contract call test', function () {
    before(function (done) {
        prepareSource(done);
    });

    for (var i = 0; i < testCases.length; i++) {

        it(testCases[i].name, function (done) {
            var testCase = testCases[caseIndex];
            prepareContractCall(testCase, function (err) {
                if (err instanceof Error) {
                    done(err);
                } else {
                    testTransferByAsync(testCase.testInput, testCase.testExpect, done);
                }
            });
        });
    }

    afterEach(function () {
        caseIndex++;
    });
});