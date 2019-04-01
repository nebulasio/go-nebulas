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
#include "fs/rocksdb_storage.h"
#include <iostream>

int main(int argc, char *argv[]) {

  neb::fs::rocksdb_storage rs;
  rs.open_database("test_write.db", neb::fs::storage_open_for_readonly);

  char c = '0';
  int64_t key = 1;
  std::thread thrd([&]() {
    while (c != 'x') {
      std::this_thread::sleep_for(std::chrono::seconds(1));
      int64_t v = rs.get<int64_t>(key);
      std::cout << "get " << v << std::endl;
    }
    rs.close_database();
  });
  std::cout << "press x to quit" << std::endl;
  std::cin >> c;
  thrd.join();

  return 0;
}
