FROM golang:alpine

RUN apk add --no-cache  git

COPY . /go/src/github.com/tmc/pqstream

RUN go get -v github.com/tmc/pqstream/cmd/pqs \
    && go get -v github.com/tmc/pqstream/cmd/pqsd

RUN go install github.com/tmc/pqstream/cmd/pqs \
    && go install github.com/tmc/pqstream/cmd/pqsd

ENV PATH $PATH:$GOPATH/bin

EXPOSE 7000

ENTRYPOINT ["pqsd"]
