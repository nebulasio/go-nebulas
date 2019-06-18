// Copyright (C) 2017 go-nebulas authors
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
#include "core/execution_context.h"
#include "core/net_ipc/client/client_context.h"
namespace neb {
namespace core {
execution_context::execution_context()
    : m_cond_var(), m_mutex(), m_ready_flag(false) {}

execution_context::~execution_context() {}

bool execution_context::is_ready() const {
  std::unique_lock<std::mutex> _l(m_mutex);
  return m_ready_flag;
}

void execution_context::wait_until_ready() {
  std::unique_lock<std::mutex> _l(m_mutex);
  while (!m_ready_flag) {
    m_cond_var.wait(_l);
  }
}

void execution_context::set_ready() {
  std::unique_lock<std::mutex> _l(m_mutex);
  m_ready_flag = true;
  m_cond_var.notify_all();
}
std::unique_ptr<execution_context> context(new client_context());
} // namespace core
} // namespace neb
