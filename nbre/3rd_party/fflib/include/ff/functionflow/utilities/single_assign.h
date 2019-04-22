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
#ifndef FF_UTILITIES_SINGLE_ASSIGN_H_
#define FF_UTILITIES_SINGLE_ASSIGN_H_
#include "ff/functionflow/common/common.h"
#include "ff/functionflow/runtime/rtcmn.h"
#include <mutex>
#include <vector>

namespace ff {
template <class T>
class single_assign {
  single_assign(const single_assign<T> &) = delete;
  single_assign<T> &operator=(const single_assign<T> &) = delete;

 public:
  single_assign() : m_oValue(), m_bIsAssigned(false) {}
  single_assign(const T &v) : m_oValue(v), m_bIsAssigned(true) {}

  single_assign<T> &operator=(const T &v) {
    if (m_bIsAssigned) return *this;
    m_bIsAssigned = true;
    m_oValue = v;
    return *this;
  }

  T &get() { return m_oValue; }

 protected:
  T m_oValue;
  std::atomic<bool> m_bIsAssigned;
};  // end class single_assign
}  // end namespace ff;
#endif
