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
#include "common/address.h"
#include "fs/proto/ir.pb.h"
namespace corepb {
class Transaction;
}
namespace neb {
namespace util {
//! check if tx is Nebulas Protocol Rerepsentation
bool is_npr_tx(const corepb::Transaction &tx);
nbre::NBREIR extract_npr(const corepb::Transaction &tx);
bool is_auth_npr(const address_t &from, const std::string &module_name);
}
} // namespace neb
