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

#include "fs/ir_manager/ir_processor.h"
#include "fs/proto/block.pb.h"
#include "fs/util.h"

typedef std::shared_ptr<corepb::Block> block_ptr_t;
typedef std::shared_ptr<neb::fs::ir_processor> nbre_storage_ptr_t;

std::string get_db_path_for_read();
std::string get_db_path_for_write();

std::string get_blockchain_path_for_read();
