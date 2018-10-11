'use strict';
var expect = require('chai').expect;
var rpc_client = require('./rpc_client/rpc_client.js');
var Wallet = require("nebulas");
var Account = Wallet.Account;

var server_address = 'localhost:8684';
var coinbase = "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8";
var sourceAccount;
var chain_id = 100;
var env = 'local';

if (env === 'testneb1') {
    chain_id = 1001;
    sourceAccount = new Wallet.Account("25a3a441a34658e7a595a0eda222fa43ac51bd223017d17b420674fb6d0a4d52");
    coinbase = "n1SAeQRVn33bamxN4ehWUT7JGdxipwn8b17";
    server_address = "35.182.48.19:8684";

}  else if (env === "local") {
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

function testRpc(testInput, testExpect, done) {
    client.GetNRByAddress(testInput.rpcInput, function (err, response) {
        if (err != null) {
            console.log("response err", err);
            try {
                expect(testExpect.errMsg).to.be.equal(err.details);
            } catch (err) {
                console.log(err);
                done(err)
                return;
            }
            done();
            return;
        } else {
            // console.log("response", response);
            try {
                // expect(response.data).to.be.a("string");
                expect(response.data).to.be.equal(testExpect.data);
            } catch (err) {
                done(err);
                return;
            };
        }
        done();
        return;
    });

}

describe('rpc: GetNRByAddress', function () {
    before(function () {
        client = rpc_client.new_client(server_address);
    });

    it('normal rpc', function (done) {
        var testInput = {
          rpcInput: {
            address: coinbase,
            height: 0
          }
        }
    
        var testExpect = {
          data: ""
        }
        
        testRpc(testInput, testExpect, done);
      })

    it('address is not exist', function (done) {
        var testInput = {
          rpcInput: {
            address: Account.NewAccount().getAddressString(),
            height: 0
          }
        }
    
        var testExpect = {
            data: ""
        }
        
        testRpc(testInput, testExpect, done);
      });
    
      it('address is invalid', function (done) {
        var testInput = {
          rpcInput: {
            address: 'b7d83b44@@3719220ec54cdb9f54c0202de68f1ebcb927b4f',
            height: 0
          }
        }
    
        var testExpect = {
          errMsg: 'address: invalid address format'
        }
        
        testRpc(testInput, testExpect, done);
      });
    
      it('address is null', function (done) {
        var testInput = {
          rpcInput: {
            address: '',
            height: 0
          }
        }
    
        var testExpect = {
          errMsg: 'address: invalid address format'
        }
        
        testRpc(testInput, testExpect, done);
      });
    
      it('address is empty', function (done) {
        var testInput = {
          rpcInput: {
            height: 0
          }
        }
    
        var testExpect = {
          errMsg: 'address: invalid address format'
        }
        
        testRpc(testInput, testExpect, done);
      })
    
      it('height is empty', function (done) {
        var testInput = {
          rpcInput: {
            address: coinbase,
          }
        }
    
        var testExpect = {
            data: ""
        }
        
        testRpc(testInput, testExpect, done);
      })
    
      it('height is negtive', function (done) {
        var testInput = {
          rpcInput: {
            address: coinbase,
            height: -1
          }
        }
    
        var testExpect = {
            data: ""
        }
        
        testRpc(testInput, testExpect, done);
      })
    
      it('height out of max', function (done) {
        var testInput = {
          rpcInput: {
            address: coinbase,
            height: 1111111111111111111111111111111111111111111111111111111
          }
        }
    
        var testExpect = {
            errMsg: 'invalid height'
        }
        
        testRpc(testInput, testExpect, done);
      })
    
      it('height is postive', function (done) {
        var testInput = {
          rpcInput: {
            address: coinbase,
            height: 1
          }
        }
    
        var testExpect = {
          data: ""
        }
        
        testRpc(testInput, testExpect, done);
      })
    
});
