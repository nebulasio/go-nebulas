'use strict';

var Wallet = require("../../../../cmd/console/neb.js/lib/wallet");
var HttpRequest = require("../../../node-request");
var Account = Wallet.Account;

var expect = require('chai').expect;
var neb = new Wallet.Neb();

var ChainID,
    accountAddr;
var coinbase = n1QZMXSZtW7BUerroSms4axNfyBGyFGkrh5;

var env = process.env.NET || 'local';
if (env === 'local') {
    neb.setRequest(new HttpRequest("http://localhost:8685"));//https://testnet.nebulas.io
	ChainID = 100;
} else if (env === 'testneb1') {
    neb.setRequest(new HttpRequest("http://35.182.48.19:8685"));
    ChainID = 1001;
} else if (env === "testneb2") {
    neb.setRequest(new HttpRequest("http://34.205.26.12:8685"));
    ChainID = 1002;
}

function newAccount(testInput, done) {
    neb.admin.newAccount(testInput.passphrase).then(resp => {
        expect(resp).to.have.property('address');
        accountAddr = resp.address;
        console.log("new account: " + JSON.stringify(resp));
        done();
    }).catch(err => done(err));
}

function testUnlockAccount(testInput, testExpect, done) {
    
    neb.admin.unlockAccount(testInput.address, testInput.passphrase).then((resp) => {
        if (testInput.canExecute) {
            expect(resp).to.have.property("result").equal(testExpect.result);
        }
        done();
    }).catch(err => {
       if (!testInput.canExecute) {
            expect(err.error.error).to.equal(testExpect.errorMsg);
        } else {
            done(err);
        }
        done();
    }).catch(err => done(err));
}

describe('UnlockAccount', () => {
    before(done => {
        newAccount({
            passphrase: "passphrase"
        }, done);
    });
    
    it("1. normal", (done) => {
        var testInput = {
            address: accountAddr,
            passphrase: "passphrase",
            canExecute: true
        }

        var testExpect = {
            result: true,
            errorMsg: ""
        }

        testUnlockAccount(testInput, testExpect, done);
    });

    it("2. illegal passphrase", (done) => {
        var testInput = {
            address: accountAddr,
            passphrase: "wrwrwqweqw",
            canExecute: false
        }

        var testExpect = {
            result: false,
            errorMsg: "could not decrypt key with given passphrase"
        }

        testUnlockAccount(testInput, testExpect, done);
    });

    it("3. empty passphrase", (done) => {
        var testInput = {
            address: accountAddr,
            passphrase: "",
            canExecute: false
        }

        var testExpect = {
            result: false,
            errorMsg: "passphrase is invalid"
        }

        testUnlockAccount(testInput, testExpect, done);
    });

    it("4. nonexistent account", (done) => {
        var testInput = {
            address: Wallet.Account.newAccount(),
            passphrase: "passphrase",
            canExecute: true
        }

        var testExpect = {
            errorMsg: "address not find"
        }

        testUnlockAccount(testInput, testExpect, done);
    });

    it("5. invalid address", (done) => {
        var testInput = {
            address: "eb31ad2d8a89a0ca693425730430bc2d63f2573b8",   // same with ""
            passphrase: "passphrase",
            canExecute: false
        }

        var testExpect = {
            errorMsg: "address: invalid address format"
        }

        testUnlockAccount(testInput, testExpect, done);
    });

});