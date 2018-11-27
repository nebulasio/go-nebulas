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
#include "common/math.h"
#include "benchmark/benchmark_instances.h"
#include <iostream>

BENCHMARK(math_benchamrk, exp_benchmark) {
  // exp
  for (int i = -9; i < 9; ++i) {
    auto e = neb::math::exp(float64(i));
    std::cout << "neb::exp(" << i << "): " << e << std::endl;
  }
}

BENCHMARK(math_benchamrk, arctan_benchmark) {
  // arctan
  for (int i = -9; i < 9; ++i) {
    auto pi = neb::math::arctan(float64(i));
    std::cout << "neb::arctan(" << i << "): " << pi << std::endl;
  }
}

BENCHMARK(math_benchamrk, sin_benchmark) {
  // sin
  for (int i = -9; i < 9; i++) {
    auto s = neb::math::sin(float64(i));
    std::cout << "neb::sin(" << i << "): " << s << std::endl;
  }
}

BENCHMARK(math_benchamrk, ln_benchmark) {
  // ln
  for (int i = 1; i < 10; ++i) {
    auto l = neb::math::ln(float64(i));
    std::cout << "neb::ln(" << i << "): " << l << std::endl;
  }
}

BENCHMARK(math_benchamrk, log2_benchmark) {
  // log2
  for (int i = 1; i < 10; ++i) {
    auto l2 = neb::math::log2(float64(i));
    std::cout << "neb::log2(" << i << "): " << l2 << std::endl;
  }
}

BENCHMARK(math_benchamrk, pow_benchmark) {
  // pow
  auto e = neb::math::constants<float64>::e();
  for (int i = -9; i < 9; ++i) {
    auto p = neb::math::pow(e, float64(i));
    std::cout << "neb::pow(e, " << i << "): " << p << std::endl;
  }
}
