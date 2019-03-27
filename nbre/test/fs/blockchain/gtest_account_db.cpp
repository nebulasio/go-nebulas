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

#include "common/configuration.h"
#include "fs/bc_storage_session.h"
#include "fs/blockchain.h"
#include "fs/blockchain/trie/trie.h"
#include <gtest/gtest.h>

TEST(test_fs, account_state) {

  std::string neb_db_path = neb::configuration::instance().neb_db_dir();
  neb::fs::bc_storage_session::instance().init(
      neb_db_path, neb::fs::storage_open_for_readonly);
  neb::fs::trie t;

  std::string root_hash_str =
      "6449f1c226e1e4d94837e1e813150b2500d0556ca67147e48c83126216362345";
  std::string addr_base58 = "n1HrPpwwH5gTA2d7QCkVjMw14YbN1NNNXHc";

  neb::bytes root_hash_bytes = neb::bytes::from_hex(root_hash_str);
  neb::bytes addr_bytes = neb::bytes::from_base58(addr_base58);

  // get block
  neb::bytes block_bytes =
      neb::fs::bc_storage_session::instance().get_bytes(root_hash_bytes);
  std::shared_ptr<corepb::Block> block = std::make_shared<corepb::Block>();
  bool ret = block->ParseFromArray(block_bytes.value(), block_bytes.size());
  if (!ret) {
    throw std::runtime_error("parse block failed");
  }
  // get block header account state
  std::string state_root_str = block->header().state_root();
  neb::bytes state_root_bytes = neb::string_to_byte(state_root_str);

  // get trie node
  neb::bytes trie_node_bytes;
  t.get_trie_node(state_root_bytes, addr_bytes, trie_node_bytes);

  std::shared_ptr<corepb::Account> corepb_account_ptr =
      std::make_shared<corepb::Account>();
  ret = corepb_account_ptr->ParseFromArray(trie_node_bytes.value(),
                                           trie_node_bytes.size());
  if (!ret) {
    throw std::runtime_error("parse corepb Account failed");
  }
  std::string addr_str = corepb_account_ptr->address();
  LOG(INFO) << neb::string_to_byte(addr_str).to_base58();
  std::string balance_str = corepb_account_ptr->balance();
  std::string hex_str = neb::string_to_byte(balance_str).to_hex();
  LOG(INFO) << hex_str;

  std::stringstream ss;
  ss << std::hex << hex_str;
  neb::wei_t tmp;
  ss >> tmp;
  LOG(INFO) << tmp;
}

TEST(test_fs, event) {}
