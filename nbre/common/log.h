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

#include <glog/logging.h>
#include <iostream>
#include <sstream>

namespace neb {

enum log_level_t {
  log_DEBUG = 0,
  log_INFO,
  log_WARNING,
  log_ERROR,
};

class log_flush_base {
public:
  log_flush_base() = default;
  virtual ~log_flush_base() = default;

  template <typename T> log_flush_base &operator<<(T const &msg) {
    m_oss << msg;
    return *this;
  }

protected:
  std::ostringstream m_oss;
};

template <log_level_t TL> class log_flush : public log_flush_base {};
template <> class log_flush<log_INFO> : public log_flush_base {
public:
  virtual ~log_flush() {
    LOG(INFO) << m_oss.str();
    google::FlushLogFiles(google::INFO);
  }
};
template <> class log_flush<log_WARNING> : public log_flush_base {
public:
  virtual ~log_flush() {
    LOG(WARNING) << m_oss.str();
    google::FlushLogFiles(google::WARNING);
  }
};
template <> class log_flush<log_ERROR> : public log_flush_base {
public:
  virtual ~log_flush() {
    LOG(ERROR) << m_oss.str();
    google::FlushLogFiles(google::ERROR);
  }
};

#define LOG_FLUSH(LEVEL) log_flush<log_##LEVEL>()

} // namespace neb
