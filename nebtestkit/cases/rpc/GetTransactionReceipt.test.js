'use strict';
var expect = require('chai').expect;
var rpc_client = require('./rpc_client/rpc_client.js');

var protocol_version = '/neb/1.0.0'
var node_version = '0.7.0'
var server_address = 'localhost:8684';
var coinbase = "n1QZMXSZtW7BUerroSms4axNfyBGyFGkrh5";
var Wallet = require("nebulas");
var sourceAccount = new Wallet.Account('d80f115bdbba5ef215707a8d7053c16f4e65588fd50b0f83369ad142b99891b5');
var chain_id = 100;
var env = '';
if (env === 'testneb1') {
    server_address = 'http://35.182.48.19:8684';
    coinbase = "0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d";
    chain_id = 1001;
} else if (env === "testneb2") {
    server_address = "http://34.205.26.12:8685";
    coinbase = "0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d";
    chain_id = 1002;
}

var api_client;
var normalOutput;
var txHash;
var nonce;

function testRpc(testInput, testExpect, done) {
    api_client.GetTransactionReceipt(testInput.rpcInput, function (err, response) {
        try {
            if (err) {
                expect(testExpect.errMsg).to.be.equal(err.details);
            } else {
                if (testInput.isNormal) {
                    //TODO:verify response
                    //  expect(response.balance).to.be.a("string");
                    normalOutput = response;
                } else {
                    if (testExpect.isNormalOutput) {
                        expect(JSON.stringify(response)).to.be.equal(JSON.stringify(normalOutput));
                    } else {
                        expect(testExpect.isNormalOutput).to.be.equal(false);
                        expect(JSON.stringify(response)).not.be.equal(JSON.stringify(normalOutput));
                        //TODO: verify response
                    }
                }

            }
            done();
        } catch (err) {
            done(err);
        }
    });

}

describe('rpc: sendTransaction', function () {
    before((done) => {
        var admin_client = rpc_client.new_client(server_address, 'AdminService');
        var args = {
            address: sourceAccount.getAddressString(),
            passphrase: "passphrase",
        }
        admin_client.UnlockAccount(args, (err, resp) => {
            expect(err).to.be.equal(null);
            console.log(err);
            done(err);
        })
    });

    before((done) => {
        api_client = rpc_client.new_client(server_address);
        api_client.GetAccountState({ address: sourceAccount.getAddressString() }, (err, resp) => {
            expect(err).to.be.equal(null);
            nonce = parseInt(resp.nonce);
            done(err);
        });
    })

    before((done) => {
        nonce = nonce + 1;
        var args = {
            from: sourceAccount.getAddressString(),
            to: coinbase,
            value: "1",
            nonce: nonce,
            gas_price: "1000000",
            gas_limit: "200000",
        },
            admin_client = rpc_client.new_client(server_address, "AdminService");
        admin_client.SendTransaction(args, (err, resp) => {
            expect(err).to.be.equal(null);
            txHash = resp.txhash;
            done(err);
        });
    });

    it('normal', function (done) {
        nonce = nonce + 1;
        var testInput = {
            rpcInput: {
                hash: txHash,
            },
            isNormal: true
        }

        var testExpect = {
            isNormalOutput: true,
        }

        testRpc(testInput, testExpect, done);
    });

    it('hash is not exist', function (done) {
        nonce = nonce + 1;
        var testInput = {
            rpcInput: {
                hash: '1c4d8ddcf2e0b41b87df09684a594468d7750e87b72f5e77320e02781a05170c',
            },
            isNormal: false
        }

        var testExpect = {
            errMsg: 'transaction not found'
        }

        testRpc(testInput, testExpect, done);
    });

    it('hash is empty', function (done) {
        nonce = nonce + 1;
        var testInput = {
            rpcInput: {

            },
            isNormal: false
        }

        var testExpect = {
            isNormalOutput: true,
            errMsg: 'invalid argument(s)'
        }

        testRpc(testInput, testExpect, done);
    });

    it('tx is null', function (done) {
        nonce = nonce + 1;
        var testInput = {
            rpcInput: {
                hash: "",
            },
            isNormal: false
        }

        var testExpect = {
            isNormalOutput: true,
            errMsg: 'invalid argument(s)'
        }

        testRpc(testInput, testExpect, done);
    });

    it('hash is different length', function (done) {
        nonce = nonce + 1;
        var testInput = {
            rpcInput: {
                hash: "sfhadfhlahdflsh",
            },
            isNormal: false
        }

        var testExpect = {
            isNormalOutput: true,
            errMsg: 'encoding/hex: invalid byte: U+0073 \'s\''
        }

        testRpc(testInput, testExpect, done);
    })

    it('hash is invalid', function (done) {
        nonce = nonce + 1;
        var testInput = {
            rpcInput: {
                hash: "1c4d8ddcf2e0b41b87df09684a@94468d7750e87b72f5e77320e0278aa05170c",
            },
            isNormal: false
        }

        var testExpect = {
            isNormalOutput: true,
            errMsg: "encoding/hex: invalid byte: U+0040 '@'"
        }

        testRpc(testInput, testExpect, done);
    })
});