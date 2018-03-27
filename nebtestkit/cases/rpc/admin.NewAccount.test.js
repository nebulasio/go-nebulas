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