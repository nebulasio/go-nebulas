'use strict';
var fs = require('fs'), 
  os = require('os'),
  path = require('path');

var tempdir = os.tmpdir();

var config_seed =
  `  p2p {
    port: {{port}}
    chain_id: 100
    version: 1
  }
  rpc {
    api_port: {{api_port}}
    management_port: {{management_port}}
    api_http_port: {{api_http_port}}
    management_http_port: {{management_http_port}}
  }
  pow {
    coinbase: "8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf"
  }
  
  storage {
    location: {{data_location}}
  }
  
  account {
    # keystore.SECP256K1 = 1
    signature: 1
  
    # keystore.SCRYPT = 1 << 4
    encrypt: 16
  
    key_dir: "keydir"
  
    test_passphrase: "passphrase"
  }
  
  influxdb {
    host: "http://localhost:8086"
    db: "nebulas"
    username: "admin"
    password: "admin"
  }
  
  metrics {
    enable: false
  }`;

var config_non_seed =
  `   p2p {
      port: {{port}}
      seed: "/ip4/{{seed_ip}}/tcp/{{seed_port}}/ipfs/QmPyr4ZbDmwF1nWxymTktdzspcBFPL6X1v3Q5nT7PGNtUN"
      chain_id: 100
      version: 1
    }
    rpc {
      api_port: {{api_port}}
      management_port: {{management_port}}
      api_http_port: {{api_http_port}}
      management_http_port: {{management_http_port}}
    }
    pow {
      coinbase: "8a209cec02cbeab7e2f74ad969d2dfe8dd24416aa65589bf"
    }
    
    storage {
      location: {{data_location}}
    }
    
    account {
      # keystore.SECP256K1 = 1
      signature: 1
    
      # keystore.SCRYPT = 1 << 4
      encrypt: 16
    
      key_dir: "keydir"
    
      test_passphrase: "passphrase"
    }
    
    influxdb {
      host: "http://localhost:8086"
      db: "nebulas"
      username: "admin"
      password: "admin"
    }
    
    metrics {
      enable: false
    }`;

var API_PORT = 51510, MANAGEMENT_PORT = 53521, API_HTTP_PORT = 8090, MANAGEMENT_HTTP_PORT = 8191;
var nonius = 1;
var now = Date.now();
var dirname = tempdir + '/nebulas/' + now;



exports.createSeedConfig = function (port) {
  mkdirsSync(dirname);
  var dataSeed = {
    port: port,
    api_port: API_PORT,
    management_port: MANAGEMENT_PORT,
    api_http_port: API_HTTP_PORT,
    management_http_port: MANAGEMENT_HTTP_PORT,
    data_location: '"' + dirname + '/data.db' + '"'
  };
  var configSeed = new Buffer(render(config_seed, dataSeed));

  fs.writeFile(dirname + '/seed.pb.txt', configSeed, { flag: 'w' }, function (err) {
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
    api_port: API_PORT + nonius,
    management_port: MANAGEMENT_PORT + nonius,
    api_http_port: API_HTTP_PORT + nonius,
    management_http_port: http_port,
    data_location: '"' + dirname + '/data.db_' + nonius + '"'
  };
  var configNonSeed = new Buffer(render(config_non_seed, dataNonSeed));
  fs.writeFile(dirname + '/nonseed_' + nonius + '.pb.txt', configNonSeed, { flag: 'w' }, function (err) {
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