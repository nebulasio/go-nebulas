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

#include <tuple>

typedef const char *name_t;
typedef uint64_t version_t;
typedef const char *address_t;
typedef uint64_t height_t;

typedef std::tuple<name_t, version_t, address_t, height_t, height_t> row_t;

std::tuple<row_t *, size_t> entry_point_auth() {
  static row_t r[] = {std::make_tuple("nr", 1, "addr1", 100, 200),
                      std::make_tuple("nr", 2, "addr2", 150, 250),
                      std::make_tuple("dip", 1, "addr1", 200, 300)};
  return std::make_tuple(r, sizeof(r) / sizeof(r[0]));
}

