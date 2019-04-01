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
#include "common/common.h"
#include "util/singleton.h"
#include <unordered_set>

namespace neb {
class benchmark_instance_base {
public:
  virtual std::string get_fixture_name() = 0;
  virtual std::string get_instance_name() = 0;
  virtual void run() const = 0;
}; // end class benchmark_instance_base

typedef std::shared_ptr<benchmark_instance_base> benchmark_instance_base_ptr;

class benchmark_instances : public util::singleton<benchmark_instances> {
public:
  void init_benchmark_instances(int argc, char *argv[]);

  virtual ~benchmark_instances();

  int run_all_benchmarks();

  size_t register_benchmark(const benchmark_instance_base_ptr &b);

protected:
  void show_all_benchmarks();
  void parse_all_enabled_fixtures(const std::string &fixture_name);

protected:
  std::vector<benchmark_instance_base_ptr> m_all_instances;
  size_t m_eval_count;
  std::string m_output_fp;
  std::unordered_set<std::string> m_enabled_fixtures;
}; // end class benchmark_instances
}

#define GEN_NAME_VAR(name) _##name##_nouse
#define BENCHMARK(fixture, name)                                               \
  class name : public neb::benchmark_instance_base {                           \
  public:                                                                      \
    virtual std::string get_fixture_name() { return #fixture; }                \
    virtual std::string get_instance_name() { return #name; }                  \
    virtual void run() const;                                                  \
  };                                                                           \
  static int GEN_NAME_VAR(name) =                                              \
      neb::benchmark_instances::instance().register_benchmark(                 \
          std::static_pointer_cast<neb::benchmark_instance_base>(              \
              std::make_shared<name>()));                                      \
  void name::run() const

