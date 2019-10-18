FROM golang:alpine as builder

RUN apk add --no-cache  git

ENV BUILD_PATH /build/main

COPY . /go/src/github.com/tmc/pqstream

RUN go get -v github.com/tmc/pqstream/cmd/pqs \
    && go get -v github.com/tmc/pqstream/cmd/pqsd

RUN mkdir -p ${BUILD_PATH}
ENV GOBIN ${BUILD_PATH}

RUN go install github.com/tmc/pqstream/cmd/pqs \
    && go install github.com/tmc/pqstream/cmd/pqsd

FROM alpine

RUN adduser -S -D -H -h /app appuser && \
    mkdir /app && \
    chown appuser:nogroup /app

USER appuser

WORKDIR /app

COPY --from=builder --chown=appuser:nogroup /build/main /app/

EXPOSE 7000

CMD ["./pqsd"]
