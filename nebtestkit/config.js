'use strict';
var fs = require('fs'),
  os = require('os'),
  path = require('path');

var tempdir = os.tmpdir();

var config_seed =
  `  network {
		listen: ["127.0.0.1:{{port}}"]
	}
	
	chain {
    chain_id: 1001
    datadir: {{data_location}}
    keydir: "keydir"
    genesis: "genesis.conf"
    coinbase: {{coinbase}}
    signature_ciphers: ["ECC_SECP256K1"]
    miner: {{miner}}
    passphrase: {{passphrase}}
	}
	
	rpc {
			rpc_listen: ["127.0.0.1:{{rpc_port}}"]
			http_listen: ["127.0.0.1:{{http_port}}"]
			http_module: ["api", "admin"]
  }
  
  app {
    log_level: "debug"
    log_file: {{log_file}}
    enable_crash_report: false
  }
	
	stats {
			enable_metrics: false
			influxdb: {
					host: "http://localhost:8086"
					db: "nebulas"
					user: "admin"
					password: "admin"
			}
	}`;

var config_normal =
  `  network {
    listen: ["127.0.0.1:{{port}}"]
    seed: ["/ip4/{{seed_ip}}/tcp/{{seed_port}}/ipfs/QmPyr4ZbDmwF1nWxymTktdzspcBFPL6X1v3Q5nT7PGNtUN"]
  }

  chain {
    chain_id: 1001
    datadir: {{data_location}}
    keydir: "keydir"
    genesis: "genesis.conf"
    coinbase: {{coinbase}}
    signature_ciphers: ["ECC_SECP256K1"]
    miner: {{miner}}
    passphrase: {{passphrase}}
  }

  rpc {
      rpc_listen: ["127.0.0.1:{{rpc_port}}"]
      http_listen: ["127.0.0.1:{{http_port}}"]
      http_module: ["api", "admin"]
  }

  app {
    log_level: "debug"
    log_file: {{log_file}}
    enable_crash_report: false
  }

  stats {
      enable_metrics: false
      influxdb: {
          host: "http://localhost:8086"
          db: "nebulas"
          user: "admin"
          password: "admin"
      }
  }`;

var now = new Date().getTime();
var dirname = tempdir + '/nebulas/' + now;
console.log(dirname);

exports.createSeedConfig = function (port, http_port, rpc_port, coinbase, miner, passphrase) {
  mkdirsSync(dirname);
  var dataSeed = {
    port: port,
    rpc_port: rpc_port,
    http_port: http_port,
    coinbase: '"' + coinbase + '"',
    miner: '"' + miner + '"',
    passphrase: '"' + passphrase + '"',
    log_file: '"' + dirname + '/seed.log' + '"',
    data_location: '"' + dirname + '/seed.db' + '"'
  };
  var configSeed = new Buffer(render(config_seed, dataSeed));
  fs.writeFileSync(dirname + '/seed.conf', configSeed, "utf-8");

  return dirname + '/seed';
};

exports.createNormalConfig = function (seed, port, http_port, rpc_port, coinbase, miner, passphrase) {
  mkdirsSync(dirname);
  var dataNonSeed = {
    port: port,
    seed_ip: seed.ip,
    seed_port: seed.port,
    rpc_port: rpc_port,
    http_port: http_port,
    coinbase: '"' + coinbase + '"',
    miner: '"' + miner + '"',
    passphrase: '"' + passphrase + '"',
    log_file: '"' + dirname + '/normal.' + (port - seed.port) + '.log' + '"',
    data_location: '"' + dirname + '/normal.' + (port - seed.port) + '.db' + '"'
  };
  var configNonSeed = new Buffer(render(config_normal, dataNonSeed));
  fs.writeFileSync(dirname + '/normal.' + (port - seed.port) + '.conf', configNonSeed, 'utf-8');

  return dirname + '/normal.' + (port - seed.port);
};

function render(tpl, data) {
  var re = /{{([^}]+)?}}/;
  var match = '';
  while (match = re.exec(tpl)) {
    tpl = tpl.replace(match[0], data[match[1]]);
  }
  return tpl;
}

function mkdirsSync(dirname) {
  if (fs.existsSync(dirname)) {
    return true;
  } else {
    if (mkdirsSync(path.dirname(dirname))) {
      fs.mkdirSync(dirname);
      return true;
    }
  }
}