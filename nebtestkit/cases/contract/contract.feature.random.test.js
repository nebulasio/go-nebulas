'use strict';

var sleep = require('system-sleep');
var FS = require('fs');
var expect = require('chai').expect;
var BigNumber = require('bignumber.js');
var HttpRequest = require('../../node-request');
var TestnetConfig = require('../testnet_config');
var Wallet = require('nebulas');

var Account = Wallet.Account;
var Transaction = Wallet.Transaction;
var env = process.argv.splice(2)[1];
var testnetConfig = new TestnetConfig(env);
var originSource = testnetConfig.sourceAccount;
var ChainID = testnetConfig.ChainId;
var coinbase = testnetConfig.coinbase;
var neb = new Wallet.Neb();
neb.setRequest(new HttpRequest(testnetConfig.apiEndPoint));


var deploy, from, contractAddr, source, fromState, coinState;
var caseIndex = 0, lastnonce = 0;

function prepareSource(done) {
    console.log("originSource address: " + originSource.getAddressString());
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
        checkTransaction(resp.txhash, 0, function (receipt) {
            console.log("tx receipt : " + JSON.stringify(receipt));
            expect(receipt).to.be.have.property('status').equal(1);

            done();
        });
    }).catch(function (err) {
        done(err);
    });
}

function cliamTokens(accounts, values, done) {
    for (var i = 0; i < accounts.length; i++) {
        console.log("acc:"+accounts[i].getAddressString()+" value:"+values[i]);
        sendTransaction(source, accounts[i], values[i], ++lastnonce);
        sleep(30);
    }
    checkCliamTokens(done);
}


function sendTransaction(from, address, value, nonce) {
    var transaction = new Transaction(ChainID, from, address, value, nonce, "1000000", "200000");
    transaction.signTransaction();
    var rawTx = transaction.toProtoString();
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

function checkTransaction(txhash, retry, done){

    var maxRetry = 45;

    // contract status and get contract_address
    var interval = setTimeout(function () {
        neb.api.getTransactionReceipt(txhash).then(function (resp) {
            retry++;

            console.log("check transaction status:" + resp.status);
            if(resp.status && resp.status === 1) {
                // clearInterval(interval);

                if (resp.contract_address) {
                    console.log("deploy private key:" + deploy.getPrivateKeyString());
                    console.log("deploy address:" + deploy.getAddressString());
                    console.log("deploy contract address:" + resp.contract_address);

                    contractAddr = resp.contract_address;
                }

                done(resp);
            } else if (resp.status && resp.status === 2) {
                if (retry > maxRetry) {
                    console.log("check transaction time out");
                    // clearInterval(interval);
                    done(resp);
                } else {
                    checkTransaction(txhash, retry++, done);
                }
            } else {
                // clearInterval(interval);
                console.log("transaction execution failed");
                done(resp);
            }
        }).catch(function (err) {
            retry++;
            console.log("check transaction not found retry " + retry);
            if (retry > maxRetry) {
                console.log(JSON.stringify(err.error));
                // clearInterval(interval);
                done(err);
            } else {
                checkTransaction(txhash, retry++, done);
            }
        });

    }, 2000);
}


function deployContract(done, caseGroup) {
    console.log("start deploying contract: " + caseGroup.groupname);

    neb.api.getAccountState(source.getAddressString()).then(function (resp) {
        console.log("source account state:" + JSON.stringify(resp));

        var accounts = new Array();
        var values = new Array();
        deploy = Account.NewAccount();
        accounts.push(deploy);
        values.push(neb.nasToBasic(1));

        from = Account.NewAccount();
        accounts.push(from);
        var fromBalance = (typeof caseGroup.fromBalance === "undefined") ? neb.nasToBasic(1) : caseGroup.fromBalance;
        values.push(fromBalance);

        cliamTokens(accounts, values, () => {
            try {
                var source = FS.readFileSync("../../../nf/nvm/test/" + caseGroup.filename, "utf-8");
                var contract = {
                    "source": source,
                    "sourceType": caseGroup.type,
                    "args": ""
                };
            
                var tx = new Transaction(testnetConfig.ChainId, deploy, deploy, "0", 1, "10000000", "2000000", contract);
                tx.signTransaction();
                var rawTx = tx.toProtoString();
            
                // console.log("contract:" + rawTx);
                neb.api.sendRawTransaction(rawTx).then(function (resp) {
                    console.log("deploy contract " + caseGroup.groupname + " return: " + JSON.stringify(resp));
            
                    checkTransaction(resp.txhash, 0, (ret) => {
                        if (ret.status && ret.status === 1) {
                            done();
                        } else {
                            done(ret);
                        }
                    });
                });
            } catch (err) {
                done(err);
            };
        });

    }).catch (err => done(err));
}

function runTest(testInput, testExpect, done) {
    var fromAcc = (typeof testInput.from === "undefined") ? from : testInput.from;
    var to = (typeof testInput.to === "undefined") ? Account.fromAddress(contractAddr) : testInput.to;

    var fromBalanceBefore, toBalanceBefore;

    neb.api.getAccountState(to.getAddressString()).then(function (resp) {
        console.log("contractAddr state before: " + JSON.stringify(resp));
        toBalanceBefore = resp.balance;
        return neb.api.getAccountState(from.getAddressString());
    }).then(resp => {
        fromState = resp;
        fromBalanceBefore = resp.balanece;
        console.log("from state before: ", JSON.stringify(resp));
        return neb.api.getAccountState(coinbase);
    }).then(function (resp) {
        console.log("coin state before: ", JSON.stringify(resp));
        coinState = resp;

        var tx = new Transaction(ChainID, fromAcc, to, testInput.value, parseInt(fromState.nonce) + testInput.nonce, testInput.gasPrice, testInput.gasLimit, testInput.contract);
        tx.from.address = fromAcc.address;
        tx.to.address = to.address;
        tx.gasPrice = new BigNumber(testInput.gasPrice);
        tx.gasLimit = new BigNumber(testInput.gasLimit);
        tx.signTransaction();
        console.log("binary tx raw before send: ", tx.toString());
        return neb.api.sendRawTransaction(tx.toProtoString());
    }).then(function (rawResp) {
        console.log("send Raw Tx return:" + JSON.stringify(rawResp));
        expect(rawResp).to.be.have.property('txhash');

        checkTransaction(rawResp.txhash, 0, function (receipt) {
            console.log("tx receipt : " + JSON.stringify(receipt));
            try {
                expect(receipt).to.not.be.a('undefined');
                if (true === testExpect.canExcuteTx) {
                    expect(receipt).to.be.have.property('status').equal(1);
                } else {
                    expect(receipt).to.be.have.property('status').equal(0);
                }

                neb.api.getAccountState(receipt.from).then(function (state) {

                    console.log("from state after: " + JSON.stringify(state));
                    // expect(state.balance).to.equal(testExpect.fromBalanceAfterTx);
                    return neb.api.getAccountState(contractAddr);
                }).then(function (state) {

                    console.log("contractAddr state after: " + JSON.stringify(state));
                    var change = new BigNumber(state.balance).minus(new BigNumber(toBalanceBefore));
                    // expect(change.toString()).to.equal(testExpect.toBalanceChange);
                    return neb.api.getAccountState(coinbase);
                }).then(function (state) {

                    console.log("get coinbase account state before tx:" + JSON.stringify(coinState));
                    console.log("get coinbase account state after tx:" + JSON.stringify(state));
                    var reward = new BigNumber(state.balance).sub(coinState.balance);
                    reward = reward.mod(new BigNumber(1.42694).mul(new BigNumber(10).pow(18)));
                    // The transaction should be only
                    // expect(reward.toString()).to.equal(testExpect.transferReward);
                    console.log("coinbase reward: " + reward.toString());
                    if (receipt.gasUsed) {
                        var txCost = new BigNumber(receipt.gasUsed).mul(receipt.gasPrice).toString(10);
                        // expect(txCost).to.equal(testExpect.transferReward);
                        console.log("tx cost gas: " + txCost.toString());
                    }

                    return neb.api.getEventsByHash(receipt.hash);
                }).then(function (events) {
                    for (var i = 0; i < events.events.length; i++) {
                        var event = events.events[i];
                        //console.log("tx event:", JSON.stringify(event,null,'\t'));
                        console.log("tx event:", event.data);
                        if (event.topic === "chain.transactionResult") {
                            var result = JSON.parse(event.data);
                            expect(result.status).to.equal(testExpect.status);

                            if (testExpect.hasOwnProperty("eventErr")){
                                console.log("Event error checked.");
                                expect(result.error).to.equal(testExpect.eventErr);
                            }
                        }
                        if (event.topic === "chain.contract.random") {
                            var result = JSON.parse(event.data);
                            expect(result.defaultSeedRandom1 == result.userSeedRandom).to.equal(testExpect.equalr1r2);
                            console.log("check equalr1r2 success");
                        }
                    }
                    done();
                }).catch(function (err) {
                    console.log("exe tx err:", err);
                    done(err);
                });
            } catch (err) {
                console.log("submit tx err:", err.message);
                done(err);
            }
        });
    }).catch(function (err) {
        if (err.error && err.error.error && testExpect.eventErr) {
            try {
                expect(err.error.error).to.equal(testExpect.eventErr)
                done();
            } catch (err) {
                done(err);
            }
            return;
        }
        done(err);
    });
}

var testCaseGroups = [];
var caseGroup = {
    "filename": "contract_date_and_random.js",
    "type": "js",
    "groupname": "case group 0: Math.random",
    "groupIndex": 0,

    cases: [
        {
            "name": "0-1. test 'undefined user seed'",
            "testInput": {
                value: "0",
                nonce: 1, 
                gasPrice: 1000000,
                gasLimit: 2000000,
                contract: {
                    function: "testRandom",
                    args: ""
                }
            },
            "testExpect": {
                canExcuteTx: false,
                toBalanceChange: "0",
                status: 0,
                eventErr: "Call: Error: input seed must be a string"
            }
        },
        {
            "name": "0-2. test 'empty user seed('')'",
            "testInput": {
                value: "0",
                nonce: 1, 
                gasPrice: 1000000,
                gasLimit: 2000000,
                contract: {
                    function: "testRandom",
                    args: "[\"\"]"
                }
            },
            "testExpect": {
                canExcuteTx: true,
                toBalanceChange: "0",
                status: 1,
                equalr1r2: false
            }
        },
        {
            "name": "0-3. test 'set user seed('abc')'",
            "testInput": {
                value: "0",
                nonce: 1, 
                gasPrice: 1000000,
                gasLimit: 2000000,
                contract: {
                    function: "testRandom",
                    args: "[\"abc\"]"
                }
            },
            "testExpect": {
                canExcuteTx: true,
                toBalanceChange: "0",
                status: 1,
                equalr1r2: false
            }
        }
    ]
};
testCaseGroups.push(caseGroup);

describe('Contract Math.random test', () => {

    before(done => prepareSource(done));

    for (var i = 0; i < testCaseGroups.length; i++) {

        // if (i != 3) {continue;}         // selectively run tests

        let caseGroup = testCaseGroups[i];
        describe(caseGroup.groupname, () => {
            before(done => {
                deployContract(done, caseGroup);
                caseIndex = 0;
            });

            
            for (var j = 0; j < caseGroup.cases.length; j++) {
                let testCase = caseGroup.cases[j];
                it(testCase.name, done => {
                    console.log("===> running case: " + JSON.stringify(testCase));
                    runTest(testCase.testInput, testCase.testExpect, done);
                });
            }

            afterEach(() => {
                caseIndex++;
                console.log("case group: " + caseGroup.groupIndex + ", index: " + caseIndex);
            });
        });
    }
});