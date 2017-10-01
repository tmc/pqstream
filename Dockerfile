FROM golang:alpine

RUN apk add --update git

RUN go get  github.com/tmc/pqstream/cmd/pqs \
    && go get github.com/tmc/pqstream/cmd/pqsd
