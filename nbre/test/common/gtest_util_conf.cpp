#include "common/configuration.h"
#include <gtest/gtest.h>

TEST(test_common_configuration, read_config) {
  EXPECT_EQ(neb::configuration::instance().exec_name(), "bar");
  EXPECT_EQ(neb::configuration::instance().runtime_library_path(), "./lib");
}
//
// TEST(test_common_json_util, throw_json) {
  // EXPECT_THROW(neb::ir_conf_reader json_reader("xxx"),
               // neb::json_general_failure);
//
  // EXPECT_THROW(neb::ir_conf_reader json_reader("../test/data/test_throw_exceptions.json"),
               // neb::json_general_failure);
//}
