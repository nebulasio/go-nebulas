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

CFLAGS=-DBUILDING_DLL -fPIC -m64 -Wall -Wextra -Wno-unused-parameter -Wno-unused-result -O3
CXXFLAGS=-DBUILDING_DLL -std=c++14 -DGOOGLE_PROTOBUF_NO_RTTI -fno-rtti -fPIC -m64 -Wall -Wextra -Wno-unused-parameter -Wno-unused-result -O3
OUTPUT_DIR=pb/

LDFLAGS=-fvisibility=hidden 
LIBS_PATH=
LIBS=-lgrpc++ -lgrpc++_error_details -lprotobuf -lpthread -lv8_libbase -lv8_libplatform -licui18n -licuuc -lglog

ifeq ($(shell uname -s),Darwin)
	CC=clang
	CXX=clang++
	LD=clang++
	DYLIB=.dylib
	CFLAGS+=-I./include
	CXXFLAGS+=-I./include
	LDFLAGS+=-Wl,-dead_strip
	LIBS_PATH+=-L../native-lib/osx
	LIB_LDFLAGS+=-current_version 1.0 -compatibility_version 1.0
	LIBS+=-lv8_base -lv8_libsampler -lv8_external_snapshot
	V8_BLOB_BIN_PATH=/usr/local/opt/v8/bin
else
	CC=gcc
	CXX=g++
	LD=g++
	DYLIB=.so
	CFLAGS+=-I./include
	CXXFLAGS+=-I./include
        LIBS_PATH=-L../native-lib/linux
	LIB_LDFLAGS+=-current_version 1.0 -compatibility_version 1.0
        LIBS+=-lv8
endif

.PHONY: engine v8_server clean v8_snapshot_bin install
all: v8_server

%.c.o : %.c
	$(CC) $(CFLAGS) -c $< -o $<.o

%.cc.o: %.cc
	$(CXX) $(CXXFLAGS) -c $< -o $<.o

%.cpp.o: %.cpp
	$(CXX) $(CXXFLAGS) -c $< -o $<.o

engine: engine.cc.o thread_engine.cc.o allocator.cc.o lib/global.cc.o lib/execution_env.cc.o lib/storage_object.cc.o lib/log_callback.cc.o lib/require_callback.cc.o lib/instruction_counter.cc.o lib/blockchain.cc.o lib/tracing.cc.o lib/file.cc.o lib/util.cc.o lib/typescript.cc.o lib/event.cc.o lib/crypto.cc.o
	$(LD) -shared $(LDFLAGS) $^ -o libnebulasv8$(DYLIB) $(LIBS_PATH) $(LIBS)

v8_server: v8_server.cc.o sha256.cc.o pb/nvm.grpc.pb.o pb/nvm.pb.o samples/memory_storage.cc.o samples/memory_modules.cc.o engine.cc.o allocator.cc.o lib/global.cc.o lib/execution_env.cc.o lib/storage_object.cc.o lib/log_callback.cc.o lib/require_callback.cc.o lib/instruction_counter.cc.o lib/blockchain.cc.o lib/fake_blockchain.cc.o lib/tracing.cc.o lib/file.cc.o lib/util.cc.o lib/typescript.cc.o lib/event.cc.o  lib/crypto.cc.o
	$(LD) $^ -o $@ $(LIBS_PATH) $(LDFLAGS) $(LIBS)

main: samples/main.cc.o samples/memory_storage.cc.o samples/memory_modules.cc.o engine.cc.o allocator.cc.o lib/global.cc.o lib/execution_env.cc.o lib/storage_object.cc.o lib/log_callback.cc.o lib/require_callback.cc.o lib/instruction_counter.cc.o lib/blockchain.cc.o lib/fake_blockchain.cc.o lib/tracing.cc.o lib/file.cc.o lib/util.cc.o lib/typescript.cc.o lib/event.cc.o  lib/crypto.cc.o
	$(LD) $(LDFLAGS) $^ -o $@ $(LIBS_PATH) $(LIBS)

install: engine
	-mkdir -p ../native-lib
	-install libnebulasv8$(DYLIB) ../native-lib/

clean:
#	-rm -f natives_blob.bin.h snapshot_blob.bin.h;
	-find . -iname '*.o' -exec rm {} \;
	-rm -f v8_server libnebulasv8$(DYLIB);

v8_snapshot_bin:
	-ln -s $(V8_BLOB_BIN_PATH)/natives_blob.bin .
	-ln -s $(V8_BLOB_BIN_PATH)/snapshot_blob.bin .
	-xxd -i natives_blob.bin natives_blob.bin.h
	-xxd -i snapshot_blob.bin snapshot_blob.bin.h
	-rm natives_blob.bin snapshot_blob.bin

rpc:
	protoc --grpc_out=$(OUTPUT_DIR) --plugin=protoc-gen-grpc=`which grpc_cpp_plugin` nvm.proto
	protoc --cpp_out $(OUTPUT_DIR) nvm.proto
	protoc --go_out=plugins=grpc:$(OUTPUT_DIR) nvm.proto
	cp pb/nvm.pb.go ../
	rm -rf pb/nvm.pb.go

