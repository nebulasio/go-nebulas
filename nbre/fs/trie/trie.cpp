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

#include "fs/trie/trie.h"
#include "fs/trie/byte_shared.h"

namespace neb {
namespace fs {

trie_node::trie_node(const neb::util::bytes &triepb_bytes) {

  std::shared_ptr<triepb::Node> triepb_node_ptr =
      std::make_shared<triepb::Node>();

  bool ret = triepb_node_ptr->ParseFromArray(triepb_bytes.value(),
                                             triepb_bytes.size());
  if (!ret) {
    throw std::runtime_error("parse triepb node failed");
  }

  m_bytes = triepb_bytes;
  // TODO sha3256 implement
  m_hash = triepb_bytes;
  for (auto &v : triepb_node_ptr->val()) {
    m_val.push_back(neb::util::string_to_byte(v));
  }
}

trie_node_type trie_node::get_trie_node_type() {

  if (m_val.size() == 16) {
    return trie_node_type::trie_node_branch;
  }
  if (m_val.size() == 3 && !m_val[0].empty()) {
    return static_cast<trie_node_type>(m_val[0][0]);
  }
  return trie_node_type::trie_node_unknown;
}

trie::trie(const std::string &neb_db_path) {
  m_storage = std::make_unique<rocksdb_storage>();
  m_storage->open_database(neb_db_path, storage_open_for_readonly);
}
trie::~trie() {
  if (!m_storage) {
    m_storage->close_database();
  }
}

std::shared_ptr<trie_node> trie::fetch_node(const neb::util::bytes &hash) {

  neb::util::bytes triepb_bytes = m_storage->get_bytes(hash);
  return std::make_shared<trie_node>(triepb_bytes);
}

neb::util::bytes trie::get_trie_node(const neb::util::bytes &root_hash,
                                     const neb::util::bytes &key) {
  auto hash = root_hash;
  auto route = key_to_route(key);

  for (int32_t route_len = static_cast<int32_t>(route.size()); route_len >= 0;
       route_len--) {
    auto root_node = fetch_node(root_hash);
    auto root_type = root_node->get_trie_node_type();
    if (route.empty() && root_type != trie_node_type::trie_node_leaf) {
      throw std::runtime_error("key/path too short");
    }

    if (root_type == trie_node_type::trie_node_branch) {
      hash = root_node->val_at(route[0]);
      assert(route.size() >= 1);
      route = neb::util::bytes(route.value() + 1, route.size() - 1);

    } else if (root_type == trie_node_type::trie_node_extension) {
      auto key_path = root_node->val_at(1);
      auto next_hash = root_node->val_at(2);

      size_t matched_len = prefix_len(key_path, next_hash);
      if (matched_len == key_path.size()) {
        hash = next_hash;
        assert(route.size() >= matched_len);
        route = neb::util::bytes(route.value() + matched_len,
                                 route.size() - matched_len);
      }
    } else if (root_type == trie_node_type::trie_node_leaf) {
      auto key_path = root_node->val_at(1);
      size_t matched_len = prefix_len(key_path, route);
      if (matched_len == key_path.size() && matched_len == route.size()) {
        return root_node->val_at(2);
      }
    }
  }

  throw std::runtime_error("key path not found");
  return hash;
}

neb::util::bytes trie::key_to_route(const neb::util::bytes &key) {

  std::unique_ptr<byte_t[]> value;
  size_t size = key.size() << 1;

  if (size > 0) {
    value = std::unique_ptr<byte_t[]>(new byte_t[size]);
    for (size_t i = 0; i < key.size(); i++) {
      byte_shared byte(key[i]);
      value.get()[i << 1] = byte.bits_high();
      value.get()[(i << 1) + 1] = byte.bits_low();
    }
  }
  return neb::util::bytes(value.get(), size);
}

neb::util::bytes trie::route_to_key(const neb::util::bytes &route) {

  std::unique_ptr<byte_t[]> value;
  size_t size = route.size() >> 1;

  if (size > 0) {
    value = std::unique_ptr<byte_t[]>(new byte_t[size]);
    for (size_t i = 0; i < size; i++) {
      byte_shared byte(route[(i << 1) + 1], route[i << 1]);
      value.get()[i] = byte.data();
    }
  }
  return neb::util::bytes(value.get(), size);
}

size_t trie::prefix_len(const neb::util::bytes &s, const neb::util::bytes &t) {
  size_t min_len = std::min(s.size(), t.size());
  for (size_t i = 0; i < min_len; i++) {
    if (s[i] != t[i]) {
      return i;
    }
  }
  return min_len;
}
} // namespace fs
} // namespace neb
