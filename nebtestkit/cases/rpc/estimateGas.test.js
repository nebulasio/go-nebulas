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
neb.setRequest(new HttpRequest("http://localhost:8685"))
var chain_id = 100;
var env = 'local';
chain_id = 1002;
if (env === 'testneb1') {
    chain_id = 1001;
    sourceAccount = new Wallet.Account("25a3a441a34658e7a595a0eda222fa43ac51bd223017d17b420674fb6d0a4d52");
    coinbase = "n1SAeQRVn33bamxN4ehWUT7JGdxipwn8b17";
    server_address = "35.182.48.19:8684";
  
  } else if (env === "testneb2") {
    chain_id = 1002;
    sourceAccount = new Wallet.Account("25a3a441a34658e7a595a0eda222fa43ac51bd223017d17b420674fb6d0a4d52");
    coinbase = "n1SAeQRVn33bamxN4ehWUT7JGdxipwn8b17";
    server_address = "34.205.26.12:8684";
  
  } else if (env === "testneb3") {
    chain_id = 1003;
    sourceAccount = new Wallet.Account("25a3a441a34658e7a595a0eda222fa43ac51bd223017d17b420674fb6d0a4d52");
    coinbase = "n1SAeQRVn33bamxN4ehWUT7JGdxipwn8b17";
    server_address = "35.177.214.138:8684";
  
  } else if (env === "testneb4") { //super node
    chain_id = 1004;
    sourceAccount = new Wallet.Account("c75402f6ffe6edcc2c062134b5932151cb39b6486a7beb984792bb9da3f38b9f");
    coinbase = "n1EzGmFsVepKduN1U5QFyhLqpzFvM9sRSmG";
    server_address = "35.154.108.11:8684";
  } else if (env === "testneb4_normalnode"){
    chain_id = 1004;
    sourceAccount = new Wallet.Account("c75402f6ffe6edcc2c062134b5932151cb39b6486a7beb984792bb9da3f38b9f");
    coinbase = "n1EzGmFsVepKduN1U5QFyhLqpzFvM9sRSmG";
    server_address = "18.197.107.228:8684";
  } else if (env === "local") {
    chain_id = 100;
    sourceAccount = new Wallet.Account("d80f115bdbba5ef215707a8d7053c16f4e65588fd50b0f83369ad142b99891b5");
    coinbase = "n1QZMXSZtW7BUerroSms4axNfyBGyFGkrh5";
    server_address = "127.0.0.1:8684";
  
  } else if (env === "maintest"){
    chain_id = 2;
    sourceAccount = new Wallet.Account("d2319a8a63b1abcb0cc6d4183198e5d7b264d271f97edf0c76cfdb1f2631848c");
    coinbase = "n1dZZnqKGEkb1LHYsZRei1CH6DunTio1j1q";
    server_address = "54.149.15.132:8684";
  } else {
    throw new Error("invalid env (" + env + ").");
}
neb.setRequest(new HttpRequest("http://127.0.0.1:8685"));

var api_client;
var normalOutput;
var txHash;
var nonce;
var contractAddress;

var maxCheckTime = 20;
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

//should not throw err
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
        var tx = new Transaction(chain_id, sourceAccount, sourceAccount.getAddressString(),testInput.value, testInput.nonce,
            testInput.gas_price, testInput.gas_limit, testInput.contract);
            tx.signTransaction();
        // console.log(tx.toString());
        return neb.api.sendRawTransaction(tx.toProtoString());   
    }).then(function (resp) {
        console.log(resp);
        checkTransaction(resp.txhash, function (receipt) {
            try {
                console.log(receipt);
                expect(receipt.status).not.to.be.a('undefined');
            } catch (err) {
                done(err);
                return;
            }
            try {
                neb.api.getAccountState(sourceAccount.getAddressString()).then(function (state) {
                    balanceAfterTx = new BigNumber(state.balance);
                    var gasConsumed = balanceBeforeTx.sub(balanceAfterTx).div(new BigNumber(testInput.gas_price));
                    expect((new BigNumber(gas)).toString()).to.be.equal(gasConsumed.toString());
                    done();
                }).catch(function (err) {
                    console.log(err);
                    console.log("silent_debugggggg");
                    done(err);
                    return;
                });
            } catch (err) {
                done(err.error);
                return;
            }
        });
    }).catch(function (err) {
        console.log("silent_debug");
        console.log(err);
        done(err.error);
        return;
    });

}



function testRpc(testInput, testExpect, done) {
    api_client.EstimateGas(testInput.rpcInput, function (err, response) {
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
                expect(response.err).equal(testExpect.exeMsg);
            } catch (err) {
                console.log("unexpected errpr :", err);
                done(err);
                return;
            }
            var gas = parseInt(response.gas);
            console.log(gas, "to verify");
            verify(gas, testInput.verifyInput, done);
        }
    });

}

describe('rpc: estimateGas', function () {
    // //unlock the sourceAccount
    // before((done) => {
    //     var admin_client = rpc_client.new_client(server_address, 'AdminService');
    //     var args = {
    //         address: sourceAccount.getAddressString(),
    //         passphrase: "passphrase",
    //     }
    //     admin_client.UnlockAccount(args, (err, resp) => {
    //         expect(err).to.be.equal(null);
    //         done(err);
    //     })
    // });
    //get nonce
    beforeEach((done) => {
        api_client = rpc_client.new_client(server_address);
        api_client.GetAccountState({ address: sourceAccount.getAddressString() }, (err, resp) => {
            try{
                expect(err).to.be.equal(null);
                nonce = parseInt(resp.nonce);
                console.log("nonce is: ", nonce);
                done(err);
            } catch(err) {
                done(err);
                return;
            }
        });
    });

    it('normal rpc', function (done) {
        nonce = nonce + 1;
        var erc20 = FS.readFileSync("./nf/nvm/test/ERC20.js", "utf-8");
        var contract = {
            "source": erc20,
            "source_type": "js",
            "args": '["NebulasToken", "NAS", 1000000000]'
        };

        var contractVerify = {
            "source": erc20,
            "sourceType": "js",
            "args": '["NebulasToken", "NAS", 1000000000]'
        };

        var testInput = {
            rpcInput: {
                from: sourceAccount.getAddressString(),
                to: sourceAccount.getAddressString(),
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "2000000",
                contract: contract,
            },
            verifyInput: {
                from: sourceAccount.getAddressString(),
                to: sourceAccount.getAddressString(),
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "2000000",
                contract: contractVerify,
            },
        }
        var testExpect = {
            exeMsg: "",
        }
        testRpc(testInput, testExpect, done);
    });

    it('value is invalid', function (done) {
        nonce = nonce + 1;
        var erc20 = FS.readFileSync("./nf/nvm/test/ERC20.js", "utf-8");
        var contract = {
            "source": erc20,
            "source_type": "js",
            "args": '["NebulasToken", "NAS", 1000000000]'
        }

        var contractVerify = {
            "source": erc20,
            "sourceType": "js",
            "args": '["NebulasToken", "NAS", 1000000000]'
        };
        var testInput = {
            rpcInput: {
                from: sourceAccount.getAddressString(),
                to: sourceAccount.getAddressString(),
                value: "0a",
                nonce: 100000000,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contract,
            },
            verifyInput: {
                from: sourceAccount.getAddressString(),
                to: sourceAccount.getAddressString(),
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contractVerify,
            },
        }

        var testExpect = {
            errMsg: 'invalid value'
        }

        testRpc(testInput, testExpect, done);
    })

    it('value is empty', function (done) {
        nonce = nonce + 1;
        var erc20 = FS.readFileSync("./nf/nvm/test/ERC20.js", "utf-8");

        var contract = {
            "source": erc20,
            "source_type": "js",
            "args": '["NebulasToken", "NAS", 1000000000]',
        };
        var contractVerify = {
            "source": erc20,
            "sourceType": "js",
            "args": '["NebulasToken", "NAS", 1000000000]'
        };
        var testInput = {
            rpcInput: {
                from: sourceAccount.getAddressString(),
                to: sourceAccount.getAddressString(),
                nonce: 100000000,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contract,
            },
            verifyInput: {
                from: sourceAccount.getAddressString(),
                to: sourceAccount.getAddressString(),
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contractVerify,
            },
        }

        var testExpect = {
            errMsg: 'invalid value'
         }
        testRpc(testInput, testExpect, done);
    });

    it('nonce is large', function (done) {
        nonce = nonce + 1;
        var erc20 = FS.readFileSync("./nf/nvm/test/ERC20.js", "utf-8");
        var contract = {
            "source": erc20,
            "source_type": "js",
            "args": '["NebulasToken", "NAS", 1000000000]'
        }
        var contractVerify = {
            "source": erc20,
            "sourceType": "js",
            "args": '["NebulasToken", "NAS", 1000000000]'
        };
        var testInput = {
            rpcInput: {
                from: sourceAccount.getAddressString(),
                to: sourceAccount.getAddressString(),
                value: "0",
                nonce: 100000000,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contract,
            },
            verifyInput: {
                from: sourceAccount.getAddressString(),
                to: sourceAccount.getAddressString(),
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contractVerify,
            },
        }
        var testExpect = {
            exeMsg: ""
        };
        testRpc(testInput, testExpect, done);
    });

    it('nonce is empty', function (done) {
        nonce = nonce + 1;
        var erc20 = FS.readFileSync("./nf/nvm/test/ERC20.js", "utf-8");
        var contract = {
            "source": erc20,
            "source_type": "js",
            "args": '["NebulasToken", "NAS", 1000000000]'
        };
        var contractVerify = {
            "source": erc20,
            "sourceType": "js",
            "args": '["NebulasToken", "NAS", 1000000000]'
        };
        var testInput = {
            rpcInput: {
                from: sourceAccount.getAddressString(),
                to: sourceAccount.getAddressString(),
                value: "0",
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contract,
            },
            verifyInput: {
                from: sourceAccount.getAddressString(),
                to: sourceAccount.getAddressString(),
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contractVerify,
            },
        }
        var testExpect = {
            exeMsg: ""
        };
        testRpc(testInput, testExpect, done);
    })

    it('nonce is small', function (done) {
        nonce = nonce + 1;
        var erc20 = FS.readFileSync("./nf/nvm/test/ERC20.js", "utf-8");
        var contract = {
            "source": erc20,
            "source_type": "js",
            "args": '["NebulasToken", "NAS", 1000000000]'
        }
        var contractVerify = {
            "source": erc20,
            "sourceType": "js",
            "args": '["NebulasToken", "NAS", 1000000000]'
        };
        var testInput = {
            rpcInput: {
                from: sourceAccount.getAddressString(),
                to: sourceAccount.getAddressString(),
                value: "0",
                nonce: 1,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contract,
            },
            verifyInput: {
                from: sourceAccount.getAddressString(),
                to: sourceAccount.getAddressString(),
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contractVerify,
            },
        }
        var testExpect = {
            exeMsg: ""
        };
        testRpc(testInput, testExpect, done);
    });

    it('gasPrice is negative', function (done) {
        nonce = nonce + 1;
        var erc20 = FS.readFileSync("./nf/nvm/test/ERC20.js", "utf-8");
        var contract = {
            "source": erc20,
            "source_type": "js",
            "args": '["NebulasToken", "NAS", 1000000000]'
        };
        var contractVerify = {
            "source": erc20,
            "sourceType": "js",
            "args": '["NebulasToken", "NAS", 1000000000]'
        };
        var testInput = {
            rpcInput: {
                from: sourceAccount.getAddressString(),
                to: sourceAccount.getAddressString(),
                value: "0",
                nonce: nonce,
                gas_price: "-1",
                gas_limit: "200000",
                contract: contract,
            },
            verifyInput: {
                from: sourceAccount.getAddressString(),
                to: sourceAccount.getAddressString(),
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contractVerify,
            },
        }

        var testExpect = {
            errMsg: 'invalid gasPrice'
        }

        testRpc(testInput, testExpect, done);
    });

    it('gas_price is less than gasPrince of tx pool', function (done) {
        nonce = nonce + 1;
        var erc20 = FS.readFileSync("./nf/nvm/test/ERC20.js", "utf-8");
        var contract = {
            "source": erc20,
            "source_type": "js",
            "args": '["NebulasToken", "NAS", 1000000000]'
        };
        var contractVerify = {
            "source": erc20,
            "sourceType": "js",
            "args": '["NebulasToken", "NAS", 1000000000]'
        };
        var testInput = {
            rpcInput: {
                from: sourceAccount.getAddressString(),
                to: sourceAccount.getAddressString(),
                value: "0",
                nonce: 1,
                gas_price: "100",
                gas_limit: "200000",
                contract: contract,
            },
            verifyInput: {
                from: sourceAccount.getAddressString(),
                to: sourceAccount.getAddressString(),
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contractVerify,
            },
        }
        var testExpect = {
            exeMsg: ""
        };
        testRpc(testInput, testExpect, done);
    });

    it('gasLimit is neg', function (done) {
        nonce = nonce + 1;
        var erc20 = FS.readFileSync("./nf/nvm/test/ERC20.js", "utf-8");
        var contract = {
            "source": erc20,
            "source_type": "js",
            "args": '["NebulasToken", "NAS", 1000000000]'
        }
        var contractVerify = {
            "source": erc20,
            "sourceType": "js",
            "args": '["NebulasToken", "NAS", 1000000000]'
        };
        var testInput = {
            rpcInput: {
                from: sourceAccount.getAddressString(),
                to: sourceAccount.getAddressString(),
                value: "0",
                nonce: 1,
                gas_price: "1000000",
                gas_limit: "-1",
                contract: contract,
            },
            verifyInput: {
                from: sourceAccount.getAddressString(),
                to: sourceAccount.getAddressString(),
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contractVerify
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
            "source": erc20,
            "source_type": "js",
            "args": '["NebulasToken", "NAS", 1000000000]'
        }
        var contractVerify = {
            "source": erc20,
            "sourceType": "js",
            "args": '["NebulasToken", "NAS", 1000000000]'
        };
        var testInput = {
            rpcInput: {
                from: sourceAccount.getAddressString(),
                to: sourceAccount.getAddressString(),
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "2000",
                contract: contract,
            },
            verifyInput: {
                from: sourceAccount.getAddressString(),
                to: sourceAccount.getAddressString(),
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contractVerify
            },
        }

        var testExpect = {
            exeMsg: ''
        }

        testRpc(testInput, testExpect, done);
    });

    it('tx failed to execute', function (done) {
        nonce = nonce + 1;
        var erc20 = FS.readFileSync("./nf/nvm/test/ERC20.initError.js", "utf-8");
        var contract = {
            "source": erc20,
            "source_type": "js",
            "args": '["NebulasToken", "NAS", 1000000000]'
        }
        var contractVerify = {
            "source": erc20,
            "sourceType": "js",
            "args": '["NebulasToken", "NAS", 1000000000]'
        };
        var testInput = {
            rpcInput: {
                from: sourceAccount.getAddressString(),
                to: sourceAccount.getAddressString(),
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contract,
            },
            verifyInput: {
                from: sourceAccount.getAddressString(),
                to: sourceAccount.getAddressString(),
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contractVerify
            },
        }

        var testExpect = {
            exeMsg: 'Deploy: fail to init'
        }

        testRpc(testInput, testExpect, done);
    });

    it('contract type is empty', function (done) {
        nonce = nonce + 1;
        var erc20 = FS.readFileSync("./nf/nvm/test/ERC20.js", "utf-8");
        var contract = {
            "source": erc20,
            "args": '["NebulasToken", "NAS", 1000000000]'
        }
        var contractVerify = {
            "source": erc20,
            "sourceType": "js",
            "args": '["NebulasToken", "NAS", 1000000000]'
        };
        var testInput = {
            rpcInput: {
                from: sourceAccount.getAddressString(),
                to: sourceAccount.getAddressString(),
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contract,
            },
            verifyInput: {
                from: sourceAccount.getAddressString(),
                to: sourceAccount.getAddressString(),
                value: "0",
                nonce: nonce,
                gas_price: "1000000",
                gas_limit: "200000",
                contract: contractVerify
            },
        }

        var testExpect = {
            errMsg: 'invalid source type of deploy payload'
        }

        testRpc(testInput, testExpect, done);
    });

});

