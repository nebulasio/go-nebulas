FROM ubuntu:16.04

ENV GOPATH /go
ENV PATH ${GOPATH}/bin:/usr/local/go/bin:${PATH}

RUN apt update && \
    apt install -y git build-essential protobuf-compiler sudo wget

# Install Go1.9.2
RUN wget https://golangtc.com/static/go/1.9.2/go1.9.2.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.9.2.linux-amd64.tar.gz && \
    mkdir go

# Install RocksDB
ENV ROCKSDB_SRC=/usr/local/src/rocksdb
RUN apt-get install -y gcc g++ libgflags-dev libsnappy-dev zlib1g-dev libbz2-dev liblz4-dev libzstd-dev && \
    git clone https://github.com/facebook/rocksdb.git ${ROCKSDB_SRC}
RUN cd ${ROCKSDB_SRC} && \
    make shared_lib && \
    make install-shared

ENV NEBULAS_SRC=${GOPATH}/src/github.com/nebulasio/go-nebulas
WORKDIR ${NEBULAS_SRC}

RUN go get -u github.com/golang/dep/cmd/dep && \
    go get -u golang.org/x/tools/cmd/goimports
