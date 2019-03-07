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
#include "cmd/dummy_neb/generator/cli_generator.h"
#include "cmd/dummy_neb/cli/pkg.h"

cli_generator::cli_generator() : generator_base(nullptr, nullptr, 0, 0) {
  m_thread = std::make_unique<std::thread>([this]() {
    ff::net::net_nervure nn;
    ff::net::typed_pkg_hub hub;
    hub.tcp_to_recv_pkg<cli_brief_req_t>(
        [this](std::shared_ptr<cli_brief_req_t> req,
               ff::net::tcp_connection_base *conn) {
          auto ack = std::make_shared<cli_brief_ack_t>();
          ack->set<p_height>(m_block->height());
          ack->set<p_account_num>(m_all_accounts->size());
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

    nn.add_pkg_hub(hub);
    nn.add_tcp_server("127.0.0.1", 0x1958);
    nn.run();
  });
}
cli_generator::~cli_generator() { m_thread->join(); }

void cli_generator::update_info(generate_block *block) {
  m_block = block;
  m_all_accounts = block->get_all_accounts();
  m_new_tx_num = m_pkgs.size();
}

std::shared_ptr<corepb::Account> cli_generator::gen_account() {
  return nullptr;
}
std::shared_ptr<corepb::Transaction> cli_generator::gen_tx() {
  while (!m_pkgs.empty()) {
    auto ret = m_pkgs.try_pop_front();
    if (!ret.first)
      continue;
    auto pkg = ret.second;
    if (pkg->type_id() == cli_submit_ir_pkg) {
      cli_submit_ir_t *req = (cli_submit_ir_t *)pkg.get();
      if (req->get<p_type>() == "nr") {
        LOG(INFO) << "got nr";
      } else if (req->get<p_type>() == "dip") {
        LOG(INFO) << "got nr";
      } else if (req->get<p_type>() == "auth") {
        LOG(INFO) << "got nr";
      }
    }
  }
  return nullptr;
}
checker_tasks::task_container_ptr_t cli_generator::gen_tasks() {
  return nullptr;
}

