#pragma once
#include "common/common.h"

namespace neb {
std::string encode_base64(const std::string &input);
std::string encode_base64(const unsigned char *pbegin,
                          const unsigned char *pend);

bool decode_base64(const std::string &input, std::string &output);
} // namespace neb
