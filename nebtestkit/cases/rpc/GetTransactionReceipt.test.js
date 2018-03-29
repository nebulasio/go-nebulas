'use strict';
var expect = require('chai').expect;
var rpc_client = require('./rpc_client/rpc_client.js');
var Wallet = require("nebulas");

var protocol_version = '/neb/1.0.0'
var node_version = '0.7.0'
var server_address = 'localhost:8684';
var coinbase = "n1QZMXSZtW7BUerroSms4axNfyBGyFGkrh5";
var sourceAccount = new Wallet.Account('d80f115bdbba5ef215707a8d7053c16f4e65588fd50b0f83369ad142b99891b5');
var chain_id = 100;
var env = 'maintest';
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
