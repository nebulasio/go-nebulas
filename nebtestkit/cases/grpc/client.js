var PROTO_PATH = __dirname + '/../../api_rpc.proto';

var grpc = require('grpc');
var api_proto = grpc.load(PROTO_PATH).rpcpb;

function main() {
  var client = new api_proto.ApiService('127.0.0.1:51510', grpc.credentials.createInsecure());
  client.NodeInfo({}, function(err, response) {
    console.log('Greeting:', response);
  });
}

main();