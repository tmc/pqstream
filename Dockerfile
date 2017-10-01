FROM golang:alpine

RUN apk add --update openssl git openssh bash

RUN wget https://github.com/google/protobuf/releases/download/v3.4.0/protoc-3.4.0-linux-x86_64.zip \
    && mkdir -p /usr/local/protoc \
    && unzip -d /usr/local/protoc  protoc-3.4.0-linux-x86_64.zip 

ENV PATH $PATH:/usr/local/protoc/bin

RUN go get  github.com/tmc/pqstream/cmd/pqs \
    && go get github.com/tmc/pqstream/cmd/pqsd
