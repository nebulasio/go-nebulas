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

#include "common/common.h"
#include "common/math.h"
#include <limits>
#include <random>

void math_exp(std::mt19937 &mt, std::uniform_int_distribution<> &dis) {
  neb::floatxx_t one =
      softfloat_cast<uint32_t, typename neb::floatxx_t::value_type>(1);

  while (true) {
    int32_t r = dis(mt);
    neb::floatxx_t x = r;
    double dx = r;
    // std::cout << x << ',' << neb::math::exp(x) << ',' << std::exp(dx)
    //<< std::endl;
    std::cout << (x / (one + neb::math::exp(x))) << ','
              << (dx / (1 + std::exp(dx))) << std::endl;
    std::this_thread::sleep_for(std::chrono::milliseconds(200));

    neb::floatxx_t y = 1 / x;
    double dy = 1 / dx;
    // std::cout << y << ',' << neb::math::exp(y) << ',' << std::exp(dy)
    //<< std::endl;
    std::cout << (y / (one + neb::math::exp(y))) << ','
              << (dy / (1 + std::exp(dy))) << std::endl;
    std::this_thread::sleep_for(std::chrono::milliseconds(200));
  }
}

double result_double(int64_t a, int64_t b, int64_t c, int64_t d, double theta,
                     double mu, double lambda, double S, double in_val,
                     double out_val) {

  auto f_weight = [](double in_val, double out_val) {
    double pi = std::acos(-1);
    double atan_val = (in_val == 0 ? pi / 2 : std::atan(out_val / in_val));
    auto ret =
        (in_val + out_val) * std::exp((-2) * std::sin(pi / 4.0 - atan_val) *
                                      std::sin(pi / 4.0 - atan_val));
    return ret;
  };
  double R = f_weight(in_val, out_val);

  auto gamma = std::pow(theta * R / (R + mu), lambda);
  auto ret =
      (S / (1 + std::exp(a + b * S))) * (gamma / (1 + std::exp(c + d * gamma)));
  return ret;
}

neb::floatxx_t result_softfloat(int64_t a, int64_t b, int64_t c, int64_t d,
                                neb::floatxx_t theta, neb::floatxx_t mu,
                                neb::floatxx_t lambda, neb::floatxx_t S,
                                neb::floatxx_t in_val, neb::floatxx_t out_val) {

  auto f_weight = [](neb::floatxx_t in_val, neb::floatxx_t out_val) {
    neb::floatxx_t pi = neb::math::constants<neb::floatxx_t>::pi();
    neb::floatxx_t atan_val =
        (in_val == 0 ? pi / 2 : neb::math::arctan(out_val / in_val));
    auto ret = (in_val + out_val) *
               neb::math::exp((-2) * neb::math::sin(pi / 4.0 - atan_val) *
                              neb::math::sin(pi / 4.0 - atan_val));
    return ret;
  };
  neb::floatxx_t R = f_weight(in_val, out_val);

  neb::floatxx_t one =
      softfloat_cast<uint32_t, typename neb::floatxx_t::value_type>(1);
  auto gamma = neb::math::pow(theta * R / (R + mu), lambda);
  auto ret = (S / (one + neb::math::exp(a + b * S))) *
             (gamma / (one + neb::math::exp(c + d * gamma)));
  return ret;
}

void result_cmp(std::mt19937 &mt, std::uniform_int_distribution<> &dis) {

  int64_t a = 3000;
  int64_t b = -1;
  int64_t c = 6;
  int64_t d = -9;
  neb::floatxx_t one =
      softfloat_cast<uint32_t, typename neb::floatxx_t::value_type>(1);
  neb::floatxx_t ten =
      softfloat_cast<uint32_t, typename neb::floatxx_t::value_type>(10);
  int32_t S = std::abs(dis(mt));
  int32_t in_val = std::abs(dis(mt));
  int32_t out_val = std::abs(dis(mt));

  auto rs = result_softfloat(a, b, c, d, one, one, one / ten, neb::floatxx_t(S),
                             neb::floatxx_t(in_val), neb::floatxx_t(out_val));
  auto rd = result_double(a, b, c, d, 1, 1, 0.1, S, in_val, out_val);
  std::cout << rs << ',' << rd << ',' << (rs - rd) << std::endl;
}

int main(int argc, char *argv[]) {
  std::random_device rd;
  std::mt19937 mt(rd());
  std::uniform_int_distribution<> dis(0, std::numeric_limits<int32_t>::max());
  math_exp(mt, dis);
  result_cmp(mt, dis);
  return 0;
}
