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
#include "core/neb_ipc/server/ipc_server_endpoint.h"
#include "common/configuration.h"
#include "core/neb_ipc/ipc_pkg.h"
#include "core/neb_ipc/server/ipc_configuration.h"
#include "fs/util.h"
#include <atomic>
#include <boost/property_tree/json_parser.hpp>
#include <boost/property_tree/ptree.hpp>
#include <condition_variable>

namespace neb {
namespace core {

ipc_server_endpoint::ipc_server_endpoint() : m_client(nullptr){};

ipc_server_endpoint::~ipc_server_endpoint() {
  if (m_thread) {
    m_thread->join();
    m_thread.reset();
  }
}

bool ipc_server_endpoint::start() {
  if (!check_path_exists()) {
    LOG(ERROR) << "nbre path not exist";
    return false;
  }
  if (!ipc_callback_holder::instance().check_all_callbacks()) {
    LOG(ERROR) << "nbre missing callback";
    return false;
  }

  std::atomic_bool init_done(false);
  std::mutex local_mutex;
  std::condition_variable local_cond_var;
  m_got_exception_when_start_nbre = false;
  m_thread = std::unique_ptr<std::thread>(new std::thread([&, this]() {
    try {
      auto shm_service_name_str =
          std::string("nbre.") +
          neb::shm_configuration::instance().shm_name_identity();
      auto shm_service_name = shm_service_name_str.c_str();
      {
        ipc_util_t us(shm_service_name, 128, 128);
        us.reset();
      }

      LOG(INFO) << "start ipc server ";
      m_ipc_server = std::unique_ptr<ipc_server_t>(
          new ipc_server_t(shm_service_name, 128, 128));

      m_ipc_server->init_local_env();
      m_request_timer =
          std::unique_ptr<api_request_timer>(new api_request_timer(
              m_ipc_server.get(), &ipc_callback_holder::instance()));

      LOG(INFO) << "nbre ipc init done!";
      add_all_callbacks();

      m_client_watcher =
          std::unique_ptr<ipc_client_watcher>(new ipc_client_watcher(
              neb::core::ipc_configuration::instance().nbre_exe_name()));
      m_client_watcher->start();

      local_mutex.lock();
      init_done = true;
      local_mutex.unlock();
      local_cond_var.notify_one();

      m_ipc_server->run();

      LOG(INFO) << "ipc server stopped!";
    } catch (const std::exception &e) {
      LOG(ERROR) << "get exception when start nbre, " << typeid(e).name()
                 << ", " << e.what();
      m_got_exception_when_start_nbre = true;
      local_cond_var.notify_one();
    } catch (...) {
      LOG(ERROR) << "get unknown exception when start nbre";
      m_got_exception_when_start_nbre = true;
      local_cond_var.notify_one();
    }
  }));

  std::unique_lock<std::mutex> _l(local_mutex);
  if (!init_done) {
    local_cond_var.wait(_l);
  }
  if (m_got_exception_when_start_nbre)
    return false;
  try {
    m_ipc_server->wait_until_client_start();
  } catch (const std::exception &e) {
    LOG(ERROR) << "get exception when wait client start, " << typeid(e).name()
               << e.what();
    return false;
  } catch (...) {
    LOG(ERROR) << "get unknown exception when wait nbre";
    return false;
  }
  return true;
}

bool ipc_server_endpoint::check_path_exists() {
  return neb::fs::exists(
      neb::core::ipc_configuration::instance().nbre_exe_name());
}

void ipc_server_endpoint::init_params(const nbre_params_t params) {
  neb::core::ipc_configuration::instance().nbre_root_dir() =
      params.m_nbre_root_dir;
  neb::core::ipc_configuration::instance().nbre_exe_name() =
      params.m_nbre_exe_name;
  neb::core::ipc_configuration::instance().neb_db_dir() = params.m_neb_db_dir;
  neb::core::ipc_configuration::instance().nbre_db_dir() = params.m_nbre_db_dir;
  neb::core::ipc_configuration::instance().nbre_log_dir() =
      params.m_nbre_log_dir;
  neb::core::ipc_configuration::instance().admin_pub_addr() =
      params.m_admin_pub_addr;
  neb::core::ipc_configuration::instance().nbre_start_height() =
      params.m_nbre_start_height;

  auto s = std::chrono::system_clock::from_time_t(0);
  auto e = std::chrono::system_clock::now();
  auto cnt = std::chrono::duration_cast<std::chrono::seconds>(e - s).count();
  cnt &= 0xff;
  neb::shm_configuration::instance().shm_name_identity() = std::to_string(cnt);
}

void ipc_server_endpoint::add_all_callbacks() {
  LOG(INFO) << "ipc server pointer: " << (void *)m_ipc_server.get();
  ipc_server_t *p = m_ipc_server.get();

  m_callbacks = &(ipc_callback_holder::instance());
  m_ipc_server->add_handler<ipc_pkg::nbre_version_ack>(
      [this](ipc_pkg::nbre_version_ack *msg) {
        m_request_timer->remove_api(msg->m_holder);
        ipc_callback_holder::instance().m_nbre_version_callback(
            ipc_status_succ, msg->m_holder, msg->get<ipc_pkg::major>(),
            msg->get<ipc_pkg::minor>(), msg->get<ipc_pkg::patch>());
      });

  m_ipc_server->add_handler<ipc_pkg::nbre_init_req>(
      [p](ipc_pkg::nbre_init_req *) {
        LOG(INFO) << "get init req ";
        ipc_pkg::nbre_init_ack *ack = p->construct<ipc_pkg::nbre_init_ack>(
            nullptr, p->default_allocator());

        std::string nbre_root_dir =
            neb::core::ipc_configuration::instance().nbre_root_dir();
        neb::ipc::char_string_t cstr_nbre_root_dir(nbre_root_dir.c_str(),
                                                   p->default_allocator());
        ack->set<ipc_pkg::nbre_root_dir>(cstr_nbre_root_dir);

        std::string nbre_exe_name =
            neb::core::ipc_configuration::instance().nbre_exe_name();
        neb::ipc::char_string_t cstr_nbre_exe_name(nbre_exe_name.c_str(),
                                                   p->default_allocator());
        ack->set<ipc_pkg::nbre_exe_name>(cstr_nbre_exe_name);

        std::string neb_db_dir =
            neb::core::ipc_configuration::instance().neb_db_dir();
        neb::ipc::char_string_t cstr_neb_db_dir(neb_db_dir.c_str(),
                                                p->default_allocator());
        ack->set<ipc_pkg::neb_db_dir>(cstr_neb_db_dir);

        std::string nbre_db_dir =
            neb::core::ipc_configuration::instance().nbre_db_dir();
        neb::ipc::char_string_t cstr_nbre_db_dir(nbre_db_dir.c_str(),
                                                 p->default_allocator());
        ack->set<ipc_pkg::nbre_db_dir>(cstr_nbre_db_dir);

        std::string nbre_log_dir =
            neb::core::ipc_configuration::instance().nbre_log_dir();
        neb::ipc::char_string_t cstr_nbre_log_dir(nbre_log_dir.c_str(),
                                                  p->default_allocator());
        ack->set<ipc_pkg::nbre_log_dir>(cstr_nbre_log_dir);

        std::string admin_pub_addr =
            neb::core::ipc_configuration::instance().admin_pub_addr().c_str();
        neb::ipc::char_string_t cstr_admin_pub_addr(admin_pub_addr.c_str(),
                                                    p->default_allocator());
        ack->set<ipc_pkg::admin_pub_addr>(cstr_admin_pub_addr);

        ack->set<ipc_pkg::nbre_start_height>(
            neb::core::ipc_configuration::instance().nbre_start_height());
        p->push_back(ack);
      });

  m_ipc_server->add_handler<ipc_pkg::nbre_ir_list_ack>(
      [this](ipc_pkg::nbre_ir_list_ack *msg) {
        m_request_timer->remove_api(msg->m_holder);

        auto ir_name_list = msg->get<ipc_pkg::ir_name_list>();
        boost::property_tree::ptree pt, root;

        for (auto &ir_name : ir_name_list) {
          boost::property_tree::ptree child;
          child.put("", ir_name);
          pt.push_back(std::make_pair("", child));
        }
        root.add_child(neb::configuration::instance().ir_list_name(), pt);

        std::stringstream ss;
        boost::property_tree::json_parser::write_json(ss, root);

        ipc_callback_holder::instance().m_nbre_ir_list_callback(
            ipc_status_succ, msg->m_holder, ss.str().c_str());
      });

  m_ipc_server->add_handler<ipc_pkg::nbre_ir_versions_ack>(
      [this](ipc_pkg::nbre_ir_versions_ack *msg) {
        m_request_timer->remove_api(msg->m_holder);

        auto ir_name = msg->get<ipc_pkg::ir_name>();
        auto ir_versions = msg->get<ipc_pkg::ir_versions>();
        boost::property_tree::ptree pt, root;

        for (auto &v : ir_versions) {
          boost::property_tree::ptree child;
          child.put("", v);
          pt.push_back(std::make_pair("", child));
        }
        root.add_child(ir_name.c_str(), pt);

        std::stringstream ss;
        boost::property_tree::json_parser::write_json(ss, root);

        ipc_callback_holder::instance().m_nbre_ir_versions_callback(
            ipc_status_succ, msg->m_holder, ss.str().c_str());
      });

  m_ipc_server->add_handler<ipc_pkg::nbre_nr_handler_ack>(
      [this](ipc_pkg::nbre_nr_handler_ack *msg) {
        m_request_timer->remove_api(msg->m_holder);
        ipc_callback_holder::instance().m_nbre_nr_handler_callback(
            ipc_status_succ, msg->m_holder,
            msg->get<ipc_pkg::nr_handler_id>().c_str());
      });

  m_ipc_server->add_handler<ipc_pkg::nbre_nr_result_ack>(
      [this](ipc_pkg::nbre_nr_result_ack *msg) {
        m_request_timer->remove_api(msg->m_holder);
        ipc_callback_holder::instance().m_nbre_nr_result_callback(
            ipc_status_succ, msg->m_holder,
            msg->get<ipc_pkg::nr_result>().c_str());
      });

  m_ipc_server->add_handler<ipc_pkg::nbre_dip_reward_ack>(
      [this](ipc_pkg::nbre_dip_reward_ack *msg) {
        m_request_timer->remove_api(msg->m_holder);
        ipc_callback_holder::instance().m_nbre_dip_reward_callback(
            ipc_status_succ, msg->m_holder,
            msg->get<ipc_pkg::dip_reward>().c_str());
      });
}

int ipc_server_endpoint::send_nbre_version_req(void *holder, uint64_t height) {
  CHECK_NBRE_STATUS(m_callbacks->m_nbre_version_callback);

  m_request_timer->issue_api(
      holder,
      [holder, height, this]() {
        ipc_pkg::nbre_version_req *req =
            m_ipc_server->construct<ipc_pkg::nbre_version_req>(
                holder, m_ipc_server->default_allocator());
        if (req == nullptr) {
          return; // will call timeout later
        }
        req->set<ipc_pkg::height>(height);
        m_ipc_server->push_back(req);
      },
      m_callbacks->m_nbre_version_callback);
  return ipc_status_succ;
}

int ipc_server_endpoint::send_nbre_ir_list_req(void *holder) {
  CHECK_NBRE_STATUS(m_callbacks->m_nbre_ir_list_callback);

  m_request_timer->issue_api(
      holder,
      [holder, this]() {
        ipc_pkg::nbre_ir_list_req *req =
            m_ipc_server->construct<ipc_pkg::nbre_ir_list_req>(
                holder, m_ipc_server->default_allocator());
        if (req == nullptr) {
          return; // will call timeout later
        }
        m_ipc_server->push_back(req);
      },
      m_callbacks->m_nbre_ir_list_callback);
  return ipc_status_succ;
}

int ipc_server_endpoint::send_nbre_ir_versions_req(void *holder,
                                                   const std::string &ir_name) {
  m_request_timer->issue_api(
      holder,
      [holder, &ir_name, this]() {
        ipc_pkg::nbre_ir_versions_req *req =
            m_ipc_server->construct<ipc_pkg::nbre_ir_versions_req>(
                holder, m_ipc_server->default_allocator());
        if (req == nullptr) {
          return; // will call timeout later
        }

        neb::ipc::char_string_t cstr_ir_name(ir_name.c_str(),
                                             m_ipc_server->default_allocator());
        LOG(INFO) << cstr_ir_name;
        req->set<ipc_pkg::ir_name>(cstr_ir_name);

        m_ipc_server->push_back(req);
      },
      m_callbacks->m_nbre_ir_versions_callback);
  return ipc_status_succ;
}

int ipc_server_endpoint::send_nbre_nr_handler_req(void *holder,
                                                  uint64_t start_block,
                                                  uint64_t end_block,
                                                  uint64_t nr_version) {
  CHECK_NBRE_STATUS(m_callbacks->m_nbre_nr_handler_callback);

  m_request_timer->issue_api(
      holder,
      [holder, start_block, end_block, nr_version, this]() {
        ipc_pkg::nbre_nr_handler_req *req =
            m_ipc_server->construct<ipc_pkg::nbre_nr_handler_req>(
                holder, m_ipc_server->default_allocator());
        if (req == nullptr) {
          return;
        }
        req->set<ipc_pkg::start_block>(start_block);
        req->set<ipc_pkg::end_block>(end_block);
        req->set<ipc_pkg::nr_version>(nr_version);
        m_ipc_server->push_back(req);
      },
      m_callbacks->m_nbre_nr_handler_callback);
  return ipc_status_succ;
}

int ipc_server_endpoint::send_nbre_nr_result_req(
    void *holder, const std::string &nr_handler_id) {
  CHECK_NBRE_STATUS(m_callbacks->m_nbre_nr_result_callback);

  m_request_timer->issue_api(
      holder,
      [holder, nr_handler_id, this]() {
        ipc_pkg::nbre_nr_result_req *req =
            m_ipc_server->construct<ipc_pkg::nbre_nr_result_req>(
                holder, m_ipc_server->default_allocator());
        if (req == nullptr) {
          return;
        }

        neb::ipc::char_string_t cstr_nr_handler_id(
            nr_handler_id.c_str(), m_ipc_server->default_allocator());
        req->set<ipc_pkg::nr_handler_id>(cstr_nr_handler_id);
        m_ipc_server->push_back(req);
      },
      m_callbacks->m_nbre_nr_result_callback);

  return ipc_status_succ;
}

int ipc_server_endpoint::send_nbre_dip_reward_req(void *holder,
                                                  uint64_t height) {
  CHECK_NBRE_STATUS(m_callbacks->m_nbre_dip_reward_callback);

  m_request_timer->issue_api(
      holder,
      [holder, height, this]() {
        ipc_pkg::nbre_dip_reward_req *req =
            m_ipc_server->construct<ipc_pkg::nbre_dip_reward_req>(
                holder, m_ipc_server->default_allocator());
        if (req == nullptr) {
          return;
        }

        req->set<ipc_pkg::height>(height);
        m_ipc_server->push_back(req);
        LOG(INFO) << "ipc server send dip reward req, height " << height;
      },
      m_callbacks->m_nbre_dip_reward_callback);

  return ipc_status_succ;
}

void ipc_server_endpoint::shutdown() {
  neb::core::command_queue::instance().send_command(
      std::make_shared<neb::core::exit_command>());
}
}// namespace core
} // namespace neb
