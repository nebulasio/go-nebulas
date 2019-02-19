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
#include "ff/net/middleware/package.h"
#include "ff/util/ntobject.h"
#include "ff/util/type_list.h"
#include <memory>

namespace ff {
namespace net {

template <uint32_t PackgeID, typename... ARGS>
class ntpackage : public package, public ::ff::util::ntobject<ARGS...> {
public:
  typedef typename util::type_list<ARGS...> type_list;

  ntpackage() : package(PackgeID), ::ff::util::ntobject<ARGS...>() {}

  virtual void archive(marshaler &ar) { archive_helper<0>::run(ar, *this); }

protected:
  template <int Index> struct archive_helper {
    template <typename VT>
    static auto run(marshaler &ar, VT &val) ->
        typename std::enable_if<(VT::type_list::len > Index), void>::type {
      ar.archive(std::get<Index>(*val.m_content));
      archive_helper<Index + 1>::run(ar, val);
    }

    template <typename VT>
    static auto run(marshaler &, VT &) ->
        typename std::enable_if<(VT::type_list::len <= Index), void>::type {}
  };
};
} // namespace net
} // namespace ff

