'use strict';

var expect = require('chai').expect;
var rpc_client = require('./rpc_client/rpc_client.js');


var coinbase,
    server_address;

var env = process.env.NET || 'local';
if (env === 'local') {
    server_address = 'localhost:8684';
    coinbase = "n1QZMXSZtW7BUerroSms4axNfyBGyFGkrh5";
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
    client.SignTransactionWithPassphrase(testInput, (err, resp) => {
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

describe("rpc: SignTransaction with passphrase", () => {
    before((done) => {
        client = rpc_client.new_client(server_address, 'AdminService');

        try {
            client.NewAccount({passphrase: "passphrase"}, (err, resp) => {
                expect(!!err).to.be.false;
                expect(resp).to.have.property('address');
                address = resp.address;
                console.log("create new account: " + address);
                done();
            });
        } catch(err) {
            done(err)
        }
    });

    it('1. normal', done => {
        var testInput = {
            transaction: {
                from: address,
                to: coinbase,
                value: "123",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "1000000",
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: false,
            errorMsg: "",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('2.  `to` illegal', done => {
        var testInput = {
             transaction: {
                from: address,
                to: "faaaa",
                value: "",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: true,
            errorMsg: "address: invalid address format",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('3.  `to` empty', done => {
        var testInput = {
            transaction: {
                from: address,
                to: "",
                value: "",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: true,
            errorMsg: "address: invalid address format",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('4.  `value` empty', done => {
        var testInput = {
            transaction: {
                from: address,
                to: coinbase,
                value: "",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: true,
            errorMsg: "invalid value",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('5.  `value` alpha', done => {
        var testInput = {
            transaction: {
                from: address,
                to: coinbase,
                value: "abc",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: true,
            errorMsg: "invalid value",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('6.  `value` neg number', done => {
        var testInput = {
            transaction: {
                from: address,
                to: coinbase,
                value: "-123",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: true,
            errorMsg: "invalid value",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('7.  `nonce` alpha', done => {
        var testInput = {
            transaction: {
                from: address,
                to: coinbase,
                value: "123",
                nonce: 1, 
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: true,
            errorMsg: "params error",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('8.  `nonce` neg number', done => {
        var testInput = {
            transaction: {
                from: address,
                to: coinbase,
                value: "123",
                nonce: "-10000", 
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: true,
            errorMsg: "params error",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('9.  `gas_price` empty', done => {
        var testInput = {
            transaction: {
                from: address,
                to: coinbase,
                value: "123",
                nonce: "10000",
                gas_price: "",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: true,
            errorMsg: "invalid gasPrice",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('10.  `gas_price` alpha', done => {
        var testInput = {
            transaction: {
                from: address,
                to: coinbase,
                value: "123",
                nonce: "10000",
                gas_price: "abcxz",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: true,
            errorMsg: "invalid gasPrice",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('11.  `gas_price` negative', done => {
        var testInput = {
            transaction: {
                from: address,
                to: coinbase,
                value: "123",
                nonce: "10000",
                gas_price: "-10000",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: true,
            errorMsg: "invalid gasPrice",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('12.  `gas_limit` empty', done => {
        var testInput = {
            transaction: {
                from: address,
                to: coinbase,
                value: "123",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "",
                contract: {}
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: true,
            errorMsg: "invalid gasLimit",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('13.  `gas_limit` alpha', done => {
        var testInput = {
            transaction: {
                from: address,
                to: coinbase,
                value: "123",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "aaz",
                contract: {}
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: true,
            errorMsg: "invalid gasLimit",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('14.  `gas_limit` negative', done => {
        var testInput = {
            transaction: {
                from: address,
                to: coinbase,
                value: "123",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "-10000",
                contract: {}
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: true,
            errorMsg: "invalid gasLimit",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('15. `contract` empty', done => {
        var testInput = {
            transaction: {
                from: address,
                to: coinbase,
                value: "123",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {}
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: true,
            errorMsg: "params error",

        }

        testSignTransaction(testInput, testExpect, done)
    });

    it('16. `contract`', done => {
        var testInput = {
            transaction: {
                from: address,
                to: coinbase,
                value: "123",
                nonce: "10000",
                gas_price: "1000000",
                gas_limit: "1000000",
                contract: {
                    "function": "save",
                }
            },
            passphrase: 'passphrase'
        }

        var testExpect = {
            hasError: false,
            errorMsg: "",

        }

        testSignTransaction(testInput, testExpect, done)
    });
});