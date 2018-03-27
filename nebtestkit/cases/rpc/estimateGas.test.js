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
var transaction = Wallet.Transaction;
var Utils = Wallet.Utils;
var Unit = Wallet.Unit;

var protocol_version = '/neb/1.0.0'
var node_version = '0.7.0'
var server_address = 'localhost:8684';
var coinbase = "n1QZMXSZtW7BUerroSms4axNfyBGyFGkrh5";
var sourceAccount = new Wallet.Account('d80f115bdbba5ef215707a8d7053c16f4e65588fd50b0f83369ad142b99891b5');
neb.setRequest(new HttpRequest("http://localhost:8685"))
var chain_id = 100;
var env = '';
if (env === 'testneb1') {

} else if (env === "testneb2") {
    server_address = "34.205.26.12:8684";
    neb.setRequest(new HttpRequest("http://34.205.26.12:8685"))
    coinbase = "n1SAeQRVn33bamxN4ehWUT7JGdxipwn8b17";
    chain_id = 1002;
    sourceAccount = new Wallet.Account("25a3a441a34658e7a595a0eda222fa43ac51bd223017d17b420674fb6d0a4d52");
}

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
                expect(response.err).equal(testExpect.resultMsg);
            } catch (err) {
                console.log("unexpected errpr :", err);
                done(err);
                return;
            }
            var gas = parseInt(response.gas);
            console.log(gas, "to verify");
            verify(gas, testInput, done);
        }
    });

}

describe('rpc: estimateGas', function () {
    //unlock the sourceAccount
    before((done) => {
        var admin_client = rpc_client.new_client(server_address, 'AdminService');
        var args = {
            address: sourceAccount.getAddressString(),
            passphrase: "passphrase",
        }
        admin_client.UnlockAccount(args, (err, resp) => {
            expect(err).to.be.equal(null);
            done(err);
        })
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
        var erc20 = FS.readFileSync("./nf/nvm/test/ERC20.js", "utf-8");
        var contract = {
            "source": erc20,
            "source_type": "js",
            "args": '["NebulasToken", "NAS", 1000000000]'
        }
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
                contract: contract,
            },
        }
        var testExpect = {
            resultMsg: "",
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
        var erc20 = FS.readFileSync("./nf/nvm/test/ERC20.js", "utf-8");
        var contract = {
            "source": erc20,
            "source_type": "js",
            "args": '["NebulasToken", "NAS", 1000000000]'
        }
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
        var erc20 = FS.readFileSync("./nf/nvm/test/ERC20.js", "utf-8");
        var contract = {
            "source": erc20,
            "source_type": "js",
            "args": '["NebulasToken", "NAS", 1000000000]'
        }
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
                contract: contract
            },
        }
        var testExpect = {
            resultMsg: ""
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
        }
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
                contract: contract
            },
        }
        var testExpect = {
            resultMsg: ""
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
                contract: contract
            },
        }
        var testExpect = {
            resultMsg: ""
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
        }
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
        var erc20 = FS.readFileSync("./nf/nvm/test/ERC20.js", "utf-8");
        var contract = {
            "source": erc20,
            "source_type": "js",
            "args": '["NebulasToken", "NAS", 1000000000]'
        }
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
                contract: contract
            },
        }
        var testExpect = {
            resultMsg: ""
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
            "source": erc20,
            "source_type": "js",
            "args": '["NebulasToken", "NAS", 1000000000]'
        }
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
                contract: contract
            },
        }

        var testExpect = {
            resultMsg: ''
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
                contract: contract
            },
        }

        var testExpect = {
            resultMsg: 'Deploy: fail to init'
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
                contract: contract
            },
        }

        var testExpect = {
            errMsg: 'invalid source type of deploy payload'
        }

        testRpc(testInput, testExpect, done);
    });

});

