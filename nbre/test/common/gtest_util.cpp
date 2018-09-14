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

TEST(test_common_util_byte, test_default) {
  neb::util::fix_bytes<> fb;

  std::string base58 = fb.to_base58(); 

  EXPECT_EQ(base58, "0");
}

TEST(test_common_util_byte, test_encode) {
  neb::util::fix_bytes<> source = neb::util::fix_bytes<>::from_base58("Hello, world");
  neb::util::fix_bytes<> want({72, 101, 108, 108, 111, 44, 32, 119, 111, 114, 108, 100});

  EXPECT_EQ(source, want);
}

TEST(test_common_util_byte, test_decode) {
  neb::util::fix_bytes<> fb({72, 101, 108, 108, 111, 44, 32, 119, 111, 114, 108, 100});

  std::string source = fb.to_base58(); 

  EXPECT_EQ(source, "Hello, world");
}

typedef struct test_source_want {
  std::string want;
  neb::util::fix_bytes<> source;
} test_source_want_t;

TEST(test_common_util_byte, test_hex) {
  neb::util::fix_bytes<> fb0({167, 255, 198, 248, 191, 30, 215, 102, 81, 193, 71, 86, 160, 97, 214, 98, 245, 128, 255, 77, 228, 59, 73, 250, 130, 216, 10, 75, 128, 248, 67, 74});
  neb::util::fix_bytes<> fb1({53, 80, 171, 169, 116, 146, 222, 56, 175, 48, 102, 240, 21, 127, 197, 50, 219, 103, 145, 179, 125, 83, 38, 44, 231, 104, 141, 204, 93, 70, 24, 86});
  neb::util::fix_bytes<> fb2({});

  test_source_want_t tsw[3] = {
    "a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a",
    fb0,
    "3550aba97492de38af3066f0157fc532db6791b37d53262ce7688dcc5d461856",
    fb1,
    "blank string", 
    fb2
  };

  for (int i = 0; i < 3; i++) {
    std::string hex = tsw[i].source.to_hex();
    EXPECT_EQ(hex, tsw[i].want);
  }
}



