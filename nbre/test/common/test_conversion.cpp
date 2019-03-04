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

#include "common/util/conversion.h"

int main() {
  neb::conversion cv;
  uint64_t x = 1ULL << 63;
  neb::floatxx_t f(x);

  f = f * 2;
  cv.from_float(f);
  std::cout << f << ',';
  std::cout << cv.data() << ',' << cv.high() << ',' << cv.low() << std::endl;

  f = f + x;
  cv.from_float(f);
  std::cout << f << ',';
  std::cout << cv.data() << ',' << cv.high() << ',' << cv.low() << std::endl;
  return 0;
}
