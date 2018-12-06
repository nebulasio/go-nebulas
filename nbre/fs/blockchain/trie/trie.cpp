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

#include "fs/blockchain/trie/trie.h"
#include "fs/blockchain/trie/byte_shared.h"

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

trie::trie(rocksdb_storage *db_ptr) : m_storage(db_ptr) {}

std::unique_ptr<trie_node> trie::fetch_node(const neb::util::bytes &hash) {

  neb::util::bytes triepb_bytes = m_storage->get_bytes(hash);
  return std::make_unique<trie_node>(triepb_bytes);
}

bool trie::get_trie_node(const neb::util::bytes &root_hash,
                         const neb::util::bytes &key,
                         neb::util::bytes &trie_node) {
  auto hash = root_hash;
  auto route = key_to_route(key);

  neb::byte_t *route_ptr = route.value();
  size_t route_size = route.size();
  neb::byte_t *end_ptr = route.value() + route_size;

  try {

    while (route_ptr <= end_ptr) {

      auto root_node = fetch_node(hash);
      auto root_type = root_node->get_trie_node_type();
      if (route_ptr == end_ptr && root_type != trie_node_type::trie_node_leaf) {
        throw std::runtime_error("key/path too short");
      }

      if (root_type == trie_node_type::trie_node_branch) {
        hash = root_node->val_at(route_ptr[0]);
        route_ptr++;

      } else if (root_type == trie_node_type::trie_node_extension) {
        auto key_path = root_node->val_at(1);
        auto next_hash = root_node->val_at(2);
        size_t left_size = end_ptr - route_ptr;

        size_t matched_len =
            prefix_len(key_path.value(), key_path.size(), route_ptr, left_size);
        if (matched_len != key_path.size()) {
          throw std::runtime_error("node extension, key path not found");
        }
        hash = next_hash;
        route_ptr += matched_len;
      } else if (root_type == trie_node_type::trie_node_leaf) {
        auto key_path = root_node->val_at(1);
        size_t left_size = end_ptr - route_ptr;
        size_t matched_len =
            prefix_len(key_path.value(), key_path.size(), route_ptr, left_size);
        if (matched_len != key_path.size() || matched_len != left_size) {
          throw std::runtime_error("node leaf, key path not found");
        }
        trie_node = root_node->val_at(2);
        return true;
      } else {
        throw std::runtime_error("unknown type, key path not found");
      }
  }

  } catch (const std::exception &e) {
    LOG(INFO) << e.what();
    return false;
  }

  throw std::runtime_error("key path not found");
}

neb::util::bytes trie::key_to_route(const neb::util::bytes &key) {

  size_t size = key.size() << 1;
  neb::util::bytes value(size);

  if (size > 0) {
    for (size_t i = 0; i < key.size(); i++) {
      byte_shared byte(key[i]);
      value[i << 1] = byte.bits_high();
      value[(i << 1) + 1] = byte.bits_low();
    }
  }
  return value;
}

neb::util::bytes trie::route_to_key(const neb::util::bytes &route) {

  size_t size = route.size() >> 1;
  neb::util::bytes value(size);

  if (size > 0) {
    for (size_t i = 0; i < size; i++) {
      byte_shared byte(route[(i << 1) + 1], route[i << 1]);
      value[i] = byte.data();
    }
  }
  return value;
}

size_t trie::prefix_len(const neb::byte_t *s, size_t s_len,
                        const neb::byte_t *t, size_t t_len) {
  size_t min_len = std::min(s_len, t_len);
  for (size_t i = 0; i < min_len; i++) {
    if (s[i] != t[i]) {
      return i;
    }
  }
  return min_len;
}
} // namespace fs
} // namespace neb
