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
#ifndef FF_PARA_PARA_GROUND_H_
#define FF_PARA_PARA_GROUND_H_
#include "ff/functionflow/common/common.h"
#include "ff/functionflow/common/error_msg.h"
#include "ff/functionflow/para/para.h"
#include "ff/functionflow/para/para_helper.h"
#include "ff/functionflow/para/paras_with_lock.h"
#include "ff/functionflow/para/wait_impl.h"
#include "ff/functionflow/runtime/rtcmn.h"
#include "ff/util/func_type_checker.h"
#include "ff/util/is_callable.h"

#include <cmath>
#include <algorithm>

namespace ff {

namespace internal {
class wait_all;
class wait_any;
}  // end namespace internal
struct auto_partitioner {};
struct simple_partitioner {};

class paragroup {
 public:
  typedef void ret_type;

 public:
  template <class PT, class WT>
  class para_accepted_wait {
    para_accepted_wait &operator=(const para_accepted_wait<PT, WT> &) = delete;

   public:
    typedef typename PT::ret_type ret_type;
    para_accepted_wait(const para_accepted_wait<PT, WT> &) = default;
    para_accepted_wait(PT &p, const WT &w) : m_refP(p), m_oWaiting(w) {}

    //! For arithmetic iterators
    template <class Iterator_t, class Functor_t>
    auto for_each(Iterator_t begin, Iterator_t end, Functor_t &&f) ->
        typename std::enable_if<
            std::is_arithmetic<
                typename std::remove_cv<Iterator_t>::type>::value,
            ::ff::internal::para_accepted_call<paragroup, void>>::type {
      FF_DEFAULT_PARTITIONER *p = nullptr;
      for_each_impl(begin, end, std::forward<Functor_t>(f), m_refP.m_pEntities,
                    p);
      // for_each_impl_general_iterator(begin, end, std::forward<Functor_t>(f),
      // m_refP.m_pEntities, p);
      return ::ff::internal::para_accepted_call<paragroup, ret_type>(m_refP);
    }

    //! For random-access iterators
    template <class Iterator_t, class Functor_t>
    auto for_each(Iterator_t begin, Iterator_t end, Functor_t &&f) ->
        typename std::enable_if<
            !std::is_arithmetic<
                typename std::remove_cv<Iterator_t>::type>::value &&
                std::is_same<typename std::iterator_traits<
                                 Iterator_t>::iterator_category,
                             std::random_access_iterator_tag>::value,
            ::ff::internal::para_accepted_call<paragroup, void>>::type {
      FF_DEFAULT_PARTITIONER *p = nullptr;
      for_each_impl(begin, end, [f](const Iterator_t &t) {
        // for_each_impl_general_iterator(begin, end, [f](const Iterator_t & t)
        // {
        f(*t);
      }, m_refP.m_pEntities, p);
      return ::ff::internal::para_accepted_call<paragroup, ret_type>(m_refP);
    }

    //! For general iterators
    template <class Iterator_t, class Functor_t>
    auto for_each(Iterator_t begin, Iterator_t end, Functor_t &&f) ->
        typename std::enable_if<
            !std::is_arithmetic<
                typename std::remove_cv<Iterator_t>::type>::value &&
                !std::is_same<typename std::iterator_traits<
                                  Iterator_t>::iterator_category,
                              std::random_access_iterator_tag>::value,
            ::ff::internal::para_accepted_call<paragroup, void>>::type {
      FF_DEFAULT_PARTITIONER *p = nullptr;
      for_each_impl_general_iterator(begin, end, [f](const Iterator_t &t) {
        f(*t);
      }, m_refP.m_pEntities, p);
      return ::ff::internal::para_accepted_call<paragroup, ret_type>(m_refP);
    }

   protected:
    PT &m_refP;
    WT m_oWaiting;
  };  // end class para_accepted_wait;

  template <class WT>
  para_accepted_wait<paragroup, WT> operator[](WT &&cond) {
    return para_accepted_wait<paragroup, WT>(*this, std::forward<WT>(cond));
  }
  //! For input index, return the corresponding task
  para<void> &operator[](int index) { return (*m_pEntities).entities[index]; }

  size_t size() const {
    if (m_pEntities) return m_pEntities->entities.size();
    return 0;
  }

  ~paragroup() {
    ::ff::internal::wait_all wa(m_pEntities);
    wa.then([]() {});
  }

  //! For arithmetic iterator
  template <class Iterator_t, class Functor_t>
  auto for_each(Iterator_t begin, Iterator_t end, Functor_t &&f) ->
      typename std::enable_if<
          std::is_arithmetic<
              typename std::remove_cv<Iterator_t>::type>::value &&
              ff::util::is_function_with_arg_type<Functor_t, Iterator_t>::value,
          ::ff::internal::para_accepted_call<paragroup, void>>::type {
    FF_DEFAULT_PARTITIONER *p = nullptr;
    for_each_impl(begin, end, std::forward<Functor_t>(f), m_pEntities, p);
    // for_each_impl_general_iterator(begin, end, std::forward<Functor_t>(f),
    // m_pEntities, p);
    return ::ff::internal::para_accepted_call<paragroup, ret_type>(*this);
  }

  //! For random-access iterator
  template <class Iterator_t, class Functor_t>
  auto for_each(Iterator_t begin, Iterator_t end, Functor_t &&f) ->
      typename std::enable_if<
          !std::is_arithmetic<
              typename std::remove_cv<Iterator_t>::type>::value &&
              std::is_same<
                  typename std::iterator_traits<Iterator_t>::iterator_category,
                  std::random_access_iterator_tag>::value &&
              ff::util::is_function_with_arg_type<
                  Functor_t,
                  typename std::iterator_traits<Iterator_t>::value_type>::value,
          ::ff::internal::para_accepted_call<paragroup, void>>::type {
    FF_DEFAULT_PARTITIONER *p = nullptr;
    for_each_impl(begin, end, [f](const Iterator_t &t) {
      // for_each_impl_general_iterator(begin, end, [f](const Iterator_t & t) {
      f(*t);
    }, m_pEntities, p);
    return ::ff::internal::para_accepted_call<paragroup, ret_type>(*this);
  }

  //! For general iterator
  template <class Iterator_t, class Functor_t>
  auto for_each(Iterator_t begin, Iterator_t end, Functor_t &&f) ->
      typename std::enable_if<
          !std::is_arithmetic<
              typename std::remove_cv<Iterator_t>::type>::value &&
              !std::is_same<
                  typename std::iterator_traits<Iterator_t>::iterator_category,
                  std::random_access_iterator_tag>::value &&
              ff::util::is_function_with_arg_type<
                  Functor_t,
                  typename std::iterator_traits<Iterator_t>::value_type>::value,
          ::ff::internal::para_accepted_call<paragroup, void>>::type {
    FF_DEFAULT_PARTITIONER *p = nullptr;
    for_each_impl_general_iterator(
        begin, end, [f](const Iterator_t &t) { f(*t); }, m_pEntities, p);
    return ::ff::internal::para_accepted_call<paragroup, ret_type>(*this);
  }

  template <class Iterator_t, class Functor_t>
  auto for_each(Iterator_t begin, Iterator_t end, Functor_t &&f) ->
      typename std::enable_if<
          !ff::util::is_callable<Functor_t>::value,
          ::ff::internal::para_accepted_call<paragroup, void>>::type {
    static_assert(Please_Check_The_Assert_Msg<Functor_t>::value,
                  FF_EM_CALL_FOR_EACH_WITHOUT_FUNCTION);
  }

  template <class Iterator_t, class Functor_t>
  auto for_each(Iterator_t begin, Iterator_t end, Functor_t &&f) ->
      typename std::enable_if<
          ff::util::is_callable<Functor_t>::value &&
              std::is_arithmetic<
                  typename std::remove_cv<Iterator_t>::type>::value &&
              !ff::util::is_function_with_arg_type<Functor_t,
                                                   Iterator_t>::value,
          ::ff::internal::para_accepted_call<paragroup, void>>::type {
    static_assert(Please_Check_The_Assert_Msg<Functor_t>::value,
                  FF_EM_CALL_FOR_EACH_WRONG_FUNCTION);
  }

  template <class Iterator_t, class Functor_t>
  auto for_each(Iterator_t begin, Iterator_t end, Functor_t &&f) ->
      typename std::enable_if<
          ff::util::is_callable<Functor_t>::value &&
              !ff::util::is_function_with_arg_type<
                  Functor_t,
                  typename std::iterator_traits<Iterator_t>::value_type>::value,
          ::ff::internal::para_accepted_call<paragroup, void>>::type {
    static_assert(Please_Check_The_Assert_Msg<Functor_t>::value,
                  FF_EM_CALL_FOR_EACH_WRONG_FUNCTION);
  }

  void clear() { m_pEntities.reset(); }
  template <class T>
  void add(T &&t) {
    static_assert(Please_Check_The_Assert_Msg<T>::value,
                  FF_EM_USE_PARACONTAINER_INSTEAD_OF_GROUP);
  }

 protected:
   typedef std::shared_ptr<::ff::internal::paras_with_lock> Entities_t;

#include "ff/functionflow/para/paragroup_general_iterator_impl.h"

#include "ff/functionflow/para/paragroup_random_access_iterator.h"

 protected:
   friend ::ff::internal::wait_all all(paragroup &pg);
   friend ::ff::internal::wait_any any(paragroup &pg);
   std::shared_ptr<::ff::internal::paras_with_lock> &all_entities() {
     return m_pEntities;
   };

 protected:
   std::shared_ptr<::ff::internal::paras_with_lock> m_pEntities;
};  // end class paragroup

}  // end namespace ff

#endif
