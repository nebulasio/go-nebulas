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
#include "util/bc_generator.h"
#include "crypto/hash.h"
#include "fs/bc_storage_session.h"
#include "fs/proto/trie.pb.h"
#include "util/chrono.h"
#include "util/json_parser.h"
#include <boost/property_tree/json_parser.hpp>
#include <boost/property_tree/ptree.hpp>

namespace neb {
namespace util {
void all_accounts::add_account(
    const std::shared_ptr<corepb::Account> &account) {
  address_t addr = get_address_from_account(account.get());
  m_all_accounts.insert(std::make_pair(addr, std::move(account)));
  m_all_addresses.push_back(addr);
}

corepb::Account *all_accounts::random_account() const {
  if (m_all_accounts.empty())
    throw std::invalid_argument("no account yet");
  std::uniform_int_distribution<uint64_t> dist(0, m_all_accounts.size() - 1);
  const address_t &addr = m_all_addresses[dist(m_random_generator)];
  auto it = m_all_accounts.find(addr);
  return it->second.get();
}

uint64_t all_accounts::get_nonce(const address_t &addr) {
  auto it = m_all_accounts.find(addr);
  if (it == m_all_accounts.end())
    return 0;
  return it->second->nonce();
}

void all_accounts::increase_nonce(const address_t &addr) {
  auto it = m_all_accounts.find(addr);
  if (it == m_all_accounts.end())
    return;

  auto account = it->second;
  account->set_nonce(account->nonce());
}

void all_accounts::increase_balance(const address_t &addr, const wei &val) {
  auto it = m_all_accounts.find(addr);
  if (it == m_all_accounts.end())
    return;

  auto account = it->second;
  wei_t b = storage_to_wei(string_to_byte(account->balance()));
  b = b + val.wei_value();
  account->set_balance(byte_to_string(wei_to_storage(b)));
}

bool all_accounts::decrease_balance(const address_t &addr, const wei &val) {
  auto it = m_all_accounts.find(addr);
  if (it == m_all_accounts.end()) {
    return false;
  }

  auto account = it->second;
  wei_t b = storage_to_wei(string_to_byte(account->balance()));
  b = b - val.wei_value();
  if (b < 0) {
    return false;
  }
  account->set_balance(byte_to_string(wei_to_storage(b)));
  return true;
}

corepb::Account *all_accounts::random_user_account() const {
  corepb::Account *account = random_account();
  address_t addr = get_address_from_account(account);
  while (!is_normal_address(addr)) {
    account = random_account();
    addr = get_address_from_account(account);
  }
  return account;
}

corepb::Account *all_accounts::random_contract_account() const {
  corepb::Account *account = random_account();
  address_t addr = get_address_from_account(account);
  uint32_t retry_max_num = m_all_accounts.size();
  while (!is_contract_address(addr)) {
    account = random_account();
    addr = get_address_from_account(account);
    if (--retry_max_num < 0) {
      return nullptr;
    }
  }
  return account;
}

address_t get_address_from_account(corepb::Account *account) {
  return to_address(account->address());
}

generate_block::generate_block(all_accounts *accounts, uint64_t height)
    : m_all_accounts(accounts), m_height(height) {}

std::shared_ptr<corepb::Account>
generate_block::gen_user_account(const nas &v) {
  std::shared_ptr<corepb::Account> ret = std::make_shared<corepb::Account>();
  address_t addr(NAS_ADDRESS_LEN);
  addr[0] = NAS_ADDRESS_MAGIC_NUM;
  addr[1] = NAS_ADDRESS_ACCOUNT_MAGIC_NUM;
  for (int i = 2; i < NAS_ADDRESS_LEN;) {
    int32_t num = std::rand();
    int32_t *p = (int32_t *)(addr.value() + i);
    *p = num;
    i += 4;
  }
  ret->set_address(address_to_string(addr));
  std::string balance = byte_to_string(wei_to_storage(v.wei_value()));
  ret->set_balance(balance);
  m_all_accounts->add_account(ret);
  return ret;
}

std::shared_ptr<corepb::Account>
generate_block::add_deploy_transaction(const address_t &owner,
                                       const bytes &payload) {
  transaction_ptr tx(new corepb::Transaction());
  tx->set_to(address_to_string(owner));
  tx->set_from(address_to_string(owner));
  corepb::Data *data = new corepb::Data();
  data->set_type("deploy");
  boost::property_tree::ptree pt;
  pt.put("Data", payload.to_base64());
  std::string payload_str;
  neb::util::json_parser::write_json(payload_str, pt);
  data->set_payload(payload_str);
  tx->set_allocated_data(data);
  tx->set_timestamp(util::now());
  tx->set_nonce(m_all_accounts->get_nonce(owner));
  wei value(1000_wei);
  bool b_ret = m_all_accounts->decrease_balance(owner, value);
  if (b_ret == false)
    return nullptr;
  tx->set_value(byte_to_string(wei_to_storage(value.wei_value())));
  m_all_accounts->increase_nonce(owner);
  neb::bytes b(tx->ByteSizeLong());
  tx->SerializeToArray(b.value(), b.size());
  tx->set_hash(byte_to_string(from_fix_bytes(crypto::sha3_256_hash(b))));
  m_transactions.push_back(tx);

  std::shared_ptr<corepb::Account> ret = std::make_shared<corepb::Account>();
  address_t addr(NAS_ADDRESS_LEN);
  addr[0] = NAS_ADDRESS_MAGIC_NUM;
  addr[1] = NAS_ADDRESS_CONTRACT_MAGIC_NUM;
  for (int i = 2; i < NAS_ADDRESS_LEN;) {
    int32_t num = std::rand();
    int32_t *p = (int32_t *)(addr.value() + i);
    *p = num;
    i += 4;
  }
  ret->set_address(address_to_string(addr));
  ret->set_birth_place(tx->hash());
  m_all_accounts->add_account(ret);

  return ret;
}
std::shared_ptr<corepb::Transaction>
generate_block::add_protocol_transaction(const address_t &owner,
                                         const bytes &payload) {
  transaction_ptr tx(new corepb::Transaction());
  tx->set_to(address_to_string(owner));
  tx->set_from(address_to_string(owner));
  corepb::Data *data = new corepb::Data();
  data->set_type("protocol");
  boost::property_tree::ptree pt;
  pt.put("Data", payload.to_base64());

  std::string payload_str;
  neb::util::json_parser::write_json(payload_str, pt);
  data->set_payload(payload_str);
  wei value(1000_wei);
  bool ret = m_all_accounts->decrease_balance(owner, value);
  if (ret == false)
    return nullptr;
  tx->set_value(byte_to_string(wei_to_storage(value.wei_value())));
  tx->set_allocated_data(data);
  tx->set_timestamp(util::now());
  tx->set_nonce(m_all_accounts->get_nonce(owner));
  m_all_accounts->increase_nonce(owner);
  neb::bytes b(tx->ByteSizeLong());
  tx->SerializeToArray(b.value(), b.size());
  tx->set_hash(byte_to_string(from_fix_bytes(crypto::sha3_256_hash(b))));
  m_transactions.push_back(tx);

  return tx;
}
std::shared_ptr<corepb::Transaction>
generate_block::add_binary_transaction(const address_t &from,
                                       const address_t &to, const nas &value) {
  transaction_ptr tx(new corepb::Transaction());
  tx->set_to(address_to_string(to));
  tx->set_from(address_to_string(from));
  corepb::Data *data = new corepb::Data();
  data->set_type("binary");
  bool ret = m_all_accounts->decrease_balance(from, wei(value));
  if (ret == false)
    return nullptr;
  m_all_accounts->increase_balance(to, wei(value));
  std::string v = byte_to_string(wei_to_storage(value.wei_value()));
  tx->set_value(v);
  tx->set_allocated_data(data);
  tx->set_timestamp(util::now());
  tx->set_nonce(m_all_accounts->get_nonce(from));
  m_all_accounts->increase_nonce(from);
  neb::bytes b(tx->ByteSizeLong());
  tx->SerializeToArray(b.value(), b.size());
  tx->set_hash(byte_to_string(from_fix_bytes(crypto::sha3_256_hash(b))));
  m_transactions.push_back(tx);

  return tx;
}

std::shared_ptr<corepb::Transaction>
generate_block::add_call_transaction(const address_t &from,
                                     const address_t &to) {
  transaction_ptr tx(new corepb::Transaction());
  tx->set_to(address_to_string(to));
  tx->set_from(address_to_string(from));
  corepb::Data *data = new corepb::Data();
  data->set_type("call");
  tx->set_allocated_data(data);
  tx->set_timestamp(util::now());
  wei value(1000_wei);
  std::string v = byte_to_string(wei_to_storage(value.wei_value()));
  tx->set_value(v);
  tx->set_nonce(m_all_accounts->get_nonce(from));
  m_all_accounts->increase_nonce(from);
  neb::bytes b(tx->ByteSizeLong());
  tx->SerializeToArray(b.value(), b.size());
  tx->set_hash(byte_to_string(from_fix_bytes(crypto::sha3_256_hash(b))));
  m_transactions.push_back(tx);

  return tx;
}

void generate_block::write_to_blockchain_db() {
  // 1. generate corepb::Block
  std::unique_ptr<corepb::BlockHeader> header =
      std::make_unique<corepb::BlockHeader>();
  header->set_timestamp(util::now());

  std::unique_ptr<corepb::Block> block = std::make_unique<corepb::Block>();
  block->set_allocated_header(header.get());
  block->set_height(m_height);
  for (auto &tx : m_transactions) {
    corepb::Transaction *etx = block->add_transactions();
    *etx = *tx;
  }
  std::string block_str = block->SerializeAsString();
  header->set_hash(
      byte_to_string(from_fix_bytes(crypto::sha3_256_hash(block_str))));

  // 2. update to LIB
  fs::blockchain::write_LIB_block(block.get());
  block->release_header();

  // 3. write all accounts to DB
  //! We use triepb::Node to write all accounts to db.
  //! This is a trick!
  //! Ideally, we should use trie to shrink db size.
  //! Yet, we only fill several fields in Account, so it's ok to write dup data.
  triepb::Node t_accounts;
  m_all_accounts->for_each_account(
      [&t_accounts](const std::shared_ptr<corepb::Account> &account) {
        std::string *s = t_accounts.add_val();
        *s = account->SerializeAsString();
      });

  std::string key = std::string("account") + std::to_string(m_height);
  std::string account_str = t_accounts.SerializeAsString();
  fs::bc_storage_session::instance().put_bytes(string_to_byte(key),
                                               string_to_byte(account_str));
}

std::vector<std::shared_ptr<corepb::Account>>
generate_block::read_accounts_in_height(block_height_t height) {
  std::string key = std::string("account") + std::to_string(height);
  auto account_str =
      fs::bc_storage_session::instance().get_bytes(string_to_byte(key));
  triepb::Node t_accounts;
  t_accounts.ParseFromArray(account_str.value(), account_str.size());
  std::vector<std::shared_ptr<corepb::Account>> ret;
  for (size_t i = 0; i < t_accounts.val_size(); ++i) {
    std::string s = t_accounts.val(i);
    std::shared_ptr<corepb::Account> account =
        std::make_shared<corepb::Account>();
    account->ParseFromString(s);
    ret.push_back(account);
  }
  return ret;
}

std::shared_ptr<corepb::Block>
generate_block::read_block_with_height(block_height_t height) {
  std::unique_ptr<corepb::Block> block =
      fs::blockchain::load_block_with_height(height);
  return std::shared_ptr<corepb::Block>(std::move(block));
}

} // namespace util
} // namespace neb
