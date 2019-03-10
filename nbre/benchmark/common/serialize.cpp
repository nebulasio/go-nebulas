// Copyright (C) 2018 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or
// modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see
// <http://www.gnu.org/licenses/>.
//
#include "benchmark/benchmark_instances.h"
#include "common/byte.h"
#include "common/math.h"
#include "fs/proto/ir.pb.h"
#include <ff/network.h>
#include <iostream>

define_nt(name, std::string);
define_nt(version, uint64_t);
define_nt(height, uint64_t);
define_nt(ir, std::string);
define_nt(ir_type, std::string);

typedef ff::net::ntpackage<1, name, version, height, ir, ir_type> my_ir_t;

BENCHMARK(ir_size, proto) {
  nbre::NBREIR ni;
  ni.set_name("nebulas");
  ni.set_version(0xffffffffffffffff);
  ni.set_height(0xffffffffffffffff);
  ni.set_ir("testnet1987");
  ni.set_ir_type("llvm");
  auto bytes_long = ni.ByteSizeLong();
  std::cout << "proto bytes: " << bytes_long << std::endl;
  neb::bytes rs(bytes_long);
  ni.SerializeToArray((void *)rs.value(), rs.size());
}

BENCHMARK(ir_size, nt) {
  my_ir_t mi;
  mi.set<name>("nebulas");
  mi.set<version>(0xffffffffffffffff);
  mi.set<height>(0xffffffffffffffff);
  mi.set<ir>("testnet");
  mi.set<ir_type>("llvm");
  ff::net::marshaler lr(ff::net::marshaler::length_retriver);
  mi.arch(lr);
  auto bytes_long = lr.get_length();
  char *buf = new char[bytes_long];
  std::cout << "nt bytes: " << bytes_long << std::endl;
  ff::net::marshaler tr(buf, bytes_long, ff::net::marshaler::seralizer);
  mi.arch(tr);
}

