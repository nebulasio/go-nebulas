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

function testLockAccount(testInput, testExpect, done) {

    client.LockAccount(testInput.args, (err, resp) => {
        try {
            // console.log(JSON.stringify(err));
            // console.log(JSON.stringify(resp));
            expect(!!err).to.equal(testExpect.hasError);

            if (err) {
                console.log(JSON.stringify(err));
                expect(err).have.property('details').equal(testExpect.errorMsg);
            } else {
                console.log(JSON.stringify(resp));
                expect(resp).to.have.property('result').equal(testExpect.result);
            }
            done();
        } catch (err) {
            done(err);
        }
    });
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

    it("3. nonexistent account", (done) => {
        var testInput = {
            args: {
                address: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8"
            }
        }

        var testExpect = {
            hasError: true,
            errorMsg: "key not unlocked"
        }

        testLockAccount(testInput, testExpect, done);
    });

    it("4. invalid address", (done) => {
        var testInput = {
            args: {
                address: "eb31ad2d8a89a0ca695730430bc2d63f2573b8" // same with ""
            }
        }

        var testExpect = {
            hasError: true,
            errorMsg: "address: invalid address"
        }

        testLockAccount(testInput, testExpect, done);
    });
});