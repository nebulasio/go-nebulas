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
#include <iostream>

namespace ff {
namespace sql {

struct sqlite3 {}; // TODO

struct cppconn {};

template <class FS> class mysql {};

struct sql_debug {
  typedef void *native_statement_type;
  void eval_sql_string(const std::string &sql) {
    std::cout << sql << std::endl;
  }

  native_statement_type prepare_sql_with_string(const std::string &sql) {
    std::cout << sql << std::endl;
    return NULL;
  }
  void eval_native_sql_stmt(native_statement_type stmt) {}
  template <typename T>
  void bind_to_native_statement(native_statement_type stmt, int index,
                                const T &value) {
    std::cout << "bind value: " << value << " to " << index << std::endl;
  }

  void begin_transaction() { std::cout << "begin transaction" << std::endl; }
  void end_transaction() { std::cout << "end transaction" << std::endl; }
};
} // namespace sql
} // namespace ff
