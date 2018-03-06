'use strict';

var expect = require('chai').expect;
var rpc_client = require('./rpc_client/rpc_client.js');


var coinbase,
    client,
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

function testGetGasPrice(testInput, testExpect, done) {

    try {
        client.GetGasPrice({}, (err, resp) => {
            try {
                expect(!!err).to.equal(testExpect.hasError);

                if (err) {
                    console.log("call return err: " + JSON.stringify(err));
                    expect(err).have.property('details').equal(testExpect.errorMsg);
                } else {
                    console.log("call return success: " + JSON.stringify(resp));
                    expect(resp).to.have.property('gas_price');
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

describe("rpc: GetGasPrice", () => {
    before(() => {
        client = rpc_client.new_client(server_address);
    });

    it('1. normal', done => {
        var testInput = {
            
        }

        var testExpect = {
            hasError: false,
            errorMsg: ''
        }

        testGetGasPrice(testInput, testExpect, done)
    });

    it('2. redundant params', done => {
        var testInput = {
            height: "3243"
        }

        var testExpect = {
            hasError: false,
            errorMsg: ''
        }

        testGetGasPrice(testInput, testExpect, done)
   
 });
});