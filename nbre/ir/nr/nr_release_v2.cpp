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

#include "common/int128_conversion.h"
#include "common/nebulas_currency.h"
#include "fs/blockchain/trie/trie.h"
#include "runtime/nr/impl/nr_impl.h"
#include <boost/property_tree/json_parser.hpp>

namespace neb {
namespace fs {

class account_db_v2 {
public:
  account_db_v2(neb::fs::account_db *adb_ptr);

  neb::wei_t get_balance(const neb::address_t &addr,
                         neb::block_height_t height);
  neb::address_t get_contract_deployer(const neb::address_t &addr,
                                       neb::block_height_t height);

  void update_height_address_val_internal(
      const std::vector<neb::fs::transaction_info_t> &txs,
      std::unordered_map<neb::address_t, neb::wei_t> &addr_balance);

  neb::wei_t get_account_balance_internal(const neb::address_t &addr,
                                          neb::block_height_t height);

  static neb::floatxx_t get_normalized_value(neb::floatxx_t value);

private:
  void init_height_address_val_internal(
      neb::block_height_t start_block,
      const std::unordered_map<neb::address_t, neb::wei_t> &addr_balance);

private:
  std::unordered_map<neb::address_t, std::vector<neb::block_height_t>>
      m_addr_height_list;
  std::unordered_map<neb::block_height_t,
                     std::unordered_map<neb::address_t, neb::wei_t>>
      m_height_addr_val;
  neb::fs::account_db *m_adb_ptr;
};

account_db_v2::account_db_v2(neb::fs::account_db *adb_ptr)
    : m_adb_ptr(adb_ptr) {}

neb::wei_t account_db_v2::get_balance(const neb::address_t &addr,
                                      neb::block_height_t height) {
  return m_adb_ptr->get_balance(addr, height);
}

neb::address_t
account_db_v2::get_contract_deployer(const neb::address_t &addr,
                                     neb::block_height_t height) {
  return m_adb_ptr->get_contract_deployer(addr, height);
}

void account_db_v2::init_height_address_val_internal(
    neb::block_height_t start_block,
    const std::unordered_map<neb::address_t, neb::wei_t> &addr_balance) {

  for (auto &ele : addr_balance) {
    std::vector<neb::block_height_t> v{start_block};
    m_addr_height_list.insert(std::make_pair(ele.first, v));

    auto iter = m_height_addr_val.find(start_block);
    if (iter == m_height_addr_val.end()) {
      std::unordered_map<address_t, wei_t> addr_val = {{ele.first, ele.second}};
      m_height_addr_val.insert(std::make_pair(start_block, addr_val));
    } else {
      auto &addr_val = iter->second;
      addr_val.insert(std::make_pair(ele.first, ele.second));
    }
  }
}

void account_db_v2::update_height_address_val_internal(
    const std::vector<neb::fs::transaction_info_t> &txs,
    std::unordered_map<neb::address_t, neb::wei_t> &addr_balance) {

  if (txs.empty()) {
    return;
  }
  auto start_block = txs.front().m_height;
  init_height_address_val_internal(start_block, addr_balance);

  for (auto it = txs.begin(); it != txs.end(); it++) {
    address_t from = it->m_from;
    address_t to = it->m_to;

    block_height_t height = it->m_height;
    wei_t tx_value = it->m_tx_value;
    wei_t value = tx_value;

    if (addr_balance.find(from) == addr_balance.end()) {
      addr_balance.insert(std::make_pair(from, 0));
    }
    if (addr_balance.find(to) == addr_balance.end()) {
      addr_balance.insert(std::make_pair(to, 0));
    }

    int32_t status = it->m_status;
    if (status) {
      addr_balance[from] -= value;
      addr_balance[to] += value;
    }

    wei_t gas_used = it->m_gas_used;
    if (gas_used != 0) {
      wei_t gas_val = gas_used * it->m_gas_price;
      addr_balance[from] -= gas_val;
    }

    if (m_height_addr_val.find(height) == m_height_addr_val.end()) {
      std::unordered_map<address_t, wei_t> addr_val = {
          {from, addr_balance[from]}, {to, addr_balance[to]}};
      m_height_addr_val.insert(std::make_pair(height, addr_val));
    } else {
      std::unordered_map<address_t, wei_t> &addr_val =
          m_height_addr_val[height];
      if (addr_val.find(from) == addr_val.end()) {
        addr_val.insert(std::make_pair(from, addr_balance[from]));
      } else {
        addr_val[from] = addr_balance[from];
      }
      if (addr_val.find(to) == addr_val.end()) {
        addr_val.insert(std::make_pair(to, addr_balance[to]));
      } else {
        addr_val[to] = addr_balance[to];
      }
    }

    if (m_addr_height_list.find(from) == m_addr_height_list.end()) {
      std::vector<block_height_t> v{height};
      m_addr_height_list.insert(std::make_pair(from, v));
    } else {
      std::vector<block_height_t> &v = m_addr_height_list[from];
      // expect reading transactions order by height asc
      if (!v.empty() && v.back() < height) {
        v.push_back(height);
      }
    }

    if (m_addr_height_list.find(to) == m_addr_height_list.end()) {
      std::vector<block_height_t> v{height};
      m_addr_height_list.insert(std::make_pair(to, v));
    } else {
      std::vector<block_height_t> &v = m_addr_height_list[to];
      if (!v.empty() && v.back() < height) {
        v.push_back(height);
      }
    }
  }
}

neb::wei_t
account_db_v2::get_account_balance_internal(const neb::address_t &address,
                                            neb::block_height_t height) {

  auto addr_it = m_addr_height_list.find(address);
  if (addr_it == m_addr_height_list.end()) {
    return get_balance(address, height);
  }

  auto height_it =
      std::lower_bound(addr_it->second.begin(), addr_it->second.end(), height);

  if (height_it == addr_it->second.end()) {
    height_it--;
    return m_height_addr_val[*height_it][address];
  }

  if (height_it == addr_it->second.begin()) {
    if (*height_it == height) {
      return m_height_addr_val[*height_it][address];
    } else {
      return get_balance(address, height);
    }
  }

  if (*height_it != height) {
    height_it--;
  }

  return m_height_addr_val[*height_it][address];
}

neb::floatxx_t account_db_v2::get_normalized_value(neb::floatxx_t value) {
  uint64_t ratio = 1000000000000000000ULL;
  return value / neb::floatxx_t(ratio);
}

class transaction_db_v2 {
public:
  transaction_db_v2(transaction_db *tdb_ptr);

  std::unique_ptr<std::vector<transaction_info_t>>
  read_transactions_from_db_with_duration(block_height_t start_block,
                                          block_height_t end_block);

  static std::unique_ptr<std::vector<transaction_info_t>>
  read_transactions_with_succ(const std::vector<transaction_info_t> &txs);

  static std::unique_ptr<std::vector<transaction_info_t>>
  read_transactions_with_address_type(
      const std::vector<transaction_info_t> &txs, byte_t from_type,
      byte_t to_type);

protected:
  transaction_db *m_tdb_ptr;
};

transaction_db_v2::transaction_db_v2(transaction_db *tdb_ptr)
    : m_tdb_ptr(tdb_ptr) {}

std::unique_ptr<std::vector<transaction_info_t>>
transaction_db_v2::read_transactions_from_db_with_duration(
    block_height_t start_block, block_height_t end_block) {
  return m_tdb_ptr->read_transactions_from_db_with_duration(start_block,
                                                            end_block);
}

std::unique_ptr<std::vector<transaction_info_t>>
transaction_db_v2::read_transactions_with_succ(
    const std::vector<transaction_info_t> &txs) {
  auto ptr = std::make_unique<std::vector<transaction_info_t>>();
  for (auto &tx : txs) {
    if (tx.m_status) {
      ptr->push_back(tx);
    }
  }
  return ptr;
}

std::unique_ptr<std::vector<transaction_info_t>>
transaction_db_v2::read_transactions_with_address_type(
    const std::vector<transaction_info_t> &txs, byte_t from_type,
    byte_t to_type) {
  return transaction_db::read_transactions_with_address_type(txs, from_type,
                                                             to_type);
}

class blockchain_api_v2 : public blockchain_api {
public:
  blockchain_api_v2();
  virtual ~blockchain_api_v2();

  virtual std::unique_ptr<std::vector<transaction_info_t>>
  get_block_transactions_api(block_height_t height);

protected:
  virtual std::unique_ptr<event_info_t>
  get_transaction_result_api(const neb::bytes &events_root,
                             const neb::bytes &tx_hash);
  std::unique_ptr<event_info_t> json_parse_event(const std::string &json);
};

blockchain_api_v2::blockchain_api_v2() : blockchain_api() {}
blockchain_api_v2::~blockchain_api_v2() {}

std::unique_ptr<std::vector<transaction_info_t>>
blockchain_api_v2::get_block_transactions_api(block_height_t height) {

  auto ret = std::make_unique<std::vector<transaction_info_t>>();
  // special for  block height 1
  if (height <= 1) {
    return ret;
  }

  auto block = blockchain::load_block_with_height(height);
  int64_t timestamp = block->header().timestamp();

  std::string events_root_str = block->header().events_root();
  neb::bytes events_root_bytes = neb::string_to_byte(events_root_str);

  for (auto &tx : block->transactions()) {
    transaction_info_t info;

    info.m_height = height;
    info.m_timestamp = timestamp;

    info.m_from = to_address(tx.from());
    info.m_to = to_address(tx.to());
    info.m_tx_value = storage_to_wei(neb::string_to_byte(tx.value()));
    info.m_gas_price = storage_to_wei(neb::string_to_byte(tx.gas_price()));
    info.m_tx_type = tx.data().type();

    // get topic chain.transactionResult
    std::string tx_hash_str = tx.hash();
    neb::bytes tx_hash_bytes = neb::string_to_byte(tx_hash_str);
    auto txs_result_ptr =
        get_transaction_result_api(events_root_bytes, tx_hash_bytes);

    info.m_status = txs_result_ptr->m_status;
    info.m_gas_used = txs_result_ptr->m_gas_used;

    ret->push_back(info);
  }
  return ret;
}

std::unique_ptr<event_info_t>
blockchain_api_v2::get_transaction_result_api(const neb::bytes &events_root,
                                              const neb::bytes &tx_hash) {
  trie t;
  neb::bytes txs_result;

  for (int64_t id = 1;; id++) {
    neb::bytes id_bytes = neb::number_to_byte<neb::bytes>(id);
    neb::bytes events_tx_hash = tx_hash;
    events_tx_hash.append_bytes(id_bytes.value(), id_bytes.size());

    neb::bytes trie_node_bytes;
    bool ret = t.get_trie_node(events_root, events_tx_hash, trie_node_bytes);
    if (!ret) {
      break;
    }
    txs_result = trie_node_bytes;
  }
  assert(!txs_result.empty());

  std::string json_str = neb::byte_to_string(txs_result);

  return json_parse_event(json_str);
}

std::unique_ptr<event_info_t>
blockchain_api_v2::json_parse_event(const std::string &json) {
  boost::property_tree::ptree pt;
  std::stringstream ss(json);
  boost::property_tree::read_json(ss, pt);

  std::string topic = pt.get<std::string>("Topic");
  assert(topic.compare("chain.transactionResult") == 0);

  std::string data_json = pt.get<std::string>("Data");
  ss = std::stringstream(data_json);
  boost::property_tree::read_json(ss, pt);

  int32_t status = pt.get<int32_t>("status");
  wei_t gas_used = boost::lexical_cast<wei_t>(pt.get<std::string>("gas_used"));

  auto ret = std::make_unique<event_info_t>(event_info_t{status, gas_used});
  return ret;
}

} // namespace fs
} // namespace neb

#define BLOCK_INTERVAL_MAGIC_NUM 128

namespace neb {
namespace rt {
namespace nr {

using transaction_db_v2_ptr_t = std::unique_ptr<neb::fs::transaction_db_v2>;
using account_db_v2_ptr_t = std::unique_ptr<neb::fs::account_db_v2>;

class nebulas_rank_v2 {
public:
  static std::vector<std::shared_ptr<nr_info_t>>
  get_nr_score(const transaction_db_v2_ptr_t &tdb_ptr,
               const account_db_v2_ptr_t &adb_ptr, const rank_params_t &rp,
               neb::block_height_t start_block, neb::block_height_t end_block);

private:
  static auto split_transactions_by_block_interval(
      const std::vector<neb::fs::transaction_info_t> &txs,
      int32_t block_interval = BLOCK_INTERVAL_MAGIC_NUM)
      -> std::unique_ptr<std::vector<std::vector<neb::fs::transaction_info_t>>>;

  static void filter_empty_transactions_this_interval(
      std::vector<std::vector<neb::fs::transaction_info_t>> &txs);

  static auto build_transaction_graphs(
      const std::vector<std::vector<neb::fs::transaction_info_t>> &txs)
      -> std::unique_ptr<std::vector<transaction_graph_ptr_t>>;

  static auto
  get_normal_accounts(const std::vector<neb::fs::transaction_info_t> &txs)
      -> std::unique_ptr<std::unordered_set<address_t>>;

  static auto get_account_balance_median(
      const std::unordered_set<address_t> &accounts,
      const std::vector<std::vector<neb::fs::transaction_info_t>> &txs,
      const account_db_v2_ptr_t &db_ptr)
      -> std::unique_ptr<std::unordered_map<address_t, floatxx_t>>;

  static auto get_account_weight(
      const std::unordered_map<address_t, neb::rt::in_out_val_t> &in_out_vals,
      const account_db_v2_ptr_t &db_ptr)
      -> std::unique_ptr<std::unordered_map<address_t, floatxx_t>>;

  static auto get_account_rank(
      const std::unordered_map<address_t, floatxx_t> &account_median,
      const std::unordered_map<address_t, floatxx_t> &account_weight,
      const rank_params_t &rp)
      -> std::unique_ptr<std::unordered_map<address_t, floatxx_t>>;

  static transaction_graph_ptr_t build_graph_from_transactions(
      const std::vector<neb::fs::transaction_info_t> &trans);

  static block_height_t get_max_height_this_block_interval(
      const std::vector<neb::fs::transaction_info_t> &txs);

  static floatxx_t f_account_weight(floatxx_t in_val, floatxx_t out_val);

  static floatxx_t f_account_rank(int64_t a, int64_t b, int64_t c, int64_t d,
                                  floatxx_t theta, floatxx_t mu,
                                  floatxx_t lambda, floatxx_t S, floatxx_t R);

}; // class nebulas_rank

std::unique_ptr<std::vector<std::vector<neb::fs::transaction_info_t>>>
nebulas_rank_v2::split_transactions_by_block_interval(
    const std::vector<neb::fs::transaction_info_t> &txs,
    int32_t block_interval) {

  auto ret =
      std::make_unique<std::vector<std::vector<neb::fs::transaction_info_t>>>();

  if (block_interval < 1 || txs.empty()) {
    return ret;
  }

  auto it = txs.begin();
  block_height_t block_first = it->m_height;
  it = txs.end();
  it--;
  block_height_t block_last = it->m_height;

  std::vector<neb::fs::transaction_info_t> v;
  it = txs.begin();
  block_height_t b = block_first;
  while (b <= block_last) {
    block_height_t h = it->m_height;
    if (h < b + block_interval) {
      v.push_back(*it++);
    } else {
      ret->push_back(v);
      v.clear();
      b += block_interval;
    }
    if (it == txs.end()) {
      ret->push_back(v);
      break;
    }
  }
  return ret;
}

void nebulas_rank_v2::filter_empty_transactions_this_interval(
    std::vector<std::vector<neb::fs::transaction_info_t>> &txs) {
  for (auto it = txs.begin(); it != txs.end();) {
    if (it->empty()) {
      it = txs.erase(it);
    } else {
      it++;
    }
  }
}

transaction_graph_ptr_t nebulas_rank_v2::build_graph_from_transactions(
    const std::vector<neb::fs::transaction_info_t> &trans) {
  auto ret = std::make_unique<neb::rt::transaction_graph>();

  for (auto ite = trans.begin(); ite != trans.end(); ite++) {
    address_t from = ite->m_from;
    address_t to = ite->m_to;
    wei_t value = ite->m_tx_value;
    int64_t timestamp = ite->m_timestamp;
    ret->add_edge(from, to, value, timestamp);
  }
  return ret;
}

std::unique_ptr<std::vector<transaction_graph_ptr_t>>
nebulas_rank_v2::build_transaction_graphs(
    const std::vector<std::vector<neb::fs::transaction_info_t>> &txs) {

  std::unique_ptr<std::vector<transaction_graph_ptr_t>> tgs =
      std::make_unique<std::vector<transaction_graph_ptr_t>>();

  for (auto it = txs.begin(); it != txs.end(); it++) {
    auto p = build_graph_from_transactions(*it);
    tgs->push_back(std::move(p));
  }
  return tgs;
}

block_height_t nebulas_rank_v2::get_max_height_this_block_interval(
    const std::vector<neb::fs::transaction_info_t> &txs) {
  if (txs.empty()) {
    return 0;
  }
  // suppose transactions in height increasing order
  return txs.back().m_height;
}

std::unique_ptr<std::unordered_set<address_t>>
nebulas_rank_v2::get_normal_accounts(
    const std::vector<neb::fs::transaction_info_t> &txs) {

  auto ret = std::make_unique<std::unordered_set<address_t>>();

  for (auto it = txs.begin(); it != txs.end(); it++) {
    auto from = it->m_from;
    ret->insert(from);

    auto to = it->m_to;
    ret->insert(to);
  }
  return ret;
}

std::unique_ptr<std::unordered_map<address_t, floatxx_t>>
nebulas_rank_v2::get_account_balance_median(
    const std::unordered_set<address_t> &accounts,
    const std::vector<std::vector<neb::fs::transaction_info_t>> &txs,
    const account_db_v2_ptr_t &db_ptr) {

  auto ret = std::make_unique<std::unordered_map<address_t, floatxx_t>>();
  std::unordered_map<address_t, std::vector<wei_t>> addr_balance_v;

  block_height_t max_height = 0;
  for (auto it = txs.begin(); it != txs.end(); it++) {
    block_height_t height = get_max_height_this_block_interval(*it);
    height = std::max(height, max_height);
    for (auto ite = accounts.begin(); ite != accounts.end(); ite++) {
      address_t addr = *ite;
      wei_t balance = db_ptr->get_account_balance_internal(addr, height);
      addr_balance_v[addr].push_back(balance);
    }
    max_height = std::max(max_height, height);
  }

  floatxx_t zero = softfloat_cast<uint32_t, typename floatxx_t::value_type>(0);
  for (auto it = addr_balance_v.begin(); it != addr_balance_v.end(); it++) {
    std::vector<wei_t> &v = it->second;
    sort(v.begin(), v.end(),
         [](const wei_t &w1, const wei_t &w2) { return w1 < w2; });
    size_t v_len = v.size();
    floatxx_t median = to_float<floatxx_t>(v[v_len >> 1]);
    if ((v_len & 0x1) == 0) {
      auto tmp = to_float<floatxx_t>(v[(v_len >> 1) - 1]);
      median = (median + tmp) / 2;
    }

    floatxx_t normalized_median = db_ptr->get_normalized_value(median);
    ret->insert(std::make_pair(it->first, math::max(zero, normalized_median)));
  }

  return ret;
}

floatxx_t nebulas_rank_v2::f_account_weight(floatxx_t in_val,
                                            floatxx_t out_val) {
  floatxx_t pi = math::constants<floatxx_t>::pi();
  floatxx_t atan_val = pi / 2.0;
  if (in_val > 0) {
    atan_val = math::arctan(out_val / in_val);
  }
  auto tmp = math::sin(pi / 4.0 - atan_val);
  return (in_val + out_val) * math::exp((-2.0) * tmp * tmp);
}

std::unique_ptr<std::unordered_map<address_t, floatxx_t>>
nebulas_rank_v2::get_account_weight(
    const std::unordered_map<address_t, neb::rt::in_out_val_t> &in_out_vals,
    const account_db_v2_ptr_t &db_ptr) {

  auto ret = std::make_unique<std::unordered_map<address_t, floatxx_t>>();

  for (auto it = in_out_vals.begin(); it != in_out_vals.end(); it++) {
    wei_t in_val = it->second.m_in_val;
    wei_t out_val = it->second.m_out_val;

    floatxx_t f_in_val = to_float<floatxx_t>(in_val);
    floatxx_t f_out_val = to_float<floatxx_t>(out_val);

    floatxx_t normalized_in_val = db_ptr->get_normalized_value(f_in_val);
    floatxx_t normalized_out_val = db_ptr->get_normalized_value(f_out_val);

    auto tmp = f_account_weight(normalized_in_val, normalized_out_val);
    ret->insert(std::make_pair(it->first, tmp));
  }
  return ret;
}

floatxx_t nebulas_rank_v2::f_account_rank(int64_t a, int64_t b, int64_t c,
                                          int64_t d, floatxx_t theta,
                                          floatxx_t mu, floatxx_t lambda,
                                          floatxx_t S, floatxx_t R) {
  floatxx_t zero = softfloat_cast<uint32_t, typename floatxx_t::value_type>(0);
  floatxx_t one = softfloat_cast<uint32_t, typename floatxx_t::value_type>(1);
  auto gamma = math::pow(theta * R / (R + mu), lambda);
  auto ret = zero;
  if (S > 0) {
    ret = (S / (one + math::pow(a / S, one / b))) * gamma;
  }
  return ret;
}

std::unique_ptr<std::unordered_map<address_t, floatxx_t>>
nebulas_rank_v2::get_account_rank(
    const std::unordered_map<address_t, floatxx_t> &account_median,
    const std::unordered_map<address_t, floatxx_t> &account_weight,
    const rank_params_t &rp) {

  auto ret = std::make_unique<std::unordered_map<address_t, floatxx_t>>();

  for (auto it_m = account_median.begin(); it_m != account_median.end();
       it_m++) {
    auto it_w = account_weight.find(it_m->first);
    if (it_w != account_weight.end()) {
      floatxx_t rank_val =
          f_account_rank(rp.m_a, rp.m_b, rp.m_c, rp.m_d, rp.m_theta, rp.m_mu,
                         rp.m_lambda, it_m->second, it_w->second);
      ret->insert(std::make_pair(it_m->first, rank_val));
    }
  }

  return ret;
}

std::vector<std::shared_ptr<nr_info_t>> nebulas_rank_v2::get_nr_score(
    const transaction_db_v2_ptr_t &tdb_ptr, const account_db_v2_ptr_t &adb_ptr,
    const rank_params_t &rp, neb::block_height_t start_block,
    neb::block_height_t end_block) {

  auto start_time = std::chrono::high_resolution_clock::now();
  LOG(INFO) << "start to read neb db transaction";
  // transactions in total and account inter transactions
  auto txs_ptr =
      tdb_ptr->read_transactions_from_db_with_duration(start_block, end_block);
  LOG(INFO) << "raw tx size: " << txs_ptr->size();
  auto inter_txs_ptr = fs::transaction_db::read_transactions_with_address_type(
      *txs_ptr, NAS_ADDRESS_ACCOUNT_MAGIC_NUM, NAS_ADDRESS_ACCOUNT_MAGIC_NUM);
  LOG(INFO) << "account to account: " << inter_txs_ptr->size();
  auto succ_inter_txs_ptr =
      tdb_ptr->read_transactions_with_succ(*inter_txs_ptr);
  LOG(INFO) << "succ account to account: " << succ_inter_txs_ptr->size();

  // graph operation
  auto txs_v_ptr = split_transactions_by_block_interval(*succ_inter_txs_ptr);
  LOG(INFO) << "split by block interval: " << txs_v_ptr->size();

  filter_empty_transactions_this_interval(*txs_v_ptr);
  auto tgs_ptr = build_transaction_graphs(*txs_v_ptr);
  if (tgs_ptr->empty()) {
    return std::vector<std::shared_ptr<nr_info_t>>();
  }
  LOG(INFO) << "we have " << tgs_ptr->size() << " subgraphs.";
  for (auto it = tgs_ptr->begin(); it != tgs_ptr->end(); it++) {
    transaction_graph *ptr = it->get();
    graph_algo::non_recursive_remove_cycles_based_on_time_sequence(
        ptr->internal_graph());
    graph_algo::merge_edges_with_same_from_and_same_to(ptr->internal_graph());
  }
  LOG(INFO) << "done with remove cycle.";

  transaction_graph *tg = neb::rt::graph_algo::merge_graphs(*tgs_ptr);
  graph_algo::merge_topk_edges_with_same_from_and_same_to(tg->internal_graph());
  LOG(INFO) << "done with merge graphs.";

  // in_out amount
  auto in_out_vals_p = graph_algo::get_in_out_vals(tg->internal_graph());
  auto in_out_vals = *in_out_vals_p;
  LOG(INFO) << "done with get in_out_vals";

  // median, weight, rank
  auto accounts_ptr = get_normal_accounts(*inter_txs_ptr);
  LOG(INFO) << "account size: " << accounts_ptr->size();

  std::unordered_map<neb::address_t, neb::wei_t> addr_balance;
  for (auto &acc : *accounts_ptr) {
    auto balance = adb_ptr->get_balance(acc, start_block);
    addr_balance.insert(std::make_pair(acc, balance));
  }
  LOG(INFO) << "done with get balance";
  adb_ptr->update_height_address_val_internal(*txs_ptr, addr_balance);
  LOG(INFO) << "done with set height address";

  auto account_median_ptr =
      get_account_balance_median(*accounts_ptr, *txs_v_ptr, adb_ptr);
  LOG(INFO) << "done with get account balance median";
  auto account_weight_ptr = get_account_weight(in_out_vals, adb_ptr);
  LOG(INFO) << "done with get account weight";

  auto account_median = *account_median_ptr;
  auto account_weight = *account_weight_ptr;
  auto account_rank_ptr = get_account_rank(account_median, account_weight, rp);
  auto account_rank = *account_rank_ptr;
  LOG(INFO) << "account rank size: " << account_rank.size();

  std::vector<std::shared_ptr<nr_info_t>> infos;
  for (auto it = accounts_ptr->begin(); it != accounts_ptr->end(); it++) {
    address_t addr = *it;
    if (in_out_vals.find(addr) != in_out_vals.end() &&
        account_median.find(addr) != account_median.end() &&
        account_weight.find(addr) != account_weight.end() &&
        account_rank.find(addr) != account_rank.end()) {
      auto in_outs = in_out_vals[addr].m_in_val + in_out_vals[addr].m_out_val;
      auto f_in_outs = to_float<floatxx_t>(in_outs);

      auto info = std::shared_ptr<nr_info_t>(
          new nr_info_t({addr, f_in_outs, account_median[addr],
                         account_weight[addr], account_rank[addr]}));
      infos.push_back(info);
    }
  }

  auto end_time = std::chrono::high_resolution_clock::now();
  LOG(INFO) << "time spend: "
            << std::chrono::duration_cast<std::chrono::seconds>(end_time -
                                                                start_time)
                   .count()
            << " seconds";
  return infos;
}

} // namespace nr
} // namespace rt
} // namespace neb

neb::rt::nr::nr_ret_type entry_point_nr_impl_v2(
    uint64_t start_block, uint64_t end_block, neb::rt::nr::version_t version,
    int64_t a, int64_t b, int64_t c, int64_t d, neb::rt::nr::nr_float_t theta,
    neb::rt::nr::nr_float_t mu, neb::rt::nr::nr_float_t lambda) {

  std::unique_ptr<neb::fs::blockchain_api_base> pba =
      std::unique_ptr<neb::fs::blockchain_api_base>(
          new neb::fs::blockchain_api_v2());
  neb::rt::nr::transaction_db_ptr_t tdb_ptr =
      std::make_unique<neb::fs::transaction_db>(pba.get());
  neb::rt::nr::account_db_ptr_t adb_ptr =
      std::make_unique<neb::fs::account_db>(pba.get());

  auto tdb_ptr_v2 = std::make_unique<neb::fs::transaction_db_v2>(tdb_ptr.get());
  auto adb_ptr_v2 = std::make_unique<neb::fs::account_db_v2>(adb_ptr.get());

  LOG(INFO) << "start block: " << start_block << " , end block: " << end_block;
  neb::rt::nr::rank_params_t rp{a, b, c, d, theta, mu, lambda};

  std::vector<std::pair<std::string, std::string>> meta_info;
  meta_info.push_back(
      std::make_pair("start_height", std::to_string(start_block)));
  meta_info.push_back(std::make_pair("end_height", std::to_string(end_block)));
  meta_info.push_back(std::make_pair("version", std::to_string(version)));

  neb::rt::nr::nr_ret_type ret;
  std::get<0>(ret) = 1;
  std::get<1>(ret) = neb::rt::meta_info_to_json(meta_info);
  std::get<2>(ret) = neb::rt::nr::nebulas_rank_v2::get_nr_score(
      tdb_ptr_v2, adb_ptr_v2, rp, start_block, end_block);

  return ret;
}

neb::rt::nr::nr_ret_type entry_point_nr_v2(neb::compatible_uint64_t start_block,
                                           neb::compatible_uint64_t end_block) {
  auto to_version_t = [](uint32_t major_version, uint16_t minor_version,
                         uint16_t patch_version) -> neb::rt::nr::version_t {
    return (0ULL + major_version) + ((0ULL + minor_version) << 32) +
           ((0ULL + patch_version) << 48);
  };

  neb::compatible_int64_t a = 100;
  neb::compatible_int64_t b = 2;
  neb::compatible_int64_t c = 6;
  neb::compatible_int64_t d = -9;
  neb::rt::nr::nr_float_t theta = 1;
  neb::rt::nr::nr_float_t mu = 1;
  neb::rt::nr::nr_float_t lambda = 2;
  return entry_point_nr_impl_v2(start_block, end_block, to_version_t(2, 0, 0),
                                a, b, c, d, theta, mu, lambda);
}

