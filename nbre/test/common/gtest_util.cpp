// Copyright (C) 2018 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or
// modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see
// <http://www.gnu.org/licenses/>.
//

#include "common/address.h"
#include "common/byte.h"
#include <gtest/gtest.h>

TEST(test_common_util, simple) {
  int32_t v = 123;
  neb::byte_t buf[4];
  neb::number_to_byte(v, buf, 4);

  int32_t ret = neb::byte_to_number<int32_t>(buf, 4);

  EXPECT_EQ(v, ret);
}

template <typename NT>
void test_number_bytes(NT v, const std::initializer_list<neb::byte_t> &l) {
  neb::bytes b = neb::number_to_byte<neb::bytes>(v);

  neb::bytes wb(l);
  EXPECT_EQ(b, wb);
  EXPECT_EQ(v, neb::byte_to_number<NT>(wb));
}
TEST(test_common_util_byte, from_uint64) {
  test_number_bytes<uint64_t>(0, {0, 0, 0, 0, 0, 0, 0, 0});
  test_number_bytes<uint64_t>(1024, {0, 0, 0, 0, 0, 0, 4, 0});
  test_number_bytes<uint64_t>(18446744073709551615u,
                              {255, 255, 255, 255, 255, 255, 255, 255});
}

TEST(test_common_util_byte, from_uint32) {
  test_number_bytes<uint32_t>(0, {0, 0, 0, 0});
  test_number_bytes<uint32_t>(1024, {0, 0, 4, 0});
  test_number_bytes<uint32_t>(4294967295u, {255, 255, 255, 255});
}

TEST(test_common_util_byte, from_uint16) {
  test_number_bytes<uint16_t>(0, {0, 0});
  test_number_bytes<uint16_t>(1024, {4, 0});
  test_number_bytes<uint16_t>(65535u, {255, 255});
}

TEST(test_common_util, from_int64) {
  test_number_bytes<int64_t>(0, {0, 0, 0, 0, 0, 0, 0, 0});
  test_number_bytes<int64_t>(1024, {0, 0, 0, 0, 0, 0, 4, 0});
  test_number_bytes<int64_t>(-1024, {255, 255, 255, 255, 255, 255, 252, 0});
  test_number_bytes<int64_t>(9223372036854775807,
                             {127, 255, 255, 255, 255, 255, 255, 255});
  test_number_bytes<int64_t>(-9223372036854775808ull,
                             {128, 0, 0, 0, 0, 0, 0, 0});
}

TEST(test_common_util, from_int32) {
  test_number_bytes<int32_t>(0, {0, 0, 0, 0});
  test_number_bytes<int32_t>(1024, {0, 0, 4, 0});
  test_number_bytes<int32_t>(-1024, {255, 255, 252, 0});
  test_number_bytes<int32_t>(2147483647, {127, 255, 255, 255});
  test_number_bytes<int32_t>(-2147483648, {128, 0, 0, 0});
}
TEST(test_common_util, from_int16) {
  test_number_bytes<int16_t>(0, {0, 0});
  test_number_bytes<int16_t>(1024, {4, 0});
  test_number_bytes<int16_t>(-1024, {252, 0});
  test_number_bytes<int16_t>(32767, {127, 255});
  test_number_bytes<int16_t>(-32768, {128, 0});
}

TEST(test_common_util_byte, to_string) {
  neb::bytes bytes({72, 101, 108, 108, 111, 44, 32, 119, 111, 114, 108, 100});
  std::string result = neb::byte_to_string(bytes);
  std::string want("Hello, world");
  EXPECT_EQ(result, want);

  neb::bytes bs = neb::string_to_byte(result);
  EXPECT_EQ(bs, bytes);
}

TEST(test_common_util_byte, test_default) {
  neb::fix_bytes<> fb;

  std::string base58 = fb.to_base58();

  EXPECT_EQ(base58, "11111111111111111111111111111111");
}

TEST(test_common_util_byte, fix_bytes_to_base58) {
  neb::fix_bytes<6> fb({32, 119, 111, 114, 108, 100});

  std::string result = fb.to_base58();
  neb::fix_bytes<6> tb = neb::fix_bytes<6>::from_base58(result);

  EXPECT_EQ(fb, tb);
}

TEST(test_common_util_byte, fix_bytes_to_hex) {
  neb::fix_bytes<6> fb({132, 11, 111, 104, 18, 100});

  std::string result = fb.to_hex();
  neb::fix_bytes<6> tb = neb::fix_bytes<6>::from_hex(result);
  EXPECT_EQ(fb, tb);

  auto tf = [](const std::string &hexstring,
               const std::initializer_list<neb::byte_t> &hexbytes) {
    neb::bytes bytes(hexbytes);
    EXPECT_EQ(bytes.to_hex(), hexstring);
    EXPECT_EQ(bytes, neb::bytes::from_hex(hexstring));
  };
  tf("a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a",
     {167, 255, 198, 248, 191, 30,  215, 102, 81, 193, 71,
      86,  160, 97,  214, 98,  245, 128, 255, 77, 228, 59,
      73,  250, 130, 216, 10,  75,  128, 248, 67, 74});
  tf("3550aba97492de38af3066f0157fc532db6791b37d53262ce7688dcc5d461856",
     {53,  80, 171, 169, 116, 146, 222, 56,  175, 48,  102,
      240, 21, 127, 197, 50,  219, 103, 145, 179, 125, 83,
      38,  44, 231, 104, 141, 204, 93,  70,  24,  86});

  tf("", {});
}

template <typename T>
void test_constructor() {
  T bytes_default;

  int result = 0;
  neb::byte_t *value = bytes_default.value();

  if (nullptr == value) {
    result = 0;
  }else{
    result = (int)(*value);
  }

  int want = 0;
  EXPECT_EQ(result, want);

  T bytes_array({132, 11, 111, 104});
  SUCCEED();

  neb::byte_t buf[32] = {53,  80, 171, 169, 116, 146, 222, 56,  175, 48,  102,
                         240, 21, 127, 197, 50,  219, 103, 145, 179, 125, 83,
                         38,  44, 231, 104, 141, 204, 93,  70,  24,  86};
  T bytes_set_length(buf, 32);
  result = bytes_set_length.size();
  want = 32;
  EXPECT_EQ(result, want);

  T bytes_copy(bytes_array);
  result = bytes_copy.size();
  want = bytes_array.size();
  EXPECT_EQ(result, want);

  T bytes_right_value(std::move(bytes_set_length));
  result = bytes_right_value.size();
  want = bytes_set_length.size();
  EXPECT_EQ(result, want);

  T bytes_assignment = bytes_default;
  result = bytes_assignment.size();
  want = bytes_default.size();
  EXPECT_EQ(result, want);

  T bytes_assignment_right_value = std::move(bytes_default);
  result = bytes_assignment_right_value.size();
  want = bytes_default.size();
  EXPECT_EQ(result, want);

  EXPECT_TRUE(bytes_assignment == bytes_default);
  EXPECT_TRUE(bytes_assignment_right_value != bytes_set_length);
}

TEST(test_common_util_byte, fix_byte_constructor) {
  test_constructor<neb::fix_bytes<>>();
  test_constructor<neb::bytes>();
}

TEST(test_common_util_byte, throw_make_array) {
  EXPECT_THROW(neb::fix_bytes<> bytes(
                   {53,  80, 171, 169, 116, 146, 222, 56,  175, 48,  102,
                    240, 21, 127, 197, 50,  219, 103, 145, 179, 125, 83,
                    38,  44, 231, 104, 141, 204, 93,  70,  24,  86,  100}),
               std::out_of_range);
}

template<typename T>
void test_throw_invalid_input() {
  EXPECT_THROW(T::from_hex("102AfbGG"), std::invalid_argument);
  EXPECT_THROW(T::from_base58("wOrld"), std::invalid_argument);
}

TEST(test_common_util_byte, throw_invalid_input) {
  test_throw_invalid_input<neb::fix_bytes<>>();
  test_throw_invalid_input<neb::bytes>();
}

template <size_t N>
void test_fixed_bytes(const std::string &hexstring,
                      const std::string &base58_string) {
  neb::fix_bytes<N> fb = neb::fix_bytes<N>::from_hex(hexstring);
  std::string result = fb.to_base58();

  EXPECT_EQ(result, base58_string);

  neb::bytes bytes = neb::bytes::from_hex(hexstring);
  result = bytes.to_base58();

  EXPECT_EQ(result, base58_string);
}

TEST(test_common_util_byte, base58_encoding_decoding) {
  test_fixed_bytes<1>("61", "2g");
  test_fixed_bytes<3>("626262", "a3gV");
  test_fixed_bytes<3>("636363", "aPEr");
  test_fixed_bytes<20>("73696d706c792061206c6f6e6720737472696e67",
                       "2cFupjhnEsSn59qHXstmK2ffpLv2");
  test_fixed_bytes<25>("00eb15231dfceb60925886b67d065299925915aeb172c06647",
                       "1NS17iag9jJgTHD1VXjvLCEnZuQ3rJDE9L");
  test_fixed_bytes<5>("516b6fcd0f", "ABnLTmg");
  test_fixed_bytes<9>("bf4f89001e670274dd", "3SEo3LWLoPntC");
  test_fixed_bytes<4>("572e4794", "3EFU7m");
  test_fixed_bytes<10>("ecac89cad93923c02321", "EJDM8drfXA6uyA");
  test_fixed_bytes<4>("10c8511e", "Rt5zm");
  test_fixed_bytes<10>("00000000000000000000", "1111111111");
  test_fixed_bytes<43>(
      "000111d38e5fc9071ffcd20b4a763cc9ae4f252bb4e48fd66a835e252ada93ff480d6dd4"
      "3dc62a641155a5",
      "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz");
}

template<typename T>
void test_base64_encoding_decoding(const std::string& input) {
  T b = T::from_base64(input);
  std::string result = b.to_base64();

  EXPECT_EQ(result, input);
}

TEST(test_common_util_byte, base64_encoding_decoding) {
  std::string input(
      "TmVidWxhcyBpcyBhIG5leHQgZ2VuZXJhdGlvbiBwdWJsaWMgYmxvY2tjaGFpbiwgYWltaW5n"
      "IGZvciBhIGNvbnRpbnVvdXNseSBpbXByb3ZpbmcgZWNvc3lzdGVtLg==");
  test_base64_encoding_decoding<neb::fix_bytes<94>>(input);
  test_base64_encoding_decoding<neb::bytes>(input);
}

TEST(test_common, address) {
  std::string base58_str("n1EjkdHBDpdjFfVaeJqMQW11RhYqrrjCZR7");
  auto ret_bytes = neb::bytes::from_base58(base58_str);
  std::string s = neb::address_to_string(ret_bytes);
  std::string t = std::to_string(ret_bytes);
  EXPECT_EQ(s, t);
}
