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
  rs.open_database("test_write.db", neb::fs::storage_open_for_readwrite);

  char c = '0';
  int64_t i = 0;
  int64_t key = 1;
  while (c != 'x') {
    rs.put(key, i);
    std::cout << "put " << i << std::endl;
    i++;
    std::cin >> c;
  }

  rs.close_database();

  return 0;
}
