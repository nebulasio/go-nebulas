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

#include "ff/sql/engine.h"
#include "ff/sql/rows.h"
#include "ff/sql/table.h"
#include <cppconn/driver.h>
#include <cppconn/exception.h>
#include <cppconn/prepared_statement.h>
#include <cppconn/resultset.h>
#include <cppconn/statement.h>
#include <sstream>
#include <thread>

namespace ff {
namespace sql {


    template<class STMT, class T> struct mysql_bind_setter{
      static void bind(STMT , int , const T& ){
        throw std::runtime_error("No support yet");
      }
    };
#define impl_mysql_bind_setter(type, method) \
    template<class STMT> \
      struct mysql_bind_setter<STMT, type>{ \
      static void bind(STMT stmt, int index, const type & value){ \
        stmt->method(index, value); \
      } \
      };
    impl_mysql_bind_setter(std::string, setString);
    impl_mysql_bind_setter(int8_t, setInt);
    impl_mysql_bind_setter(uint8_t, setUInt);
    impl_mysql_bind_setter(int16_t, setInt);
    impl_mysql_bind_setter(uint16_t, setUInt);
    impl_mysql_bind_setter(int32_t, setInt);
    impl_mysql_bind_setter(uint32_t, setUInt);
    impl_mysql_bind_setter(int64_t, setInt64);
    impl_mysql_bind_setter(uint64_t, setUInt64);
    impl_mysql_bind_setter(double, setDouble);
    impl_mysql_bind_setter(float, setDouble);
#undef impl_mysql_bind_setter

    template<class T>
      struct mysql_rs_getter{};
#define impl_mysql_rs_getter(type, method) \
    template<> \
      struct mysql_rs_getter<type>{ \
        template <typename RST> \
          static type get(RST r, const std::string & name){ \
            return r->method(name); \
          } \
      }; \

    impl_mysql_rs_getter(std::string, getString);
    impl_mysql_rs_getter(double, getDouble);
    impl_mysql_rs_getter(float, getDouble);
    impl_mysql_rs_getter(int64_t, getInt64);
    impl_mysql_rs_getter(uint64_t, getUInt64);
    impl_mysql_rs_getter(int32_t, getInt);
    impl_mysql_rs_getter(uint32_t, getUInt);
    impl_mysql_rs_getter(int16_t, getInt);
    impl_mysql_rs_getter(uint16_t, getUInt);
    impl_mysql_rs_getter(int8_t, getInt);
    impl_mysql_rs_getter(uint8_t, getUInt);
#undef impl_mysql_rs_getter

    template<typename... ARGS>
      struct mysql_record_setter{};

    template<>
      struct mysql_record_setter<>{
        template<typename RT, typename RST>
          static void set(RT & , RST  ){
          }
      };
    template<typename T, typename... ARGS>
      struct mysql_record_setter<T, ARGS...>{
        template<typename RT, typename RST>
          static void set(RT & row, RST r){
            row.template set<T>(mysql_rs_getter<typename T::type>::get(r, T::name));
            mysql_record_setter<ARGS...>::set(row, r);
          }
      };

    template<typename... ARGS> class lazy_eval_string_impl{
      public:
          void to_string(std::stringstream & )const {
          }
    };

    template<typename T, typename... ARGS>
      class lazy_eval_string_impl<T, ARGS...>{
        public:
          lazy_eval_string_impl(const T & t, const ARGS& ...args): m_t(t), m_params(args...){}
          void to_string(std::stringstream & ss) const{
            ss<<m_t;
            m_params.to_string(ss);
          }
        protected:
          const T & m_t;
          lazy_eval_string_impl<ARGS...> m_params;
      };
    template<typename... ARGS>
      lazy_eval_string_impl<ARGS...> lazy_eval_string(const ARGS& ...args){
        return lazy_eval_string_impl<ARGS...>(args...);
      }
template <> class mysql<cppconn> {
  public:
    typedef std::shared_ptr<::sql::PreparedStatement > native_statement_type;
    typedef std::shared_ptr<::sql::ResultSet>  query_result_type;

    mysql(const std::string & url, const std::string & usrname, const std::string & passwd, const std::string &dbname)
    : m_url(url), m_usrname(usrname), m_passwd(passwd), m_dbname(dbname), m_is_worker_thread(false){

    m_sql_driver = get_driver_instance();
    m_sql_conn.reset(m_sql_driver->connect(m_url, m_usrname, m_passwd));
    m_sql_conn->setSchema(m_dbname);
    m_local_thread_id = std::this_thread::get_id();
  }

    virtual ~mysql(){
      if(m_is_worker_thread){
        m_sql_driver->threadEnd();
      }
    }

  protected:
    mysql(mysql<cppconn> * engine)
      : m_sql_driver(engine->m_sql_driver)
        , m_url(engine->m_url)
        , m_usrname(engine->m_usrname)
        , m_passwd(engine->m_passwd)
        , m_dbname(engine->m_dbname)
        , m_is_worker_thread(true){
          m_sql_driver->threadInit();
          m_sql_conn.reset(m_sql_driver->connect(m_url, m_usrname, m_passwd));
          m_sql_conn->setSchema(m_dbname);
          m_local_thread_id = std::this_thread::get_id();
        }
  public:

  std::shared_ptr<mysql<cppconn> > thread_copy(){
    std::shared_ptr<mysql<cppconn> > ret;
    ret.reset(new mysql<cppconn>(this));
    return ret;
    }

  void eval_sql_string(const std::string &sql) {
    if(sql.empty()) return ;
    check_local_thread(lazy_eval_string("can't call ", sql, " in another thread. Please check thread_copy()."));
    std::shared_ptr<::sql::Statement> stmt(m_sql_conn->createStatement());
    stmt->execute(sql);
  }

  query_result_type eval_sql_query_string(const std::string &sql) {
    check_local_thread(lazy_eval_string("can't call ", sql, " in another thread. Please check thread_copy()."));
    std::shared_ptr<::sql::Statement> stmt(m_sql_conn->createStatement());
    query_result_type ret;
    ret.reset(stmt->executeQuery(sql));
    return ret;
  }

  template<typename... ARGS>
    row_collection<ARGS...> parse_records(query_result_type query_result){
      row_collection<ARGS...> rows;

      while(query_result->next()){
        typename row_collection<ARGS...>::row_type r;
        mysql_record_setter<ARGS...>::set(r, query_result);
        rows.push_back(std::move(r));
      }
      return rows;
    }

  native_statement_type prepare_sql_with_string(const std::string &sql) {
    check_local_thread(lazy_eval_string("can't call ", sql, " in another thread. Please check thread_copy()."));
    native_statement_type stmt;
    stmt.reset(m_sql_conn->prepareStatement(sql.c_str()));
    return stmt;
  }

  query_result_type eval_native_sql_stmt(native_statement_type stmt) {
    check_local_thread(lazy_eval_string("can't call ", "executeQuery", " in another thread. Please check thread_copy()."));
    query_result_type ret;
    ret.reset(stmt->executeQuery());
    return ret;
  }
  template <typename T>
  void bind_to_native_statement(native_statement_type stmt, int index,
                                const T &value) {
    mysql_bind_setter<native_statement_type, T>::bind(stmt, index, value);
  }

  void begin_transaction() {
    eval_sql_string("START TRANSACTION");
  }
  void end_transaction() {
    eval_sql_string("COMMIT");
  }

  protected:
  template<typename T>
  void check_local_thread(const T & msg){
    std::thread::id local_id = std::this_thread::get_id();
    if(local_id != m_local_thread_id){
      std::stringstream ss;
      msg.to_string(ss);
      throw std::runtime_error(ss.str());
    }
  }

protected:
  ::sql::Driver *m_sql_driver;
  std::shared_ptr<::sql::Connection> m_sql_conn;

  const std::string m_url;
  const std::string m_usrname;
  const std::string m_passwd;
  const std::string m_dbname;
  bool m_is_worker_thread;
  std::thread::id m_local_thread_id;
};

using default_engine = mysql<cppconn>;

template <typename TM, typename... ARGS>
using default_table = table<default_engine, TM, ARGS...>;
} // namespace sql
} // namespace ff

