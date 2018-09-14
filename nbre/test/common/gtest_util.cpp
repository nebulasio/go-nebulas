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
  EXPECT_TRUE(b1 == neb::util::bytes({0, 0, 0, 0, 0, 0, 0, 0}));

  uint64_t v2 = 1024;
  neb::util::bytes b2 = neb::util::number_to_byte<neb::util::bytes>(v2);
  EXPECT_TRUE(b2 == neb::util::bytes({0, 0, 0, 0, 0, 0, 4, 0}));

  uint64_t v3 = 18446744073709551615u;
  neb::util::bytes b3 = neb::util::number_to_byte<neb::util::bytes>(v3);
  EXPECT_TRUE(b2 == neb::util::bytes({255, 255, 255, 255, 255, 255, 255, 255}));
}
