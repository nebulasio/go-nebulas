'use strict';

var Wallet = require('../../../cmd/console/neb.js/lib/wallet.js');
var sleep = require("system-sleep");
var HttpRequest = require("../../node-request");
var FS = require("fs");

var expect = require('chai').expect;
var BigNumber = require('bignumber.js');

var Neb = Wallet.Neb;
var neb = new Neb();

var ChainID;
var source, deploy, from, fromState, contract;

var coinbase, coinState;
var testCases = new Array();

//local
var env = "local";
if (env == 'local'){
	neb.setRequest(new HttpRequest("http://127.0.0.1:8685"));//https://testnet.nebulas.io
	ChainID = 100;
    source = new Wallet.Account("a6e5eb290e1438fce79f5cb8774a72621637c2c9654c8b2525ed1d7e4e73653f");
    coinbase = "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8";
}else if(env == 'testneb1'){
	neb.setRequest(new HttpRequest("http://35.182.48.19:8685"));
	ChainID = 1001;
    source = new Wallet.Account("43181d58178263837a9a6b08f06379a348a5b362bfab3631ac78d2ac771c5df3");
    coinbase = "0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d";
}else if(env == "testneb2"){
	neb.setRequest(new HttpRequest("http://34.205.26.12:8685"));
	ChainID = 1002;
    source = new Wallet.Account("43181d58178263837a9a6b08f06379a348a5b362bfab3631ac78d2ac771c5df3");
    coinbase = "0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d";
}else{
	console.log("please input correct env local testneb1 testneb2");
	return;
}

var lastnonce = 0;

function prepareContractCall(done) {
    neb.api.getAccountState(source.getAddressString()).then(function (resp) {
        console.log("source account state:" + JSON.stringify(resp));
        lastnonce = parseInt(resp.nonce);

        var accounts = new Array();
        if (typeof contract === "undefined") {
            deploy = Wallet.Account.NewAccount();
            accounts.push(deploy);
        }

        from = Wallet.Account.NewAccount();
        accounts.push(from);

        cliamTokens(accounts, neb.nasToBasic(1), function () {
            if (typeof contract === "undefined") {
                deployContract(done);
            } else {
                done();
            }
        });

    });
}

function testContractCall(testInput, testExpect, done) {
    var fromAcc = (typeof testInput.from === "undefined") ? from : testInput.from;
    var to = (typeof testInput.to === "undefined") ? contract : testInput.to;

    neb.api.getAccountState(from.getAddressString()).then(function (resp) {
        fromState = resp;
        console.log("from state:", JSON.stringify(resp));
        return neb.api.getAccountState(coinbase);
    }).catch (function (err) {
        done(err);
    }).then(function (resp) {
        console.log("coin state:", JSON.stringify(resp));
        coinState = resp;

        var tx = new Wallet.Transaction(ChainID, fromAcc, to, testInput.value, parseInt(fromState.nonce) + testInput.nonce, testInput.gasPrice, testInput.gasLimit, testInput.contract);
        if (testInput.sign) {
            tx.signTransaction();
        }
        // console.log("tx raw:", tx.toString());
        return neb.api.sendRawTransaction(tx.toProtoString());
    }).then(function (rawResp) {
        if (true === testExpect.canSendTx) {
            console.log("send Rax Tx:" + JSON.stringify(rawResp));
            expect(rawResp).to.be.have.property('txhash');
            checkTransaction(rawResp.txhash, function (receipt) {

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
                            return neb.api.getAccountState(contract);
                        }).then(function (state) {

                            console.log("get contractAddr account state :" + JSON.stringify(state));
                            expect(state.balance).to.equal(testExpect.toBalanceAfterTx);
                            return neb.api.getAccountState(coinbase);
                        }).then(function (state) {

                            console.log("get coinbase account state before tx:" + JSON.stringify(coinState));
                            console.log("get coinbase account state after tx:" + JSON.stringify(state));
                            var reward = new BigNumber(state.balance).sub(coinState.balance);
                            reward = reward.mod(new BigNumber(0.48).mul(new BigNumber(10).pow(18)));
                            // The transaction should be only
                            expect(reward.toString()).to.equal(testExpect.transferReward);
                            if (receipt.gasUsed) {
                                var txCost = new BigNumber(receipt.gasUsed).mul(receipt.gasPrice).toString(10);
                                expect(txCost).to.equal(testExpect.transferReward);
                            }

                            return neb.api.getEventsByHash(receipt.hash);
                        }).then(function (events) {
                            for (var i = 0; i < events.length; i++) {
                                var event = events[i];
                                if (event.topic == "chain.transactionResult") {
                                    var result = JSON.parse(event.data);
                                    expect(result.status).to.equal(testExpect.status);
                                }
                            }
                            done();
                        }).catch(function (err) {
                            console.log("exe tx err:", err);
                            done(err);
                        });
                    } else {
                        expect(receipt).to.be.a('undefined');
                        done();
                    }
                } catch (err) {
                    console.log("submit tx err:", err);
                    done(err);
                }
            });
        } else {
            console.log("send tx:", rawResp);
            done(new Error("send tx should failed"));
        }
    }).catch(function (err) {
        if (true === testExpect.canSendTx) {
            done(err);
        } else {
            console.log("send tx err:", err);
            done();
        }
    });

}

function cliamTokens(accounts, value, done) {
    for (var i = 0; i < accounts.length; i++) {
        sendTransaction(source, accounts[i], value, ++lastnonce);
        sleep(30);
    }
    checkCliamTokens(done);
}

function sendTransaction(from, address, value, nonce) {
    var transaction = new Wallet.Transaction(ChainID, from, address, value, nonce);
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

function deployContract(done){

    // create contract
    var source = FS.readFileSync("../nf/nvm/test/NRC20.js", "utf-8");
    var contract = {
        "source": source,
        "sourceType": "js",
        "args": "[\"StandardToken\", \"NRC\", 18, \"1000000000\"]"
    };

    var transaction = new Wallet.Transaction(ChainID, deploy, deploy, "0", 1, "10000000", "20000000000", contract);
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

    // contract status and get contract_address 
    var interval = setInterval(function () {
		// console.log("getTransactionReceipt hash:"+txhash);
		neb.api.getTransactionReceipt(txhash).then(function (resp) {

			console.log("check transaction status:" + resp.status);
			
			if(resp.status && resp.status === 1) {
                clearInterval(interval);

                if (resp.contract_address) {
                    console.log("deploy private key:" + deploy.getPrivateKeyString());
                    console.log("deploy address:" + deploy.getAddressString());
                    console.log("deploy contract address:" + resp.contract_address);
                    // console.log("deploy receipt:" + JSON.stringify(resp));

                    contract = resp.contract_address;

                    // checkNRCBalance(resp.from, resp.contract_address);
                }

                done(resp);
			}
		}).catch(function (err) {
			retry++;
			console.log("check transaction retry:", retry);
			if (retry > 10) {
				console.log(JSON.stringify(err.error));
				clearInterval(interval);
                done(err);
			}
		});

    }, 2000);
}

function checkNRCBalance(address, contractAddress) {
    var contract = {
        "function": "balanceOf",
        "args": "[\"" + address + "\"]"
    };

    neb.api.call(address, contractAddress, "0", 2, "0", "0", contract).then(function (resp) {
        console.log("balance of NRC:" + JSON.stringify(resp));
    });
}

var testCase = {
    "name": "normal call",
    "testInput": {
        sign: true,
        from: from,
        to: contract,
        value: "1",
        nonce: 1,
        gasPrice: 1000000,
        gasLimit: 2000000,
        contract: {
            "function": "name",
            "args": ""
        }
    },
    "testExpect": {
        canSendTx: true,
        canSubmitTx: true,
        canExcuteTx: true,
        status: 1,
        fromBalanceAfterTx: '999999979873999999',
        toBalanceAfterTx: '1',
        transferReward: '20126000000'
    }
};
testCases.push(testCase);


var invalidFrom = Wallet.Account.NewAccount();
invalidFrom.address = Wallet.CryptoUtils.toBuffer("12af");
testCase = {
    "name": "from address invalid",
    "testInput": {
        sign: true,
        from: invalidFrom,
        to: contract,
        value: "1",
        nonce: 1,
        gasPrice: 1000000,
        gasLimit: 2000000,
        contract: {
            "function": "name",
            "args": ""
        }
    },
    "testExpect": {
        canSendTx: false,
        canSubmitTx: false,
        canExcuteTx: false,
        status: 0,
        fromBalanceAfterTx: neb.nasToBasic(1),
        toBalanceAfterTx: '0',
        transferReward: '0'
    }
};
testCases.push(testCase);

var invalidTo = Wallet.Account.NewAccount();
invalidTo.address = Wallet.CryptoUtils.toBuffer("12af");
testCase = {
    "name": "to address invalid",
    "testInput": {
        sign: true,
        from: from,
        to: invalidTo,
        value: "1",
        nonce: 1,
        gasPrice: 1000000,
        gasLimit: 2000000,
        contract: {
            "function": "name",
            "args": ""
        }
    },
    "testExpect": {
        canSendTx: false,
        canSubmitTx: false,
        canExcuteTx: false,
        status: 0,
        fromBalanceAfterTx: neb.nasToBasic(1),
        toBalanceAfterTx: '0',
        transferReward: '0'
    }
};
testCases.push(testCase);

testCase = {
    "name": "not contract address",
    "testInput": {
        sign: true,
        from: from,
        to: Wallet.Account.NewAccount(),
        value: "1",
        nonce: 1,
        gasPrice: 1000000,
        gasLimit: 2000000,
        contract: {
            "function": "name",
            "args": ""
        }
    },
    "testExpect": {
        canSendTx: false,
        canSubmitTx: false,
        canExcuteTx: false,
        status: 0,
        fromBalanceAfterTx: neb.nasToBasic(1),
        toBalanceAfterTx: '0',
        transferReward: '0'
    }
};
testCases.push(testCase);

testCase = {
    "name": "signature invalid",
    "testInput": {
        sign: false,
        from: from,
        to: contract,
        value: "1",
        nonce: 1,
        gasPrice: 1000000,
        gasLimit: 2000000,
        contract: {
            "function": "name",
            "args": ""
        }
    },
    "testExpect": {
        canSendTx: false,
        canSubmitTx: false,
        canExcuteTx: false,
        status: 0,
        fromBalanceAfterTx: neb.nasToBasic(1),
        toBalanceAfterTx: '0',
        transferReward: '0'
    }
};
testCases.push(testCase);

testCase = {
    "name": "nonce < from.nonce + 1",
    "testInput": {
        sign: true,
        from: from,
        to: contract,
        value: "1",
        nonce: 0,
        gasPrice: 1000000,
        gasLimit: 2000000,
        contract: {
            "function": "name",
            "args": ""
        }
    },
    "testExpect": {
        canSendTx: false,
        canSubmitTx: false,
        canExcuteTx: false,
        status: 0,
        fromBalanceAfterTx: neb.nasToBasic(1),
        toBalanceAfterTx: '0',
        transferReward: '0'
    }
};
testCases.push(testCase);

testCase = {
    "name": "nonce = from.nonce + 1",
    "testInput": {
        sign: true,
        from: from,
        to: contract,
        value: "1",
        nonce: 1,
        gasPrice: 1000000,
        gasLimit: 2000000,
        contract: {
            "function": "name",
            "args": ""
        }
    },
    "testExpect": {
        canSendTx: true,
        canSubmitTx: true,
        canExcuteTx: true,
        status: 1,
        fromBalanceAfterTx: "999999979873999999",
        toBalanceAfterTx: '1',
        transferReward: '20126000000'
    }
};
testCases.push(testCase);

testCase = {
    "name": "nonce > from.nonce + 1",
    "testInput": {
        sign: true,
        from: from,
        to: contract,
        value: "1",
        nonce: 2,
        gasPrice: 1000000,
        gasLimit: 2000000,
        contract: {
            "function": "name",
            "args": ""
        }
    },
    "testExpect": {
        canSendTx: true,
        canSubmitTx: false,
        canExcuteTx: false,
        status: 2,
        fromBalanceAfterTx: neb.nasToBasic(1),
        toBalanceAfterTx: '0',
        transferReward: '0'
    }
};
testCases.push(testCase);

testCase = {
    "name": "gasPrice = 0",
    "testInput": {
        sign: true,
        from: from,
        to: contract,
        value: "1",
        nonce: 1,
        gasPrice: 0,
        gasLimit: 2000000,
        contract: {
            "function": "name",
            "args": ""
        }
    },
    "testExpect": {
        canSendTx: true,
        canSubmitTx: true,
        canExcuteTx: true,
        status: 1,
        fromBalanceAfterTx: "999999979873999999",
        toBalanceAfterTx: '1',
        transferReward: '20126000000'
    }
};
testCases.push(testCase);

testCase = {
    "name": "gasPrice > 0 && gasPrice < txpool.gasPrice",
    "testInput": {
        sign: true,
        from: from,
        to: contract,
        value: "1",
        nonce: 1,
        gasPrice: 10000,
        gasLimit: 2000000,
        contract: {
            "function": "name",
            "args": ""
        }
    },
    "testExpect": {
        canSendTx: false,
        canSubmitTx: false,
        canExcuteTx: false,
        status: 0,
        fromBalanceAfterTx: neb.nasToBasic(1),
        toBalanceAfterTx: '0',
        transferReward: '0'
    }
};
testCases.push(testCase);

testCase = {
    "name": "gasPrice = txpool.gasPrice",
    "testInput": {
        sign: true,
        from: from,
        to: contract,
        value: "1",
        nonce: 1,
        gasPrice: 1000000,
        gasLimit: 2000000,
        contract: {
            "function": "name",
            "args": ""
        }
    },
    "testExpect": {
        canSendTx: false,
        canSubmitTx: false,
        canExcuteTx: false,
        status: 0,
        fromBalanceAfterTx: "999999979873999999",
        toBalanceAfterTx: '1',
        transferReward: '20126000000'
    }
};
testCases.push(testCase);

testCase = {
    "name": "gasPrice > txpool.gasPrice",
    "testInput": {
        sign: true,
        from: from,
        to: contract,
        value: "1",
        nonce: 1,
        gasPrice: 2000000,
        gasLimit: 2000000,
        contract: {
            "function": "name",
            "args": ""
        }
    },
    "testExpect": {
        canSendTx: false,
        canSubmitTx: false,
        canExcuteTx: false,
        status: 0,
        fromBalanceAfterTx: "999999979873999999",
        toBalanceAfterTx: '1',
        transferReward: '20126000000'
    }
};
testCases.push(testCase);

testCase = {
    "name": "gasLimit = 0",
    "testInput": {
        sign: true,
        from: from,
        to: contract,
        value: "1",
        nonce: 1,
        gasPrice: 2000000,
        gasLimit: 2000000,
        contract: {
            "function": "name",
            "args": ""
        }
    },
    "testExpect": {
        canSendTx: false,
        canSubmitTx: false,
        canExcuteTx: false,
        status: 0,
        fromBalanceAfterTx: "999999979873999999",
        toBalanceAfterTx: '1',
        transferReward: '20126000000'
    }
};
testCases.push(testCase);

describe('contract call test', function () {
    beforeEach(function (done) {
        prepareContractCall(function (result) {
            if (result instanceof Error) {
                done(result);
            } else {
                done();
            }
        });
    });

    it(testCases[0].name, function (done) {
        testContractCall(testCases[0].testInput, testCases[0].testExpect, done);
    });

    it(testCases[1].name, function (done) {
        testContractCall(testCases[1].testInput, testCases[1].testExpect, done);
    });
});
