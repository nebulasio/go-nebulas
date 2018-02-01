# Stage 1 of 2 stage [build]
FROM golang:stretch

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

WORKDIR /go/src/github.com/nebulasio/go-nebulas

RUN apt-get update && \
    apt-get -yy -q install git build-essential protobuf-compiler sudo 

# Build Nebulas
COPY . /go/src/github.com/nebulasio/go-nebulas

RUN go get -u github.com/golang/dep/cmd/dep && \
    go get -u golang.org/x/tools/cmd/goimports && \
    make dep && \
    make deploy-v8 && \
    make build

# Stage 2 of 2 [copy assets to thin image]
FROM debian:stretch

WORKDIR /nebulas

ENV PATH /usr/local/bin:$PATH

# Copy v8 libs from builder
COPY --from=0 /usr/local/lib /usr/local/lib
RUN ldconfig

# Copy neb and crash reporter from builder
COPY --from=0 /go/src/github.com/nebulasio/go-nebulas/neb /usr/local/bin/neb
COPY --from=0 /go/src/github.com/nebulasio/go-nebulas/nebulas_crashreporter /usr/local/bin/nebulas_crashreporter

# Copy conf & keydir to enable bootstrapping
COPY --from=0 /go/src/github.com/nebulasio/go-nebulas/keydir /nebulas/keydir
COPY --from=0 /go/src/github.com/nebulasio/go-nebulas/conf /nebulas/conf

ENTRYPOINT ["/usr/local/bin/neb"]

