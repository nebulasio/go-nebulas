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
#include "ff/sql/columns.h"
#include "ff/sql/common.h"
#include <sstream>

namespace ff {
namespace sql {
namespace internal {

template <typename T> struct dump_col_type_creation {
  static void dump(std::stringstream &ss) { ss << " BLOB "; }
};

#define impl_table_dump_types(cpptype, sqltype)                                \
  template <> struct dump_col_type_creation<cpptype> {                         \
    static void dump(std::stringstream &ss) { ss << " " << sqltype << " "; }   \
  }
impl_table_dump_types(uint64_t, "BIGINT UNSIGNED");
impl_table_dump_types(int64_t, "BIGINT");
impl_table_dump_types(uint32_t, "INT UNSIGNED");
impl_table_dump_types(int32_t, "INT");
impl_table_dump_types(int16_t, "SMALLINT");
impl_table_dump_types(uint16_t, "SMALLINT UNSIGNED");
impl_table_dump_types(int8_t, "TINYINT");
impl_table_dump_types(uint8_t, "TINYINT UNSIGNED");
impl_table_dump_types(float, "FLOAT");
impl_table_dump_types(double, "DOUBLE");
impl_table_dump_types(std::string, "VARCHAR(20)");

//////////////////////////
} // namespace internal

enum {
  key_type,
  index_type,
  column_type,
};
template <typename T> struct extract_col_type {
  typedef typename T::type ct;
  const static int value = std::conditional<
      std::is_base_of<key<ct>, T>::value, util::int_number_type<key_type>,
      typename std::conditional<std::is_base_of<index<ct>, T>::value,
                                util::int_number_type<index_type>,
                                util::int_number_type<column_type>>::type>::
      type::value;
};
template <typename T, int V = extract_col_type<T>::value>
struct dump_col_creation {};

template <typename T> struct dump_col_creation<T, key_type> {
  static void dump(std::stringstream &ss) {
    ss << T::name;
    internal::dump_col_type_creation<typename T::type>::dump(ss);
    ss << " primary key";
  }
};

template <typename T> struct dump_col_creation<T, index_type> {
  static void dump(std::stringstream &ss) {
    ss << T::name;
    internal::dump_col_type_creation<typename T::type>::dump(ss);
  }
};
template <typename T> struct dump_col_creation<T, column_type> {
  static void dump(std::stringstream &ss) {
    ss << T::name;
    internal::dump_col_type_creation<typename T::type>::dump(ss);
  }
};
} // namespace sql
} // namespace ff
