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

#include "runtime/nr/impl/nr_handler.h"
#include "common/configuration.h"
#include "core/ir_warden.h"
#include "fs/proto/ir.pb.h"
#include "jit/jit_driver.h"
#include "runtime/nr/impl/nebulas_rank.h"
#include <ff/functionflow.h>

namespace neb {
namespace rt {
namespace nr {

nr_handler::nr_handler() {}

void nr_handler::run_if_default(block_height_t start_block,
                                block_height_t end_block,
                                const std::string &nr_handle) {
  ff::para<> p;
  p([this, start_block, end_block, nr_handle]() {
    try {
      jit_driver &jd = jit_driver::instance();
      auto nr_ret = jd.run_ir<nr_ret_type>(
          "nr", start_block, neb::configuration::instance().nr_func_name(),
          start_block, end_block);
      m_nr_result.set(nr_handle, nr_ret);
    } catch (const std::exception &e) {
      LOG(INFO) << "jit driver execute nr failed " << e.what();
    }
  });
}

void nr_handler::run_if_specify(block_height_t start_block,
                                block_height_t end_block, uint64_t nr_version,
                                const std::string &nr_handle) {
  ff::para<> p;
  p([this, start_block, end_block, nr_version, nr_handle]() {
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
  });
}

void nr_handler::start(const std::string &nr_handle) {

  if (!nr_handle.empty() && m_nr_result.exists(nr_handle)) {
    return;
  }

  neb::bytes nr_handle_bytes = neb::bytes::from_hex(nr_handle);
  size_t bytes = sizeof(uint64_t) / sizeof(byte_t);
  assert(nr_handle_bytes.size() == 3 * bytes);

  const auto &s = neb::bytes(nr_handle_bytes.value(), bytes);
  const auto &e = neb::bytes(nr_handle_bytes.value() + bytes, bytes);
  const auto &v = neb::bytes(nr_handle_bytes.value() + 2 * bytes, bytes);

  uint64_t start_block = neb::byte_to_number<uint64_t>(s);
  uint64_t end_block = neb::byte_to_number<uint64_t>(e);
  uint64_t nr_version = neb::byte_to_number<uint64_t>(v);

  if (!nr_version) {
    run_if_default(start_block, end_block, nr_handle);
    return;
  }

  run_if_specify(start_block, end_block, nr_version, nr_handle);
}

nr_ret_type nr_handler::get_nr_result(const std::string &nr_handle) {

  nr_ret_type nr_result;
  auto ret = m_nr_result.get(nr_handle, nr_result);
  if (!ret) {
    LOG(INFO) << std::string(
        "{\"err\":\"nr hash expired or nr result not complete yet\"}");
  }
  return nr_result;
}
} // namespace nr
} // namespace rt
} // namespace neb
