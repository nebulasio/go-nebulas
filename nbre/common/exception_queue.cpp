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
#include "common/exception_queue.h"
#include "common/configuration.h"
#include "common/ir_conf_reader.h"
#include "fs/storage.h"

namespace neb {
void exception_queue::push_back(neb_exception::neb_exception_type type,
                                const char *what) {
  std::unique_lock<std::mutex> _l(m_mutex);
  bool was_empty = m_exceptions.empty();
  m_exceptions.push_back(std::make_shared<neb_exception>(type, what));
  _l.unlock();
  if (was_empty) {
    m_cond_var.notify_one();
  }
}
void exception_queue::push_back(const std::exception &p) {
  std::unique_lock<std::mutex> _l(m_mutex);
  bool was_empty = m_exceptions.empty();
  m_exceptions.push_back(std::make_shared<neb_exception>(
      neb_exception::neb_std_exception, p.what()));
  _l.unlock();
  if (was_empty) {
    m_cond_var.notify_one();
  }
}

neb_exception_ptr exception_queue::pop_front() {
  std::unique_lock<std::mutex> _l(m_mutex);

  if (m_exceptions.empty()) {
    m_cond_var.wait(_l);
  }
  if (m_exceptions.empty()) {
    return nullptr;
  }
  neb_exception_ptr ret = m_exceptions.front();
  m_exceptions.erase(m_exceptions.begin());
  return ret;
}

void exception_queue::catch_exception(const std::function<void()> &func) {
#define EC(a)                                                                  \
  LOG(ERROR) << "got exception: " << e.what();                                 \
  exception_queue::instance().push_back(neb_exception::a, e.what());

  try {
    func();
  } catch (const neb::configure_general_failure &e) {
    EC(neb_configure_general_failure);
  } catch (const neb::json_general_failure &e) {
    EC(neb_json_general_failure);
  } catch (const neb::fs::storage_exception_no_such_key &e) {
    EC(neb_storage_exception_no_such_key);
  } catch (const neb::fs::storage_exception_no_init &e) {
    EC(neb_storage_exception_no_init);
  } catch (const neb::fs::storage_general_failure &e) {
    EC(neb_storage_general_failure);
  } catch (const std::exception &e) {
    exception_queue::instance().push_back(e);
  }

#undef EC
}
} // namespace neb
