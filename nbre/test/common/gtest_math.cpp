#include "common/math.h"
#include <gtest/gtest.h>

TEST(test_common_math, simple) {
  float64 e = neb::math::constants<float64>::e();
  std::cout << "e: " << e << std::endl;
  float64 pi = neb::math::constants<float64>::pi();
  std::cout << "pi: " << pi << std::endl;
  float64 ln2 = neb::math::constants<float64>::ln2();
  std::cout << "ln2: " << ln2 << std::endl;

  auto t = neb::math::exp(float64(2));
  std::cout << "e^2: " << t << std::endl;

  float64 ie = e.integer_val();
  float64 ipi = pi.integer_val();
  float64 iln2 = ln2.integer_val();
  std::cout << "ie: " << ie << std::endl;
  std::cout << "ipi: " << ipi << std::endl;
  std::cout << "iln2: " << iln2 << std::endl;
  std::cout << "xxxxxxxxxxxxx" << std::endl;

  auto epi = neb::math::exp(pi);
  auto pi_4 = neb::math::arctan(float64(1));
  auto sin = neb::math::sin(pi / 4);

  auto lne = neb::math::ln(e);
  std::cout << "e^pi: " << epi << std::endl;
  std::cout << "pi/4: " << pi_4 << std::endl;
  std::cout << "sin(pi/4): " << sin << std::endl;
  std::cout << "ln(e): " << lne << std::endl;
}
