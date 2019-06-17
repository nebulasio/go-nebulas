
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
#include "core/net_ipc/client/client_driver_base.h"
#include "common/address.h"
#include "common/configuration.h"
#include "compatible/compatible_checker.h"
#include "compatible/db_checker.h"
#include "core/ir_warden.h"
#include "core/net_ipc/client/client_context.h"
#include "fs/bc_storage_session.h"
#include "fs/blockchain.h"
#include "fs/ir_manager/api/ir_api.h"
#include "fs/ir_manager/api/ir_list.h"
#include "fs/ir_manager/ir_processor.h"
#include "fs/rocksdb_session_storage.h"
#include "fs/storage.h"
#include "jit/jit_driver.h"
#include "runtime/auth/auth_handler.h"
#include "runtime/auth/auth_table.h"
#include "runtime/util.h"
#include "runtime/version.h"
#include "util/persistent_flag.h"
#include "util/persistent_type.h"
#include <boost/filesystem.hpp>
#include <boost/property_tree/json_parser.hpp>
#include <boost/property_tree/ptree.hpp>
#include <ff/functionflow.h>

namespace neb {
namespace core {
namespace internal {

client_driver_base::client_driver_base() : m_exit_flag(false) {}
client_driver_base::~client_driver_base() {
  LOG(INFO) << "to destroy client driver base";
  if (m_timer_thread) {
    m_timer_thread->join();
  }
}

bool client_driver_base::init() {

  m_context = (client_context *)context.get();

  ff::initialize(4);
  m_client = std::unique_ptr<nipc_client>(new nipc_client());
  LOG(INFO) << "ipc client construct";
  add_handlers();

  //! we should make share wait_until_sync first

  bool ret = m_client->start();
  if (!ret)
    return ret;

  m_ipc_conn = m_client->connection();

  auto p = std::make_shared<nbre_init_req>();

  LOG(INFO) << "to send nbre_init_req";
  m_ipc_conn->send(p);

  return true;
}

void client_driver_base::run() {
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

void client_driver_base::handle_exception(
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

void client_driver_base::init_timer_thread() {
  if (m_timer_thread) {
    return;
  }
  m_timer_thread = std::unique_ptr<std::thread>(new std::thread([this]() {
    boost::asio::io_service io_service;

    m_timer_loop =
        std::unique_ptr<util::timer_loop>(new util::timer_loop(&io_service));
    m_timer_loop->register_timer_and_callback(
        neb::configuration::instance().ir_warden_time_interval(),
        []() { ir_warden::instance().on_timer(); });

    m_timer_loop->register_timer_and_callback(
        1, []() { jit_driver::instance().timer_callback(); });

    io_service.run();
  }));
}

void client_driver_base::init_nbre() {

  m_context->m_bc_storage = std::make_unique<fs::rocksdb_session_storage>();
  fs::rocksdb_session_storage *rss =
      (fs::rocksdb_session_storage *)m_context->blockchain_storage();
  rss->init(configuration::instance().neb_db_dir(),
            fs::storage_open_for_readonly);

  m_context->m_nbre_storage = std::make_unique<fs::rocksdb_storage>();
  fs::rocksdb_storage *rs = (fs::rocksdb_storage *)m_context->nbre_storage();
  rs->open_database(configuration::instance().nbre_db_dir(),
                    fs::storage_open_for_readwrite);

  m_context->m_blockchain =
      std::make_unique<fs::blockchain>(m_context->blockchain_storage());

  m_context->m_ir_processor = std::make_unique<fs::ir_processor>(
      m_context->nbre_storage(), m_context->blockchain());

  m_context->m_compatible_checker =
      std::make_unique<compatible::compatible_checker>();
  m_context->m_compatible_checker->init();

  m_context->set_ready();


  compatible::db_checker dc;
  dc.update_db_if_needed();

  neb::block_height_t height = 1;
  try {
    auto tmp = rs->get(neb::configuration::instance().nbre_max_height_name());
    height = neb::byte_to_number<neb::block_height_t>(tmp);
  } catch (const std::exception &e) {
  }
  LOG(INFO) << "init dip params with height " << height;
#if 0
  neb::rt::dip::dip_handler::instance().check_dip_params(height);
#endif
}
} // end namespace internal
} // namespace core
} // namespace neb
