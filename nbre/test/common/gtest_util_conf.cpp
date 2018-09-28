#include "common/configuration.h"
#include <gtest/gtest.h>

TEST(test_common_configuration, read_config) {
  EXPECT_EQ(neb::configuration::instance().exec_name(), "");
  EXPECT_EQ(neb::configuration::instance().runtime_library_path(), "");

  const char *argv[3] = {"", "--ini-file",
                         "../test/data/test_configuration.ini"};

  neb::configuration::instance().init_with_args(3, argv);
  EXPECT_EQ(neb::configuration::instance().exec_name(), "bar");
  EXPECT_EQ(neb::configuration::instance().runtime_library_path(), "./lib");
}

TEST(test_common_configuration, throw_config) {
  const char *argv1[3] = {"", "--ini-file", "../test/data/test_xxxx.ini"};

  EXPECT_THROW(neb::configuration::instance().init_with_args(3, argv1),
               neb::configure_general_failure);

  const char *argv2[3] = {"", "--xxxx", "../test/data/test_configuration.ini"};

  EXPECT_THROW(neb::configuration::instance().init_with_args(3, argv2),
               neb::configure_general_failure);
}
