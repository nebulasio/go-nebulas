#include "common/util/base64.h"
#include <boost/archive/iterators/base64_from_binary.hpp>
#include <boost/archive/iterators/binary_from_base64.hpp>
#include <boost/archive/iterators/transform_width.hpp>

namespace neb {

bool encode_base64(const std::string &input, std::string &output) {
  typedef boost::archive::iterators::base64_from_binary<
      boost::archive::iterators::transform_width<std::string::const_iterator, 6,
                                                 8>>
      Base64EncodeIterator;

  std::stringstream result;

  std::copy(Base64EncodeIterator(input.begin()), Base64EncodeIterator(input.end()),
       std::ostream_iterator<char>(result));

  output = result.str();

  return output.empty() == false;
}

bool decode_base64(const std::string &input, std::string &output) {
  // using namespace boost::archive::iterators;
  typedef boost::archive::iterators::transform_width<
      boost::archive::iterators::binary_from_base64<
          std::string::const_iterator>,
      8, 6>
      Base64DecodeIterator;
  std::stringstream result;

  try {
    copy(Base64DecodeIterator(input.begin()), Base64DecodeIterator(input.end()),
         std::ostream_iterator<char>(result));
  } catch (...) {
    return false;
  }

  output = result.str();

  return output.empty() == false;
}
} // namespace neb

