'use strict';
var expect = require('chai').expect;
var rpc_client = require('./rpc_client/rpc_client.js');
var Wallet  = new require('nebulas');

var protocol_version = '/neb/1.0.0'
var node_version = '0.7.0'
var server_address = 'localhost:8684';
var coinbase = "n1QZMXSZtW7BUerroSms4axNfyBGyFGkrh5";
var sourceAccount = new Wallet.Account("d80f115bdbba5ef215707a8d7053c16f4e65588fd50b0f83369ad142b99891b5");
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

var api_client;
var normalOutput;
var nonce;

function testRpc(testInput, testExpect, done) {
  var admin_client = rpc_client.new_client(server_address, "AdminService");
  admin_client.sendTransaction(testInput.rpcInput, function (err, response) {
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
        console.log("silent_debug" + JSON.stringify(response));
        if (testInput.isNormal) {
          //TODO:verify response
          //  expect(response.balance).to.be.a("string");
          normalOutput = response;
        } else {
          if (testExpect.isNormalOutput) {
            expect(JSON.stringify(response)).to.be.equal(JSON.stringify(normalOutput));
          } else {
            expect(testExpect.isNormalOutput).to.be.equal(false);
            expect(JSON.stringify(response)).not.be.equal(JSON.stringify(normalOutput));
          }
        }
      } catch (err) {
        done(err);
        return;
      };
    }
    done();
    return;
  });

}

describe('rpc: sendTransaction', function () {
  before((done) => {
    var admin_client = rpc_client.new_client(server_address, 'AdminService');
    var args = {
      address: sourceAccount.getAddressString(),
      passphrase: "passphrase",
    }
    admin_client.UnlockAccount(args, (err, resp) => {
      expect(err).to.be.equal(null);
      console.log(err);
      done(err);
    })
  });

  before((done) => {
    api_client = rpc_client.new_client(server_address);
    api_client.GetAccountState({ address: sourceAccount.getAddressString() }, (err, resp) => {
      expect(err).to.be.equal(null);
      nonce = parseInt(resp.nonce);
      done(err);
    });
  })

  it('normal rpc', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount.getAddressString(),
        to: coinbase,
        value: "1",
        nonce: nonce,
        gas_price: "1000000",
        gas_limit: "200000",
      },
      isNormal: true
    }

    var testExpect = {
      isNormalOutput: true
    }

    testRpc(testInput, testExpect, done);
  })

  it('from address is empty', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        to: coinbase,
        value: "1",
        nonce: nonce,
        gas_price: "1000000",
        gas_limit: "200000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "address: invalid address format"
    }

    testRpc(testInput, testExpect, done);
  })

  it('from address is short', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: "eb31ad2d30bc2d63f25738",
        to: coinbase,
        value: "1",
        nonce: nonce,
        gas_price: "1000000",
        gas_limit: "200000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "address: invalid address format"
    }

    testRpc(testInput, testExpect, done);
  })

  it('from address is long', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: "eb31ad2d8a89a0ca6935c308d5e425730430bc2d63f257383",
        to: coinbase,
        value: "1",
        nonce: nonce,
        gas_price: "1000000",
        gas_limit: "200000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "address: invalid address format"
    }

    testRpc(testInput, testExpect, done);
  })

  it('from address is invalid', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: "xxeb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8",
        to: coinbase,
        value: "1",
        nonce: nonce,
        gas_price: "1000000",
        gas_limit: "200000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "address: invalid address format"
    }

    testRpc(testInput, testExpect, done);
  })


  it('to address is empty', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount.getAddressString(),
        value: "1",
        nonce: nonce,
        gas_price: "1000000",
        gas_limit: "200000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "address: invalid address format"
    }

    testRpc(testInput, testExpect, done);
  })

  it('to address is short', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount.getAddressString(),
        to: "eb31ad2d8a893b8",
        value: "1",
        nonce: nonce,
        gas_price: "1000000",
        gas_limit: "200000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "address: invalid address format"
    }

    testRpc(testInput, testExpect, done);
  })

  it('to address is longer', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount.getAddressString(),
        to: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f25733b8",
        value: "1",
        nonce: nonce,
        gas_price: "1000000",
        gas_limit: "200000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "address: invalid address format"
    }

    testRpc(testInput, testExpect, done);
  })

  it('value is empty', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount.getAddressString(),
        to: coinbase,
        nonce: nonce,
        gas_price: "1000000",
        gas_limit: "200000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "invalid value"
    }

    testRpc(testInput, testExpect, done);
  });

  it('value is invalid', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount.getAddressString(),
        to: coinbase,
        value: "3$",
        nonce: nonce,
        gas_price: "1000000",
        gas_limit: "200000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "invalid value"
    }

    testRpc(testInput, testExpect, done);
  })

  it('value is out of uint128', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount.getAddressString(),
        to: coinbase,
        value: "111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111",
        nonce: nonce,
        gas_price: "1000000",
        gas_limit: "200000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "invalid value"
    }

    testRpc(testInput, testExpect, done);
  })

  it('value is neg', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount.getAddressString(),
        to: coinbase,
        value: "-1",
        nonce: nonce,
        gas_price: "1000000",
        gas_limit: "200000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "invalid value"
    }

    testRpc(testInput, testExpect, done);
  });

  it('gasLimit is empty', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount.getAddressString(),
        to: coinbase,
        value: "1",
        nonce: nonce,
        gas_price: "1000000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "invalid gasLimit"
    }

    testRpc(testInput, testExpect, done);
  });

  it('gasLimit is neg', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount.getAddressString(),
        to: coinbase,
        value: "1",
        nonce: nonce,
        gas_limit: "-1",
        gas_price: "1000000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "invalid gasLimit"
    }

    testRpc(testInput, testExpect, done);
  });

  it('gasLimit is invalid', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount.getAddressString(),
        to: coinbase,
        value: "1",
        nonce: nonce,
        gas_limit: "$d",
        gas_price: "1000000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "invalid gasLimit"
    }

    testRpc(testInput, testExpect, done);
  });

  it('gasLimit is out of uint128', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount.getAddressString(),
        to: coinbase,
        value: "1",
        nonce: nonce,
        gas_limit: "111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111",
        gas_price: "1000000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "invalid gasLimit"
    }

    testRpc(testInput, testExpect, done);
  });

  it('gasPrice is out of uint128', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount.getAddressString(),
        to: coinbase,
        value: "1",
        nonce: nonce,
        gas_limit: "200000",
        gas_price: "111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "invalid gasPrice"
    }

    testRpc(testInput, testExpect, done);
  });

  it('gasPrice is neg', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount.getAddressString(),
        to: coinbase,
        value: "1",
        nonce: nonce,
        gas_limit: "200000",
        gas_price: "-1",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "invalid gasPrice"
    }

    testRpc(testInput, testExpect, done);
  });

  it('gasPrice is invalid', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount.getAddressString(),
        to: coinbase,
        value: "1",
        nonce: nonce,
        gas_limit: "200000",
        gas_price: "@#",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "invalid gasPrice"
    }

    testRpc(testInput, testExpect, done);
  })

  it('gasPrice is empty', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount.getAddressString(),
        to: coinbase,
        value: "1",
        nonce: nonce,
        gas_limit: "200000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "invalid gasPrice"
    }

    testRpc(testInput, testExpect, done);
  });

  it('nonce is out of uint64', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount.getAddressString(),
        to: coinbase,
        value: "1",
        nonce: 1000000000000000000000000000000000000000000000000000000000000000000000000000000000000,
        gas_limit: "200000",
        gas_price: "1000000"
      },
      isNormal: false
    }

    var testExpect = {
      isNormalOutput: false
    }

    testRpc(testInput, testExpect, done);
  })

  it('nonce is neg', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount.getAddressString(),
        to: coinbase,
        value: "1",
        nonce: -1,
        gas_limit: "200000",
        gas_price: "1000000"
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: 'transaction\'s nonce is invalid, should bigger than the from\'s nonce'
    }

    testRpc(testInput, testExpect, done);
  })

  it('nonce is empty', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount.getAddressString(),
        to: coinbase,
        value: "1",
        gas_limit: "200000",
        gas_price: "1000000"
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "transaction's nonce is invalid, should bigger than the from's nonce"
    }

    testRpc(testInput, testExpect, done);
  })

  it('nonce is invalid', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount.getAddressString(),
        to: coinbase,
        value: "1",
        nonce: "@",
        gas_limit: "200000",
        gas_price: "1000000"
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "transaction's nonce is invalid, should bigger than the from's nonce"
    }

    testRpc(testInput, testExpect, done);
  });

  it('from address is unlock', function (done) {
    nonce = nonce + 1;

    var admin_client = rpc_client.new_client(server_address, 'AdminService');
    var args = {
      address: sourceAccount.getAddressString(),
    }
    var lock = 0;
    admin_client.LockAccount(args, (err, resp) => {
      console.log(err);
      if (err != null) {
        console.log(err)
        done(err);
      }
    });
    var testInput = {
      rpcInput: {
        from: sourceAccount.getAddressString(),
        to: coinbase,
        value: "1",
        nonce: nonce,
        gas_price: "1000000",
        gas_limit: "200000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "account is locked"
    }
    setTimeout(() => { testRpc(testInput, testExpect, done); }, 100);
  })

});