// Copyright (C) 2018 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or
// modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see
// <http://www.gnu.org/licenses/>.
//
#pragma once
#include "common/common.h"
#include "common/util/singleton.h"
#include <algorithm>
#include <condition_variable>
#include <exception>
#include <thread>

namespace neb {

class neb_exception {
public:
  inline neb_exception(const std::string &msg) : m_msg(msg) {}
  inline const char *what() const throw() { return m_msg.c_str(); }

protected:
  std::string m_msg;
};

typedef std::shared_ptr<neb_exception> neb_exception_ptr;

class exception_queue : public neb::util::singleton<exception_queue> {
public:
  void push_back(const std::exception &p);

  inline bool empty() const {
    std::lock_guard<std::mutex> _l(m_mutex);
    return m_exceptions.empty();
  }
  inline size_t size() const {
    std::lock_guard<std::mutex> _l(m_mutex);
    return m_exceptions.size();
  }

  neb_exception_ptr pop_front();

  inline void for_each(const std::function<void(neb_exception_ptr p)> &func) {
    std::lock_guard<std::mutex> _l(m_mutex);
    std::for_each(m_exceptions.begin(), m_exceptions.end(), func);
  }

protected:
  std::vector<neb_exception_ptr> m_exceptions;
  mutable std::mutex m_mutex;
  std::condition_variable m_cond_var;
};
}

