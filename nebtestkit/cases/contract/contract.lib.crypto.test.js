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
                        console.log("tx event data:", event.data);
                        if (event.topic === "chain.transactionResult") {
                            var result = JSON.parse(event.data);
                            expect(result.status).to.equal(testExpect.status);

                            if (testExpect.hasOwnProperty("eventErr")){
                                console.log("Event error checked.");
                                expect(result.error).to.equal(testExpect.eventErr);
                            }

                            if (testExpect.hasOwnProperty("result")){
                                console.log("Result checked.");
                                expect(result.execute_result).to.equal(testExpect.result);
                            }
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
    "filename": "contract_crypto.js",
    "type": "js",
    "groupname": "case group 0",
    "groupIndex": 0,

    cases: [
        {
            "name": "0-1. test sha256",
            "testInput": {
                value: "0",
                nonce: 1, 
                gasPrice: 1000000,
                gasLimit: 2000000,
                contract: {
                    function: "testSha256",
                    args: "[\"Nebulas is a next generation public blockchain, aiming for a continuously improving ecosystem.\"]"
                }
            },
            "testExpect": {
                canExcuteTx: true,
                toBalanceChange: "0",
                status: 1,
                // eventErr: "Call: Error: input seed must be a string",
                result: "\"a32d6d686968192663b9c9e21e6a3ba1ba9b2e288470c2f98b790256530933e0\""
            }
        },
        {
            "name": "0-2. test sha3256",
            "testInput": {
                value: "0",
                nonce: 1, 
                gasPrice: 1000000,
                gasLimit: 2000000,
                contract: {
                    function: "testSha3256",
                    args: "[\"Nebulas is a next generation public blockchain, aiming for a continuously improving ecosystem.\"]"
                }
            },
            "testExpect": {
                canExcuteTx: true,
                toBalanceChange: "0",
                status: 1,
                result: "\"564733f9f3e139b925cfb1e7e50ba8581e9107b13e4213f2e4708d9c284be75b\""
            }
        },
        {
            "name": "0-3. test ripemd160",
            "testInput": {
                value: "0",
                nonce: 1, 
                gasPrice: 1000000,
                gasLimit: 2000000,
                contract: {
                    function: "testRipemd160",
                    args: "[\"Nebulas is a next generation public blockchain, aiming for a continuously improving ecosystem.\"]"
                }
            },
            "testExpect": {
                canExcuteTx: true,
                toBalanceChange: "0",
                status: 1,
                result: "\"4236aa9974eb7b9ddb0f7a7ed06d4bf3d9c0e386\""
            }
        },
        {
            "name": "0-4. test recoverAddress",
            "testInput": {
                value: "0",
                nonce: 1, 
                gasPrice: 1000000,
                gasLimit: 2000000,
                contract: {
                    function: "testRecoverAddress",
                    args: "[1,\"564733f9f3e139b925cfb1e7e50ba8581e9107b13e4213f2e4708d9c284be75b\",\"d80e282d165f8c05d8581133df7af3c7c41d51ec7cd8470c18b84a31b9af6a9d1da876ab28a88b0226707744679d4e180691aca6bdef5827622396751a0670c101\"]"
                }
            },
            "testExpect": {
                canExcuteTx: true,
                toBalanceChange: "0",
                status: 1,
                result: "\"n1F8QbdnhqpPXDPFT2c9a581tpia8iuF7o2\""
            }
        },
        {
            "name": "0-5. test recoverAddress invalid alg",
            "testInput": {
                value: "0",
                nonce: 1, 
                gasPrice: 1000000,
                gasLimit: 2000000,
                contract: {
                    function: "testRecoverAddress",
                    args: "[10,\"564733f9f3e139b925cfb1e7e50ba8581e9107b13e4213f2e4708d9c284be75b\",\"d80e282d165f8c05d8581133df7af3c7c41d51ec7cd8470c18b84a31b9af6a9d1da876ab28a88b0226707744679d4e180691aca6bdef5827622396751a0670c101\"]"
                }
            },
            "testExpect": {
                canExcuteTx: true,
                toBalanceChange: "0",
                status: 1,
                result: "null"
            }
        },
        {
            "name": "0-6. test md5",
            "testInput": {
                value: "0",
                nonce: 1, 
                gasPrice: 1000000,
                gasLimit: 2000000,
                contract: {
                    function: "testMd5",
                    args: "[\"Nebulas is a next generation public blockchain, aiming for a continuously improving ecosystem.\"]"
                }
            },
            "testExpect": {
                canExcuteTx: true,
                toBalanceChange: "0",
                status: 1,
                result: "\"9954125a33a380c3117269cff93f76a7\""
            }
        },
        {
            "name": "0-7. test base64",
            "testInput": {
                value: "0",
                nonce: 1, 
                gasPrice: 1000000,
                gasLimit: 2000000,
                contract: {
                    function: "testBase64",
                    args: "[\"https://y.qq.com/portal/player_radio.html#id=99\"]"
                }
            },
            "testExpect": {
                canExcuteTx: true,
                toBalanceChange: "0",
                status: 1,
                result: "\"aHR0cHM6Ly95LnFxLmNvbS9wb3J0YWwvcGxheWVyX3JhZGlvLmh0bWwjaWQ9OTk=\""
            }
        }
    ]
};
testCaseGroups.push(caseGroup);

describe('Contract crypto test', () => {

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