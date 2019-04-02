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
#include "cmd/dummy_neb/dummy_callback.h"
#include "cmd/dummy_neb/dummy_common.h"
#include "cmd/dummy_neb/dummy_driver.h"
#include "common/configuration.h"
#include "fs/util.h"
#include <boost/process.hpp>
#include <boost/program_options.hpp>

namespace po = boost::program_options;
namespace bp = boost::process;
po::variables_map get_variables_map(int argc, char *argv[]) {
  po::options_description desc("Dummy neb");
  // clang-format off
  desc.add_options()("help", "show help message")
    ("list-dummies", "show all dummy names")
    ("run-dummy", po::value<std::string>()->default_value("default_random"), "run a dummy with name (from list-dummies, default [default_random])")
    ("block-interval", po::value<uint64_t>()->default_value(3), "block interval with seconds")
    ("without-clean-db", po::value<bool>()->default_value(true), "run a dummy without clean previous db")
    ("clean-dummy-db", po::value<std::string>(), "clean the db file of a dummy")
    ("glog-log-to-stderr", po::value<bool>()->default_value(false), "glog to stderr")
    ("use-test-blockchain", po::value<bool>()->default_value(true), "use test blockchain")
    ("nipc-listen", po::value<std::string>()->default_value("127.0.0.1"), "nipc listen")
    ("nipc-port", po::value<uint16_t>()->default_value(6987), "nipc port")
    ("rpc-listen", po::value<std::string>()->default_value("127.0.0.1"), "nipc listen")
    ("rpc-port", po::value<uint16_t>()->default_value(0x1958), "nipc port")
    ("enable-nbre-killer", po::value<uint64_t>()->default_value(24*60), "kill nbre periodically in miniutes");
  // clang-format on

  po::variables_map vm;
  po::store(po::parse_command_line(argc, argv, desc), vm);
  po::notify(vm);

  if (vm.count("help")) {
    std::cout << desc << "\n";
    exit(1);
  }

  return vm;
}

void init_dummy_driver(dummy_driver &dd, const std::string &rpc_listen,
                       uint16_t rpc_port) {
  auto default_dummy = std::make_shared<random_dummy>(
      "default_random", 20, 10000_nas, 0.05, rpc_listen, rpc_port);
  default_dummy->enable_auth_gen_with_ratio(1, 1);
  default_dummy->enable_nr_ir_with_ratio(1, 20);
  default_dummy->enable_dip_ir_with_ratio(1, 20);
  default_dummy->enable_call_tx_with_ratio(1, 1);
  dd.add_dummy(default_dummy);

  // auto stress = std::make_shared<stress_dummy>(
  //"stress", 20, 10000_nas, 100, 10, 10000, 1000, rpc_listen, rpc_port);
  // dd.add_dummy(stress);
}

void init_and_start_nbre(const address_t &auth_admin_addr,
                         const std::string &neb_db_dir,
                         const std::string &nipc_listen, uint16_t nipc_port) {

  const char *root_dir = neb::configuration::instance().nbre_root_dir().c_str();
  std::string nbre_path = neb::fs::join_path(root_dir, "bin/nbre");

  set_recv_nbre_version_callback(nbre_version_callback);
  set_recv_nbre_ir_list_callback(nbre_ir_list_callback);
  set_recv_nbre_ir_versions_callback(nbre_ir_versions_callback);
  set_recv_nbre_nr_handle_callback(nbre_nr_handle_callback);
  set_recv_nbre_nr_result_by_handle_callback(nbre_nr_result_callback);
  set_recv_nbre_nr_result_by_height_callback(nbre_nr_result_by_height_callback);
  set_recv_nbre_nr_sum_callback(nbre_nr_sum_callback);
  set_recv_nbre_dip_reward_callback(nbre_dip_reward_callback);

  uint64_t nbre_start_height = 1;
  std::string auth_addr_str = auth_admin_addr.to_base58();
  neb::configuration::instance().neb_db_dir() = neb_db_dir;
  nbre_params_t params{
      root_dir,
      nbre_path.c_str(),
      neb::configuration::instance().neb_db_dir().c_str(),
      neb::configuration::instance().nbre_db_dir().c_str(),
      neb::configuration::instance().nbre_log_dir().c_str(),
      auth_addr_str.c_str(),
      nbre_start_height,
  };

  LOG(INFO) << "auth admin addr: " << auth_admin_addr.to_hex();
  LOG(INFO) << "init auth admin addr: " << params.m_admin_pub_addr;
  params.m_nipc_listen = nipc_listen.c_str();
  params.m_nipc_port = nipc_port;

  auto ret = start_nbre_ipc(params);
  if (ret != ipc_status_succ) {
    nbre_ipc_shutdown();
    exit(-1);
  }
}

int main(int argc, char *argv[]) {

  dummy_driver dd;
  std::thread quiter_thrd([&]() {
    char c = 'a';
    while (c != 'x') {
      std::cin >> c;
    }
    neb::core::command_queue::instance().send_command(
        std::make_shared<neb::core::exit_command>());
    nbre_ipc_shutdown();
    LOG(INFO) << "nbre_ipc_shutdown shut down done";

    dd.shutdown();
  });
  struct _t {
    _t(std::thread &t) : thread(t) {}
    ~_t() {
      thread.join();
      neb::fs::bc_storage_session::instance().release();
    }
    std::thread &thread;
  } __(quiter_thrd);

  std::string root_dir = neb::configuration::instance().nbre_root_dir();
  neb::configuration::instance().neb_db_dir() =
      neb::fs::join_path(root_dir, std::string("dummy_db"));

  FLAGS_logtostderr = false;

  po::variables_map vm = get_variables_map(argc, argv);
  neb::glog_log_to_stderr = vm["glog-log-to-stderr"].as<bool>();
  neb::use_test_blockchain = vm["use-test-blockchain"].as<bool>();
  LOG(INFO) << "log-to-stderr: " << neb::glog_log_to_stderr;

  std::string rpc_listen = vm["rpc-listen"].as<std::string>();
  uint16_t rpc_port = vm["rpc-port"].as<uint16_t>();

  init_dummy_driver(dd, rpc_listen, rpc_port);
  if (vm.count("list-dummies")) {
    auto dummies = dd.get_all_dummy_names();
    std::for_each(dummies.begin(), dummies.end(), [](const std::string &s) {
      std::cout << "\t" << s << std::endl;
    });
    return 1;
  }

  if (vm.count("run-dummy")) {
    std::string dummy_name = vm["run-dummy"].as<std::string>();
    uint64_t interval = vm["block-interval"].as<uint64_t>();
    bool without_clean_db = vm["without-clean-db"].as<bool>();
    if (!without_clean_db) {
      dd.reset_dummy(dummy_name);
    }
    auto dummy = dd.get_dummy_with_name(dummy_name);
    std::string db_path = dummy->db_path();
    bc_storage_session::instance().init(db_path,
                                        neb::fs::storage_open_for_readwrite);
    auto admin_addr = dummy->get_auth_admin_addr();

    std::thread thrd([&dd, dummy_name, interval]() {
      dd.run(dummy_name, interval);
      nbre_ipc_shutdown();
    });

    std::this_thread::sleep_for(std::chrono::seconds(interval));

    std::string nipc_listen = vm["nipc-listen"].as<std::string>();
    uint16_t nipc_port = vm["nipc-port"].as<uint16_t>();
    init_and_start_nbre(admin_addr, db_path, nipc_listen, nipc_port);
    LOG(INFO) << "init nbre done!";
    thrd.join();
    LOG(INFO) << "to quit nbre";
    return 0;
  }
  if (vm.count("clean-dummy-db")) {
    std::string dummy_name = vm["clean-dummy-db"].as<std::string>();
    dd.reset_dummy(dummy_name);
    return 0;
  }

  return 0;
}
