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
#include "test/common/ipc/ipc_instance.h"
#include "common/configuration.h"
#include "fs/util.h"
#include <boost/asio/ip/host_name.hpp>
#include <boost/date_time/posix_time/posix_time.hpp>
#include <boost/process/child.hpp>
#include <boost/program_options.hpp>
#include <chrono>
#include <iostream>

namespace po = boost::program_options;

namespace neb {
ipc_instances::~ipc_instances() {}

void ipc_instances::init_ipc_instances(int argc, char *argv[]) {
  po::options_description desc("Do IPC tests");
  desc.add_options()("help", "show help message")(
      "list", "show all IPC instances")("fixture", po::value<std::string>(),
                                        "enabled fixure, can be a "
                                        "fixture name shown in \"list\" "
                                        "option [default: all]")(
      "instance", po::value<std::string>(),
      "server | client")("debug", "debug mode without catching exceptions.")(
      "enable-log", "enable glog to stderr");

  po::variables_map vm;
  po::store(po::parse_command_line(argc, argv, desc), vm);
  po::notify(vm);
  if (vm.count("help")) {
    std::cout << desc << std::endl;
    exit(1);
  }
  if (vm.count("list")) {
    show_all_instances();
    exit(1);
  }

  std::string fixture_name = "all";
  if (vm.count("fixture")) {
    fixture_name = vm["fixture"].as<std::string>();
  }
  m_enabled_fixture_name = fixture_name;
  if (fixture_name != std::string("all") && !vm.count("instance")) {
    std::cout << "miss *instance*" << std::endl;
    std::cout << desc << std::endl;
    exit(1);
  }
  if (vm.count("debug")) {
    m_debug_mode = true;
  } else {
    m_debug_mode = false;
  }
  if (vm.count("instance")) {
    m_enabled_instance_name = vm["instance"].as<std::string>();
  }
  if (vm.count("enable-log")) {
    FLAGS_logtostderr = true;
    ::google::SetStderrLogging(google::GLOG_INFO);
    ::google::SetLogDestination(
        google::GLOG_INFO, configuration::instance().nbre_root_dir().c_str());
  } else {
    FLAGS_logtostderr = false;
    ::google::SetLogDestination(
        google::GLOG_ERROR, configuration::instance().nbre_root_dir().c_str());
  }
  ::google::InitGoogleLogging(argv[0]);
  LOG(INFO) << argv[0] << " started! ";
  m_exe_name =
      neb::fs::join_path(neb::fs::cur_full_path(), std::string(argv[0]));
}

void ipc_instances::show_all_instances() {
  std::map<std::string, std::vector<std::string>> all;
  for (auto it = m_all_instances.begin(); it != m_all_instances.end(); ++it) {
    std::string fixture = (*it)->get_fixture_name();
    std::string name = (*it)->get_instance_name();

    if (all.find(fixture) == all.end()) {
      all.insert(std::make_pair(fixture, std::vector<std::string>()));
    }
    all[fixture].push_back(name);
  }

  for (decltype(all)::iterator it = all.begin(); it != all.end(); ++it) {
    std::cout << "|-" << it->first << std::endl;
    for (size_t i = 0; i < it->second.size(); ++i) {
      std::cout << "|----" << it->second[i] << std::endl;
    }
  }
}

int ipc_instances::run_all_ipc_instances() {
  std::string fp = m_exe_name;

  LOG(INFO) << "file path: " << fp;
  std::unordered_set<std::string> done_fixtures;
  std::unordered_set<std::string> has_prelude_fixtures;

  if (m_enabled_fixture_name == std::string("all")) {
    int exit_code = 0;
    for (auto it = m_all_instances.begin(); it != m_all_instances.end(); ++it) {
      std::string fixture = (*it)->get_fixture_name();
      std::string instance = (*it)->get_instance_name();
      if (instance == std::string("prelude")) {
        has_prelude_fixtures.insert(fixture);
      }
    }

    for (auto it = m_all_instances.begin(); it != m_all_instances.end(); ++it) {
      std::string fixture = (*it)->get_fixture_name();
      if (done_fixtures.find(fixture) != done_fixtures.end())
        continue;
      done_fixtures.insert(fixture);

      if (has_prelude_fixtures.find(fixture) != has_prelude_fixtures.end()) {
        boost::process::child prelude(fp, "--fixture", fixture, "--instance",
                                      "prelude");
        prelude.wait();
      }

      boost::process::child server(fp, "--fixture", fixture, "--instance",
                                   "server");
      boost::process::child client(fp, "--fixture", fixture, "--instance",
                                   "client");

      if (!server.valid()) {
        std::cout << "invalid server process for " << fixture << std::endl;
      }
      if (!client.valid()) {
        std::cout << "invalid client process for " << fixture << std::endl;
      }

      int tmp_exit_code = 0;
      if (server.valid()) {
        server.join();
        tmp_exit_code += server.exit_code();
        if (server.exit_code() == 0) {
          std::cerr << ::neb::tcolor::green << now() << " Success: " << fixture
                    << "."
                    << "server" << ::neb::tcolor::reset << std::endl;
        } else {
          std::cerr << ::neb::tcolor::red << now() << " Fail: " << fixture
                    << "."
                    << "server" << ::neb::tcolor::reset << std::endl;
        }
      }
      if (client.valid()) {
        client.join();
        tmp_exit_code += server.exit_code();
        if (client.exit_code() == 0) {
          std::cerr << ::neb::tcolor::green << now() << " Success: " << fixture
                    << "."
                    << "client" << ::neb::tcolor::reset << std::endl;
        } else {
          std::cerr << ::neb::tcolor::red << now() << " Fail: " << fixture
                    << "."
                    << "client" << ::neb::tcolor::reset << std::endl;
        }
      }
      exit_code += tmp_exit_code;
    }
    return exit_code;
  } else {
    return run_one_ipc_instance(m_enabled_fixture_name,
                                m_enabled_instance_name);
  }
}

int ipc_instances::run_one_ipc_instance(const std::string &fixture,
                                        const std::string &instance) {

  bool found = false;
  for (auto it = m_all_instances.begin(); it != m_all_instances.end(); ++it) {
    std::string cur_fixture = (*it)->get_fixture_name();
    std::string cur_instance = (*it)->get_instance_name();
    if (cur_fixture == fixture && cur_instance == instance) {
      found = true;
      std::cerr << ::neb::tcolor::reset << now() << " Start " << fixture << "."
                << instance << ::neb::tcolor::reset << std::endl;
      if (m_debug_mode) {
        (*it)->run();
      } else {
        try {
          (*it)->run();
        } catch (const std::exception &e) {
          std::cerr << ::neb::tcolor::red << now() << " Exception: " << fixture
                    << "." << instance << " " << e.what()
                    << ::neb::tcolor::reset << std::endl;
          return -1;
        }
      }
    }
  }
  if (!found) {
    std::cout << "cannot find " << instance << " for " << fixture << std::endl;
    return -1;
  }
  return 0;
}
size_t ipc_instances::register_ipc_instance(const ipc_instance_base_ptr &b) {
  m_all_instances.push_back(b);
  return m_all_instances.size();
}
} // end namespace neb
