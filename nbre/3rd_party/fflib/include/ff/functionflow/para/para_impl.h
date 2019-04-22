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
#ifndef FF_PARA_PARA_IMPL_H_
#define FF_PARA_PARA_IMPL_H_

#include "ff/functionflow/common/common.h"
#include "ff/functionflow/runtime/rtcmn.h"
#include "ff/functionflow/runtime/runtime.h"
#include "ff/functionflow/runtime/taskbase.h"

namespace ff {
namespace internal {
template <class RT>
class para_impl_base;

template <class T>
class para_ret {
 public:
  para_ret(para_impl_base<T>& p) : m_refP(p), m_oValue() {}

  T& get() { return m_oValue; }
  void set(T& v) { m_oValue = v; }
  void set(T&& v) { m_oValue = v; }

 protected:
  para_impl_base<T>& m_refP;
  T m_oValue;
};  // end class para_ret;

template <class RT>
class para_impl_base : public ff::rt::task_base {
 public:
  para_impl_base()
      : ff::rt::task_base(), m_oRet(*this), m_iES(exe_state::exe_init) {}

  virtual ~para_impl_base() {}
  // virtual void	run(){}

  exe_state get_state() { return m_iES.load(); }
  bool check_if_over() {
    if (m_iES.load() == exe_state::exe_over) return true;
    return false;
  }
  RT& get() { return m_oRet.get(); }

 protected:
  para_ret<RT> m_oRet;
  std::atomic<exe_state> m_iES;
};

template <class RT>
class para_impl : public para_impl_base<RT> {
 public:
  template <class F>
  para_impl(F&& f)
      : para_impl_base<RT>(), m_oFunc(std::move(f)) {}

  virtual void run() {
    m_iES.store(exe_state::exe_run);
    m_oRet.set(m_oFunc());
    m_iES.store(exe_state::exe_over);
  }

 protected:
  using para_impl_base<RT>::m_iES;
  using para_impl_base<RT>::m_oRet;
  std::function<RT()> m_oFunc;
};  // end class para_impl

template <class RT>
using para_impl_base_ptr = std::shared_ptr<para_impl_base<RT>>;

template <class RT>
using para_impl_ptr = std::shared_ptr<para_impl<RT>>;

template <class ret_type, class F>
auto make_para_impl(F&& f) ->
    typename std::enable_if<std::is_void<ret_type>::value,
                            internal::para_impl_ptr<ret_type>>::type {
  auto p = std::make_shared<internal::para_impl<ret_type>>(std::forward<F>(f));
  return p;
}
template <class ret_type, class F>
auto make_para_impl(F&& f) ->
    typename std::enable_if<!std::is_void<ret_type>::value,
                            internal::para_impl_ptr<ret_type>>::type {
  auto p = std::make_shared<internal::para_impl<ret_type>>(std::forward<F>(f));
  return p;
}

template <class RT, class WT, class FT>
class para_impl_wait : public para_impl_base<RT> {
 public:
  para_impl_wait(WT& w, FT&& f)
      : para_impl_base<RT>(), m_pFunc(f), m_oWaitingPT(w) {}
  virtual ~para_impl_wait() {}
  virtual void run() {
    m_iES = exe_state::exe_run;
    if (m_oWaitingPT.get_state() != exe_state::exe_over) {
      ff::rt::task_base::need_to_reschedule() = true;
      return;
    }
    ff::rt::task_base::need_to_reschedule() = false;
    m_oRet.set(m_oWaitingPT.internal_then(m_pFunc));
    m_iES.store(exe_state::exe_over);
  }
  RT& get() { return m_oRet.get(); }

 protected:
  typedef typename std::remove_reference<FT>::type FT_t;
  using para_impl_base<RT>::m_iES;
  using para_impl_base<RT>::m_oRet;
  FT_t m_pFunc;
  WT m_oWaitingPT;
};  // end class para_impl_wait_base;

template <class RT, class WT, class FT>
using para_impl_wait_ptr = std::shared_ptr<para_impl_wait<RT, WT, FT>>;

template <class RT>
void schedule(para_impl_base_ptr<RT> p) {
  ::ff::rt::schedule(std::dynamic_pointer_cast<ff::rt::task_base>(p));
}
#include "ff/functionflow/para/para_impl_void.h"
}  // end namespace internal
}  // end namespace ff
#endif
