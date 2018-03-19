'use strict';

var FS = require('fs');
var expect = require('chai').expect;
var process = require('child_process');

var neb = './neb';
var newAccount;

function countSubstr(str, substr) {
    var reg = new RegExp(substr, "g");
    return str.match(reg) ? str.match(reg).length : 0;//若match返回不为null，则结果为true，输出match返回的数组(["test","test"])的长度  
}

describe('neb genesis dump', () => {

    it('neb genesis dump', (done) => {
        process.exec(neb + ' genesis dump', (err, stdout, stderr) => {
            try {
                expect(err).to.be.equal(null);
                expect(stderr).to.be.equal("");

                var genesisConf = FS.readFileSync('./conf/default/genesis.conf','utf-8');

                var pos = genesisConf.indexOf('meta');
                expect(pos).to.be.greaterThan(0);
                var expectGenesis = genesisConf.slice(pos);
                var expectGenesisTxt = expectGenesis.replace(/[^a-z0-9A-Z]/g, "");


                var genesis = stdout.slice(stdout.indexOf('\"meta\": {'));
                var genesisTxt = genesis.replace(/[^a-z0-9A-Z]/g, "");
                //console.log(genesisTxt);

                expect(genesisTxt).to.be.equal(expectGenesisTxt);
            } catch (err) {
                done(err);
                return;
            }
            done();
        });
    });
});
