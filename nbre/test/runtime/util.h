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

#include "common/common.h"

static std::vector<std::string> split_by_comma(const std::string &str,
                                               char comma) {
  std::vector<std::string> v;
  std::stringstream ss(str);
  std::string token;

  while (getline(ss, token, comma)) {
    v.push_back(token);
  }
  return v;
}

template <typename T> std::string mem_bytes(T x) {
  auto buf = reinterpret_cast<unsigned char *>(&x);
  std::stringstream ss;
  for (auto i = 0; i < sizeof(x); i++) {
    ss << std::hex << std::setw(2) << std::setfill('0')
       << static_cast<unsigned int>(buf[i]);
  }
  return ss.str();
}

