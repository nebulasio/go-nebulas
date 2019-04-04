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

#include "common/common.h"
#include "common/configuration.h"
#include "core/net_ipc/client/client_driver.h"
#include "fs/bc_storage_session.h"
#include "fs/storage_holder.h"
#include "fs/util.h"
#include "util/controller.h"
#include <boost/program_options.hpp>

namespace po = boost::program_options;

po::variables_map get_variables_map(int argc, char *argv[]) {
  po::options_description desc("NBRE (Nebulas Blockchain Runtime Environment)");
  // clang-format off
  desc.add_options()("help", "show help message")
    ("use-test-blockchain", "use test blockchain")
    ("log-to-stderr", "glog to stderr")
    ("log-dir", po::value<std::string>(), "nbre log dir")
    ("ipc-ip", po::value<std::string>(), "ipc network ip")
    ("ipc-port", po::value<std::uint16_t>(), "ipc network port");

  // clang-format on

  po::variables_map vm;
  po::store(po::parse_command_line(argc, argv, desc), vm);
  po::notify(vm);
  if (vm.count("help")) {
    std::cout << desc << "\n";
    exit(1);
  }

  if (vm.count("use-test-blockchain")) {
    neb::use_test_blockchain = true;
  } else {
    neb::use_test_blockchain = false;
  }

  if (vm.count("log-to-stderr")) {
    neb::glog_log_to_stderr = true;
  } else {
    neb::glog_log_to_stderr = false;
  }
  LOG(INFO) << "log-to-stderr: " << neb::glog_log_to_stderr;

  if (!vm.count("log-dir")) {
    std::cout << "You must specify \"log-dir\"!" << std::endl;
    exit(1);
  }
  if (!vm.count("ipc-ip")) {
    std::cout << "You must specify \"ipc-ip\"!" << std::endl;
    exit(1);
  }
  if (!vm.count("ipc-port")) {
    std::cout << "You must specify \"ipc-port\"!" << std::endl;
    exit(1);
  }

  neb::configuration::instance().nbre_log_dir() =
      vm["log-dir"].as<std::string>();
  neb::configuration::instance().nipc_listen() = vm["ipc-ip"].as<std::string>();
  neb::configuration::instance().nipc_port() = vm["ipc-port"].as<uint16_t>();

  return vm;
}

int main(int argc, char *argv[]) {
  FLAGS_logtostderr = false;
  neb::program_name = "nbre";

  get_variables_map(argc, argv);
  if (neb::glog_log_to_stderr) {
    FLAGS_log_dir = neb::configuration::instance().nbre_log_dir();
    google::InitGoogleLogging("nbre-client");
    LOG(INFO) << "log dir client " << FLAGS_log_dir;
  }
  LOG(INFO) << "nbre started!";

#ifdef NDEBUG
  neb::core::client_driver d;
  d.init();
  d.run();
  LOG(INFO) << "to quit nbre";
#else

  std::thread thrd([]() {
    neb::core::client_driver d;
    d.init();
    d.run();
    LOG(INFO) << "to quit nbre";
  });

  // neb::util::elfin ef;
  // ef.run();
  // LOG(INFO) << "ef run done";
  thrd.join();
#endif

  neb::fs::bc_storage_session::instance().release();
  neb::fs::storage_holder::instance().release();
  return 0;
}
