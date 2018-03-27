'use strict';

var sleep = require("system-sleep");
var HttpRequest = require("../../node-request");
var FS = require("fs");

var expect = require('chai').expect;
var BigNumber = require('bignumber.js');

var Nebulas = require('nebulas');
var Account = Nebulas.Account;
var Transaction = Nebulas.Transaction;
var CryptoUtils = Nebulas.CryptoUtils;
var Utils = Nebulas.Utils;
var Neb = Nebulas.Neb;
var neb = new Neb();

var ChainID;
var originSource, source, deploy, from, fromState, contractAddr;

var coinbase, coinState;
var testCases = new Array();
var caseIndex = 0;

// mocha cases/contract/xxx testneb1 -t 200000
var args = process.argv.splice(2);
var env = args[1];
if (env !== "local" && env !== "testneb1" && env !== "testneb2" && env !== "testneb3") {
    env = "local";
}
console.log("env:", env);

if (env == 'local'){
    neb.setRequest(new HttpRequest("http://127.0.0.1:8685"));//https://testnet.nebulas.io
    ChainID = 100;
    originSource = new Account("d80f115bdbba5ef215707a8d7053c16f4e65588fd50b0f83369ad142b99891b5");
    coinbase = "n1QZMXSZtW7BUerroSms4axNfyBGyFGkrh5";

}else if(env == 'testneb1'){
    neb.setRequest(new HttpRequest("http://35.182.48.19:8685"));
    ChainID = 1001;
    originSource = new Account("25a3a441a34658e7a595a0eda222fa43ac51bd223017d17b420674fb6d0a4d52");
    coinbase = "n1SAeQRVn33bamxN4ehWUT7JGdxipwn8b17";
}else if(env == "testneb2"){
    neb.setRequest(new HttpRequest("http://34.205.26.12:8685"));
    ChainID = 1002;
    originSource = new Account("25a3a441a34658e7a595a0eda222fa43ac51bd223017d17b420674fb6d0a4d52");
    coinbase = "n1SAeQRVn33bamxN4ehWUT7JGdxipwn8b17";
}else if(env == "testneb3"){
    neb.setRequest(new HttpRequest("http://35.177.214.138:8685"));
    ChainID = 1003;
    originSource = new Account("25a3a441a34658e7a595a0eda222fa43ac51bd223017d17b420674fb6d0a4d52");
    coinbase = "n1SAeQRVn33bamxN4ehWUT7JGdxipwn8b17";
}else{
    console.log("please input correct env local testneb1 testneb2 testneb3");
    return;
}

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

function testTransfer(testInput, testExpect, done) {
    var from = (Utils.isNull(testInput.from)) ? deploy : testInput.from;
    var to = Account.NewAccount();
    var fromBalance, toBalance;

    balanceOfNRC20(from.getAddressString()).then(function(resp) {
        fromBalance = JSON.parse(resp.result);
        console.log("from balance:", fromBalance);

        return balanceOfNRC20(to.getAddressString());
    }).then(function (resp) {
        toBalance = JSON.parse(resp.result);
        console.log("to balance:", toBalance);

        return neb.api.getAccountState(from.getAddressString());
    }).then(function (resp) {
        console.log("from state:", JSON.stringify(resp));

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
        var tx = new Transaction(ChainID, from, contractAddr, "0", parseInt(resp.nonce) + 1, "1000000", "2000000", contract);
        tx.signTransaction();

        console.log("raw tx:", tx.toString());
        return neb.api.sendRawTransaction(tx.toProtoString());
    }).then(function (resp) {
        console.log("send raw tx:", resp);
        checkTransaction(resp.txhash, function (receipt) {
            var resetContract = false;
            try {
                expect(receipt).to.be.have.property('status').equal(testExpect.status);

                balanceOfNRC20(from.getAddressString()).then(function (resp) {
                    var balance = JSON.parse(resp.result);
                    console.log("after from balance:", balance);

                    if (testExpect.status === 1) {
                        var balanceNumber = new BigNumber(fromBalance).sub(testInput.transferValue);
                        expect(balanceNumber.toString(10)).to.equal(balance);
                    } else {
                        expect(balance).to.equal(fromBalance);
                    }

                    if (balance === "0") {
                        resetContract = true;
                    }

                    return balanceOfNRC20(to.getAddressString());
                }).then(function (resp) {
                    var balance = JSON.parse(resp.result);
                    console.log("after to balance:", balance);

                    if (testExpect.status === 1) {
                        var balanceNumber = new BigNumber(toBalance).plus(testInput.transferValue);
                        expect(balanceNumber.toString(10)).to.equal(balance);
                    } else {
                        expect(toBalance).to.equal(balance);
                    }

                    return neb.api.getEventsByHash(receipt.hash);
                }).then(function (events) {
                    // console.log("tx events:", events);
                    for (var i = 0; i < events.events.length; i++) {
                        var event = events.events[i];
                        console.log("tx event:", event);
                        if (event.topic == "chain.transactionResult") {
                            var result = JSON.parse(event.data);
                            expect(result.status).to.equal(testExpect.status);
                        }
                    }
                    expect(events.events.length).to.equal(5);
                    if (resetContract) {
                        contractAddr = null;
                    }
                    done();
                }).catch(function (err) {
                    if (resetContract) {
                        contractAddr = null;
                    }
                    done(err);
                })
            } catch (err) {
                if (resetContract) {
                    contractAddr = null;
                }
                done(err);
            }
        });
    }).catch(function(err) {
        done(err);
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

var testCase = {
    "name": "1. transfer mult event",
    "testInput": {
        isTransfer: true,
        transferValue: "1",
        function: "transferforMultEvent",
        args: ""
    },
    "testExpect": {
        status: 1
    }
};
testCases.push(testCase);
/*testCase = {
    "name": "1. transfer mult event",
    "testInput": {
        isTransfer: true,
        function: "transferforMultEvent",
        args: ""
    },
    "testExpect": {
        status: 0
    }
};
testCases.push(testCase);*/

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
                    testTransfer(testCase.testInput, testCase.testExpect, done);
                }
            });
        });
    }

    afterEach(function () {
        caseIndex++;
    });
});
