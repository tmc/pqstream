import pqstream_pb2_grpc
import pqstream_pb2
import grpc

channel = grpc.insecure_channel('localhost:7000')
stub = pqstream_pb2_grpc.PQStreamStub(channel)

request = pqstream_pb2.ListenRequest()
request.table_regexp = ".*"

for event in stub.Listen(request):
  print(event.payload)

