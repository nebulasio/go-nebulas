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
#include "benchmark/benchmark_instances.h"
#include <boost/asio/ip/host_name.hpp>
#include <boost/date_time/posix_time/posix_time.hpp>
#include <boost/program_options.hpp>
#include <boost/property_tree/json_parser.hpp>
#include <boost/property_tree/ptree.hpp>
#include <chrono>
#include <iostream>

#include <sys/sysinfo.h>
#include <sys/types.h>

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

  struct sysinfo mem_start, mem_end;
  auto f_mem = [](boost::property_tree::ptree &pt,
                  const struct sysinfo &mem_start,
                  const struct sysinfo &mem_end) {
    auto mem_unit = mem_start.mem_unit;
    pt.put("totalram", (mem_end.totalram - mem_start.totalram) * mem_unit);
    pt.put("freeram", (mem_end.freeram - mem_start.freeram) * mem_unit);
    pt.put("sharedram", (mem_end.sharedram - mem_start.sharedram) * mem_unit);
    pt.put("bufferram", (mem_end.bufferram - mem_start.bufferram) * mem_unit);
    pt.put("totalswap", (mem_end.totalswap - mem_start.totalswap) * mem_unit);
    pt.put("freeswap", (mem_end.freeswap - mem_start.freeswap) * mem_unit);
    pt.put("totalhigh", (mem_end.totalhigh - mem_start.totalhigh) * mem_unit);
    pt.put("freeshigh", (mem_end.freehigh - mem_start.freehigh) * mem_unit);
  };

  struct proc_stat_t {
    uint64_t user;
    uint64_t nice;
    uint64_t system;
    uint64_t idle;
    uint64_t iowait;
    uint64_t irq;
    uint64_t softrq;
  };
  struct proc_stat_t cpu_start, cpu_end;
  auto procstat = [](struct proc_stat_t &cpu_info) {
    std::ifstream file("/proc/stat");
    std::string ignore;
    file >> ignore >> cpu_info.user >> cpu_info.nice >> cpu_info.system >>
        cpu_info.idle >> cpu_info.iowait >> cpu_info.irq >> cpu_info.softrq;
  };

  auto f_cpu = [](boost::property_tree::ptree &pt,
                  const struct proc_stat_t &cpu_start,
                  const struct proc_stat_t &cpu_end) {
    auto delta_idle = cpu_end.idle - cpu_start.idle;
    auto delta_iowait = cpu_end.iowait - cpu_start.iowait;
    auto delta_total =
        cpu_end.user + cpu_end.nice + cpu_end.system + cpu_end.idle +
        cpu_end.iowait + cpu_end.irq + cpu_end.softrq -
        (cpu_start.user + cpu_start.nice + cpu_start.system + cpu_start.idle +
         cpu_start.iowait + cpu_start.irq + cpu_start.softrq);
    decltype(delta_idle) usage = 0;
    if (delta_total > 0) {
      usage = 1 - (delta_idle + delta_iowait) / delta_total;
    }
    pt.put("cpu_usage", usage);
  };

  for (auto it = m_all_instances.begin(); it != m_all_instances.end(); ++it) {
    std::string fixture = (*it)->get_fixture_name();
    std::string name = (*it)->get_instance_name();
    if (m_enabled_fixtures.find(fixture) == m_enabled_fixtures.end())
      continue;

    boost::property_tree::ptree results;
    for (size_t i = 0; i < m_eval_count; ++i) {
      procstat(cpu_start);
      sysinfo(&mem_start);
      start = std::chrono::system_clock::now();
      (*it)->run();
      end = std::chrono::system_clock::now();
      sysinfo(&mem_end);
      procstat(cpu_end);
      auto elapsed_seconds =
          std::chrono::duration_cast<std::chrono::nanoseconds>(end - start)
              .count();
      boost::property_tree::ptree time_result;
      time_result.put("nano_seconds", elapsed_seconds);
      f_cpu(time_result, cpu_start, cpu_end);
      f_mem(time_result, mem_start, mem_end);
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
