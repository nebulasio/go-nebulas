'use strict';

var Wallet = require("nebulas");

var expect = require('chai').expect;
var rpc_client = require('./rpc_client/rpc_client.js');


var coinbase,
    sourceAccount,
    toAddress = Wallet.Account.NewAccount(),
    txhash,
    client,
    chain_id,
    server_address;

var env = process.env.NET || 'local';
env = "maintest";
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

function testGetEventsByHash(testInput, testExpect, done) {
    try {
        client.GetEventsByHash(testInput, (err, resp) => {
            try {
                expect(!!err).to.equal(testExpect.hasError);

                if (err) {
                    console.log("call return err: " + JSON.stringify(err));
                    expect(err).have.property('details').string(testExpect.errorMsg);
                } else {
                    console.log("call return success: " + JSON.stringify(resp));
                    expect(resp).to.have.property('events');
                }
                done();
            } catch (err) {
                done(err);
            }
        });
    } catch(err) {
        if (testExpect.hasError) {
            try {
                expect(err.toString()).to.have.string(testExpect.errorMsg);
                done()
                return;
            } catch(er) {}
        } 
        done(err)
    }
}

function checkTx(hash, done) {
    client.GetTransactionReceipt({hash: txhash}, (err, resp) => {
        console.log("err: " + JSON.stringify(err))
        console.log("resp: " + JSON.stringify(resp))
        
        if ((err && expect(err).to.have.property("details").equal("transaction not found")) || resp.status == 2) {
            setTimeout(() => {
                checkTx(hash, done)
            }, 2000); 
            return;
        } 
        if (resp.status == 1) {
            console.log("tx send done: " + JSON.stringify(resp));
            done();
        } else {
            done(new Error("tx failed"));
        }
    });
}

describe("rpc: GetEventsByHash", () => {
    before((done) => {
        client = rpc_client.new_client(server_address);

        try {
            /* client.NewAccount({passphrase: "passphrase"}, (err, resp) => {
                expect(!!err).to.be.false;
                expect(resp).to.have.property('address');
                toAddress = resp.address;
                console.log("create new `to` account: " + address); */
                
                client.GetAccountState({address: sourceAccount.getAddressString()}, (err, resp) => {
                    console.log("sourceAccount state: " + JSON.stringify(resp));
                    var coinbaseNonce = resp.nonce;

                    var adminclient = rpc_client.new_client(server_address, 'AdminService');
                    adminclient.SendTransactionWithPassphrase({
                        transaction: {
                            from: sourceAccount.getAddressString(),
                            to: toAddress.getAddressString(),
                            value: "100000000",
                            nonce: parseInt(coinbaseNonce) + 1,
                            gas_price: "1000000",
                            gas_limit: "2000000"
                        },
                        passphrase: 'passphrase'
                    }, (err, resp) => {
                        if (err) {
                            console.log("send tx error: " + JSON.stringify(err));
                            done(err)
                            return;
                        }
                        console.log("send tx response: " + JSON.stringify(resp));
                        try {
                            expect(resp).to.have.property('txhash');
                            txhash = resp.txhash;

                            // check tx done
                            checkTx(txhash, done);
                        } catch (err) {
                            done(err);
                        }
                    });
                });
            // });
        } catch(err) {
            done(err)
        }
    });

    it('1. nonexistent `hash`', done => {
        var testInput = {
            hash: "02930f09029f0f"
        }

        var testExpect = {
            hasError: true,
            errorMsg: 'invalid argument(s)'
        }

        testGetEventsByHash(testInput, testExpect, done);
    });

    it('2. odd length `hash`', done => {
        var testInput = {
            hash: "02930f09029ff"
        }

        var testExpect = {
            hasError: true,
            errorMsg: 'encoding/hex: odd length hex string'
        }

        testGetEventsByHash(testInput, testExpect, done);
    });

    it('3. non-hexadecimal char `hash`', done => {
        var testInput = {
            hash: "02930fg09029ff"
        }

        var testExpect = {
            hasError: true,
            errorMsg: 'encoding/hex: invalid byte'
        }

        testGetEventsByHash(testInput, testExpect, done);
    });

    it('4. empty `hash`', done => {
        var testInput = {
            hash: ""
        }

        var testExpect = {
            hasError: true,
            errorMsg: 'please input valid hash'
        }

        testGetEventsByHash(testInput, testExpect, done);
    });

    it('5. missing `hash`', done => {
        var testInput = {}

        var testExpect = {
            hasError: true,
            errorMsg: 'please input valid hash'
        }

        testGetEventsByHash(testInput, testExpect, done);
    });

    it('6. normal `hash`', done => {
        var testInput = {
            hash: txhash
        }

        var testExpect = {
            hasError: false,
            errorMsg: ''
        }

        testGetEventsByHash(testInput, testExpect, done);
    });
})