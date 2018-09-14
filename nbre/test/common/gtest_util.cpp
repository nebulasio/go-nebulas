#include "common/util/byte.h"
#include <gtest/gtest.h>

TEST(test_common_util, simple) {
  int32_t v = 123;
  neb::byte_t buf[4];
  neb::util::number_to_byte(v, buf, 4);

  int32_t ret = neb::util::byte_to_number<int32_t>(buf, 4);

  EXPECT_EQ(v, ret);
}

TEST(test_common_util_byte, from_uint64) {
  uint64_t v1 = 0;
  neb::util::bytes b1 = neb::util::number_to_byte<neb::util::bytes>(v1);
  EXPECT_TRUE(b1 == neb::util::bytes({0, 0, 0, 0, 0, 0, 0, 0})) << " 0 failed";

  uint64_t v2 = 1024;
  neb::util::bytes b2 = neb::util::number_to_byte<neb::util::bytes>(v2);
  EXPECT_TRUE(b2 == neb::util::bytes({0, 0, 0, 0, 0, 0, 4, 0}))
      << " 1024 failed";

  uint64_t v3 = 18446744073709551615u;
  neb::util::bytes b3 = neb::util::number_to_byte<neb::util::bytes>(v3);
  EXPECT_TRUE(b3 == neb::util::bytes({255, 255, 255, 255, 255, 255, 255, 255}))
      << "uint64 max failed";
}

TEST(test_common_util_byte, fix_bytes_default) {
  neb::util::fix_bytes<> fb;

  std::string base58 = fb.to_base58();

  EXPECT_EQ(base58, "0");
}

TEST(test_common_util_byte, fix_bytes_Encode) {
  neb::util::fix_bytes<> source = neb::util::fix_bytes<>::from_base58("Hello, world");
  neb::util::fix_bytes<> want({72, 101, 108, 108, 111, 44, 32, 119, 111, 114, 108, 100});

  EXPECT_EQ(source, want);
}

TEST(test_common_util_byte, fix_bytes_Decode) {
  neb::util::fix_bytes<> fb({72, 101, 108, 108, 111, 44, 32, 119, 111, 114, 108, 100});

  std::string source = fb.to_base58();

  EXPECT_EQ(source, "Hello, world");
}


