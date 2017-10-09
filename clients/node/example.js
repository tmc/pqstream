var grpc = require('grpc');
var path = require('path')
var protoDescriptor = grpc.load(path.resolve(__dirname ,'..', '..', 'pqstream.proto'));
var pqs = protoDescriptor.pqs
var ListenRequest = pqs.ListenRequest;
var PQStream  = pqs.PQStream;

function main() {
  var client = new PQStream('0.0.0.0:7000', grpc.credentials.createInsecure());
  var request = new ListenRequest()
  request.table_regexp = ".*"

  var call = client.listen(request)
  call.on('data', function( data) {
    console.log("data received", data)
  })

  call.on('end', function() {
    // The server has finished sending
  });
  call.on('status', function(status) {
    // process status
  });
  call.on("error", function(err) {
    console.log(err)
  })
}

main();