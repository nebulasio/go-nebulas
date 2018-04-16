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

//require("\x00");

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
    var source = FS.readFileSync("../nf/nvm/test/contract_require.js", "utf-8");
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

    async.auto({
        executeContract: function(callback){
            var RR = requireTemplate(from.getAddressString(), testInput.function);
            RR.then(function(resp) {
                console.log("resp:" + resp.result);
                console.log("data:" + JSON.stringify(resp));
                //data = JSON.parse(resp);
                //console.log("to balance:", toBalance);
                //expect(resp).to.be.have.property('status').equal(testExpect.status);
                expect(resp).to.be.have.property('execute_err');
                //console.log("--exce:" + testInput.testExpect.result);
                //expect(resp).to.be.have.property('result').equal(testExpect.result);
                callback(null, null);
            }).catch(function(err){
                //console.log("getToBalance err:", err);

                callback(err, null);
            })
        },
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

function requireTemplate(address, arg) {
    var contract = {
        "function": arg,
        "args": "[\"" + address + "\"]"
    };
    return neb.api.call(address, contractAddr, "0", 1, "1000000", "200000", contract)
}

function requireNULL(address) {
    var contract = {
        "function": "requireNULL",
        "args": "[\"" + address + "\"]"
    };
    return neb.api.call(address, contractAddr, "0", 1, "1000000", "200000", contract)
}
function requireNotExistPath(address) {
    var contract = {
        "function": "requireNotExistPath",
        "args": "[\"" + address + "\"]"
    };
    return neb.api.call(address, contractAddr, "0", 1, "1000000", "200000", contract)
}
function requireCurPath(address) {
    var contract = {
        "function": "requireCurPath",
        "args": "[\"" + address + "\"]"
    };
    return neb.api.call(address, contractAddr, "0", 1, "1000000", "200000", contract)
}
function requireNotExistFile(address) {
    var contract = {
        "function": "requireNotExistFile",
        "args": "[\"" + address + "\"]"
    };
    return neb.api.call(address, contractAddr, "0", 1, "1000000", "200000", contract)
}

var testCase = {
    "name": "1. require NULL",
    "testInput": {
        function: "requireNULL",
        args: ""
    },
    "testExpect": {
        status: 0,
        result: "Error: require path is not in lib"
    }
};
    
testCases.push(testCase);

testCase = {
    "name": "2. requireNotExistPath",
    "testInput": {
        function: "requireNotExistPath",
        args: ""
    },
    "testExpect": {
        status: 0,
        result: "Error: require path is not in lib"
    }
};
testCases.push(testCase);

testCase = {
    "name": "3. requireCurPath",
    "testInput": {
        function: "requireCurPath",
        args: ""
    },
    "testExpect": {
        status: 0,
        result: "Error: require path is not in lib"
    }
};
testCases.push(testCase);

testCase = {
    "name": "4. requireNotExistFile",
    "testInput": {
        function: "requireNotExistFile",
        args: ""
    },
    "testExpect": {
        status: 0,
        result: "Error: require path is invalid absolutepath"
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
                    //done();
                    testTransferByAsync(testCase.testInput, testCase.testExpect, done);
                }
            });
        });
    }

    afterEach(function () {
        caseIndex++;
    });
});