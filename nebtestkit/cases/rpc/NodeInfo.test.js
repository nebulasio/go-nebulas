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

describe('rpc: NodeInfo', function () {
  before(function () {
    client = rpc_client.new_client(server_address);
  });

  it('normal rpc', function (done) {
    client.NodeInfo({}, function (err, response) {
      if (err != null) {
        done(err);
        return;
      } else {
        try {
          expect(response.chain_id).to.be.equal(chain_id);
          expect(response.version).to.be.a('number');
          expect(response.peer_count).to.be.a('number');
          expect(response.synchronized).to.be.a('boolean');
          expect(response.bucket_size).to.be.a('number');
          expect(response.relay_cache_size).to.be.a('number');
          expect(response.stream_store_size).to.be.a('number');
          expect(response.stream_store_extend_size).to.be.a('number');
          expect(response.protocol_version).to.equal(protocol_version);
          expect(response).to.be.have.property('route_table');
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
