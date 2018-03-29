'use strict';
var expect = require('chai').expect;
var rpc_client = require('./rpc_client/rpc_client.js');
var Wallet = require("nebulas");

var protocol_version = '/neb/1.0.0'
var node_version = '0.7.0'
var server_address = 'localhost:8684';
var sourceAccount;
var coinbase = "n1QZMXSZtW7BUerroSms4axNfyBGyFGkrh5";
var ChainID = 100;
var env = 'maintest';

if (env === 'testneb1') {
  ChainID = 1001;
  sourceAccount = new Wallet.Account("25a3a441a34658e7a595a0eda222fa43ac51bd223017d17b420674fb6d0a4d52");
  coinbase = "n1SAeQRVn33bamxN4ehWUT7JGdxipwn8b17";
  server_address = "35.182.48.19:8684";

} else if (env === "testneb2") {
  ChainID = 1002;
  sourceAccount = new Wallet.Account("25a3a441a34658e7a595a0eda222fa43ac51bd223017d17b420674fb6d0a4d52");
  coinbase = "n1SAeQRVn33bamxN4ehWUT7JGdxipwn8b17";
  server_address = "34.205.26.12:8684";

} else if (env === "testneb3") {
  ChainID = 1003;
  sourceAccount = new Wallet.Account("25a3a441a34658e7a595a0eda222fa43ac51bd223017d17b420674fb6d0a4d52");
  coinbase = "n1SAeQRVn33bamxN4ehWUT7JGdxipwn8b17";
  server_address = "35.177.214.138:8684";

} else if (env === "testneb4") { //super node
  ChainID = 1004;
  sourceAccount = new Wallet.Account("c75402f6ffe6edcc2c062134b5932151cb39b6486a7beb984792bb9da3f38b9f");
  coinbase = "n1EzGmFsVepKduN1U5QFyhLqpzFvM9sRSmG";
  server_address = "35.154.108.11:8684";
} else if (env === "testneb4_normalnode"){
  ChainID = 1004;
  sourceAccount = new Wallet.Account("c75402f6ffe6edcc2c062134b5932151cb39b6486a7beb984792bb9da3f38b9f");
  coinbase = "n1EzGmFsVepKduN1U5QFyhLqpzFvM9sRSmG";
  server_address = "18.197.107.228:8684";
} else if (env === "local") {
  ChainID = 100;
  sourceAccount = new Wallet.Account("d80f115bdbba5ef215707a8d7053c16f4e65588fd50b0f83369ad142b99891b5");
  coinbase = "n1QZMXSZtW7BUerroSms4axNfyBGyFGkrh5";
  server_address = "127.0.0.1:8684";

} else if (env === "maintest"){
  ChainID = 2;
  sourceAccount = new Wallet.Account("d2319a8a63b1abcb0cc6d4183198e5d7b264d271f97edf0c76cfdb1f2631848c");
  coinbase = "n1dZZnqKGEkb1LHYsZRei1CH6DunTio1j1q";
  server_address = "54.149.15.132:8684";
} else {
  throw new Error("invalid env (" + env + ").");
}


var client;

console.log("running chain_id: ", ChainID, " coinbase:", coinbase, " server_address:", server_address);
describe('rpc: getNebState', function () {
  before(function () {
    client = rpc_client.new_client(server_address);
  });

  it('normal rpc', function (done) {
      client.GetNebState({}, function (err, response) {
      if (err != null) {
        done(err);
        return;
      } else {
        try {
          //         verify_respone(response)
          expect(response.chain_id).to.be.equal(ChainID);
          expect(response.chain_id).to.be.a('number');
          expect(response.tail).to.be.a('string');
          expect(response.lib).to.be.a('string');
          expect(response.height).to.be.a('string');
          expect(response.protocol_version).to.equal(protocol_version);
          expect(response.synchronized).to.be.a('boolean');
          expect(response.version).to.equal(node_version);
        } catch (err) {
          done(err);
          return;
        }
        done()
        return;
      }
    });
  })

});
