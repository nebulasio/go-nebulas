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
VERSION?=0.1
COMMIT=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

CURRENT_DIR=$(shell pwd)
BUILD_DIR=${CURRENT_DIR}
BINARY=neb

VET_REPORT=vet.report
LINT_REPORT=lint.report

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS = -ldflags "-X main.VERSION=${VERSION} -X main.COMMIT=${COMMIT} -X main.BRANCH=${BRANCH}"

# Build the project
.PHONY: build clean dep lint run

all: clean vet fmt lint build

dep:
	cd ${BUILD_DIR}
	dep ensure

build:
	cd cmd/neb; go build ${LDFLAGS} -o ../../${BINARY}

vet:
	go vet $$(go list ./...) 2>&1 | tee ${VET_REPORT}

fmt:
	go fmt $$(go list ./...)

lint:
	golint $$(go list ./...) | sed "s:^${BUILD_DIR}/::" | tee ${LINT_REPORT}

clean:
	rm -f ${VET_REPORT}
	rm -f ${LINT_REPORT}
	rm -f ${BINARY}
