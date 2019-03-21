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
#include "cmd/dummy_neb/generator/nr_ir_generator.h"
#include "fs/proto/ir.pb.h"
#include <sstream>

static std::string ir_slice =
    "#include \"runtime/nr/impl/nr_impl.h\" \n"
    "neb::rt::nr::nr_ret_type entry_point_nr(neb::compatible_uint64_t "
    "start_block, \n"
    "                           neb::compatible_uint64_t end_block) { \n"
    "  neb::rt::nr::nr_ret_type ret; \n"
    "  std::get<0>(ret) = 0; \n"
    "  if (start_block > end_block) { \n"
    "    std::get<1>(ret) = std::string(\"{\\\"err\\\":\\\"start height must "
    "less than end height\\\"}\"); \n"
    "    return ret; \n"
    "  } \n"
    "  uint64_t block_nums_of_a_day = 24 * 3600 / 15; \n"
    "  uint64_t days = 7; \n"
    "  uint64_t block_interval = days * block_nums_of_a_day; \n"
    "  if (start_block + block_interval < end_block) { \n"
    "    std::get<1>(ret) = std::string(\"{\\\"err\\\":\\\"nr block interval "
    "out of range\\\"}\"); \n"
    "    return ret; \n"
    "  } \n"
    "  auto to_version_t = [](uint32_t major_version, uint16_t minor_version, "
    "\n"
    "                         uint16_t patch_version) -> "
    "neb::rt::nr::version_t { \n"
    "    return (0ULL + major_version) + ((0ULL + minor_version) << 32) + \n"
    "           ((0ULL + patch_version) << 48); \n"
    "  }; \n"
    "  neb::compatible_int64_t a = VAR_a; \n"
    "  neb::compatible_int64_t b = VAR_b; \n"
    "  neb::compatible_int64_t c = VAR_c; \n"
    "  neb::compatible_int64_t d = VAR_d; \n"
    "  neb::rt::nr::nr_float_t theta = VAR_theta; \n"
    "  neb::rt::nr::nr_float_t mu = VAR_mu; \n"
    "  neb::rt::nr::nr_float_t lambda = VAR_lambda; \n"
    "  return neb::rt::nr::entry_point_nr_impl(start_block, end_block, \n"
    "                                          to_version_t(VERSION_major, "
    "VERSION_minor, VERSION_patch), a, b, c, "
    "d, \n"
    "                                          theta, mu, lambda); \n"
    "} \n";

std::string gen_ir_with_params(int64_t a, int64_t b, int64_t c, int64_t d,
                               float theta, float mu, float lambda,
                               int32_t major_version, int32_t minor_version,
                               int32_t patch_version) {
  std::string ret = ir_slice;
  std::stringstream ss;
#define R(t, name)                                                             \
  ss << t;                                                                     \
  boost::replace_all(ret, name, std::to_string(t));                            \
  ss.clear();

  R(a, "VAR_a");
  R(b, "VAR_b");
  R(c, "VAR_c");
  R(d, "VAR_d");
  R(theta, "VAR_theta");
  R(mu, "VAR_mu");
  R(lambda, "VAR_lambda");
  R(major_version, "VERSION_major");
  R(minor_version, "VERSION_minor");
  R(patch_version, "VERSION_patch");
#undef R

  return ret;
}

nr_ir_generator::nr_ir_generator(generate_block *block, const address_t &addr)
    : generator_base(block->get_all_accounts(), block, 0, 1),
      m_nr_admin_addr(addr) {
  m_a = 100;
  m_b = 2;
  m_c = 6;
  m_d = -9;
  m_theta = 1;
  m_mu = 1;
  m_lambda = 2;
}

nr_ir_generator::~nr_ir_generator() {}

std::shared_ptr<corepb::Account> nr_ir_generator::gen_account() {
  return nullptr;
}
std::shared_ptr<corepb::Transaction> nr_ir_generator::gen_tx() {

  std::string payload =
      gen_ir_with_params(m_a, m_b, m_c, m_d, m_theta, m_mu, m_lambda,
                         m_major_version, m_minor_version, m_patch_version);
  nbre::NBREIR ir;
  ir.set_name("nr");
  neb::version v(m_major_version, m_minor_version, m_patch_version);
  ir.set_version(v.data());
  ir.set_height(m_block->height());
  ir.set_ir(payload);
  ir.set_ir_type(neb::ir_type::cpp);

  std::string ir_str = ir.SerializeAsString();
  return m_block->add_protocol_transaction(m_nr_admin_addr,
                                           neb::string_to_byte(ir_str));
}
checker_tasks::task_container_ptr_t nr_ir_generator::gen_tasks() {
  return nullptr;
}

