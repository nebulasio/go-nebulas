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
#include "fs/items_storage.h"
#include "fs/storage.h"

namespace neb {
namespace fs {
namespace internal {
items_storage_base::items_storage_base(storage *db,
                                       const std::string &key_prefix,
                                       const std::string &last_item_key,
                                       size_t block_trunk_size)
    : m_db(db), m_key_prefix(key_prefix), m_last_item_key(last_item_key),
      m_block_trunk_size(block_trunk_size) {}

void items_storage_base::append_item(const bytes &item) {
  if (is_empty_index_table()) {
    read_index_table();
  }

  boost::unique_lock<boost::shared_mutex> _l(m_mutex);

  item_info_t last_live_block = get_last_live_block_info();
  item_contents_t block_content =
      read_block_with_key(last_live_block.get<p_item_content>());
  block_content.get<p_item_contents>().push_back(item);
  last_live_block.set<p_item_count>(last_live_block.get<p_item_count>() + 1);
  update_index_table(last_live_block);
  write_block_with_key(last_live_block.get<p_item_content>(), block_content);
  write_index_table();
}

void items_storage_base::read_index_table() {
  boost::unique_lock<boost::shared_mutex> _l(m_mutex);
  try {
    bytes bs = m_db->get_bytes(string_to_byte(m_last_item_key));
    auto bs_str = std::to_string(bs);
    m_index_table.deserialize_from_string(bs_str);
  } catch (...) {
    // ignore if not exist
  }
}
void items_storage_base::write_index_table() {
  auto bs_str = m_index_table.serialize_to_string();
  m_db->put_bytes(string_to_byte(m_last_item_key), string_to_byte(bs_str));
}

item_info_t items_storage_base::get_last_live_block_info() {
  std::vector<item_info_t> infos = m_index_table.get<p_item_keys>();

  auto new_info = [this, &infos]() {
    item_info_t item;
    bytes key = string_to_byte(m_key_prefix + std::to_string(0));
    item.set<p_item_content>(key);
    item.set<p_item_count>(0);
    infos.push_back(item);
    m_index_table.set<p_item_keys>(infos);
    return item;
  };
  if (infos.size() == 0) {
    return new_info();
  } else {
    item_info_t &last = infos[infos.size() - 1];
    if (last.get<p_item_count>() == m_block_trunk_size) {
      return new_info();
    }
    return last;
  }
}

item_contents_t items_storage_base::read_block_with_key(const bytes &key) {
  item_contents_t ret;
  try {
    auto bs = m_db->get_bytes(key);
    auto bs_str = byte_to_string(bs);
    ret.deserialize_from_string(bs_str);
  } catch (...) {
    // ignore if not exist
  }
  return ret;
}
void items_storage_base::write_block_with_key(const bytes &key,
                                              item_contents_t &contents) {
  auto bs_str = contents.serialize_to_string();
  m_db->put_bytes(key, string_to_byte(bs_str));
}

void items_storage_base::update_index_table(const item_info_t &info) {
  std::vector<item_info_t> infos = m_index_table.get<p_item_keys>();
  for (int32_t i = static_cast<int32_t>(infos.size()) - 1; i > 0; i--) {
    auto &ti = infos[i];
    if (ti.get<p_item_content>() == info.get<p_item_content>()) {
      ti.set<p_item_count>(info.get<p_item_count>());
      break;
    }
  }
  m_index_table.set<p_item_keys>(infos);
}
std::vector<bytes> items_storage_base::get_all_items() {

  if (is_empty_index_table()) {
    read_index_table();
  }
  boost::shared_lock<boost::shared_mutex> _l(m_mutex);

  std::vector<bytes> items;
  for (auto &info : m_index_table.get<p_item_keys>()) {
    item_contents_t tc = read_block_with_key(info.get<p_item_content>());
    for (auto &ic : tc.get<p_item_contents>()) {
      items.push_back(ic);
    }
  }
  return items;
}
} // namespace internal
} // namespace fs
} // namespace neb
