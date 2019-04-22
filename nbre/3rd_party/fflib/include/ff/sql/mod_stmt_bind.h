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
#include "ff/sql/rows.h"
#include "ff/util/type_list.h"

namespace ff {
namespace sql {

template <typename TT, int index> struct traverse_row_for_bind {
  typedef typename TT::engine_type engine_type;
  typedef typename engine_type::native_statement_type native_statement_type;

  template <typename VT>
  static auto run(engine_type *engine, native_statement_type stmt,
                  const VT &val) ->
      typename std::enable_if<(VT::type_list::len > index), void>::type {
    typedef typename util::get_type_at_index_in_typelist<typename VT::type_list,
                                                         index>::type cur_type;
    engine->bind_to_native_statement(stmt, index + 1,
                                     val.template get<cur_type>());
    traverse_row_for_bind<TT, index + 1>::run(engine, stmt, val);
  }

  template <typename VT>
  static auto run(engine_type *engine, native_statement_type stmt,
                  const VT &val) ->
      typename std::enable_if<(VT::type_list::len <= index), void>::type {}
};
//////////////////////////
template <typename TT, int index, typename TL>
struct traverse_row_for_bind_and_put_key_to_last {
  typedef typename TT::engine_type engine_type;
  typedef typename engine_type::native_statement_type native_statement_type;

  template <typename VT>
  static auto run(engine_type *engine, native_statement_type stmt,
                  const VT &val, int &next_index) ->
      typename std::enable_if<(VT::type_list::len > index), void>::type {
    typedef
        typename util::get_type_at_index_in_typelist<typename VT::type_list,
                                                     index>::type current_type;

    if (std::is_base_of<key<typename current_type::type>,
                        current_type>::value) {
      traverse_row_for_bind_and_put_key_to_last<TT, index + 1, TL>::run(
          engine, stmt, val, next_index);
      engine->bind_to_native_statement(stmt, next_index + 1,
                                       val.template get<current_type>());
      next_index++;
    } else {
      engine->bind_to_native_statement(stmt, next_index + 1,
                                       val.template get<current_type>());
      next_index++;
      traverse_row_for_bind_and_put_key_to_last<TT, index + 1, TL>::run(
          engine, stmt, val, next_index);
    }
  }

  template <typename VT>
  static auto run(engine_type *engine, native_statement_type stmt,
                  const VT &val, int &next_index) ->
      typename std::enable_if<(VT::type_list::len <= index), void>::type {}
};

} // namespace sql
} // namespace ff
