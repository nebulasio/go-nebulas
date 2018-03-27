'use strict';

var Wallet = require("nebulas");

var expect = require('chai').expect;
var rpc_client = require('./rpc_client/rpc_client.js');


var coinbase,
    sourceAccount,
    toAddress = Wallet.Account.NewAccount(),
    txhash,
    client,
    server_address;

var env = process.env.NET || 'local';
if (env === 'local') {
    server_address = 'localhost:8684';
    coinbase = "n1QZMXSZtW7BUerroSms4axNfyBGyFGkrh5";
    sourceAccount = new Wallet.Account("d80f115bdbba5ef215707a8d7053c16f4e65588fd50b0f83369ad142b99891b5");
} else if (env === 'testneb1') {
    server_address = '35.182.48.19:8684';
    coinbase = "0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d";
    sourceAccount = new Wallet.Account("43181d58178263837a9a6b08f06379a348a5b362bfab3631ac78d2ac771c5df3");
} else if (env === "testneb2") {
    server_address = "34.205.26.12:8684";
    coinbase = "0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d";
    sourceAccount = new Wallet.Account("43181d58178263837a9a6b08f06379a348a5b362bfab3631ac78d2ac771c5df3");
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