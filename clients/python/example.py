#!/usr/bin/env python

import pqstream_pb2_grpc
import pqstream_pb2
import grpc

def run(messages):
  channel = grpc.insecure_channel('localhost:7000')
  stub = pqstream_pb2_grpc.PQStreamStub(channel)

  request = pqstream_pb2.ListenRequest()
  request.table_regexp = ".*"

  i = 0
  for event in stub.Listen(request):
    print("Received change, payload follows:")
    print(event.payload)
    i = i + 1
    if i >= messages:
      return

if __name__ == "__main__":
  print("Now receiving 100 messages.")
  run(100)
