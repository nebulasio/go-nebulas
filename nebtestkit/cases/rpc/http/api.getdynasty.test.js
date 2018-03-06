'use strict';

var Wallet = require("../../../../cmd/console/neb.js/lib/wallet");
var HttpRequest = require("../../../node-request");
var Account = Wallet.Account;

var expect = require('chai').expect;
var neb = new Wallet.Neb();

var ChainID,
    coinbase,
    sourceAccount;

var env = process.env.NET || 'local';
if (env === 'local') {
    neb.setRequest(new HttpRequest("http://127.0.0.1:8685"));//https://testnet.nebulas.io
	ChainID = 100;
    sourceAccount = new Wallet.Account("a6e5eb290e1438fce79f5cb8774a72621637c2c9654c8b2525ed1d7e4e73653f");
    coinbase = "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8";
} else if (env === 'testneb1') {
    neb.setRequest(new HttpRequest("http://35.182.48.19:8685"));
    ChainID = 1001;
    sourceAccount = new Wallet.Account("43181d58178263837a9a6b08f06379a348a5b362bfab3631ac78d2ac771c5df3");
    coinbase = "0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d";
} else if (env === "testneb2") {
    neb.setRequest(new HttpRequest("http://34.205.26.12:8685"));
    ChainID = 1002;
    sourceAccount = new Wallet.Account("43181d58178263837a9a6b08f06379a348a5b362bfab3631ac78d2ac771c5df3");
    coinbase = "0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d";
}

function testGetDynasty(testInput, testExpect, done) {
    neb.api.getDynasty().then(resp => {
        console.log("call return success: " + JSON.stringify(resp));
        done();
    }).catch(err => {
        console.log("call return error: " + JSON.stringify(err));
        done(err);
    });
}

describe('http: GetDynasty', () => {

    it('1. normal', done => {
        var testInput = {
            height: 0
        }

        var testExpect = {

        }

        testGetDynasty(testInput, testExpect, done);
    });
});