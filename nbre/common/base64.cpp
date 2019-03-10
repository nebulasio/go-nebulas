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
#include "common/base64.h"
#include <boost/archive/iterators/base64_from_binary.hpp>
#include <boost/archive/iterators/binary_from_base64.hpp>
#include <boost/archive/iterators/transform_width.hpp>
#include <boost/beast/core/detail/base64.hpp>

namespace neb {

std::string encode_base64(const std::string &input) {
  return ::boost::beast::detail::base64_encode(input);
}

std::string encode_base64(const unsigned char *pbegin,
                          const unsigned char *pend) {
  return ::boost::beast::detail::base64_encode(pbegin, pend - pbegin);
}

bool decode_base64(const std::string &input, std::string &output) {
  output = ::boost::beast::detail::base64_decode(input);
  return output.empty() == false;
}
} // namespace neb

