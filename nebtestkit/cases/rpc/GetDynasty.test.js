'use strict';

var expect = require('chai').expect;
var rpc_client = require('./rpc_client/rpc_client.js');


var coinbase,
    client,
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

function testGetDynasty(testInput, testExpect, done) {

    try {
        client.GetDynasty(testInput, (err, resp) => {
            try {
                expect(!!err).to.equal(testExpect.hasError);

                if (err) {
                    console.log("call return err: " + JSON.stringify(err));
                    expect(err).have.property('details').equal(testExpect.errorMsg);
                } else {
                    console.log("call return success: " + JSON.stringify(resp));
                    expect(resp).to.have.property('miners');
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

describe("rpc: GetDynasty", () => {
    before(() => {
        client = rpc_client.new_client(server_address);
    });

    it('1. `height` missing', done => {
        var testInput = {
            
        }

        var testExpect = {
            hasError: false,
            errorMsg: ''
        }

        testGetDynasty(testInput, testExpect, done)
    });

    it('2. `height` empty string', done => {
        var testInput = {
            height: ""
        }

        var testExpect = {
            hasError: true,
            errorMsg: 'Error:'
        }

        testGetDynasty(testInput, testExpect, done)
    });

    it('3. `height` 0', done => {
        var testInput = {
            height: 0
        }

        var testExpect = {
            hasError: false,
            errorMsg: ''
        }

        testGetDynasty(testInput, testExpect, done)
    });

    it('4. `height` negative', done => {
        var testInput = {
            height: -100000
        }

        var testExpect = {
            hasError: false,
            errorMsg: ''
        }

        testGetDynasty(testInput, testExpect, done)
    });

    it('5. `height` bigger than now', done => {
        var testInput = {
            height: 10000000000
        }

        var testExpect = {
            hasError: true,
            errorMsg: 'block not found'
        }

        testGetDynasty(testInput, testExpect, done)
    });
});