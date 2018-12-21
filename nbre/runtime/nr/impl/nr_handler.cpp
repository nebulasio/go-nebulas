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
#include "core/ir_warden.h"
#include "fs/proto/ir.pb.h"
#include "jit/jit_driver.h"
#include "runtime/nr/impl/nebulas_rank.h"
#include <ff/ff.h>

namespace neb {
namespace rt {
namespace nr {

nr_handler::nr_handler() {}

std::string nr_handler::get_nr_handler_id() {
  std::unique_lock<std::mutex> _l(m_sync_mutex);
  return m_nr_handler_id;
}

void nr_handler::start(std::string nr_handler_id) {
  std::unique_lock<std::mutex> _l(m_sync_mutex);
  m_nr_handler_id = nr_handler_id;
  if (!m_nr_handler_id.empty() &&
      m_nr_result.find(m_nr_handler_id) != m_nr_result.end()) {
    m_nr_handler_id.clear();
    return;
  }

  ff::para<> p;
  p([this]() {
    neb::util::bytes nr_handler_bytes =
        neb::util::bytes::from_hex(m_nr_handler_id);

    size_t bytes = sizeof(uint64_t) / sizeof(byte_t);
    assert(nr_handler_bytes.size() == 3 * bytes);

    uint64_t start_block = neb::util::byte_to_number<uint64_t>(
        neb::util::bytes(nr_handler_bytes.value(), bytes));
    uint64_t end_block = neb::util::byte_to_number<uint64_t>(
        neb::util::bytes(nr_handler_bytes.value() + bytes, bytes));
    uint64_t nr_version = neb::util::byte_to_number<uint64_t>(
        neb::util::bytes(nr_handler_bytes.value() + 2 * bytes, bytes));

    try {
      std::string nr_name = "nr";
      std::vector<nbre::NBREIR> irs;
      auto ir = neb::core::ir_warden::instance().get_ir_by_name_version(
          nr_name, nr_version);
      irs.push_back(*ir);

      jit_driver &jd = jit_driver::instance();
      std::stringstream ss;
      ss << nr_name << nr_version;

      //  TODO func name
      std::string nr_result = jd.run<std::string>(
          ss.str(), irs, "_Z14entry_point_nrB5cxx11mm", start_block, end_block);

      auto it_nr_infos = nebulas_rank::json_to_nr_info(nr_result);
      nr_result = nebulas_rank::nr_info_to_json(*it_nr_infos,
                                                {{"start_height", start_block},
                                                 {"end_height", end_block},
                                                 {"version", nr_version}});

      m_nr_result.insert(std::make_pair(m_nr_handler_id, nr_result));
      m_nr_handler_id.clear();
    } catch (const std::exception &e) {
      LOG(INFO) << "jit driver execute nr failed " << e.what();
    }

  });
}

std::string nr_handler::get_nr_result(const std::string &nr_handler_id) {
  std::unique_lock<std::mutex> _l(m_sync_mutex);

  auto nr_result = m_nr_result.find(nr_handler_id);
  if (nr_result == m_nr_result.end()) {
    return std::string("{\"err\":\"not complete yet\"}");
  }
  return nr_result->second;
}
} // namespace nr
} // namespace rt
} // namespace neb
