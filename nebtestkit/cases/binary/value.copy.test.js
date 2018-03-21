'use strict';

var HttpRequest = require("../../node-request");

var Wallet = require("../../../cmd/console/neb.js/lib/wallet");
var Neb = Wallet.Neb;
var neb = new Neb();
var Account = Wallet.Account;
var Transaction = Wallet.Transaction;
var Unit = Wallet.Unit;
var utils = require("../../../cmd/console/neb.js/lib/wallet").Utils;


var expect = require('chai').expect;
var BigNumber = require('bignumber.js');
neb.setRequest(new HttpRequest("http://localhost:8685"));
var ChainID = 100;
var sourceAccount = new Account("a6e5eb290e1438fce79f5cb8774a72621637c2c9654c8b2525ed1d7e4e73653f");
/*
 * make sure every node of testnet has the same coinbase, and substitute the address below
 */
var coinbase = "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8";
var coinState;


// mocha cases/contract/xxx testneb1 -t 200000
console.log("process.argv: ", process.argv);
var args = process.argv.splice(2);
//console.log("process: ", process);
console.log("argv.splice(2): ", args);
//args.forEach(function(item,index){console.log("arg[",index,"] = ", item);} ); //to see what args is

var env = args[1];
if (env !== "local" && env !== "testneb1" && env !== "testneb2" && env !== "testneb3") {
    env = "local";
}
console.log("env:", env);


if (env === 'testneb1') {
    neb.setRequest(new HttpRequest("http://35.182.48.19:8685"));
    ChainID = 1001;
    sourceAccount = new Wallet.Account("43181d58178263837a9a6b08f06379a348a5b362bfab3631ac78d2ac771c5df3");
    coinbase = "0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d";
} else if (env === "testneb2") {
    neb.setRequest(new HttpRequest("http://34.205.26.12:8685"));
    ChainID = 1002;
    sourceAccount = new Wallet.Account("43181d58178263837a9a6b08f06379a348a5b362bfab3631ac78d2ac771c5df3");
    coinbase = "0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d";
} else if (env === "testneb3") {
    neb.setRequest(new HttpRequest("http://35.177.214.138:8685"));
    ChainID = 1003;
    sourceAccount = new Wallet.Account("43181d58178263837a9a6b08f06379a348a5b362bfab3631ac78d2ac771c5df3");
    coinbase = "0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d";
}

var from;
var fromState;
var initFromBalance = 10;

/*
 * set this value according to the status of your testnet.
 * the smaller the value, the faster the test, with the risk of causing error
 */

var maxCheckTime = 40;
var checkTimes = 0;

function checkTransaction(hash, callback) {
    checkTimes += 1;

    if (checkTimes > maxCheckTime) {
        console.log("check tx receipt timeout:" + hash);
        checkTimes = 0;
        callback();
        return;
    }
    neb.api.getTransactionReceipt(hash).then(function (resp) { //what's resp? the return value of getTransactionReceipt?
        console.log("tx receipt status:" + resp.status);  //status is not a property of resp?
        if (resp.status === 2) {
            setTimeout(function () {
                checkTransaction(hash, callback);
            }, 2000);
        } else {
            checkTimes = 0;         //
            callback(resp);
        }
    }).catch(function (err) {           //what's this err? an err thrown by getTransactionReceipt?
        console.log("fail to get tx receipt hash: " + hash);
        console.log("it may because the tx is being packing, we are going on to check it!");
        console.log(err.error);
        setTimeout(function () {
            checkTransaction(hash, callback);
        }, 2000);
    });
}

function testTransfer(testInput, testExpect, done) {
    neb.api.getAccountState(from.getAddressString()).then(function (state) {

        fromState = state;
        console.log("from state:" + JSON.stringify(state));
        return neb.api.getAccountState(coinbase);
    }).then(function (resp) {                   //is resp equals neb.api.getAccountState(coinbase)?

        var toAddr = Account.NewAccount();      //todo to be optimized,
        if (testInput.isSameAddr === true) {
            toAddr = from;
        }

        coinState = resp;
        console.log("get coinbase state before tx:" + JSON.stringify(resp));


        var tx;

        if (!testInput.hasOwnProperty("payloadLength")){
            tx = new Transaction(ChainID, from, toAddr, Unit.nasToBasic(testInput.transferValue), parseInt(fromState.nonce) + testInput.nonceIncrement, testInput.gasPrice, testInput.gasLimit);
        } else {
            var payloadContent = new Array(testInput.payloadLength + 1).join("s");
            console.log("payloadcontent:" + payloadContent)
            tx = new Transaction(ChainID, from, toAddr, Unit.nasToBasic(testInput.transferValue), parseInt(fromState.nonce) + testInput.nonceIncrement, testInput.gasPrice, testInput.gasLimit, payloadContent);
        }

        if(testInput.hasOwnProperty("overrideFromAddr")) {
            tx.from.address = Wallet.CryptoUtils.bufferToHex(testInput.overrideFromAddr);
            console.log("--> override tx.from.address with: " + testInput.overrideFromAddr);
        }

        if(testInput.hasOwnProperty("overrideToAddr")) {
            tx.to.address = Wallet.CryptoUtils.bufferToHex(testInput.overrideToAddr);
            console.log("--> override tx.to.address with: " + testInput.overrideToAddr);
        }

        if (testInput.hasOwnProperty("overrideGasLimit")){
            tx.gasLimit = utils.toBigNumber(testInput.overrideGasLimit);
            console.log("--> override tx.gasLimit: " + tx.gasLimit);
        }

        if (testInput.hasOwnProperty("overrideGasPrice")){
            tx.gasPrice = utils.toBigNumber(testInput.overrideGasPrice);
            console.log("--> override tx.gasPrice: " + tx.gasPrice);
        }

        tx.signTransaction();

        if(testInput.hasOwnProperty("overrideSignature")){
            tx.sign = testInput.overrideSignature;
        }

        return neb.api.sendRawTransaction(tx.toProtoString());

    }).catch(function (err) {
        if (true === testExpect.canSendTx) {
            done(err);              //done(err) ??
        } else {
            console.log("cannot send tx, err: "+err)
            if (testExpect.hasOwnProperty("errMsg")){
                expect(testExpect.errMsg).to.be.equal(err.error.error);
            }
            done();
        }
    }).then(function (resp) {

        console.log("resp11:" + JSON.stringify(resp));

        if (true === testExpect.canSendTx) {
            console.log("send Raw Tx:" + JSON.stringify(resp));
            expect(resp).to.be.have.property('txhash');
            checkTransaction(resp.txhash, function (receipt) {

                try {
                    if (true === testExpect.canSubmitTx) {
                        expect(receipt).to.not.be.a('undefined');
                        if (true === testExpect.canExcuteTx) {
                            expect(receipt).to.be.have.property('status').equal(1);
                        } else {
                            expect(receipt).to.not.have.property('status');
                        }
                        console.log("tx receipt : " + JSON.stringify(receipt));
                        neb.api.getAccountState(receipt.from).then(function (state) {

                            console.log("get from account state :" + JSON.stringify(state));
                            expect(state.balance).to.equal(testExpect.fromBalanceAfterTx);
                            //expect(state.balance).to.not.equal(testExpect.fromBalanceAfterTx);
                            return neb.api.getAccountState(receipt.to);
                        }).then(function (state) {

                            console.log("get to account state :" + JSON.stringify(state));
                            expect(state.balance).to.equal(testExpect.toBalanceAfterTx);
                            return neb.api.getAccountState(coinbase);
                        }).then(function (state) {

                            console.log("get coinbase account state after tx:" + JSON.stringify(state));
                            var reward = new BigNumber(state.balance).sub(coinState.balance);
                            reward = reward.mod(new BigNumber(0.48).mul(new BigNumber(10).pow(18)));
                            // The transaction should be only
                            expect(reward.toString()).to.equal(testExpect.transferReward);
                            return neb.api.getEventsByHash(resp.txhash);
                        }).then(function (eventResult) {
                            console.log("[eventCheck] event[0] topic: " + JSON.stringify(eventResult.events[0].topic));
                            if (ChainID !== 100) {
                                expect(eventResult.events[0].topic).to.equal("chain.transactionResult");
                            }
                            if (eventResult.hasOwnProperty('eventError')) {
                                expect(eventResult.events[0].error).to.equal(eventResult.eventError);
                            }
                            done();
                        }).catch(function (err) {
                            console.log(JSON.stringify(err));
                            done(err);
                        });
                    } else {
                        expect(receipt).to.be.a('undefined');
                        done();
                    }
                } catch (err) {
                    console.log(JSON.stringify(err));
                    done(err);
                }
            });
        } else {
            console.log(JSON.stringify(resp))
            expect(resp).to.be.a('undefined');
        }
    }).catch(function (err) {

        console.log(JSON.stringify(err));
        done(err);
    });
}

//create a new account as "from" in the following transaction.
//then transfer 10 nas to this "from" account
//
function prepare(done) {
    from = Account.NewAccount();
    neb.api.getAccountState(sourceAccount.getAddressString()).then(function (resp) {

        console.log("source state:" + JSON.stringify(resp));
        var tx = new Transaction(ChainID, sourceAccount, from, Unit.nasToBasic(initFromBalance), parseInt(resp.nonce) + 1);
        tx.signTransaction();
        // console.log("source tx:" + tx.toString());
        return neb.api.sendRawTransaction(tx.toProtoString());
    }).then(function (resp) {
        checkTransaction(resp.txhash, function (resp) {
            try {
                expect(resp).to.be.have.property('status').equal(1); //expect(resp).to.be.have.property('status',1)
                console.log("complete from address claim.");
                done();
            } catch (err) {
                done(err);
            }
        });
    }).catch(function (err) {
        console.log("claim token failed:" + JSON.stringify(err));
        done(err);
    });
};

describe('normal transaction', function () {


    it('[nonce check] nonce < 0', function (done) {

        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: 1,
            gasPrice: -1,
            nonceIncrement: -10000000000
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: false,
            fromBalanceAfterTx: '-1',
            toBalanceAfterTx: '-1',
            transferReward: '-1',
            //errMsg: 'transaction\'s nonce is invalid, should bigger than the from\'s nonce'
            errMsg: 'invalid transaction hash'  //TODO is this error right?
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });


});
