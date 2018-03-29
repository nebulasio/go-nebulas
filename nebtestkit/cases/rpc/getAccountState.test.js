'use strict';
var expect = require('chai').expect;
var rpc_client = require('./rpc_client/rpc_client.js');
var Wallet = require("nebulas");
var Account = Wallet.Account;
var protocol_version = '/neb/1.0.0'
var node_version = '0.7.0';


var server_address = 'localhost:8684';
var coinbase = "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8";
var sourceAccount;
var chain_id = 100;
var env = 'maintest';

if (env === 'testneb1') {
  chain_id = 1001;
  sourceAccount = new Wallet.Account("25a3a441a34658e7a595a0eda222fa43ac51bd223017d17b420674fb6d0a4d52");
  coinbase = "n1SAeQRVn33bamxN4ehWUT7JGdxipwn8b17";
  server_address = "35.182.48.19:8684";

} else if (env === "testneb2") {
  chain_id = 1002;
  sourceAccount = new Wallet.Account("25a3a441a34658e7a595a0eda222fa43ac51bd223017d17b420674fb6d0a4d52");
  coinbase = "n1SAeQRVn33bamxN4ehWUT7JGdxipwn8b17";
  server_address = "34.205.26.12:8684";

} else if (env === "testneb3") {
  chain_id = 1003;
  sourceAccount = new Wallet.Account("25a3a441a34658e7a595a0eda222fa43ac51bd223017d17b420674fb6d0a4d52");
  coinbase = "n1SAeQRVn33bamxN4ehWUT7JGdxipwn8b17";
  server_address = "35.177.214.138:8684";

} else if (env === "testneb4") { //super node
  chain_id = 1004;
  sourceAccount = new Wallet.Account("c75402f6ffe6edcc2c062134b5932151cb39b6486a7beb984792bb9da3f38b9f");
  coinbase = "n1EzGmFsVepKduN1U5QFyhLqpzFvM9sRSmG";
  server_address = "35.154.108.11:8684";
} else if (env === "testneb4_normalnode"){
  chain_id = 1004;
  sourceAccount = new Wallet.Account("c75402f6ffe6edcc2c062134b5932151cb39b6486a7beb984792bb9da3f38b9f");
  coinbase = "n1EzGmFsVepKduN1U5QFyhLqpzFvM9sRSmG";
  server_address = "18.197.107.228:8684";
} else if (env === "local") {
  chain_id = 100;
  sourceAccount = new Wallet.Account("d80f115bdbba5ef215707a8d7053c16f4e65588fd50b0f83369ad142b99891b5");
  coinbase = "n1QZMXSZtW7BUerroSms4axNfyBGyFGkrh5";
  server_address = "127.0.0.1:8684";

} else if (env === "maintest"){
  chain_id = 2;
  sourceAccount = new Wallet.Account("d2319a8a63b1abcb0cc6d4183198e5d7b264d271f97edf0c76cfdb1f2631848c");
  coinbase = "n1dZZnqKGEkb1LHYsZRei1CH6DunTio1j1q";
  server_address = "54.149.15.132:8684";
} else {
  throw new Error("invalid env (" + env + ").");
}

var client;
var normalOutput;

function testRpc(testInput, testExpect, done) {
  client.getAccountState(testInput.rpcInput, function (err, response) {
    if (err != null) {
      try {
        expect(testExpect.errMsg).to.be.equal(err.details);
      } catch (err) {
        console.log(err);
        done(err)
        return;
      }
      console.log(err);
      done();
      return;
    } else {
      try {
        if (testInput.isNormal) {
          normalOutput = response;
        } else {
          if (testExpect.isNormalOutput) {
            expect(JSON.stringify(response)).to.be.equal(JSON.stringify(normalOutput));
          } else {
            expect(testExpect.isNormalOutput).to.be.equal(false);
            expect(JSON.stringify(response)).not.be.equal(JSON.stringify(normalOutput));
            expect(response.balance).to.be.a("string");
            //In JS, uint64 is converted to string
            expect(response.nonce).to.be.a('string');
          }     
        }
      } catch (err
      ) {
        done(err);
        return;
      };
    }
    done();
    return;
  });

}

describe('rpc: getAccountState', function () {
  before(function () {
    client = rpc_client.new_client(server_address);
  });
  
  it('normal rpc', function (done) {
    var testInput = {
      rpcInput: {
        address: coinbase,
        height: 0
      },
      isNormal: true
    }

    var testExpect = {
      isNormalOutput: true
    }
    
    testRpc(testInput, testExpect, done);
  })

  console.log(Account.NewAccount().getAddressString());
  it('address is not exist', function (done) {
    var testInput = {
      rpcInput: {
        address: Account.NewAccount().getAddressString(),
        height: 0
      },
      isNormal: false
    }

    var testExpect = {
      isNormalOutput: false,
    }
    
    testRpc(testInput, testExpect, done);
  });

  it('address is invalid', function (done) {
    var testInput = {
      rpcInput: {
        address: 'b7d83b44@@3719220ec54cdb9f54c0202de68f1ebcb927b4f',
        height: 0
      },
      isNormal: false
    }

    var testExpect = {
      isNormalOutput: false,
      errMsg: 'address: invalid address format'
    }
    
    testRpc(testInput, testExpect, done);
  });

  it('address is null', function (done) {
    var testInput = {
      rpcInput: {
        address: '',
        height: 0
      },
      isNormal: false
    }

    var testExpect = {
      isNormalOutput: false,
      errMsg: 'address: invalid address format'
    }
    
    testRpc(testInput, testExpect, done);
  });

  it('address is empty', function (done) {
    var testInput = {
      rpcInput: {
        height: 0
      },
      isNormal: false
    }

    var testExpect = {
      isNormalOutput: false,
      errMsg: 'address: invalid address format'
    }
    
    testRpc(testInput, testExpect, done);
  })

  it('height is empty', function (done) {
    var testInput = {
      rpcInput: {
        address: coinbase,
      },
      isNormal: false
    }

    var testExpect = {
      isNormalOutput: true
    }
    
    testRpc(testInput, testExpect, done);
  })

  it('height is negtive', function (done) {
    var testInput = {
      rpcInput: {
        address: coinbase,
        height: -1
      },
      isNormal: false
    }

    var testExpect = {
      isNormalOutput: true//todo: to check
    }
    
    testRpc(testInput, testExpect, done);
  })

  it('height out of max', function (done) {
    var testInput = {
      rpcInput: {
        address: coinbase,
        height: 1111111111111111111111111111111111111111111111111111111
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: 'block not found'
    }
    
    testRpc(testInput, testExpect, done);
  })

  it('height is postive', function (done) {
    var testInput = {
      rpcInput: {
        address: coinbase,
        height: 2
      },
      isNormal: false
    }

    var testExpect = {
      isNormalOutput: false
    }
    
    testRpc(testInput, testExpect, done);
  })

});
