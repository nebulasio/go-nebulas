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
#include "cmd/dummy_neb/dummies/random_dummy.h"
#include "cmd/dummy_neb/dummy_callback.h"
#include "cmd/dummy_neb/generator/checkers.h"
#include "fs/util.h"

random_dummy::random_dummy(const std::string &name, int initial_account_num,
                           nas initial_nas, double account_increase_ratio,
                           const std::string &rpc_listen, uint16_t rpc_port)
    : dummy_base(name), m_initial_account_num(initial_account_num),
      m_initial_nas(initial_nas),
      m_account_increase_ratio(account_increase_ratio), m_auth_ratio(0),
      m_rpc_listen(rpc_listen), m_rpc_port(rpc_port) {
  m_cli_generator = std::make_unique<cli_generator>();
  m_thread = std::make_unique<std::thread>([this]() {
    ff::net::net_nervure nn;
    m_p_nn = &nn;
    ff::net::typed_pkg_hub hub;
    hub.tcp_to_recv_pkg<cli_brief_req_t>(
        [this](std::shared_ptr<cli_brief_req_t> req,
               ff::net::tcp_connection_base *conn) {
          auto ack = std::make_shared<cli_brief_ack_t>();
          ack->set<p_height>(m_current_height);
          ack->set<p_account_num>(m_all_accounts.size());
          // ack->set<p_nr_ir_status>(std::string());
          // ack->set<p_auth_ir_status>(std::string());
          // ack->set<p_dip_ir_status>(std::string());
          auto status = checker_tasks::instance().status();
          std::string fp =
              neb::fs::join_path(neb::configuration::instance().nbre_db_dir(),
                                 "checker_status.txt");
          std::ofstream ifp(fp);
          ifp << status;
          ifp.close();
          ack->set<p_checker_status>(fp);
          conn->send(ack);
        });

    hub.tcp_to_recv_pkg<cli_submit_ir_t>(
        [this](std::shared_ptr<cli_submit_ir_t> req,
               ff::net::tcp_connection_base *conn) {
          auto ack = std::make_shared<cli_submit_ack_t>();
          ack->set<p_result>("got ir");
          conn->send(ack);
          m_pkgs.push_back(req);
        });
    hub.to_recv_pkg<nbre_nr_handle_req>(
        [this](std::shared_ptr<nbre_nr_handle_req> req) {
          m_pkgs.push_back(req);
        });

    hub.to_recv_pkg<nbre_nr_result_by_handle_req>(
        [this](std::shared_ptr<nbre_nr_result_by_handle_req> req) {
          m_pkgs.push_back(req);
        });

    hub.to_recv_pkg<nbre_nr_result_by_height_req>(
        [this](std::shared_ptr<nbre_nr_result_by_height_req> req) {
          LOG(INFO) << "dummy server recv cli nr result req with height "
                    << req->get<p_height>();
          m_pkgs.push_back(req);
        });

    hub.to_recv_pkg<nbre_nr_sum_req>(
        [this](std::shared_ptr<nbre_nr_sum_req> req) {
          m_pkgs.push_back(req);
        });
    hub.to_recv_pkg<nbre_dip_reward_req>(
        [this](std::shared_ptr<nbre_dip_reward_req> req) {
          m_pkgs.push_back(req);
        });

    nn.get_event_handler()
        ->listen<::ff::net::event::more::tcp_server_accept_connection>(
            [this](::ff::net::tcp_connection_base_ptr conn) { m_conn = conn; });

    nn.add_pkg_hub(hub);
    nn.add_tcp_server(m_rpc_listen, m_rpc_port);
    nn.run();
  });
}

random_dummy::~random_dummy() {
  if (m_p_nn) {
    m_p_nn->stop();
  }
  LOG(INFO) << "to kill thread";
  if (m_thread)
    m_thread->join();
  LOG(INFO) << "kill thread done";
}

std::shared_ptr<generate_block> random_dummy::generate_LIB_block() {

  handle_cli_pkgs();

  std::shared_ptr<generate_block> ret =
      std::make_shared<generate_block>(&m_all_accounts, m_current_height);

  if (m_current_height == 0) {
    genesis_generator g(ret.get(), m_initial_account_num, m_initial_nas);
    g.run();
  } else {
    int account_num = m_account_increase_ratio * m_initial_account_num;
    int tx_num = account_num + std::rand() % m_initial_account_num;
    LOG(INFO) << "transaction num is : " << tx_num;
    m_tx_gen = std::make_unique<transaction_generator>(
        &m_all_accounts, ret.get(),
        m_account_increase_ratio * m_initial_account_num, tx_num);
    m_tx_gen->run();

    if (std::abs(m_auth_ratio) > 1e-9 &&
        std::rand() % 1000 < m_auth_ratio * 1000 &&
        m_current_height % m_auth_interval == 0) {
      m_auth_gen =
          std::make_unique<auth_table_generator>(&m_all_accounts, ret.get());
      address_t nr_admin_addr = m_auth_admin_addr;
      address_t dip_admin_addr = m_auth_admin_addr;
      m_nr_admin_addr = nr_admin_addr;
      m_dip_admin_addr = dip_admin_addr;
      m_auth_gen->set_auth_admin_addr(m_auth_admin_addr);
      m_auth_gen->set_nr_admin_addr(nr_admin_addr);
      m_auth_gen->set_dip_admin_addr(dip_admin_addr);
      m_auth_gen->run();
    }

    if (std::abs(m_nr_ratio) > 1e-9 && std::rand() % 1000 < m_nr_ratio * 1000 &&
        m_current_height % m_nr_interval == 0) {
      if (m_nr_admin_addr.empty()) {
        m_nr_admin_addr = m_auth_admin_addr;
      }
      m_nr_gen = std::make_unique<nr_ir_generator>(ret.get(), m_nr_admin_addr);
      random_increase_version(m_nr_version);
      m_nr_gen->m_major_version = m_nr_version.major_version();
      m_nr_gen->m_minor_version = m_nr_version.minor_version();
      m_nr_gen->m_patch_version = m_nr_version.patch_version();

      m_nr_gen->run();
    }

    if (std::abs(m_dip_ratio) > 1e-9 &&
        std::rand() % 1000 < m_dip_ratio * 1000 &&
        m_current_height % m_dip_interval == 0) {
      if (m_dip_admin_addr.empty()) {
        m_dip_admin_addr = m_auth_admin_addr;
      }
      m_dip_gen =
          std::make_unique<dip_ir_generator>(ret.get(), m_dip_admin_addr);
      random_increase_version(m_dip_version);
      m_dip_gen->m_major_version = m_dip_version.major_version();
      m_dip_gen->m_minor_version = m_dip_version.minor_version();
      m_dip_gen->m_patch_version = m_dip_version.patch_version();
      m_dip_gen->m_nr_version = m_nr_version.data();

      m_dip_gen->run();
    }

    if (std::abs(m_contract_ratio) > 1e-9 &&
        std::rand() % 1000 < m_contract_ratio * 1000) {
      m_contract_gen = std::make_unique<contract_generator>(ret.get(), 1);
      m_contract_gen->run();
    }
    if (std::abs(m_call_ratio) > 1e-9 &&
        std::rand() % 1000 < m_call_ratio * 1000) {
      m_call_gen = std::make_unique<call_tx_generator>(
          ret.get(), std::rand() % (m_all_accounts.size() / 5));
      m_call_gen->run();
    }

    m_cli_generator->m_auth_admin_addr = m_auth_admin_addr;
    m_cli_generator->m_nr_admin_addr = m_nr_admin_addr;
    m_cli_generator->m_dip_admin_addr = m_dip_admin_addr;
    m_cli_generator->update_info(ret.get());
    m_cli_generator->run();
    m_auth_admin_addr = m_cli_generator->m_auth_admin_addr;
    m_nr_admin_addr = m_cli_generator->m_nr_admin_addr;
    m_dip_admin_addr = m_cli_generator->m_dip_admin_addr;
  }

  m_current_height++;
  return ret;
}

address_t
random_dummy::enable_auth_gen_with_ratio(double auth_ratio,
                                         block_height_t block_interval) {
  if (m_current_height == 0)
    generate_LIB_block();
  corepb::Account *admin_account = m_all_accounts.random_user_account();
  address_t admin_addr = neb::to_address(admin_account->address());
  m_auth_admin_addr = admin_addr;
  m_auth_ratio = auth_ratio;
  m_auth_interval = block_interval;
  return admin_addr;
}
void random_dummy::enable_nr_ir_with_ratio(double nr_ratio,
                                           block_height_t block_interval) {
  if (m_current_height == 0)
    generate_LIB_block();
  m_nr_ratio = nr_ratio;
  m_nr_interval = block_interval;
}
void random_dummy::enable_dip_ir_with_ratio(double dip_ratio,
                                            block_height_t block_interval) {
  if (m_current_height == 0)
    generate_LIB_block();
  m_dip_ratio = dip_ratio;
  m_dip_interval = block_interval;
}

void random_dummy::enable_call_tx_with_ratio(double contract_ratio,
                                             double call_ratio) {
  if (m_current_height == 0)
    generate_LIB_block();
  m_contract_ratio = contract_ratio;
  m_call_ratio = call_ratio;
}

std::shared_ptr<checker_task_base> random_dummy::generate_checker_task() {
  std::shared_ptr<checker_task_base> ret;
  if (m_current_height > 0 && m_current_height % 2 == 0) {
    int m = std::rand() % 5;
    switch (m) {
    case 0:
      if (!m_version_checker) {
        m_version_checker = std::make_shared<nbre_version_checker>();
        ret = m_version_checker;
      }
      break;
    case 1: {
      uint64_t t1 = std::rand() % m_current_height;
      uint64_t t2 = std::rand() % m_current_height;
      ret = std::make_shared<nbre_nr_handle_check>(std::min(t1, t2),
                                                   std::max(t1, t2));
      break;
    }
    case 2:
      break;
    case 3:
      ret = std::make_shared<nbre_dip_reward_check>(std::rand() %
                                                    m_current_height);
      break;
    case 4:
      break;
    }

    if (ret) {
      checker_tasks::instance().add_task(ret);
    }
  }
  return ret;
}

address_t random_dummy::get_auth_admin_addr() {
  if (m_current_height == 0)
    generate_LIB_block();
  if (m_auth_admin_addr.empty()) {
    m_auth_admin_addr = m_all_accounts.random_user_addr();
  }
  return m_auth_admin_addr;
}

void random_dummy::handle_cli_pkgs() {
  while (!m_pkgs.empty()) {
    auto ret = m_pkgs.try_pop_front();
    if (!ret.first)
      continue;
    auto pkg = ret.second;
    if (pkg->type_id() == cli_submit_ir_pkg) {
      m_cli_generator->append_pkg(pkg);
    } else if (pkg->type_id() == nbre_nr_handle_req_pkg) {
      nbre_nr_handle_req *req = (nbre_nr_handle_req *)pkg.get();

      callback_handler::instance().add_nr_handler(
          req->get<p_holder>(),
          [this](uint64_t holder, const char *nr_handle_id) {
            std::shared_ptr<nbre_nr_handle_ack> ack =
                std::make_shared<nbre_nr_handle_ack>();
            ack->set<p_holder>(holder);
            ack->set<p_nr_handle>(std::string(nr_handle_id));
            m_conn->send(ack);
          });
      LOG(INFO) << "forward nr handle req";
      ipc_nbre_nr_handle(reinterpret_cast<void *>(req->get<p_holder>()),
                         req->get<p_start_block>(), req->get<p_end_block>(),
                         req->get<p_nr_version>());

    } else if (pkg->type_id() == nbre_nr_result_by_handle_req_pkg) {
      nbre_nr_result_by_handle_req *req =
          (nbre_nr_result_by_handle_req *)pkg.get();
      callback_handler::instance().add_nr_result_handler(
          req->get<p_holder>(), [this](uint64_t holder, const char *nr_result) {
            std::shared_ptr<nbre_nr_result_by_handle_ack> ack =
                std::make_shared<nbre_nr_result_by_handle_ack>();
            ack->set<p_holder>(holder);
            ack->set<p_nr_result>(std::string(nr_result));
            m_conn->send(ack);
          });
      ipc_nbre_nr_result_by_handle(
          reinterpret_cast<void *>(req->get<p_holder>()),
          req->get<p_nr_handle>().c_str());
    } else if (pkg->type_id() == nbre_nr_result_by_height_req_pkg) {
      LOG(INFO) << "handle pkg nr result by height req";
      nbre_nr_result_by_height_req *req =
          (nbre_nr_result_by_height_req *)pkg.get();
      callback_handler::instance().add_nr_result_by_height_handler(
          req->get<p_holder>(), [this](uint64_t holder, const char *nr_result) {
            std::shared_ptr<nbre_nr_result_by_height_ack> ack =
                std::make_shared<nbre_nr_result_by_height_ack>();
            ack->set<p_holder>(holder);
            ack->set<p_nr_result>(std::string(nr_result));
            m_conn->send(ack);
          });
      ipc_nbre_nr_result_by_height(
          reinterpret_cast<void *>(req->get<p_holder>()), req->get<p_height>());
    } else if (pkg->type_id() == nbre_nr_sum_req_pkg) {
      nbre_nr_sum_req *req = (nbre_nr_sum_req *)pkg.get();
      callback_handler::instance().add_nr_sum_handler(
          req->get<p_holder>(), [this](uint64_t holder, const char *nr_sum) {
            std::shared_ptr<nbre_nr_sum_ack> ack =
                std::make_shared<nbre_nr_sum_ack>();
            ack->set<p_holder>(holder);
            ack->set<p_nr_sum>(std::string(nr_sum));
            m_conn->send(ack);
          });
      ipc_nbre_nr_sum(reinterpret_cast<void *>(req->get<p_holder>()),
                      req->get<p_height>());
    } else if (pkg->type_id() == nbre_dip_reward_req_pkg) {
      nbre_dip_reward_req *req = (nbre_dip_reward_req *)pkg.get();
      callback_handler::instance().add_dip_reward_handler(
          req->get<p_holder>(),
          [this](uint64_t holder, const char *dip_reward) {
            std::shared_ptr<nbre_dip_reward_ack> ack =
                std::make_shared<nbre_dip_reward_ack>();
            ack->set<p_holder>(holder);
            ack->set<p_dip_reward>(std::string(dip_reward));
            m_conn->send(ack);
          });
      ipc_nbre_dip_reward(reinterpret_cast<void *>(req->get<p_holder>()),
                          req->get<p_height>(), req->get<p_version>());
    } else {
      LOG(INFO) << "pkg type id " << pkg->type_id() << " not found";
    }
  }
}
