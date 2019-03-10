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
#include "common/common.h"
#include "common/configuration.h"
#include "common/version.h"
#include "core/command.h"
#include "core/net_ipc/ipc_interface.h"
#include "core/net_ipc/nipc_pkg.h"
#include "fs/bc_storage_session.h"
#include "fs/proto/block.pb.h"
#include "fs/proto/ir.pb.h"
#include "fs/proto/trie.pb.h"
#include "util/bc_generator.h"
#include "util/quitable_thread.h"
#include "util/singleton.h"
#include <algorithm>
#include <boost/algorithm/string/replace.hpp>
#include <ff/functionflow.h>

using bc_storage_session = neb::fs::bc_storage_session;
using generate_block = neb::util::generate_block;
using all_accounts = neb::util::all_accounts;
using nas = neb::nas;
using address_t = neb::address_t;
using block_height_t = neb::block_height_t;

