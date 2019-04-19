FROM ubuntu:16.04

ENV GOPATH /go
ENV PATH ${GOPATH}/bin:/usr/local/go/bin:${PATH}

RUN apt update && \
    apt install -y git build-essential protobuf-compiler sudo wget

# Install Go1.11.5
RUN wget https://dl.google.com/go/go1.11.5.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.11.5.linux-amd64.tar.gz

ENV NEBULAS_SRC=${GOPATH}/src/github.com/nebulasio/go-nebulas
WORKDIR ${NEBULAS_SRC}

# RUN go get -u golang.org/x/tools/cmd/goimports

ENV LD_LIBRARY_PATH=${NEBULAS_SRC}/native-lib:${LD_LIBRARY_PATH}