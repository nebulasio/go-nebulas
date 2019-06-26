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
#include "common/address.h"
#include "common/nebulas_currency.h"
#include "fs/blockchain.h"
#include "fs/blockchain/blockchain_api.h"
#include "fs/proto/block.pb.h"
#include <random>

namespace neb {
namespace util {

class all_accounts {
public:
  void add_account(const std::shared_ptr<corepb::Account> &account);

  corepb::Account *random_contract_account() const;
  corepb::Account *random_user_account() const;

  inline address_t random_contract_addr() const {
    auto p = random_contract_account();
    if (!p) {
      return address_t();
    }
    return to_address(p->address());
  }
  inline address_t random_user_addr() const {
    auto p = random_user_account();
    if (!p) {
      return address_t();
    }
    return to_address(p->address());
  }

  inline size_t size() const { return m_all_accounts.size(); }

  uint64_t get_nonce(const address_t &addr);
  void increase_nonce(const address_t &addr);

  template <typename Func> void for_each_account(Func &&f) {
    std::for_each(
        m_all_accounts.begin(), m_all_accounts.end(),
        [f](const std::pair<address_t, std::shared_ptr<corepb::Account>>
                &item) { f(item.second); });
  }

  void increase_balance(const address_t &addr, const wei &val);
  bool decrease_balance(const address_t &addr, const wei &val);

protected:
  corepb::Account *random_account() const;

protected:
  std::unordered_map<address_t, std::shared_ptr<corepb::Account>>
      m_all_accounts;
  std::vector<address_t> m_all_addresses;
  mutable std::default_random_engine m_random_generator;
};

address_t get_address_from_account(corepb::Account *account);

class generate_block {
public:
  generate_block(fs::blockchain *bc);
  generate_block(all_accounts *accounts, uint64_t height);

  std::shared_ptr<corepb::Account> gen_user_account(const nas &v = 10000_nas);

  std::shared_ptr<corepb::Account>
  add_deploy_transaction(const address_t &owner, const bytes &payload);

  std::shared_ptr<corepb::Transaction>
  add_protocol_transaction(const address_t &owner, const bytes &payload);

  std::shared_ptr<corepb::Transaction>
  add_binary_transaction(const address_t &from, const address_t &to,
                         const nas &value);

  std::shared_ptr<corepb::Transaction>
  add_call_transaction(const address_t &from, const address_t &to);

  void write_to_blockchain_db();

  inline uint64_t height() const { return m_height; }

  all_accounts *get_all_accounts() { return m_all_accounts; }

  const std::vector<std::shared_ptr<corepb::Transaction>> &
  all_transactions() const {
    return m_transactions;
  }

  static std::vector<std::shared_ptr<corepb::Account>>
  read_accounts_in_height(block_height_t height);
  static std::vector<std::shared_ptr<corepb::Account>> read_accounts_in_LIB();

  static std::shared_ptr<corepb::Block>
  read_block_with_height(block_height_t height);
  static std::shared_ptr<corepb::Block> read_LIB_block();

protected:
  all_accounts *m_all_accounts;
  uint64_t m_height;
  typedef std::shared_ptr<corepb::Transaction> transaction_ptr;
  std::vector<transaction_ptr> m_transactions;
  fs::blockchain *m_blockchain;
};
} // namespace util
} // namespace neb
