
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

#include "common/util/util.h"

namespace neb {

std::shared_ptr<std::vector<std::string>>
string_util::split_by_comma(const std::string &s, char comma) {

  std::vector<std::string> v;
  std::stringstream ss(s);
  std::string token;

  while (getline(ss, token, comma)) {
    v.push_back(token);
  }
  return std::make_shared<std::vector<std::string>>(v);
}
} // namespace neb
