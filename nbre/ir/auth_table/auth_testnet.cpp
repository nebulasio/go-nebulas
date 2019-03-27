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

#include <string>
#include <tuple>
#include <vector>

typedef std::string name_t;
typedef std::string address_t;
typedef uint64_t height_t;

typedef std::tuple<name_t, address_t, height_t, height_t> row_t;

std::vector<row_t> entry_point_auth() {

  // testnet admin address: n1UodK5h3o7yHFLHe9Vq4N3WZGUthsWm6j7
  auto admin_addr = {0x19, 0x57, 0x9c, 0xbd, 0xfd, 0x7c, 0x04, 0xad, 0x91,
                     0x44, 0x41, 0x5f, 0x48, 0xe1, 0xe5, 0x08, 0x54, 0x0d,
                     0xb6, 0x99, 0x61, 0x70, 0xae, 0xd4, 0x78, 0xae};

  std::vector<row_t> auth_table = {
      std::make_tuple("nr", std::string(admin_addr.begin(), admin_addr.end()),
                      1562800ULL, 3000000ULL),
      std::make_tuple("dip", std::string(admin_addr.begin(), admin_addr.end()),
                      1562800ULL, 3000000ULL)};
  return auth_table;
}

