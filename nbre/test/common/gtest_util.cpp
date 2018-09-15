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

//TEST(test_common_util_byte, positive_to_base58) {
  // neb::util::bytes bytes({72, 101, 108, 108, 111, 44, 32, 119, 111, 114, 108, 100});
  // std::string result = bytes.to_base58();
  // std::string want("Hello, world");
  // EXPECT_EQ(result, want);
//}

//TEST(test_common_util_byte, positive_to_hex) {
  // neb::util::bytes bytes({72, 101, 108, 108, 111, 44, 32, 119, 111, 114, 108, 100});
  // std::string result = bytes.to_base58();
  // std::string want("Hello, world");
  // EXPECT_EQ(result, want);
//}

// TEST(test_common_util_byte, positive_from_base58) {
  // neb::util::bytes want({72, 101, 108, 108, 111, 44, 32, 119, 111, 114, 108, 100});
  // neb::util::bytes result = neb::util::bytes::from_base58("0");
  // EXPECT_EQ(result, want);
// }

TEST(test_common_util_byte, test_default) {
  neb::util::fix_bytes<> fb;

  std::string base58 = fb.to_base58();

  EXPECT_EQ(base58, "11111111111111111111111111111111");
}

TEST(test_common_util_byte, positive_fix_bytes_to_base58) {
  neb::util::fix_bytes<6> fb({32, 119, 111, 114, 108, 100});

  std::string result = fb.to_base58();
  std::string want = "HAi6xaJX";

  EXPECT_EQ(result, want);
}

TEST(test_common_util_byte, positive_fix_bytes_to_base58_2) {
  neb::util::fix_bytes<6> fb({ 0x47, 0x4b, 0xc9 ,0x32});

  std::string result = fb.to_base58();
  std::string want = "HAi6xaJX";

  EXPECT_EQ(result, want);
}

TEST(test_common_util_byte, positive_fix_bytes_to_hex) {
  neb::util::fix_bytes<6> fb({132, 11, 111, 104, 18, 100});

  std::string result = fb.to_hex();
  std::string want = "0x7fffb0b23c48";

  EXPECT_EQ(result, want);
}

TEST(test_common_util_byte, positive_fix_bytes_size) {
  neb::util::fix_bytes<12> fb({72, 101, 108, 108, 111, 44, 32, 119, 111, 114, 108, 100});

  int result = fb.size();
  int want = 12;

  EXPECT_EQ(result, want);
}

TEST(test_common_util_byte, positive_fix_bytes_value) {
  neb::util::fix_bytes<12> fb({72, 101, 108, 108, 111, 44, 32, 119, 111, 114, 108, 100});

  neb::byte_t *result = fb.value();
  std::cout <<  printf("----------ok");
  neb::byte_t want[4] = {1, 2, 3, 4};

  EXPECT_EQ(result, want);
}

TEST(test_common_util_byte, positive_fix_bytes_from_base58) {
  neb::util::fix_bytes<12> result = neb::util::fix_bytes<12>::from_base58("2NEp7TZsLFA2wMeK");
  neb::util::fix_bytes<12> want({72, 101, 108, 108, 111, 44, 32, 119, 111, 114, 108, 100});

  EXPECT_EQ(result, want);
}

TEST(test_common_util_byte, positive_fix_bytes_from_hex) {
  neb::util::fix_bytes<12> result = neb::util::fix_bytes<12>::from_hex("10AB");
  neb::util::fix_bytes<12> want({72, 101, 108, 108, 111, 44, 32, 119, 111, 114, 108, 100});

  EXPECT_EQ(result, want);
}

TEST(test_common_util_byte, negative_fix_bytes_to_hex) {
  neb::util::fix_bytes<12> fb({72, 101, 108, 108, 111, 44, 32, 119, 111, 114, 108, 100});

  std::string result = fb.to_hex();
  std::string want = "Hello, world";

  EXPECT_EQ(result, want);
}

TEST(test_common_util_byte, negative_fix_bytes_from_hex) {
  neb::util::fix_bytes<12> result = neb::util::fix_bytes<12>::from_hex("Hello, world");
  neb::util::fix_bytes<12> want({72, 101, 108, 108, 111, 44, 32, 119, 111, 114, 108, 100});

  EXPECT_EQ(result, want);
}



