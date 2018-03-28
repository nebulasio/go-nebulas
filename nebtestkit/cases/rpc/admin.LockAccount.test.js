'use strict';

var expect = require('chai').expect;
var rpc_client = require('./rpc_client/rpc_client.js');
var Wallet = require("nebulas");

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

function testLockAccount(testInput, testExpect, done) {

    try {
        client.LockAccount(testInput.args, (err, resp) => {
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

describe("rpc: LockAccount", () => {
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
                address: address
            }
        }

        var testExpect = {
            result: true,
            hasError: false
        }

        // unlock first
        client.UnlockAccount({
                address: address,
                passphrase: "passphrase"
            }, (err, resp) => {
                try {
                    expect(resp).to.have.property('result').equal(true);
                    testLockAccount(testInput, testExpect, done);
                } catch(err) {
                    console.log("unlock account failed");
                    done(err);
                }
        });
    });

    it("2. not unlocked", (done) => {
        var testInput = {
            args: {
                address: address
            }
        }

        var testExpect = {
            hasError: true,
            errorMsg: "key not unlocked"
        }

        testLockAccount(testInput, testExpect, done);
    });

    it("3. empty `address`", (done) => {
        var testInput = {
            args: {
                address: ""
            }
        }

        var testExpect = {
            hasError: true,
            errorMsg: "address: invalid address format"
        }

        testLockAccount(testInput, testExpect, done);
    });

    it("4. nonexistent `address`", (done) => {
        var testInput = {
            args: {
                address: "key not unlocked"
            }
        }

        var testExpect = {
            hasError: true,
            errorMsg: "address: invalid address format"
        }

        testLockAccount(testInput, testExpect, done);
    });

    it("5. invalid `address`", (done) => {
        var testInput = {
            args: {
                address: "n1QZMXSZtW7BUerroSms4axNfyBGyFGkrh3" // same with ""
            }
        }

        var testExpect = {
            hasError: true,
            errorMsg: "address: invalid address checksum"
        }

        testLockAccount(testInput, testExpect, done);
    });
});