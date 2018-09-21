#include "common/util/base64.h"
#include <boost/archive/iterators/base64_from_binary.hpp>
#include <boost/archive/iterators/binary_from_base64.hpp>
#include <boost/archive/iterators/transform_width.hpp>
#include <boost/beast/core/detail/base64.hpp>

namespace neb {

std::string encode_base64(const std::string &input) {
  return ::boost::beast::detail::base64_encode(input);
}

std::string encode_base64(const unsigned char *pbegin,
                          const unsigned char *pend) {
  return ::boost::beast::detail::base64_encode(pbegin, pend - pbegin);
}

bool decode_base64(const std::string &input, std::string &output) {
  output = ::boost::beast::detail::base64_decode(input);
  return output.empty() == false;
}
} // namespace neb

