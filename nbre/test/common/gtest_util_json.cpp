#include "common/ir_conf_reader.h"
#include <gtest/gtest.h>

TEST(test_common_json_util, read_json) {
  neb::ir_conf_reader json_reader("../test/data/test_nasir.json");
  EXPECT_EQ(json_reader.self_ref().name(), "xxx");
  EXPECT_EQ(json_reader.self_ref().version().major_version(), 1);
  EXPECT_EQ(json_reader.self_ref().version().minor_version(), 2);
  EXPECT_EQ(json_reader.self_ref().version().patch_version(), 3);
  EXPECT_EQ(json_reader.available_height(), 15);
  EXPECT_EQ(json_reader.ir_fp(), "../runtime/bar.bc");
  EXPECT_EQ(json_reader.depends()[0].name(), "yyy");
  EXPECT_EQ(json_reader.depends()[0].version().major_version(), 11);
  EXPECT_EQ(json_reader.depends()[0].version().minor_version(), 12);
  EXPECT_EQ(json_reader.depends()[0].version().patch_version(), 13);
  EXPECT_EQ(json_reader.depends()[1].name(), "zzz");
  EXPECT_EQ(json_reader.depends()[1].version().major_version(), 21);
  EXPECT_EQ(json_reader.depends()[1].version().minor_version(), 22);
  EXPECT_EQ(json_reader.depends()[1].version().patch_version(), 23);
}

TEST(test_common_json_util, throw_json) {
  EXPECT_THROW(neb::ir_conf_reader json_reader("xxx"),
               neb::json_general_failure);

  EXPECT_THROW(neb::ir_conf_reader json_reader("../test/data/test_throw_exceptions.json"),
               neb::json_general_failure);
}
