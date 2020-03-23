
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

SHELL := /bin/bash 

VERSION?=3.0.0

COMMIT=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

CURRENT_DIR=$(shell pwd)
BUILD_DIR=${CURRENT_DIR}
BINARY=neb

VET_REPORT=vet.report
LINT_REPORT=lint.report
TEST_REPORT=test.report
TEST_XUNIT_REPORT=test.report.xml

OS := $(shell uname -s)
ifeq ($(OS),Darwin)
	DYLIB=.dylib
	INSTALL=install
	LDCONFIG=
	NEBBINARY=$(BINARY)
	BUUILDLOG=
	DBCHECK=brew list | grep -c rocksdb
else
	DYLIB=.so
	INSTALL=sudo install
	LDCONFIG=sudo /sbin/ldconfig
	NEBBINARY=$(BINARY)-$(COMMIT)
	BUUILDLOG=-rm -f $(BINARY); ln -s $(BINARY)-$(COMMIT) $(BINARY)
	DBCHECK=ldconfig -p | grep -c rocksdb
endif

NATIVELIB := native-lib
ifeq ($(NATIVELIB),$(wildcard $(NATIVELIB)))
    CGO_CFLAGS=
	CGO_LDFLAGS=CGO_LDFLAGS="-L$(CURRENT_DIR)/native-lib -lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy -llz4 -lzstd"
else
    CGO_CFLAGS=
	CGO_LDFLAGS=
endif

# ifeq ($(shell command -v dep 2> /dev/null || echo "uninstalled"),uninstalled)
# 	DEPINSTALL=mkdir -p $(GOPATH)/bin && curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
# else
# 	DEPINSTALL=
# endif

ifeq ($(shell command -v go 2> /dev/null || echo "uninstalled"),uninstalled)
	GOCHECK=$(error  "go not installed. run `source setup.sh` first!")
else
	# GOCHECK=$(info  "go installed.")
	GOCHECK=
endif

ifneq ($(DBCHECK),0)
	DBCHECK=
else
	DBCHECK=$(error  "rocksdb not installed. run `source setup.sh` first!")
endif

ifneq ($(shell ls $(CURRENT_DIR)/nf/nvm/native-lib |grep -c libnebulas),0)
	LIBCHECK=
else
	LIBCHECK=$(error  "nvm not installed. run `source setup.sh` first!")
endif

# $(warning  $(CGO_LDFLAGS))

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS = -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.branch=${BRANCH} -X main.compileAt=`date +%s`"

# Build the project
.PHONY: build build-linux clean dep lint run test vet link-libs all

all: clean vet fmt lint build test

#dep:
#	$(DEPINSTALL) dep ensure -v

check: check-go check-db check-lib

check-go:
	$(GOCHECK)
    
check-db:
	$(DBCHECK)

check-lib:
	$(LIBCHECK)

build: check build-neb

build-neb:
	cd cmd/neb; GOPROXY=https://goproxy.io $(CGO_CFLAGS) $(CGO_LDFLAGS) go build $(LDFLAGS) -o ../../$(NEBBINARY)
	$(BUUILDLOG)

#build-linux:
#	cd cmd/neb; GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o ../../$(BINARY)-linux

LIST := ./account/... ./cmd/... ./common/... ./consensus ./core/... ./crypto/... ./metrics/... ./neblet/... ./net/... ./nf/... ./rpc/... ./storage/... ./sync/... ./util/... ./nip/... ./nr/...
# LIST := $(ls -d */|grep -Ev "vendor|logs|nebtestkit|.db")/...

test:
	$(CGO_CFLAGS) $(CGO_LDFLAGS) go test $(LIST) 2>&1 | tee $(TEST_REPORT); go2xunit -fail -input $(TEST_REPORT) -output $(TEST_XUNIT_REPORT)

vet:
	go vet $$(go list $(LIST)) 2>&1 | tee $(VET_REPORT)

fmt:
	goimports -w $$(go list -f "{{.Dir}}" $(LIST) | grep -v /vendor/)

lint:
	golint $$(go list $(LIST)) | sed "s:^$(BUILD_DIR)/::" | tee $(LINT_REPORT)

clean:
	-rm -f $(VET_REPORT)
	-rm -f $(LINT_REPORT)
	-rm -f $(TEST_REPORT)
	-rm -f $(TEST_XUNIT_REPORT)
	-rm -f $(BINARY)
	-rm -f $(BINARY)-$(COMMIT)

