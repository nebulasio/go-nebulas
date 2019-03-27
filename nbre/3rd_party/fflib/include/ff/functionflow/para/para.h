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
#include "ff/functionflow/common/common.h"
#include "ff/functionflow/para/bin_wait_func_deducer.h"
#include "ff/functionflow/para/exception.h"
#include "ff/functionflow/para/is_wait_compatible_with_then.h"
#include "ff/functionflow/para/para_helper.h"
#include "ff/functionflow/para/para_impl.h"
#include "ff/functionflow/para/para_wait_traits.h"
#include "ff/functionflow/runtime/rtcmn.h"
#include "ff/util/function_traits.h"

namespace ff {

using namespace ff::util;
namespace internal {
template <typename DT, typename RT>
class para_common {
 public:
  typedef RT ret_type;

 public:
#include "ff/functionflow/para/para_accepted_wait.h"
   para_common() : m_pImpl(nullptr){};
   ~para_common() {}
   template <class WT>
   auto operator[](WT &&cond) -> typename std::enable_if<
       is_para_or_wait<typename std::remove_reference<WT>::type>::value,
       para_accepted_wait<DT, WT>>::type {
     if (cond.get_state() == exe_state::exe_empty)
       throw empty_para_exception();
     return para_accepted_wait<DT, WT>(*(static_cast<DT *>(this)),
                                       std::forward<WT>(cond));
   }
   template <class WT>
   auto operator[](WT &&cond) -> typename std::enable_if<
       !is_para_or_wait<typename std::remove_reference<WT>::type>::value,
       para_accepted_wait<DT, para<void>>>::type {
     static_assert(Please_Check_The_Assert_Msg<WT>::value,
                   FF_EM_WRONG_USE_SQPAREN);
   }
   template <class F> auto exe(F &&f) -> para_accepted_call<DT, ret_type> {
     if (m_pImpl)
       throw used_para_exception();
     auto pp = make_para_impl<ret_type>(std::forward<F>(f));
     m_pImpl = std::dynamic_pointer_cast<para_impl_base<ret_type>>(pp);
     schedule(m_pImpl);
     return para_accepted_call<DT, ret_type>(*(static_cast<DT *>(this)));
   }
   template <class F>
   auto operator()(F &&f) -> typename std::enable_if<
       std::is_same<typename ::ff::util::function_res_traits<F>::ret_type,
                    ret_type>::value,
       para_accepted_call<DT, ret_type>>::type {
     return exe(std::forward<F>(f));
   }

   template <class F>
   auto operator()(F &&f) -> typename std::enable_if<
       !std::is_same<typename ::ff::util::function_res_traits<F>::ret_type,
                     ret_type>::value,
       para_accepted_call<DT, ret_type>>::type {
     static_assert(Please_Check_The_Assert_Msg<F>::value,
                   FF_EM_CALL_WITH_TYPE_MISMATCH);
   }
   exe_state get_state() {
     if (m_pImpl)
       return m_pImpl->get_state();
     return exe_state::exe_empty;
   }
   bool check_if_over() {
     if (m_pImpl)
       return m_pImpl->check_if_over();
     return false;
   }

   ::ff::internal::para_impl_base_ptr<ret_type> get_internal_impl() {
     return m_pImpl;
   }

   template <class F> void then(const F &) {
     static_assert(Please_Check_The_Assert_Msg<F>::value,
                   FF_EM_CALL_THEN_WITHOUT_CALL_PAREN);
   }

 protected:
   ::ff::internal::para_impl_base_ptr<ret_type> m_pImpl;
};  // end class para_common

}  // end namespace internal

template <typename RT = void>
class para : public ::ff::internal::para_common<para<RT>, RT> {
public:
  typedef RT ret_type;
  auto get() -> typename std::enable_if<!std::is_void<RT>::value, RT>::type & {
    return ::ff::internal::para_common<para<RT>, RT>::m_pImpl->get();
  }
  template <class FT>
  auto internal_then(FT &&f) -> typename std::enable_if<
      std::is_void<typename function_res_traits<FT>::ret_type>::value &&
          is_function_with_arg_type<FT, ret_type>::value,
      void>::type {
    f(m_pImpl->get());
  }

  template <class FT>
  auto internal_then(FT &&f) -> typename std::enable_if<
      is_function_with_arg_type<FT, RT>::value,
      typename std::remove_reference<
          typename function_res_traits<FT>::ret_type>::type &&>::type {
    return std::move(f(m_pImpl->get()));
  }

 protected:
   using ::ff::internal::para_common<para<RT>, RT>::m_pImpl;
};  // end class para;

template <>
class para<void> : public ::ff::internal::para_common<para<void>, void> {
public:
  typedef void ret_type;
  template <class FT>
  auto internal_then(FT &&f) -> typename std::enable_if<
      std::is_void<typename function_res_traits<FT>::ret_type>::value &&
          is_function_with_arg_type<FT, void>::value,
      void>::type {
    f();
  }

  template <class FT>
  auto internal_then(FT &&f) -> typename std::enable_if<
      is_function_with_arg_type<FT, void>::value &&
          !std::is_void<typename function_res_traits<FT>::ret_type>::value,
      typename std::remove_reference<
          typename function_res_traits<FT>::ret_type>::type>::type {
    return f();
  }
};  // end class para;

template <class T>
class para<para<T>> {
 public:
  para() {
    static_assert(Please_Check_The_Assert_Msg<para<T>>::value,
                  FF_EM_CALL_NO_SUPPORT_FOR_PARA);
  };
};  // end class para;

}  // end namespace ff
