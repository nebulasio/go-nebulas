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
#ifndef FF_UTILITIES_THREAD_LOCAL_VAR_H_
#define FF_UTILITIES_THREAD_LOCAL_VAR_H_
#include "ff/functionflow/common/common.h"
#include "ff/functionflow/runtime/rtcmn.h"
#include <vector>

namespace ff {
template <typename T>
class thread_local_var {
 public:
  thread_local_var() {
    for (size_t i = 0; i < rt::concurrency(); ++i) {
      m_oVars.push_back(T());
    }
  }
  thread_local_var(const T& t) {
    for (size_t i = 0; i < rt::concurrency(); ++i) {
      m_oVars.push_back(T());
    }
  }

  inline T& current() {
    thread_local static thrd_id_t id = ff::rt::get_thrd_id();
    return m_oVars[id];
  }

  inline const T& current() const {
    thread_local static thrd_id_t id = ff::rt::get_thrd_id();
    return m_oVars[id];
  }

  inline void reset() const {
    for (size_t i = 0; i < rt::concurrency(); ++i) {
      m_oVars[i] = T();
    }
  }

  template <typename FT>
  inline auto for_each(FT &&f) ->
      typename std::enable_if<util::is_function_with_arg_type<FT, T>::value,
                              void>::type {
    for (size_t i = 0; i < m_oVars.size(); ++i) {
      f(m_oVars[i]);
    }
  }

  template <typename F>
  auto for_each(F &&f) ->
      typename std::enable_if<!util::is_callable<F>::value, void>::type {
    static_assert(Please_Check_The_Assert_Msg<F>::value,
                  FF_EM_CALL_FOR_EACH_WITH_NO_FUNC);
  }

  template <typename F>
  auto for_each(F &&f) ->
      typename std::enable_if<util::is_callable<F>::value &&
                                  !util::is_function_with_arg_type<F, T>::value,
                              void>::type {
    static_assert(Please_Check_The_Assert_Msg<F>::value,
                  FF_EM_CALL_FOR_EACH_WITH_WRONG_PARAM);
  }

 protected:
  std::vector<T> m_oVars;
};  // end class thread_local_var
}  // end namespace ff

#endif
