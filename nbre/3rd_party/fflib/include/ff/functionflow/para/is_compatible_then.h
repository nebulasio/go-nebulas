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
#ifndef FF_PARA_IS_COMPATIBLE_THEN_H_
#define FF_PARA_IS_COMPATIBLE_THEN_H_

#include "ff/functionflow/common/common.h"
#include "ff/functionflow/para/bin_wait_func_deducer.h"
#include "ff/util/func_type_checker.h"
#include "ff/util/is_callable.h"

namespace ff {
namespace internal {

template <class F, class RT1, class RT2>
struct is_compatible_then {
  const static bool is_cpt_with_and =
      ::ff::util::is_callable<F>::value &&
      ::ff::util::is_function_with_two_arg_type<F, RT1, RT2>::value;

  const static bool is_cpt_with_or =
      ::ff::util::is_callable<F>::value &&
      ::ff::util::is_function_with_two_arg_type<F, int,
                                                std::tuple<RT1, RT2>>::value;
};

template <class F, class RT1>
struct is_compatible_then<F, RT1, void> {
  const static bool is_cpt_with_and =
      ::ff::util::is_callable<F>::value &&
      ::ff::util::is_function_with_two_arg_type<F, RT1, void>::value;
  const static bool is_cpt_with_or =
      ::ff::util::is_callable<F>::value &&
      ::ff::util::is_function_with_two_arg_type<F, bool, RT1>::value;
};

template <class F, class RT1>
struct is_compatible_then<F, void, RT1> {
  const static bool is_cpt_with_and =
      ::ff::util::is_callable<F>::value &&
      ::ff::util::is_function_with_two_arg_type<F, RT1, void>::value;
  const static bool is_cpt_with_or =
      ::ff::util::is_callable<F>::value &&
      ::ff::util::is_function_with_two_arg_type<F, bool, RT1>::value;
};

template <class F>
struct is_compatible_then<F, void, void> {
  const static bool is_cpt_with_and =
      ::ff::util::is_callable<F>::value &&
      ::ff::util::is_function_with_two_arg_type<F, void, void>::value;
  const static bool is_cpt_with_or =
      ::ff::util::is_callable<F>::value &&
      ::ff::util::is_function_with_two_arg_type<F, void, void>::value;
};

}  // end namespace internal
}  // end namespace ff

#endif
