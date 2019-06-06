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
#include "common/configuration.h"
#include "core/ir_warden.h"
#include "core/neb_ipc/server/ipc_configuration.h"
#include "fs/fs_storage.h"
#include "fs/ir_manager/api/ir_api.h"
#include "jit/jit_driver.h"
#include "runtime/dip/dip_handler.h"
#include "runtime/nr/impl/nr_handler.h"
#include "runtime/version.h"
#include <ff/functionflow.h>

namespace neb {
namespace core {
namespace internal {

driver_base::driver_base() : m_exit_flag(false) {}

bool driver_base::init() {
  ff::initialize(8);
  m_client = std::unique_ptr<ipc_client_endpoint>(new ipc_client_endpoint());
  LOG(INFO) << "ipc client construct";
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

void driver_base::init_timer_thread() {
  if (m_timer_thread) {
    return;
  }
  m_timer_thread = std::unique_ptr<std::thread>(new std::thread([this]() {
    boost::asio::io_service io_service;

    m_timer_loop = std::unique_ptr<timer_loop>(new timer_loop(&io_service));
    m_timer_loop->register_timer_and_callback(
        neb::configuration::instance().ir_warden_time_interval(),
        []() { ir_warden::instance().on_timer(); });

    m_timer_loop->register_timer_and_callback(
        1, []() { jit_driver::instance().timer_callback(); });

    io_service.run();
  }));
}

void driver_base::init_nbre() {

  LOG(INFO) << "call func init_nbre";
  auto rs = neb::fs::fs_storage::instance().nbre_db_ptr();
  neb::rt::dip::dip_handler::instance().read_dip_reward_from_storage();

  try {
    auto nbre_max_height_bytes =
        rs->get(neb::configuration::instance().nbre_max_height_name());
    auto nbre_max_height =
        neb::util::byte_to_number<block_height_t>(nbre_max_height_bytes);
    LOG(INFO) << "nbre max height " << nbre_max_height;
    neb::rt::dip::dip_handler::instance().init_dip_params(nbre_max_height);
    LOG(INFO) << "init dip params done when init nbre";
  } catch (const std::exception &e) {
    LOG(INFO) << "nbre max height not init " << e.what();
  }
  LOG(INFO) << "driver init nbre done";
}

} // end namespace internal

driver::driver() : internal::driver_base() {}

void driver::add_handlers() {
  m_client->add_handler<ipc_pkg::nbre_version_req>(
      [this](ipc_pkg::nbre_version_req *req) {
        neb::core::ipc_pkg::nbre_version_ack *ack =
            m_ipc_conn->construct<neb::core::ipc_pkg::nbre_version_ack>(
                req->m_holder, m_ipc_conn->default_allocator());
        if (ack == nullptr) {
          return;
        }
        neb::util::version v = neb::rt::get_version();
        ack->set<neb::core::ipc_pkg::major>(v.major_version());
        ack->set<neb::core::ipc_pkg::minor>(v.minor_version());
        ack->set<neb::core::ipc_pkg::patch>(v.patch_version());
        m_ipc_conn->push_back(ack);
      });

  m_client->add_handler<ipc_pkg::nbre_init_ack>(
      [this](ipc_pkg::nbre_init_ack *ack) {
        LOG(INFO) << "get init ack";

        ipc_configuration::instance().nbre_root_dir() =
            ack->get<ipc_pkg::nbre_root_dir>().c_str();
        ipc_configuration::instance().nbre_exe_name() =
            ack->get<ipc_pkg::nbre_exe_name>().c_str();
        ipc_configuration::instance().neb_db_dir() =
            ack->get<ipc_pkg::neb_db_dir>().c_str();
        ipc_configuration::instance().nbre_db_dir() =
            ack->get<ipc_pkg::nbre_db_dir>().c_str();
        ipc_configuration::instance().nbre_log_dir() =
            ack->get<ipc_pkg::nbre_log_dir>().c_str();
        ipc_configuration::instance().nbre_start_height() =
            ack->get<ipc_pkg::nbre_start_height>();

        std::string addr_base58 = ack->get<ipc_pkg::admin_pub_addr>().c_str();
        neb::util::bytes addr_bytes =
            neb::util::bytes::from_base58(addr_base58);
        ipc_configuration::instance().admin_pub_addr() =
            neb::util::byte_to_string(addr_bytes);

        LOG(INFO) << ipc_configuration::instance().nbre_db_dir();
        LOG(INFO) << ipc_configuration::instance().admin_pub_addr();
        FLAGS_log_dir = ipc_configuration::instance().nbre_log_dir();
        google::InitGoogleLogging("nbre-client");

        init_nbre();
        init_timer_thread();
        ir_warden::instance().wait_until_sync();
      });

  m_client->add_handler<ipc_pkg::nbre_ir_list_req>(
      [this](ipc_pkg::nbre_ir_list_req *req) {
        neb::core::ipc_pkg::nbre_ir_list_ack *ack =
            m_ipc_conn->construct<neb::core::ipc_pkg::nbre_ir_list_ack>(
                req->m_holder, m_ipc_conn->default_allocator());
        if (ack == nullptr) {
          return;
        }

        auto rs = neb::fs::fs_storage::instance().nbre_db_ptr();
        auto irs_ptr = neb::fs::ir_api::get_ir_list(rs);
        for (auto &ir : *irs_ptr) {
          neb::ipc::char_string_t ir_name(ir.c_str(),
                                          m_ipc_conn->default_allocator());
          ack->get<neb::core::ipc_pkg::ir_name_list>().push_back(ir_name);
        }
        m_ipc_conn->push_back(ack);
      });

  m_client->add_handler<ipc_pkg::nbre_ir_versions_req>(
      [this](ipc_pkg::nbre_ir_versions_req *req) {
        neb::core::ipc_pkg::nbre_ir_versions_ack *ack =
            m_ipc_conn->construct<neb::core::ipc_pkg::nbre_ir_versions_ack>(
                req->m_holder, m_ipc_conn->default_allocator());
        if (ack == nullptr) {
          return;
        }

        auto ir_name = req->get<ipc_pkg::ir_name>();
        ack->set<neb::core::ipc_pkg::ir_name>(ir_name);

        auto rs = neb::fs::fs_storage::instance().nbre_db_ptr();
        auto ir_versions_ptr =
            neb::fs::ir_api::get_ir_versions(ir_name.c_str(), rs);

        for (auto &v : *ir_versions_ptr) {
          ack->get<neb::core::ipc_pkg::ir_versions>().push_back(v);
        }
        m_ipc_conn->push_back(ack);
      });

  m_client->add_handler<ipc_pkg::nbre_nr_handler_req>(
      [this](ipc_pkg::nbre_nr_handler_req *req) {
        neb::core::ipc_pkg::nbre_nr_handler_ack *ack =
            m_ipc_conn->construct<neb::core::ipc_pkg::nbre_nr_handler_ack>(
                req->m_holder, m_ipc_conn->default_allocator());
        if (ack == nullptr) {
          return;
        }

        if (!neb::rt::nr::nr_handler::instance().get_nr_handler_id().empty()) {
          ack->set<neb::core::ipc_pkg::nr_handler_id>(
              std::string("nr handler not available").c_str());
          m_ipc_conn->push_back(ack);
          return;
        }

        uint64_t start_block = req->get<ipc_pkg::start_block>();
        uint64_t end_block = req->get<ipc_pkg::end_block>();
        uint64_t nr_version = req->get<ipc_pkg::nr_version>();

        std::stringstream ss;
        ss << neb::util::number_to_byte<neb::util::bytes>(start_block).to_hex()
           << neb::util::number_to_byte<neb::util::bytes>(end_block).to_hex()
           << neb::util::number_to_byte<neb::util::bytes>(nr_version).to_hex();

        neb::ipc::char_string_t cstr_handler_id(
            ss.str().c_str(), m_ipc_conn->default_allocator());
        ack->set<neb::core::ipc_pkg::nr_handler_id>(cstr_handler_id);
        neb::rt::nr::nr_handler::instance().start(start_block, end_block,
                                                  nr_version);
        m_ipc_conn->push_back(ack);

      });

  m_client->add_handler<ipc_pkg::nbre_nr_result_req>(
      [this](ipc_pkg::nbre_nr_result_req *req) {
        neb::core::ipc_pkg::nbre_nr_result_ack *ack =
            m_ipc_conn->construct<neb::core::ipc_pkg::nbre_nr_result_ack>(
                req->m_holder, m_ipc_conn->default_allocator());
        if (ack == nullptr) {
          return;
        }

        std::string nr_handler_id = req->get<ipc_pkg::nr_handler_id>().c_str();
        neb::ipc::char_string_t cstr_nr_result(
            neb::rt::nr::nr_handler::instance()
                .get_nr_result(nr_handler_id)
                .c_str(),
            m_ipc_conn->default_allocator());
        ack->set<neb::core::ipc_pkg::nr_result>(cstr_nr_result);
        m_ipc_conn->push_back(ack);
      });

  m_client->add_handler<ipc_pkg::nbre_dip_reward_req>(
      [this](ipc_pkg::nbre_dip_reward_req *req) {
        neb::core::ipc_pkg::nbre_dip_reward_ack *ack =
            m_ipc_conn->construct<neb::core::ipc_pkg::nbre_dip_reward_ack>(
                req->m_holder, m_ipc_conn->default_allocator());
        if (ack == nullptr) {
          return;
        }

        auto height = req->get<ipc_pkg::height>();
        LOG(INFO) << "ipc client receive dip reward req, height " << height;
        neb::ipc::char_string_t cstr_dip_reward(
            neb::rt::dip::dip_handler::instance()
                .get_dip_reward(height)
                .c_str(),
            m_ipc_conn->default_allocator());
        ack->set<neb::core::ipc_pkg::dip_reward>(cstr_dip_reward);
        m_ipc_conn->push_back(ack);
      });
}
}
} // namespace neb
