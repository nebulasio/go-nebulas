#ifndef FF_UTILITIES_THREAD_LOCAL_VAR_H_
#define FF_UTILITIES_THREAD_LOCAL_VAR_H_
#include "ff/common/common.h"
#include "ff/runtime/rtcmn.h"
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
  inline auto for_each(FT&& f) ->
      typename std::enable_if<utils::is_function_with_arg_type<FT, T>::value,
                              void>::type {
    for (size_t i = 0; i < m_oVars.size(); ++i) {
      f(m_oVars[i]);
    }
  }

  template <typename F>
  auto for_each(F&& f) ->
      typename std::enable_if<!utils::is_callable<F>::value, void>::type {
    static_assert(Please_Check_The_Assert_Msg<F>::value,
                  FF_EM_CALL_FOR_EACH_WITH_NO_FUNC);
  }

  template <typename F>
  auto for_each(F&& f) -> typename std::enable_if<
      utils::is_callable<F>::value &&
          !utils::is_function_with_arg_type<F, T>::value,
      void>::type {
    static_assert(Please_Check_The_Assert_Msg<F>::value,
                  FF_EM_CALL_FOR_EACH_WITH_WRONG_PARAM);
  }

 protected:
  std::vector<T> m_oVars;
};  // end class thread_local_var
}  // end namespace ff

#endif
