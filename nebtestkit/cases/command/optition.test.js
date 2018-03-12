'use strict';

var sleep = require("system-sleep");
var process = require('child_process');
var expect = require('chai').expect;
var FS = require('fs');
var rpc_client = require('../rpc/rpc_client/rpc_client.js');
var neb = './neb';
var server_address = '127.0.0.1:8684';
var newAccount;

function countSubstr(str, substr) {
    var reg = new RegExp(substr, "g");
    return str.match(reg) ? str.match(reg).length : 0;//若match返回不为null，则结果为true，输出match返回的数组(["test","test"])的长度  
}

describe('neb options', () => {
    before('create dir.tmp', (done) => {
        process.exec('mv keydir keydir.tmp ', (err, stdout, stderr) => {
            try {
                expect(err).to.be.equal(null);
                expect(stderr).to.be.equal("");
            } catch (err) {
                done(err);
                return;
            }
            done();
        });
    }); 

    before('create network.tmp', (done) => {
        process.exec('cp conf/network/ed25519key network.tmp', (err, stdout, stderr) => {
            try {
                expect(err).to.be.equal(null);
                expect(stderr).to.be.equal("");
            } catch (err) {
                done(err);
                return;
            }
            done();
        });
    }); 

    it('neb options all', (done) => {
        var neb_process = process.exec(neb + //app.config
                                             ' --app.crashreport=false' + 
                                             ' --app.logfile logs.tmp' +
                                             ' --app.loglevel info' +
                                             ' --app.reporturl https://127.0.0.1' +
                                             ' --app.pprof.listen 127.0.0.1' +
                                             ' --app.pprof.memprofile memprofile.tmp' +
                                             ' --app.pprof.cpuprofile cpuprofile.tmp' +
                                             //block chain config
                                             ' --chain.ciphers cipher' + 
                                             ' --chain.coinbase 75e4e5a71d647298b88928d8cb5da43d90ab1a6c52d0905f' +
                                             ' --chain.gaslimit 1000' + 
                                             ' --chain.gasprice 1000' + 
                                             ' --chain.datadir data.tmp' + 
                                             ' --chain.keydir keydir.tmp' +
                                             ' --chain.startmine=false' + 
                                             //network
                                             ' --network.seed /ip4/127.0.0.1/tcp/8680/ipfs/QmP7HDFcYmJL12Ez4ZNVCKjKedfE7f48f1LAkUc3Whz4j1' +
                                             ' --network.listen 0.0.0.0:8991' + 
                                             ' --network.key network.tmp' +
                                             //rpc
                                             ' --rpc.listen 0.0.0.0:8684' +
                                             ' --rpc.http 0.0.0.0:8685' +
                                             ' --rpc.module api' +
                                             //stats
                                             ' --stats.dbhost 127.0.0.1:12345' +
                                             ' --stats.dbname test_db_123' +
                                             ' --stats.dbuser silent' +
                                             ' --stats.dbpassword 123456' +
                                             ' --stats.enable=false',

                                             
                                             (err, stdout, stderr) => {
            console.log(stdout);
            if (err.signal != 'SIGKILL') {
                console.log("unexpected error");
                console.log(err);
            }
        });
        try {
            var client = rpc_client.new_client(server_address);
        } catch (err) {
            done(err);
            return;
        }
        sleep(1000);
        try {
            client.getConfig({}, (err, resp) => {
                try {
                    expect(err).to.be.equal(null);
                    console.log(resp.config);
                    //app config
                    expect(resp.config.app.log_level).to.be.equal('info');
                    expect(resp.config.app.log_file).to.be.equal('logs.tmp');
                    expect(resp.config.app.enable_crash_report).to.be.equal(false);//enable_crash_report = false
                    expect(resp.config.app.crash_report_url).to.be.equal('https://127.0.0.1');
                    expect(resp.config.app.pprof.http_listen).to.be.equal('127.0.0.1');
                    expect(resp.config.app.pprof.memprofile).to.be.equal('memprofile.tmp');
                    expect(resp.config.app.pprof.cpuprofile).to.be.equal('cpuprofile.tmp');

                    //blockChain
                    expect(JSON.stringify(resp.config.chain.signature_ciphers)).equal('["cipher"]');
                    expect(resp.config.chain.coinbase).equal('75e4e5a71d647298b88928d8cb5da43d90ab1a6c52d0905f');
                    expect(resp.config.chain.gas_limit).equal('1000');
                    expect(resp.config.chain.gas_price).equal('1000');
                    expect(resp.config.chain.datadir).equal('data.tmp');
                    expect(resp.config.chain.keydir).equal('keydir.tmp');
                    expect(resp.config.chain.start_mine).equal(false);

                    //network
                    expect(JSON.stringify(resp.config.network.seed)).equal('["/ip4/127.0.0.1/tcp/8680/ipfs/QmP7HDFcYmJL12Ez4ZNVCKjKedfE7f48f1LAkUc3Whz4j1"]');
                    expect(JSON.stringify(resp.config.network.listen)).equal('["0.0.0.0:8991"]');
                    expect(resp.config.network.private_key).equal('network.tmp');

                    //rpc
                    expect(JSON.stringify(resp.config.rpc.rpc_listen)).equal('["0.0.0.0:8684"]');
                    expect(JSON.stringify(resp.config.rpc.http_listen)).equal('["0.0.0.0:8685"]');
                    expect(JSON.stringify(resp.config.rpc.http_module)).equal('["api"]');

                    //stats
                    expect(resp.config.stats.influxdb.host).equal('127.0.0.1:12345');
                    expect(resp.config.stats.influxdb.db).equal('test_db_123');
                    expect(resp.config.stats.influxdb.user).equal('silent');
                    expect(resp.config.stats.influxdb.password).equal('123456');
                    expect(resp.config.stats.enable_metrics).equal(false);
                    

                    //

                } catch (err) {
                    done(err);
                    neb_process.kill(9);
                    return;
                }
                neb_process.kill(9);
                done();
            });
        } catch (err) {
            neb_process.kill(9);
            done(err);
        }
    });

    after('remove the logs.tmp', (done) => {
        process.exec('rm -rf logs.tmp', (err, stdout, stderr) => {
            try {
                expect(err).to.be.equal(null);
                expect(stderr).to.be.equal("");
            } catch (err) {
                done(err);
                return;
            }
            done();
        });
    });
    after('remove the memprofile.tmp', (done) => {
        process.exec('rm -rf memprofile.tmp', (err, stdout, stderr) => {
            try {
                expect(err).to.be.equal(null);
                expect(stderr).to.be.equal("");
            } catch (err) {
                done(err);
                return;
            }
            done();
        });
    });

    after('remove the cpuprofile.tmp', (done) => {
        process.exec('rm -rf cpuprofile.tmp', (err, stdout, stderr) => {
            try {
                expect(err).to.be.equal(null);
                expect(stderr).to.be.equal("");
            } catch (err) {
                done(err);
                return;
            }
            done();
        });
    });

    after('remove the data.tmp', (done) => {
        process.exec('rm -rf data.tmp', (err, stdout, stderr) => {
            try {
                expect(err).to.be.equal(null);
                expect(stderr).to.be.equal("");
            } catch (err) {
                done(err);
                return;
            }
            done();
        });
    });

    after('move back the keydir', (done) => {
        process.exec('mv keydir.tmp keydir', (err, stdout, stderr) => {
            try {
                expect(err).to.be.equal(null);
                expect(stderr).to.be.equal("");
            } catch (err) {
                done(err);
                return;
            }
            done();
        });
    });

    after('remove network.tmp', (done) => {
        process.exec('rm network.tmp', (err, stdout, stderr) => {
            try {
                expect(err).to.be.equal(null);
                expect(stderr).to.be.equal("");
            } catch (err) {
                done(err);
                return;
            }
            done();
        });
    }); 
});
