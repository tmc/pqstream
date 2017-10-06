#!/usr/bin/env python

"""PQstream Python example client

This client connects to a running PQstream instance running at
localhost on port 7000 and processes, when called from the command
line, 100 messages and then exits. The payload of the messages will be
printed to stdout.
"""

import pqstream_pb2_grpc
import pqstream_pb2
import grpc

def run(messages):
    """process a set number of messages"""
    channel = grpc.insecure_channel('localhost:7000')
    stub = pqstream_pb2_grpc.PQStreamStub(channel)

    request = pqstream_pb2.ListenRequest()
    request.table_regexp = ".*"

    i = 0
    for event in stub.Listen(request):
        print("Received change, payload follows:")
        print("OP: {0}".format(event.op))
        print("Table: {0}.{1}".format(event.schema, event.table))
        print(event.payload)
        i = i + 1
        if i >= messages:
            return

if __name__ == "__main__":
    print("Now receiving 100 messages.")
    run(100)
