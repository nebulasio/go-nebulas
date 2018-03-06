'use strict';
var expect = require('chai').expect;
var FS = require('fs');
var BigNumber = require('bignumber.js');
var HttpRequest = require("../../node-request");
var rpc_client = require('./rpc_client/rpc_client.js');
var cryptoUtils = require('../../../cmd/console/neb.js/lib/utils/crypto-utils.js');
var Wallet = require("../../../cmd/console/neb.js/lib/wallet");
var Neb = Wallet.Neb;
var neb = new Neb();
var Account = Wallet.Account;
var transaction = Wallet.Transaction;
var Utils = Wallet.Utils;
var Unit = Wallet.Unit;

var protocol_version = '/neb/1.0.0'
var node_version = '0.7.0'
var server_address = 'localhost:8684';
var coinbase = "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8";
var sourceAccount = new Wallet.Account('a6e5eb290e1438fce79f5cb8774a72621637c2c9654c8b2525ed1d7e4e73653f');
neb.setRequest(new HttpRequest("http://localhost:8685"))
var chain_id = 100;
var env = '';
if (env === 'testneb1') {
    server_address = '35.182.48.19:8684';
    neb.setRequest(new HttpRequest("http://35.182.48.19:8685"))
    coinbase = "0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d";
    chain_id = 1001;
    sourceAccount = new Wallet.Account("43181d58178263837a9a6b08f06379a348a5b362bfab3631ac78d2ac771c5df3");
} else if (env === "testneb2") {
    server_address = "34.205.26.12:8684";
    neb.setRequest(new HttpRequest("http://34.205.26.12:8685"))
    coinbase = "0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d";
    chain_id = 1002;
    sourceAccount = new Wallet.Account("43181d58178263837a9a6b08f06379a348a5b362bfab3631ac78d2ac771c5df3");
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
        console.log(err);
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
        api_client.SendTransaction(testInput.verifyInput, function (err, resp) {
            try {
                expect(err).to.be.equal(null);
            } catch (err) {
                done(err);
                return;
            }
            checkTransaction(resp.txhash, function (receipt) {
                try {
                    expect(receipt.status).to.be.equal(1);
                } catch (err) {
                    done(err);
                    rturn;
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
        retrun;
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
            if (testInput.isNormal) {
                //TODO:verify response
                //  expect(response.balance).to.be.a("string");
                normalOutput = response;
            } else {
                if (testExpect.isNormalOutput) {
                    try {
                        expect(JSON.stringify(response)).to.be.equal(JSON.stringify(normalOutput));
                    } catch (err) {
                        done(err);
                        return;
                    }
                    done();
                } else {
                    try {
                        expect(testExpect.isNormalOutput).to.be.equal(false);
                        expect(JSON.stringify(response)).not.be.equal(JSON.stringify(normalOutput));
                        var gas = parseInt(response.gas);
                    } catch (err) {
                        done(err);
                        retrun;
                    }
                    console.log(gas);
                    verify(gas, testInput, done);
                    //TODO: verify response
                }
            }
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
            isNormal: false
        }

        var testExpect = {
            isNormalOutput: false
        }

        testRpc(testInput, testExpect, done);
    })

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
            isNormal: false
        }

        var testExpect = {
            isNormalOutput: false,
            errMsg: 'uint128: invalid string to uint128'
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
            isNormal: false
        }

        var testExpect = {
            isNormalOutput: false,
            errMsg: 'uint128: invalid string to uint128'
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
            isNormal: false
        }

        var testExpect = {
            isNormalOutput: false,
        }

        testRpc(testInput, testExpect, done);
    })

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
            isNormal: false
        }

        var testExpect = {
            isNormalOutput: false,
        }

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
            isNormal: false
        }

        var testExpect = {
            isNormalOutput: false,
        }

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
            isNormal: false
        }

        var testExpect = {
            errMsg: 'uint128: underflow'
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
            isNormal: false
        }

        var testExpect = {
            isNormalOutput: false,
        }

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
            isNormal: false
        }

        var testExpect = {
            isNormalOutput: false,
            errMsg: 'uint128: underflow'
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
            isNormal: false
        }

        var testExpect = {
            isNormalOutput: false,
            errMsg: 'out of gas limit'
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
            isNormal: false
        }

        var testExpect = {
            isNormalOutput: false,
            errMsg: 'execution failed'
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
            isNormal: false
        }

        var testExpect = {
            isNormalOutput: false,
            errMsg: 'unsupported source type'
        }

        testRpc(testInput, testExpect, done);
    });

});