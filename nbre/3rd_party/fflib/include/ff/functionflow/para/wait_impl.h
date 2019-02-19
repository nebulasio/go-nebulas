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
#ifndef FF_PARA_WAIT_IMPL_H_
#define FF_PARA_WAIT_IMPL_H_
#include "ff/functionflow/common/common.h"
#include "ff/functionflow/para/bin_wait_func_deducer.h"
#include "ff/functionflow/para/para.h"
#include "ff/functionflow/para/paracontainer.h"
#include "ff/util/tuple_type.h"

namespace ff {
template <class RT>
class para;
namespace internal {
template <class T1, class T2>
class wait_and {
 public:
  typedef typename std::remove_reference<T1>::type T1_t;
  typedef typename std::remove_reference<T2>::type T2_t;
  typedef typename T1_t::ret_type RT1_t;
  typedef typename T2_t::ret_type RT2_t;
  typedef bin_wait_func_deducer<typename T1_t::ret_type,
                                typename T2_t::ret_type> deduct_t;
  typedef typename deduct_t::and_type ret_type;

 public:
  wait_and(T1&& t1, T2&& t2) : m_1(t1), m_2(t2), m_iES(exe_state::exe_init) {}

  template <class FT>
  auto then(FT &&f) -> typename std::enable_if<
      std::is_void<typename function_res_traits<FT>::ret_type>::value &&
          is_compatible_then<FT, RT1_t, RT2_t>::is_cpt_with_and &&
          !util::function_args_traits<FT>::is_no_args,
      void>::type {
    if (!check_if_over()) {
      ::ff::rt::yield_and_ret_until([this]() { return check_if_over(); });
    }
    deduct_t::void_func_and(std::forward<FT>(f), m_1, m_2);
  }

  template <class FT>
  auto then(FT &&f) -> typename std::enable_if<
      !std::is_void<typename function_res_traits<FT>::ret_type>::value &&
          is_compatible_then<FT, RT1_t, RT2_t>::is_cpt_with_and &&
          !util::function_args_traits<FT>::is_no_args,
      typename std::remove_reference<
          typename function_res_traits<FT>::ret_type>::type>::type {
    if (!check_if_over())
      ::ff::rt::yield_and_ret_until([this]() { return check_if_over(); });
    return deduct_t::ret_func_and(std::forward<FT>(f), m_1, m_2);
  }

  template <class FT>
  auto then(FT &&f) -> typename std::enable_if<
      std::is_void<typename function_res_traits<FT>::ret_type>::value &&
          util::function_args_traits<FT>::is_no_args,
      void>::type {
    bool b = check_if_over();
    if (!b)
      ::ff::rt::yield_and_ret_until([this]() { return check_if_over(); });
    f();
  }

  template <class FT>
  auto then(FT &&f) -> typename std::enable_if<
      !std::is_void<typename function_res_traits<FT>::ret_type>::value &&
          util::function_args_traits<FT>::is_no_args,
      typename std::remove_reference<
          typename function_res_traits<FT>::ret_type>::type>::type {
    bool b = check_if_over();
    if (!b) ::ff::rt::yield_and_ret_until([this]() { return check_if_over(); });
    return f();
  }

  template <class FT>
  auto then(FT &&f) -> typename std::enable_if<
      !util::function_args_traits<FT>::is_no_args &&
          !is_compatible_then<FT, RT1_t, RT2_t>::is_cpt_with_and,
      typename std::remove_reference<
          typename function_res_traits<FT>::ret_type>::type>::type {
    static_assert(Please_Check_The_Assert_Msg<FT>::value,
                  FF_EM_THEN_FUNC_TYPE_MISMATCH);
  }

  template <class FT>
  auto internal_then(FT &&f) -> typename std::enable_if<
      std::is_void<typename function_res_traits<FT>::ret_type>::value &&
          is_compatible_then<FT, RT1_t, RT2_t>::is_cpt_with_and &&
          !util::function_args_traits<FT>::is_no_args,
      void>::type {
    deduct_t::void_func_and(std::forward<FT>(f), m_1, m_2);
  }

  template <class FT>
  auto internal_then(FT &&f) -> typename std::enable_if<
      !std::is_void<typename function_res_traits<FT>::ret_type>::value &&
          is_compatible_then<FT, RT1_t, RT2_t>::is_cpt_with_and &&
          !util::function_args_traits<FT>::is_no_args,
      typename std::remove_reference<
          typename function_res_traits<FT>::ret_type>::type>::type {
    return deduct_t::ret_func_and(std::forward<FT>(f), m_1, m_2);
  }

  template <class FT>
  auto internal_then(FT &&f) -> typename std::enable_if<
      std::is_void<typename function_res_traits<FT>::ret_type>::value &&
          util::function_args_traits<FT>::is_no_args,
      void>::type {
    f();
  }

  template <class FT>
  auto internal_then(FT &&f) -> typename std::enable_if<
      !std::is_void<typename function_res_traits<FT>::ret_type>::value &&
          util::function_args_traits<FT>::is_no_args,
      typename std::remove_reference<
          typename function_res_traits<FT>::ret_type>::type>::type {
    return f();
  }

  auto get() -> typename deduct_t::and_type {
    return deduct_t::wrap_ret_for_and(m_1, m_2);
  }

  exe_state get_state() {
    if (m_iES != exe_state::exe_over)
      m_iES = exe_state_and(m_1.get_state(), m_2.get_state());
    return m_iES;
  }
  bool check_if_over() {
    if (m_iES == exe_state::exe_over) return true;
    m_iES = exe_state_and(m_1.get_state(), m_2.get_state());
    if (m_iES == exe_state::exe_over) return true;
    return false;
  }

 protected:
  T1_t m_1;
  T2_t m_2;
  exe_state m_iES;
};  // end class wait_and

template <class T1, class T2>
class wait_or {
 public:
  typedef typename std::remove_reference<T1>::type T1_t;
  typedef typename std::remove_reference<T2>::type T2_t;
  typedef typename T1_t::ret_type RT1_t;
  typedef typename T2_t::ret_type RT2_t;
  typedef bin_wait_func_deducer<typename T1_t::ret_type,
                                typename T2_t::ret_type> deduct_t;
  typedef typename deduct_t::or_type ret_type;

 public:
  wait_or(T1&& t1, T2&& t2) : m_1(t1), m_2(t2), m_iES(exe_state::exe_init) {}

  template <class FT>
  auto then(FT &&f) -> typename std::enable_if<
      std::is_void<typename function_res_traits<FT>::ret_type>::value &&
          is_compatible_then<FT, RT1_t, RT2_t>::is_cpt_with_or &&
          !util::function_args_traits<FT>::is_no_args,
      void>::type {
    bool b = check_if_over();
    if (!b) ::ff::rt::yield_and_ret_until([this]() { return check_if_over(); });
    deduct_t::void_func_or(std::forward<FT>(f), m_1, m_2);
  }

  template <class FT>
  auto then(FT &&f) -> typename std::enable_if<
      !std::is_void<typename function_res_traits<FT>::ret_type>::value &&
          is_compatible_then<FT, RT1_t, RT2_t>::is_cpt_with_or &&
          !util::function_args_traits<FT>::is_no_args,
      typename std::remove_reference<
          typename function_res_traits<FT>::ret_type>::type>::type {
    bool b = check_if_over();
    if (!b) ::ff::rt::yield_and_ret_until([this]() { return check_if_over(); });
    return deduct_t::ret_func_or(std::forward<FT>(f), m_1, m_2);
  }

  template <class FT>
  auto then(FT &&f) -> typename std::enable_if<
      std::is_void<typename function_res_traits<FT>::ret_type>::value &&
          util::function_args_traits<FT>::is_no_args,
      void>::type {
    bool b = check_if_over();
    if (!b)
      ::ff::rt::yield_and_ret_until([this]() { return check_if_over(); });
    f();
  }

  template <class FT>
  auto then(FT &&f) -> typename std::enable_if<
      !std::is_void<typename function_res_traits<FT>::ret_type>::value &&
          util::function_args_traits<FT>::is_no_args,
      typename std::remove_reference<
          typename function_res_traits<FT>::ret_type>::type>::type {
    bool b = check_if_over();
    if (!b) ::ff::rt::yield_and_ret_until([this]() { return check_if_over(); });
    return f();
  }

  template <class FT>
  auto then(FT &&f) -> typename std::enable_if<
      !util::is_no_args_function<FT>::value &&
          !is_compatible_then<FT, RT1_t, RT2_t>::is_cpt_with_or,
      typename std::remove_reference<
          typename function_res_traits<FT>::ret_type>::type>::type {
    static_assert(Please_Check_The_Assert_Msg<FT>::value,
                  FF_EM_THEN_FUNC_TYPE_MISMATCH);
  }

  template <class FT>
  auto internal_then(FT &&f) -> typename std::enable_if<
      std::is_void<typename function_res_traits<FT>::ret_type>::value &&
          is_compatible_then<FT, RT1_t, RT2_t>::is_cpt_with_or &&
          !util::function_args_traits<FT>::is_no_args,
      void>::type {
    deduct_t::void_func_or(std::forward<FT>(f), m_1, m_2);
  }

  template <class FT>
  auto internal_then(FT &&f) -> typename std::enable_if<
      !std::is_void<typename function_res_traits<FT>::ret_type>::value &&
          is_compatible_then<FT, RT1_t, RT2_t>::is_cpt_with_or &&
          !util::function_args_traits<FT>::is_no_args,
      typename std::remove_reference<
          typename function_res_traits<FT>::ret_type>::type>::type {
    return deduct_t::ret_func_or(std::forward<FT>(f), m_1, m_2);
  }

  template <class FT>
  auto internal_then(FT &&f) -> typename std::enable_if<
      std::is_void<typename function_res_traits<FT>::ret_type>::value &&
          util::function_args_traits<FT>::is_no_args,
      void>::type {
    f();
  }

  template <class FT>
  auto internal_then(FT &&f) -> typename std::enable_if<
      !std::is_void<typename function_res_traits<FT>::ret_type>::value &&
          util::function_args_traits<FT>::is_no_args,
      typename std::remove_reference<
          typename function_res_traits<FT>::ret_type>::type>::type {
    return f();
  }

  auto get() -> typename deduct_t::or_type {
    return deduct_t::wrap_ret_for_or(m_1, m_2);
  }

  exe_state get_state() {
    if (m_iES != exe_state::exe_over)
      m_iES = exe_state_or(m_1.get_state(), m_2.get_state());
    return m_iES;
  }
  bool check_if_over() {
    if (m_iES == exe_state::exe_over) return true;
    m_iES = exe_state_or(m_1.get_state(), m_2.get_state());
    if (m_iES == exe_state::exe_over) return true;
    return false;
  }

 protected:
  T1_t m_1;
  T2_t m_2;
  exe_state m_iES;
};  // end class wait_or

class wait_all {
 public:
  typedef void ret_type;
  wait_all(std::shared_ptr<::ff::internal::paras_with_lock> ps);

  template <class FT>
  auto then(FT&& f) -> void {
    bool b = check_if_over();
    if (!b) ::ff::rt::yield_and_ret_until([this]() { return check_if_over(); });
    f();
  }

  exe_state get_state();
  bool check_if_over();

 protected:
   std::shared_ptr<::ff::internal::paras_with_lock> all_ps;
   exe_state m_iES;
};  // end class wait_all

class wait_any {
 public:
  typedef void ret_type;
  wait_any(std::shared_ptr<::ff::internal::paras_with_lock> ps);

  template <class FT>
  auto then(FT&& f) -> void {
    bool b = check_if_over();
    if (!b) ::ff::rt::yield_and_ret_until([this]() { return check_if_over(); });
    f();
  }
  exe_state get_state();
  bool check_if_over();

 protected:
   std::shared_ptr<::ff::internal::paras_with_lock> all_ps;
   exe_state m_iES;
};  // end class wait_any

}  // end namespace internal
}  // end namespace ff

#endif
