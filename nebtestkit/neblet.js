var config = require('./config');
var fs = require('fs');
var Neb = require('../cmd/console/neb.js/lib/neb.js');
var HttpRequest = require("./node-request");

var Neblet = function (ip, port, http_port, rpc_port, coinbase, miner, passphrase) {
    this.ip = ip;
    this.port = port;
    this.rpc_port = rpc_port;
    this.http_port = http_port;
    this.coinbase = coinbase;
    this.miner = miner;
    this.passphrase = passphrase;
};

Neblet.prototype = {

    Init: function (seed) {
        if (seed) {
            this._configName = config.createNormalConfig(
                seed, this.port, this.http_port, this.rpc_port,
                this.coinbase, this.miner, this.passphrase);
        } else {
            this._configName = config.createSeedConfig(
                this.port, this.http_port, this.rpc_port,
                this.coinbase, this.miner, this.passphrase);
        }
    },

    Start: function () {
        var spawn = require('child_process').spawn;
        var neb = spawn('./neb', ['-c', this._configName + '.conf']);
        var debug = this._configName + '.debug.log';
        var err = this._configName + '.err.log';
        neb.stdout.on('data', function (data) {
            fs.writeFile(debug, data, {
                flag: 'a'
            }, function (err) {
                if (err) {
                    console.error(err);
                }
            });
        });
        neb.stderr.on('data', function (data) {
            fs.writeFile(err, data, {
                flag: 'a'
            }, function (err) {
                if (err) {
                    console.error(err);
                }
            });
        });
        this.neb = neb;

        return neb;
    },

    RPC: function () {

        var httpRequest = new HttpRequest('http://' + this.ip + ':' + this.http_port);
        return new Neb(httpRequest);
    },

    Coinbase: function () {
        return this.coinbase;
    },
    Passphrase: function () {
        return this.passphrase;
    },
    Kill: function () {
        this.neb.kill('SIGINT');
    },
};

module.exports = Neblet;