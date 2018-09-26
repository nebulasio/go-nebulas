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
#include "benchmark/common/benchmark_instances.h"

namespace neb {
benchmark_instances::~benchmark_instances() {
  for (auto it = m_all_instances.begin(); it != m_all_instances.end(); ++it) {
    (*it)->run();
  }
}

void benchmark_instances::init_benchmark_instances(int argc, char *argv[]) {}
int benchmark_instances::run_all_benchmarks() { return 0; }

size_t
benchmark_instances::register_benchmark(const benchmark_instance_base_ptr &b) {
  m_all_instances.push_back(b);
  return m_all_instances.size();
}
} // end namespace neb
