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

describe('neb config', () => {

    it('neb config new', (done) => {
        process.exec(neb + ' config new ./config.tmp', (err, stdout, stderr) => {
            try {
                expect(err).to.be.equal(null);
                expect(stderr).to.be.equal("");

                var genesisConf = FS.readFileSync('./config.tmp','utf-8');
                console.log(stdout);
                console.log(genesisConf);
                
                expect(stdout).to.be.equal('create default config ./config.tmp\n');
                expect(genesisConf).to.be.equal('\n\tnetwork {\n\t\tlisten: ["127.0.0.1:8680"]\n\t}\n\n\tchain '+
                    '{\n\t\tchain_id: 100\n\t\tdatadir: "data.db"\n\t\tgenesis: "conf/default/genesis.conf"\n\t'+
                    '\tkeydir: "keydir"\n\t\tcoinbase: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8"\n\t'+
                    '\tsignature_ciphers: ["ECC_SECP256K1"]\n\t}\n\n\trpc {\n\t\trpc_listen: ["127.0.0.1:8684"]\n'+
                    '\t\thttp_listen: ["127.0.0.1:8685"]\n\t\thttp_module: ["api","admin"]\n\t}\n\n  \tapp {\n\t' + 
                    '\tlog_level: "info"\n    \tlog_file: "logs"\n    \tenable_crash_report: false\n  \t}\n\n\t' + 
                    'stats {\n\t\tenable_metrics: false\n\t}\n\t');
            } catch (err) {
                done(err);
                return;
            }
            done();
        });
    });
    after('rm tmp file', (done) => {
        process.exec('rm -f ./config.tmp', (err, stdout, stderr) => {
            try {
                expect(err).to.be.equal(null);
                expect(stderr).to.be.equal("");
            } catch(err) {
                done(err);
                return;
            }
            done();
        })
    })
});
