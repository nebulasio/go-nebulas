/***********************************************
  The MIT License (MIT)

  Copyright (c) 2012 Athrun Arthur <athrunarthur@gmail.com>

  Permission is hereby granted, free of charge, to any person obtaining a copy
  of this software and associated documentation files (the "Software"), to deal
  in the Software without restriction, including without limitation the rights
  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
  copies of the Software, and to permit persons to whom the Software is
  furnished to do so, subject to the following conditions:

  The above copyright notice and this permission notice shall be included in
  all copies or substantial portions of the Software.

  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
  THE SOFTWARE.
 *************************************************/
#include "ff/net/middleware/pkg_packer.h"
#include "ff/net/network/net_buffer.h"

namespace ff {
namespace net {
length_packer::~length_packer(){};

void length_packer::pack(net_buffer &oSendBuffer, const char *pBuf,
                         size_t len) {
  uint32_t buflen = static_cast<uint32_t>(len);
  oSendBuffer.write_buffer((char *)&buflen, sizeof(buflen));
  oSendBuffer.write_buffer(pBuf, buflen);
}

void length_packer::pack(net_buffer &oSendBuffer, const package_ptr &pkg) {
  marshaler lr(marshaler::length_retriver);

  pkg->arch(lr);
  uint32_t len = lr.get_length();

  oSendBuffer.reserve_idle(static_cast<size_t>(len + sizeof(len)));
  char *pBuf = boost::asio::buffer_cast<char *>(oSendBuffer.writeable());
  seralize(len, pBuf);
  oSendBuffer.filled() += sizeof(len);
  pBuf = boost::asio::buffer_cast<char *>(oSendBuffer.writeable());
  marshaler s(pBuf, oSendBuffer.idle(), marshaler::seralizer);
  pkg->arch(s);
  oSendBuffer.filled() += len;

  // LOG(INFO) << "length_packer::bond(), seralize pkg: "
  //<< print_buf(boost::asio::buffer_cast<const char *>(
  // oSendBuffer.readable()),
  // oSendBuffer.filled());
}

std::list<shared_buffer> length_packer::split(net_buffer &oRecvBuffer) {
  std::list<shared_buffer> resPkgs;
  if (oRecvBuffer.length() <= 4)
    return resPkgs;

  uint32_t len;
  const char *pBuf =
      boost::asio::buffer_cast<const char *>(oRecvBuffer.readable());
  uint32_t buf_len = oRecvBuffer.length();
  size_t bi = 0;
  deseralize(pBuf + bi, len);

  while (buf_len - bi >= sizeof(len) && buf_len - bi - sizeof(len) >= len) {
    // LOG(INFO) << "length_packer::split() buffer is "
    //<< print_buf(boost::asio::buffer_cast<const char *>(
    // oRecvBuffer.readable()),
    // oRecvBuffer.filled());
    // LOG(INFO) << "length_packer::split() " << "split pkg with len:" << len;

    bi += sizeof(len);
    shared_buffer sb(pBuf + bi, len);
    resPkgs.push_back(sb);
    bi += len;

    deseralize(pBuf + bi, len);
  }
  oRecvBuffer.read() = bi;
  if (oRecvBuffer.read() == oRecvBuffer.filled()) {
    oRecvBuffer.read() = 0;
    oRecvBuffer.filled() = 0;
  }
  return resPkgs;
}
} // namespace net
} // namespace ff
