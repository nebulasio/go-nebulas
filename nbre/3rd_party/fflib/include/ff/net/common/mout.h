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
#include <sstream>

namespace ff {
namespace net {
class mout_stream {
public:
  class mout_internal_stream {
  public:
    typedef std::basic_ostream<char, std::char_traits<char>> ostream_t;

    mout_internal_stream();

    ~mout_internal_stream();
    template <class T> mout_internal_stream &operator<<(const T &t) {
      (*m_ss) << t;
      return *this;
    }
    inline mout_internal_stream &operator<<(ostream_t &(*pfn)(ostream_t &)) {
      (*m_ss) << pfn;
      return *this;
    }

  protected:
    std::stringstream *m_ss;
    static std::mutex s_out_mutex;
  }; // end class mout_internal_stream
  template <class T> mout_internal_stream operator<<(const T &t) {
    mout_internal_stream tm;
    tm << t;
    return tm;
  }
};
extern mout_stream mout;
} // namespace net
} // namespace ff
