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
#include "ff/functionflow/para/wait.h"
#include "ff/functionflow/common/common.h"
#include "ff/functionflow/para/wait_impl.h"

namespace ff {
namespace internal {

wait_all::wait_all(std::shared_ptr<::ff::internal::paras_with_lock> ps)
    : all_ps(ps), m_iES(exe_state::exe_init){};

exe_state wait_all::get_state() {
  if (!all_ps) m_iES = exe_state::exe_over;
  if (m_iES != exe_state::exe_over) {
    m_iES = exe_state::exe_over;
    all_ps->lock.lock();
    for (auto p = all_ps->entities.begin(); p != all_ps->entities.end(); ++p) {
      m_iES = exe_state_and(m_iES, p->get_state());
    }
    all_ps->lock.unlock();
  }
  return m_iES;
}

bool wait_all::check_if_over() {
  if (m_iES == exe_state::exe_over) return true;
  get_state();
  if (m_iES == exe_state::exe_over) return true;
  return false;
}

wait_any::wait_any(std::shared_ptr<::ff::internal::paras_with_lock> ps)
    : all_ps(ps), m_iES(exe_state::exe_init){};

exe_state wait_any::get_state() {
  if (!all_ps) m_iES = exe_state::exe_over;

  if (m_iES != exe_state::exe_over) {
    m_iES = exe_state::exe_wait;
    all_ps->lock.lock();
    if(all_ps->entities.size() == 0){
      m_iES = exe_state::exe_over;
    }else{
      for (auto p = all_ps->entities.begin(); p != all_ps->entities.end(); ++p)
        m_iES = exe_state_or(m_iES, p->get_state());
    }
    all_ps->lock.unlock();
  }
  return m_iES;
}

bool wait_any::check_if_over() {
  if (m_iES == exe_state::exe_over) return true;
  get_state();
  if (m_iES == exe_state::exe_over) return true;
  return false;
}
}  // end namespace internal

::ff::internal::wait_all all(paragroup &pg) {
  return ::ff::internal::wait_all(pg.all_entities());
}
::ff::internal::wait_any any(paragroup &pg) {
  return ::ff::internal::wait_any(pg.all_entities());
}
::ff::internal::wait_all all(paracontainer &pc) {
  return ::ff::internal::wait_all(pc.all_entities());
}
::ff::internal::wait_any any(paracontainer &pc) {
  return ::ff::internal::wait_any(pc.all_entities());
}
}  // end namespace ff
