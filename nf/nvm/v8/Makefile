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

CFLAGS=-DBUILDING_DLL -g -fPIC -m64 -Wall -Wextra -Wno-unused-parameter -Wno-unused-result -O2
CXXFLAGS=-DBUILDING_DLL -g -std=c++11 -fno-rtti -fPIC -m64 -Wall -Wextra -Wno-unused-parameter -Wno-unused-result -O2

LDFLAGS=-fvisibility=hidden
LIBS_PATH=
LIBS=-lpthread -lv8_libbase -lv8_libplatform -licui18n -licuuc

ifeq ($(shell uname -s),Darwin)
	CC=clang
	CXX=clang++
	LD=clang++
	DYLIB=.dylib
	LDFLAGS+=-Wl,-dead_strip
	LIB_LDFLAGS+=-current_version 1.0 -compatibility_version 1.0
	LIBS+=-lv8_base -lv8_libsampler -lv8_external_snapshot
	V8_BLOB_BIN_PATH=/usr/local/opt/v8/bin
else
# original gcc version is 4.8
	CC=gcc
	CXX=g++
	LD=g++
	DYLIB=.so
	CFLAGS+=-I/usr/local/v8/6.2.414.40/include
	CXXFLAGS+=-I/usr/local/v8/6.2.414.40/include
	LIBS_PATH=-L/usr/local/v8/6.2.414.40/lib
	LIBS+=-lv8
	V8_BLOB_BIN_PATH=/usr/local/v8/6.2.414.40/bin
endif

.PHONY: engine main clean v8_snapshot_bin install
all: main

%.c.o : %.c
	$(CC) $(CFLAGS) -c $< -o $<.o

%.cc.o: %.cc
	$(CXX) $(CXXFLAGS) -c $< -o $<.o

%.cpp.o: %.cpp
	$(CXX) $(CXXFLAGS) -c $< -o $<.o

main: samples/main.cc.o samples/memory_storage.cc.o samples/memory_modules.cc.o engine.cc.o allocator.cc.o lib/global.cc.o lib/execution_env.cc.o lib/storage_object.cc.o lib/log_callback.cc.o lib/require_callback.cc.o lib/instruction_counter.cc.o lib/blockchain.cc.o lib/fake_blockchain.cc.o lib/tracing.cc.o lib/file.cc.o lib/util.cc.o lib/typescript.cc.o lib/event.cc.o  lib/crypto.cc.o
	$(LD) $(LDFLAGS) $^ -o $@ $(LIBS_PATH) $(LIBS)


engine: engine.cc.o thread_engine.cc.o allocator.cc.o lib/random.cc.o lib/global.cc.o lib/execution_env.cc.o lib/storage_object.cc.o lib/log_callback.cc.o lib/require_callback.cc.o lib/instruction_counter.cc.o lib/blockchain.cc.o lib/tracing.cc.o lib/file.cc.o lib/util.cc.o lib/typescript.cc.o lib/event.cc.o lib/crypto.cc.o
	$(LD) -shared $(LDFLAGS) $^ -o libnebulasv8$(DYLIB) $(LIBS_PATH) $(LIBS)

install: engine
	-mkdir -p ../native-lib
	-install libnebulasv8$(DYLIB) ../native-lib/
clean:
	-rm -f natives_blob.bin.h snapshot_blob.bin.h;
	-find . -iname '*.o' -exec rm {} \;
	-rm -f main libnebulasv8$(DYLIB);

v8_snapshot_bin:
	-ln -s $(V8_BLOB_BIN_PATH)/natives_blob.bin .
	-ln -s $(V8_BLOB_BIN_PATH)/snapshot_blob.bin .
	-xxd -i natives_blob.bin natives_blob.bin.h
	-xxd -i snapshot_blob.bin snapshot_blob.bin.h
	-rm natives_blob.bin snapshot_blob.bin
