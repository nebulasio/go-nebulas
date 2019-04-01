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
#include "ff/sql/rows.h"
#include <sstream>

namespace ff {
namespace sql {

struct cond_stmt {
public:
  virtual void dump_to_sql_string(std::stringstream &ss) const = 0;
    };


    template<typename T1, typename T2>
      struct and_cond_stmt : public cond_stmt{
        and_cond_stmt(const T1 & t1, const T2 & t2): stmt1(t1), stmt2(t2){}
        T1 stmt1;
        T2 stmt2;

        typedef typename util::merge_type_list<
            typename T1::cols_type, typename T2::cols_type>::type cols_type;

        void traverse_and_call_back(const std::function<void (cond_stmt *)> & callback){
          stmt1.traverse_and_call_back(callback);
          callback(this);
          stmt2.traverse_and_call_back(callback);
        }

        virtual void dump_to_sql_string(std::stringstream &ss) const {
          ss<<" (";
          stmt1.dump_to_sql_string(ss);
          ss<<" and ";
          stmt2.dump_to_sql_string(ss);
          ss<<") ";
        }
      };
    template< typename T1, typename T2>
      struct or_cond_stmt : public cond_stmt{
        or_cond_stmt(const T1 & t1, const T2 & t2): stmt1(t1), stmt2(t2){}
        void traverse_and_call_back(const std::function<void (cond_stmt *)> & callback){
          stmt1.traverse_and_call_back(callback);
          callback(this);
          stmt2.traverse_and_call_back(callback);
        }
        T1 stmt1;
        T2 stmt2;
        typedef typename util::merge_type_list<
            typename T1::cols_type, typename T2::cols_type>::type cols_type;
        virtual void dump_to_sql_string(std::stringstream &ss) const {
          ss<<" (";
          stmt1.dump_to_sql_string(ss);
          ss<<" or ";
          stmt2.dump_to_sql_string(ss);
          ss<<") ";
        }
      };


    enum sql_cond_type{
      eq_cond,
      ne_cond,
      ge_cond,
      le_cond,
    };

    template<typename T, sql_cond_type CT>
      struct basic_cond_stmt : public cond_stmt{
        basic_cond_stmt(const typename T::type & value) : m_value (value){}
        void traverse_and_call_back(const std::function<void (cond_stmt *)> & callback){
          callback(this);
        }
        typename T::type m_value;

        typedef util::type_list<T> cols_type;
        const static sql_cond_type cond_type = CT;
      };

    template<typename T>
      struct eq_cond_stmt : public basic_cond_stmt<T, eq_cond>{
        eq_cond_stmt(const typename T::type & value): basic_cond_stmt<T, eq_cond>(value){}

        virtual void dump_to_sql_string(std::stringstream &ss) const {
          ss << T::name << " = ? ";
        }
      };
    template<typename T>
      struct ne_cond_stmt : public basic_cond_stmt<T, ne_cond>{
        ne_cond_stmt(const typename T::type & value): basic_cond_stmt<T, ne_cond>(value){}
        virtual void dump_to_sql_string(std::stringstream &ss) const {
          ss<<T::name<<" != ? ";
        }
      };
    template<typename T>
      struct ge_cond_stmt : public basic_cond_stmt<T, ge_cond>{
        ge_cond_stmt(const typename T::type & value): basic_cond_stmt<T, ge_cond>(value){}
        virtual void dump_to_sql_string(std::stringstream &ss) const {
          ss<<T::name<<" >= ? ";
        }
      };
    template<typename T>
      struct le_cond_stmt : public basic_cond_stmt<T, le_cond>{
        le_cond_stmt(const typename T::type & value): basic_cond_stmt<T, le_cond>(value){}
        virtual void dump_to_sql_string(std::stringstream &ss) const {
          ss<<T::name<<" <= ? ";
        }
      };

    template<typename TT, typename... ARGS>
      struct statement{
        public:
          typedef typename TT::engine_type engine_type;
          statement(engine_type *engine) : m_engine(engine) {}
          virtual row_collection<ARGS...> eval() = 0;
        protected:
          engine_type *m_engine;
      };

#include "ff/sql/where_stmt_bind.hpp"

      template <typename TT, typename... ARGS>
      struct limit_statement : public statement<TT, ARGS...> {
      public:
        typedef typename TT::engine_type engine_type;
        limit_statement(engine_type *engine, const std::string &prev_sql,
                        int64_t count)
            : statement<TT, ARGS...>(engine), m_prev_sql(prev_sql),
              m_count(count) {}

        virtual row_collection<ARGS...> eval() {
          std::stringstream ss;
          ss << get_eval_sql_string() << ";";
          auto ret = m_engine->eval_sql_query_string(ss.str());
          return m_engine->template parse_records<ARGS...>(ret);
        }

      protected:
        std::string get_eval_sql_string() const {
          std::stringstream ss;
          ss << m_prev_sql << " LIMIT " << m_count;
          return ss.str();
        }

      protected:
        using statement<TT, ARGS...>::m_engine;
        std::string m_prev_sql;
        int64_t m_count;
      };

      struct desc {
        constexpr static const char *name = "DESC";
      };
      struct asc {
        constexpr static const char *name = "ASC";
      };

      template <typename TT, typename CT, typename ORDER, typename... ARGS>
      struct order_statement : public statement<TT, ARGS...> {
      public:
        typedef typename TT::engine_type engine_type;
        order_statement(engine_type *engine, const std::string &prev_sql)
            : statement<TT, ARGS...>(engine), m_prev_sql(prev_sql) {}

        limit_statement<TT, ARGS...> limit(int64_t count) {
          if (count <= 0) {
            throw std::runtime_error("limit count must be larger than 0");
          }
          return limit_statement<TT, ARGS...>(m_engine, get_eval_sql_string(),
                                              count);
        }
        virtual row_collection<ARGS...> eval() {
          std::stringstream ss;
          ss << get_eval_sql_string() << ";";
          auto ret = m_engine->eval_sql_query_string(ss.str());
          return m_engine->template parse_records<ARGS...>(ret);
        }

      protected:
        std::string get_eval_sql_string() const {
          std::stringstream ss;
          ss << m_prev_sql << " order by " << CT::name << " " << ORDER::name;
          return ss.str();
        }

      protected:
        using statement<TT, ARGS...>::m_engine;
        std::string m_prev_sql;
      };

      template <typename TT, typename CST, typename... ARGS>
      struct cst_limit_statement : public statement<TT, ARGS...> {
      public:
        typedef typename TT::engine_type engine_type;
        cst_limit_statement(engine_type *engine, const std::string &prev_sql,
                            int64_t count, CST cst)
            : statement<TT, ARGS...>(engine), m_prev_sql(prev_sql),
              m_count(count), m_cst(cst) {}

        virtual row_collection<ARGS...> eval() {
          std::stringstream ss;
          ss << get_eval_sql_string();
          ss << "; ";
          auto native_statmenet = m_engine->prepare_sql_with_string(ss.str());
          int index = 0;
          traverse_cond_for_bind<TT, CST>::run(m_engine, native_statmenet,
                                               m_cst, index);
          auto ret = m_engine->eval_native_sql_stmt(native_statmenet);
          return m_engine->template parse_records<ARGS...>(ret);
        }

      protected:
        std::string get_eval_sql_string() const {
          std::stringstream ss;
          ss << m_prev_sql << " LIMIT " << m_count;
          return ss.str();
        }

      protected:
        using statement<TT, ARGS...>::m_engine;
        std::string m_prev_sql;
        int64_t m_count;
        CST m_cst;
      };

      template <typename TT, typename CT, typename ORDER, typename CST,
                typename... ARGS>
      struct cst_order_statement : public statement<TT, ARGS...> {
      public:
        typedef typename TT::engine_type engine_type;
        cst_order_statement(engine_type *engine, const std::string &prev_sql,
                            CST cst)
            : statement<TT, ARGS...>(engine), m_prev_sql(prev_sql), m_cst(cst) {
        }

        cst_limit_statement<TT, CST, ARGS...> limit(int64_t count) {
          if (count <= 0) {
            throw std::runtime_error("limit count must be larger than 0");
          }
          return cst_limit_statement<TT, CST, ARGS...>(
              m_engine, get_eval_sql_string(), count, m_cst);
        }
        virtual row_collection<ARGS...> eval() {
          std::stringstream ss;
          ss << get_eval_sql_string();
          ss << "; ";
          auto native_statmenet = m_engine->prepare_sql_with_string(ss.str());
          int index = 0;
          traverse_cond_for_bind<TT, CST>::run(m_engine, native_statmenet,
                                               m_cst, index);
          auto ret = m_engine->eval_native_sql_stmt(native_statmenet);
          return m_engine->template parse_records<ARGS...>(ret);
        }

      protected:
        std::string get_eval_sql_string() const {
          std::stringstream ss;
          ss << m_prev_sql << " order by " << CT::name << " " << ORDER::name;
          return ss.str();
        }

      protected:
        using statement<TT, ARGS...>::m_engine;
        std::string m_prev_sql;
        CST m_cst;
      };
      template <typename TT, typename CST, typename... ARGS>
      struct where_statement : public statement<TT, ARGS...> {
      public:
        typedef typename TT::engine_type engine_type;
        where_statement(engine_type *engine, const std::string &sql,
                        const CST &cst)
            : statement<TT, ARGS...>(engine), m_prev_sql(sql), m_cst(cst) {}

        virtual row_collection<ARGS...> eval() {
          std::stringstream ss;
          ss << get_eval_sql_string();
          ss << "; ";
          auto native_statmenet = m_engine->prepare_sql_with_string(ss.str());
          int index = 0;
          traverse_cond_for_bind<TT, CST>::run(m_engine, native_statmenet,
                                               m_cst, index);
          auto ret = m_engine->eval_native_sql_stmt(native_statmenet);
          return m_engine->template parse_records<ARGS...>(ret);
        }

        cst_limit_statement<TT, CST, ARGS...> limit(int64_t count) {
          if (count <= 0) {
            throw std::runtime_error("limit count must be larger than 0");
            return;
          }
          return cst_limit_statement<TT, CST, ARGS...>(
              m_engine, get_eval_sql_string(), count, m_cst);
        }

        template <typename CT, typename ORDER>
        cst_order_statement<TT, CT, ORDER, CST, ARGS...> order_by() {
          static_assert(util::is_contain_types<typename TT::cols_type,
                                               util::type_list<CT>>::value,
                        "Can't use rows that is not in table for order by");
          return cst_order_statement<TT, CT, ORDER, CST, ARGS...>(
              m_engine, get_eval_sql_string(), m_cst);
        }

      protected:
        std::string get_eval_sql_string() const {
          std::stringstream ss;
          ss << m_prev_sql << " where ";

          m_cst.dump_to_sql_string(ss);
          return ss.str();
        }

      protected:
        using statement<TT, ARGS...>::m_engine;
        CST m_cst;
        std::string m_prev_sql;
      };//end class where_statement

    template<typename TT>
      struct update_statement : public statement<TT>{
        public:
          typedef typename TT::engine_type engine_type;
          update_statement(engine_type *engine, const std::string &sql)
              : statement<TT>(engine), m_prev_sql(sql) {}
          virtual  row_collection<> eval(){
            throw std::bad_exception();
          }
        template <typename CST>
          where_statement<TT, CST> where(const CST & cst){
          static_assert(
              util::is_contain_types<typename TT::cols_type,
                                     typename CST::cols_type>::value,
              "Can't use rows that is not in table for *update where*");
          return where_statement<TT, CST>(m_engine, m_prev_sql, cst);
          }

        protected:
          using statement<TT>::m_engine;
          std::string m_prev_sql;
      };

    template<typename TT>
      struct delete_statement : public statement<TT>{
        public:
          typedef typename TT::engine_type engine_type;
          delete_statement(engine_type *engine, const std::string &sql)
              : statement<TT>(engine), m_prev_sql(sql) {}
          virtual  row_collection<> eval(){
            throw std::bad_exception();
          }
        template <typename CST>
          where_statement<TT, CST> where(const CST & cst){
          static_assert(
              util::is_contain_types<typename TT::cols_type,
                                     typename CST::cols_type>::value,
              "Can't use rows that is not in table for *delete where*");
          return where_statement<TT, CST>(m_engine, m_prev_sql, cst);
          }

        protected:
          using statement<TT>::m_engine;
          std::string m_prev_sql;
      };


      template <typename TT, typename... ARGS>
      struct select_statement : public statement<TT, ARGS...> {
      public:
        typedef typename TT::engine_type engine_type;
        select_statement(engine_type *engine)
            : statement<TT, ARGS...>(engine) {}
        template <typename CST>
        where_statement<TT, CST, ARGS...> where(const CST &cst) {
          static_assert(
              util::is_contain_types<typename TT::cols_type,
                                     typename CST::cols_type>::value,
              "Can't use rows that is not in table for *select where*");
          return where_statement<TT, CST, ARGS...>(m_engine,
                                                   get_eval_sql_string(), cst);
        }

        template <typename CT, typename ORDER>
        order_statement<TT, CT, ORDER, ARGS...> order_by() {
          static_assert(util::is_contain_types<typename TT::cols_type,
                                               util::type_list<CT>>::value,
                        "Can't use rows that is not in table for order by");
          return order_statement<TT, CT, ORDER, ARGS...>(m_engine,
                                                         get_eval_sql_string());
        }

        limit_statement<TT, ARGS...> limit(int64_t count) {
          if (count <= 0) {
            throw std::runtime_error("limit count must be larger than 0");
          }
          return limit_statement<TT, ARGS...>(m_engine, get_eval_sql_string(),
                                              count);
        }

        virtual row_collection<ARGS...> eval(){
          std::stringstream ss;
          ss<<get_eval_sql_string()<<";";
          auto ret = m_engine->eval_sql_query_string(ss.str());
          return m_engine->template parse_records<ARGS...>(ret);
        }

        protected:
        std::string get_eval_sql_string() const{
          std::stringstream ss;
          ss<<"select ";
          recursive_dump_col_name<ARGS...>(ss);
          ss<<" from "<<TT::meta_type::table_name;
          return ss.str();
        }
        template<typename T, typename T1, typename... TS>
          void recursive_dump_col_name(std::stringstream & ss)const {
            ss<<T::name<<", ";
            recursive_dump_col_name<T1, TS...>(ss);
          }
        template<typename T>
          void recursive_dump_col_name(std::stringstream & ss)const{
            ss<<T::name;
          }

          using statement<TT, ARGS...>::m_engine;
      };//end class select_statement

  }
}
template <typename T1, typename T2>
auto operator&&(const T1 &t1, const T2 &t2) ->
    typename std::enable_if<std::is_base_of<ff::sql::cond_stmt, T1>::value &&
                                std::is_base_of<ff::sql::cond_stmt, T2>::value,
                            ff::sql::and_cond_stmt<T1, T2>>::type {
  return ff::sql::and_cond_stmt<T1, T2>(t1, t2);
         }

         template <typename T1, typename T2>
         auto operator||(const T1 &t1, const T2 &t2) -> typename std::enable_if<
             std::is_base_of<ff::sql::cond_stmt, T1>::value &&
                 std::is_base_of<ff::sql::cond_stmt, T2>::value,
             ff::sql::or_cond_stmt<T1, T2>>::type {
           return ff::sql::or_cond_stmt<T1, T2>(t1, t2);
         }
