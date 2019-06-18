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

#include "core/nr_handler.h"
#include "common/configuration.h"
#include "core/ir_warden.h"
#include "fs/proto/ir.pb.h"
#include "jit/jit_driver.h"
#include "runtime/util.h"
#include <ff/functionflow.h>

namespace neb {
namespace core {

nr_handler::nr_handler() {}

#if 0
void nr_handler::run_if_default(block_height_t start_block,
                                block_height_t end_block,
                                const std::string &nr_handle) {
  // ff::para<> p;
  // p([this, start_block, end_block, nr_handle]() {
  try {
    jit_driver &jd = jit_driver::instance();
    auto nr_ret = jd.run_ir<nr_ret_type>(
        "nr", start_block, neb::configuration::instance().nr_func_name(),
        start_block, end_block);
    m_nr_result.set(nr_handle, nr_ret);
    } catch (const std::exception &e) {
      LOG(INFO) << "jit driver execute nr failed " << e.what();
    }
    //});
}

void nr_handler::run_if_specify(block_height_t start_block,
                                block_height_t end_block, uint64_t nr_version,
                                const std::string &nr_handle) {
  // ff::para<> p;
  // p([this, start_block, end_block, nr_version, nr_handle]() {
  try {
    std::string nr_name = "nr";
    std::vector<nbre::NBREIR> irs;
    auto ir = neb::core::ir_warden::instance().get_ir_by_name_version(
        nr_name, nr_version);
    irs.push_back(*ir);

    std::stringstream ss;
    ss << nr_name << nr_version;
    std::string name_version = ss.str();

    jit_driver &jd = jit_driver::instance();
    auto nr_ret = jd.run<nr_ret_type>(
        name_version, irs, neb::configuration::instance().nr_func_name(),
        start_block, end_block);
    m_nr_result.set(nr_handle, nr_ret);
    } catch (const std::exception &e) {
      LOG(INFO) << "jit driver execute nr failed " << e.what();
    }
    //});
}

#endif

void nr_handler::start(block_height_t start_block, block_height_t end_block,
                       uint64_t nr_version) {
#if 0
  std::string nr_handle = param_to_key(start_block, end_block, nr_version);
  if (!nr_handle.empty() && m_nr_result.exists(nr_handle)) {
    return;
  }
  m_running_nr_handler_mutex.lock();
  if (m_running_nr_handlers.find(nr_handle) != m_running_nr_handlers.end()) {
    m_running_nr_handler_mutex.unlock();
    return;
  }
  m_running_nr_handlers.insert(nr_handle);
  m_running_nr_handler_mutex.unlock();

  auto remove_nr_hander_from_running = [this, nr_handle]() {
    m_running_nr_handler_mutex.lock();
    m_running_nr_handlers.erase(nr_handle);
    m_running_nr_handler_mutex.unlock();
  };
  if (!nr_version) {
    run_if_default(start_block, end_block, nr_handle);
    remove_nr_hander_from_running();
    return;
  }

  run_if_specify(start_block, end_block, nr_version, nr_handle);
  remove_nr_hander_from_running();
#endif
}

std::string nr_handler::get_nr_handle(block_height_t start_block,
                                      block_height_t end_block,
                                      uint64_t version) {
  return rt::param_to_key(start_block, end_block, version);
}

std::string nr_handler::get_nr_handle(block_height_t height) {
  throw std::runtime_error("no impl");
}

rt::nr::nr_ret_type nr_handler::get_nr_result(const std::string &nr_handle) {

  rt::nr::nr_ret_type nr_result;
#if 0
  bool status = m_checker.get_nr_result(nr_result, nr_handle);
  if (!status) {
    nr_result = m_cache.get_nr_score(nr_handle);
  }
#endif
  return nr_result;
}
bool nr_handler::get_nr_sum(floatxx_t &nr_sum, const std::string &handle) {

  //! TODO: we need compute NR if not exist
  nr_sum = floatxx_t();
  rt::nr::nr_ret_type nr_result = get_nr_result(handle);
  if (!core::is_succ(nr_result)) {
    return false;
  }
  const std::vector<nr_item> &nrs = nr_result->get<p_nr_items>();
  for (auto &ts : nrs) {
    nr_sum += ts.get<p_nr_item_score>();
  }
  return true;
}
bool nr_handler::get_nr_addr_list(std::vector<address_t> &nr_addrs,
                                  const std::string &handle) {
  //! TODO: we need compute NR if not exist
  nr_addrs.clear();
  rt::nr::nr_ret_type nr_result = get_nr_result(handle);
  if (!core::is_succ(nr_result)) {
    return false;
  }
  const std::vector<nr_item> &nrs = nr_result->get<p_nr_items>();
  for (auto &ts : nrs) {
    nr_addrs.push_back(to_address(ts.get<p_nr_item_addr>()));
  }
  return true;
}
} // namespace core
} // namespace neb
