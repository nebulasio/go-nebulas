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
#include "common/util/byte.h"
#include "fs/proto/trie.pb.h"
#include "fs/rocksdb_storage.h"

namespace neb {
namespace fs {

enum class trie_node_type {
  trie_node_unknown = 0,
  trie_node_extension,
  trie_node_leaf,
  trie_node_branch,
};

class trie_node {
public:
  trie_node(const neb::util::bytes &triepb_bytes);

  trie_node_type get_trie_node_type();

  inline neb::util::bytes val_at(size_t index) { return m_val[index]; }

private:
  neb::util::bytes m_hash;
  neb::util::bytes m_bytes;
  std::vector<neb::util::bytes> m_val;
};

class trie {
public:
  trie(const std::string &neb_db_path);
  ~trie();

  neb::util::bytes get_trie_node(const neb::util::bytes &root_hash,
                                 const neb::util::bytes &key);

private:
  std::shared_ptr<trie_node> fetch_node(const neb::util::bytes &hash);

public:
  static neb::util::bytes key_to_route(const neb::util::bytes &key);
  static neb::util::bytes route_to_key(const neb::util::bytes &route);
  static size_t prefix_len(const neb::util::bytes &s,
                           const neb::util::bytes &t);

private:
  std::unique_ptr<rocksdb_storage> m_storage;
};
} // namespace fs
} // namespace neb
