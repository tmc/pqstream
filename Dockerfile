FROM golang:latest

EXPOSE 7000

WORKDIR /go/src/github.com/tmc/pqstream
COPY . .
WORKDIR cmd/pqsd
# "go get -d -v ./..."
RUN go-wrapper download
# "go install -v ./..."
RUN go-wrapper install
WORKDIR /go/src/github.com/tmc/pqstream/cmd/pqs
RUN go-wrapper download
RUN go-wrapper install

WORKDIR /tmp

ENTRYPOINT ["pqsd"]
CMD ["-h"]
