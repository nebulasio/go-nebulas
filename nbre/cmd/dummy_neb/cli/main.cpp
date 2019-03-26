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
#include "cmd/dummy_neb/cli/pkg.h"
#include "common/configuration.h"
#include "fs/util.h"
#include "util/controller.h"
#include <boost/process.hpp>
#include <boost/program_options.hpp>
#include <ff/network.h>

namespace po = boost::program_options;
namespace bp = boost::process;

po::variables_map get_variables_map(int argc, char *argv[]) {
  po::options_description desc("Dummy CLI tool");
  // clang-format off
  desc.add_options()("help", "show help message")
    ("brief", "show current dummy brief")
    ("payload", po::value<std::string>(), "payload file path")
    ("submit", po::value<std::string>(), "auth, nr, dip")
    ("query", po::value<std::string>(), "nr, nr-result, dip-reward")
    ("start-block", po::value<uint64_t>(), "start block height")
    ("end-block", po::value<uint64_t>(), "end block height")
    ("version", po::value<std::string>()->default_value("0.0.1"), "x.x.x")
    ("handle", po::value<std::string>(), "request handle")
    ("height", po::value<uint64_t>()->default_value(0), "request height")
    ("rpc-listen", po::value<std::string>()->default_value("127.0.0.1"), "nipc listen")
    ("rpc-port", po::value<uint16_t>()->default_value(0x1958), "nipc port")
    ("kill-nbre", "kill nbre immediatelly");
    /*
    ("run-dummy", po::value<std::string>()->default_value("default_random"), "run a dummy with name (from list-dummies, default [default_random])")
    ("block-interval", po::value<uint64_t>()->default_value(3), "block interval with seconds")
    ("without-clean-db", "run a dummy without clean previous db")
    ("clean-dummy-db", po::value<std::string>(), "clean the db file of a dummy")
    ("enable-nbre-killer", po::value<uint64_t>()->default_value(24*60), "kill nbre periodically in miniutes");
    */
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

class cli_executor {
public:
  cli_executor(const std::string &rpc_listen, uint16_t rpc_port)
      : m_rpc_listen(rpc_listen), m_rpc_port(rpc_port) {}

  void send_brief_req() {
    std::shared_ptr<cli_brief_req_t> req = std::make_shared<cli_brief_req_t>();
    m_package = req;
    start_and_join();
  }
  void send_submit_ir(const std::string &type, const std::string &fp) {
    std::ifstream ifs;
    ifs.open(fp);
    if (!ifs.is_open()) {
      std::cout << "cannot open file: " << fp << std::endl;
    }
    ifs.seekg(0, ifs.end);
    std::ifstream::pos_type size = ifs.tellg();
    neb::bytes buf(size);

    ifs.seekg(0, ifs.beg);
    ifs.read((char *)buf.value(), buf.size());
    ifs.close();

    std::shared_ptr<cli_submit_ir_t> req = std::make_shared<cli_submit_ir_t>();
    req->set<p_type>(type);
    req->set<p_payload>(neb::byte_to_string(buf));
    m_package = req;
    start_and_join();
  }

  void send_nr_req(uint64_t start_block, uint64_t end_block, uint64_t version) {
    std::shared_ptr<nbre_nr_handle_req> req =
        std::make_shared<nbre_nr_handle_req>();
    req->set<p_holder>(reinterpret_cast<uint64_t>(this));
    req->set<p_start_block>(start_block);
    req->set<p_end_block>(end_block);
    req->set<p_nr_version>(version);
    m_package = req;
    start_and_join();
  }
  void send_nr_result_req(const std::string &handle) {
    std::shared_ptr<nbre_nr_result_by_handle_req> req =
        std::make_shared<nbre_nr_result_by_handle_req>();
    req->set<p_holder>(reinterpret_cast<uint64_t>(this));
    req->set<p_nr_handle>(handle);
    m_package = req;
    start_and_join();
  }

  void send_nr_result_by_height_req(uint64_t height) {
    std::shared_ptr<nbre_nr_result_by_height_req> req =
        std::make_shared<nbre_nr_result_by_height_req>();
    req->set<p_holder>(reinterpret_cast<uint64_t>(this));
    req->set<p_height>(height);
    m_package = req;
    start_and_join();
  }

  void send_nr_sum_req(uint64_t height) {
    std::shared_ptr<nbre_nr_sum_req> req = std::make_shared<nbre_nr_sum_req>();
    req->set<p_holder>(reinterpret_cast<uint64_t>(this));
    req->set<p_height>(height);
    m_package = req;
    start_and_join();
  }

  void send_dip_reward_req(uint64_t height, uint64_t version) {
    std::shared_ptr<nbre_dip_reward_req> req =
        std::make_shared<nbre_dip_reward_req>();
    req->set<p_holder>(reinterpret_cast<uint64_t>(this));
    req->set<p_height>(height);
    req->set<p_version>(version);
    m_package = req;
    start_and_join();
  }

protected:
  void start_and_join() {

    ff::net::net_nervure nn;

    ff::net::typed_pkg_hub hub;
    ff::net::tcp_connection_base_ptr conn;
    nn.get_event_handler()->listen<::ff::net::event::tcp_get_connection>(
        [&, this](::ff::net::tcp_connection_base *conn) {
          conn->send(m_package);
        });

    hub.to_recv_pkg<cli_brief_ack_t>([&](std::shared_ptr<cli_brief_ack_t> ack) {
      std::cout << "\t height: " << ack->get<p_height>() << std::endl;
      std::cout << "\t account num: " << ack->get<p_account_num>() << std::endl;
      std::cout << "\t " << ack->get<p_checker_status>() << std::endl;
      ;
      conn->close();
      exit(-1);
    });

    hub.to_recv_pkg<cli_submit_ack_t>(
        [&](std::shared_ptr<cli_submit_ack_t> ack) {
          std::cout << "\t result: " << ack->get<p_result>() << std::endl;
          conn->close();
          exit(-1);
        });
    hub.to_recv_pkg<nbre_nr_result_by_handle_ack>(
        [&](std::shared_ptr<nbre_nr_result_by_handle_ack> ack) {
          std::cout << "\t " << ack->get<p_nr_result>() << std::endl;
          conn->close();
          exit(-1);
        });
    hub.to_recv_pkg<nbre_nr_handle_ack>(
        [&](std::shared_ptr<nbre_nr_handle_ack> ack) {
          std::cout << "\t" << ack->get<p_nr_handle>() << std::endl;
          conn->close();
          exit(-1);
        });

    hub.to_recv_pkg<nbre_nr_result_by_height_ack>(
        [&](std::shared_ptr<nbre_nr_result_by_height_ack> ack) {
          std::cout << "\t " << ack->get<p_nr_result>() << std::endl;
          conn->close();
          exit(-1);
        });
    hub.to_recv_pkg<nbre_nr_sum_ack>([&](std::shared_ptr<nbre_nr_sum_ack> ack) {
      std::cout << "\t " << ack->get<p_nr_sum>() << std::endl;
      conn->close();
      exit(-1);
    });
    hub.to_recv_pkg<nbre_dip_reward_ack>(
        [&](std::shared_ptr<nbre_dip_reward_ack> ack) {
          std::cout << "\t" << ack->get<p_dip_reward>() << std::endl;
          conn->close();
          exit(-1);
        });
    nn.add_pkg_hub(hub);
    conn = nn.add_tcp_client(m_rpc_listen, m_rpc_port);

    nn.run();
  }

protected:
  std::shared_ptr<ff::net::package> m_package;
  std::string m_rpc_listen;
  uint16_t m_rpc_port;
};

int main(int argc, char *argv[]) {
  po::variables_map vm = get_variables_map(argc, argv);
  std::string rpc_listen = vm["rpc-listen"].as<std::string>();
  uint16_t rpc_port = vm["rpc-port"].as<uint16_t>();

  if (vm.count("submit")) {
    std::string type = vm["submit"].as<std::string>();
    if (type != "nr" && type != "auth" && type != "dip") {
      std::cout << "invalid type " << type << std::endl;
      exit(-1);
    }
    if (!vm.count("payload")) {
      std::cout << "no payload " << std::endl;
      exit(-1);
    }
    std::string fp = vm["payload"].as<std::string>();

    cli_executor ce(rpc_listen, rpc_port);
    ce.send_submit_ir(type, fp);
  } else if (vm.count("brief")) {
    cli_executor ce(rpc_listen, rpc_port);
    ce.send_brief_req();
  } else if (vm.count("query")) {
    std::string type = vm["query"].as<std::string>();
    if (type != "nr" && type != "nr-result" && type != "nr-sum" &&
        type != "dip-reward") {
      std::cout << "invalid type " << type << std::endl;
      exit(-1);
    }
    if (type == "nr") {
      if (!vm.count("start-block") || !vm.count("end-block") ||
          !vm.count("version")) {
        std::cout << "no start, end block, or version" << std::endl;
        exit(-1);
      }
      auto start_block = vm["start-block"].as<uint64_t>();
      auto end_block = vm["end-block"].as<uint64_t>();
      auto version_str = vm["version"].as<std::string>();
      neb::version v;
      v.from_string(version_str);
      cli_executor ce(rpc_listen, rpc_port);
      ce.send_nr_req(start_block, end_block, v.data());
    }
    if (type == "nr-result") {
      if (vm.count("handle")) {
        auto handle = vm["handle"].as<std::string>();
        cli_executor ce(rpc_listen, rpc_port);
        ce.send_nr_result_req(handle);
      } else if (vm.count("height")) {
        auto height = vm["height"].as<uint64_t>();
        LOG(INFO) << "cli query nr-result by height " << height;
        cli_executor ce(rpc_listen, rpc_port);
        ce.send_nr_result_by_height_req(height);
      }
    }
    if (type == "nr-sum") {
      if (vm.count("height")) {
        auto height = vm["height"].as<uint64_t>();
        LOG(INFO) << "cli query nr-sum by height " << height;
        cli_executor ce(rpc_listen, rpc_port);
        ce.send_nr_sum_req(height);
      }
    }
    if (type == "dip-reward") {
      auto height = vm["height"].as<uint64_t>();
      auto version_str = vm["version"].as<std::string>();
      neb::version v;
      v.from_string(version_str);
      cli_executor ce(rpc_listen, rpc_port);
      ce.send_dip_reward_req(height, v.data());
    }
  } else if (vm.count("kill-nbre")) {
    neb::util::magic_wand mw;
    mw.kill_nbre();
  }
  return 0;
}
