#include "common/configuration.h"
#include <gtest/gtest.h>

TEST(test_common_configuration, read_config) {
  EXPECT_EQ(neb::configuration::instance().exec_name(), "bar");
  EXPECT_EQ(neb::configuration::instance().runtime_library_path(), "./lib");
}
