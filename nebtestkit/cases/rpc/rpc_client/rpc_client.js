'use strict';

var grpc = require('grpc');


    function new_client(server_address, service) {
        var PROTO_PATH = __dirname + '/./rpc.proto';
        var grpc = require('grpc');
        var rpc_proto = grpc.load(PROTO_PATH).rpcpb;
        var client = (service === 'AdminService') ? new rpc_proto.AdminService(server_address, grpc.credentials.createInsecure())
                : new rpc_proto.ApiService(server_address, grpc.credentials.createInsecure());
        return client;
    }


module.exports = {
    new_client: new_client
};