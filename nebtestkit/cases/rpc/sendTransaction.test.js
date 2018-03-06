'use strict';
var expect = require('chai').expect;
var rpc_client = require('./rpc_client/rpc_client.js');

var protocol_version = '/neb/1.0.0'
var node_version = '0.7.0'
var server_address = 'localhost:8684';
var coinbase = "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8";
var sourceAccount = '1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c';
var chain_id = 100;
var env = '';
if (env === 'testneb1') {
  server_address = 'http://35.182.48.19:8684';
  coinbase = "0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d";
  chain_id = 1001;
} else if (env === "testneb2") {
  server_address = "http://34.205.26.12:8685";
  coinbase = "0b9cd051a6d7129ab44b17833c63fe4abead40c3714cde6d";
  chain_id = 1002;
}

var api_client;
var normalOutput;
var nonce;

function testRpc(testInput, testExpect, done) {
  api_client.sendTransaction(testInput.rpcInput, function (err, response) {
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
      address: sourceAccount,
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
    api_client.GetAccountState({ address: sourceAccount }, (err, resp) => {
      expect(err).to.be.equal(null);
      nonce = parseInt(resp.nonce);
      done(err);
    });
  })

  it('normal rpc', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount,
        to: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8",
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
        to: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8",
        value: "1",
        nonce: nonce,
        gas_price: "1000000",
        gas_limit: "200000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "address: invalid address"
    }

    testRpc(testInput, testExpect, done);
  })

  it('from address is short', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f25738",
        to: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8",
        value: "1",
        nonce: nonce,
        gas_price: "1000000",
        gas_limit: "200000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "address: invalid address"
    }

    testRpc(testInput, testExpect, done);
  })

  it('from address is long', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: "eb31ad2d8a89a0ca6935c308d5e425730430bc2d63f257383",
        to: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8",
        value: "1",
        nonce: nonce,
        gas_price: "1000000",
        gas_limit: "200000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "address: invalid address"
    }

    testRpc(testInput, testExpect, done);
  })

  it('from address is invalid', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: "xxeb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8",
        to: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8",
        value: "1",
        nonce: nonce,
        gas_price: "1000000",
        gas_limit: "200000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "address: invalid address"
    }

    testRpc(testInput, testExpect, done);
  })


  it('to address is empty', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount,
        value: "1",
        nonce: nonce,
        gas_price: "1000000",
        gas_limit: "200000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "address: invalid address"
    }

    testRpc(testInput, testExpect, done);
  })

  it('to address is short', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount,
        to: "eb31ad2d8a89a0ca6935c308d542570430bc2d63f2573b8",
        value: "1",
        nonce: nonce,
        gas_price: "1000000",
        gas_limit: "200000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "address: invalid address"
    }

    testRpc(testInput, testExpect, done);
  })

  it('to address is longer', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount,
        to: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f25733b8",
        value: "1",
        nonce: nonce,
        gas_price: "1000000",
        gas_limit: "200000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "address: invalid address"
    }

    testRpc(testInput, testExpect, done);
  })

  it('value is empty', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount,
        to: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8",
        nonce: nonce,
        gas_price: "1000000",
        gas_limit: "200000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "uint128: invalid string to uint128"
    }

    testRpc(testInput, testExpect, done);
  });

  it('value is invalid', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount,
        to: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8",
        value: "3$",
        nonce: nonce,
        gas_price: "1000000",
        gas_limit: "200000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "uint128: invalid string to uint128"
    }

    testRpc(testInput, testExpect, done);
  })

  it('value is out of uint128', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount,
        to: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8",
        value: "111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111",
        nonce: nonce,
        gas_price: "1000000",
        gas_limit: "200000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "uint128: overflow"
    }

    testRpc(testInput, testExpect, done);
  })

  it('value is neg', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount,
        to: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8",
        value: "-1",
        nonce: nonce,
        gas_price: "1000000",
        gas_limit: "200000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "uint128: underflow"
    }

    testRpc(testInput, testExpect, done);
  });

  it('gasLimit is empty', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount,
        to: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8",
        value: "1",
        nonce: nonce,
        gas_price: "1000000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "uint128: invalid string to uint128"
    }

    testRpc(testInput, testExpect, done);
  });

  it('gasLimit is neg', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount,
        to: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8",
        value: "1",
        nonce: nonce,
        gas_limit: "-1",
        gas_price: "1000000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "uint128: underflow"
    }

    testRpc(testInput, testExpect, done);
  });

  it('gasLimit is invalid', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount,
        to: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8",
        value: "1",
        nonce: nonce,
        gas_limit: "$d",
        gas_price: "1000000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "uint128: invalid string to uint128"
    }

    testRpc(testInput, testExpect, done);
  });

  it('gasLimit is out of uint128', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount,
        to: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8",
        value: "1",
        nonce: nonce,
        gas_limit: "111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111",
        gas_price: "1000000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "uint128: overflow"
    }

    testRpc(testInput, testExpect, done);
  });

  it('gasPrice is out of uint128', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount,
        to: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8",
        value: "1",
        nonce: nonce,
        gas_limit: "200000",
        gas_price: "111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "uint128: overflow"
    }

    testRpc(testInput, testExpect, done);
  });

  it('gasPrice is neg', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount,
        to: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8",
        value: "1",
        nonce: nonce,
        gas_limit: "200000",
        gas_price: "-1",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "uint128: underflow"
    }

    testRpc(testInput, testExpect, done);
  });

  it('gasPrice is invalid', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount,
        to: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8",
        value: "1",
        nonce: nonce,
        gas_limit: "200000",
        gas_price: "@#",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "uint128: invalid string to uint128"
    }

    testRpc(testInput, testExpect, done);
  })

  it('gasPrice is empty', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount,
        to: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8",
        value: "1",
        nonce: nonce,
        gas_limit: "200000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "uint128: invalid string to uint128"
    }

    testRpc(testInput, testExpect, done);
  });

  it('nonce is out of uint64', function (done) {
    nonce = nonce + 1;
    var testInput = {
      rpcInput: {
        from: sourceAccount,
        to: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8",
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
        from: sourceAccount,
        to: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8",
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
        from: sourceAccount,
        to: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8",
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
        from: sourceAccount,
        to: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8",
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
      address: sourceAccount,
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
        from: sourceAccount,
        to: "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8",
        value: "1",
        nonce: nonce,
        gas_price: "1000000",
        gas_limit: "200000",
      },
      isNormal: false
    }

    var testExpect = {
      errMsg: "key not unlocked"
    }
    setTimeout(() => { testRpc(testInput, testExpect, done); }, 100);
  })

});