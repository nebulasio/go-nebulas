'use strict';

var grpc = require('grpc');


    function new_client(server_address) {
        var PROTO_PATH = __dirname + '/./rpc.proto';
        var grpc = require('grpc');
        var rpc_proto = grpc.load(PROTO_PATH).rpcpb;
        var client = new rpc_proto.ApiService(server_address, grpc.credentials.createInsecure());
        return client;
    }


module.exports = {
    new_client: new_client
};