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

CC=clang
CFLAGS=-g `llvm-config --cflags`
CXXFLAGS=-g `llvm-config --cxxflags`
LD=clang++
LDFLAGS=`llvm-config --cxxflags --ldflags --libs core executionengine mcjit asmprinter asmparser native irreader --system-libs`

all: main

%.c.o : %.c
	$(CC) $(CFLAGS) -c $< -o $<.o

%.cpp.o: %.cpp
	$(CC) $(CXXFLAGS) -c $< -o $<.o

main: main.cpp.o
	$(LD) $< $(LDFLAGS) -o $@

engine_test: engine.cpp.o engine_test.cpp.o memory_manager.cpp.o
	$(LD) $^ $(LDFLAGS) -o $@

sum: sum.c.o
	$(LD) $^ $(LDFLAGS) -o $@

clean:
	-rm -f *.o main sum engine_test

ir:
	clang -S -emit-llvm sample.c
