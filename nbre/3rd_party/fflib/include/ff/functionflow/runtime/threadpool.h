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
#ifndef FF_RUNTIME_THREAD_POOL_H_
#define FF_RUNTIME_THREAD_POOL_H_
#include <thread>
#include <vector>
#include <memory>
#include <functional>
#include <iostream>

namespace ff {
namespace rt {
class threadpool {
 public:
  threadpool() : m_oThreads(){};

  template <class F>
  void run(int thrd_num, F&& func) {
    for (int i = 0; i < thrd_num; i++) {
      m_oThreads.push_back(std::thread(func));
    }
  }

  template <class F>
  void run(F&& func) {
    m_oThreads.push_back(std::thread(func));
  }

  void join() {
    for (size_t i = 0; i < m_oThreads.size(); ++i) {
      if (m_oThreads[i].joinable()) m_oThreads[i].join();
    }
  }

 protected:
  std::vector<std::thread> m_oThreads;
};  // end class threadpool;
}  // end namespace rt
}  // end namespace ff

#endif
