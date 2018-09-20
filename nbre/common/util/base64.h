#pragma once
#include "common/common.h"

namespace neb {
bool encode_base64(const std::string &input, std::string &output);
bool decode_base64(const std::string &input, std::string &output);
} // namespace neb
