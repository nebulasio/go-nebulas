# Copyright (C) 2017 go-nebulas authors
#
# This file is part of the go-nebulas library.
#
# the go-nebulas library is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# the go-nebulas library is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with the go-nebulas library.  If not, see <http://www.gnu.org/licenses/>.
#
PB = $(wildcard *.proto)
GO = $(PB:.proto=.pb.go)
INCLUDE = -I ~/go/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/ -I ~/go/src -I ./

all: $(GO) nvm

%.pb.go: %.proto
	protoc -I/usr/local/include -I. -I../../../../../../src -I../../vendor/github.com/nebulasio/grpc-gateway/third_party/googleapis --gogo_out=plugins=grpc:. --grpc-gateway_out=logtostderr=true:. $<
	sed 's/grpc-ecosystem/nebulasio/g' rpc.pb.gw.go > tmp
	mv tmp rpc.pb.gw.go

nvm:
	protoc $(INCLUDE) --grpc_out=./cxx --plugin=protoc-gen-grpc=`which grpc_cpp_plugin` rpc.proto
	protoc $(INCLUDE) --cpp_out ./cxx rpc.proto

clean:
	rm *.go cxx/*.cc cxx/*.h
