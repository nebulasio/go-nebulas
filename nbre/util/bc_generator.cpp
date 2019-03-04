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
#include "util/chrono.h"

namespace neb {
namespace util {
void all_accounts::add_account(
    const std::shared_ptr<corepb::Account> &account) {
  address_t addr = get_address_from_account(account.get());
  m_all_accounts.insert(std::make_pair(addr, std::move(account)));
  m_all_addresses.push_back(addr);
}

corepb::Account *all_accounts::random_account() const {
  std::uniform_int_distribution<uint64_t> dist(0, m_all_accounts.size());
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
  while (!is_contract_address(addr)) {
    account = random_account();
    addr = get_address_from_account(account);
  }
  return account;
}

address_t get_address_from_account(corepb::Account *account) {
  return to_address(account->address());
}

generate_block::generate_block(all_accounts *accounts, uint64_t height)
    : m_all_accounts(accounts), m_height(height) {}

std::shared_ptr<corepb::Account>
generate_block::gen_user_Account(const nas &v) {
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
  std::string balance = util::byte_to_string(wei_to_storage(v.wei_value()));
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
  data->set_payload(util::byte_to_string(payload));
  tx->set_allocated_data(data);
  tx->set_timestamp(util::now());
  tx->set_nonce(m_all_accounts->get_nonce(owner));
  m_all_accounts->increase_nonce(owner);
  neb::util::bytes b(tx->ByteSizeLong());
  tx->SerializeToArray(b.value(), b.size());
  tx->set_hash(util::byte_to_string(from_fix_bytes(crypto::sha3_256_hash(b))));
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
  data->set_payload(util::byte_to_string(payload));
  tx->set_allocated_data(data);
  tx->set_timestamp(util::now());
  tx->set_nonce(m_all_accounts->get_nonce(owner));
  m_all_accounts->increase_nonce(owner);
  neb::util::bytes b(tx->ByteSizeLong());
  tx->SerializeToArray(b.value(), b.size());
  tx->set_hash(util::byte_to_string(from_fix_bytes(crypto::sha3_256_hash(b))));
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
  std::string v = util::byte_to_string(wei_to_storage(value.wei_value()));
  tx->set_value(v);
  tx->set_allocated_data(data);
  tx->set_timestamp(util::now());
  tx->set_nonce(m_all_accounts->get_nonce(from));
  m_all_accounts->increase_nonce(from);
  neb::util::bytes b(tx->ByteSizeLong());
  tx->SerializeToArray(b.value(), b.size());
  tx->set_hash(util::byte_to_string(from_fix_bytes(crypto::sha3_256_hash(b))));
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
  tx->set_nonce(m_all_accounts->get_nonce(from));
  m_all_accounts->increase_nonce(from);
  neb::util::bytes b(tx->ByteSizeLong());
  tx->SerializeToArray(b.value(), b.size());
  tx->set_hash(util::byte_to_string(from_fix_bytes(crypto::sha3_256_hash(b))));
  m_transactions.push_back(tx);

  return tx;
}
} // namespace util
} // namespace neb
