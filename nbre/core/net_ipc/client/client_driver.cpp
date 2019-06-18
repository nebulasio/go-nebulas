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
#include "jit/jit_driver.h"
#if 0
#include "runtime/dip/dip_handler.h"
#endif
#include "runtime/util.h"
#include "runtime/version.h"
#include <boost/filesystem.hpp>
#include <boost/property_tree/json_parser.hpp>
#include <boost/property_tree/ptree.hpp>
#include <ff/functionflow.h>

namespace neb {
namespace core {

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

          auto irs = context->ir_processor()->get_ir_names();

          boost::property_tree::ptree pt, root;
          for (auto &ir : irs) {
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

          auto ir_versions =
              core::context->ir_processor()->get_ir_versions(ir_name);

          boost::property_tree::ptree pt, root;
          for (auto &v : ir_versions) {
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

void client_driver::add_nr_handlers() {
#if 0
  m_client->add_handler<nbre_nr_handle_req>(
      [this](std::shared_ptr<nbre_nr_handle_req> req) {
        LOG(INFO) << "recv nbre_nr_handle_req";
        try {
          uint64_t start_block = req->get<p_start_block>();
          uint64_t end_block = req->get<p_end_block>();
          uint64_t nr_version = req->get<p_nr_version>();

          auto handle = rt::param_to_key(start_block, end_block, nr_version);

          auto ack = new_ack_pkg<nbre_nr_handle_ack>(req);
          ack->set<p_nr_handle>(handle);
          m_nr_handler->start(start_block, end_block, nr_version);
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
          auto nr_ret = m_nr_handler->get_nr_result(nr_handle);
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
#endif
}

void client_driver::add_dip_handlers() {
#if 0
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
#endif
}
} // namespace core
} // namespace neb
