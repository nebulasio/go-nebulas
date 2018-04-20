FROM centos

RUN yum -y update && yum -y install epel-release
RUN yum -y update && yum -y install which make git protobuf-compiler gflags-devel snappy-devel zlib-devel bzip2-devel  gcc-c++  libstdc++-devel wget

RUN wget https://golangtc.com/static/go/1.9.2/go1.9.2.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.9.2.linux-amd64.tar.gz

RUN mkdir go
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

# Install RocksDB
RUN mkdir /dependency && cd /dependency
RUN git clone https://github.com/facebook/rocksdb.git
RUN cd rocksdb && make shared_lib && make install-shared
ENV LD_LIBRARY_PATH $LD_LIBRARY_PATH:/usr/local/lib

# Build Nebulas
RUN go get -u github.com/golang/dep/cmd/dep && \
    go get -u golang.org/x/tools/cmd/goimports

WORKDIR /go/src/github.com/nebulasio
RUN git clone https://github.com/nebulasio/go-nebulas.git

WORKDIR /go/src/github.com/nebulasio/go-nebulas
RUN git checkout master

# Install dependency
RUN wget http://ory7cn4fx.bkt.clouddn.com/vendor.tar.gz
RUN tar xzf vendor.tar.gz
## Or use `make dep` if you have ladders.

RUN install nf/nvm/native-lib/*.so /usr/local/lib/ && make build

ENTRYPOINT ["/go/src/github.com/nebulasio/go-nebulas/neb"]
