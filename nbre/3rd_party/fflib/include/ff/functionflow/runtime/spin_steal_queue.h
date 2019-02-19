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
#ifndef FF_RUNTIME_SPIN_STEALING_QUEUE_H_
#define FF_RUNTIME_SPIN_STEALING_QUEUE_H_
#include "ff/functionflow/common/common.h"
#include "ff/functionflow/utilities/scope_guard.h"
#include "ff/functionflow/utilities/spin_lock.h"
namespace ff {
namespace rt {
template <class T, size_t N>
class spin_stealing_queue {
  const static uint64_t INITIAL_SIZE = 1 << N;
  const static int64_t mask = (1 << N) - 1;

 public:
  spin_stealing_queue()
      : array(new T[1 << N]), head(0), tail(0), steal_lock() {}
  ~spin_stealing_queue() {
    if (array != nullptr) {
      delete[] array;
    }
  }

  bool push_back(const T& val) {
    std::lock_guard<ff::spinlock> __l(steal_lock);
    if (head - tail == mask) return false;
    array[head & mask] = val;
    head++;
    return true;
  }

  bool pop(T& val) {
    std::lock_guard<ff::spinlock> __l(steal_lock);
    if (head == tail) {
      return false;
    }

    head--;
    val = array[head & mask];
    return true;
  }

  bool steal(T& val) {
    std::lock_guard<ff::spinlock> __l(steal_lock);
    if (tail == head) return false;
    val = array[tail & mask];
    tail++;
    return true;
  }

  uint64_t size() {
    scope_guard __l([this]() { steal_lock.lock(); },
                    [this]() { steal_lock.unlock(); });
    return head - tail;
  }

 protected:
  T* array;
  int64_t head;
  int64_t tail;
  ff::spinlock steal_lock;
};  // end class nonblocking_stealing_queue

}  // end namespace rt;
}  // end namespace ff

#endif
