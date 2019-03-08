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
    ("query", po::value<std::string>(), "nr")
    ("start-block", po::value<uint64_t>(), "start block height")
    ("end-block", po::value<uint64_t>(), "end block height")
    ("version", po::value<std::string>(), "x.x.x")
    ("handle", po::value<std::string>(), "request handle")
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
    neb::util::bytes buf(size);

    ifs.seekg(0, ifs.beg);
    ifs.read((char *)buf.value(), buf.size());
    ifs.close();

    std::shared_ptr<cli_submit_ir_t> req = std::make_shared<cli_submit_ir_t>();
    req->set<p_type>(type);
    req->set<p_payload>(neb::util::byte_to_string(buf));
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
    std::shared_ptr<nbre_nr_result_req> req =
        std::make_shared<nbre_nr_result_req>();
    req->set<p_holder>(reinterpret_cast<uint64_t>(this));
    req->set<p_nr_handle>(handle);
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
      conn->close();
      exit(-1);
    });

    hub.to_recv_pkg<cli_submit_ack_t>(
        [&](std::shared_ptr<cli_submit_ack_t> ack) {
          std::cout << "\t result: " << ack->get<p_result>() << std::endl;
          conn->close();
          exit(-1);
        });
    hub.to_recv_pkg<nbre_nr_result_ack>(
        [&](std::shared_ptr<nbre_nr_result_ack> ack) {
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
    nn.add_pkg_hub(hub);
    conn = nn.add_tcp_client("127.0.0.1", 0x1958);

    nn.run();
  }

protected:
  std::shared_ptr<ff::net::package> m_package;
};

int main(int argc, char *argv[]) {
  po::variables_map vm = get_variables_map(argc, argv);
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

    cli_executor ce;
    ce.send_submit_ir(type, fp);
  } else if (vm.count("brief")) {
    cli_executor ce;
    ce.send_brief_req();
  } else if (vm.count("query")) {
    std::string type = vm["query"].as<std::string>();
    if (type != "nr" && type != "auth" && type != "dip" &&
        type != "nr-result") {
      std::cout << "invalid type " << type << std::endl;
      exit(-1);
    }
    if (type == "nr") {
      if (!vm.count("start-block") && !vm.count("end-block") &&
          !vm.count("version")) {
        std::cout << "no start, end block, or version" << std::endl;
        exit(-1);
      }
      auto start_block = vm["start-block"].as<uint64_t>();
      auto end_block = vm["end-block"].as<uint64_t>();
      auto version_str = vm["version"].as<std::string>();
      neb::util::version v;
      v.from_string(version_str);
      cli_executor ce;
      ce.send_nr_req(start_block, end_block, v.data());
    }
    if (type == "nr-result") {
      if (!vm.count("handle")) {
        std::cout << "no handle" << std::endl;
        exit(-1);
      }
      auto handle = vm["handle"].as<std::string>();
      cli_executor ce;
      ce.send_nr_result_req(handle);
    }
  } else if (vm.count("kill-nbre")) {
    neb::util::magic_wand mw;
    mw.kill_nbre();
  }
  return 0;
}
