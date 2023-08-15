FROM golang:1.21-alpine as gotools

RUN mkdir /build

COPY . /build/go-sfomuseum-twitter-publish

RUN apk update && apk upgrade \
    && apk add git \
    #
    && cd /build \
    && git clone https://github.com/aaronland/gocloud-blob.git \
    && cd gocloud-blob \
    && go build -mod vendor -ldflags="-s -w" -o /usr/local/bin/copy cmd/copy/main.go \
    #
    && cd /build \
    && git clone https://github.com/sfomuseum/runtimevar.git \
    && cd runtimevar \
    && go build -mod vendor -ldflags="-s -w" -o /usr/local/bin/runtimevar cmd/runtimevar/main.go \
    #
    && cd /build \
    && cd go-sfomuseum-twitter-publish \
    && go build -mod vendor -ldflags="-s -w" -o /usr/local/bin/twitter-publish cmd/twitter-publish/main.go \
    #
    && cd / \
    && rm -rf build

FROM alpine

RUN mkdir /usr/local/data
RUN mkdir -p /usr/local/sfomuseum/bin

RUN apk update && apk upgrade

COPY --from=gotools /usr/local/bin/copy /usr/local/sfomuseum/bin/copy
COPY --from=gotools /usr/local/bin/runtimevar /usr/local/sfomuseum/bin/runtimevar
COPY --from=gotools /usr/local/bin/twitter-publish /usr/local/sfomuseum/bin/twitter-publish

COPY docker/publish-tweets /usr/local/sfomuseum/bin/publish-tweets    