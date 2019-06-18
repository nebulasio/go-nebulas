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

define_nt(p_item_content, neb::bytes);
define_nt(p_item_count, size_t);
typedef ff::net::ntpackage<1, p_item_content, p_item_count> item_info_t;
define_nt(p_item_keys, std::vector<item_info_t>);
define_nt(p_item_contents, std::vector<neb::bytes>);
typedef ::ff::net::ntpackage<1, p_item_keys> item_index_table_t;
typedef ::ff::net::ntpackage<2, p_item_contents> item_contents_t;

namespace neb {
namespace fs {
class storage;
namespace internal {
class items_storage_base {
public:
  items_storage_base(storage *db, const std::string &key_prefix,
                     const std::string &last_item_key,
                     size_t block_trunk_size = 16);

  virtual void append_item(const bytes &item);
  virtual std::vector<bytes> get_all_items() const;

protected:
  void read_index_table();
  void write_index_table();
  inline bool is_empty_index_table() const {
    return m_index_table.get<p_item_keys>().empty();
  }
  item_info_t get_last_live_block_info();
  item_contents_t read_block_with_key(const bytes &key) const;
  void write_block_with_key(const bytes &key, item_contents_t &contents);
  void update_index_table(const item_info_t &info);

protected:
  storage *m_db;
  std::string m_key_prefix;
  std::string m_last_item_key;
  size_t m_block_trunk_size;
  mutable item_index_table_t m_index_table;
  mutable boost::shared_mutex m_mutex;
  };
  } // namespace internal

  } // namespace fs
} // namespace neb
