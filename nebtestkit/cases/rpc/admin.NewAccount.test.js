'use strict';

var expect = require('chai').expect;
var rpc_client = require('./rpc_client/rpc_client.js');
var Wallet = require('nebulas');

var coinbase,
    sourceAccount,
    chain_id,
    server_address;

var env = process.env.NET || 'local';
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

var client;

function testNewAccount(testInput, testExpect, done) {

    try {
        client.NewAccount(testInput.args, (err, resp) => {
            try {
                expect(!!err).to.equal(testExpect.hasError);

                if (err) {
                    console.log("call return err: " + JSON.stringify(err));
                    expect(err).have.property('details').equal(testExpect.errorMsg);
                } else {
                    console.log("call return success: " + JSON.stringify(resp));
                    expect(resp).to.have.property('address');
                }
                done();
            } catch (err) {
                done(err);
            }
        });
    } catch(err) {
        console.log("call failed:" + err.toString())
        if (testExpect.callFailed) {
            try {
                expect(err.toString()).to.have.string(testExpect.errorMsg);
                done();
            } catch(er) {
                done(er);
            }
        } else {
            done(err)
        }
    }
}

describe("rpc: NewAccount", () => {
    before(() => {
        client = rpc_client.new_client(server_address, 'AdminService');
    });

    it("1. legal `passphrase` (len>0)", (done) => {
        
        var testInput = {
            args: {
                passphrase: "passphrase"
            }
        }

        var testExpect = {
            hasError: false
        }
        
        testNewAccount(testInput, testExpect, done);
    });

    it("2. empty `passphrase`", (done) => {
        var testInput = {
            args: {
                passphrase: ""
            }
        }

        var testExpect = {
            hasError: true,
            errorMsg: "passphrase is invalid"
        }

        testNewAccount(testInput, testExpect, done);
    });
    
    it("3. redundant param", (done) => {
        var testInput = {
            args: {
                fsa: "",
                more1: "more1"
            }
        }

        var testExpect = {
            callFailed: true,
            errorMsg: "Error: "
        }

        testNewAccount(testInput, testExpect, done);
    }); 
});