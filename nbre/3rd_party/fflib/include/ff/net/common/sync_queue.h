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
#pragma once
#include <boost/noncopyable.hpp>
#include <boost/thread/locks.hpp>
#include <boost/thread/mutex.hpp>
#include <list>

namespace ff {
namespace net {
#define GUARD_LOCK boost::lock_guard<boost::mutex> _l(m_oMutex)

template <class Ty_> class sync_queue : public boost::noncopyable {
public:
  typedef std::list<Ty_> Container_t;
  typedef Ty_ Elem_t;

public:
  sync_queue() : m_oQueue(), m_oMutex() {}

  std::size_t size() {
    GUARD_LOCK;
    return m_oQueue.size();
  }

  void push(const Ty_ &val) {
    GUARD_LOCK;
    m_oQueue.push_back(val);
  }

  bool pop(Ty_ &val) {
    GUARD_LOCK;
    if (m_oQueue.empty())
      return false;

    val = m_oQueue.front();
    m_oQueue.pop_front();
    return true;
  }
  boost::mutex &mutex() { return m_oMutex; }
  std::list<Ty_> &content() { return m_oQueue; }

protected:
  std::list<Ty_> m_oQueue;
  mutable boost::mutex m_oMutex;
};
#undef GUARD_LOCK
}
} // namespace ff

