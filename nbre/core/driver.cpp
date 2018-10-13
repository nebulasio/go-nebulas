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
#include "core/driver.h"
#include "core/ir_warden.h"
#include "jit/jit_driver.h"
#include <ff/ff.h>

namespace neb {
namespace core {
driver::driver() : m_exit_flag(false) {}

bool driver::init() {
  m_client = std::unique_ptr<ipc_client>(new ipc_client());
  add_handlers();

  //! we should make share wait_until_sync first
  ir_warden::instance().async_run();
  ir_warden::instance().wait_until_sync();

  m_client->start();
  m_ipc_conn = m_client->ipc_connection();

  return true;
}

void driver::run() {
  neb::core::command_queue::instance().listen_command<neb::core::exit_command>(
      this, [this](const std::shared_ptr<neb::core::exit_command> &) {
        m_exit_flag = true;
      });
  neb::exception_queue &eq = neb::exception_queue::instance();
  while (!m_exit_flag) {
    std::shared_ptr<neb::neb_exception> ep = eq.pop_front();
    // TODO handle exceptions here
  }
}
void driver::add_handlers() {
  m_client->add_handler<nbre_version_req>([this](nbre_version_req *req) {
    ff::para<void> p;
    p([req, this]() {
      module_info mi = pkg_type_to_module_info<nbre_version_req>();
      neb::block_height_t height = req->m_height;
      auto irs = neb::core::ir_warden::instance().get_ir_by_name_height(
          mi.module_name, height);
      jit_driver d;
      d.run(this, irs, mi.func_name);

    });

  });
}
void driver::handle_exception(const std::shared_ptr<neb::neb_exception> &p) {
  switch (p->type()) {
  case neb_exception::neb_std_exception:
    break;
  default:
    break;
  }
}
}
}
