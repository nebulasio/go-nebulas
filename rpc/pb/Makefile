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

all: $(GO)

%.pb.go: %.proto
	protoc -I/usr/local/include -I. -I../../../../../../src -I../../../../../../pkg/mod/github.com/nebulasio/grpc-gateway@v1.11.2/third_party/googleapis --gogo_out=plugins=grpc:. --grpc-gateway_out=logtostderr=true:. $<
clean:
	rm -rf *.go
