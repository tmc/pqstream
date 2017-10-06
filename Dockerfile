FROM golang:alpine

COPY . /go/src/github.com/tmc/pqstream

RUN go install github.com/tmc/pqstream/cmd/pqs \
    && go install github.com/tmc/pqstream/cmd/pqsd
