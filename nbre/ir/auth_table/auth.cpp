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

  // admin address base58: n1S9RrRPC46T9byYBS868YuZgzqGuiPCY1m
  auto admin_addr = {0x19, 0x57, 0x7f, 0x94, 0x5d, 0xb3, 0x03, 0x0b, 0xfc,
                     0x3e, 0x49, 0x71, 0xa5, 0x1e, 0x54, 0x24, 0x04, 0xf0,
                     0xe6, 0x09, 0x0a, 0x75, 0xfb, 0xfd, 0xc5, 0xc0};

  // nr submit address: n1SgxtoDrExXD3B9M9hH2p1aP4JrXE2dP8q
  auto nr_submit = {0x19, 0x57, 0x85, 0x8b, 0x2c, 0x2a, 0xdc, 0x74, 0xd0,
                    0xa9, 0xd1, 0x6e, 0x79, 0x80, 0x25, 0x3c, 0x82, 0x9a,
                    0x34, 0x4b, 0xc8, 0x7a, 0x7c, 0x40, 0xd3, 0xae};

  // dip submit address: n1TBKKEmKT6AKWDiTHXZ9FzYNpZKnPGARnC
  auto dip_submit = {0x19, 0x57, 0x8a, 0xe7, 0xdd, 0xca, 0xaa, 0x65, 0x41,
                     0xa4, 0x7d, 0x0a, 0x76, 0x7c, 0xb7, 0x6d, 0x2b, 0xe8,
                     0x1d, 0xa9, 0x50, 0xec, 0x2c, 0xb7, 0xa8, 0xf5};

  std::vector<row_t> auth_table = {
      std::make_tuple("nr", std::string(nr_submit.begin(), nr_submit.end()),
                      2307000ULL, 8067000ULL),
      std::make_tuple("dip", std::string(dip_submit.begin(), dip_submit.end()),
                      2307000ULL, 8067000ULL)};
  return auth_table;
}

