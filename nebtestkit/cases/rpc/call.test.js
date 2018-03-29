'use strict';
var expect = require('chai').expect;
var FS = require('fs');
var BigNumber = require('bignumber.js');
var HttpRequest = require("../../node-request");
var rpc_client = require('./rpc_client/rpc_client.js');
var Wallet = require("nebulas");
var Neb = Wallet.Neb;
var neb = new Neb();
var Account = Wallet.Account;
var Transaction = Wallet.Transaction;
var Utils = Wallet.Utils;
var Unit = Wallet.Unit;

var protocol_version = '/neb/1.0.0'
var node_version = '0.7.0'
var server_address = 'localhost:8684';
var coinbase = "n1QZMXSZtW7BUerroSms4axNfyBGyFGkrh5";
var sourceAccount = new Wallet.Account('d80f115bdbba5ef215707a8d7053c16f4e65588fd50b0f83369ad142b99891b5');

var ChainID = 100;
var env = 'maintest';


if (env === 'testneb1') {
  ChainID = 1001;
  sourceAccount = new Wallet.Account("25a3a441a34658e7a595a0eda222fa43ac51bd223017d17b420674fb6d0a4d52");
  coinbase = "n1SAeQRVn33bamxN4ehWUT7JGdxipwn8b17";
  server_address = "35.182.48.19:8684";

} else if (env === "testneb2") {
  ChainID = 1002;
  sourceAccount = new Wallet.Account("25a3a441a34658e7a595a0eda222fa43ac51bd223017d17b420674fb6d0a4d52");
  coinbase = "n1SAeQRVn33bamxN4ehWUT7JGdxipwn8b17";
  server_address = "34.205.26.12:8684";

} else if (env === "testneb3") {
    ChainID = 1003;
  sourceAccount = new Wallet.Account("25a3a441a34658e7a595a0eda222fa43ac51bd223017d17b420674fb6d0a4d52");
  coinbase = "n1SAeQRVn33bamxN4ehWUT7JGdxipwn8b17";
  server_address = "35.177.214.138:8684";

} else if (env === "testneb4") { //super node
    ChainID = 1004;
  sourceAccount = new Wallet.Account("c75402f6ffe6edcc2c062134b5932151cb39b6486a7beb984792bb9da3f38b9f");
  coinbase = "n1EzGmFsVepKduN1U5QFyhLqpzFvM9sRSmG";
  server_address = "35.154.108.11:8684";
} else if (env === "testneb4_normalnode"){
    ChainID = 1004;
  sourceAccount = new Wallet.Account("c75402f6ffe6edcc2c062134b5932151cb39b6486a7beb984792bb9da3f38b9f");
  coinbase = "n1EzGmFsVepKduN1U5QFyhLqpzFvM9sRSmG";
  server_address = "18.197.107.228:8684";
} else if (env === "local") {
  ChainID = 100;
  sourceAccount = new Wallet.Account("d80f115bdbba5ef215707a8d7053c16f4e65588fd50b0f83369ad142b99891b5");
  coinbase = "n1QZMXSZtW7BUerroSms4axNfyBGyFGkrh5";
  server_address = "127.0.0.1:8684";

} else if (env === "maintest"){
    ChainID = 2;
  sourceAccount = new Wallet.Account("d2319a8a63b1abcb0cc6d4183198e5d7b264d271f97edf0c76cfdb1f2631848c");
  coinbase = "n1dZZnqKGEkb1LHYsZRei1CH6DunTio1j1q";
  server_address = "54.149.15.132:8684";
} else {
  throw new Error("invalid env (" + env + ").");
}
neb.setRequest(new HttpRequest("http://54.149.15.132:8685"));

var api_client;
var normalOutput;
var txHash;
var nonce;
var contractAddress;
var toAddress = Wallet.Account.NewAccount();

var maxCheckTime = 40;
var checkTimes = 0;

console.log("running chain_id: ", ChainID, " coinbase:", coinbase, " server_address:", server_address);

function checkTransaction(hash, callback) {
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
            callback(resp);
        }
    }).catch(function (err) {
        console.log(err.error);
        console.log("maybe packing");
        setTimeout(function () {
            checkTransaction(hash, callback);
        }, 2000);
    });
}


function verify(gas, testInput, done) {
    try {
        var balanceBeforeTx, balanceAfterTx;
    } catch (err) {
        done(err);
        return;
    }

    neb.api.getAccountState(sourceAccount.getAddressString()).then(function (state) {
        try {
            balanceBeforeTx = new BigNumber(state.balance);
        } catch (err) {
            done(err)
            return;
        }
        var admin_client = rpc_client.new_client(server_address, 'AdminService');
        admin_client.SendTransaction(testInput.verifyInput, function (err, resp) {
            try {
                expect(err).to.be.equal(null);
            } catch (err) {
                done(err);
                return;
            }
            checkTransaction(resp.txhash, function (receipt) {
                try {
                    expect(receipt.status).not.to.be.a('undefined');
                } catch (err) {
                    done(err);
                    return;
                }
                try {
                    neb.api.getAccountState(sourceAccount.getAddressString()).then(function (state) {
                        balanceAfterTx = new BigNumber(state.balance);
                        var gasConsumed = balanceBeforeTx.sub(balanceAfterTx).div(new BigNumber(testInput.verifyInput.gas_price));
                        expect((new BigNumber(gas)).toString()).to.be.equal(gasConsumed.toString());
                    }).catch(function (err) {
                        done(err);
                        return;
                    });
                    done()
                } catch (err) {
                    done(err);
                    return;
                }
            });
        });
    }).catch(function (err) {
        done(err);
        return;
    });

}



function testRpc(testInput, testExpect, done) {
    api_client.Call(testInput.rpcInput, function (err, response) {
        if (err) {
            try {
                expect(testExpect.errMsg).to.be.equal(err.details);
            } catch (err) {
                done(err);
                return;
            }
            done();
        } else {
            console.log(JSON.stringify(response));
            try {
                expect(testExpect.resultMsg).to.be.equal(response.result);
                expect(response.execute_err).equal(testExpect.exeErr);
            } catch (err) {
                console.log("unexpected errpr :", err);
                done(err);
                return;
            }
            var gas = parseInt(response.estimate_gas);
            console.log("to verify");
            verify(gas, testInput, done);
        }
    });
}

describe('rpc: Call', function () {
    before('deploy contract', function (done) {
        try {
            neb.api.getAccountState(sourceAccount.getAddressString()).then(function(resp) {
                console.log("----step0. get source account state: " + JSON.stringify(resp));
                var contractSource = FS.readFileSync("./nf/nvm/test/transfer_value_from_contract.js", "UTF-8");
                var contract = {
                    'source': contractSource,
                    "sourceType": "js",
                    "arges": ''
                };
                nonce = parseInt(resp.nonce);
                nonce = nonce + 1;
                var tx = new Transaction(ChainID, sourceAccount, sourceAccount, 0, nonce, 1000000, 20000000, contract);
                tx.signTransaction();
                console.log(tx.toProtoString());
                return neb.api.sendRawTransaction(tx.toProtoString());
            }).then(function(resp) {
                console.log("----step1. deploy contract: " + JSON.stringify(resp));
                contractAddress = resp.contract_address;
                checkTransaction(resp.txhash, function(resp) {
                    expect(resp).to.not.be.a('undefined');
                    console.log("----step2. have been on chain");
                    done();
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

    before ('send 10 nas to contract address', function (done) {
        nonce = nonce + 1;
        console.log(contractAddress);
        var tx = new Transaction(ChainID, sourceAccount, contractAddress, Unit.nasToBasic(10), nonce, 1000000, 2000000);
        // tx.to = contractAddress;
        tx.signTransaction();
        console.log(tx.toString());
        // console.log("silent_debug");
        neb.api.sendRawTransaction(tx.toProtoString()).then(function(resp) {
            console.log("----step3. send nas to contract address: ", resp);
            checkTransaction(resp.txhash, function(resp) {
                expect(resp).to.not.be.a('undefined');
                console.log("----step4. have been on chain");
                done();
            });
        }).catch(function(err) {
            console.log("unexpected err: " + err);
            done(err);
        });
    });

    //get nonce
    beforeEach((done) => {
        api_client = rpc_client.new_client(server_address);
        api_client.GetAccountState({ address: sourceAccount.getAddressString() }, (err, resp) => {
            expect(err).to.be.equal(null);
            nonce = parseInt(resp.nonce);
            done(err);
        });
    });

    it('normal rpc', function (done) {
        nonce = nonce + 1;
        var contract = {
            "function": "transferSpecialValue",
            "args": "[\"" + toAddress.getAddressString() + "\", \"5000000000000000000\"]"
        };
        var testInput = {
            rpcInput: {
                from: sourceAccount.getAddressString(),
                to: contractAddress,
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contract,
            },
            verifyInput: {
                from: sourceAccount.getAddressString(),
                to: contractAddress,
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contract,
            },
        }
        var testExpect = {
            exeErr: "",
            resultMsg: "0",
        }
        testRpc(testInput, testExpect, done);
    });
    
    it('call function returns an error ', function (done) {
        nonce = nonce + 1;
        var contract = {
            "function": "transferSpecialValue",
            "args": "[\"" + toAddress.getAddressString() + "\", \"11000000000000000000\"]"
        };
        var testInput = {
            rpcInput: {
                from: sourceAccount.getAddressString(),
                to: contractAddress,
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contract,
            },
            verifyInput: {
                from: sourceAccount.getAddressString(),
                to: contractAddress,
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contract,
            },
        }
        var testExpect = {
            exeErr: "Call: Error: transfer failed.",
            resultMsg: "Error: transfer failed.",
        }
        testRpc(testInput, testExpect, done);
    });

    it('call function success but balanace is not enough ', function (done) {
        nonce = nonce + 1;
        var contract = {
            "function": "transferSpecialValue",
            "args": "[\"" + toAddress.getAddressString() + "\", \"5000000000000000000\"]"
        };
        var testInput = {
            rpcInput: {
                from: Wallet.Account.NewAccount().getAddressString(),
                to: contractAddress,
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contract,
            },
            verifyInput: {
                from: sourceAccount.getAddressString(),
                to: contractAddress,
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contract,
            },
        }
        var testExpect = {
            exeErr: "insufficient balance",
            resultMsg: "0",
        }
        testRpc(testInput, testExpect, done);
    });

    it('value is invalid', function (done) {
        nonce = nonce + 1;
        var contract = {
            "function": "transferSpecialValue",
            "args": "[\"" + toAddress.getAddressString() + "\", \"5000000000000000000\"]"
        };
        var testInput = {
            rpcInput: {
                from: sourceAccount.getAddressString(),
                to: contractAddress,
                value: "0a",
                nonce: 100000000,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contract,
            },
            verifyInput: {
                from: sourceAccount.getAddressString(),
                to: contractAddress,
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contract
            },
        }

        var testExpect = {
            errMsg: 'invalid value'
        }

        testRpc(testInput, testExpect, done);
    })

    it('value is empty', function (done) {
        nonce = nonce + 1;
        var contract = {
            "function": "transferSpecialValue",
            "args": "[\"" + toAddress.getAddressString() + "\", \"5000000000000000000\"]"
        };
        var testInput = {
            rpcInput: {
                from: sourceAccount.getAddressString(),
                to: contractAddress,
                nonce: 100000000,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contract,
            },
            verifyInput: {
                from: sourceAccount.getAddressString(),
                to: contractAddress,
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contract
            },
        }

        var testExpect = {
            errMsg: 'invalid value'
         }

        testRpc(testInput, testExpect, done);
    })

    it('nonce is large', function (done) {
        nonce = nonce + 1;
        var contract = {
            "function": "transferSpecialValue",
            "args": "[\"" + toAddress.getAddressString() + "\", \"5000000000000000000\"]"
        };
        var testInput = {
            rpcInput: {
                from: sourceAccount.getAddressString(),
                to: contractAddress,
                value: "0",
                nonce: 100000000,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contract,
            },
            verifyInput: {
                from: sourceAccount.getAddressString(),
                to: contractAddress,
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contract
            },
        }
        var testExpect = {
            exeErr: "",
            resultMsg: "0"
        };
        testRpc(testInput, testExpect, done);
    });

    it('nonce is empty', function (done) {//todo: to check
        nonce = nonce + 1;
        var contract = {
            "function": "transferSpecialValue",
            "args": "[\"" + toAddress.getAddressString() + "\", \"5000000000000000000\"]"
        };
        var testInput = {
            rpcInput: {
                from: sourceAccount.getAddressString(),
                to: contractAddress,
                value: "0",
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contract,
            },
            verifyInput: {
                from: sourceAccount.getAddressString(),
                to: contractAddress,
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contract
            },
        }
        var testExpect = {
            exeErr: "",
            resultMsg: "0"
        };
        testRpc(testInput, testExpect, done);
    })

    it('nonce is small', function (done) {
        nonce = nonce + 1;
        var contract = {
            "function": "transferSpecialValue",
            "args": "[\"" + toAddress.getAddressString() + "\", \"5000000000000000000\"]"
        };
        var testInput = {
            rpcInput: {
                from: sourceAccount.getAddressString(),
                to: contractAddress,
                value: "0",
                nonce: 1,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contract,
            },
            verifyInput: {
                from: sourceAccount.getAddressString(),
                to: contractAddress,
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contract
            },
        }
        var testExpect = {
            exeErr: "",
            resultMsg: "0"
        };
        testRpc(testInput, testExpect, done);
    });

    it('gasPrice is negative', function (done) {
        nonce = nonce + 1;
        var contract = {
            "function": "transferSpecialValue",
            "args": "[\"" + toAddress.getAddressString() + "\", \"5000000000000000000\"]"
        };
        var testInput = {
            rpcInput: {
                from: sourceAccount.getAddressString(),
                to: contractAddress,
                value: "0",
                nonce: nonce,
                gas_price: "-1",
                gas_limit: "200000",
                contract: contract,
            },
            verifyInput: {
                from: sourceAccount.getAddressString(),
                to: contractAddress,
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contract
            },
        }

        var testExpect = {
            errMsg: 'invalid gasPrice'
        }

        testRpc(testInput, testExpect, done);
    });

    it('gas_price is less than gasPrince of tx pool', function (done) {
        nonce = nonce + 1;
        var contract = {
            "function": "transferSpecialValue",
            "args": "[\"" + toAddress.getAddressString() + "\", \"5000000000000000000\"]"
        };
        var testInput = {
            rpcInput: {
                from: sourceAccount.getAddressString(),
                to: contractAddress,
                value: "0",
                nonce: 1,
                gas_price: "100",
                gas_limit: "200000",
                contract: contract,
            },
            verifyInput: {
                from: sourceAccount.getAddressString(),
                to: contractAddress,
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contract
            },
        }
        var testExpect = {
            exeErr: "",
            resultMsg: "0"
        };
        testRpc(testInput, testExpect, done);
    });

    it('gasLimit is neg', function (done) {
        nonce = nonce + 1;
        var contract = {
            "function": "transferSpecialValue",
            "args": "[\"" + toAddress.getAddressString() + "\", \"5000000000000000000\"]"
        };
        var testInput = {
            rpcInput: {
                from: sourceAccount.getAddressString(),
                to: contractAddress,
                value: "0",
                nonce: 1,
                gas_price: "1000000",
                gas_limit: "-1",
                contract: contract,
            },
            verifyInput: {
                from: sourceAccount.getAddressString(),
                to: contractAddress,
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contract
            },
        }

        var testExpect = {
            errMsg: 'invalid gasLimit'
        }

        testRpc(testInput, testExpect, done);
    });

    it('gasLimit is sufficient', function (done) {
        nonce = nonce + 1;
        var erc20 = FS.readFileSync("./nf/nvm/test/ERC20.js", "utf-8");
        var contract = {
            "function": "transferSpecialValue",
            "args": "[\"" + toAddress.getAddressString() + "\", \"5000000000000000000\"]"
        };
        var testInput = {
            rpcInput: {
                from: sourceAccount.getAddressString(),
                to: contractAddress,
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "2000",
                contract: contract,
            },
            verifyInput: {
                from: sourceAccount.getAddressString(),
                to: contractAddress,
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contract
            },
        }

        var testExpect = {
            exeErr: '',
            resultMsg: '0'
        }

        testRpc(testInput, testExpect, done);
    });
    
    it('args is less than that required', function (done) {
        nonce = nonce + 1;
        var contract = {
            "function": "transferSpecialValue",
            "args": "[\"" + toAddress.getAddressString() + "\"]"
        };
        var testInput = {
            rpcInput: {
                from: sourceAccount.getAddressString(),
                to: contractAddress,
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contract,
            },
            verifyInput: {
                from: sourceAccount.getAddressString(),
                to: contractAddress,
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contract
            },
        }

        var testExpect = {
            exeErr: 'Call: BigNumber Error: new BigNumber() not a number: undefined',
            resultMsg: 'BigNumber Error: new BigNumber() not a number: undefined',

        }

        testRpc(testInput, testExpect, done);
    });

});

