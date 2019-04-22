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

#ifndef FF_RUNTIME_GTWSQ_FIXED_H_
#define FF_RUNTIME_GTWSQ_FIXED_H_
#include "ff/functionflow/common/common.h"
#include "ff/functionflow/runtime/rtcmn.h"

namespace ff {
namespace rt {

// N, 2^N.
template <class T, size_t N>
class default_work_stealing_queue {
  const static int64_t INITIAL_SIZE = 1 << N;
  const static int64_t MASK = (1 << N) - 1;

 public:
  default_work_stealing_queue() : head(0), tail(0), array(new T[1 << N]) {
    if (array == nullptr) {
      assert(false && "Allocation Failed!");
      exit(-1);
    }
  }

  ~default_work_stealing_queue() {
    if (array != nullptr) {
      delete[] array;
      array = nullptr;
    }
  }

  bool push_back(const T& val) {
    auto h = head;
    if (h - tail.load(std::memory_order_relaxed) == MASK) return false;
    array[h & MASK] = val;
    head = h + 1;
    return true;
  }
  bool pop(T& val) {
    auto h = head;
    if (h <= tail.load(std::memory_order_relaxed)) {
      head = tail.load(std::memory_order_relaxed);
      return false;
    }

    head = h - 1;
    h = head;
    std::atomic_thread_fence(std::memory_order_seq_cst);
    //;__sync_synchronize(); //This two lines are the magic
    auto t = tail.load(std::memory_order_relaxed);

    if (h < t) {
      head = tail.load(std::memory_order_relaxed);
      return false;
    }
    val = array[h & MASK];
    if (h > t) return true;
    bool res = true;
    if (!std::atomic_compare_exchange_strong_explicit(
            &tail, &h, h + 1, std::memory_order_relaxed,
            std::memory_order_relaxed))
      res = false;
    head = tail.load(std::memory_order_relaxed);
    return res;
  }

  bool steal(T& val) {
    int64_t t = tail.load(std::memory_order_relaxed);
    int s = head - t;
    if (s <= 0) {
      return false;
    }
    val = array[t & MASK];
    if (!std::atomic_compare_exchange_strong_explicit(
            &tail, &t, t + 1, std::memory_order_relaxed,
            std::memory_order_relaxed))
      return false;
    return true;
  }

  inline int64_t size() {
    return (head - tail.load(std::memory_order_relaxed));
  }
  inline int64_t get_head() const { return head; }
  inline int64_t get_tail() const {
    return tail.load(std::memory_order_relaxed);
  }

 protected:
  int64_t head;
  std::atomic<int64_t> tail;
  T* array;
};  // end class default_work_stealing_queue

}  // end namespace rt
}  // end namespace ff
#endif
