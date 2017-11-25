var config = require('./config');
var fs = require('fs');
var Neb = require('../cmd/console/neb.js/lib/neb.js');
var HttpRequest = require("../cmd/console/neb.js/lib/httprequest.js");

var Neblet = function (ip, port, http_port) {
    this.ip = ip;
    this.port = port;
    this.http_port = http_port;
};

Neblet.prototype = {

    Init: function (seed) {
        if (seed) {
            this._configName = config.createNonSeedConfig(seed, this.port, this.http_port);
        } else {
            this._configName = config.createSeedConfig(this.port, this.http_port);;
        }
    },

    Start: function () {
        var spawn = require('child_process').spawn;
        var neb = spawn('./neb', ['-c', this._configName + '.pb.txt']);
        var logPath = this._configName + '.log';
        neb.stdout.on('data', function (data) {
            fs.writeFile(logPath, data, { flag: 'a' }, function (err) {
                if (err) {
                    console.error(err);
                }
            })
        });

        return neb;
    },

    NebJs: function () {
        var httpRequest = new HttpRequest('http://' + this.ip + ':' + this.http_port);
        return new Neb(httpRequest);
    },
};

module.exports = Neblet;