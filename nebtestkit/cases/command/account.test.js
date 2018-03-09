'use strict';

var expect = require('chai').expect;
var process = require('child_process');

var neb = './neb';
var newAccount;

describe('neb account', () => {

    it('account new', (done) => {
        process.exec(neb + ' account new 123456', (err, stdout, stderr) => {
            try {
                expect(err).to.be.equal(null);
                expect(stderr).to.be.equal("");
                var expectStdout = /Address: ([0-9]|[a-f]){48}\n$/;
                expect(expectStdout.test(stdout)).to.be.equal(true);
                newAccount = stdout.slice(9, 57);
            } catch (err) {
                done(err);
                return;
            }
            done();
        });
    });

    it('account list', (done) => {
        process.exec(neb + ' account list', (err, stdout, stderr) => {
            try {
                expect(err).to.be.equal(null);
                expect(stderr).to.be.equal("");
                var expectStdout = /^(Account #[0-9]+: ([0-9]|[a-f]){48}\n)+$/
                expect(expectStdout.test(stdout)).to.be.equal(true);
            } catch(err) {
                done(err);
                return;
            }
            done();
        });
    });

    it('account update', (done) => {
        var child = process.exec(neb + ' account update ' + newAccount, (err, stdout, stderr) => {
            try {
                expect(err).to.be.equal(null);
                expect(stderr).to.be.equal("");
                console.log(stdout);
                var expectStdout = 'Please input current passhprase\nPassphrase: \n'+
                    'Please give a new password. Do not forget this password.\n'+
                    'Passphrase: \nRepeat passphrase: \nUpdated address: ' + newAccount +'\n';
                expect(expectStdout).to.be.equal(stdout);
            } catch(err) {
                done(err);
                return;
            }
            done();
        });
        child.stdin.write("123456\n123\n123\n");
    });

    it('account import', (done) => {
        var child = process.exec(neb + ' account import keydir/' + newAccount, (err, stdout, stderr) => {
            try {
                expect(err).to.be.equal(null);
                expect(stderr).to.be.equal("");
                var expectStdout = 'Passphrase: \nImport address: ' + newAccount + '\n';
                expect(expectStdout).to.be.equal(stdout);
            } catch (err) {
                done(err);
                return;
            }
            done();
        });
        child.stdin.write('123\n');
    });

    after('remove the accout file', (done) => {
        var child = process.exec('rm -f ./keydir/' + newAccount, (err, stdout, stderr) => {
            try {
                expect(err).to.be.equal(null);
                expect(stderr).to.be.equal("");
                done();
            } catch (err) {
                done(err);
            }
        });
    });
});
