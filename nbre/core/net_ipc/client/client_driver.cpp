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
#include "core/net_ipc/client/client_driver.h"
#include "common/address.h"
#include "common/configuration.h"
#include "compatible/compatible_checker.h"
#include "compatible/db_checker.h"
#include "core/ir_warden.h"
#include "core/net_ipc/client/client_context.h"
#include "fs/bc_storage_session.h"
#include "fs/ir_manager/api/ir_api.h"
#include "fs/rocksdb_session_storage.h"
#include "fs/storage_holder.h"
#include "jit/jit_driver.h"
#include "runtime/dip/dip_handler.h"
#include "runtime/nr/impl/nr_handler.h"
#include "runtime/util.h"
#include "runtime/version.h"
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
  m_context->m_bc_storage->init(configuration::instance().neb_db_dir(),
                                fs::storage_open_for_readonly);

  m_context->m_nbre_storage = std::make_unique<fs::rocksdb_storage>();
  m_context->m_nbre_storage->open_database(
      configuration::instance().nbre_db_dir(), fs::storage_open_for_readwrite);

  m_context->m_blockchain =
      std::make_unique<fs::blockchain>(m_context->blockchain_storage());

  m_context->m_ir_processor = std::make_unique<fs::ir_processor>(
      m_context->nbre_storage(), m_context->blockchain());

  m_context->m_compatible_checker =
      std::make_unique<compatible::compatible_checker>();
  m_context->m_compatible_checker->init();

  m_context->set_ready();

  auto *rs = m_context->m_nbre_storage.get();

  compatible::db_checker dc;
  dc.update_db_if_needed();

  neb::block_height_t height = 1;
  try {
    auto tmp = rs->get(neb::configuration::instance().nbre_max_height_name());
    height = neb::byte_to_number<neb::block_height_t>(tmp);
  } catch (const std::exception &e) {
  }
  LOG(INFO) << "init dip params with height " << height;
  neb::rt::dip::dip_handler::instance().check_dip_params(height);
}
} // end namespace internal

client_driver::client_driver() : internal::client_driver_base() {}
client_driver::~client_driver() { LOG(INFO) << "to destroy client driver"; }

void client_driver::add_handlers() {

  LOG(INFO) << "client is " << m_client.get();
  m_client->add_handler<nbre_version_req>(
      [this](std::shared_ptr<nbre_version_req> req) {
        LOG(INFO) << "recv nbre_version_req";
        auto ack = new_ack_pkg<nbre_version_ack>(req);
        if (ack == nullptr) {
          return;
        }

        neb::version v = neb::rt::get_version();
        ack->set<p_major>(v.major_version());
        ack->set<p_minor>(v.minor_version());
        ack->set<p_patch>(v.patch_version());
        m_ipc_conn->send(ack);
      });

  m_client->add_handler<nbre_init_ack>([this](
                                           std::shared_ptr<nbre_init_ack> ack) {
    LOG(INFO) << "recv nbre_init_ack";
    try {
      configuration::instance().nbre_root_dir() =
          ack->get<p_nbre_root_dir>().c_str();
      configuration::instance().nbre_exe_name() =
          ack->get<p_nbre_exe_name>().c_str();
      configuration::instance().neb_db_dir() = ack->get<p_neb_db_dir>().c_str();
      configuration::instance().nbre_db_dir() =
          ack->get<p_nbre_db_dir>().c_str();
      configuration::instance().nbre_log_dir() =
          ack->get<p_nbre_log_dir>().c_str();
      configuration::instance().nbre_start_height() =
          ack->get<p_nbre_start_height>();

      std::string addr = ack->get<p_admin_pub_addr>().c_str();
      // neb::util::bytes addr_bytes =
      // neb::util::bytes::from_base58(addr_base58);
      configuration::instance().admin_pub_addr() = to_address(addr);

      LOG(INFO) << configuration::instance().nbre_db_dir();
      LOG(INFO) << configuration::instance().neb_db_dir();
      // LOG(INFO) << addr_base58;

      init_nbre();
      init_timer_thread();
      ir_warden::instance().wait_until_sync();
    } catch (const std::exception &e) {
      LOG(ERROR) << "got exception " << typeid(e).name()
                 << " with what: " << e.what();
    }
  });

  m_client->add_handler<nbre_ir_list_req>(
      [this](std::shared_ptr<nbre_ir_list_req> req) {
        LOG(INFO) << "recv nbre_ir_list_req";
        try {
          auto ack = new_ack_pkg<nbre_ir_list_ack>(req);

          auto rs = neb::fs::storage_holder::instance().nbre_db_ptr();
          auto irs_ptr = neb::fs::ir_api::get_ir_list(rs);

          boost::property_tree::ptree pt, root;
          for (auto &ir : *irs_ptr) {
            boost::property_tree::ptree child;
            child.put("", ir);
            pt.push_back(std::make_pair("", child));
          }
          root.add_child(neb::configuration::instance().ir_list_name(), pt);
          std::stringstream ss;
          boost::property_tree::json_parser::write_json(ss, root);

          ack->set<p_ir_name_list>(ss.str());
          m_ipc_conn->send(ack);

        } catch (const std::exception &e) {
          LOG(ERROR) << "got exception " << typeid(e).name()
                     << " with what: " << e.what();
        }
      });

  m_client->add_handler<nbre_ir_versions_req>(
      [this](std::shared_ptr<nbre_ir_versions_req> req) {
        LOG(INFO) << "recv nbre_ir_versions_req";
        try {
          auto ack = new_ack_pkg<nbre_ir_versions_ack>(req);
          auto ir_name = req->get<p_ir_name>();

          auto rs = neb::fs::storage_holder::instance().nbre_db_ptr();
          auto ir_versions_ptr =
              neb::fs::ir_api::get_ir_versions(ir_name.c_str(), rs);

          boost::property_tree::ptree pt, root;
          for (auto &v : *ir_versions_ptr) {
            boost::property_tree::ptree child;
            child.put("", v);
            pt.push_back(std::make_pair("", child));
          }
          root.add_child(ir_name.c_str(), pt);

          std::stringstream ss;
          boost::property_tree::json_parser::write_json(ss, root);

          m_ipc_conn->send(ack);
        } catch (const std::exception &e) {
          LOG(ERROR) << "got exception " << typeid(e).name()
                     << " with what: " << e.what();
        }
      });

  m_client->add_handler<nbre_nr_handle_req>(
      [this](std::shared_ptr<nbre_nr_handle_req> req) {
        LOG(INFO) << "recv nbre_nr_handle_req";
        try {
          uint64_t start_block = req->get<p_start_block>();
          uint64_t end_block = req->get<p_end_block>();
          uint64_t nr_version = req->get<p_nr_version>();

          auto handle = param_to_key(start_block, end_block, nr_version);

          auto ack = new_ack_pkg<nbre_nr_handle_ack>(req);
          ack->set<p_nr_handle>(handle);
          neb::rt::nr::nr_handler::instance().start(start_block, end_block,
                                                    nr_version);
          m_ipc_conn->send(ack);

        } catch (const std::exception &e) {
          LOG(ERROR) << "got exception " << typeid(e).name()
                     << " with what: " << e.what();
        }
      });

  m_client->add_handler<nbre_nr_result_by_handle_req>(
      [this](std::shared_ptr<nbre_nr_result_by_handle_req> req) {
        LOG(INFO) << "recv nbre_nr_result_by_handle_req";
        try {
          auto ack = new_ack_pkg<nbre_nr_result_by_handle_ack>(req);
          std::string nr_handle = req->get<p_nr_handle>();
          auto nr_ret =
              neb::rt::nr::nr_handler::instance().get_nr_result(nr_handle);
          if (!std::get<0>(nr_ret)) {
            ack->set<p_nr_result>("");
            m_ipc_conn->send(ack);
            return;
          }

          auto str_ptr = std::get<1>(nr_ret)->serialize_to_string();
          ack->set<p_nr_result>(str_ptr);
          m_ipc_conn->send(ack);
        } catch (const std::exception &e) {
          LOG(ERROR) << "got exception " << typeid(e).name()
                     << " with what: " << e.what();
        }
      });

  m_client->add_handler<nbre_nr_result_by_height_req>(
      [this](std::shared_ptr<nbre_nr_result_by_height_req> req) {
        LOG(INFO) << "recv nbre_nr_result_by_height_req";
        try {
          auto ack = new_ack_pkg<nbre_nr_result_by_height_ack>(req);
          auto height = req->get<p_height>();
          auto ret_ptr =
              neb::rt::dip::dip_handler::instance().get_nr_result(height);
          LOG(INFO) << "nr result \n" << *ret_ptr;
          ack->set<p_nr_result>(*ret_ptr);
          m_ipc_conn->send(ack);
        } catch (const std::exception &e) {
          LOG(ERROR) << "got exception " << typeid(e).name()
                     << " with what: " << e.what();
        }
      });

  m_client->add_handler<nbre_nr_sum_req>(
      [this](std::shared_ptr<nbre_nr_sum_req> req) {
        LOG(INFO) << "recv nbre_nr_sum_req";
        try {
          auto ack = new_ack_pkg<nbre_nr_sum_ack>(req);
          auto height = req->get<p_height>();
          auto ret_ptr =
              neb::rt::dip::dip_handler::instance().get_nr_sum(height);
          LOG(INFO) << "nr sum \n" << *ret_ptr;
          ack->set<p_nr_sum>(*ret_ptr);
          m_ipc_conn->send(ack);
        } catch (const std::exception &e) {
          LOG(ERROR) << "got exception " << typeid(e).name()
                     << " with what: " << e.what();
        }
      });

  m_client->add_handler<nbre_dip_reward_req>(
      [this](std::shared_ptr<nbre_dip_reward_req> req) {
        LOG(INFO) << "recv nbre_dip_reward_req";
        try {
          auto ack = new_ack_pkg<nbre_dip_reward_ack>(req);
          auto height = req->get<p_height>();
          auto ret_ptr =
              neb::rt::dip::dip_handler::instance().get_dip_reward(height);
          ack->set<p_dip_reward>(*ret_ptr);
          m_ipc_conn->send(ack);
        } catch (const std::exception &e) {
          LOG(ERROR) << "got exception " << typeid(e).name()
                     << " with what: " << e.what();
        }
      });

  m_client->add_handler<nbre_ir_transactions_req>(
      [this](std::shared_ptr<nbre_ir_transactions_req> req) {
        try {
          ir_warden::instance().on_receive_ir_transactions(req);
        } catch (const std::exception &e) {
          LOG(ERROR) << "got exception " << typeid(e).name()
                     << " with what: " << e.what();
        }
      });
}
} // namespace core
} // namespace neb
