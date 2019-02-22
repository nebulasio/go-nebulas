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
#include "ff/util/preprocessor.h"
#include "ff/util/type_list.h"
#include <type_traits>

namespace ff {
namespace util {
namespace internal {
template <typename T> struct nt_traits {};
} // namespace internal

template <typename TL> struct nt_extract_content_type_list {};

template <typename T1, typename T2, typename... TS>
struct nt_extract_content_type_list<type_list<T1, T2, TS...>> {
  typedef typename merge_type_list<
      type_list<typename internal::nt_traits<T1>::type>,
      typename nt_extract_content_type_list<type_list<T2, TS...>>::type>::type
      type;
};

template <typename T1> struct nt_extract_content_type_list<type_list<T1>> {
  typedef type_list<typename internal::nt_traits<T1>::type> type;
};

template <> struct nt_extract_content_type_list<type_list<>> {
  typedef type_list<> type;
};

} // namespace util
} // namespace ff

#define define_nt(...) JOIN(define_nt_impl_, PP_NARG(__VA_ARGS__))(__VA_ARGS__)

#define define_nt_impl_2(_name, _dtype)                                        \
  struct _name {};                                                             \
  template <> struct ::ff::util::internal::nt_traits<_name> {                  \
    constexpr static const char *name = #_name;                                \
    typedef _dtype type;                                                       \
  };

#define define_nt_impl_3(_name, _dtype, _tname)                                \
  struct _name {};                                                             \
  template <> struct ::ff::util::nt_traits<_name> {                            \
    constexpr static const char *name = _tname;                                \
    typedef _dtype type;                                                       \
  };                                                                           \
  \
