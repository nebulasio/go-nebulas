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

#include "runtime/dip/dip_impl.h"
#include "runtime/dip/dip_reward.h"
#include <vector>

int main(int argc, char *argv[]) {

  using dip_info_t = neb::rt::dip::dip_info_t;
  using floatxx_t = neb::floatxx_t;
  std::vector<dip_info_t> v;
  v.push_back(dip_info_t{"addr1", floatxx_t(1.1)});
  v.push_back(dip_info_t{"addr2", floatxx_t(2.1)});

  std::cout << neb::rt::dip::dip_reward::dip_info_to_json(v) << std::endl;
  return 0;
}
