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
#include "ff/net/network/net_buffer.h"
#include <sstream>
namespace ff {
namespace net {
net_buffer::net_buffer(int iInitSize)
    : m_oBuffer(iInitSize), m_iToWriteBufIndex(0), m_iToReadBufIndex(0) {}

void net_buffer::write_buffer(const char *pBuf, size_t len) {
  if (len >= idle())
    m_oBuffer.resize(m_iToWriteBufIndex + len + 64);
  memcpy(&m_oBuffer[m_iToWriteBufIndex], pBuf, len);
  m_iToWriteBufIndex += len;
}

size_t net_buffer::read_buffer(char *pBuf, size_t len) {
  if (filled() < len)
    len = filled();
  if (len == 0)
    return 0;
  memcpy(pBuf, &m_oBuffer[0], len);
  return len;
}

void net_buffer::erase_buffer(size_t len) {
  if (len == 0)
    return;
  if (len >= m_iToWriteBufIndex)
    len = m_iToWriteBufIndex;
  m_oBuffer.erase(m_oBuffer.begin(), m_oBuffer.begin() + len);
  m_iToReadBufIndex += len;
  if (m_iToReadBufIndex == m_iToWriteBufIndex) {
    m_iToReadBufIndex = 0;
    m_iToWriteBufIndex = 0;
  }
}

boost::asio::const_buffer net_buffer::readable() const {
  return boost::asio::const_buffer(m_oBuffer.data() + m_iToReadBufIndex,
                                   m_iToWriteBufIndex);
}
boost::asio::mutable_buffer net_buffer::writeable() {
  if (idle() < BUFFER_INC_STEP)
    m_oBuffer.resize(size() + BUFFER_INC_STEP);
  return boost::asio::mutable_buffer(m_oBuffer.data() + m_iToWriteBufIndex,
                                     idle());
}

void net_buffer::append_buffer(boost::asio::const_buffer buf) {
  write_buffer(boost::asio::buffer_cast<const char *>(buf),
               boost::asio::buffer_size(buf));
}

void net_buffer::reserve(size_t r) { m_oBuffer.reserve(r); }

void net_buffer::reserve_idle(size_t r) {
  if (idle() < r) {
    m_oBuffer.resize(filled() + r);
  }
}
std::string print_buf(const char *pBuf, size_t len) {
  std::stringstream ss;
  for (size_t i = 0; i < len; ++i) {
    uint8_t v = (uint8_t)pBuf[i];
    ss << std::hex << v / 16 << v % 16 << " ";
  }
  return ss.str();
}

} // namespace net

} // namespace ff
