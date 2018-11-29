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

#include "fs/blockchain/nebulas_currency.h"

neb::fs::nas operator"" _nas(long double v) { return neb::fs::nas(v); }
neb::fs::nas operator"" _nas(const char *s) {
  return neb::fs::nas(std::atoi(s));
}

neb::fs::wei operator"" _wei(long double v) { return neb::fs::wei(v); }
neb::fs::wei operator"" _wei(const char *s) {
  return neb::fs::wei(std::atoi(s));
}

std::ostream &operator<<(std::ostream &os, const neb::fs::nas &obj) {
  os << obj.value() << "nas";
  return os;
}

std::ostream &operator<<(std::ostream &os, const neb::fs::wei &obj) {
  os << obj.value() << "wei";
  return os;
}
