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
#include "benchmark/common/benchmark_instances.h"
#include <boost/asio/ip/host_name.hpp>
#include <boost/date_time/posix_time/posix_time.hpp>
#include <boost/program_options.hpp>
#include <boost/property_tree/json_parser.hpp>
#include <boost/property_tree/ptree.hpp>
#include <chrono>
#include <iostream>

namespace po = boost::program_options;

namespace neb {
benchmark_instances::~benchmark_instances() {}

void benchmark_instances::init_benchmark_instances(int argc, char *argv[]) {
  po::options_description desc("Do benchmark");
  desc.add_options()("help", "show help message")(
      "list", "show all benchmarks")("output", po::value<std::string>(),
                                     "output file of benchmark result")(
      "eval-count", po::value<int>(), "benchmark eval count [default: 10]")(
      "fixture", po::value<std::string>(), "enabled benchmark fixure, can be a "
                                           "fixture name shown in \"list\" "
                                           "option [default: all]");

  po::variables_map vm;
  po::store(po::parse_command_line(argc, argv, desc), vm);
  po::notify(vm);
  if (vm.count("help")) {
    std::cout << desc << std::endl;
    exit(1);
  }
  if (vm.count("list")) {
    show_all_benchmarks();
    exit(1);
  }

  std::string fixture_name = "all";
  if (vm.count("fixture")) {
    fixture_name = vm["fixture"].as<std::string>();
  }
  parse_all_enabled_fixtures(fixture_name);

  if (!vm.count("output")) {
    std::cout << "You must specify \"output\"!" << std::endl;
    exit(1);
  }
  m_output_fp = vm["output"].as<std::string>();

  m_eval_count = 10;
  if (vm.count("eval-count")) {
    m_eval_count = vm["eval-count"].as<int>();
  }
}
void benchmark_instances::parse_all_enabled_fixtures(
    const std::string &fixture_name) {
  std::for_each(m_all_instances.begin(), m_all_instances.end(),
                [&](const benchmark_instance_base_ptr &p) {
                  if (p->get_fixture_name() == fixture_name ||
                      fixture_name == std::string("all")) {
                    m_enabled_fixtures.insert(p->get_fixture_name());
                  }
                });
}
void benchmark_instances::show_all_benchmarks() {
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
    std::cout << "|" << it->first << std::endl;
    for (size_t i = 0; i < it->second.size(); ++i) {
      std::cout << "|----" << it->second[i] << std::endl;
    }
  }
}
int benchmark_instances::run_all_benchmarks() {
  boost::property_tree::ptree property_tree;

  std::chrono::time_point<std::chrono::system_clock> start, end;
  auto prefix_str = [](const std::string &fixture, const std::string &name) {
    return fixture + std::string(".") + name;
  };

  property_tree.put("unit", "microseconds");
  property_tree.put("time-zone", "UTC");
  property_tree.put("time",
                    boost::posix_time::to_iso_string(
                        boost::posix_time::second_clock::universal_time()));
  property_tree.put("host", boost::asio::ip::host_name());

  for (auto it = m_all_instances.begin(); it != m_all_instances.end(); ++it) {
    std::string fixture = (*it)->get_fixture_name();
    std::string name = (*it)->get_instance_name();
    if (m_enabled_fixtures.find(fixture) == m_enabled_fixtures.end())
      continue;

    boost::property_tree::ptree results;
    for (size_t i = 0; i < m_eval_count; ++i) {
      start = std::chrono::system_clock::now();
      (*it)->run();
      end = std::chrono::system_clock::now();
      auto elapsed_seconds =
          std::chrono::duration_cast<std::chrono::microseconds>(end - start)
              .count();
      boost::property_tree::ptree time_result;
      time_result.put("", elapsed_seconds);
      results.push_back(std::make_pair("", time_result));
    }
    property_tree.add_child(prefix_str(fixture, name), results);
  }

  boost::property_tree::write_json(m_output_fp, property_tree);
  return 0;
}

size_t
benchmark_instances::register_benchmark(const benchmark_instance_base_ptr &b) {
  m_all_instances.push_back(b);
  return m_all_instances.size();
}
} // end namespace neb
