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
#include "common/configuration.h"
#include "fs/util.h"
#include <vector>

std::string cur_path;
std::string db_read_path;
std::string db_write_path;

std::unique_ptr<neb::fs::rocksdb_storage> db_read_ptr;
std::unique_ptr<neb::fs::rocksdb_storage> db_write_ptr;

BENCHMARK(rocksdb_storage, rocksdb_storage_init) {
  cur_path = neb::configuration::instance().root_dir();
  db_read_path = neb::fs::join_path(cur_path, "test/data/read-data.db");
  db_write_path = neb::fs::join_path(cur_path, "test/data/write-data.db");

  db_read_ptr = std::make_unique<neb::fs::rocksdb_storage>();
  db_write_ptr = std::make_unique<neb::fs::rocksdb_storage>();
}

BENCHMARK(rocksdb_storage, open_close_database) {
  db_read_ptr->open_database(db_read_path, neb::fs::storage_open_for_readonly);
  db_read_ptr->close_database();
}

BENCHMARK(rocksdb_storage, put_bytes) {
  db_write_ptr->open_database(db_write_path,
                              neb::fs::storage_open_for_readwrite);

  std::string key = "benchmark_rocksdb_storage";
  size_t eval_count = 10;
  for (size_t i = 0; i < eval_count; i++) {
    std::string k = key + '_' + std::to_string(i);
    std::string v = neb::util::string_to_byte(k).to_hex();
    db_write_ptr->put_bytes(neb::util::string_to_byte(k),
                            neb::util::string_to_byte(v));
  }
  db_write_ptr->close_database();
}

BENCHMARK(rocksdb_storage, get_bytes) {
  db_write_ptr->open_database(db_write_path,
                              neb::fs::storage_open_for_readonly);

  std::string key = "benchmark_rocksdb_storage";
  size_t eval_count = 10;
  for (size_t i = 0; i < eval_count; i++) {
    std::string k = key + '_' + std::to_string(i);
    db_write_ptr->get_bytes(neb::util::string_to_byte(k));
  }
  db_write_ptr->close_database();
}

BENCHMARK(rocksdb_storage, del_bytes) {
  db_write_ptr->open_database(db_write_path,
                              neb::fs::storage_open_for_readwrite);

  std::string key = "benchmark_rocksdb_storage";
  size_t eval_count = 10;
  for (size_t i = 0; i < eval_count; i++) {
    std::string k = key + '_' + std::to_string(i);
    db_write_ptr->del_by_bytes(neb::util::string_to_byte(k));
  }
  db_write_ptr->close_database();
}

BENCHMARK(rocksdb_storage, put) {
  db_write_ptr->open_database(db_write_path,
                              neb::fs::storage_open_for_readwrite);

  std::string key = "benchmark_rocksdb_storage";
  size_t eval_count = 10;
  for (size_t i = 0; i < eval_count; i++) {
    std::string k = key + '_' + std::to_string(i);
    std::string v = neb::util::string_to_byte(k).to_hex();
    db_write_ptr->put(k, neb::util::string_to_byte(v));
  }
  db_write_ptr->close_database();
}

BENCHMARK(rocksdb_storage, get) {
  db_write_ptr->open_database(db_write_path,
                              neb::fs::storage_open_for_readonly);

  std::string key = "benchmark_rocksdb_storage";
  size_t eval_count = 10;
  for (size_t i = 0; i < eval_count; i++) {
    std::string k = key + '_' + std::to_string(i);
    db_write_ptr->get(k);
  }
  db_write_ptr->close_database();
}

BENCHMARK(rocksdb_storage, del) {
  db_write_ptr->open_database(db_write_path,
                              neb::fs::storage_open_for_readwrite);

  std::string key = "benchmark_rocksdb_storage";
  size_t eval_count = 10;
  for (size_t i = 0; i < eval_count; i++) {
    std::string k = key + '_' + std::to_string(i);
    db_write_ptr->del(k);
  }
  db_write_ptr->close_database();
}

BENCHMARK(rocksdb_storage, rocksdb_storage_destroy) {
  db_read_ptr.reset(nullptr);
  db_write_ptr.reset(nullptr);
}
