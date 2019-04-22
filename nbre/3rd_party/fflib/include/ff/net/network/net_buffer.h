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
#pragma once
#include "ff/net/common/common.h"

namespace ff {
namespace net {

#define BUFFER_INC_STEP 256
class net_buffer {
public:
  net_buffer(int iInitSize = BUFFER_INC_STEP);

  void write_buffer(const char *pBuf, size_t len);

  size_t read_buffer(char *pBuf, size_t len);

  void erase_buffer(size_t len);

  boost::asio::const_buffer readable() const;

  boost::asio::mutable_buffer writeable();

  void append_buffer(boost::asio::const_buffer buf);
  void reserve(size_t r);
  void reserve_idle(size_t r);
  inline const char *buffer() const { return &m_oBuffer[m_iToWriteBufIndex]; }
  inline size_t length() const {
    return m_iToWriteBufIndex - m_iToReadBufIndex;
  }
  inline size_t size() const { return m_oBuffer.size(); }
  inline void resize(size_t s) { m_oBuffer.resize(s); }
  inline size_t capacity() const { return m_oBuffer.capacity(); }
  inline const size_t &filled() const { return m_iToWriteBufIndex; }
  inline size_t &filled() { return m_iToWriteBufIndex; }
  inline const size_t &read() const { return m_iToReadBufIndex; }
  inline size_t &read() { return m_iToReadBufIndex; }
  inline size_t idle() const { return size() - filled(); }

protected:
  std::vector<char> m_oBuffer;
  size_t m_iToWriteBufIndex;
  size_t m_iToReadBufIndex;
}; // end class NetBuffer

std::string print_buf(const char *pBuf, size_t len);
}
} // end namespace ff
