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
#ifndef FF_RUNTIME_RUNTIME_H_
#define FF_RUNTIME_RUNTIME_H_
#include "ff/functionflow/common/common.h"
#include "ff/functionflow/runtime/rtcmn.h"
#include "ff/functionflow/runtime/task_queue.h"
#include "ff/functionflow/runtime/threadpool.h"
#include "ff/functionflow/utilities/simo_queue.h"

namespace ff {

void initialize(size_t concurrency = 0);

namespace rt {

class threadpool;
class runtime;
typedef runtime *runtime_ptr;

class runtime {
 protected:
  runtime();
  runtime(const runtime &) = delete;

 public:
  virtual ~runtime();
  static runtime_ptr instance();

  void schedule(task_base_ptr p);
  bool take_one_task(task_base_ptr &p);

  bool steal_one_task(task_base_ptr &p);
  void run_task(task_base_ptr &p);

  bool is_idle();

  thrd_id_t get_idle();
  std::tuple<uint64_t, uint64_t> current_task_counter();

 protected:
  void thread_run();
  void init_for_no_ff_thread();
  static void init();

 protected:
  std::unique_ptr<threadpool> m_pTP;
  std::vector<std::unique_ptr<work_stealing_queue> > m_oQueues;
  typedef simo_queue<task_base_ptr, 8> simo_queue_t;
  std::vector<std::unique_ptr<simo_queue_t> > m_oWQueues;
  //    thread_local static work_stealing_queue *
  //    m_pLQueue;
  std::atomic<bool> m_bAllThreadsQuit;
  static runtime_ptr s_pInstance;
  static std::once_flag s_oOnce;
  std::atomic_int m_sleep_counter;
  std::mutex m_wakeup_mutex;
  std::condition_variable m_wakeup;
  std::mutex m_queue_mutex;
};  // end class runtime

class runtime_deletor {
 public:
  runtime_deletor(runtime *pRT) : m_pRT(pRT){};
  ~runtime_deletor() { delete m_pRT; };
  static std::shared_ptr<runtime_deletor> s_pInstance;

 protected:
  runtime *m_pRT;
};

//! Get the number of exe_over_tasks and scheduled_tasks
inline std::tuple<uint64_t, uint64_t> current_task_counter() {
  static runtime_ptr r = runtime::instance();
  return r->current_task_counter();
}

void schedule(task_base_ptr p);

template <class Func>
void yield_and_ret_until(Func &&f) {
  int cur_id = get_thrd_id();
  runtime_ptr r = runtime::instance();
  bool b = f();
  task_base_ptr pTask;

  while (!b) {
    if (r->take_one_task(pTask)) {
      r->run_task(pTask);
    } else {
      yield();
    }
    b = f();
  }
}
}  // end namespace rt

}  // end namespace ff
#endif
