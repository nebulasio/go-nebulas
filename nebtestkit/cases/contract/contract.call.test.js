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
var source, deploy, from, fromState, contractAddr;

var coinbase, coinState;
var testCases = new Array();
var caseIndex = 0;

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

function prepareContractCall(testCase, done) {
    neb.api.getAccountState(source.getAddressString()).then(function (resp) {
        console.log("source account state:" + JSON.stringify(resp));
        lastnonce = parseInt(resp.nonce);

        var accounts = new Array();
        var values = new Array();
        // if (typeof contractAddr === "undefined") {
            deploy = Wallet.Account.NewAccount();
            accounts.push(deploy);
            values.push(neb.nasToBasic(1));
        // }

        from = Wallet.Account.NewAccount();
        accounts.push(from);

        var fromBalance = (typeof testCase.testInput.fromBalance === "undefined") ? neb.nasToBasic(1) : testCase.testInput.fromBalance;
        values.push(fromBalance);
        cliamTokens(accounts, values, function () {
            // if (typeof contractAddr === "undefined") {
                deployContract(done);
            // } else {
            //     done();
            // }
        });

    });
}

function testContractCall(testInput, testExpect, done) {
    var fromAcc = (typeof testInput.from === "undefined") ? from : testInput.from;
    var to = (typeof testInput.to === "undefined") ? Wallet.Account.fromAddress(contractAddr) : testInput.to;

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
        // test invalid address
        tx.from.address = fromAcc.address;
        tx.to.address = to.address;
        tx.gasPrice = new BigNumber(testInput.gasPrice);
        tx.gasLimit = new BigNumber(testInput.gasLimit);
        if (testInput.sign) {
            tx.signTransaction();
        }
        // console.log("tx raw:", tx.toString());
        return neb.api.sendRawTransaction(tx.toProtoString());
    }).then(function (rawResp) {
        if (true === testExpect.canSendTx) {
            console.log("send Raw Tx:" + JSON.stringify(rawResp));
            expect(rawResp).to.be.have.property('txhash');
            checkTransaction(rawResp.txhash, function (receipt) {
                console.log("tx receipt : " + JSON.stringify(receipt));
                try {
                    if (true === testExpect.canSubmitTx) {
                        expect(receipt).to.not.be.a('undefined');
                        if (true === testExpect.canExcuteTx) {
                            expect(receipt).to.be.have.property('status').equal(1);
                        } else {
                            expect(receipt.from).to.not.be.a('undefined');
                            expect(receipt).to.not.have.property('status');
                        }

                        neb.api.getAccountState(receipt.from).then(function (state) {

                            console.log("get from account state :" + JSON.stringify(state));
                            expect(state.balance).to.equal(testExpect.fromBalanceAfterTx);
                            return neb.api.getAccountState(contractAddr);
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
                                    console.log("tx event:", event.data);
                                }
                            }
                            done();
                        }).catch(function (err) {
                            console.log("exe tx err:", err);
                            done(err);
                        });
                    } else {
                        if (receipt.status) {
                            expect(receipt.status).to.equal(2);
                        }
                        console.log("transaction can send but submit failed");
                        done();
                    }
                } catch (err) {
                    console.log("submit tx err:", err.message);
                    done(err);
                }
            });
        } else {
            console.log("send tx unexpected:", rawResp);
            done(new Error("send tx should failed"));
        }
    }).catch(function (err) {
        if (true === testExpect.canSendTx) {
            done(err);
        } else {
            console.log("send tx failed:", err.message);
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
    var maxRetry = 15;

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
    "name": "1. normal call",
    "testInput": {
        sign: true,
        from: from,
        to: contractAddr,
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
    "name": "2. from address invalid",
    "testInput": {
        sign: true,
        from: invalidFrom,
        to: contractAddr,
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
    "name": "3. to address invalid",
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
    "name": "4. not contract address",
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
    "name": "5. signature invalid",
    "testInput": {
        sign: false,
        from: from,
        to: contractAddr,
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
    "name": "6. nonce < from.nonce + 1",
    "testInput": {
        sign: true,
        from: from,
        to: contractAddr,
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
    "name": "7. nonce = from.nonce + 1",
    "testInput": {
        sign: true,
        from: from,
        to: contractAddr,
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
    "name": "8. nonce > from.nonce + 1",
    "testInput": {
        sign: true,
        from: from,
        to: contractAddr,
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
    "name": "9. gasPrice = 0",
    "testInput": {
        sign: true,
        from: from,
        to: contractAddr,
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
        canSendTx: false,
        canSubmitTx: false,
        canExcuteTx: false,
        status: 1,
        fromBalanceAfterTx: neb.nasToBasic(1),
        toBalanceAfterTx: '0',
        transferReward: '0'
    }
};
testCases.push(testCase);

testCase = {
    "name": "10. gasPrice > 0 && gasPrice < txpool.gasPrice",
    "testInput": {
        sign: true,
        from: from,
        to: contractAddr,
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
    "name": "11. gasPrice = txpool.gasPrice",
    "testInput": {
        sign: true,
        from: from,
        to: contractAddr,
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
    "name": "12. gasPrice > txpool.gasPrice",
    "testInput": {
        sign: true,
        from: from,
        to: contractAddr,
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
        canSendTx: true,
        canSubmitTx: true,
        canExcuteTx: true,
        status: 1,
        fromBalanceAfterTx: "999999959747999999",
        toBalanceAfterTx: '1',
        transferReward: '40252000000'
    }
};
testCases.push(testCase);

testCase = {
    "name": "13. gasLimit = 0",
    "testInput": {
        sign: true,
        from: from,
        to: contractAddr,
        value: "1",
        nonce: 1,
        gasPrice: 1000000,
        gasLimit: 0,
        contract: {
            "function": "name",
            "args": ""
        }
    },
    "testExpect": {
        canSendTx: true,
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
    "name": "14. gasLimit < TxBaseGasCount + gasCountOfPayload",
    "testInput": {
        sign: true,
        from: from,
        to: contractAddr,
        value: "1",
        nonce: 1,
        gasPrice: 1000000,
        gasLimit: 20000,
        contract: {
            "function": "name",
            "args": ""
        }
    },
    "testExpect": {
        canSendTx: true,
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
    "name": "15. gasLimit = TxBaseGasCount + gasCountOfPayload",
    "testInput": {
        sign: true,
        from: from,
        to: contractAddr,
        value: "1",
        nonce: 1,
        gasPrice: 1000000,
        gasLimit: 20029,
        contract: {
            "function": "name",
            "args": ""
        }
    },
    "testExpect": {
        canSendTx: true,
        canSubmitTx: true,
        canExcuteTx: false,
        status: 1,
        fromBalanceAfterTx: "999999979971000000",
        toBalanceAfterTx: '0',
        transferReward: '20029000000'
    }
};
testCases.push(testCase);

testCase = {
    "name": "16. gasLimit < TxBaseGasCount + gasCountOfPayload + gasCountOfpayloadExecuted",
    "testInput": {
        sign: true,
        from: from,
        to: contractAddr,
        value: "1",
        nonce: 1,
        gasPrice: 1000000,
        gasLimit: 20100,
        contract: {
            "function": "name",
            "args": ""
        }
    },
    "testExpect": {
        canSendTx: true,
        canSubmitTx: true,
        canExcuteTx: false,
        status: 0,
        fromBalanceAfterTx: "999999979900000000",
        toBalanceAfterTx: '0',
        transferReward: '20100000000'
    }
};
testCases.push(testCase);

testCase = {
    "name": "17. gasLimit = TxBaseGasCount + gasCountOfPayload + gasCountOfpayloadExecuted",
    "testInput": {
        sign: true,
        from: from,
        to: contractAddr,
        value: "1",
        nonce: 1,
        gasPrice: 1000000,
        gasLimit: 20126,
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
    "name": "18. gasLimit > TxBaseGasCount + gasCountOfPayload + gasCountOfpayloadExecuted",
    "testInput": {
        sign: true,
        from: from,
        to: contractAddr,
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
    "name": "19. gasLimit > txpool.gasLimit",
    "testInput": {
        sign: true,
        from: from,
        to: contractAddr,
        value: "1",
        nonce: 1,
        gasPrice: 1000000,
        gasLimit: "500000000000",
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
    "name": "20. balanceOfFrom < gasPrice*gasLimit",
    "testInput": {
        fromBalance: "1",
        sign: true,
        from: from,
        to: contractAddr,
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
        canSubmitTx: false,
        canExcuteTx: false,
        status: 0,
        fromBalanceAfterTx: "1",
        toBalanceAfterTx: '0',
        transferReward: '0'
    }
};
testCases.push(testCase);

testCase = {
    "name": "21. balanceOfFrom = gasPrice*gasLimit",
    "testInput": {
        fromBalance: "20126000000",
        sign: true,
        from: from,
        to: contractAddr,
        value: "1",
        nonce: 1,
        gasPrice: 1000000,
        gasLimit: 20126,
        contract: {
            "function": "name",
            "args": ""
        }
    },
    "testExpect": {
        canSendTx: true,
        canSubmitTx: true,
        canExcuteTx: false,
        status: 0,
        fromBalanceAfterTx: "0",
        toBalanceAfterTx: '0',
        transferReward: '20126000000'
    }
};
testCases.push(testCase);

testCase = {
    "name": "22. balanceOfFrom < (TxBaseGasCount + TxPayloadBaseGasCount[payloadType] + gasCountOfPayload + gasCountOfPayloadExecuted) * gasPrice + valueOfTx",
    "testInput": {
        fromBalance: "20126100000",
        sign: true,
        from: from,
        to: contractAddr,
        value: "1000000",
        nonce: 1,
        gasPrice: 1000000,
        gasLimit: 20126,
        contract: {
            "function": "name",
            "args": ""
        }
    },
    "testExpect": {
        canSendTx: true,
        canSubmitTx: true,
        canExcuteTx: false,
        status: 0,
        fromBalanceAfterTx: "100000",
        toBalanceAfterTx: '0',
        transferReward: '20126000000'
    }
};
testCases.push(testCase);

testCase = {
    "name": "23. balanceOfFrom = (TxBaseGasCount + TxPayloadBaseGasCount[payloadType] + gasCountOfPayload + gasCountOfPayloadExecuted) * gasPrice + valueOfTx",
    "testInput": {
        fromBalance: "20127000000",
        sign: true,
        from: from,
        to: contractAddr,
        value: "1000000",
        nonce: 1,
        gasPrice: 1000000,
        gasLimit: 20126,
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
        fromBalanceAfterTx: "0",
        toBalanceAfterTx: '1000000',
        transferReward: '20126000000'
    }
};
testCases.push(testCase);

testCase = {
    "name": "24. balanceOfFrom > (TxBaseGasCount + TxPayloadBaseGasCount[payloadType] + gasCountOfPayload + gasCountOfPayloadExecuted) * gasPrice + valueOfTx",
    "testInput": {
        fromBalance: "20128000000",
        sign: true,
        from: from,
        to: contractAddr,
        value: "1000000",
        nonce: 1,
        gasPrice: 1000000,
        gasLimit: 20126,
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
        fromBalanceAfterTx: "1000000",
        toBalanceAfterTx: '1000000',
        transferReward: '20126000000'
    }
};
testCases.push(testCase);

testCase = {
    "name": "25. function not found",
    "testInput": {
        sign: true,
        from: from,
        to: contractAddr,
        value: "1",
        nonce: 1,
        gasPrice: 1000000,
        gasLimit: 2000000,
        contract: {
            "function": "functionNotFound",
            "args": ""
        }
    },
    "testExpect": {
        canSendTx: true,
        canSubmitTx: true,
        canExcuteTx: false,
        status: 0,
        fromBalanceAfterTx: "999999979866000000",
        toBalanceAfterTx: '0',
        transferReward: '20134000000'
    }
};
testCases.push(testCase);

testCase = {
    "name": "26. args more",
    "testInput": {
        sign: true,
        from: from,
        to: contractAddr,
        value: "1",
        nonce: 1,
        gasPrice: 1000000,
        gasLimit: 2000000,
        contract: {
            "function": "name",
            "args": "[1]"
        }
    },
    "testExpect": {
        canSendTx: true,
        canSubmitTx: true,
        canExcuteTx: true,
        status: 1,
        fromBalanceAfterTx: "999999979858999999",
        toBalanceAfterTx: '1',
        transferReward: '20141000000'
    }
};
testCases.push(testCase);

testCase = {
    "name": "27. args less",
    "testInput": {
        sign: true,
        from: from,
        to: contractAddr,
        value: "1",
        nonce: 1,
        gasPrice: 1000000,
        gasLimit: 2000000,
        contract: {
            "function": "balanceOf",
            "args": ""
        }
    },
    "testExpect": {
        canSendTx: true,
        canSubmitTx: true,
        canExcuteTx: true,
        status: 1,
        fromBalanceAfterTx: "999999979853999999",
        toBalanceAfterTx: '1',
        transferReward: '20146000000'
    }
};
testCases.push(testCase);

testCase = {
    "name": "28. args err",
    "testInput": {
        sign: true,
        from: from,
        to: contractAddr,
        value: "1",
        nonce: 1,
        gasPrice: 1000000,
        gasLimit: 2000000,
        contract: {
            "function": "transfer",
            "args": "[\"asda\", \"asda\"]"
        }
    },
    "testExpect": {
        canSendTx: true,
        canSubmitTx: true,
        canExcuteTx: false,
        status: 1,
        fromBalanceAfterTx: "999999979831000000",
        toBalanceAfterTx: '0',
        transferReward: '20169000000'
    }
};
testCases.push(testCase);

testCase = {
    "name": "29. execution failed",
    "testInput": {
        sign: true,
        from: from,
        to: contractAddr,
        value: "1",
        nonce: 1,
        gasPrice: 1000000,
        gasLimit: 2000000,
        contract: {
            "function": "transfer",
            "args": "[\"0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d\", 1]"
        }
    },
    "testExpect": {
        canSendTx: true,
        canSubmitTx: true,
        canExcuteTx: false,
        status: 0,
        fromBalanceAfterTx: "999999979505000000",
        toBalanceAfterTx: '0',
        transferReward: '20495000000'
    }
};
testCases.push(testCase);

testCase = {
    "name": "30. execution success",
    "testInput": {
        sign: true,
        from: from,
        to: contractAddr,
        value: "1",
        nonce: 1,
        gasPrice: 1000000,
        gasLimit: 2000000,
        contract: {
            "function": "balanceOf",
            "args": "[\"0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d\"]"
        }
    },
    "testExpect": {
        canSendTx: true,
        canSubmitTx: true,
        canExcuteTx: true,
        status: 1,
        fromBalanceAfterTx: "999999979787999999",
        toBalanceAfterTx: '1',
        transferReward: '20212000000'
    }
};
testCases.push(testCase);


describe('contract call test', function () {
    // beforeEach(function (done) {
    //     prepareContractCall(function (result) {
    //         if (result instanceof Error) {
    //             done(result);
    //         } else {
    //             done();
    //         }
    //     });
    // });


    // var testCase = testCases[29];
    // it(testCase.name, function (done) {
    //     prepareContractCall(testCase, function (err) {
    //         if (err instanceof Error) {
    //             done(err);
    //         } else {
    //             testContractCall(testCase.testInput, testCase.testExpect, done);
    //         }
    //     });
    // });


    for (var i = 0; i < testCases.length; i++) {

        it(testCases[i].name, function (done) {
            var testCase = testCases[caseIndex];
            prepareContractCall(testCase, function (err) {
                if (err instanceof Error) {
                    done(err);
                } else {
                    testContractCall(testCase.testInput, testCase.testExpect, done);
                }
            });
        });
    }

    afterEach(function () {
        caseIndex++;
        console.log("case index:", caseIndex);
    });
});
