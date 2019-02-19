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
#include "ff/functionflow/runtime/runtime.h"
#include "ff/functionflow/runtime/rtcmn.h"

namespace ff {
extern bool g_initialized_flag;
void initialize(size_t concurrency) {
  rt::set_concurrency(concurrency);
  rt::runtime::instance();
}

namespace rt {
std::shared_ptr<runtime_deletor> runtime_deletor::s_pInstance(nullptr);
runtime_ptr runtime::s_pInstance(nullptr);
std::once_flag runtime::s_oOnce;

void schedule(task_base_ptr p) {
  static runtime_ptr r = runtime::instance();
  r->schedule(p);
}
void yield() { std::this_thread::yield(); }

runtime::runtime()
    : m_pTP(new threadpool()), m_oQueues(), m_bAllThreadsQuit(false){};

runtime::~runtime() {
  m_bAllThreadsQuit = true;
  {
    std::unique_lock<std::mutex> _l(m_wakeup_mutex);
    m_wakeup.notify_all();
  }
  m_pTP->join();
}

runtime_ptr runtime::instance() {
  if (!s_pInstance) std::call_once(s_oOnce, runtime::init);
  return s_pInstance;
}


void runtime::init() {
  s_pInstance = new runtime();
  runtime_deletor::s_pInstance = std::make_shared<runtime_deletor>(s_pInstance);
  auto thrd_num = concurrency();
  for (int i = 0; i < thrd_num; ++i) {
    s_pInstance->m_oQueues.push_back(
        std::unique_ptr<work_stealing_queue>(new work_stealing_queue()));
    s_pInstance->m_oWQueues.push_back(
        std::unique_ptr<simo_queue_t>(new simo_queue_t()));
  }
  s_pInstance->m_sleep_counter = 0;

  set_local_thrd_id(0);

  for (int i = 1; i < thrd_num; ++i) {
    s_pInstance->m_pTP->run([i]() {
      auto r = runtime::instance();
      set_local_thrd_id(i);
      r->thread_run();
    });
  }
  g_initialized_flag = true;
}

void runtime::init_for_no_ff_thread() {
  m_queue_mutex.lock();
  m_oWQueues.push_back(std::unique_ptr<simo_queue_t>(new simo_queue_t()));
  m_oQueues.push_back(
      std::unique_ptr<work_stealing_queue>(new work_stealing_queue()));
  m_queue_mutex.unlock();
  size_t t = std::atomic_fetch_add(&s_current_concurrency, (size_t)1);
  set_local_thrd_id(t);
}
void runtime::schedule(task_base_ptr p) {
  thread_local static int i = get_thrd_id();
  if (i == invalid_thrd_id) {
    init_for_no_ff_thread();
    i = get_thrd_id();
  }
  if (!m_oQueues[i]->push_back(p)) {
    run_task(p);
  } else if (m_sleep_counter > 1) {
    m_wakeup.notify_one();
  }
}

bool runtime::take_one_task(task_base_ptr &pTask) {
  bool b = false;
  thread_local static int i = get_thrd_id();
  thread_local static uint64_t ct = 0;
  b = m_oQueues[i]->pop(pTask);
  if (!b) {
    ct++;
    if ((ct & 0x1) == 0) {
      b = m_oWQueues[i]->pop(pTask);
      if (!b) {
        b = steal_one_task(pTask);
      }
    } else {
      b = steal_one_task(pTask);
      if (!b) {
        b = m_oWQueues[i]->pop(pTask);
      }
    }
  }
  return b;
}

void runtime::run_task(task_base_ptr &pTask) {
  thread_local static int cur_id = get_thrd_id();
  pTask->run();
  while (pTask->need_to_reschedule() && !m_oWQueues[cur_id]->push(pTask))
    pTask->run();
}
void runtime::thread_run() {
  bool flag = false;
  thread_local static int cur_id = get_thrd_id();
  task_base_ptr pTask;
  size_t dis = 1;
  while (!m_bAllThreadsQuit) {
    size_t ts = m_oQueues.size();
    int8_t retry_counter = 3;
    while (retry_counter > 0) {
      flag = take_one_task(pTask);
      if (flag) {
        run_task(pTask);
        retry_counter = 3;
      }
      if (!flag) {
        yield();
        retry_counter--;
      }
    }
    std::unique_lock<std::mutex> _l(m_wakeup_mutex);
    if (!m_bAllThreadsQuit) {
      m_sleep_counter++;
      m_wakeup.wait(_l);
      m_sleep_counter--;
    }
  }
}

bool runtime::steal_one_task(task_base_ptr &pTask) {
  thread_local static int cur_id = get_thrd_id();
  size_t dis = 1;
  size_t ts = m_oQueues.size();
  while ((cur_id + dis) % ts != cur_id) {
    if (m_oQueues[(cur_id + dis) % ts]->steal(pTask)) {
      return true;
    } else if (m_oWQueues[(cur_id + dis) % ts]->pop(pTask)) {
      return true;
    }
    dis++;
  }
  return false;
}
}  // end namespace rt
}  // end namespace ff
