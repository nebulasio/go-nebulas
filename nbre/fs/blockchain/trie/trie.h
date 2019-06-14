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
#include "crypto/hash.h"
#include "fs/blockchain/trie/byte_shared.h"
#include "fs/proto/trie.pb.h"

namespace neb {
namespace fs {

typedef int32_t trie_node_type;
constexpr static int32_t trie_node_unknown = 0;
constexpr static int32_t trie_node_extension = 1;
constexpr static int32_t trie_node_leaf = 2;
constexpr static int32_t trie_node_branch = 3;
class storage;
class trie_node {
public:
  trie_node(trie_node_type type);
  trie_node(const neb::bytes &triepb_bytes);

  trie_node(const std::vector<neb::bytes> &val);

  trie_node_type get_trie_node_type();

  inline const neb::bytes &val_at(size_t index) const { return m_val[index]; }
  inline neb::bytes &val_at(size_t index) { return m_val[index]; }

  void change_to_type(trie_node_type new_type);

  inline const hash_t &hash() const { return m_hash; }

  inline hash_t &hash() { return m_hash; }

  std::unique_ptr<triepb::Node> to_proto() const;

private:
  std::vector<neb::bytes> m_val;
  hash_t m_hash;
};

typedef std::unique_ptr<trie_node> trie_node_ptr;

class trie {
public:
  trie(storage *db, const hash_t &hash);
  trie(storage *db);

  bool get_trie_node(const neb::bytes &root_hash, const neb::bytes &key,
                     neb::bytes &trie_node);

  hash_t put(const hash_t &key, const neb::bytes &val);

  trie_node_ptr create_node(const std::vector<neb::bytes> &val);

  void commit_node(trie_node *node);

  inline const hash_t &root_hash() const { return m_root_hash; }
  inline hash_t &root_hash() { return m_root_hash; }
  inline bool empty() const { return m_root_hash.size() == 0; }

private:
  std::unique_ptr<trie_node> fetch_node(const neb::bytes &hash);
  std::unique_ptr<trie_node> fetch_node(const hash_t &hash);

  hash_t update(const hash_t &root, const neb::bytes &route,
                const neb::bytes &val);

  hash_t update_when_meet_branch(trie_node *root_node, const neb::bytes &route,
                                 const neb::bytes &val);
  hash_t update_when_meet_ext(trie_node *root_node, const neb::bytes &route,
                              const neb::bytes &val);
  hash_t update_when_meet_leaf(trie_node *root_node, const neb::bytes &route,
                               const neb::bytes &val);

public:
  static neb::bytes route_to_key(const neb::bytes &route);
  template <typename T> static neb::bytes key_to_route(const T &key) {

    size_t size = key.size() << 1;
    neb::bytes value(size);

    if (size > 0) {
      for (size_t i = 0; i < key.size(); i++) {
        byte_shared byte(key[i]);
        value[i << 1] = byte.bits_high();
        value[(i << 1) + 1] = byte.bits_low();
      }
    }
    return value;
  }
  template <typename T1, typename T2>
  static size_t prefix_len(const T1 &s, const T2 &t) {
    auto s_len = s.size();
    auto t_len = t.size();
    size_t min_len = std::min(s_len, t_len);
    for (size_t i = 0; i < min_len; i++) {
      if (s[i] != t[i]) {
        return i;
      }
    }
    return min_len;
  }
  template <typename T1>
  static size_t prefix_len(const T1 &s, const byte_t *t, size_t t_len) {
    auto s_len = s.size();
    size_t min_len = std::min(s_len, t_len);
    for (size_t i = 0; i < min_len; i++) {
      if (s[i] != t[i]) {
        return i;
      }
    }
    return min_len;
  }

protected:
  hash_t m_root_hash;
  storage *m_storage;
};
} // namespace fs
} // namespace neb
