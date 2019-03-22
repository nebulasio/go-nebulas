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
#pragma once
#include "cmd/dummy_neb/dummies/dummy_base.h"
#include "cmd/dummy_neb/generator/generators.h"

class random_dummy : public dummy_base {
public:
  random_dummy(const std::string &name, int initial_account_num,
               nas initial_nas, double account_increase_ratio,
               const std::string &rpc_listen, uint16_t rpc_port);

  address_t enable_auth_gen_with_ratio(double auth_ratio,
                                       block_height_t block_interval);
  void enable_nr_ir_with_ratio(double nr_ratio, block_height_t block_interval);
  void enable_dip_ir_with_ratio(double dip_ratio,
                                block_height_t block_interval);

  void enable_call_tx_with_ratio(double contract_ratio, double call_ratio);

  virtual ~random_dummy();
  virtual std::shared_ptr<generate_block> generate_LIB_block();

  virtual std::shared_ptr<checker_task_base> generate_checker_task();

  virtual address_t get_auth_admin_addr();

protected:
  void handle_cli_pkgs();

protected:
  all_accounts m_all_accounts;
  std::unique_ptr<transaction_generator> m_tx_gen;
  std::unique_ptr<genesis_generator> m_genesis_gen;
  std::unique_ptr<auth_table_generator> m_auth_gen;
  std::unique_ptr<nr_ir_generator> m_nr_gen;
  std::unique_ptr<dip_ir_generator> m_dip_gen;
  std::unique_ptr<call_tx_generator> m_call_gen;
  std::unique_ptr<contract_generator> m_contract_gen;
  std::unique_ptr<cli_generator> m_cli_generator;
  int m_initial_account_num;
  nas m_initial_nas;
  double m_account_increase_ratio;
  double m_auth_ratio;
  double m_nr_ratio;
  double m_dip_ratio;
  double m_contract_ratio;
  double m_call_ratio;

  block_height_t m_auth_interval;
  block_height_t m_nr_interval;
  block_height_t m_dip_interval;

  std::string m_rpc_listen;
  uint16_t m_rpc_port;

  neb::version m_nr_version;
  neb::version m_dip_version;

  address_t m_auth_admin_addr;
  address_t m_nr_admin_addr;
  address_t m_dip_admin_addr;

  std::unique_ptr<std::thread> m_thread;
  neb::util::wakeable_queue<std::shared_ptr<ff::net::package>> m_pkgs;
  ff::net::tcp_connection_base_ptr m_conn;

  std::shared_ptr<nbre_version_checker> m_version_checker;
  ff::net::net_nervure *m_p_nn;
};
