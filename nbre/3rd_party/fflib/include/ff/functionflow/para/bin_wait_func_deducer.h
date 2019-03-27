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
#include "ff/functionflow/common/common.h"
#include "ff/functionflow/para/is_compatible_then.h"
#include "ff/util/function_traits.h"
#include <tuple>

namespace ff {
template <class RT>
class para;
namespace internal {
template <class RT1, class RT2>
struct bin_wait_func_deducer {
  typedef std::tuple<RT1, RT2> pair;
  typedef pair and_type;
  typedef std::tuple<int, pair> or_type;

  template <class FT, class T1, class T2>
  static void void_func_and(FT&& f, T1&& t1, T2&& t2) {
    f(t1.get(), t2.get());
  }
  template <class FT, class T1, class T2>
  static auto ret_func_and(FT &&f, T1 &&t1, T2 &&t2) ->
      typename std::remove_reference<
          typename util::function_res_traits<FT>::ret_type>::type {
    return f(t1.get(), t2.get());
  }

  template <class FT, class T1, class T2>
  static void void_func_or(FT&& f, T1&& t1, T2&& t2) {
    int i = 0;
    if (t1.get_state() == exe_state::exe_over)
      i = 0;
    else if (t2.get_state() == exe_state::exe_over)
      i = 1;
    f(i, std::make_tuple(t1.get(), t2.get()));
  }

  template <class FT, class T1, class T2>
  static auto ret_func_or(FT &&f, T1 &&t1, T2 &&t2) ->
      typename std::remove_reference<
          typename util::function_res_traits<FT>::ret_type>::type {
    int i = 0;
    if (t1.get_state() == exe_state::exe_over)
      i = 0;
    else if (t2.get_state() == exe_state::exe_over)
      i = 1;
    return f(i, std::make_tuple(t1.get(), t2.get()));
  }

  template <class T1, class T2>
  static and_type wrap_ret_for_and(T1&& t1, T2&& t2) {
    return pair(t1.get(), t2.get());
  }

  template <class T1, class T2>
  static or_type wrap_ret_for_or(T1&& t1, T2&& t2) {
    int i = 0;
    if (t1.get_state() == exe_state::exe_over)
      i = 0;
    else if (t2.get_state() == exe_state::exe_over)
      i = 1;
    return std::make_tuple(i, std::make_tuple(t1.get(), t2.get()));
  }

};  // end struct bin_wait_func_deducer;
template <class RT2>
struct bin_wait_func_deducer<void, RT2> {
  typedef RT2 pair;
  typedef pair and_type;
  typedef std::tuple<bool, RT2> or_type;
  template <class FT, class T1, class T2>
  static void void_func_and(FT &&f, T1 &&, T2 &&t2) {
    f(t2.get());
  }
  template <class FT, class T1, class T2>
  static auto ret_func_and(FT &&f, T1 &&, T2 &&t2) ->
      typename std::remove_reference<
          typename util::function_res_traits<FT>::ret_type>::type {
    return f(t2.get());
  }

  template <class FT, class T1, class T2>
  static void void_func_or(FT &&f, T1 &&, T2 &&t2) {
    int i = 0;
    if (t2.get_state() == exe_state::exe_over) i = 1;
    f(i == 1, t2.get());
  }
  template <class FT, class T1, class T2>
  static auto ret_func_or(FT &&f, T1 &&, T2 &&t2) ->
      typename std::remove_reference<
          typename util::function_res_traits<FT>::ret_type>::type {
    int i = 0;
    if (t2.get_state() == exe_state::exe_over) i = 1;
    return f(i == 1, t2.get());
  }

  template <class T1, class T2>
  static and_type wrap_ret_for_and(T1 &&, T2 &&t2) {
    return pair(t2.get());
  }

  template <class T1, class T2> static or_type wrap_ret_for_or(T1 &&, T2 &&t2) {
    int i = 0;
    if (t2.get_state() == exe_state::exe_over) i = 1;
    return std::make_tuple(i == 1, t2.get());
  }
};  // end struct bin_wait_func_deducer;
template <class RT1>
struct bin_wait_func_deducer<RT1, void> {
  typedef RT1 pair;
  typedef pair and_type;
  typedef std::tuple<bool, RT1> or_type;

  template <class FT, class T1, class T2>
  static void void_func_and(FT &&f, T1 &&t1, T2 &&) {
    f(t1.get());
  }
  template <class FT, class T1, class T2>
  static auto ret_func_and(FT &&f, T1 &&t1, T2 &&) ->
      typename std::remove_reference<
          typename ::ff::util::function_res_traits<FT>::ret_type>::type {
    return f(t1.get());
  }

  template <class FT, class T1, class T2>
  static void void_func_or(FT &&f, T1 &&t1, T2 &&) {
    int i = 0;
    if (t1.get_state() == exe_state::exe_over) i = 1;
    f(i == 1, t1.get());
  }
  template <class FT, class T1, class T2>
  static auto ret_func_or(FT &&f, T1 &&t1, T2 &&) ->
      typename std::remove_reference<
          typename util::function_res_traits<FT>::ret_type>::type {
    int i = 0;
    if (t1.get_state() == exe_state::exe_over) i = 1;
    return f(i == 1, t1.get());
  }

  template <class T1, class T2>
  static and_type wrap_ret_for_and(T1 &&t1, T2 &&) {
    return pair(t1.get());
  }

  template <class T1, class T2> static or_type wrap_ret_for_or(T1 &&t1, T2 &&) {
    int i = 0;
    if (t1.get_state() == exe_state::exe_over) i = 1;
    return std::make_tuple(i == 1, t1.get());
  }
};  // end struct bin_wait_func_deducer;

template <>
struct bin_wait_func_deducer<void, void> {
  typedef void pair;
  typedef pair and_type;
  typedef pair or_type;

  template <class FT, class T1, class T2>
  static void void_func_and(FT &&f, T1 &&, T2 &&) {
    f();
  }
  template <class FT, class T1, class T2>
  static auto ret_func_and(FT &&f, T1 &&, T2 &&) ->
      typename std::remove_reference<
          typename util::function_res_traits<FT>::ret_type>::type {
    return f();
  }

  template <class FT, class T1, class T2>
  static void void_func_or(FT &&f, T1 &&, T2 &&) {
    f();
  }
  template <class FT, class T1, class T2>
  static auto ret_func_or(FT &&f, T1 &&, T2 &&) ->
      typename std::remove_reference<
          typename util::function_res_traits<FT>::ret_type>::type {
    return f();
  }

  template <class T1, class T2>
  static and_type wrap_ret_for_and(T1 &&, T2 &&) {}

  template <class T1, class T2> static or_type wrap_ret_for_or(T1 &&, T2 &&) {}
};  // end struct bin_wait_func_deducer;

}  // end namespace internal
}  // end namespace ff
