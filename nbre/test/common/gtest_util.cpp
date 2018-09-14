#include "common/util/byte.h"
#include <gtest/gtest.h>

TEST(test_common_util, simple) {
  int32_t v = 123;
  neb::byte_t buf[4];
  neb::util::number_to_byte(v, buf, 4);

  int32_t ret = neb::util::byte_to_number<int32_t>(buf, 4);

  EXPECT_EQ(v, ret);
}
