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
#include "ff/net/common/archive.h"

namespace ff {
namespace net {

marshaler::marshaler(const char *buf, size_t len, marshaler_type at)
    : m_iAT(at), m_iBase(0), m_pWriteBuf(NULL), m_pReadBuf(NULL),
      m_iBufLen(len) {
  if (m_iAT == deseralizer) {
    m_pReadBuf = buf;
  } else if (m_iAT == seralizer) {
    assert(0);
  }
}

marshaler::marshaler(char *buf, size_t len, marshaler_type at)
    : m_iAT(at), m_iBase(0), m_pWriteBuf(buf), m_pReadBuf(NULL),
      m_iBufLen(len) {
  if (m_iAT == deseralizer) {
    m_pReadBuf = buf;
  } else if (m_iAT == seralizer) {
    m_pWriteBuf = buf;
  }
}

marshaler::marshaler(marshaler_type at)
    : m_iAT(at), m_iBase(0), m_pWriteBuf(NULL), m_pReadBuf(NULL), m_iBufLen(0) {
  assert(m_iAT == length_retriver);
}

void marshaler::archive(std::string &s) {
  size_t len = s.size();
  switch (get_marshaler_type()) {
  case seralizer:
    len = s.size();
    std::memcpy(m_pWriteBuf + m_iBase, (const char *)&len, sizeof(size_t));
    m_iBase += sizeof(size_t);
    std::memcpy(m_pWriteBuf + m_iBase, s.c_str(), len);
    m_iBase += len;
    break;
  case deseralizer:
    std::memcpy((char *)&len, m_pReadBuf + m_iBase, sizeof(size_t));
    m_iBase += sizeof(size_t);
    s = std::string(m_pReadBuf + m_iBase, len);
    m_iBase += len;
    break;
  case length_retriver:
    m_iBase += (sizeof(size_t) + s.size());
    break;
  }
}
} // namespace net
} // namespace ff
