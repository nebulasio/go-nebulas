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

std::unique_ptr<neb::fs::nbre_storage> nbre_ptr;

BENCHMARK(nbre_storage, nbre_storage_init) {
  nbre_ptr.reset(nullptr);
  nbre_ptr =
      std::make_unique<neb::fs::nbre_storage>(db_write_path, db_read_path);
}

BENCHMARK(nbre_storage, read_nbre_by_height) {
  nbre_ptr->read_nbre_by_height("nr", 1000);
}

BENCHMARK(nbre_storage, read_nbre_by_name_version) {
  nbre_ptr->read_nbre_by_name_version("nr", 2LL << 48);
}

BENCHMARK(nbre_storage, write_nbre) { nbre_ptr->write_nbre(); }

BENCHMARK(nbre_storage, is_latest_irreversible_block) {
  nbre_ptr->is_latest_irreversible_block();
}

BENCHMARK(nbre_storage, nbre_storage_destroy) { nbre_ptr.reset(nullptr); }
