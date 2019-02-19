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
#include "ff/sql/engine.h"
#include "ff/sql/mod_stmt_bind.h"
#include "ff/sql/rows.h"
#include "ff/sql/stmt.h"
#include "ff/sql/table_create.h"
#include <sstream>

namespace ff {
namespace sql {

template <typename ET, typename TM, typename... ARGS> class table {
public:
  typedef ET engine_type;
  typedef TM meta_type;
  typedef util::type_list<ARGS...> cols_type;
  typedef row_type_base<ARGS...> row_type;
  typedef row_collection<ARGS...> row_collection_type;
  typedef table<ET, TM, ARGS...> self_type;

  template <typename... TARGS>
  static select_statement<self_type, TARGS...> select(engine_type *engine) {
    static_assert(util::is_contain_types<typename util::type_list<ARGS...>,
                                         util::type_list<TARGS...>>::value,
                  "Can't select row which is not in table");
    select_statement<self_type, TARGS...> ss(engine);
    return ss;
  }

  static void create_table(engine_type *engine) {
    std::stringstream ss;
    ss << "create table if not exists " << TM::table_name << " (";
    recursive_dump_col_creation<ARGS...>(ss);
    ss << ";";
    engine->eval_sql_string(ss.str());
    ss.str(std::string());
    recursive_dump_for_index<engine_type, ARGS...>(engine, ss);
    engine->eval_sql_string(ss.str());
  }
  static void clear_table(engine_type *engine) {
    std::stringstream ss;
    ss << "delete from " << TM::table_name << ";";
    engine->eval_sql_string(ss.str());
  }
  static void drop_table(engine_type *engine) {
    std::stringstream ss;
    ss << "drop table " << TM::table_name << ";";
    engine->eval_sql_string(ss.str());
  }
  static bool insert_or_ignore_rows(engine_type *engine,
                                    const row_collection_type &rows) {
    if (rows.size() == 0) {
      return true;
    }
    std::stringstream ss;
    ss << "insert or ignore into " << TM::table_name << " (";
    recursive_dump_col_name<ARGS...>(ss);
    ss << ") values (";
    for (int i = 0; i < util::type_list<ARGS...>::len; i++) {
      if (i != 0) {
        ss << ", ";
      }
      ss << "?";
    }
    ss << ");";
    engine->begin_transaction();
    for (size_t i = 0; i < rows.size(); ++i) {
      typename engine_type::native_statement_type stmt =
          engine->prepare_sql_with_string(ss.str());
      row_type rt = rows[i];
      traverse_row_for_bind<self_type, 0>::run(stmt, rt);
      engine->eval_native_sql_stmt(stmt);
    }

    engine->end_transaction();
    return true;
  }
  static bool insert_or_replace_rows(engine_type *engine,
                                     const row_collection_type &rows) {
    if (rows.size() == 0) {
      return true;
    }
    std::stringstream ss;
    ss << "replace into " << TM::table_name << " (";
    recursive_dump_col_name<ARGS...>(ss);
    ss << ") values (";
    for (int i = 0; i < util::type_list<ARGS...>::len; i++) {
      if (i != 0) {
        ss << ", ";
      }
      ss << "?";
    }
    ss << ");";
    engine->begin_transaction();
    for (size_t i = 0; i < rows.size(); ++i) {
      typename engine_type::native_statement_type stmt =
          engine->prepare_sql_with_string(ss.str());
      const row_type &rt = rows[i];
      traverse_row_for_bind<self_type, 0>::run(engine, stmt, rt);
      engine->eval_native_sql_stmt(stmt);
    }

    engine->end_transaction();
    return true;
  }

  template <typename... RS>
  static bool update_rows(engine_type *engine,
                          const row_collection<RS...> &rows) {
    static_assert(util::is_contain_types<util::type_list<ARGS...>,
                                         util::type_list<RS...>>::value,
                  "Can't update rows which are not in table");
    static_assert(
        !util::is_same_type<void, typename get_key_column_type<
                                      util::type_list<RS...>>::type>::value,
        "Have to specify a key column for update");
    if (rows.size() == 0) {
      return true;
    }
    std::stringstream ss;
    ss << "update " << TM::table_name << " set ";
    recurseive_dump_update_item_and_ignore_key<RS...>(ss);
    ss << " where ";
    recurseive_dump_update_item_for_only_where<RS...>(ss);
    ss << ";";
    std::string sql = ss.str();

    engine->begin_transaction();

    for (size_t i = 0; i < rows.size(); ++i) {
      typename engine_type::native_statement_type stmt =
          engine->prepare_sql_with_string(ss.str());
      const typename row_collection<RS...>::row_type &rt = rows[i];
      int next_index = 0;
      traverse_row_for_bind_and_put_key_to_last<
          self_type, 0, util::type_list<RS...>>::run(engine, stmt, rt,
                                                     next_index);
      engine->eval_native_sql_stmt(stmt);
    }
    engine->end_transaction();
    return true;
  }

  static delete_statement<self_type> delete_rows(engine_type *engine) {
    std::stringstream ss;
    ss << "delete from " << TM::table_name << ";";
    std::string sql = ss.str();
    return delete_statement<self_type>(engine, sql);
  }

protected:
#include "ff/sql/table_dump.hpp"
};
} // namespace sql
} // namespace ff
