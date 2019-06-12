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
#include "fs/items_storage.h"

namespace neb {
namespace fs {

template <class ItemType>
class nt_items_storage : public internal::items_storage_base {
public:
  typedef ItemType item_type;
  nt_items_storage(storage *db, const std::string &key_prefix,
                   const std::string &latest_item_key,
                   size_t block_trunk_size = 16)
      : items_storage_base(db, key_prefix, latest_item_key, block_trunk_size) {}

  virtual void append_item(item_type &item) {
    auto bs_str = item.serialize_to_string();
    append_item(string_to_byte(bs_str));
  }
  virtual std::vector<item_type> &get_all_items() {
    std::vector<item_type> its;
    std::vector<bytes> bs = get_all_items();
    for (bytes &b : bs) {
      auto b_str = std::to_string(b);
      item_type t;
      t.deserialize_from_string(b_str);
      its.push_back(t);
    }
    return its;
  }
};

} // namespace fs
} // namespace neb
