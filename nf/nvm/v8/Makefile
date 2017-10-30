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
CXX=clang++
LD=clang++

CFLAGS=-g
CXXFLAGS=-g -std=c++11

LDFLAGS=
LIBS=-lv8 -lv8_libbase -lv8_libplatform

all: main

%.c.o : %.c
	$(CC) $(CFLAGS) -c $< -o $<.o

%.cc.o: %.cc
	$(CXX) $(CXXFLAGS) -c $< -o $<.o

%.cpp.o: %.cpp
	$(CXX) $(CXXFLAGS) -c $< -o $<.o

main: main.cc.o
	$(LD) $(LDFLAGS) $(LIBS) $^ -o $@

engine_test: engine.cc.o lib/log_callback.cc.o lib/require_callback.cc.o engine_test.cc.o
	$(LD) $(LDFLAGS) $(LIBS) $^ -o $@

clean:
	-find . -iname '*.o' -exec rm {} \;
