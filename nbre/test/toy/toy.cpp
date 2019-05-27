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

#include "common/byte.h"
#include "common/common.h"
#include "common/math.h"
#include <boost/program_options.hpp>
#include <iomanip>
#include <sstream>

namespace po = boost::program_options;

template <typename T> unsigned char *to_mem_bytes(T &x) {
  auto ret = reinterpret_cast<unsigned char *>(&x);
  std::stringstream ss;
  for (size_t i = 0; i < sizeof(x); i++) {
    ss << std::hex << std::setw(2) << std::setfill('0')
       << static_cast<unsigned int>(ret[i]);
  }
  std::cout << ss.str() << std::endl;
  return ret;
}

template <typename T> T from_mem_bytes(unsigned char *buf) {
  return *reinterpret_cast<T *>(buf);
}

int main(int argc, char *argv[]) {

  po::options_description desc("Address base58 to string bytes");
  desc.add_options()("help", "show help message")(
      "address", po::value<std::string>(), "address base58 encoding");

  po::variables_map vm;
  po::store(po::parse_command_line(argc, argv, desc), vm);
  po::notify(vm);

  if (vm.count("help")) {
    std::cout << desc << "\n";
    return 1;
  }

  neb::floatxx_t zero =
      softfloat_cast<uint32_t, typename neb::floatxx_t::value_type>(0);
  neb::floatxx_t one =
      softfloat_cast<uint32_t, typename neb::floatxx_t::value_type>(1);

  neb::floatxx_t f_xx = zero;
  auto ret = to_mem_bytes(f_xx);
  unsigned char buf[] = {0x0, 0x0, 0x0, 0x0};
  // std::cout << std::memcmp(ret, buf, sizeof(buf)) << std::endl;

  f_xx = neb::math::sqrt(f_xx);
  ret = to_mem_bytes(f_xx);

  f_xx = from_mem_bytes<neb::floatxx_t>(ret);
  std::cout << f_xx << std::endl;

  return 0;
}
