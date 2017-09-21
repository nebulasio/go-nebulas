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
VERSION?=0.1.0
COMMIT=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

CURRENT_DIR=$(shell pwd)
BUILD_DIR=${CURRENT_DIR}
BINARY=neb

VET_REPORT=vet.report
LINT_REPORT=lint.report
TEST_REPORT=test.report
TEST_XUNIT_REPORT=test.report.xml

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS = -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.branch=${BRANCH} -X main.compileAt=`date +%s`"

# Build the project
.PHONY: build build-linux clean dep lint run test

all: clean vet fmt lint build test

dep:
	dep ensure

build:
	cd cmd/neb; go build ${LDFLAGS} -o ../../${BINARY}

build-linux:
	cd cmd/neb; GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ../../${BINARY}-linux

test:
	go test -v ./... 2>&1 | tee ${TEST_REPORT}; go2xunit -input ${TEST_REPORT} -output ${TEST_XUNIT_REPORT};

vet:
	go vet $$(go list ./...) 2>&1 | tee ${VET_REPORT}

fmt:
	go fmt $$(go list ./...)

lint:
	golint $$(go list ./...) | sed "s:^${BUILD_DIR}/::" | tee ${LINT_REPORT}

clean:
	rm -f ${VET_REPORT}
	rm -f ${LINT_REPORT}
	rm -f ${TEST_REPORT}
	rm -f ${TEST_XUNIT_REPORT}
	rm -f ${BINARY}*
