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

#pragma once

#include "common/math/softfloat.hpp"
#include "common/optional.h"
#include <algorithm>
#include <array>
#include <boost/multiprecision/cpp_int.hpp>
#include <boost/thread/shared_lock_guard.hpp>
#include <boost/thread/shared_mutex.hpp>
#include <cstdint>
#include <ff/network.h>
#include <glog/logging.h>
#include <iostream>
#include <memory>
#include <mutex>
#include <queue>
#include <string>
#include <thread>
#include <tuple>
#include <unordered_map>
#include <unordered_set>
#include <vector>

namespace neb {

typedef std::string hex_hash_t;
typedef uint8_t byte_t;
typedef uint64_t block_height_t;
extern std::string program_name;

typedef uint64_t version_t;
typedef boost::multiprecision::int128_t int128_t;
typedef boost::multiprecision::uint128_t uint128_t;
typedef int128_t wei_t;

typedef float32 floatxx_t;

constexpr int32_t tx_status_fail = 0;
constexpr int32_t tx_status_succ = 1;
constexpr int32_t tx_status_special = 2; // only in the genesis block

namespace tcolor {
const static char *red = "\033[1;31m";
const static char *green = "\033[1;32m";
const static char *yellow = "\033[1;33m";
const static char *blue = "\033[1;34m";
const static char *magenta = "\033[1;35m";
const static char *cyan = "\033[1;36m";
const static char *reset = "\033[0m";
}

namespace ir_type {
const static std::string invalid = "invalid";
const static std::string llvm = "llvm";
const static std::string cpp = "cpp";
};
}

namespace std {
std::string to_string(const neb::floatxx_t &v);
}

namespace ff {
namespace net {
template <> class udt_marshaler<neb::floatxx_t> {
public:
  static size_t seralize(char *buf, neb::floatxx_t &v) {
    typename neb::floatxx_t::value_type t = v.value();
    std::memcpy(buf, &t, sizeof(t));
    return sizeof(t);
  }
  static size_t deseralize(const char *buf, size_t len, neb::floatxx_t &v) {
    typename neb::floatxx_t::value_type t;
    std::memcpy((char *)&t, buf, sizeof(t));
    v = neb::floatxx_t(t);
    return sizeof(t);
  }
  static size_t length(neb::floatxx_t &) {
    return sizeof(typename neb::floatxx_t::value_type);
  }
};

} // namespace net
} // namespace ff
