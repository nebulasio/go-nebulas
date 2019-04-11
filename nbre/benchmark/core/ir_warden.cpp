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
#include "benchmark/core/common.h"

std::unique_ptr<neb::core::ir_warden> ir_warden_ptr;

BENCHMARK(ir_warden, ir_warden_init) {
  ir_warden_ptr.reset();
  ir_warden_ptr = std::make_unique<neb::core::ir_warden>();
}

BENCHMARK(ir_warden, get_ir_by_name_version) {
  ir_warden_ptr->get_ir_by_name_version("nr", 1LL << 48);
  ir_warden_ptr->get_ir_by_name_version("nr", 2LL << 48);
  ir_warden_ptr->get_ir_by_name_version("nr", 3LL << 48);
  ir_warden_ptr->get_ir_by_name_version("dip", 1);
}

BENCHMARK(ir_warden, get_ir_by_name_height) {
  size_t eval_count = 10;
  for (size_t h = 1; h < eval_count; h++) {
    ir_warden_ptr->get_ir_by_name_height("nr", h);
    ir_warden_ptr->get_ir_by_name_height("dip", h);
  }
}

BENCHMARK(ir_warden, is_sync_already) { ir_warden_ptr->is_sync_already(); }

BENCHMARK(ir_warden, ir_warden_destroy) { ir_warden_ptr.reset(); }
