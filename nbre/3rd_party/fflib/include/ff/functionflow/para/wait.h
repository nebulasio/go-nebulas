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
#ifndef FF_PARA_WAIT_H_
#define FF_PARA_WAIT_H_
#include "ff/functionflow/common/common.h"
#include "ff/functionflow/para/bin_wait_func_deducer.h"
#include "ff/functionflow/para/para.h"
#include "ff/functionflow/para/para_wait_traits.h"
#include "ff/functionflow/para/paracontainer.h"
#include "ff/functionflow/para/paragroup.h"
#include "ff/functionflow/para/wait_impl.h"
#include "ff/util/tuple_type.h"

namespace ff {

template <class T1, class T2>
auto operator&&(T1 &&t1, T2 &&t2) -> typename std::enable_if<
    is_para_or_wait<typename std::remove_reference<T1>::type>::value &&
        is_para_or_wait<typename std::remove_reference<T2>::type>::value,
    ::ff::internal::wait_and<T1, T2>>::type {
  return ::ff::internal::wait_and<T1, T2>(std::forward<T1>(t1),
                                          std::forward<T2>(t2));
}

template <class T1, class T2>
auto operator||(T1 &&t1, T2 &&t2) -> typename std::enable_if<
    is_para_or_wait<typename std::remove_reference<T1>::type>::value &&
        is_para_or_wait<typename std::remove_reference<T2>::type>::value,
    ::ff::internal::wait_or<T1, T2>>::type {
  return ::ff::internal::wait_or<T1, T2>(std::forward<T1>(t1),
                                         std::forward<T2>(t2));
}

template <class T1, class T2>
auto operator&&(T1 &&t1, T2 &&t2) -> typename std::enable_if<
    (!is_para_or_wait<typename std::remove_reference<T1>::type>::value &&
     is_para_or_wait<typename std::remove_reference<T2>::type>::value) ||
        (is_para_or_wait<typename std::remove_reference<T1>::type>::value &&
         !is_para_or_wait<typename std::remove_reference<T2>::type>::value),
    ::ff::internal::wait_and<para<void>, para<void>>>::type {
  static_assert(Please_Check_The_Assert_Msg<T1>::value,
                FF_EM_COMBINE_PARA_AND_OTHER);
}

template <class T1, class T2>
auto operator||(T1 &&t1, T2 &&t2) -> typename std::enable_if<
    (!is_para_or_wait<typename std::remove_reference<T1>::type>::value &&
     is_para_or_wait<typename std::remove_reference<T2>::type>::value) ||
        (is_para_or_wait<typename std::remove_reference<T1>::type>::value &&
         !is_para_or_wait<typename std::remove_reference<T2>::type>::value),
    ::ff::internal::wait_or<para<void>, para<void>>>::type {
  static_assert(Please_Check_The_Assert_Msg<T1>::value,
                FF_EM_COMBINE_PARA_AND_OTHER);
}

auto all(paragroup &pg) -> ::ff::internal::wait_all;
auto any(paragroup &pg) -> ::ff::internal::wait_any;

auto all(paracontainer &pc) -> ::ff::internal::wait_all;
auto any(paracontainer &pc) -> ::ff::internal::wait_any;

}  // end namespace ff

#endif
