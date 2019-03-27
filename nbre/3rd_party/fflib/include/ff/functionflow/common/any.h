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
#ifndef FF_COMMON_ANY_H_
#define FF_COMMON_ANY_H_

#include "ff/functionflow/common/common.h"

namespace ff {
namespace internal {
class any_base {};

template <class T>
class any_impl : public any_base {
 public:
  any_impl(const T &t) : m_val(t){};
  T &get() { return m_val; }
  const T &get() const { return m_val; }

 protected:
  T m_val;
};
}  // end namespace internal

class any_value {
 public:
  template <class T>
  any_value(const T &t)
      : m_pVal(nullptr) {
    m_pVal = std::shared_ptr<internal::any_base>(new internal::any_impl<T>(t));
  };

  template <class T>
  T &get() {
    internal::any_impl<T> *p =
        static_cast<internal::any_impl<T> *>(m_pVal.get());
    return p->get();
  };

  template <class T>
  const T &get() const {
    const internal::any_impl<T> *p =
        static_cast<const internal::any_impl<T> *>(m_pVal.get());
    return p->get();
  };

 protected:
  std::shared_ptr<internal::any_base> m_pVal;
};
}  // end namespace ff

#endif
