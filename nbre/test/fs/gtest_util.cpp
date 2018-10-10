#include "fs/util.h"
#include <gtest/gtest.h>
#include <iostream>

TEST(test_fs_util, simple) {
  std::string cur_path = neb::fs::cur_full_path();
  std::string cur_dir = neb::fs::cur_dir();
  std::string tmp_dir = neb::fs::tmp_dir();

  EXPECT_TRUE(cur_path.size() > 0);
  EXPECT_TRUE(cur_dir.size() > 0);
  EXPECT_TRUE(tmp_dir.size() > 0);

  std::string rocksdb_data_path =
      neb::fs::join_path(cur_dir, "test/data/read-data.db/CURRENT");

  EXPECT_TRUE(neb::fs::exists(rocksdb_data_path));
}
