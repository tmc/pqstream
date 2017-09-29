FROM golang:1.8.3-alpine3.6

RUN apk add --update openssl git openssh bash

RUN wget https://github.com/google/protobuf/releases/download/v3.4.0/protoc-3.4.0-linux-x86_64.zip \
    && mkdir -p /usr/local/protoc \
    && unzip -d /usr/local/protoc  protoc-3.4.0-linux-x86_64.zip 

ENV PATH $PATH:/usr/local/protoc/bin

RUN echo -e "[url \"git@github.com:\"]\n\tinsteadOf = https://github.com" >> /root/.gitconfig

RUN mkdir /root/.ssh && echo "StrictHostKeyChecking no " > /root/.ssh/config

RUN go get -u github.com/tmc/pqstream/cmd/pqs \
    && go get -u github.com/tmq/pqstream/cmd/pqsd

