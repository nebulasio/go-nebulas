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
#include "fs/storage.h"
#include <exception>

namespace neb {
namespace fs {

trie_node::trie_node(trie_node_type type) { change_to_type(type); }
trie_node::trie_node(const neb::bytes &triepb_bytes) {

  std::unique_ptr<triepb::Node> triepb_node_ptr =
      std::make_unique<triepb::Node>();

  bool ret = triepb_node_ptr->ParseFromArray(triepb_bytes.value(),
                                             triepb_bytes.size());
  if (!ret) {
    throw std::runtime_error("parse triepb node failed");
  }

  for (const std::string &v : triepb_node_ptr->val()) {
    m_val.push_back(neb::string_to_byte(v));
  }
}

trie_node::trie_node(const std::vector<neb::bytes> &val) : m_val(val) {}

trie_node_type trie_node::get_trie_node_type() {

  if (m_val.size() == 16) {
    return trie_node_branch;
  }
  if (m_val.size() == 3 && !m_val[0].empty()) {
    return static_cast<trie_node_type>(m_val[0][0]);
  }
  return trie_node_unknown;
}

void trie_node::change_to_type(trie_node_type new_type) {
  if (new_type == get_trie_node_type())
    return;
  if (new_type == trie_node_extension &&
      get_trie_node_type() == trie_node_leaf) {
    m_val.clear();
    return;
  }
  if (new_type == trie_node_leaf &&
      get_trie_node_type() == trie_node_extension) {
    m_val.clear();
    return;
  }

  if (new_type == trie_node_extension || new_type == trie_node_leaf) {
    m_val.resize(3);
    m_val.shrink_to_fit();
    return;
  }
  if (new_type == trie_node_leaf) {
    m_val.resize(16);
    m_val.shrink_to_fit();
  }
}

std::unique_ptr<triepb::Node> trie_node::to_proto() const {
  auto ret = std::make_unique<triepb::Node>();
  for (const neb::bytes &v : m_val) {
    std::string *val = ret->add_val();
    *val = std::to_string(v);
  }
  return ret;
}

trie::trie(storage *db, const hash_t &hash)
    : m_root_hash(hash), m_storage(db) {}

trie::trie(storage *db) : m_storage(db) {}

trie_node_ptr trie::create_node(const std::vector<neb::bytes> &val) {
  auto ret = std::make_unique<trie_node>(val);
  commit_node(ret.get());
  return ret;
}

void trie::commit_node(trie_node *node) {
  if (node == nullptr) {
    return;
  }
  auto pb = node->to_proto();
  size_t s = pb->ByteSizeLong();
  neb::bytes bs(s);
  pb->SerializeToArray(bs.value(), s);

  node->hash() = crypto::sha3_256_hash(bs);

  m_storage->put_bytes(bytes(node->hash().value(), node->hash().size()), bs);
}

hash_t trie::put(const hash_t &key, const neb::bytes &val) {
  auto new_hash = update(m_root_hash, key_to_route(key), val);
  m_root_hash = new_hash;
  return new_hash;
}

hash_t trie::update(const hash_t &root, const neb::bytes &route,
                    const neb::bytes &val) {
  if (root.empty()) {
    std::vector<neb::bytes> value(
        {number_to_byte<bytes>(trie_node_leaf), route, val});
    auto node = create_node(value);
    return node->hash();
  }

  auto root_node = fetch_node(root);
  auto type = root_node->get_trie_node_type();
  switch (type) {
  case trie_node_extension:
    return update_when_meet_ext(root_node.get(), route, val);
    break;
  case trie_node_leaf:
    return update_when_meet_leaf(root_node.get(), route, val);
    break;
  case trie_node_branch:
    return update_when_meet_branch(root_node.get(), route, val);
    break;
  default:
    throw std::invalid_argument("unknown node type");
  }
}
hash_t trie::update_when_meet_branch(trie_node *root_node,
                                     const neb::bytes &route,
                                     const neb::bytes &val) {
  auto new_hash = update(to_fix_bytes<hash_t>(root_node->val_at(route[0])),
                         neb::bytes(route.value() + 1, route.size() - 1), val);

  root_node->val_at(route[0]) = from_fix_bytes(new_hash);
  commit_node(root_node);
  return root_node->hash();
}

hash_t trie::update_when_meet_ext(trie_node *root_node, const neb::bytes &route,
                                  const neb::bytes &val) {
  bytes path = root_node->val_at(1);
  bytes next = root_node->val_at(2);
  if (path.size() > route.size()) {
    throw std::invalid_argument("wrong key, too short");
  }
  auto match_len = prefix_len(path, route);
  if (match_len == path.size()) {
    auto new_hash =
        update(to_fix_bytes<hash_t>(next),
               bytes(route.value() + match_len, route.size() - match_len), val);
    root_node->val_at(2) = from_fix_bytes(new_hash);
    commit_node(root_node);
    return root_node->hash();
  }
  std::unique_ptr<trie_node> br_node =
      std::make_unique<trie_node>(trie_node_branch);

  if (match_len > 0 || path.size() == 1) {
    br_node->val_at(path[match_len]) = next;
    if (match_len > 0 && match_len + 1 < path.size()) {
      std::vector<neb::bytes> value(
          {number_to_byte<bytes>(trie_node_extension),
           bytes(path.value() + match_len + 1, path.size() - match_len - 1),
           next});
      auto ext_node = create_node(value);
      br_node->val_at(path[match_len]) = from_fix_bytes(ext_node->hash());
    }

    // a branch to hold the new node
    br_node->val_at(route[match_len]) = from_fix_bytes(update(
        hash_t(),
        bytes(route.value() + match_len + 1, route.size() - match_len - 1),
        val));

    commit_node(br_node.get());

    // if no common prefix, replace the ext node with the new branch node
    if (match_len == 0) {
      return br_node->hash();
    }

    // use the new branch node as the ext node's sub-trie
    root_node->val_at(1) = bytes(path.value(), match_len);
    root_node->val_at(2) =
        bytes(br_node->hash().value(), br_node->hash().size());
    commit_node(root_node);
    return root_node->hash();
  }

  // 4. matchLen = 0 && len(path) > 1, 12... meets 23... => branch - ext - ...
  root_node->val_at(1) = bytes(path.value() + 1, path.size() - 1);
  commit_node(root_node);

  br_node->val_at(path[match_len]) = from_fix_bytes(root_node->hash());
  br_node->val_at(route[match_len]) = from_fix_bytes(update(
      hash_t(),
      bytes(route.value() + match_len + 1, route.size() - match_len - 1), val));
  commit_node(br_node.get());
  return br_node->hash();
}

hash_t trie::update_when_meet_leaf(trie_node *root_node,
                                   const neb::bytes &route,
                                   const neb::bytes &val) {
  neb::bytes path = root_node->val_at(1);
  neb::bytes leaf_val = root_node->val_at(2);
  if (path.size() > route.size()) {
    throw std::invalid_argument("wrong key, too short");
  }
  auto match_len = prefix_len(path, route);

  // node exists, update its value
  if (match_len == path.size()) {
    if (route.size() > match_len) {
      throw std::invalid_argument("wrong key, too long");
    }
    root_node->val_at(2) = val;
    commit_node(root_node);
    return root_node->hash();
  }

  auto br_node = std::make_unique<trie_node>(trie_node_branch);
  br_node->val_at(path[match_len]) = from_fix_bytes(
      update(hash_t(),
             bytes(path.value() + match_len + 1, path.size() - match_len - 1),
             leaf_val));
  br_node->val_at(route[match_len]) = from_fix_bytes(update(
      hash_t(),
      bytes(route.value() + match_len + 1, route.size() - match_len - 1), val));

  commit_node(br_node.get());

  // if no common prefix, replace the leaf node with the new branch node
  if (match_len == 0) {
    return br_node->hash();
  }

  root_node->change_to_type(trie_node_extension);
  root_node->val_at(0) = number_to_byte<bytes>(trie_node_extension);
  root_node->val_at(1) = bytes(path.value(), match_len);
  root_node->val_at(2) = from_fix_bytes(br_node->hash());
  commit_node(root_node);
  return root_node->hash();
}

std::unique_ptr<trie_node> trie::fetch_node(const hash_t &hash) {
  neb::bytes triepb_bytes = m_storage->get_bytes(from_fix_bytes(hash));
  return std::make_unique<trie_node>(triepb_bytes);
}

std::unique_ptr<trie_node> trie::fetch_node(const neb::bytes &hash) {
  neb::bytes triepb_bytes = m_storage->get_bytes(hash);
  return std::make_unique<trie_node>(triepb_bytes);
}

bool trie::get_trie_node(const neb::bytes &root_hash, const neb::bytes &key,
                         neb::bytes &trie_node) {
  auto hash = root_hash;
  auto route = key_to_route(key);

  neb::byte_t *route_ptr = route.value();
  size_t route_size = route.size();
  neb::byte_t *end_ptr = route.value() + route_size;

  try {

    while (route_ptr <= end_ptr) {

      auto root_node = fetch_node(hash);
      auto root_type = root_node->get_trie_node_type();
      if (route_ptr == end_ptr && root_type != trie_node_leaf) {
        throw std::runtime_error("key/path too short");
      }

      if (root_type == trie_node_branch) {
        hash = root_node->val_at(route_ptr[0]);
        route_ptr++;

      } else if (root_type == trie_node_extension) {
        auto key_path = root_node->val_at(1);
        auto next_hash = root_node->val_at(2);
        size_t left_size = end_ptr - route_ptr;

        size_t matched_len = prefix_len(key_path, route_ptr, left_size);
        if (matched_len != key_path.size()) {
          throw std::runtime_error("node extension, key path not found");
        }
        hash = next_hash;
        route_ptr += matched_len;
      } else if (root_type == trie_node_leaf) {
        auto key_path = root_node->val_at(1);
        size_t left_size = end_ptr - route_ptr;
        size_t matched_len = prefix_len(key_path, route_ptr, left_size);
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
    // LOG(INFO) << e.what();
    return false;
  }

  throw std::runtime_error("key path not found");
}

neb::bytes trie::route_to_key(const neb::bytes &route) {

  size_t size = route.size() >> 1;
  neb::bytes value(size);

  if (size > 0) {
    for (size_t i = 0; i < size; i++) {
      byte_shared byte(route[(i << 1) + 1], route[i << 1]);
      value[i] = byte.data();
    }
  }
  return value;
}

} // namespace fs
} // namespace neb
