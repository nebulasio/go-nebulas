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

#ifndef FF_PARA_IS_WAIT_COMPATIBLE_WITH_THEN_H
#define FF_PARA_IS_WAIT_COMPATIBLE_WITH_THEN_H

#include "ff/functionflow/para/is_compatible_then.h"

namespace ff {
// forward declaration
template <class RT>
class para;
namespace internal {
template <class RT1, class RT2>
class wait_and;
template <class RT1, class RT2>
class wait_or;

template <class WT, class F>
struct is_wait_compatible_with_func {
  const static bool value = false;
};

template <class T1, class T2, class F>
struct is_wait_compatible_with_func<wait_and<T1, T2>, F> {
  typedef typename std::remove_reference<T1>::type T1_t;
  typedef typename std::remove_reference<T2>::type T2_t;
  typedef typename T1_t::ret_type RT1_t;
  typedef typename T2_t::ret_type RT2_t;
  const static bool value =
      is_compatible_then<F, RT1_t, RT2_t>::is_cpt_with_and;
};

template <class T1, class T2, class F>
struct is_wait_compatible_with_func<wait_or<T1, T2>, F> {
  typedef typename std::remove_reference<T1>::type T1_t;
  typedef typename std::remove_reference<T2>::type T2_t;
  typedef typename T1_t::ret_type RT1_t;
  typedef typename T2_t::ret_type RT2_t;
  const static bool value = is_compatible_then<F, RT1_t, RT2_t>::is_cpt_with_or;
};

template <class RT, class F>
struct is_wait_compatible_with_func<para<RT>, F> {
  const static bool value = util::is_function_with_arg_type<F, RT>::value;
};
}  // end namespace internal
}  // end namespace ff
#endif  // FUNCTIONFLOW_IS_WAIT_COMPATIBLE_WITH_THEN_H
