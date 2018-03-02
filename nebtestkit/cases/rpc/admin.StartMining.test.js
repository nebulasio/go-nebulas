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

function testStartMining(testInput, testExpect, done) {

    client.StartMining(testInput, (err, resp) => {
        try {
            expect(!!err).to.equal(testExpect.hasError);

            if (err) {
                console.log(JSON.stringify(err));
                expect(err).have.property('details').equal(testExpect.errorMsg);
            } else {
                console.log(JSON.stringify(resp));
                console.log("start success.")
                expect(resp).to.have.property('result').equal(testExpect.result);
            }
            done();
        } catch (err) {
            done(err);
        }
    });
}

describe("rpc: StartMining", () => {
    before((done) => {
        client = rpc_client.new_client(server_address, 'AdminService');
        try {
            client.StartMining({passphrase: "passphrase"}, (err, resp) => {
                done();
            });
        }catch (err) {
            done(err)
        }
    });

    describe('already started', () => {
        it('1. invalid passphrase', done => {
            var testInput = {
                passphrase: "absad",
            }

            var testExpect = {
                result: true,
                hasError: true,
                errorMsg: 'consensus has already been started'
            }

            testStartMining(testInput, testExpect, done);
        });

        it('2. empty passphrase', done => {
            var testInput = {
                passphrase: "",
            }

            var testExpect = {
                result: true,
                hasError: true,
                errorMsg: 'consensus has already been started'
            }

            testStartMining(testInput, testExpect, done);
        });

        it('3. valid passphrase', done => {
            var testInput = {
                passphrase: "passphrase",
            }

            var testExpect = {
                result: true,
                hasError: true,
                errorMsg: 'consensus has already been started'
            }

            testStartMining(testInput, testExpect, done);
        });
    });

    describe('already stopped', () => {
        beforeEach(done => {
            try {
                client.StopMining({}, (err, resp) => {
                    
                    done();
                });
            }catch (err) {
                done(err)
            }
        });

        it('4. invalid passphrase', done => {
            var testInput = {
                passphrase: "absad",
            }

            var testExpect = {
                result: true,
                hasError: true,
                errorMsg: 'could not decrypt key with given passphrase'
            }

            testStartMining(testInput, testExpect, done);
        });

        it('5. empty passphrase', done => {
            var testInput = {
                passphrase: "",
            }

            var testExpect = {
                result: true,
                hasError: true,
                errorMsg: 'passphrase is invalid'
            }

            testStartMining(testInput, testExpect, done);
        });

        it('6. valid passphrase', done => {
            var testInput = {
                passphrase: "passphrase",
            }

            var testExpect = {
                result: true,
                hasError: false,
                errorMsg: 'consensus has already been started'
            }

            testStartMining(testInput, testExpect, done);
        });
    });
});