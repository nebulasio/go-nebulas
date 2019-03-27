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
#ifndef FF_RUNTIME_HAZARD_POINTER_H_
#define FF_RUNTIME_HAZARD_POINTER_H_
#include "ff/functionflow/runtime/rtcmn.h"
#include <atomic>
#include <mutex>
#ifdef FUNCTION_FLOW_DEBUG
#include <sstream>
#endif

//! This hazard pointer is specifical for FF!!
namespace ff {
namespace rt {
template <class T>
class hp_owner {
 public:
  hp_owner() : m_oflag() {
#ifdef CLANG_LLVM
    m_pPointers = new std::atomic<T *>[ff::rt::concurrency()];
#else
    std::call_once(m_oflag, [this]() {
      m_pPointers = new std::atomic<T *>[ff::rt::concurrency()];
    });
#endif
  }

  ~hp_owner() { delete[] m_pPointers; }

  std::atomic<T *> &get_hazard_pointer() {
    return m_pPointers[ff::rt::get_thrd_id()];
  }

  //! return true if other thread had it.
  bool outstanding_hazard_pointer_for(T *p) {
    thread_local static thrd_id_t id = ff::rt::get_thrd_id();
    if (!p) return false;
    for (int i = 0; i < ff::rt::concurrency(); i++) {
      if (i == id) continue;
      if (m_pPointers[i].load(std::memory_order_acquire) == p) return true;
    }
    return false;
  }
#ifdef FUNCTION_FLOW_DEBUG
  std::string str() {
    std::stringstream ss;
    for (int i = 0; i < ff::rt::concurrency(); ++i) {
      ss << i << ":" << m_pPointers[i].load() << ";";
    }
    return ss.str();
  }
#endif
 protected:
  std::once_flag m_oflag;
  std::atomic<T *> *m_pPointers;
};  // end class hp_owner

}  // end namespace rt
}  // end namespace ff

#endif
