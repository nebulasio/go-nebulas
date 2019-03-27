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
#include "cmd/dummy_neb/dummies/stress_dummy.h"
#include "cmd/dummy_neb/dummy_callback.h"
#include "cmd/dummy_neb/generator/checkers.h"
#include "fs/util.h"

stress_dummy::stress_dummy(const std::string &name, int initial_account_num,
                           nas initial_nas, int account_num, int contract_num,
                           int transfer_tx_num, int call_tx_num,
                           const std::string &rpc_listen, uint16_t rpc_port)
    : dummy_base(name), m_initial_account_num(initial_account_num),
      m_initial_nas(initial_nas), m_account_num(account_num),
      m_contract_num(contract_num), m_transfer_tx_num(transfer_tx_num),
      m_call_tx_num(call_tx_num), m_rpc_listen(rpc_listen),
      m_rpc_port(rpc_port) {

  m_cli_generator = std::make_unique<cli_generator>();
  m_thread = std::make_unique<std::thread>([this]() {
    ff::net::net_nervure nn;
    m_p_nn = &nn;
    ff::net::typed_pkg_hub hub;

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

stress_dummy::~stress_dummy() {
  if (m_p_nn) {
    m_p_nn->stop();
  }
  LOG(INFO) << "to kill thread";
  if (m_thread)
    m_thread->join();
  LOG(INFO) << "kill thread done";
}

void stress_dummy::handle_cli_pkgs() {
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

void stress_dummy::init_ir_and_accounts() {
  LOG(INFO) << "init ir and accounts";
  std::shared_ptr<generate_block> ret =
      std::make_shared<generate_block>(&m_all_accounts, m_current_height);
  for (auto i = 0; i < m_account_num; i++) {
    ret->gen_user_account();
  }
  LOG(INFO) << "gen user account done";

  m_contract_gen =
      std::make_unique<contract_generator>(ret.get(), m_initial_account_num);
  m_contract_gen->run();

  // gen auth
  m_auth_gen =
      std::make_unique<auth_table_generator>(&m_all_accounts, ret.get());
  m_auth_gen->set_auth_admin_addr(m_auth_admin_addr);
  m_auth_gen->set_nr_admin_addr(m_auth_admin_addr);
  m_auth_gen->set_dip_admin_addr(m_auth_admin_addr);
  m_auth_gen->run();
  LOG(INFO) << "gen auth done";

  // gen nr
  neb::version nr_v(0);
  m_nr_gen = std::make_unique<nr_ir_generator>(ret.get(), m_auth_admin_addr);
  random_increase_version(nr_v);
  m_nr_gen->m_major_version = nr_v.major_version();
  m_nr_gen->m_minor_version = nr_v.minor_version();
  m_nr_gen->m_patch_version = nr_v.patch_version();
  m_nr_gen->run();
  LOG(INFO) << "gen nr done";

  // gen dip
  neb::version dip_v(0);
  m_dip_gen = std::make_unique<dip_ir_generator>(ret.get(), m_auth_admin_addr);
  random_increase_version(dip_v);
  m_dip_gen->m_major_version = dip_v.major_version();
  m_dip_gen->m_minor_version = dip_v.minor_version();
  m_dip_gen->m_patch_version = dip_v.patch_version();
  m_dip_gen->m_nr_version = nr_v.data();
  m_dip_gen->run();
  LOG(INFO) << "gen dip done";
}

std::shared_ptr<generate_block> stress_dummy::generate_LIB_block() {

  handle_cli_pkgs();

  std::shared_ptr<generate_block> ret =
      std::make_shared<generate_block>(&m_all_accounts, m_current_height);

  if (m_current_height == 0) {
    genesis_generator g(ret.get(), m_initial_account_num, m_initial_nas);
    g.run();
    m_current_height++;
    return ret;
  }

  if (m_current_height == 1) {
    init_ir_and_accounts();
  }

  int32_t week_blocks = 7 * 24 * 3600 / 15;
  double tx_ratio = m_account_num * 1.0 / week_blocks;
  double contract_ratio = m_contract_num * 1.0 / week_blocks;
  double call_ratio = m_call_tx_num * 1.0 / week_blocks;

  int32_t tx_num = tx_ratio;
  int32_t contract_num = contract_ratio;
  int32_t call_num = call_ratio;

  // binary tx
  if (tx_ratio > 1.0) {
    tx_ratio -= tx_num;
  }
  if (std::rand() % week_blocks < tx_ratio * week_blocks) {
    tx_num++;
  }
  m_tx_gen = std::make_unique<transaction_generator>(&m_all_accounts, ret.get(),
                                                     0, tx_num);
  m_tx_gen->run();

  // deploy tx
  if (contract_ratio > 1.0) {
    contract_ratio -= contract_num;
  }
  if (std::rand() % week_blocks < contract_ratio * week_blocks) {
    contract_num++;
  }
  m_contract_gen =
      std::make_unique<contract_generator>(ret.get(), contract_num);
  m_contract_gen->run();

  // call tx
  if (call_ratio > 1.0) {
    call_ratio -= call_num;
  }
  if (std::rand() % week_blocks < call_ratio * week_blocks) {
    call_num++;
  }
  m_call_gen = std::make_unique<call_tx_generator>(ret.get(), call_num);
  m_call_gen->run();

  m_current_height++;
  return ret;
}

std::shared_ptr<checker_task_base> stress_dummy::generate_checker_task() {
  auto ret = std::make_shared<nbre_version_checker>();
  return ret;
}

address_t stress_dummy::get_auth_admin_addr() {
  if (m_current_height == 0) {
    generate_LIB_block();
  }
  if (m_auth_admin_addr.empty()) {
    m_auth_admin_addr = m_all_accounts.random_user_addr();
  }
  return m_auth_admin_addr;
}

