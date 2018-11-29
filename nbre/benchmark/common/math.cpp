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
#include <cmath>
#include <iostream>

BENCHMARK(math_benchamrk_plus, softfloat_plus) {
  // exp
  auto e1 = neb::math::exp(float64(1));
  auto e2 = neb::math::exp(float64(2));
  for (int i = -9000; i < 9000; ++i) {
    e1 = e1 + e2;
  }
}
BENCHMARK(math_benchamrk_plus, std_plus) {
  // exp
  auto e1 = std::exp(1);
  auto e2 = std::exp(2);
  for (int i = -9000; i < 9000; ++i) {
    e1 = e1 + e2;
  }
}
BENCHMARK(math_benchamrk_plus1, softfloat_plus1) {
  // exp
  auto e1 = neb::math::exp(float64(1));
  auto e2 = neb::math::exp(float64(2));
  for (int i = -9000; i < 9000; ++i) {
    e1 = e1 - e2;
  }
}
BENCHMARK(math_benchamrk_plus1, std_plus1) {
  // exp
  auto e1 = std::exp(1);
  auto e2 = std::exp(2);
  for (int i = -9000; i < 9000; ++i) {
    e1 = e1 - e2;
  }
}

BENCHMARK(math_benchamrk_mul1, softfloat_mul1) {
  // exp
  auto e1 = neb::math::exp(float64(1));
  auto e2 = neb::math::exp(float64(2));
  for (int i = -9000; i < 9000; ++i) {
    e1 = e1 * e2;
  }
}

BENCHMARK(math_benchamrk_mul1, std_mul1) {
  // exp
  auto e1 = std::exp(1);
  auto e2 = std::exp(2);
  for (int i = -9000; i < 9000; ++i) {
    e1 = e1 * e2;
  }
}

BENCHMARK(math_benchamrk_div1, softfloat_div1) {
  // exp
  auto e1 = neb::math::exp(float64(1));
  auto e2 = neb::math::exp(float64(2));
  for (int i = -9000; i < 9000; ++i) {
    e1 = e1 / e2;
  }
}
BENCHMARK(math_benchamrk_div1, std_div1) {
  // exp
  auto e1 = std::exp(1);
  auto e2 = std::exp(2);
  for (int i = -9000; i < 9000; ++i) {
    e1 = e1 / e2;
  }
}

BENCHMARK(math_benchamrk_exp, softfloat_exp) {
  // exp
  for (int i = -9000; i < 9000; ++i) {
    auto e = neb::math::exp(float64(10));
    std::cout << "neb::exp(" << 10 << "): " << e << std::endl;
  }
}
BENCHMARK(math_benchamrk_exp, std_exp) {
  // exp
  for (int i = -9000; i < 9000; ++i) {
    auto e = std::exp(10);
    std::cout << "std::exp(" << 10 << "): " << e << std::endl;
  }
}

BENCHMARK(math_benchamrk_arctan, softfloat_arctan) {
  // arctan
  for (int i = -9000; i < 9000; ++i) {
    auto pi = neb::math::arctan(float64(10));
    std::cout << "neb::arctan(" << 10 << "): " << pi << std::endl;
  }
}
BENCHMARK(math_benchamrk_arctan, std_arctan) {
  // arctan
  for (int i = -9000; i < 9000; ++i) {
    auto pi = std::atan(10);
    std::cout << "std::atan(" << 10 << "): " << pi << std::endl;
  }
}

BENCHMARK(math_benchamrk_sin, softfloat_sin) {
  // sin
  for (int i = -9000; i < 9000; i++) {
    auto s = neb::math::sin(float64(10));
    std::cout << "neb::sin(" << 10 << "): " << s << std::endl;
  }
}
BENCHMARK(math_benchamrk_sin, std_sin) {
  // sin
  for (int i = -9000; i < 9000; i++) {
    auto s = std::sin(10);
    std::cout << "std::sin(" << 10 << "): " << s << std::endl;
  }
}

BENCHMARK(math_benchamrk_ln, softfloat_ln) {
  // ln
  for (int i = 1; i < 10000; ++i) {
    auto l = neb::math::ln(float64(10));
    std::cout << "neb::ln(" << 10 << "): " << l << std::endl;
  }
}
BENCHMARK(math_benchamrk_ln, std_ln) {
  // ln
  for (int i = 1; i < 10000; ++i) {
    auto l = std::log(10);
    std::cout << "std::log(" << 10 << "): " << l << std::endl;
  }
}

BENCHMARK(math_benchamrk_log2, softfloat_log2) {
  // log2
  for (int i = 1; i < 10000; ++i) {
    auto l2 = neb::math::log2(float64(10));
    std::cout << "neb::log2(" << 10 << "): " << l2 << std::endl;
  }
}
BENCHMARK(math_benchamrk_log2, std_log2) {
  // log2
  for (int i = 1; i < 10000; ++i) {
    auto l2 = std::log2(10);
    std::cout << "std::log2(" << 10 << "): " << l2 << std::endl;
  }
}

BENCHMARK(math_benchamrk_pow, softfloat_pow) {
  // pow
  auto e = neb::math::constants<float64>::e();
  for (int i = -90; i < 90; ++i) {
    auto p = neb::math::pow(e, float64(10));
    std::cout << "neb::pow(e, " << 10 << "): " << p << std::endl;
  }
}
BENCHMARK(math_benchamrk_pow, std_pow) {
  // pow
  auto e = std::exp(1);
  for (int i = -9000; i < 9000; ++i) {
    auto p = std::pow(e, 10);
    std::cout << "std::pow(e, " << 10 << "): " << p << std::endl;
  }
}
