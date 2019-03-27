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

#pragma once

#include "fs/blockchain.h"
#include "fs/rocksdb_storage.h"
#include <memory>
#include <string>

extern std::string cur_path;
extern std::string db_read_path;
extern std::string db_write_path;

extern std::unique_ptr<neb::fs::rocksdb_storage> db_read_ptr;
extern std::unique_ptr<neb::fs::rocksdb_storage> db_write_ptr;

extern std::unique_ptr<neb::fs::blockchain> blockchain_ptr;
extern std::unique_ptr<neb::fs::nbre_storage> nbre_ptr;
