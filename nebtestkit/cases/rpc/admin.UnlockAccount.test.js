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

function testUnlockAccount(testInput, testExpect, done) {
    try {
        client.UnlockAccount(testInput.args, (err, resp) => {
            try {
                expect(!!err).to.equal(testExpect.hasError);

                if (err) {
                    console.log("call return error: " + JSON.stringify(err));
                    expect(err).have.property('details').equal(testExpect.errorMsg);
                } else {
                    console.log("call return success: " + JSON.stringify(resp));
                    expect(resp).to.have.property('result').equal(testExpect.result);
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

describe("rpc: UnlockAccount", () => {
    before((done) => {
        client = rpc_client.new_client(server_address, 'AdminService');

        client.NewAccount({passphrase: "passphrase"}, (err, resp) => {
            try {
                expect(!!err).to.be.false;
                expect(resp).to.have.property('address');
                address = resp.address;
                console.log("create new account: " + address);
                done();
            } catch(err) {
                done(err)
            }
        });
    });

    it("1. normal", (done) => {
        var testInput = {
            args: {
                address: address,
                passphrase: "passphrase",
            }
        }

        var testExpect = {
            result: true,
            hasError: false
        }

        testUnlockAccount(testInput, testExpect, done);
    });

    it("2. wrong `passphrase`", (done) => {
        var testInput = {
            args: {
                address: address,
                passphrase: "wrwrwqweqw"
            }
        }

        var testExpect = {
            hasError: true,
            errorMsg: "could not decrypt key with given passphrase"
        }

        testUnlockAccount(testInput, testExpect, done);
    });

    it("3. empty `passphrase`", (done) => {
        var testInput = {
            args: {
                address: address,
                passphrase: ""
            }
        }

        var testExpect = {
            hasError: true,
            errorMsg: "passphrase is invalid"
        }

        testUnlockAccount(testInput, testExpect, done);
    });

    it("4. nonexistent `address`", (done) => {
        var testInput = {
            args: {
                address: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8",
                passphrase: "passphrase"
            }
        }

        var testExpect = {
            hasError: true,
            errorMsg: "address not find"
        }

        testUnlockAccount(testInput, testExpect, done);
    });

    it("5. invalid `address`", (done) => {
        var testInput = {
            args: {
                address: "eb31ad2d8a89a0ca693425730430bc2d63f2573b8",   // same with ""
                passphrase: "passphrase",
            }
        }

        var testExpect = {
            hasError: true,
            errorMsg: "address: invalid address"
        }

        testUnlockAccount(testInput, testExpect, done);
    });

    it("6. missing `address`", (done) => {
        var testInput = {
            args: {
                // address: address,
                passphrase: "passphrase"
            }
        }

        var testExpect = {
            hasError: true,
            errorMsg: 'address: invalid address'
        }

        testUnlockAccount(testInput, testExpect, done);
    });

    it("7. missing `passphrase`", (done) => {
        var testInput = {
            args: {
                address: address,
                // passphrase: "passphrase"
            }
        }

        var testExpect = {
            hasError: true,
            errorMsg: 'passphrase is invalid'
        }

        testUnlockAccount(testInput, testExpect, done);
    });

    it("8. redundant param", (done) => {
        var testInput = {
            args: {
                address: address,
                passphrase: "passphrase",
                test: "gtes"
            }
        }

        var testExpect = {
            callFailed: true,
            errorMsg: 'Error:'
        }

        testUnlockAccount(testInput, testExpect, done);
    });
});