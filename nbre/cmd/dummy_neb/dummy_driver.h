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
#include "cmd/dummy_neb/dummies/dummies.h"
#include "cmd/dummy_neb/dummy_common.h"
#include "cmd/dummy_neb/generator/generators.h"
#include "util/timer_loop.h"

class dummy_driver {
public:
  dummy_driver();
  ~dummy_driver();
  void add_dummy(const std::shared_ptr<dummy_base> &dummy);
  std::vector<std::string> get_all_dummy_names() const;

  std::shared_ptr<dummy_base>
  get_dummy_with_name(const std::string &dummy_name) const;

  void run(const std::string &dummy_name, uint64_t block_interval_seconds);
  void shutdown();

  void check_all_dummy_tasks(const std::string &dummy_name);
  void check_dummy_task(const std::string &dummy_name,
                        const std::string &task_name);

  void reset_dummy(const std::string &dummy_name);

protected:
  boost::asio::io_service m_io_service;
  std::unordered_map<std::string, std::shared_ptr<dummy_base>> m_all_dummies;
  std::unique_ptr<neb::util::timer_loop> m_block_gen_timer;
  std::unique_ptr<neb::util::timer_loop> m_checker_gen_timer;
  uint64_t m_block_interval_seconds;
};
