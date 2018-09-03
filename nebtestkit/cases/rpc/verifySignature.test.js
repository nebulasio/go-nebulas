'use strict';
var expect = require('chai').expect;
var rpc_client = require('./rpc_client/rpc_client.js');
var Wallet = require("nebulas");
var Account = Wallet.Account;
var protocol_version = '/neb/1.0.0'
var node_version = '0.7.0';


var server_address = 'localhost:8684';
var coinbase = "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8";
var sourceAccount;
var chain_id = 100;
var env = 'local';

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

var client;
var normalOutput;

function testRpc(testInput, testExpect, done) {
    client.verifySignature(testInput.rpcInput, function (err, response) {
        if (err != null) {
            try {
                expect(testExpect.errMsg).to.be.equal(err.details);
            } catch (err) {
                console.log(err);
                done(err)
                return;
            }
            console.log(err);
            done();
            return;
        } else {
            try {
                if (testInput.isNormal) {
                    normalOutput = response;
                    console.log("response for normal input is: " + JSON.stringify(response))
                    expect(JSON.stringify(response)).to.be.equal(JSON.stringify(testExpect.rpcOutput));
                } else {
                    if (testExpect.isNormalOutput) { //非正常输入，正常输出？
                        expect(JSON.stringify(response)).to.be.equal(JSON.stringify(normalOutput));
                    } else {
                        expect(testExpect.isNormalOutput).to.be.equal(false);
                        expect(JSON.stringify(response)).not.be.equal(JSON.stringify(normalOutput));
                        expect(response.result).to.be.a("boolean");
                        //In JS, uint64 is converted to string
                        expect(response.address).to.be.equal('n1HUbJZ45Ra5jrRqWvfVaRMiBMB3CACGhqc');
                    }
                }
            } catch (err) {
                done(err);
                return;
            };
        }
        done();
        return;
    });

}

describe('rpc: verifySignature', function () {
    before(function () {
        client = rpc_client.new_client(server_address);
    });

    it('normal rpc', function (done) {
        var testInput = {
            rpcInput: {
                msg: '9dedc6db0d895e346355f2c702a7a8e462993fee16a1ec8847b2852d49245564',
                signature: 'ee291ab49ba4ad1c5874a3842bcf02ce3e948ea0938289835eea353394297a166d8b3a93c4d10ffb115a30466c4499fd38e2288586efb36f9fb83399350d3ce600',
                address: 'n1HUbJZ45Ra5jrRqWvfVaRMiBMB3CACGhqc',
                alg: 1
            },
            isNormal: true
        }

        var testExpect = {
            isNormalOutput: true,
            rpcOutput: {
                result: true,
                address: 'n1HUbJZ45Ra5jrRqWvfVaRMiBMB3CACGhqc'
            },
        }

        testRpc(testInput, testExpect, done);
    })

    it('empty alg', function (done) {
        var testInput = {
            rpcInput: {
                msg: '9dedc6db0d895e346355f2c702a7a8e462993fee16a1ec8847b2852d49245564',
                signature: 'ee291ab49ba4ad1c5874a3842bcf02ce3e948ea0938289835eea353394297a166d8b3a93c4d10ffb115a30466c4499fd38e2288586efb36f9fb83399350d3ce600',
                address: 'n1HUbJZ45Ra5jrRqWvfVaRMiBMB3CACGhqc',
            },
            isNormal: true
        }

        var testExpect = {
            isNormalOutput: true,
            rpcOutput: {
                result: true,
                address: 'n1HUbJZ45Ra5jrRqWvfVaRMiBMB3CACGhqc'
            },
        }

        testRpc(testInput, testExpect, done);
    })

    it('error rpc', function (done) {
        var testInput = {
            rpcInput: {
                msg: '9dedc6db0d895e346355f2c702a7a8e462993fee16a1ec8847b2852d49245564',
                signature: 'ee291ab49ba4ad1c5874a3842bcf02ce3e948ea0938289835eea353394297a166d8b3a93c4d10ffb115a30466c4499fd38e2288586efb36f9fb83399350d3ce600',
                address: 'none_exist_address',
                alg: 1
            },
            isNormal: true
        }

        var testExpect = {
            isNormalOutput: true,
            rpcOutput: {
                result: false,
                address: 'n1HUbJZ45Ra5jrRqWvfVaRMiBMB3CACGhqc'
            },
        }

        testRpc(testInput, testExpect, done);
    })

    it('empty address', function (done) {
        var testInput = {
            rpcInput: {
                msg: '9dedc6db0d895e346355f2c702a7a8e462993fee16a1ec8847b2852d49245564',
                signature: 'ee291ab49ba4ad1c5874a3842bcf02ce3e948ea0938289835eea353394297a166d8b3a93c4d10ffb115a30466c4499fd38e2288586efb36f9fb83399350d3ce600',
                alg: 1
            },
            isNormal: true
        }

        var testExpect = {
            isNormalOutput: false,
            rpcOutput: {
                result: false,
                address: 'n1HUbJZ45Ra5jrRqWvfVaRMiBMB3CACGhqc'
            },
        }

        testRpc(testInput, testExpect, done);
    });

    it('invalid message: odd length', function (done) {
        var testInput = {
            rpcInput: {
                msg: '9dedc6db0d895e346355f2c702a7a8e462993fee16a1ec8847b2852d4924556', //
                signature: 'ee291ab49ba4ad1c5874a3842bcf02ce3e948ea0938289835eea353394297a166d8b3a93c4d10ffb115a30466c4499fd38e2288586efb36f9fb83399350d3ce600',
                alg: 1
            },
            isNormal: false
        }

        var testExpect = {
            isNormalOutput: false,
            errMsg: "encoding/hex: odd length hex string"
        }

        testRpc(testInput, testExpect, done);
    });

    it('invalid message: not a 32Byte', function (done) {
        var testInput = {
            rpcInput: {
                msg: '9dedc6db0d895e346355f2c702a7a8e462993fee16a1ec8847b2852d492455',
                signature: 'ee291ab49ba4ad1c5874a3842bcf02ce3e948ea0938289835eea353394297a166d8b3a93c4d10ffb115a30466c4499fd38e2288586efb36f9fb83399350d3ce600',
                alg: 1
            },
            isNormal: false
        }

        var testExpect = {
            isNormalOutput: false,
            errMsg: "invalid message length, need 32 bytes"
        }

        testRpc(testInput, testExpect, done);
    });

    it('invalid message: not a 32Byte', function (done) {
        var testInput = {
            rpcInput: {
                msg: '9dedc6db0d895e346355f2c702a7a8e462993fee16a1ec8847b2852d492455',
                signature: 'ee291ab49ba4ad1c5874a3842bcf02ce3e948ea0938289835eea353394297a166d8b3a93c4d10ffb115a30466c4499fd38e2288586efb36f9fb83399350d3ce600',
                alg: 2
            },
            isNormal: false
        }

        var testExpect = {
            isNormalOutput: false,
            errMsg: "invalid Algorithm"
        }

        testRpc(testInput, testExpect, done);
    });


});
