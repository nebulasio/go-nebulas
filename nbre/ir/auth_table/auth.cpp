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
typedef uint64_t version_t;
typedef std::string address_t;
typedef uint64_t height_t;

typedef std::tuple<name_t, version_t, address_t, height_t, height_t> row_t;

std::vector<row_t> entry_point_auth() {

  auto to_version_t = [](uint32_t major_version, uint16_t minor_version,
                         uint16_t patch_version) -> version_t {
    return (0ULL + major_version) + ((0ULL + minor_version) << 32) +
           ((0ULL + patch_version) << 48);
  };

  auto admin_addr = {0x19, 0x57, 0x73, 0x3f, 0x7b, 0x52, 0xad, 0x3a, 0x99,
                     0xc5, 0x56, 0x84, 0xb3, 0x9f, 0x4c, 0x31, 0xe0, 0x5c,
                     0x35, 0x0c, 0x5d, 0xaa, 0xb0, 0x7e, 0x87, 0x5b};

  std::vector<row_t> auth_table = {
      std::make_tuple("nr", to_version_t(0, 0, 1),
                      std::string(admin_addr.begin(), admin_addr.end()), 1ULL,
                      200000ULL),
      std::make_tuple("nr", to_version_t(0, 0, 2),
                      std::string(admin_addr.begin(), admin_addr.end()), 1ULL,
                      200000ULL),
      std::make_tuple("nr", to_version_t(0, 0, 3),
                      std::string(admin_addr.begin(), admin_addr.end()), 1ULL,
                      200000ULL),
      std::make_tuple("dip", to_version_t(0, 0, 1),
                      std::string(admin_addr.begin(), admin_addr.end()), 1ULL,
                      200000ULL),
      std::make_tuple("dip", to_version_t(0, 0, 2),
                      std::string(admin_addr.begin(), admin_addr.end()), 1ULL,
                      200000ULL),
      std::make_tuple("dip", to_version_t(0, 0, 3),
                      std::string(admin_addr.begin(), admin_addr.end()), 1ULL,
                      200000ULL)};
  return auth_table;
}

