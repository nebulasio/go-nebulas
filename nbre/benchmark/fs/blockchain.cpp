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
#include "benchmark/fs/common.h"

std::unique_ptr<neb::fs::blockchain> blockchain_ptr;

BENCHMARK(blockchain, blockchain_init) {
  blockchain_ptr = std::make_unique<neb::fs::blockchain>(
      db_read_path, neb::fs::storage_open_for_readonly);
}

BENCHMARK(blockchain, load_tail_block) { blockchain_ptr->load_tail_block(); }

BENCHMARK(blockchain, load_LIB_block) { blockchain_ptr->load_LIB_block(); }

BENCHMARK(blockchain, load_block_with_height) {
  blockchain_ptr->load_block_with_height(100);
}

BENCHMARK(blockchain, blockchain_destroy) { blockchain_ptr.reset(nullptr); }
