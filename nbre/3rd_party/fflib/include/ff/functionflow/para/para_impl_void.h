#include "ff/functionflow/runtime/taskbase.h"

template <>
class para_impl_base<void> : public ff::rt::task_base {
 public:
  para_impl_base() : ff::rt::task_base(), m_iES(exe_state::exe_init) {}

  virtual ~para_impl_base() {}
  // virtual void	run(){}

  exe_state get_state() { return m_iES.load(); }
  bool check_if_over() {
    if (m_iES.load() == exe_state::exe_over) return true;
    return false;
  }

 protected:
  std::atomic<exe_state> m_iES;
};

template <>
class para_impl<void> : public para_impl_base<void> {
 public:
  template <class F>
  para_impl(F&& f)
      : para_impl_base<void>(), m_oFunc(std::forward<F>(f)) {}

  virtual ~para_impl() {}

  virtual void run() {
    m_iES.store(exe_state::exe_run);
    m_oFunc();
    m_iES.store(exe_state::exe_over);
  }

 protected:
  using para_impl_base<void>::m_iES;
  std::function<void()> m_oFunc;
};  // end class para_impl

template <class WT, class FT>
class para_impl_wait<void, WT, FT> : public para_impl_base<void> {
 public:
  para_impl_wait(WT& w, FT&& f)
      : para_impl_base<void>(), m_pFunc(f), m_oWaitingPT(w) {}
  virtual ~para_impl_wait() {}
  virtual void run() {
    m_iES = exe_state::exe_run;
    if (m_oWaitingPT.get_state() != exe_state::exe_over) {
      ff::rt::task_base::need_to_reschedule() = true;
      return;
    }
    ff::rt::task_base::need_to_reschedule() = false;
    m_oWaitingPT.internal_then(m_pFunc);
    m_iES.store(exe_state::exe_over);
  }

 protected:
  typedef typename std::remove_reference<FT>::type FT_t;
  using para_impl_base<void>::m_iES;
  FT_t m_pFunc;
  WT m_oWaitingPT;
};
