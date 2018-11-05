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
namespace internal {

driver_base::driver_base() : m_exit_flag(false) {}

bool driver_base::init() {
  m_client = std::unique_ptr<ipc_client_endpoint>(new ipc_client_endpoint());
  add_handlers();

  //! we should make share wait_until_sync first

  bool ret = m_client->start();
  if (!ret)
    return ret;

  m_ipc_conn = m_client->ipc_connection();

  auto p = m_ipc_conn->construct<ipc_pkg::nbre_init_req>(
      nullptr, m_ipc_conn->default_allocator());
  m_ipc_conn->push_back(p);

  return true;
}

void driver_base::run() {
  neb::core::command_queue::instance().listen_command<neb::core::exit_command>(
      this, [this](const std::shared_ptr<neb::core::exit_command> &) {
        m_exit_flag = true;
      });
  neb::exception_queue &eq = neb::exception_queue::instance();
  while (!m_exit_flag) {
    std::shared_ptr<neb::neb_exception> ep = eq.pop_front();
    handle_exception(ep);
  }
}

void driver_base::handle_exception(
    const std::shared_ptr<neb::neb_exception> &p) {
  switch (p->type()) {
    LOG(ERROR) << p->what();
  case neb_exception::neb_std_exception:
    neb::core::command_queue::instance().send_command(
        std::make_shared<neb::core::exit_command>());
    break;
  case neb_exception::neb_shm_queue_failure:
    neb::core::command_queue::instance().send_command(
        std::make_shared<neb::core::exit_command>());
    break;
  case neb_exception::neb_shm_service_failure:
    neb::core::command_queue::instance().send_command(
        std::make_shared<neb::core::exit_command>());
    break;
  case neb_exception::neb_shm_session_already_start:
    neb::core::command_queue::instance().send_command(
        std::make_shared<neb::core::exit_command>());
    break;
  case neb_exception::neb_shm_session_timeout:
    neb::core::command_queue::instance().send_command(
        std::make_shared<neb::core::exit_command>());
    break;
  case neb_exception::neb_shm_session_failure:
    neb::core::command_queue::instance().send_command(
        std::make_shared<neb::core::exit_command>());
    break;
  case neb_exception::neb_configure_general_failure:
    neb::core::command_queue::instance().send_command(
        std::make_shared<neb::core::exit_command>());
    break;
  case neb_exception::neb_json_general_failure:
    neb::core::command_queue::instance().send_command(
        std::make_shared<neb::core::exit_command>());
    break;
  case neb_exception::neb_storage_exception_no_such_key:
    neb::core::command_queue::instance().send_command(
        std::make_shared<neb::core::exit_command>());
    break;
  case neb_exception::neb_storage_exception_no_init:
    neb::core::command_queue::instance().send_command(
        std::make_shared<neb::core::exit_command>());
    break;
  case neb_exception::neb_storage_general_failure:
    neb::core::command_queue::instance().send_command(
        std::make_shared<neb::core::exit_command>());
    break;
  default:
    break;
  }
}
} // end namespace internal

driver::driver() : internal::driver_base() {}
void driver::add_handlers() {
  m_client->add_handler<ipc_pkg::nbre_version_req>(
      [this](ipc_pkg::nbre_version_req *req) {
        ff::para<void> p;
        p([req, this]() {
          LOG(INFO) << " to start jit driver for data";
          using mi = pkg_type_to_module_info<ipc_pkg::nbre_version_req>;
          neb::block_height_t height = req->get<ipc_pkg::height>();
          auto irs = neb::core::ir_warden::instance().get_ir_by_name_height(
              mi::module_name, height);
          jit_driver d;
          d.run(this, irs, mi::func_name, req);
        });
      });

  m_client->add_handler<ipc_pkg::nbre_init_ack>(
      [this](ipc_pkg::nbre_init_ack *ack) {
        std::string s = ack->get<ipc_pkg::admin_pub_addr>().c_str();
        LOG(INFO) << "got admin_pub_addr: " << s;
        configuration::instance().auth_table_nas_addr() = s;
        ir_warden::instance().async_run();
        ir_warden::instance().wait_until_sync();
      });
}
}
} // namespace neb
