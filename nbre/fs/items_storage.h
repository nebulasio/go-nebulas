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
#include "common/byte.h"
#include "common/common.h"
#include "fs/rocksdb_storage.h"

define_nt(p_item_prev_block_key, neb::bytes);

namespace neb {
namespace fs {
template <class ItemType, class ItemContainerType> class items_storage {
public:
  typedef ItemType item_type;
  typedef ItemContainerType item_containter_type;
  items_storage(const std::string &key_prefix,
                const std::string &latest_item_key,
                size_t block_trunk_size = 4);

  virtual void append_item(const item_type &item);
  virtual std::vector<item_type> &get_all_items();

protected:
  typedef ::ff::net::ntpackage<1, p_item_prev_block_key, ItemContainerType>
      block_t;
};
} // namespace fs
} // namespace neb
