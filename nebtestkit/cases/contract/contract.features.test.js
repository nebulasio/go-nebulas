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

var maxCheckTime = 50;
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
                var contractSource = FS.readFileSync("./nf/nvm/test/test_contract_features.js", "UTF-8");
                var contract = {
                    'source': contractSource,
                    "sourceType": "js",
                    "args": ''
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
                    try {
                        expect(resp).to.not.be.a('undefined');
                        expect(resp.status).to.be.equal(1);
                        console.log("----step2. have been on chain");
                        done();
                    } catch(err) {
                        console.log(err);
                        done(err);
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

    it ('1. test feature getAccoutState 1/2', function (done) {
        var contract = {
            'function': 'testGetAccountState1',
            "args": '[]',
        };
        neb.api.call(sourceAccount.getAddressString(), contractAddress, Unit.nasToBasic(10), nonce, 200000, 1000000, contract).then(function(resp){
            console.log(JSON.stringify(resp));
            expect(resp.execute_err).to.be.equal("");
            done();
        }).catch(function(err) {
            console.log(err);
            done(err);
        });
    });

    it ('2. test feature getAccoutState 2/2', function (done) {
        var contract = {
            'function': 'testGetAccountState2',
            "args": '[]',
        };
        neb.api.call(sourceAccount.getAddressString(), contractAddress, Unit.nasToBasic(10), nonce, 200000, 1000000, contract).then(function(resp){
            console.log(JSON.stringify(resp));
            expect(resp.execute_err).to.be.equal("Call: Blockchain.getAccountState(), parse address failed");
            done();
        }).catch(function(err) {
            console.log(err);
            done(err);
        });
    });

    it ('3. test feature getPreBlockHash 1/5[offset is 1]', function (done) {
        var offset = 1;
        var contract = {
            'function': 'testGetPreBlockHash',
            "args": '[' + offset + ']',
        };
        var hash;
        var height;
        neb.api.call(sourceAccount.getAddressString(), contractAddress, Unit.nasToBasic(10), nonce, 200000, 1000000, contract).then(function(resp){
            console.log(JSON.stringify(resp));
            var result = JSON.parse(resp.result);
            hash = result.hash;
            height = result.height;
            return neb.api.getBlockByHeight(height - offset);
        }).then(function(resp) {
            // console.log(JSON.stringify(resp));
            expect(resp.hash).to.be.equal(hash);
            done();
        }).catch(function(err) {
            console.log(err);
            done(err);
        });
    });

    it ('4. test feature getPreBlockHash 2/5[offset is "1"]', function (done) {
        var offset = 1;
        var contract = {
            'function': 'testGetPreBlockHash',
            "args": "[\"" + offset + "\"]",
        };
        var hash;
        var height;
        neb.api.call(sourceAccount.getAddressString(), contractAddress, Unit.nasToBasic(10), nonce, 200000, 1000000, contract).then(function(resp){
            console.log(JSON.stringify(resp));
            var result = JSON.parse(resp.result);
            hash = result.hash;
            height = result.height;
            return neb.api.getBlockByHeight(height - offset);
        }).then(function(resp) {
            // console.log(JSON.stringify(resp));
            expect(resp.hash).to.be.equal(hash);
            done();
        }).catch(function(err) {
            console.log(err);
            done(err);
        });
    });

    it ('5. test feature getPreBlockHash 3/5[offset is "#1"]', function (done) {
        var offset = "#1";
        var contract = {
            'function': 'testGetPreBlockHash',
            "args": "[\"" + offset + "\"]",
        };
        neb.api.call(sourceAccount.getAddressString(), contractAddress, Unit.nasToBasic(10), nonce, 200000, 1000000, contract).then(function(resp){
            console.log(JSON.stringify(resp));
            expect(resp.execute_err).to.be.equal("Call: getPreBlockHash: invalid offset");
            done();
        }).catch(function(err) {
            console.log(err);
            done(err);
        });
    });

    it ('6. test feature getPreBlockHash 4/5[offset is 0]', function (done) {
        var offset = 0;
        var contract = {
            'function': 'testGetPreBlockHash',
            "args": '[' + offset + ']',
        };
        neb.api.call(sourceAccount.getAddressString(), contractAddress, Unit.nasToBasic(10), nonce, 200000, 1000000, contract).then(function(resp){
            console.log(JSON.stringify(resp));
            expect(resp.execute_err).to.be.equal("Call: getPreBlockHash: invalid offset");
            done();
        }).catch(function(err) {
            console.log(err);
            done(err);
        });
    });

    it ('7. test feature getPreBlockHash 5/5[offset is too large]', function (done) {
        var offset = 11111111111111;
        var contract = {
            'function': 'testGetPreBlockHash',
            "args": '[' + offset + ']',
        };
        neb.api.call(sourceAccount.getAddressString(), contractAddress, Unit.nasToBasic(10), nonce, 200000, 1000000, contract).then(function(resp){
            console.log(JSON.stringify(resp));
            expect(resp.execute_err).to.be.equal("Call: getPreBlockHash: block not exist");
            done();
        }).catch(function(err) {
            console.log(err);
            done(err);
        });
    });

    it ('8. test feature getPreBlockSeed 1/5[offset is 1]', function (done) {
        var offset = 1;
        var contract = {
            'function': 'testGetPreBlockSeed',
            "args": '[' + offset + ']',
        };
        var seed;
        var height;
        neb.api.call(sourceAccount.getAddressString(), contractAddress, Unit.nasToBasic(10), nonce, 200000, 1000000, contract).then(function(resp){
            console.log(JSON.stringify(resp));
            var result = JSON.parse(resp.result);
            seed = result.seed;
            console.log(seed);
            height = result.height;
            return neb.api.getBlockByHeight(height - offset);
        }).then(function(resp) {
            console.log(JSON.stringify(resp));
            expect(resp.randomSeed).to.be.equal(seed);
            done();
        }).catch(function(err) {
            console.log(err);
            done(err);
        });
    });

    it ('9. test feature getPreBlockSeed 2/5[offset is "1"]', function (done) {
        var offset = 1;
        var contract = {
            'function': 'testGetPreBlockSeed',
            "args": "[\"" + offset + "\"]",
        };
        var seed;
        var height;
        neb.api.call(sourceAccount.getAddressString(), contractAddress, Unit.nasToBasic(10), nonce, 200000, 1000000, contract).then(function(resp){
            console.log(JSON.stringify(resp));
            var result = JSON.parse(resp.result);
            seed = result.seed;
            height = result.height;
            return neb.api.getBlockByHeight(height - offset);
        }).then(function(resp) {
            // console.log(JSON.stringify(resp));
            expect(resp.randomSeed).to.be.equal(seed);
            done();
        }).catch(function(err) {
            console.log(err);
            done(err);
        });
    });

    it ('10. test feature getPreBlockSeed 3/5[offset is 0]', function (done) {
        var offset = 0;
        var contract = {
            'function': 'testGetPreBlockSeed',
            "args": '[' + offset + ']',
        };
        neb.api.call(sourceAccount.getAddressString(), contractAddress, Unit.nasToBasic(10), nonce, 200000, 1000000, contract).then(function(resp){
            console.log(JSON.stringify(resp));
            expect(resp.execute_err).to.be.equal("Call: getPreBlockSeed: invalid offset");
            done();
        }).catch(function(err) {
            console.log(err);
            done(err);
        });
    });

    it ('11. test feature getPreBlockSeed 4/5[offset is too large]', function (done) {
        var offset = 11111111111111;
        var contract = {
            'function': 'testGetPreBlockSeed',
            "args": '[' + offset + ']',
        };
        neb.api.call(sourceAccount.getAddressString(), contractAddress, Unit.nasToBasic(10), nonce, 200000, 1000000, contract).then(function(resp){
            console.log(JSON.stringify(resp));
            expect(resp.execute_err).to.be.equal("Call: getPreBlockSeed: block not exist");
            done();
        }).catch(function(err) {
            console.log(err);
            done(err);
        });
    });
    it ('12. test feature getPreBlockSeed 5/5[offset is "#1"]', function (done) {
        var offset = "#1";
        var contract = {
            'function': 'testGetPreBlockSeed',
            "args": "[\"" + offset + "\"]",
        };
        neb.api.call(sourceAccount.getAddressString(), contractAddress, Unit.nasToBasic(10), nonce, 200000, 1000000, contract).then(function(resp){
            console.log(JSON.stringify(resp));
            expect(resp.execute_err).to.be.equal("Call: getPreBlockSeed: invalid offset");
            done();
        }).catch(function(err) {
            console.log(err);
            done(err);
        });
    });
});
