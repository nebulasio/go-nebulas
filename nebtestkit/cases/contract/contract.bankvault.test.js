'use strict';

var HttpRequest = require("../../node-request");

var Wallet = require("nebulas");
var Account = Wallet.Account;
var Transaction = Wallet.Transaction;
var Utils = Wallet.Utils;
var Unit = Wallet.Unit;

var expect = require('chai').expect;
var BigNumber = require('bignumber.js');
var FS = require("fs");

var neb = new Wallet.Neb();
neb.setRequest(new HttpRequest("http://localhost:8685"));
var ChainID = 100;
var sourceAccount = new Account("a6e5eb290e1438fce79f5cb8774a72621637c2c9654c8b2525ed1d7e4e73653f");

var coinbase = "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8";
var coinState;

// deploy a new contract each run 
var redeploy = process.env.REDEPLOY || true;
var scriptType = process.env.script || 'js';
var env = process.env.NET || 'local';
var blockInterval = 15;

// mocha cases/contract/xxx testneb1 -t 200000
var args = process.argv.splice(2);
env = args[1];
if (env !== "local" && env !== "testneb1" && env !== "testneb2" && env !== "testneb3") {
    env = "local";
}
console.log("env:", env);

if (env === 'local') {
    neb.setRequest(new HttpRequest("http://127.0.0.1:8685"));//https://testnet.nebulas.io
	ChainID = 100;
    sourceAccount = new Wallet.Account("1d3fe06a53919e728315e2ccca41d4aa5b190845a79007797517e62dbc0df454");
    coinbase = "n1QZMXSZtW7BUerroSms4axNfyBGyFGkrh5";
    if (!redeploy) {
        contractAddr = "e5a02995444628cf7935e3ef8b613299a2d97e6d188ce808";  //js
        // contractAddr = "557fccaf2f05d4b46156e9b98ca9726f1c9e91d95d3830a7";  //ts
    }

} else if (env === 'testneb1') {
    neb.setRequest(new HttpRequest("http://35.182.48.19:8685"));
    ChainID = 1001;
    sourceAccount = new Wallet.Account("43181d58178263837a9a6b08f06379a348a5b362bfab3631ac78d2ac771c5df3");
    coinbase = "0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d";
    if (!redeploy) {
        contractAddr = "a506c769c3e8832f9bfaea99ba667d0ebb44a79136696045";
    }

} else if (env === "testneb2") {
    neb.setRequest(new HttpRequest("http://34.205.26.12:8685"));
    ChainID = 1002;
    sourceAccount = new Wallet.Account("25a3a441a34658e7a595a0eda222fa43ac51bd223017d17b420674fb6d0a4d52");
    coinbase = "n1SAeQRVn33bamxN4ehWUT7JGdxipwn8b17";
    if (!redeploy) {
        contractAddr = "";
    }
}else if (env === "testneb3") {
    neb.setRequest(new HttpRequest("http://35.177.214.138:8685"));
    ChainID = 1003;
    sourceAccount = new Wallet.Account("43181d58178263837a9a6b08f06379a348a5b362bfab3631ac78d2ac771c5df3");
    coinbase = "0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d";
    if (!redeploy) {
        contractAddr = "";
    }
}



var from;
var fromState;
var contractAddr;
var initFromBalance = 10;



/*
 * set this value according to the status of your testnet.
 * the smaller the value, the faster the test, with the risk of causing error
 */
var maxCheckTime = 15;
var checkTimes = 0;
var beginCheckTime;

function checkTransaction(hash, callback) {
    console.log("==>checkTransaction");
    if (checkTimes === 0) {
        beginCheckTime = new Date().getTime();
    }
    checkTimes += 1;
    if (checkTimes > maxCheckTime) {
        console.log("check tx receipt timeout: " + hash);
        checkTimes = 0;
        callback();
        return
    }

    neb.api.getTransactionReceipt(hash).then(function(resp) {
        console.log("tx receipt: " + JSON.stringify(resp))
        if (resp.status === 2) {
            setTimeout(function(){
                checkTransaction(hash, callback)
            }, 5000)
        } else {
            checkTimes = 0;
            var endCheckTime = new Date().getTime();
            console.log("check tx time: : " + (endCheckTime - beginCheckTime) / 1000);
            callback(resp);
            return
        }

    }).catch(function (err) {
        console.log("fail to get tx receipt hash: '" + hash + "' probably being packing, continue checking...")
        console.log(err);
        console.log(err.error);
        setTimeout(function(){
            checkTransaction(hash, callback)
        }, 5000)
    });
}

function deployContract(testInput, done) {
    console.log("==> deployContract");
    neb.api.getAccountState(from.getAddressString()).then(function(state){
        fromState = state;
        console.log("from state: " + JSON.stringify(fromState));
    }).then(function(){
        var filepath = "../nf/nvm/test/bank_vault_contract." + testInput.contractType;
        console.log("deploying contract: " + filepath);
        var bankvault = FS.readFileSync(filepath, "utf-8");
        var contract = {
            "source": bankvault,
            "sourceType": testInput.contractType,
            "args": ""
        }

        var tx = new Transaction(ChainID, from, from, 0, parseInt(fromState.nonce) + 1, 1000000, testInput.gasLimit, contract);
        tx.signTransaction();
        return neb.api.sendRawTransaction(tx.toProtoString());
    }).catch (function(err){
        console.log(err.error);
        done(err);
    }).then(function(resp){
        console.log("111: result", resp);
        expect(resp).to.be.have.property('txhash');
        expect(resp).to.be.have.property('contract_address');
        contractAddr = resp.contract_address;
        console.log("deploy contract return: " + JSON.stringify(resp));

        checkTransaction(resp.txhash, function(resp){
            expect(resp).to.not.be.a('undefined');
            expect(resp).to.be.have.property('status').equal(1);

            neb.api.getAccountState(contractAddr).then(function(state){
                console.log("contract: " + contractAddr + ", state: " + JSON.stringify(state));
                done();
            });
        });
    });
}
var fromBalanceBefore,
    contractBalanceBefore,
    fromBalanceChange,
    contractBalanceChange,
    coinBalanceBefore,
    gasUsed,
    gasPrice,
    txInfo = {};

function testSave(testInput, testExpect, done) {
    console.log("==> testSave");

    var call = {
        "function": testInput.func,
        "args": testInput.args
    }

    neb.api.getAccountState(contractAddr).then(function(state){
        console.log("[before save] contract state: " + JSON.stringify(state));
        contractBalanceBefore = new BigNumber(state.balance);

        return neb.api.getAccountState(from.getAddressString());
    }).then(function(state){
        console.log("[before save] from state: " + JSON.stringify(state));
        fromBalanceBefore = state.balance;
        txInfo.call = call;
        txInfo.nonce = parseInt(state.nonce) + 1;

        neb.api.getAccountState(coinbase).then(function(state){
            console.log("[before save] coinbase state:" + JSON.stringify(state));
            coinBalanceBefore = state.balance;
        });
       
        return neb.api.estimateGas(from.getAddressString(), contractAddr,
                testInput.value, txInfo.nonce, testInput.gasPrice, testInput.gasLimit, txInfo.call);

    }).then(function(resp){
        expect(resp).to.have.property('gas');
        gasUsed = resp.gas;

        var tx = new Transaction(ChainID, from, contractAddr, 
            testInput.value, txInfo.nonce, 0, testInput.gasLimit, call);
        tx.signTransaction();
        return neb.api.sendRawTransaction(tx.toProtoString());
    }).then(function(resp){
        expect(resp).to.have.property('txhash');
        checkTransaction(resp.txhash, function(receipt){

            if (testExpect.canExecuteTx) {
                expect(receipt).to.have.property('status').equal(1);
            } else {
                expect(receipt).to.not.have.property('status');
            }

            neb.api.getAccountState(receipt.from).then(function(state){

                // from balance change
                fromBalanceChange = new BigNumber(fromBalanceBefore).sub(new BigNumber(state.balance));
                console.log("[after save] from state: " + JSON.stringify(state) + ", balance change: " + fromBalanceChange);

                return neb.api.getAccountState(receipt.to);
            }).then(function(cstate){
                contractBalanceChange = new BigNumber(cstate.balance).sub(contractBalanceBefore);
                console.log("[after save] contract state: " + JSON.stringify(cstate));

                return neb.api.getAccountState(coinbase);
            }).then(function(state){
                var coinbalancechange = new BigNumber(state.balance).sub(new BigNumber(coinBalanceBefore))
                    .mod(new BigNumber(1.4269).mul(new BigNumber(10).pow(new BigNumber(18))));
                console.log("[after save] coinbase state:" + JSON.stringify(state) + ", balance change: " + coinbalancechange);
                return neb.api.gasPrice();
            }).then(function(resp){
                expect(resp).to.have.property('gas_price');
                console.log("[after save] gas price:" + resp['gas_price'] + ", gas used: " + gasUsed);
                gasPrice = resp['gas_price'];
                var isEqual = fromBalanceChange.equals(new BigNumber(gasUsed)
                                                .mul(new BigNumber(gasPrice))
                                                .add(new BigNumber(testInput.value)));     
                expect(isEqual).to.be.true;

                isEqual = contractBalanceChange.equals(new BigNumber(testInput.value));
                expect(isEqual).to.be.true; 
                done();
            }).catch(function(err){
                done(err);
            });
        });
    }).catch(function(err){
        if (testExpect.hasError) {
            try {
                expect(err.error).to.have.property('error').equal(testExpect.errorMsg);
                done();
            } catch (e) {
                done(err)
            }
        } else {
            console.log(err.error);
            done(err)
        }
    });
}

function testTakeout(testInput, testExpect, done) {
    console.log("==>testTakeout");
    neb.api.getAccountState(contractAddr).then(function(state){
        console.log("[before take] contract state: " + JSON.stringify(state));
        contractBalanceBefore = new BigNumber(state.balance);

        return neb.api.getAccountState(from.getAddressString());
    }).then(function(state){
        console.log("[before take] from state: " + JSON.stringify(state));
        fromBalanceBefore = state.balance;

        neb.api.getAccountState(coinbase).then(function(state){
            console.log("[before take] coinbase state:" + JSON.stringify(state));
            coinBalanceBefore = state.balance;
        });

        var call = {
            "function": testInput.func,
            "args": testInput.args
        }
        var tx = new Transaction(ChainID, from, contractAddr, 
            testInput.value, parseInt(state.nonce) + 1, 0, testInput.gasLimit, call);
        tx.signTransaction();

        neb.api.estimateGas(from.getAddressString(), contractAddr,
            testInput.value, parseInt(state.nonce) + 1, testInput.gasPrice, testInput.gasLimit, call).then(function(resp){
                        expect(resp).to.have.property('gas');
                        gasUsed = resp.gas;
        }).catch(function(err) {
            // do nothing
        });

        return neb.api.sendRawTransaction(tx.toProtoString());
    }).then(function(resp){
        expect(resp).to.have.property('txhash');
        checkTransaction(resp.txhash, function(receipt){
            try {
                console.log("22: receipt", receipt);
                if (testExpect.canExecuteTx) {
                    expect(receipt).to.have.property('status').equal(1);
                } else {
                    expect(receipt).to.have.property('status').equal(0);
                }

                console.log("333");
                neb.api.getAccountState(receipt.from).then(function (state) {

                    // from balance change
                    fromBalanceChange = new BigNumber(state.balance).sub(new BigNumber(fromBalanceBefore));
                    console.log("[after take] from state: " + JSON.stringify(state) + ", balance change: " + fromBalanceChange);

                    return neb.api.getAccountState(receipt.to);
                }).then(function (cstate) {

                    return neb.api.getAccountState(coinbase).then(function (state) {

                        var coinbalancechange = new BigNumber(state.balance).sub(new BigNumber(coinBalanceBefore))
                            .mod(new BigNumber(1.4269).mul(new BigNumber(10).pow(new BigNumber(18))));
                        console.log("[after take] coinbase state:" + JSON.stringify(state) + ", balance change: " + coinbalancechange);

                        var chg = contractBalanceBefore.sub(new BigNumber(cstate.balance));
                        console.log("[after take] contract state: " + JSON.stringify(cstate) + ", balance change: " + chg);

                        if (testExpect.canExecuteTx) {
                            var isEqual = chg.equals(new BigNumber(testExpect.takeBalance));
                            expect(isEqual).to.be.true;
                        }

                        return neb.api.getEventsByHash(receipt.hash);
                    });

                }).then(function (evtResp) {

                    if (!testExpect.canExecuteTx) {

                        expect(evtResp.events[0].topic).to.equal(testExpect.eventTopic);
                        done();

                    } else {
                        neb.api.gasPrice().then(function (resp) {
                            expect(resp).to.have.property('gas_price');
                            console.log("[after take] gas price:" + resp['gas_price'] + ", gas used: " + gasUsed);
                            gasPrice = resp['gas_price'];

                            var t = new BigNumber(testExpect.takeBalance).sub(new BigNumber(gasUsed)
                                .mul(new BigNumber(gasPrice)));
                            var isEqual = fromBalanceChange.equals(t);

                            expect(isEqual).to.be.true;
                            done();
                        }).catch(function (err) {
                            done(err);
                        });
                    }
                }).catch(function (err) {
                    console.log(err.error);
                    done(err)
                });
            } catch (err) {
                console.log(JSON.stringify(err));
                done(err)
            }
        })
    }).catch(function(err){
        console.log(err.error);
        if (testExpect.hasError) {
            try {
                expect(err.error).to.have.property('error').equal(testExpect.errorMsg)
                done();
            } catch (e) {
                done(err)
            }
        } else {
            done(err)
        }
    });
}

function testVerifyAddress(testInput, testExpect, done) {
    console.log("==> testVerifyAddress");
    neb.api.getAccountState(contractAddr).then(function(state){
        console.log("[before verify] contract state: " + JSON.stringify(state));
        contractBalanceBefore = new BigNumber(state.balance);

        return neb.api.getAccountState(from.getAddressString());
    }).then(function(state){
        console.log("[before verify] from state: " + JSON.stringify(state));
        fromBalanceBefore = state.balance;

        neb.api.getAccountState(coinbase).then(function(state){
            console.log("[before verify] coinbase state:" + JSON.stringify(state));
            coinBalanceBefore = state.balance;
        });

        var call = {
            "function": testInput.func,
            "args": testInput.args
        }
        return neb.api.call(from.getAddressString(), contractAddr, testInput.value, parseInt(state.nonce) + 1, testInput.gasPrice, testInput.gasLimit, call);
    }).then(resp => {
        console.log("response: " + JSON.stringify(resp));
        expect(resp).to.have.property('result')
        expect(JSON.parse(resp.result)).to.have.property('valid').equal(testExpect.valid);
        done();
    }).catch(err => done(err));
}

function claimNas(contractType, done) {
    console.log("==> claimNas");
    from = Account.NewAccount();

    console.log("from addr:" + from.getAddressString());
    console.log("source addr:" + sourceAccount.getAddressString());

    neb.api.getAccountState(sourceAccount.getAddressString()).then(function(resp){

        console.log("source state:" + JSON.stringify(resp));
        var tx = new Transaction(ChainID, sourceAccount, from, Unit.nasToBasic(initFromBalance), parseInt(resp.nonce) + 1, 0, 2000000);
        tx.signTransaction();
        return neb.api.sendRawTransaction(tx.toProtoString());
    }).then(function(resp){
        console.log("send claim tx result:" + JSON.stringify(resp));
        checkTransaction(resp.txhash, function(resp){
            expect(resp).to.not.be.a('undefined');
            expect(resp).to.be.have.property('status').equal(1);
            console.log("complete from address claim.")

            if (redeploy) {
                var testInput = {
                    contractType: contractType,
                    gasLimit: 2000000,
                    gasPrice: 1000000,
                };
                redeploy = false;
                deployContract(testInput, done);
            } else {
                done();
            }
        });
    }).catch(function(err){
        console.log("claim token failed: " + JSON.stringify(err))
        done(err)
    });
}

describe('bankvault.' + scriptType, function() {
    

    describe('1. take: normal', function() {
        before(function(done){
            claimNas(scriptType, done);
        });

        it('save before take', function(done){
            var testInput = {
                gasLimit: 2000000,
                gasPrice: 1000000,

                func: "save",
                args: "[0]",
                value: 20000000000,
            };

            var testExpect = {
                canExecuteTx: true
            };

            testSave(testInput, testExpect, done);
        });

        it('call takeout()', function(done){

            // take
            var testInput = {
                gasLimit: 2000000,
                gasPrice: 1000000,

                func: "takeout",
                args: "[10000000000]",
                value: "0"  //no use
            }

            var testExpect = {
                canExecuteTx: true,
                takeBalance: "10000000000"
            }

            testTakeout(testInput, testExpect, done);
        });
    });

    describe("2. take: insufficient balance", function(){
        before(function(done){
            claimNas(scriptType, done);
        });

        it('save before take', function(done){
            var testInput = {
                gasLimit: 2000000,
                gasPrice: 1000000,

                func: "save",
                args: "[0]",
                value: 20000000000,
            }
    
            var testExpect = {
                canExecuteTx: true
            }
    
            testSave(testInput, testExpect, done);
        });
    
        it('2-1. transfer > balance', function(done){
            // take
            var testInput = {
                gasLimit: 2000000,
                gasPrice: 1000000,
                func: "takeout",
                args: "[40000000000]",
                value: "0"  //no use
            }
    
            var testExpect = {
                canExecuteTx: false,  // actually, should be `false`
                takeBalance: '40000000000',  // same with testInput.args[0]
                // eventTopic: 'chain.executeTxFailed',
                eventTopic: 'chain.transactionResult',
                hasError: true,
                errorMsg: 'execution failed'
            }
    
            testTakeout(testInput, testExpect, done);
        });

        it('2-2. transfer = 0', function(done){
            // take
            var testInput = {
                gasLimit: 2000000,
                gasPrice: 1000000,

                func: "takeout",
                args: "[0]",
                value: "0"  //no use
            }
    
            var testExpect = {
                canExecuteTx: true,  // actually, should be `false`
                takeBalance: '0',  // same with testInput.args[0]
            }
    
            testTakeout(testInput, testExpect, done);
        });

        it('2-3. transfer < 0', function(done){
            // take
            var testInput = {
                gasLimit: 2000000,
                gasPrice: 1000000,

                func: "takeout",
                args: "[-40000000000]",
                value: "0"  //no use
            }
    
            var testExpect = {
                canExecuteTx: false,  // actually, should be `false`
                takeBalance: '-40000000000',  // same with testInput.args[0]
                // eventTopic: 'chain.executeTxFailed',
                eventTopic: 'chain.transactionResult',
                hasError: true,
                errorMsg: 'execution failed'
            }
    
            testTakeout(testInput, testExpect, done);
        });
    });

    describe("3. take: lt-height", () => {
        before(done => {
            claimNas(scriptType, done);
        });

        it('save(40) before take', done => {
            var testInput = {
                gasLimit: 2000000,
                gasPrice: 1000000,

                func: "save",
                args: "[15]",
                value: 20000000000,
            }
    
            var testExpect = {
                canExecuteTx: true
            }
    
            testSave(testInput, testExpect, done);
        });
    
        it('call takeout()', done => {
            // take
            var testInput = {
                gasLimit: 2000000,
                gasPrice: 1000000,

                func: "takeout",
                args: "[10000000000]",
                value: "0"  //no use
            }
    
            var testExpect = {
                canExecuteTx: false,
                takeBalance: '10000000000',  // same with testInput.args[0]
                // eventTopic: 'chain.executeTxFailed'
                eventTopic: 'chain.transactionResult'
            }
    
            testTakeout(testInput, testExpect, done);
        });

        it('call takeout() after 25 blocks', done => {
            // take
            var testInput = {
                gasLimit: 2000000,
                gasPrice: 1000000,

                func: "takeout",
                args: "[10000000000]",
                value: "0"  //no use
            }
    
            var testExpect = {
                canExecuteTx: true,
                takeBalance: '10000000000'  // same with testInput.args[0]
            }
            
            setTimeout(() => {
                testTakeout(testInput, testExpect, done);
            }, 25 * blockInterval * 1000);
            
        });
    });
    
    describe("4. take: no deposit", function(){
        before(function(done){
            claimNas(scriptType, done);
        });

        it('call takeout()', function(done){
    
            // take
            var testInput = {
                gasLimit: 2000000,
                gasPrice: 1000000,

                func: "takeout",
                args: "[10000000000]",
                value: "0"  //no use
            }
    
            var testExpect = {
                canExecuteTx: false , 
                takeBalance: '0',
                // eventTopic: 'chain.executeTxFailed'
                eventTopic: 'chain.transactionResult'
            }
    
            testTakeout(testInput, testExpect, done);
        });
    });

    describe("5. save", function(){
        before(function(done){
            claimNas(scriptType, done);
        });

        it('5-1. save non-negative value', function(done){
            var testInput = {
                gasLimit: 2000000,
                gasPrice: 1000000,

                func: "save",
                args: "[0]",
                value: 10000000000,
            }
    
            var testExpect = {
                canExecuteTx: true
            }
    
            testSave(testInput, testExpect, done);
        });

        it('5-2. save negative value', function(done){
            var testInput = {
                gasLimit: 2000000,
                gasPrice: 1000000,

                func: "save",
                args: "[0]",
                value: -20000000000,
            }
    
            var testExpect = {
                hasError: true,
                errorMsg: "invalid value"
            }
    
            testSave(testInput, testExpect, done);
        });

        it('5-3. save negative height', done => {
            var testInput = {
                gasLimit: 2000000,
                gasPrice: 1000000,
                func: "save",
                args: "[-500]",
                value: 20000000000,
            }
    
            var testExpect = {
                canExecuteTx: true
            }
    
            testSave(testInput, testExpect, done);
        });
    });

    describe('6. verifyAddress', () => {

        before(function(done){
            claimNas(scriptType, done);
        });

        it('6-1. legal address', done => {
            var testInput = {
                value: "0",
                gasLimit: 2000000,
                gasPrice: 1000000,
                func: "verifyAddress",
                args: "[\"n1QZMXSZtW7BUerroSms4axNfyBGyFGkrh5\"]"
            }
    
            var testExpect = {
                valid: true
            }

            testVerifyAddress(testInput, testExpect, done);
        });

        it('6-3. illegal address', done => {
            var testInput = {
                value: "0",
                gasLimit: 2000000,
                gasPrice: 1000000,
                func: "verifyAddress",
                args: "[\"slfjlsalfksdflsfjks\"]"
            }
    
            var testExpect = {
                valid: false
            }

            testVerifyAddress(testInput, testExpect, done);
        });
    });
});

