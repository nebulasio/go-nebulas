'use strict';

var expect = require('chai').expect;
var rpc_client = require('./rpc_client/rpc_client.js');


var coinbase,
    server_address;

var env = process.env.NET || 'local';
if (env === 'local') {
    server_address = 'localhost:8684';
    coinbase = "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8";
} else if (env === 'testneb1') {
    server_address = '35.182.48.19:8684';
    coinbase = "0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d";
} else if (env === "testneb2") {
    server_address = "34.205.26.12:8684";
    coinbase = "0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d";
}

var client,
    address;

function testSignTransaction(testInput, testExpect, done) {
    client.SignTransaction(testInput.args, (err, resp) => {
        try {
            expect(!!err).to.equal(testExpect.hasError);

            if (err) {
                console.log(JSON.stringify(err));
                expect(err).have.property('details').equal(testExpect.errorMsg);
            } else {
                console.log(JSON.stringify(resp));
                expect(resp).to.have.property('data');
            }
            done();
        } catch (err) {
            done(err);
        }
    });
}

describe("rpc: SignTransaction", () => {
    before((done) => {
        client = rpc_client.new_client(server_address, 'AdminService');

        try {
            client.NewAccount({passphrase: "passphrase"}, (err, resp) => {
                expect(!!err).to.be.false;
                expect(resp).to.have.property('address');
                address = resp.address;
                console.log("create new account: " + address);

                client.UnlockAccount({
                        address: address,
                        passphrase: "passphrase"
                    }, (err, resp) => {
                        try {
                            expect(resp).to.have.property('result').equal(true);
                            done();
                        } catch(err) {
                            console.log("unlock account failed");
                            done(err);
                        }
                });
            });
        } catch(err) {
            done(err)
        }
    });

    it('1. normal', done => {
        var testInput = {
            args: {
                from: address,
                to: "0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d",
                value: "123",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {}
            }
        }

        var testExpect = {
            hasError: false,
            errorMsg: "",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('2.  `to`', done => {
        var testInput = {
            args: {
                from: address,
                to: "faaaa",
                value: "",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {}
            }
        }

        var testExpect = {
            hasError: true,
            errorMsg: "address: invalid address",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('3.  `value', done => {
        var testInput = {
            args: {
                from: address,
                to: "0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d",
                value: "100",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {}
            }
        }

        var testExpect = {
            hasError: false,
            errorMsg: "",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('4.  `nonce`', done => {
        var testInput = {
            args: {
                from: address,
                to: "0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d",
                value: "123",
                nonce: "abasdx", // ""  --- error
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {}
            }
        }

        var testExpect = {
            hasError: false,
            errorMsg: "",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('5.  `gas_price`', done => {
        var testInput = {
            args: {
                from: address,
                to: "0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d",
                value: "123",
                nonce: "10000",
                gas_price: "abcxz",
                gas_limit: "1000000",
                contract: {}
            }
        }

        var testExpect = {
            hasError: false,
            errorMsg: "",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('6.  `gas_limit`', done => {
        var testInput = {
            args: {
                from: address,
                to: "0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d",
                value: "123",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "aaz",
                contract: {}
            }
        }

        var testExpect = {
            hasError: false,
            errorMsg: "",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('7. `contract`', done => {
        var testInput = {
            args: {
                from: address,
                to: "0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d",
                value: "123",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {
                    "function": "save",
                }
            }
        }

        var testExpect = {
            hasError: false,
            errorMsg: "",

        }

        testSignTransaction(testInput, testExpect, done)
    });
});