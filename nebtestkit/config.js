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
		chain_id: 100
		datadir: {{data_location}}
		keydir: "keydir"
		coinbase: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8"
		signature_ciphers: ["ECC_SECP256K1"]
	}
	
	rpc {
			rpc_listen: ["127.0.0.1:{{rpc_port}}"]
			http_listen: ["127.0.0.1:{{http_port}}"]
			http_module: ["api", "admin"]
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

var config_non_seed =
  `  network {
    listen: ["127.0.0.1:{{port}}"]
    seed: ["/ip4/{{seed_ip}}/tcp/{{seed_port}}/ipfs/QmPyr4ZbDmwF1nWxymTktdzspcBFPL6X1v3Q5nT7PGNtUN"]
  }

  chain {
    chain_id: 100
    datadir: {{data_location}}
    keydir: "keydir"
    coinbase: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8"
    signature_ciphers: ["ECC_SECP256K1"]
  }

  rpc {
      rpc_listen: ["127.0.0.1:{{rpc_port}}"]
      http_listen: ["127.0.0.1:{{http_port}}"]
      http_module: ["api", "admin"]
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

var RPC_PORT = 51510, HTTP_PORT = 8191;
var nonius = 1;
var now = Date.now();
var dirname = tempdir + '/nebulas/' + now;



exports.createSeedConfig = function (port) {
  mkdirsSync(dirname);
  var dataSeed = {
    port: port,
    rpc_port: RPC_PORT,
    http_port: HTTP_PORT,
    data_location: '"' + dirname + '/data.db' + '"'
  };
  var configSeed = new Buffer(render(config_seed, dataSeed));

  fs.writeFile(dirname + '/seed.conf', configSeed, { flag: 'w' }, function (err) {
    if (err) {
      console.error(err);
    } else {
      // console.log('generate default config file success.');
    }
  });

  return dirname + '/seed';
};

exports.createNonSeedConfig = function (seed, port, http_port) {
  mkdirsSync(dirname);
  var nonius = nextNonius();
  var dataNonSeed = {
    port: port,
    seed_ip: seed.ip,
    seed_port: seed.port,
    rpc_port: RPC_PORT + nonius,
    http_port: http_port,
    data_location: '"' + dirname + '/data.db_' + nonius + '"'
  };
  var configNonSeed = new Buffer(render(config_non_seed, dataNonSeed));
  fs.writeFile(dirname + '/nonseed_' + nonius + '.conf', configNonSeed, { flag: 'w' }, function (err) {
    if (err) {
      console.error(err);
    } else {
      // console.log('generate normal config file success.');
    }
  });

  return dirname + '/nonseed_' + nonius;
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

function nextNonius() {
  return nonius++;
}