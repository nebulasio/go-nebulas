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

  std::string serialize_to_string() {
    marshaler lr(marshaler::length_retriver);
    arch(lr);
    size_t s = lr.get_length();

    std::string ret(s, 0);
    marshaler sc((char *)ret.data(), s, marshaler::seralizer);
    arch(sc);
    return ret;
  }

  void deserialize_from_string(std::string &s) {
    marshaler sc(s.data(), s.size(), marshaler::deseralizer);
    arch(sc);
  }

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

template <uint32_t PackageID, typename... ARGS>
class udt_marshaler<ntpackage<PackageID, ARGS...>> {
public:
  static size_t seralize(char *buf, ntpackage<PackageID, ARGS...> &v) {
    marshaler lr(marshaler::length_retriver);
    v.arch(lr);
    size_t s = lr.get_length();

    marshaler sc(buf, s, marshaler::seralizer);
    v.arch(sc);
    return s;
  }
  static size_t deseralize(const char *buf, size_t len,
                           ntpackage<PackageID, ARGS...> &v) {
    marshaler sc(buf, len, marshaler::deseralizer);
    v.arch(sc);
    marshaler lr(marshaler::length_retriver);
    v.arch(lr);
    size_t s = lr.get_length();
    return s;
  }
  static size_t length(ntpackage<PackageID, ARGS...> &v) {
    marshaler lr(marshaler::length_retriver);
    v.arch(lr);
    size_t s = lr.get_length();
    return s;
  }
};
} // namespace net
} // namespace ff

