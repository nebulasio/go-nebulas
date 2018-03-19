'use strict';

var expect = require('chai').expect;
var process = require('child_process');

var neb = './neb';
var newAccount;

function countSubstr(str, substr) {
    var reg = new RegExp(substr, "g");
    return str.match(reg) ? str.match(reg).length : 0;//若match返回不为null，则结果为true，输出match返回的数组(["test","test"])的长度  
}

describe('neb dump', () => {

    it('neb dump', (done) => {
        process.exec(neb + ' dump 1', (err, stdout, stderr) => {
            try {
                expect(err).to.be.equal(null);
                expect(stderr).to.be.equal("");
                var expectStdout = /blockchain dump: \[\{"height": [0-9]+, "hash": "([0-9]|[a-f]){64}"/;
                expect(expectStdout.test(stdout)).to.be.equal(true);
                newAccount = stdout.slice(9, 57);
            } catch (err) {
                done(err);
                return;
            }
            done();
        });
    });

    it('neb dump 2', (done) => {
        process.exec(neb + ' dump 2', (err, stdout, stderr) => {
            try {
                expect(err).to.be.equal(null);
                expect(stderr).to.be.equal("");
                //   console.log(stdout);
                var dump_str = /"height": [0-9]+, "hash": "([0-9]|[a-f]){64}"/;
                var dump_count = countSubstr(stdout, dump_str);

                expect(dump_count).to.be.equal(2);
                newAccount = stdout.slice(9, 57);
            } catch (err) {
                done(err);
                return;
            }
            done();
        });
    })
});
