'use strict';

var expect = require('chai').expect;
var BigNumber = require('bignumber.js');
var HttpRequest = require("../../node-request");
var TestNetConfig = require("../testnet_config.js");
var Wallet;
try {
    Wallet = require("../../neb.js");
} catch (e) {
    Wallet = require("nebulas");
}
var utils = Wallet.Utils;
var Neb = Wallet.Neb;
var Account = Wallet.Account;
var Transaction = Wallet.Transaction;
var Unit = Wallet.Unit;

// mocha cases/contract/xxx testneb2 -t 2000000

var args = process.argv.splice(2);
var env = args[1];
var testNetConfig = new TestNetConfig(env);

// global vars.
var neb = new Neb();
var ChainID = testNetConfig.ChainId;
var sourceAccount = testNetConfig.sourceAccount;
var coinbase = testNetConfig.coinbase;
var apiEndPoint = testNetConfig.apiEndPoint;
neb.setRequest(new HttpRequest(apiEndPoint));


var coinState;


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
    neb.api.getTransactionReceipt(hash).then(function (resp) {
        console.log("0. tx receipt status:" + resp.status);
        if (resp.status === 2) {
            setTimeout(function () {
                checkTransaction(hash, callback);
            }, 2000);
        } else {
            checkTimes = 0;
            callback(resp);
        }
    }).catch(function (err) {
        console.log("1. fail to get tx receipt hash: " + hash);
        console.log("2. it may because the tx is being packing, we are going on to check it!");
        console.log("3. " + JSON.stringify(err));
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
    }).then(function (resp) {

        var toAddr = Account.NewAccount();
        if (testInput.isSameAddr === true) {
            toAddr = from;
        }

        coinState = resp;
        console.log("get coinbase state before tx:" + JSON.stringify(resp));


        var tx;

        if (!testInput.hasOwnProperty("payloadLength")) {
            tx = new Transaction(ChainID, from, toAddr, Unit.nasToBasic(testInput.transferValue), parseInt(fromState.nonce) + testInput.nonceIncrement, testInput.gasPrice, testInput.gasLimit);
        } else {
            var payloadContent = new Array(testInput.payloadLength + 1).join("s");
            console.log("payloadcontent:" + payloadContent)
            tx = new Transaction(ChainID, from, toAddr, Unit.nasToBasic(testInput.transferValue), parseInt(fromState.nonce) + testInput.nonceIncrement, testInput.gasPrice, testInput.gasLimit, payloadContent);
        }

        if (testInput.hasOwnProperty("overrideFromAddr")) {
            tx.from.address = Wallet.CryptoUtils.bufferToHex(testInput.overrideFromAddr);
            console.log("--> override tx.from.address with: " + testInput.overrideFromAddr);
        }

        if (testInput.hasOwnProperty("overrideToAddr")) {
            tx.to.address = Wallet.CryptoUtils.bufferToHex(testInput.overrideToAddr);
            console.log("--> override tx.to.address with: " + testInput.overrideToAddr);
        }

        if (testInput.hasOwnProperty("overrideGasLimit")) {
            tx.gasLimit = utils.toBigNumber(testInput.overrideGasLimit);
            console.log("--> override tx.gasLimit: " + tx.gasLimit);
        }

        if (testInput.hasOwnProperty("overrideGasPrice")) {
            tx.gasPrice = utils.toBigNumber(testInput.overrideGasPrice);
            console.log("--> override tx.gasPrice: " + tx.gasPrice);
        }

        tx.signTransaction();

        if (testInput.hasOwnProperty("overrideSignature")) {
            tx.sign = testInput.overrideSignature;
        } else if (testInput.fakeSign) {
            //repalce the privkey to sign
            console.log("this is the right signature:" + tx.sign.toString('hex'));
            console.log("repalce the privkey and sign another signatrue...")
            var newAccount = new Wallet.Account("a6e5eb222e4538fce79f5cb8774a72621637c2c9654c8b2525ed1d7e4e73653f");
            var privKey = tx.from.privKey
            tx.from.privKey = newAccount.privKey
            tx.signTransaction();
            console.log("now signatrue is: " + tx.sign.toString('hex'));
            tx.from.privKey = privKey;
        }

        console.log("tx to be sent: " + tx.toString());
        //console.log("tx to be send: " + JSON.stringify(JSON.parse(tx.toString()),null,'\t')); //json format output
        return neb.api.sendRawTransaction(tx.toProtoString());

    }).catch(function (err) {
        if (true === testExpect.canSendTx) {
            done(err);
        } else {
            console.log("cannot send tx, err: ", err);
            if (testExpect.hasOwnProperty("errMsg")) {
                expect(err.error.error).to.be.equal(testExpect.errMsg);
                // expect(testExpect.errMsg).to.be.equal(err);
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
                            expect(receipt).to.be.have.property('status').equal(0);
                        }
                        console.log("tx receipt : " + JSON.stringify(receipt));
                        neb.api.getAccountState(receipt.from).then(function (state) {

                            console.log("get from account state :" + JSON.stringify(state));
                            expect(state.balance).to.equal(testExpect.fromBalanceAfterTx);
                            return neb.api.getAccountState(receipt.to);
                        }).then(function (state) {

                            console.log("get to account state :" + JSON.stringify(state));
                            expect(state.balance).to.equal(testExpect.toBalanceAfterTx);
                            return neb.api.getAccountState(coinbase);
                        }).then(function (state) {

                            console.log("get coinbase account state after tx:" + JSON.stringify(state));
                            var reward = new BigNumber(state.balance).sub(coinState.balance);
                            reward = reward.mod(new BigNumber(1.42694).mul(new BigNumber(10).pow(18)));
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
                        console.log("transaction can send but submit failed");
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
        //TODO test case should fail: a tx which is expected "canNotSendTX" is send
        console.log(JSON.stringify(err));
        done(err);
    });
}


function prepare(done) {
    from = Account.NewAccount();
    neb.api.getAccountState(sourceAccount.getAddressString()).then(function (resp) {
        console.log("source state:" + JSON.stringify(resp));
        var tx = new Transaction(ChainID, sourceAccount, from, Unit.nasToBasic(initFromBalance), parseInt(resp.nonce) + 1);
        tx.signTransaction();
        // console.log("source tx:" + tx.toString());
        return neb.api.sendRawTransaction(tx.toProtoString());
    }).then(function (resp) {
        console.log("prepare: ", resp);
        checkTransaction(resp.txhash, function (resp) {
            try {
                expect(resp).to.be.have.property('status').equal(1);
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
    it('normal transfer11', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: -1,
            gasPrice: -1,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: true,
            canSubmitTx: true,
            canExcuteTx: true,
            fromBalanceAfterTx: '8999999980000000000',
            toBalanceAfterTx: '1000000000000000000',
            transferReward: '20000000000'
        };

        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });

    it('[address invalid] invalid fromAddr (length is odd)', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: -1,
            gasPrice: -1,
            nonceIncrement: 1,
            overrideFromAddr: "01239120abdcde01239120abdcde"
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: false,
            fromBalanceAfterTx: '8999999980000000000',
            toBalanceAfterTx: '1000000000000000000',
            transferReward: '20000000000',
            errMsg: 'address: invalid address format'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });

    it('[address invalid] invalid fromAddr (length is even)', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: -1,
            gasPrice: -1,
            nonceIncrement: 1,
            overrideFromAddr: "01239120abdcde01239120abdcde1"
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: false,
            fromBalanceAfterTx: '8999999980000000000',
            toBalanceAfterTx: '1000000000000000000',
            transferReward: '20000000000',
            errMsg: 'address: invalid address format'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });

    it('[address invalid] invalid fromAddr (length exceed limits)', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: -1,
            gasPrice: -1,
            nonceIncrement: 1,
            overrideFromAddr: "01239120abdcde01239120abdcde01239120abdcde01239120abdcde01239120abdcde01239120abdcde01239120abdcde01239120abdcde01239120abdcde01239120abdcde"
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: false,
            fromBalanceAfterTx: '8999999980000000000',
            toBalanceAfterTx: '1000000000000000000',
            transferReward: '20000000000',
            errMsg: 'address: invalid address format'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });

    it('[address invalid] invalid toAddr (length is odd)', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: -1,
            gasPrice: -1,
            nonceIncrement: 1,
            overrideToAddr: "01239120abdcde01239120abdcde"
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: false,
            fromBalanceAfterTx: '8999999980000000000',
            toBalanceAfterTx: '1000000000000000000',
            transferReward: '20000000000',
            errMsg: 'address: invalid address format'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });

    it('[address invalid] invalid toAddr (length is even)', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: -1,
            gasPrice: -1,
            nonceIncrement: 1,
            overrideToAddr: "01239120abdcde01239120abdcde1"
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: false,
            fromBalanceAfterTx: '8999999980000000000',
            toBalanceAfterTx: '1000000000000000000',
            transferReward: '20000000000',
            errMsg: 'address: invalid address format'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });

    it('[address invalid] invalid toAddr (length exceed limits)', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: -1,
            gasPrice: -1,
            nonceIncrement: 1,
            overrideToAddr: "01239120abdcde01239120abdcde01239120abdcde01239120abdcde01239120abdcde01239120abdcde01239120abdcde01239120abdcde01239120abdcde01239120abdcde"
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: false,
            fromBalanceAfterTx: '8999999980000000000',
            toBalanceAfterTx: '1000000000000000000',
            transferReward: '20000000000',
            errMsg: 'address: invalid address format'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });

    it('[address invalid] from addr empty', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: -1,
            gasPrice: -1,
            nonceIncrement: 1,
            overrideFromAddr: ""
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: false,
            fromBalanceAfterTx: '8999999980000000000',
            toBalanceAfterTx: '1000000000000000000',
            transferReward: '20000000000',
            errMsg: 'address: invalid address format'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });

    it('[address invalid] to addr empty', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: -1,
            gasPrice: -1,
            nonceIncrement: 1,
            overrideToAddr: ""
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: false,
            fromBalanceAfterTx: '8999999980000000000',
            toBalanceAfterTx: '1000000000000000000',
            transferReward: '20000000000',
            errMsg: 'address: invalid address format'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });

    it('[address] from & to are same', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: true,
            gasLimit: -1,
            gasPrice: -1,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: true,
            canSubmitTx: true,
            canExcuteTx: true,
            fromBalanceAfterTx: '9999999980000000000',
            toBalanceAfterTx: '9999999980000000000',
            transferReward: '20000000000'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });

    it('[signature] invalid signature (wrong sig)', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: -1,
            gasPrice: -1,
            nonceIncrement: 1,
            overrideSignature: "some_wrong_sig"
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: true,
            fromBalanceAfterTx: '8999999980000000000',
            toBalanceAfterTx: '1000000000000000000',
            transferReward: '20000000000',
            errMsg: 'invalid signature'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });

    it('[signature] invalid signature (empty sig)', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: -1,
            gasPrice: -1,
            nonceIncrement: 1,
            overrideSignature: ""
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: true,
            fromBalanceAfterTx: '8999999980000000000',
            toBalanceAfterTx: '1000000000000000000',
            transferReward: '20000000000',
            errMsg: 'invalid signature'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });

    it('[signature] invalid signature (fake sig)', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: -1,
            gasPrice: -1,
            nonceIncrement: 1,
            fakeSign: true
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: false,
            fromBalanceAfterTx: '',
            toBalanceAfterTx: '',
            transferReward: '',
            errMsg: 'invalid transaction signer'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });

    it('[gasLimit insufficient] 0 < gasLimit = TxBaseGasCount-1', function (done) {

        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: 19999,
            gasPrice: -1,
            nonceIncrement: 1
        };

        var testExpect = {
            canSendTx: true,
            canSubmitTx: false,
            canExcuteTx: true,
            fromBalanceAfterTx: '8999999980000000000',
            toBalanceAfterTx: '1000000000000000000',
            transferReward: '20000000000'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });

    it('[gasLimit insufficient] 1 = gasLimit < TxBaseGasCount', function (done) {

        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: 1,
            gasPrice: -1,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: true,
            canSubmitTx: false,
            canExcuteTx: false,
            fromBalanceAfterTx: '-1',
            toBalanceAfterTx: '-1',
            transferReward: '-1'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });


    it('[gasLimit sufficient] gasLimit<0', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: 0,
            overrideGasLimit: -100,
            gasPrice: -1,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: false,
            fromBalanceAfterTx: '8999999980000000000',
            toBalanceAfterTx: '1000000000000000000',
            transferReward: '20000000000',
            errMsg: 'invalid gas limit, should be in (0, 5*10^10]'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });

    it('[gasLimit sufficient] gasLimit=0', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: 0,
            overrideGasLimit: 0,
            gasPrice: -1,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: false,
            fromBalanceAfterTx: '8999999980000000000',
            toBalanceAfterTx: '1000000000000000000',
            transferReward: '20000000000',
            errMsg: 'invalid gas limit, should be in (0, 5*10^10]'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });

    it('[gasLimit sufficient] gasLimit > TransactionMaxGase', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: 50000000001,
            gasPrice: -1,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: false,
            fromBalanceAfterTx: '10000000000000000000',
            toBalanceAfterTx: '0',
            transferReward: '0',
            errMsg: 'invalid gas limit, should be in (0, 5*10^10]'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });

    it('[gasLimit sufficient] gasLimit=TxBaseGasCount', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: 20000,
            gasPrice: -1,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: true,
            canSubmitTx: true,
            canExcuteTx: true,
            fromBalanceAfterTx: '8999999980000000000',
            toBalanceAfterTx: '1000000000000000000',
            transferReward: '20000000000'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });

    it('[gasLimit sufficient] gasLimit>TxBaseGasCount', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: 20001,
            gasPrice: -1,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: true,
            canSubmitTx: true,
            canExcuteTx: true,
            fromBalanceAfterTx: '8999999980000000000',
            toBalanceAfterTx: '1000000000000000000',
            transferReward: '20000000000'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });

    it('[gasPrice insufficient] gasPrice<0', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: -1,
            gasPrice: -1,
            overrideGasPrice: -1,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: false,
            fromBalanceAfterTx: '8999999980000000000',
            toBalanceAfterTx: '1000000000000000000',
            transferReward: '20000000000',
            errMsg: 'invalid gas price, should be in (0, 10^12]'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });

    it('[gasPrice sufficient] gasPrice=0', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: 0,
            gasPrice: 0,
            overrideGasPrice: 0,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: false,
            fromBalanceAfterTx: '8999999980000000000',
            toBalanceAfterTx: '1000000000000000000',
            transferReward: '20000000000',
            errMsg: 'invalid gas price, should be in (0, 10^12]'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });

    it('[gasPrice sufficient] gasPrice = txPool.gasPrice', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: 0,
            gasPrice: 1000000,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: true,
            canSubmitTx: true,
            canExcuteTx: true,
            fromBalanceAfterTx: '8999999980000000000',
            toBalanceAfterTx: '1000000000000000000',
            transferReward: '20000000000'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });
    it('[gasPrice sufficient] gasPrice = 2 * txPool.baseGasPrice', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: -1,
            gasPrice: 2000000,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: true,
            canSubmitTx: true,
            canExcuteTx: true,
            fromBalanceAfterTx: '8999999960000000000',
            toBalanceAfterTx: '1000000000000000000',
            transferReward: '40000000000'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });

    it('gas price is too small', function (done) {

        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: -1,
            gasPrice: 1,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: false,
            fromBalanceAfterTx: '-1',
            toBalanceAfterTx: '-1',
            transferReward: '-1',
            errMsg: 'below the gas price'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });

    it('[gasPrice insufficient] gasPrice < txPool.gasPrice', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: 0,
            gasPrice: 99999,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: false,
            fromBalanceAfterTx: '',
            toBalanceAfterTx: '',
            transferReward: '',
            errMsg: 'below the gas price'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });

    it('[balanceOfFrom insufficient] gasPrice * gasLimit <= balanceOfFrom < valueOfTx ', function (done) {

        var testInput = {
            transferValue: 10.1,
            isSameAddr: false,
            gasLimit: -1,
            gasPrice: -1,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: true,
            canSubmitTx: true,
            canExcuteTx: false,
            fromBalanceAfterTx: '9999999980000000000',
            toBalanceAfterTx: '0',
            transferReward: '20000000000',
            eventError: 'insufficient balance'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });

    it('[balanceOfFrom insufficient] balanceOfFrom < TxBaseGasCount * gasPrice + valueOfTx', function (done) {

        var testInput = {
            transferValue: 9.999999999999,
            isSameAddr: false,
            gasLimit: -1,
            gasPrice: -1,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: true,
            canSubmitTx: true,
            canExcuteTx: false,
            fromBalanceAfterTx: '9999999980000000000',
            toBalanceAfterTx: '0',
            transferReward: '20000000000'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });



    it('[balanceOfFrom insufficient] balanceOfFrom < gasPrice * gasLimit', function (done) {

        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: 50000000000,
            gasPrice: 210000000,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: true,
            canSubmitTx: false,
            canExcuteTx: false,
            fromBalanceAfterTx: '-1',
            toBalanceAfterTx: '-1',
            transferReward: '-1'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });

    it('[balanceOfFrom insufficient] from = to && balanceOfFrom < valueOfTx', function (done) {

        var testInput = {
            transferValue: 15,
            isSameAddr: true,
            gasLimit: -1,
            gasPrice: -1,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: true,
            canSubmitTx: true,
            canExcuteTx: false,
            fromBalanceAfterTx: '9999999980000000000',
            toBalanceAfterTx: '9999999980000000000',
            transferReward: '20000000000',
            eventError: 'insufficient balance'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });

    });



    it('[nonce check] nonce < from.nonce + 1', function (done) {

        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: -1,
            gasPrice: -1,
            nonceIncrement: 0
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: false,
            fromBalanceAfterTx: '-1',
            toBalanceAfterTx: '-1',
            transferReward: '-1',
            errMsg: 'transaction\'s nonce is invalid, should bigger than the from\'s nonce'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });


    it('[nonce check] nonce > from.nonce + 1', function (done) {

        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: -1,
            gasPrice: -1,
            nonceIncrement: 2
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: true,
            canSubmitTx: false,
            canExcuteTx: false,
            fromBalanceAfterTx: '-1',
            toBalanceAfterTx: '-1',
            transferReward: '-1'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });


    it('[nonce check] nonce < 0', function (done) {

        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: -1,
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
            errMsg: 'invalid transaction hash' //TODO is this error right?
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });


    it('[balanceOfFrom sufficient] balanceOfFrom = GasLimit * GasPrice + valueOfTx ', function (done) {

        var testInput = {
            transferValue: 9.99999998,
            isSameAddr: false,
            gasLimit: 20000,
            gasPrice: 1000000,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: true,
            canSubmitTx: true,
            canExcuteTx: true,
            fromBalanceAfterTx: '0',
            toBalanceAfterTx: '9999999980000000000',
            transferReward: '20000000000'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });


    it('[balanceOfFrom sufficient] (GasLimit * GasPrice + valueOfTx) > balanceOfFrom ' +
        '&& balanceOfFrom = TxBaseGasCount * GasPrice + valueOfTx  ', function (done) {

        var testInput = {
            transferValue: 9.99999998,
            isSameAddr: false,
            gasLimit: 400000,
            gasPrice: 1000000,
            nonceIncrement: 1
        };

        // The TxBaseGasCount in neb is 20000

        var testExpect = {
            canSendTx: true,
            canSubmitTx: true,
            canExcuteTx: false,
            fromBalanceAfterTx: '9999999980000000000',
            toBalanceAfterTx: '0',
            transferReward: '20000000000'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });


    it('[ payload > 0 ] normal transfer', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: 30000,
            gasPrice: -1,
            nonceIncrement: 1,
            payloadLength: 99
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: true,
            canSubmitTx: true,
            canExcuteTx: true,
            fromBalanceAfterTx: '8999999979964000000',
            toBalanceAfterTx: '1000000000000000000',
            transferReward: '20036000000'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });

    it('[ payload > 0 gaslimit ] TxBaseGasCount > gasLimit = minGasPerTransaction ', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: 20001,
            gasPrice: -1,
            nonceIncrement: 1,
            payloadLength: 99
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: true,
            canSubmitTx: false,
            canExcuteTx: false,
            fromBalanceAfterTx: '8999999979964000000',
            toBalanceAfterTx: '1000000000000000000',
            transferReward: '20036000000'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });

    it('[ payload > 0 gaslimit ] TxBaseGasCount > gasLimit = TxBaseGasCount-1 ', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: 20035,
            gasPrice: -1,
            nonceIncrement: 1,
            payloadLength: 99
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: true,
            canSubmitTx: false,
            canExcuteTx: false,
            fromBalanceAfterTx: '8999999979964000000',
            toBalanceAfterTx: '1000000000000000000',
            transferReward: '20036000000'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });



    it('[ payload > 0 ] TxBaseGasCount = gasLimit ', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: 20036,
            gasPrice: -1,
            nonceIncrement: 1,
            payloadLength: 99
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: true,
            canSubmitTx: true,
            canExcuteTx: true,
            fromBalanceAfterTx: '8999999979964000000',
            toBalanceAfterTx: '1000000000000000000',
            transferReward: '20036000000'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });

    // TxBaseGasCount = min + data.length
    it('[ data.length > 0 && balance ] balanceOfFrom = TxBaseGasCount * GasPrice + valueOfTx ', function (done) {
        var testInput = {
            transferValue: 9.999999979964,
            isSameAddr: false,
            gasLimit: 20036,
            gasPrice: -1,
            nonceIncrement: 1,
            payloadLength: 99
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: true,
            canSubmitTx: true,
            canExcuteTx: true,
            fromBalanceAfterTx: '0',
            toBalanceAfterTx: '9999999979964000000',
            transferReward: '20036000000'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });



    it('[ data.length > 0 ] balanceOfFrom < GasLimit * GasPrice ', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            gasLimit: 50000000001,
            gasPrice: 10000000,
            nonceIncrement: 1,
            payloadLength: 99
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: false,
            fromBalanceAfterTx: '8999999979964000000',
            toBalanceAfterTx: '1000000000000000000',
            transferReward: '20036000000',
            errMsg: 'invalid gas limit, should be in (0, 5*10^10]'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });


    it('[ payload > 0 ] (GasLimit * GasPrice + valueOfTx) > balanceOfFrom ' +
        '&& balanceOfFrom = ( TxBaseGasCount + GasCountOfPayload )* gasPrice + valueOfTx', function (done) {
        var testInput = {
            transferValue: 9.999999979964,
            isSameAddr: false,
            gasLimit: 30000,
            gasPrice: -1,
            nonceIncrement: 1,
            payloadLength: 99
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: true,
            canSubmitTx: true,
            canExcuteTx: false,
            fromBalanceAfterTx: '9999999979964000000',
            toBalanceAfterTx: '0',
            transferReward: '20036000000'
        };
        prepare(function (err) {
            if (err instanceof Error) {
                done(err);
            } else {
                testTransfer(testInput, testExpect, done);
            }
        });
    });

    it('[ payload > 0 ] gasPrice * gasLimit <= balanceOfFrom  ' +
        '&& balanceOfFrom < valueOfTx + (TxBaseGasCount  + GasCountOfPayload)* gasPrice', function (done) {
        var testInput = {
            transferValue: 9.999999999,
            isSameAddr: false,
            gasLimit: 30000,
            gasPrice: -1,
            nonceIncrement: 1,
            payloadLength: 99
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: true,
            canSubmitTx: true,
            canExcuteTx: false,
            fromBalanceAfterTx: '9999999979964000000',
            toBalanceAfterTx: '0',
            transferReward: '20036000000'
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
