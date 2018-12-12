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

  auto admin_addr = {0x19, 0x57, 0x8a, 0x4d, 0x92, 0x74, 0xbd, 0x92, 0xd6,
                     0x83, 0xc7, 0x11, 0x7e, 0x0c, 0xc2, 0xc1, 0xec, 0x3a,
                     0x14, 0x04, 0xf4, 0x71, 0x3a, 0x9e, 0x0c, 0xbd};

  std::vector<row_t> auth_table = {
      std::make_tuple("nr", to_version_t(0, 0, 1),
                      std::string(admin_addr.begin(), admin_addr.end()), 1ULL,
                      200000ULL),
      std::make_tuple("nr", to_version_t(0, 1, 0),
                      std::string(admin_addr.begin(), admin_addr.end()), 1ULL,
                      200000ULL),
      std::make_tuple("dip", to_version_t(0, 0, 1),
                      std::string(admin_addr.begin(), admin_addr.end()), 1ULL,
                      200000ULL)};
  return auth_table;
}

