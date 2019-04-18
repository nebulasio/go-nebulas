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
#include "ff/net/common/common.h"

namespace ff {
namespace net {
class event_handler {
public:
  template <class ETy_> void listen(const typename ETy_::Handler_t &h) {
    size_t v = ETy_::identifier;
    if (m_oHandlers.find(v) == m_oHandlers.end()) {
      m_oHandlers[v] = std::vector<boost::any>();
    }
    m_oHandlers[v].push_back(boost::any(h));
  }

  template <class ETy_, class... TS> void triger(TS... t) {
    size_t v = ETy_::identifier;
    if (m_oHandlers.find(v) == m_oHandlers.end())
      return;
    for (auto it : m_oHandlers[v]) {
      typename ETy_::Handler_t h =
          boost::any_cast<typename ETy_::Handler_t>(it);
      h(t...);
    }
}


protected:
  typedef std::map<size_t, std::vector<boost::any>> ETHandlers_t;
  ETHandlers_t m_oHandlers;
};

} // namespace net
} // namespace ff

