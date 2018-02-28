'use strict';

var HttpRequest = require("../../node-request");

var Wallet = require("../../../cmd/console/neb.js/lib/wallet");
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
var redeploy = process.env.REDEPLOY || false;
var scriptType = process.env.script || 'ts';
var env = process.env.NET || 'local';
if (env === 'local') {
    neb.setRequest(new HttpRequest("http://127.0.0.1:8685"));//https://testnet.nebulas.io
	ChainID = 100;
    sourceAccount = new Wallet.Account("a6e5eb290e1438fce79f5cb8774a72621637c2c9654c8b2525ed1d7e4e73653f");
    coinbase = "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8";
    if (!redeploy) {
        // txhash='f8588fabf049afa4646eb8135c0cd09f96c90fa3dd18d2bfe2816541ec81263e'
        contractAddr = "bada3c0992c3b42d8b5ecc0fb122cd0c00a27a09573ebdee";
        // ts a60c51edc4a361cddfb405a8ed9703c106321ce15b66b90c
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
var maxCheckTime = 100;
var checkTimes = 0;
var beginCheckTime;

function checkTransaction(hash, callback) {
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
            }, 2000)
        } else {
            checkTimes = 0;
            var endCheckTime = new Date().getTime();
            console.log("check tx time: : " + (endCheckTime - beginCheckTime) / 1000);
            callback(resp);
        }

    }).catch(function (err) {
        console.log("fail to get tx receipt hash: '" + hash + "' probably being packing, continue checking...")
        console.log(err.error)
        setTimeout(function(){
            checkTransaction(hash, callback)
        }, 2000)
    });
}

function deployContract(testInput, done) {
    neb.api.getAccountState(from.getAddressString()).then(function(state){
        fromState = state;
        console.log("from state: " + JSON.stringify(fromState));
    }).then(function(){
        var filepath = "./nf/nvm/test/bank_vault_contract." + testInput.contractType;
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
    coinBalanceBefore,
    gasUsed,
    gasPrice;

function testSave(testInput, testExpect, done) {
    neb.api.getAccountState(contractAddr).then(function(state){
        console.log("[before save] contract state: " + JSON.stringify(state));
        contractBalanceBefore = new BigNumber(state.balance);

        return neb.api.getAccountState(from.getAddressString());
    }).then(function(state){
        console.log("[before save] from state: " + JSON.stringify(state));
        fromBalanceBefore = state.balance;

        neb.api.getAccountState(coinbase).then(function(state){
            console.log("[before save] coinbase state:" + JSON.stringify(state));
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
            testInput.value, parseInt(state.nonce) + 1, 0, testInput.gasLimit, call).then(function(resp){
                        expect(resp).to.have.property('gas');
                        gasUsed = resp.gas;
         });

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
                console.log("[after save] contract state: " + JSON.stringify(cstate));

                neb.api.getAccountState(coinbase).then(function(state){
                    var coinbalancechange = new BigNumber(state.balance).sub(new BigNumber(coinBalanceBefore))
                        .mod(new BigNumber(0.48).mul(new BigNumber(10).pow(new BigNumber(18))));
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

                    isEqual = new BigNumber(cstate.balance).sub(contractBalanceBefore)
                                    .equals(new BigNumber(testInput.value));
                    expect(isEqual).to.be.true;
                    done();
                }).catch(function(err){
                    done(err);
                });
            }).catch(function(err){
                done(err)
            });
        });
    }).catch(function(err){
        console.log(err.error);
        done(err)
    });
}

function testTakeout(testInput, testExpect, done) {
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
            testInput.value, parseInt(state.nonce) + 1, 0, testInput.gasLimit, call).then(function(resp){
                        expect(resp).to.have.property('gas');
                        gasUsed = resp.gas;
        });

        return neb.api.sendRawTransaction(tx.toProtoString());
    }).then(function(resp){
        expect(resp).to.have.property('txhash');
        checkTransaction(resp.txhash, function(receipt){

            try {
                if (testExpect.canExecuteTx) {
                    expect(receipt).to.have.property('status').equal(1);
                } else {
                    expect(receipt).to.not.have.property('status');
                }
            }catch(err) {
                done(err)
            }
            
            neb.api.getAccountState(receipt.from).then(function(state){

                // from balance change
                fromBalanceChange = new BigNumber(state.balance).sub(new BigNumber(fromBalanceBefore));
                console.log("[after take] from state: " + JSON.stringify(state) + ", balance change: " + fromBalanceChange);

                return neb.api.getAccountState(receipt.to);
            }).then(function(cstate){
                

               return neb.api.getAccountState(coinbase).then(function(state){
                    
                    var coinbalancechange = new BigNumber(state.balance).sub(new BigNumber(coinBalanceBefore))
                        .mod(new BigNumber(0.48).mul(new BigNumber(10).pow(new BigNumber(18))));
                    console.log("[after take] coinbase state:" + JSON.stringify(state) + ", balance change: " + coinbalancechange);

                    var chg = contractBalanceBefore.sub(new BigNumber(cstate.balance));
                    console.log("[after take] contract state: " + JSON.stringify(cstate) + ", balance change: " + chg);

                    if (testExpect.canExecuteTx) {
                        var isEqual = chg.equals(new BigNumber(testExpect.takeBalance));
                        expect(isEqual).to.be.true;
                    }
    
                    return neb.api.getEventsByHash(receipt.hash);
                
                });
                
            }).then(function(evtResp){

                if (!testExpect.canExecuteTx) {
                    expect(evtResp.events[0].topic).to.equal(testExpect.eventTopic);
                    done();
                } else {
                    neb.api.gasPrice().then(function(resp){
                        expect(resp).to.have.property('gas_price');
                        console.log("[after take] gas price:" + resp['gas_price'] + ", gas used: " + gasUsed);
                        gasPrice = resp['gas_price'];
                        
                        var t = new BigNumber(testExpect.takeBalance).sub(new BigNumber(gasUsed)
                                .mul(new BigNumber(gasPrice)));
                        var isEqual = fromBalanceChange.equals(t);
                        console.log("[after take] nas cost : " + t);
    
                        expect(isEqual).to.be.true;
                        done();
                    }).catch(function(err){
                        done(err);
                    });
                }
            }).catch(function(err){
                console.log(err.error);
                done(err)
            });
        });
    }).catch(function(err){
        console.log(err.error);
        done(err)
    });
}

function claimNas(contractType, done) {
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
                    gasLimit: 2000000
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

describe('bankvault test suits', function() {
    

    describe('take-normal', function() {
        before(function(done){
            claimNas(scriptType, done);
        });

        it('save before take', function(done){
            var testInput = {
                gasLimit: 2000000,
                func: "save",
                args: "[0]",
                value: 20000000000,
            }

            var testExpect = {
                canExecuteTx: true
            }

            testSave(testInput, testExpect, done);
        });

        it('call takeout()', function(done){

            // take
            var testInput = {
                gasLimit: 2000000,
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

    describe("take-insufficient", function(){
        before(function(done){
            claimNas(scriptType, done);
        });

        it('save before take', function(done){
            var testInput = {
                gasLimit: 2000000,
                func: "save",
                args: "[0]",
                value: 20000000000,
            }
    
            var testExpect = {
                canExecuteTx: true
            }
    
            testSave(testInput, testExpect, done);
        });
    
        it('call takeout()', function(done){
            // take
            var testInput = {
                gasLimit: 2000000,
                func: "takeout",
                args: "[40000000000]",
                value: "0"  //no use
            }
    
            var testExpect = {
                canExecuteTx: true,  // actually, should be `false`
                takeBalance: '40000000000'  // same with testInput.args[0]
            }
    
            testTakeout(testInput, testExpect, done);
        });
    });

    describe("take-lt-height", function(){
        before(function(done){
            claimNas(scriptType, done);
        });

        it('save before take', function(done){
            var testInput = {
                gasLimit: 2000000,
                func: "save",
                args: "[100000]",
                value: 20000000000,
            }
    
            var testExpect = {
                canExecuteTx: true
            }
    
            testSave(testInput, testExpect, done);
        });
    
        it('call takeout()', function(done){
            // take
            var testInput = {
                gasLimit: 2000000,
                func: "takeout",
                args: "[10000000000]",
                value: "0"  //no use
            }
    
            var testExpect = {
                canExecuteTx: true,  // actually, should be `false`
                takeBalance: '10000000000'  // same with testInput.args[0]
            }
    
            testTakeout(testInput, testExpect, done);
        });
    });
    
    describe("take-nodeposite", function(){
        before(function(done){
            claimNas(scriptType, done);
        });

        it('call takeout()', function(done){
    
            // take
            var testInput = {
                gasLimit: 2000000,
                func: "takeout",
                args: "[10000000000]",
                value: "0"  //no use
            }
    
            var testExpect = {
                canExecuteTx: false , // actually, should be `false`
                takeBalance: '0',
                eventTopic: 'chain.executeTxFailed'
            }
    
            testTakeout(testInput, testExpect, done);
        });
    });

    describe("save", function(){
        before(function(done){
            claimNas(scriptType, done);
        });

        it('save non-negative', function(done){
            var testInput = {
                gasLimit: 2000000,
                func: "save",
                args: "[0]",
                value: 0//20000000000,
            }
    
            var testExpect = {
                canExecuteTx: true
            }
    
            testSave(testInput, testExpect, done);
        });

        it('save negative', function(done){
            var testInput = {
                gasLimit: 2000000,
                func: "save",
                args: "[0]",
                value: -20000000000,
            }
    
            var testExpect = {
                canExecuteTx: true
            }
    
            testSave(testInput, testExpect, done);
        });
    });
});

