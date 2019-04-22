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
#ifndef FF_RUNTIME_SIMO_QUEUE_H_
#define FF_RUNTIME_SIMO_QUEUE_H_

#include "ff/functionflow/common/common.h"
namespace ff {
namespace rt {
// N, 2^N
//! This queue is for single-thread's push, and multiple-threads' pop.
//! This queue is capability-fixed
template <class T, size_t N>
class simo_queue {
  const static int64_t MASK = (1 << N) - 1;

 public:
  simo_queue() : array(nullptr), cap(0), head(0), tail(0) {
    array = new T[1 << N];
    cap = 1 << N;
  }

  bool push(const T& val) {
    if (head - tail >= MASK) return false;
    array[head & MASK] = val;
    head++;
    return true;
  }
  bool pop(T& val) {
    auto t = tail;
    if (t == head) return false;
    val = array[t & MASK];
    while (!__sync_bool_compare_and_swap(&tail, t, t + 1)) {
      t = tail;
      if (t == head) return false;
      val = array[t & MASK];
    }
    return true;
  }
  size_t size() const { return head - tail; }

 protected:
  T* array;
  int64_t cap;
  int64_t head;
  int64_t tail;
};  // end class simo_queue;
}
}  // end namespace ff
#endif
