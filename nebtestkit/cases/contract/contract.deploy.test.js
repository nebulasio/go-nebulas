'use strict';

//var cryptoUtils = require('../../../cmd/console/neb.js/lib/utils/crypto-utils.js');
var Wallet = require("nebulas");
var cryptoUtils = Wallet.CryptoUtils;
var HttpRequest = require('../../node-request');
var Neb = Wallet.Neb;
var Account = Wallet.Account;
var Transaction = Wallet.Transaction;
var Utils = Wallet.Utils;
var Unit = Wallet.Unit;

var expect = require('chai').expect;
var BigNumber = require('bignumber.js');
var FS = require("fs");
var TestNetConfig = require("../testnet_config.js");

//mocha cases/contract/contract.deploy.test.js testneb2 -t 200000

var args = process.argv.splice(2);
var env = args[1];
if (env !== "local" && env !== "testneb1" && env !== "testneb2" && env !== "testneb3" && env !== "maintest") {
    env = "local";
}
var testNetConfig = new TestNetConfig(env);

var coinState,
    from,
    fromState,
    contractAddr,
    initFromBalance = 10;

var neb = new Neb();
var ChainID = testNetConfig.ChainId;
var sourceAccount = testNetConfig.sourceAccount;
var coinbase = testNetConfig.coinbase;
var apiEndPoint = testNetConfig.apiEndPoint;
neb.setRequest(new HttpRequest(apiEndPoint));



/*
 * set this value according to the status of your testnet.
 * the smaller the value, the faster the test, with the risk of causing error
 */

var maxCheckTime = 30;
var checkTimes = 0;
var beginCheckTime;


//callback must not throw any error!!!
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
        console.log("error details: " + JSON.stringify(err.error));
        console.log("maybe packing!");
        setTimeout(function () {
            checkTransaction(hash, callback);
        }, 2000);
    });
}

function testContractDeploy(testInput, testExpect, done) {
    neb.api.getAccountState(from.getAddressString()).then(function (state) {
        fromState = state;
        console.log("1. get from state:" + JSON.stringify(state));
        return neb.api.getAccountState(coinbase);
    }).then(function (resp) {
        coinState = resp;
        console.log("2. get coinbase state before tx:" + JSON.stringify(resp));
        var erc20 = FS.readFileSync("../nf/nvm/test/ERC20.js", "utf-8");
        var contract = {
            "source": erc20,
            "sourceType": "js",
            "args": '["NebulasToken", "NAS", 1000000000]'
        };

        if (testInput.canNotInit) {
            contract.source = FS.readFileSync("../nf/nvm/test/ERC20.initError.js", "utf-8");
        } else if (testInput.isSourceEmpty) {
            contract.source = "";
        } else if (testInput.isSyntaxErr) {
            contract.source = FS.readFileSync("../nf/nvm/test/ERC20.syntaxError.js", "utf-8");
        } else if (testInput.isParamErr) {
            contract.args = '["NebulasToken", 100,"NAS", 1000000000]';
        } else if (testInput.isNormalJs) {
            contract.source = "console.log('this is not contract but a normal js file')";
        } else if (testInput.NoInitFunc) {
            contract.source = FS.readFileSync("../nf/nvm/test/contract.noInitFunc.test.js", "utf-8");
        } else if (testInput.isTypeEmpty) {
            contract.sourceType = "";
        } else if (testInput.isTypeWrong) {
            contract.sourceType = "c";
        } else {
            if (testInput.isSourceTypeTs) {
                contract.source = FS.readFileSync("../nf/nvm/test/bank_vault_contract.ts", "utf-8");
            }
            if (testInput.isTypeTs) {
                contract.sourceType = "ts";
            }
        }

        var toAddr = Account.NewAccount();
        if (testInput.isSameAddr === true) {
            toAddr = from;
        }
        var tx = new Transaction(ChainID, from, toAddr, Unit.nasToBasic(testInput.transferValue), parseInt(fromState.nonce) + testInput.nonceIncrement, testInput.gasPrice, testInput.gasLimit, contract);
        if (testInput.isAddressInvalid) {
            tx.from.address = cryptoUtils.toBuffer("0x23");
            tx.to.address = cryptoUtils.toBuffer("0x23");
        } else if (testInput.isAddressNull) {
            tx.from.address = cryptoUtils.bufferToHex("");
            tx.to.address = cryptoUtils.bufferToHex("");
        }
        
        if (testInput.hasOwnProperty("rewritePrice")) {
            tx.gasPrice = testInput.rewritePrice
        }

        if (testInput.hasOwnProperty("rewriteGasLimit")) {
            tx.gasLimit = testInput.rewriteGasLimit
        }
        
        tx.signTransaction();
        if (testInput.isSignErr) {
            tx.sign = "wrong signature";
        } else if (testInput.isSignNull) {
            tx.sign = '';
        } else if (testInput.isSignFake) {
            //repalce the privkey to sign
            console.log("this is the right signature:" + tx.sign.toString('hex'));
            console.log("repalce the privkey and sign a fake signatrue...")
            var newAccount = new Wallet.Account("a6e5eb222e4538fce79f5cb8774a72621637c2c9654c8b2525ed1d7e4e73653f");
            var privKey = tx.from.privKey
            tx.from.privKey = newAccount.privKey
            tx.signTransaction();
            console.log("now signatrue is: " + tx.sign.toString('hex'));
            tx.from.privKey = privKey;
        }
        //print dataLen to calculate gas 
        console.log("tx.data.payload.length: " + tx.data.payload.length);
        return neb.api.sendRawTransaction(tx.toProtoString());
    }).catch(function (err) {
        //console.log("--------------------", err);
        if (true === testExpect.canSendTx) {
            console.log(JSON.stringify(err))
            done(err);
        } else {
            console.log("sendRawTx failed reason:" + JSON.stringify(err.error));
            if (testExpect.hasOwnProperty("errMsg")) {
                //expect(testExpect.errMsg).to.be.equal(err.error.error);
                expect(testExpect.errMsg).to.be.equal(err.error.error);
            }
            done();
        }
    }).then(function (resp) {

        if (true === testExpect.canSendTx) {
            //console.log("send Rax Tx:" + resp);
            expect(resp).to.be.have.property('txhash');
            expect(resp).to.be.have.property('contract_address');
            var toAddr = resp.contract_address;

            checkTransaction(resp.txhash, function (receipt) {

                try {
                    if (true === testExpect.canSubmitTx) {
                        expect(receipt).to.not.be.a('undefined');
                        if (true === testExpect.canExcuteTx) {
                            console.log(receipt);
                            expect(receipt).have.property('status').equal(1);
                        } else {
                            expect(receipt).have.property('status').equal(0);
                        }
                        neb.api.getAccountState(receipt.from).then(function (state) {

                            console.log("get from account state :" + JSON.stringify(state));
                            expect(state.balance).to.equal(testExpect.fromBalanceAfterTx);
                            return neb.api.getAccountState(toAddr);
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
                        }).then(function (events) {
                            //console.log("[eventCheck] events[]: " + JSON.stringify(eventResult.events,null,'\t'));

                            for (var i = 0; i < events.events.length; i++) {
                                var event = events.events[i];
                                //console.log("tx event:", JSON.stringify(event,null,'\t'));
                                console.log("tx event:", event.data);
                                if (event.topic === "chain.transactionResult") {
                                    var result = JSON.parse(event.data);
                                    expect(result.status).to.equal(testExpect.status);

                                    if (testExpect.hasOwnProperty("eventError")){
                                        console.log("Event error checked.");
                                        expect(result.error).to.equal(testExpect.eventError);
                                    }
                                }
                            }
                            done();
                        }).catch(function (err) {

                            console.log(JSON.stringify(err));
                            done(err);
                        });
                    } else {
                        console.log("transaction can send but submit failed");
                        expect(receipt).to.be.a('undefined');
                        done();
                    }
                } catch (err) {
                    console.log(JSON.stringify(err));
                    done(err);
                }
            });
        } else {
            expect(resp).to.be.a('undefined');
        }
    }).catch(function (err) {

        console.log(JSON.stringify(err.error));
        done(err);
    });
}

function prepare(done) {
    from = Account.NewAccount();
    console.log("source address: " + sourceAccount.getAddressString());
    neb.api.getAccountState(sourceAccount.getAddressString()).then(function (resp) {

        console.log("source state:" + JSON.stringify(resp));
        var tx = new Transaction(ChainID, sourceAccount, from, Unit.nasToBasic(initFromBalance), parseInt(resp.nonce) + 1);
        tx.signTransaction();
        //console.log("source tx:" + tx.toString());
        return neb.api.sendRawTransaction(tx.toProtoString());
    }).then(function (resp) {
        console.log("sendRawTx resp" + JSON.stringify(resp))

        checkTransaction(resp.txhash, function (resp) {
            try {
                console.log("complete from address claim.");
                expect(resp).to.be.have.property('status').equal(1);
                done();
            } catch (err) {
                done(err);
            }
        });
    }).catch(function (err) {
        console.log("claim token failed:" + JSON.stringify(err));
        done(err);
    });
}

describe('contract deploy', function () {
 
    it('111 normal deploy', function (done) {

        var testInput = {
            transferValue: 1,
            isSameAddr: true,
            canInit: true,
            gasLimit: 2000000,
            gasPrice: -1,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: true,
            canSubmitTx: true,
            canExcuteTx: true,
            status: 1,
            fromBalanceAfterTx: '9999999977563000000',
            toBalanceAfterTx: '0',
            transferReward: '22437000000'
        };
        prepare((err) => {
            if (err) {
                done(err);
            } else {
                testContractDeploy(testInput, testExpect, done);
            }
        });
    });

    it('from & to is different', function (done) {

        var testInput = {
            transferValue: 1,
            isSameAddr: false,
            canInit: true,
            gasLimit: 2000000,
            gasPrice: -1,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: false,
            status: 0,
            fromBalanceAfterTx: '',
            toBalanceAfterTx: '',
            transferReward: ''
        };
        prepare((err) => {
            if (err) {
                done(err);
            } else {
                testContractDeploy(testInput, testExpect, done);
            }
        });
    });

    it('[address invalid] address invalid', function (done) {

        var testInput = {
            transferValue: 1,
            isSameAddr: true,
            isAddressInvalid: true,
            gasLimit: 2000000,
            gasPrice: -1,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: false,
            status: 0,
            fromBalanceAfterTx: '-1',
            toBalanceAfterTx: '-1',
            transferReward: '-1'
        };
        prepare((err) => {
            if (err) {
                done(err);
            } else {
                testContractDeploy(testInput, testExpect, done);
            }
        });
    });

    it('address is null', function (done) {

        var testInput = {
            transferValue: 1,
            isSameAddr: true,
            isAddressNull: true,
            gasLimit: 2000000,
            gasPrice: -1,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: false,
            status: 0,
            fromBalanceAfterTx: '-1',
            toBalanceAfterTx: '-1',
            transferReward: '-1'
        };
        prepare((err) => {
            if (err) {
                done(err);
            } else {
                testContractDeploy(testInput, testExpect, done);
            }
        });
    });

    it('signature invalid', function (done) {

        var testInput = {
            transferValue: 1,
            isSameAddr: true,
            isSignErr: true,
            gasLimit: 2000000,
            gasPrice: -1,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: false,
            status: 0,
            fromBalanceAfterTx: '',
            toBalanceAfterTx: '',
            transferReward: ''
        };
        prepare((err) => {
            if (err) {
                done(err);
            } else {
                testContractDeploy(testInput, testExpect, done);
            }
        });
    });

    it('signature is null', function (done) {

        var testInput = {
            transferValue: 1,
            isSameAddr: true,
            isSignNull: true,
            gasLimit: 2000000,
            gasPrice: -1,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: false,
            status: 0,
            fromBalanceAfterTx: '',
            toBalanceAfterTx: '',
            transferReward: ''
        };
        prepare((err) => {
            if (err) {
                done(err);
            } else {
                testContractDeploy(testInput, testExpect, done);
            }
        });
    });

    it('signature is fake', function (done) {

        var testInput = {
            transferValue: 1,
            isSameAddr: true,
            isSignFake: true,
            gasLimit: 2000000,
            gasPrice: -1,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: false,
            status: 0,
            fromBalanceAfterTx: '',
            toBalanceAfterTx: '',
            transferReward: ''
        };
        prepare((err) => {
            if (err) {
                done(err);
            } else {
                testContractDeploy(testInput, testExpect, done);
            }
        });
    });

    it('[balance sufficient] sssbalanceOfFrom = (TxBaseGasCount + TxPayloadBaseGasCount[payloadType] + ' +
        'gasCountOfPayloadExecuted) * gasPrice + valueOfTx', function (done) {

            var testInput = {
                transferValue: 9.999999977563000000,
                isSameAddr: true,
                canInit: true,
                gasLimit: 22437,
                gasPrice: -1,
                nonceIncrement: 1
            };
            //can calc value by previous params
            var testExpect = {
                canSendTx: true,
                canSubmitTx: true,
                canExcuteTx: true,
                status: 1,
                fromBalanceAfterTx: '9999999977563000000',
                toBalanceAfterTx: '0',
                transferReward: '22437000000'
            };
            prepare((err) => {
                if (err) {
                    done(err);
                } else {
                    testContractDeploy(testInput, testExpect, done);
                }
            });
        });


    it('[balance insufficient] balanceOfFrom < gasPrice*gasLimit', function (done) { //todo: check result

        var testInput = {
            transferValue: 1,
            isSameAddr: true,
            gasLimit: 50000000000,
            gasPrice: 21000000000,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: true,
            canSubmitTx: false,
            canExcuteTx: false,
            status: 0,
            fromBalanceAfterTx: '',
            toBalanceAfterTx: '',
            transferReward: ''
        };
        prepare((err) => {
            if (err) {
                done(err);
            } else {
                testContractDeploy(testInput, testExpect, done);
            }
        });
    });

    it('[balance insufficient] balanceOfFrom < (TxBaseGasCount + TxPayloadBaseGasCount[payloadType] +'
         + 'gasCountOfPayloadExecuted) * gasPrice + valueOfTx', function (done) {

            var testInput = {
                transferValue: 9.999999999999,
                isSameAddr: true,
                gasLimit: 2000000,
                gasPrice: -1,
                nonceIncrement: 1
            };
            //can calc value by previous params
            var testExpect = {
                canSendTx: true,
                canSubmitTx: true,
                canExcuteTx: false,
                status: 0,
                fromBalanceAfterTx: '9999999977692000000',
                toBalanceAfterTx: '0',
                transferReward: '22308000000',
                eventError: 'insufficient balance'
            };
            prepare((err) => {
                if (err) {
                    done(err);
                } else {
                    testContractDeploy(testInput, testExpect, done);
                }
            });
        });

    //No need to verfiy the case when balance is below than (TxBaseGasCount + TxPayloadBaseGasCount[payloadType]
    // + gasCountOfPayloadExecuted * gasPrice, becuase we have case verfiy that balance must  be larger than gasPrice
    // * gasLimit, and case verify what happen when gasLimit is sufficient. 

    

    it('[gas price insufficient] gas price = 0', function (done) {

        var testInput = {
            transferValue: 1,
            isSameAddr: true,
            gasLimit: 2000000,
            gasPrice: 0,
            rewritePrice: 0,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: false,
            status: 0,
            fromBalanceAfterTx: '',
            toBalanceAfterTx: '',
            transferReward: ''
        };
        prepare((err) => {
            if (err) {
                done(err);
            } else {
                testContractDeploy(testInput, testExpect, done);
            }
        });
    });

    it('[gas price insufficient] gas price < txpool.gasPrice', function (done) {

        var testInput = {
            transferValue: 1,
            isSameAddr: true,
            canInit: true,
            gasLimit: 2000000,
            gasPrice: 1,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: false,
            status: 0,
            fromBalanceAfterTx: '',
            toBalanceAfterTx: '',
            transferReward: ''
        };
        prepare((err) => {
            if (err) {
                done(err);
            } else {
                testContractDeploy(testInput, testExpect, done);
            }
        });
    });

    it('[gas price sufficient] gas price = txpool.gasPrice', function (done) {

        var testInput = {
            transferValue: 1,
            isSameAddr: true,
            gasLimit: 2000000,
            gasPrice: 1000000,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: true,
            canSubmitTx: true,
            canExcuteTx: true,
            status: 1,
            fromBalanceAfterTx: '9999999977563000000',
            toBalanceAfterTx: '0',
            transferReward: '22437000000'
        };
        prepare((err) => {
            if (err) {
                done(err);
            } else {
                testContractDeploy(testInput, testExpect, done);
            }
        });
    });

    it('[gas price sufficient] gas price > txpool.gasPrice', function (done) {

        var testInput = {
            transferValue: 1,
            isSameAddr: true,
            gasLimit: 2000000,
            gasPrice: 2000000,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: true,
            canSubmitTx: true,
            canExcuteTx: true,
            status: 1,
            fromBalanceAfterTx: '9999999955126000000',
            toBalanceAfterTx: '0',
            transferReward: '44874000000'
        };
        prepare((err) => {
            if (err) {
                done(err);
            } else {
                testContractDeploy(testInput, testExpect, done);
            }
        });
    });

    it('[gas price sufficient] gas price > Max gas price', function (done) {

        var testInput = {
            transferValue: 1,
            isSameAddr: true,
            gasLimit: 2000000,
            gasPrice: 1000000000001,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: false,
            status: 0,
            fromBalanceAfterTx: '',
            toBalanceAfterTx: '',
            transferReward: '',
            errMsg: 'invalid gas price, should be in (0, 10^12]'
        };
        prepare((err) => {
            if (err) {
                done(err);
            } else {
                testContractDeploy(testInput, testExpect, done);
            }
        });
    });

    it('[gasLimit sufficient] gasLimit > TxBaseGasCount + gasCountOfPayload' +
        '+ gasCountOfpayloadExecuted', function (done) {

            var testInput = {
                transferValue: 1,
                isSameAddr: true,
                gasLimit: 2000000,
                gasPrice: -1,
                nonceIncrement: 1
            };
            //can calc value by previous params
            var testExpect = {
                canSendTx: true,
                canSubmitTx: true,
                canExcuteTx: true,
                status: 1,
                fromBalanceAfterTx: '9999999977563000000',
                toBalanceAfterTx: '0',
                transferReward: '22437000000'
            };
            prepare((err) => {
                if (err) {
                    done(err);
                } else {
                    testContractDeploy(testInput, testExpect, done);
                }
            });
        });

    it('[gasLimit sufficient] gasLimit = TxBaseGasCount + gasCountOfPayload' +
        '+ gasCountOfpayloadExecuted', function (done) {

            var testInput = {
                transferValue: 1,
                isSameAddr: true,
                gasLimit: 22437,
                gasPrice: -1,
                nonceIncrement: 1
            };
            //can calc value by previous params
            var testExpect = {
                canSendTx: true,
                canSubmitTx: true,
                canExcuteTx: true,
                status: 1,
                fromBalanceAfterTx: '9999999977563000000',
                toBalanceAfterTx: '0',
                transferReward: '22437000000'
            };
            prepare((err) => {
                if (err) {
                    done(err);
                } else {
                    testContractDeploy(testInput, testExpect, done);
                }
            });
        });

    it('[gasLimit insufficient] TxBaseGasCount + gasCountOfPayload < gasLimit < TxBaseGasCount' +
        ' + gasCountOfPayload + gasCountOfpayloadExecuted', function (done) {

            var testInput = {
                transferValue: 1,
                isSameAddr: true,
                gasLimit: 22309,
                gasPrice: -1,
                nonceIncrement: 1
            };
            //can calc value by previous params
            var testExpect = {
                canSendTx: true,
                canSubmitTx: true,
                canExcuteTx: false,
                status: 0,
                fromBalanceAfterTx: '9999999977691000000',
                toBalanceAfterTx: '0',
                transferReward: '22309000000',
                eventError: 'insufficient gas'
            };
            prepare((err) => {
                if (err) {
                    done(err);
                } else {
                    testContractDeploy(testInput, testExpect, done);
                }
            });
        });

    it('[gasLimit insufficient] gasLimit = TxBaseGasCount + gasCountOfPayload', function (done) {

            var testInput = {
                transferValue: 1,
                isSameAddr: true,
                gasLimit: 22308,
                gasPrice: -1,
                nonceIncrement: 1
            };  
            //can calc value by previous params
            var testExpect = {
                canSendTx: true,
                canSubmitTx: true,
                canExcuteTx: false,
                status: 0,
                fromBalanceAfterTx: '9999999977692000000',
                toBalanceAfterTx: '0',
                transferReward: '22308000000',
                eventError: 'out of gas limit'
            };
            prepare((err) => {
                if (err) {
                    done(err);
                } else {
                    testContractDeploy(testInput, testExpect, done);
                }
            });
        });

        it('[gasLimit insufficient] TxBaseGasCount < gasLimit < TxBaseGasCount + gasCountOfPayload', function (done) {

            var testInput = {
                transferValue: 1,
                isSameAddr: true,
                gasLimit: 22307,
                gasPrice: -1,
                nonceIncrement: 1
            };  
            //can calc value by previous params
            var testExpect = {
                canSendTx: true,
                canSubmitTx: true,
                canExcuteTx: false,
                status: 0,
                fromBalanceAfterTx: '9999999977693000000',
                toBalanceAfterTx: '0',
                transferReward: '22307000000',
                eventError: 'out of gas limit'
            };
            prepare((err) => {
                if (err) {
                    done(err);
                } else {
                    testContractDeploy(testInput, testExpect, done);
                }
            });
        });

        it('[gasLimit insufficient] gasLimit = TxBaseGasCount ', function (done) {

            var testInput = {
                transferValue: 1,
                isSameAddr: true,
                gasLimit: 22248,
                gasPrice: -1,
                nonceIncrement: 1
            };  
            //can calc value by previous params
            var testExpect = {
                canSendTx: true,
                canSubmitTx: true,
                canExcuteTx: false,
                status: 0,
                fromBalanceAfterTx: '9999999977752000000',
                toBalanceAfterTx: '0',
                transferReward: '22248000000',
                eventError: 'out of gas limit'
            };
            prepare((err) => {
                if (err) {
                    done(err);
                } else {
                    testContractDeploy(testInput, testExpect, done);
                }
            });
        });
        
        it('[gasLimit insufficient] gasLimit < TxBaseGasCount', function (done) {
    
            var testInput = {
                transferValue: 1,
                isSameAddr: true,
                gasLimit: 22247,
                gasPrice: -1,
                nonceIncrement: 1
            };
            //can calc value by previous params
            var testExpect = {
                canSendTx: true,
                canSubmitTx: false,
                canExcuteTx: false,
                status: 0,
                fromBalanceAfterTx: '',
                toBalanceAfterTx: '',
                transferReward: ''
            };
            prepare((err) => {
                if (err) {
                    done(err);
                } else {
                    testContractDeploy(testInput, testExpect, done);
                }
            });
        });

    it('[gasLimit insufficient] gasLimit = 0', function (done) {

         var testInput = {
            transferValue: 1,
            isSameAddr: true,
            gasLimit: -1,
            gasPrice: -1,
            nonceIncrement: 1,
            rewriteGasLimit: 0
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: false,
            status: 0,
            fromBalanceAfterTx: '',
            toBalanceAfterTx: '',
            transferReward: '',
            errMsg: 'invalid gas limit, should be in (0, 5*10^10]'
        };
        prepare((err) => {
            if (err) {
                done(err);
            } else {
                testContractDeploy(testInput, testExpect, done);
            }
        });
    });

    it('gasLimit out of max ', function (done) {

        var testInput = {
            transferValue: 1,
            isSameAddr: true,
            gasLimit: 1000000000000000000,
            gasPrice: -1,
            nonceIncrement: 1
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: false,
            status: 0,
            fromBalanceAfterTx: '',
            toBalanceAfterTx: '',
            transferReward: '',
            errMsg: 'invalid gas limit, should be in (0, 5*10^10]'
        };
        prepare((err) => {
            if (err) {
                done(err);
            } else {
                testContractDeploy(testInput, testExpect, done);
            }
        });
    });

    it('nonce is below', function (done) {

        var testInput = {
            transferValue: 1,
            isSameAddr: true,
            gasLimit: 2000000,
            gasPrice: -1,
            nonceIncrement: 0
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: false,
            canSubmitTx: false,
            canExcuteTx: false,
            status: 0,
            fromBalanceAfterTx: '-1',
            toBalanceAfterTx: '-1',
            transferReward: '-1',
            errMsg: 'transaction\'s nonce is invalid, should bigger than the from\'s nonce'
        };
        prepare((err) => {
            if (err) {
                done(err);
            } else {
                testContractDeploy(testInput, testExpect, done);
            }
        });
    });

    it('nonce is bigger', function (done) {

        var testInput = {
            transferValue: 1,
            isSameAddr: true,
            gasLimit: 2000000,
            gasPrice: -1,
            nonceIncrement: 2
        };
        //can calc value by previous params
        var testExpect = {
            canSendTx: true,
            canSubmitTx: false,
            canExcuteTx: false,
            status: 0,
            fromBalanceAfterTx: '',
            toBalanceAfterTx: '',
            transferReward: ''
        };
        prepare((err) => {
            if (err) {
                done(err);
            } else {
                testContractDeploy(testInput, testExpect, done);
            }
        });
    });

    it('contract fail to init ', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: true,
            canNotInit: true,
            gasLimit: 2000000,
            gasPrice: -1,
            nonceIncrement: 1
        };

        var testExpect = {
            canSendTx: true,
            canSubmitTx: true,
            canExcuteTx: false,
            status: 0,
            fromBalanceAfterTx: '9999999977526000000',
            toBalanceAfterTx: '0',
            transferReward: '22474000000',
            eventError: 'Deploy: fail to init'
        };
        prepare((err) => {
            if (err) {
                done(err);
            } else {
                testContractDeploy(testInput, testExpect, done);
            }
        });
    });

    it('source code syntax error ', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: true,
            isSyntaxErr: true,
            gasLimit: 2000000,
            gasPrice: -1,
            nonceIncrement: 1
        };

        var testExpect = {
            canSendTx: true,
            canSubmitTx: true,
            canExcuteTx: false,
            status: 0,
            fromBalanceAfterTx: '9999999977694000000',
            toBalanceAfterTx: '0',
            transferReward: '22306000000',
            eventError: 'inject tracing instructions failed'
        };
        prepare((err) => {
            if (err) {
                done(err);
            } else {
                testContractDeploy(testInput, testExpect, done);
            }
        });
    });

    it('source is normal js file but not contract', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: true,
            isNormalJs: true,
            gasLimit: 2000000,
            gasPrice: -1,
            nonceIncrement: 1
        };

        var testExpect = {
            canSendTx: true,
            canSubmitTx: true,
            canExcuteTx: false,
            status: 0,
            fromBalanceAfterTx: '9999999979792000000',
            toBalanceAfterTx: '0',
            transferReward: '20208000000'
        };
        prepare((err) => {
            if (err) {
                done(err);
            } else {
                testContractDeploy(testInput, testExpect, done);
            }
        });
    });

    it('no init function in contract', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: true,
            NoInitFunc: true,
            gasLimit: 2000000,
            gasPrice: -1,
            nonceIncrement: 1
        };

        var testExpect = {
            canSendTx: true,
            canSubmitTx: true,
            canExcuteTx: false,
            status: 0,
            fromBalanceAfterTx: '9999999977295000000',
            toBalanceAfterTx: '0',
            transferReward: '22705000000'
        };
        prepare((err) => {
            if (err) {
                done(err);
            } else {
                testContractDeploy(testInput, testExpect, done);
            }
        });
    });

    it('contract args error', function (done) {//TODO: modify erc20
        var testInput = {
            transferValue: 1,
            isSameAddr: true,
            isParamErr: true,
            gasLimit: 2000000,
            gasPrice: -1,
            nonceIncrement: 1
        };

        var testExpect = {
            canSendTx: true,
            canSubmitTx: true,
            canExcuteTx: true,
            status: 1,
            fromBalanceAfterTx: '9999999977566000000',
            toBalanceAfterTx: '0',
            transferReward: '22434000000'
        };
        prepare((err) => {
            if (err) {
                done(err);
            } else {
                testContractDeploy(testInput, testExpect, done);
            }
        });
    });

    it('source type is wrong', function (done) { //todo: => canSubmitTx: false, (file: ".c")
        var testInput = {
            transferValue: 1,
            isSameAddr: true,
            isTypeWrong: true,
            gasLimit: 2000000,
            gasPrice: -1,
            nonceIncrement: 1
        };

        var testExpect = {
            canSendTx: true,
            canSubmitTx: true,
            canExcuteTx: false,
            status: 0,
            fromBalanceAfterTx: '9999999977753000000',
            toBalanceAfterTx: '0',
            transferReward: '22247000000'
        };
        prepare((err) => {
            if (err) {
                done(err);
            } else {
                testContractDeploy(testInput, testExpect, done);
            }
        });
    });

    it('source type is empty', function (done) { //todo: => canSubmitTx: false,

        var testInput = {
            transferValue: 1,
            isSameAddr: true,
            isTypeEmpty: true,
            gasLimit: 2000000,
            gasPrice: -1,
            nonceIncrement: 1
        };

        var testExpect = {
            canSendTx: true,
            canSubmitTx: true,
            canExcuteTx: false,
            status: 0,
            fromBalanceAfterTx: '9999999977754000000',
            toBalanceAfterTx: '0',
            transferReward: '22246000000'
        };
        prepare((err) => {
            if (err) {
                done(err);
            } else {
                testContractDeploy(testInput, testExpect, done);
            }
        });
    });

    it('source is js but type is ts', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: true,
            isTypeTs: true,
            gasLimit: 2000000,
            gasPrice: -1,
            nonceIncrement: 1
        };

        var testExpect = {
            canSendTx: true,
            canSubmitTx: true,
            canExcuteTx: true,
            status: 1,
            fromBalanceAfterTx: '9999999977563000000',
            toBalanceAfterTx: '0',
            transferReward: '22437000000'
        };
        prepare((err) => {
            if (err) {
                done(err);
            } else {
                testContractDeploy(testInput, testExpect, done);
            }
        });
    });

    it('source is ts but type is js', function (done) {
        var testInput = {
            transferValue: 1,
            isSameAddr: true,
            isSourceTypeTs: true,
            gasLimit: 2000000,
            gasPrice: -1,
            nonceIncrement: 1
        };

        var testExpect = {
            canSendTx: true,
            canSubmitTx: true,
            canExcuteTx: false,
            status: 0,
            fromBalanceAfterTx: '9999999976573000000',
            toBalanceAfterTx: '0',
            transferReward: '23427000000'
        };
        prepare((err) => {
            if (err) {
                done(err);
            } else {
                testContractDeploy(testInput, testExpect, done);
            }
        });
    });

});
