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
#include "ff/sql/common.h"
#include "ff/sql/stmt.h"
#include "ff/util/type_list.h"

namespace ff {
namespace sql {

template <typename T> struct column_base { typedef T type; };
template <typename T> struct column : public column_base<T> {};

template <typename T> class index : public column_base<T> {};

template <typename T> class key : public column_base<T> {};

// template <typename TT, typename CT> struct table_column {
// typedef TT table_type;
// typedef CT column_type;
// typedef typename CT::type type;
//};

template <typename TL> struct get_key_column_type {};

template <typename T, typename... TS>
struct get_key_column_type<ff::util::type_list<T, TS...>> {
  typedef typename std::conditional<
      std::is_base_of<key<typename T::type>, T>::value, T,
      typename get_key_column_type<ff::util::type_list<TS...>>::type>::type
      type;
};

template <> struct get_key_column_type<ff::util::type_list<>> {
  typedef void type;
};
} // namespace sql
} // namespace ff

#define define_column(_name, _type, _dtype, _tname)                            \
  struct _name : public ::ff::sql::_type<_dtype> {                             \
    constexpr static const char *name = _tname;                                \
    static ff::sql::eq_cond_stmt<_name> eq(const _dtype &value) {              \
      return ff::sql::eq_cond_stmt<_name>(value);                              \
    }                                                                          \
    static ff::sql::ne_cond_stmt<_name> ne(const _dtype &value) {              \
      return ff::sql::ne_cond_stmt<_name>(value);                              \
    }                                                                          \
    static ff::sql::le_cond_stmt<_name> le(const _dtype &value) {              \
      return ff::sql::le_cond_stmt<_name>(value);                              \
    }                                                                          \
    static ff::sql::ge_cond_stmt<_name> ge(const _dtype &value) {              \
      return ff::sql::ge_cond_stmt<_name>(value);                              \
    }                                                                          \
  };                                                                           \
  namespace ff {                                                               \
  namespace util {                                                             \
  namespace internal {                                                         \
  template <> struct nt_traits<_name> {                                        \
    constexpr static const char *name = _tname;                                \
    typedef _dtype type;                                                       \
  };                                                                           \
  }                                                                            \
  }                                                                            \
  }

