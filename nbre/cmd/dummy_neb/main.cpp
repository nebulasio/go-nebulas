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
#include "cmd/dummy_neb/dummy_common.h"
#include "cmd/dummy_neb/dummy_driver.h"
#include "common/configuration.h"
#include "fs/util.h"
#include <boost/process.hpp>
#include <boost/program_options.hpp>

std::mutex local_mutex;
std::condition_variable local_cond_var;
bool to_quit = false;

void nbre_version_callback(ipc_status_code isc, void *handler, uint32_t major,
                           uint32_t minor, uint32_t patch) {
  LOG(INFO) << "got version: " << major << ", " << minor << ", " << patch;
  std::unique_lock<std::mutex> _l(local_mutex);
  to_quit = true;
  _l.unlock();
  // local_cond_var.notify_one();
}

void nbre_ir_list_callback(ipc_status_code isc, void *handler,
                           const char *ir_name_list) {
  LOG(INFO) << ir_name_list;
  std::unique_lock<std::mutex> _l(local_mutex);
  to_quit = true;
  _l.unlock();
}

void nbre_ir_versions_callback(ipc_status_code isc, void *handler,
                               const char *ir_versions) {
  LOG(INFO) << ir_versions;
  std::unique_lock<std::mutex> _l(local_mutex);
  to_quit = true;
  _l.unlock();
}

void nbre_nr_handle_callback(ipc_status_code isc, void *holder,
                             const char *nr_handle_id) {
  LOG(INFO) << nr_handle_id;
  std::unique_lock<std::mutex> _l(local_mutex);
  to_quit = true;
  _l.unlock();
}

void nbre_nr_result_callback(ipc_status_code isc, void *holder,
                             const char *nr_result) {
  LOG(INFO) << nr_result;
  std::unique_lock<std::mutex> _l(local_mutex);
  to_quit = true;
  _l.unlock();
}

void nbre_dip_reward_callback(ipc_status_code isc, void *holder,
                              const char *dip_reward) {
  LOG(INFO) << dip_reward;
  std::unique_lock<std::mutex> _l(local_mutex);
  to_quit = true;
  _l.unlock();
}
#if 0
int main(int argc, char *argv[]) {
  FLAGS_logtostderr = true;

  //::google::InitGoogleLogging(argv[0]);
  neb::glog_log_to_stderr = true;
  neb::use_test_blockchain = true;

  const char *root_dir = neb::configuration::instance().nbre_root_dir().c_str();
  std::string nbre_path = neb::fs::join_path(root_dir, "bin/nbre");

  set_recv_nbre_version_callback(nbre_version_callback);
  set_recv_nbre_ir_list_callback(nbre_ir_list_callback);
  set_recv_nbre_ir_versions_callback(nbre_ir_versions_callback);
  set_recv_nbre_nr_handle_callback(nbre_nr_handle_callback);
  set_recv_nbre_nr_result_callback(nbre_nr_result_callback);
  set_recv_nbre_dip_reward_callback(nbre_dip_reward_callback);

  nbre_params_t params{root_dir,
                       nbre_path.c_str(),
                       neb::configuration::instance().neb_db_dir().c_str(),
                       neb::configuration::instance().nbre_db_dir().c_str(),
                       neb::configuration::instance().nbre_log_dir().c_str(),
                       "auth address here!"};
  params.m_nipc_port = 6987;

  auto ret = start_nbre_ipc(params);
  if (ret != ipc_status_succ) {
    to_quit = false;
    nbre_ipc_shutdown();
    return -1;
  }

  uint64_t height = 100;

  ipc_nbre_version(&local_mutex, height);
  ipc_nbre_ir_list(&local_mutex);
  // ipc_nbre_ir_versions(&local_mutex, "dip");

  ipc_nbre_nr_handle(&local_mutex, 6600, 6650,
                     neb::util::version(0, 1, 0).data());
  while (true) {
    ipc_nbre_nr_result(&local_mutex,
                       "00000000000019c800000000000019fa0000000100000000");
    std::this_thread::sleep_for(std::chrono::seconds(1));
  }

  // while (true) {
  // ipc_nbre_dip_reward(&local_mutex, 60000);
  // std::this_thread::sleep_for(std::chrono::seconds(1));
  //}
  std::unique_lock<std::mutex> _l(local_mutex);
  if (to_quit) {
    return 0;
  }
  local_cond_var.wait(_l);

  return 0;
}
#endif

namespace po = boost::program_options;
namespace bp = boost::process;
po::variables_map get_variables_map(int argc, char *argv[]) {
  po::options_description desc("Dummy neb");
  // clang-format off
  desc.add_options()("help", "show help message")
    ("list-dummies", "show all dummy names")
    ("run-dummy", po::value<std::string>()->default_value("default_random"), "run a dummy with name (from list-dummies, default [default_random])")
    ("block-interval", po::value<uint64_t>()->default_value(3), "block interval with seconds")
    ("without-clean-db", "run a dummy without clean previous db")
    ("clean-dummy-db", po::value<std::string>(), "clean the db file of a dummy")
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

void init_dummy_driver(dummy_driver &dd) {
  auto default_dummy =
      std::make_shared<random_dummy>("default_random", 20, 10000_nas, 0.05);
  default_dummy->enable_auth_gen_with_ratio(1);
  // default_dummy->enable_auth_gen_with_ratio(0.5);
  dd.add_dummy(default_dummy);
}

void init_and_start_nbre(const address_t &auth_admin_addr,
                         const std::string &neb_db_dir) {

  const char *root_dir = neb::configuration::instance().nbre_root_dir().c_str();
  std::string nbre_path = neb::fs::join_path(root_dir, "bin/nbre");

  // set_recv_nbre_version_callback(nbre_version_callback);
  // set_recv_nbre_ir_list_callback(nbre_ir_list_callback);
  // set_recv_nbre_ir_versions_callback(nbre_ir_versions_callback);
  // set_recv_nbre_nr_handle_callback(nbre_nr_handle_callback);
  // set_recv_nbre_nr_result_callback(nbre_nr_result_callback);
  // set_recv_nbre_dip_reward_callback(nbre_dip_reward_callback);

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
  params.m_nipc_listen = "127.0.0.1";
  params.m_nipc_port = 6987;

  auto ret = start_nbre_ipc(params);
  if (ret != ipc_status_succ) {
    to_quit = false;
    nbre_ipc_shutdown();
    exit(-1);
  }
}
int main(int argc, char *argv[]) {

  std::string root_dir = neb::configuration::instance().nbre_root_dir();
  neb::configuration::instance().neb_db_dir() =
      neb::fs::join_path(root_dir, std::string("dummy_db"));

  std::thread quiter([]() {
    char c = 'a';
    while (c != 'x')
      std::cin >> c;
    nbre_ipc_shutdown();
    exit(-1);
  });

  FLAGS_logtostderr = true;

  //::google::InitGoogleLogging(argv[0]);
  neb::glog_log_to_stderr = false;
  neb::use_test_blockchain = true;

  dummy_driver dd;
  init_dummy_driver(dd);
  po::variables_map vm = get_variables_map(argc, argv);
  if (vm.count("list_dummies")) {
    auto dummies = dd.get_all_dummy_names();
    std::for_each(dummies.begin(), dummies.end(), [](const std::string &s) {
      std::cout << "\t" << s << std::endl;
    });
    return 1;
  }


  if (vm.count("run-dummy")) {
    std::string dummy_name = vm["run-dummy"].as<std::string>();
    uint64_t interval = vm["block-interval"].as<uint64_t>();
    if (!vm.count("without-clean-db")) {
      dd.reset_dummy(dummy_name);
    }
    auto dummy = dd.get_dummy_with_name(dummy_name);
    std::string db_path = dummy->db_path();
    bc_storage_session::instance().init(db_path,
                                        neb::fs::storage_open_for_readwrite);
    std::thread thrd([&dd, dummy_name, interval]() {
      dd.run(dummy_name, interval);
      nbre_ipc_shutdown();
    });

    std::this_thread::sleep_for(std::chrono::seconds(interval));

    init_and_start_nbre(dummy->get_auth_admin_addr(), db_path);
    LOG(INFO) << "init nbre done!";
    thrd.join();
    return 0;
  }
  if (vm.count("clean-dummy-db")) {
    std::string dummy_name = vm["clean-dummy-db"].as<std::string>();
    dd.reset_dummy(dummy_name);
    return 0;
  }

  return 0;
}
