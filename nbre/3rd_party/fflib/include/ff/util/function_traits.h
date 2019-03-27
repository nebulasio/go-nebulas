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
#include "ff/util/func_type_checker.h"
#include "ff/util/is_callable.h"
#include <functional>
#include <type_traits>

namespace ff {
namespace util {
template <class F>
struct deduce_function {};

template <class Ret, class C, class... Args>
struct deduce_function<Ret (C::*)(Args...) const> {
  typedef std::function<Ret(Args...)> type;
  typedef Ret ret_type;
};

template <class Ret, class C, class... Args>
struct deduce_function<Ret (C::*)(Args...)> {
  typedef std::function<Ret(Args...)> type;
  typedef Ret ret_type;
};

template <class Ret, class... Args>
struct deduce_function<Ret(Args...)> {
  typedef std::function<Ret(Args...)> type;
  typedef Ret ret_type;
};

template <class F>
struct is_no_args_function {
  const static bool value = false;
};

template <class Ret, class C>
struct is_no_args_function<Ret (C::*)(void) const> {
  const static bool value = true;
};

template <class Ret, class C>
struct is_no_args_function<Ret (C::*)(void)> {
  const static bool value = true;
};

template <class Ret>
struct is_no_args_function<Ret(void)> {
  const static bool value = true;
};
template <class Ret>
struct is_no_args_function<Ret (*)(void)> {
  const static bool value = true;
};

template <class F, bool flag>
struct functor_res_traits {
  typedef typename deduce_function<decltype(
      &std::remove_reference<F>::type::operator())>::ret_type ret_type;
};

template <class F>
struct functor_res_traits<F, false> {
  typedef void ret_type;
};

template <class F, bool flag>
struct bind_res_traits {
  typedef typename F::result_type ret_type;
};

template <class F>
struct bind_res_traits<F, false> {
  typedef void ret_type;
};

namespace internal {
template <class F, bool flag>
struct function_res_traits_impl {
  typedef typename deduce_function<F>::ret_type ret_type;
};
template <class F>
struct function_res_traits_impl<F, false> {
  typedef void ret_type;
};

template <class F>
struct function_res_traits_impl<F *, true> {
  typedef typename deduce_function<F>::ret_type ret_type;
};
template <class F>
struct function_res_traits_impl<F &, true> {
  typedef typename deduce_function<F>::ret_type ret_type;
};
} // namespace internal

template <class F>
struct function_res_traits {
  typedef typename std::remove_reference<F>::type FT;
  const static bool s_is_class = std::is_class<FT>::value;
  const static bool s_is_bind_expr = std::is_bind_expression<FT>::value;

  typedef typename std::conditional<
      s_is_class,
      typename std::conditional<
          s_is_bind_expr,
          typename bind_res_traits<FT, s_is_bind_expr>::ret_type,
          typename functor_res_traits<FT, !s_is_bind_expr &&
                                              s_is_class>::ret_type>::type,
      typename internal::function_res_traits_impl<FT, !s_is_class>::ret_type>::
      type ret_type;
};

template <class F, bool flag>
struct functor_args_traits {
  const static bool is_no_args = is_no_args_function<decltype(
      &std::remove_reference<F>::type::operator())>::value;
};

template <class F>
struct functor_args_traits<F, false> {
  const static bool is_no_args = false;
};

namespace internal {
template <class F, bool flag>
struct function_args_traits_impl {
  const static bool is_no_args = is_no_args_function<F>::value;
};

template <class F>
struct function_args_traits_impl<F, false> {
  const static bool is_no_args = false;
};

template <class F>
struct function_args_traits_impl<F *, true> {
  const static bool is_no_args = is_no_args_function<F>::value;
};
template <class F>
struct function_args_traits_impl<F &, true> {
  const static bool is_no_args = is_no_args_function<F>::value;
};
} // namespace internal

template <class F>
struct function_args_traits {
  typedef typename std::remove_reference<F>::type FT;
  const static bool s_is_class = std::is_class<FT>::value;

  const static bool is_no_args =
      functor_args_traits<FT, s_is_class>::is_no_args ||
      internal::function_args_traits_impl<FT, !s_is_class>::is_no_args;
};

} // end namespace util
};  // end namespace ff
