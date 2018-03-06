'use strict';
var expect = require('chai').expect;
var rpc_client = require('./rpc_client/rpc_client.js');

var protocol_version = '/neb/1.0.0'
var node_version = '0.7.0'
var server_address = 'localhost:8684';
var coinbase = "eb31ad2d8a89a0ca6935c308d5425730430bc2d63f2573b8";
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

var client;

describe('rpc: blockDump', function () {
  before(function () {
    client = rpc_client.new_client(server_address);
  });

  it('normal rpc', function (done) {
    client.blockDump({count: 111111111}, function (err, response) {
      if (err != null) {
        done(err);
        return;
      } else {
        try {
          //         verify_respone(response)
          console.log(response);
          //TODO：modify the rpc
          /*
          expect(response.chain_id).to.be.equal(chain_id);
          expect(response.chain_id).to.be.a('number');
          expect(response.tail).to.be.a('string');
          expect(response.height).to.be.a('string');
          expect(response.coinbase).to.be.equal(coinbase);
          expect(response.peer_count).to.be.a('number');
          expect(response.is_mining).to.equal(false);
          expect(response.protocol_version).to.equal(protocol_version);
          expect(response.synchronized).to.be.an('boolean');
          expect(response.version).to.equal(node_version);
          */
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
