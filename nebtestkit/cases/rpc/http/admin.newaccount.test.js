'use strict';

var Wallet = require("../../../../cmd/console/neb.js/lib/wallet");
var HttpRequest = require("../../../node-request");
var Account = Wallet.Account;

var BigNumber = require('bignumber.js');
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

function testNewAccount(testInput, testExpect, done) {
    neb.admin.newAccount(testInput.passphrase).then((resp) => {
        expect(resp).to.have.property('address');
        console.log("NewAccount('" + testInput.passphrase + "') return: " + JSON.stringify(resp));
        
        expect(Account.isValidAddress(resp.address)).to.equal(testExpect.validAddr);
        done();
    }).catch(err => {
        if (testExpect.failed) {
            expect(err.error.error).to.equal(testExpect.errorMsg);
            done();
        } else {
            done(err);
        }
    });
}

describe('NewAccount', () => {
    it("1. legal passphrase", (done) => {
        var testInput = {
            passphrase: "passphrase",
        }

        var testExpect = {
            validAddr: true,
            failed: false
        }

        testNewAccount(testInput, testExpect, done);
    });

    it("2. empty passphrase", (done) => {
        var testInput = {
            passphrase: "",
        }

        var testExpect = {
            validAddr: false,
            failed: true,
            errorMsg: "passphrase is invalid"
        }

        testNewAccount(testInput, testExpect, done);
    });

    // it("3. illegal char ", (done) => {
    //     var testInput = {
    //         passphrase: "、、||」|【」、",
    //     }

    //     var testExpect = {
    //         validAddr: false,
    //         failed: true,
    //         errorMsg: "passphrase is invalid"
    //     }

    //     testNewAccount(testInput, testExpect, done);
    // });
});