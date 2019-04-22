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
template <typename TT, typename T> struct traverse_cond_for_bind {
  typedef typename TT::engine_type engine_type;
  typedef typename engine_type::native_statement_type native_statement_type;

  template <typename VT, sql_cond_type CT>
  static void run(engine_type *engine, native_statement_type stmt,
                  const basic_cond_stmt<VT, CT> &cst, int &index) {
    static_assert(std::is_base_of<cond_stmt, T>::value, "T must be cond_stmt");
    engine->bind_to_native_statement(stmt, index + 1, cst.m_value);
    index++;
  }
};

template <typename TT, typename T1, typename T2>
struct traverse_cond_for_bind<TT, and_cond_stmt<T1, T2>> {
  typedef typename TT::engine_type engine_type;
  typedef typename engine_type::native_statement_type native_statement_type;

  static void run(engine_type *engine, native_statement_type stmt,
                  const and_cond_stmt<T1, T2> &cst, int &index) {
    traverse_cond_for_bind<TT, T1>::run(engine, stmt, cst.stmt1, index);
    traverse_cond_for_bind<TT, T2>::run(engine, stmt, cst.stmt2, index);
  }
};

template <typename TT, typename T1, typename T2>
struct traverse_cond_for_bind<TT, or_cond_stmt<T1, T2>> {
  typedef typename TT::engine_type engine_type;
  typedef typename engine_type::native_statement_type native_statement_type;

  static void run(engine_type *engine, native_statement_type stmt,
                  const or_cond_stmt<T1, T2> &cst, int &index) {
    traverse_cond_for_bind<TT, T1>::run(engine, stmt, cst.stmt1, index);
    traverse_cond_for_bind<TT, T2>::run(engine, stmt, cst.stmt2, index);
  }
};
