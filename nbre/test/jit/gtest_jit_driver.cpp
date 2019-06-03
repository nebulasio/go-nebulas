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
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the // GNU General
// Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see
// <http://www.gnu.org/licenses/>.
//

#include "common/configuration.h"
#include "fs/storage_holder.h"
#include "jit/cpp_ir.h"
#include "jit/jit_driver.h"
#include "test/jit/ir/gen_ir.h"
#include <gtest/gtest.h>

TEST(test_jit_driver, get_mangled_entry_point) {
  neb::jit_driver &jd = neb::jit_driver::instance();

  auto auth_func_name = neb::configuration::instance().auth_func_name();
  auto auth_func_mangled = jd.get_mangled_entry_point(auth_func_name);
  EXPECT_EQ(auth_func_mangled, "_Z16entry_point_authB5cxx11v");

  auto nr_func_name = neb::configuration::instance().nr_func_name();
  auto nr_func_mangled = jd.get_mangled_entry_point(nr_func_name);
  EXPECT_EQ(nr_func_mangled, "_Z14entry_point_nrB5cxx11yy");

  auto dip_func_name = neb::configuration::instance().dip_func_name();
  auto dip_func_mangled = jd.get_mangled_entry_point(dip_func_name);
  EXPECT_EQ(dip_func_mangled, "_Z15entry_point_dipB5cxx11y");
}

TEST(test_jit_driver, get_mangled_entry_point_illegal) {
  neb::jit_driver &jd = neb::jit_driver::instance();

  auto ret = jd.get_mangled_entry_point(std::string());
  EXPECT_EQ(ret, std::string());

  ret = jd.get_mangled_entry_point("_Z16entry_point_authB5cxx11v");
  EXPECT_EQ(ret, std::string());
}

void check_auth_ret(const neb::auth_table_t &ret) {
  EXPECT_EQ(ret.size(), 2);

  auto tmp = ret[0];
  EXPECT_EQ(std::get<0>(tmp), "nr");
  EXPECT_EQ(neb::to_address(std::get<1>(tmp)),
            neb::base58_to_address("n1KxWR8ycXg7Kb9CPTtNjTTEpvka269PniB"));
  EXPECT_EQ(std::get<2>(tmp), 100ULL);
  EXPECT_EQ(std::get<3>(tmp), 200ULL);

  tmp = ret[1];
  EXPECT_EQ(std::get<0>(tmp), "dip");
  EXPECT_EQ(neb::to_address(std::get<1>(tmp)),
            neb::base58_to_address("n1Wt2VbPAR6TttM17HQXscCyWBrFe36HeYC"));
  EXPECT_EQ(std::get<2>(tmp), 100ULL);
  EXPECT_EQ(std::get<3>(tmp), 200ULL);
}

TEST(test_jit_driver, run) {
  neb::jit_driver &jd = neb::jit_driver::instance();

  auto auth_func_name = neb::configuration::instance().auth_func_name();

  auto ir = gen_auth_ir();
  std::stringstream ss;
  ss << ir.name() << ir.version();

  neb::cpp::cpp_ir ci(std::make_pair(ss.str(), ir.ir()));
  neb::bytes ir_bytes = ci.llvm_ir_content();
  ir.set_ir(neb::byte_to_string(ir_bytes));

  std::vector<nbre::NBREIR> irs;
  irs.push_back(ir);

  auto ret = jd.run<neb::auth_table_t>(ss.str(), irs, auth_func_name);
  check_auth_ret(ret);
}

TEST(test_jit_driver, run_if_exists) {
  neb::jit_driver &jd = neb::jit_driver::instance();

  auto auth_func_name = neb::configuration::instance().auth_func_name();

  auto ir = gen_auth_ir();
  std::stringstream ss;
  ss << ir.name() << ir.version();
  ss << auth_func_name;

  neb::cpp::cpp_ir ci(std::make_pair(ss.str(), ir.ir()));
  neb::bytes ir_bytes = ci.llvm_ir_content();
  ir.set_ir(neb::byte_to_string(ir_bytes));

  std::vector<nbre::NBREIR> irs;
  irs.push_back(ir);

  auto ret = jd.run_if_exists<neb::auth_table_t>(ir, auth_func_name);
  EXPECT_EQ(ret.first, false);

  jd.run<neb::auth_table_t>(ss.str(), irs, auth_func_name);
  ret = jd.run_if_exists<neb::auth_table_t>(ir, auth_func_name);
  EXPECT_EQ(ret.first, true);
  check_auth_ret(ret.second);
}

TEST(test_jit_driver, run_ir) {
  neb::jit_driver &jd = neb::jit_driver::instance();

  auto auth_func_name = neb::configuration::instance().auth_func_name();

  auto ir = gen_auth_ir();
  std::stringstream ss;
  ss << ir.name() << ir.version();
  ss << auth_func_name;

  neb::cpp::cpp_ir ci(std::make_pair(ss.str(), ir.ir()));
  neb::bytes ir_bytes = ci.llvm_ir_content();
  ir.set_ir(neb::byte_to_string(ir_bytes));

  auto size = ir.ByteSizeLong();
  neb::bytes buf(size);
  ir.SerializeToArray((void *)buf.value(), buf.size());

  auto *rs = neb::fs::storage_holder::instance().nbre_db_ptr();
  neb::fs::ir_manager_helper::update_ir_list(ir.name(), rs);
  neb::fs::ir_manager_helper::update_ir_versions(ir.name(), ir.version(), rs);
  neb::fs::ir_manager_helper::deploy_ir(ir.name(), ir.version(), buf, rs);

  auto ret = jd.run_ir<neb::auth_table_t>(ir.name(), 150, auth_func_name);
  check_auth_ret(ret);

  EXPECT_THROW(jd.run_ir<neb::auth_table_t>(ir.name(), 149, auth_func_name),
               std::invalid_argument);
  EXPECT_THROW(jd.run_ir<neb::auth_table_t>(ir.name(), 123, auth_func_name),
               std::invalid_argument);
  EXPECT_THROW(jd.run_ir<neb::auth_table_t>(ir.name(), 101, auth_func_name),
               std::invalid_argument);
  EXPECT_THROW(jd.run_ir<neb::auth_table_t>(ir.name(), 100, auth_func_name),
               std::invalid_argument);
  EXPECT_THROW(jd.run_ir<neb::auth_table_t>(ir.name(), 99, auth_func_name),
               std::invalid_argument);
  EXPECT_THROW(jd.run_ir<neb::auth_table_t>(ir.name(), 66, auth_func_name),
               std::invalid_argument);
  EXPECT_THROW(jd.run_ir<neb::auth_table_t>(ir.name(), 1, auth_func_name),
               std::invalid_argument);
  EXPECT_THROW(jd.run_ir<neb::auth_table_t>(ir.name(), 0, auth_func_name),
               std::invalid_argument);

  ret = jd.run_ir<neb::auth_table_t>(ir.name(), 151, auth_func_name);
  check_auth_ret(ret);

  ret = jd.run_ir<neb::auth_table_t>(ir.name(), 200, auth_func_name);
  check_auth_ret(ret);

  ret = jd.run_ir<neb::auth_table_t>(ir.name(), 300, auth_func_name);
  check_auth_ret(ret);
}
