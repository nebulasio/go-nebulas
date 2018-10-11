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
#include "common/ipc/shm_service.h"
#include "core/command.h"
#include "core/exception_handler.h"
#include "test/common/ipc/ipc_test.h"
#include <thread>

typedef neb::ipc::shm_service_server<128 * 1024> shm_server_t;
typedef neb::ipc::shm_service_client<128 * 1024> shm_client_t;
IPC_PRELUDE(test_service_simple) {
  std::string base_name = "test_session_simple";
  boost::interprocess::named_mutex::remove(base_name.c_str());
  shm_server_t s(base_name, 128, 128);
  LOG(INFO) << "to reset";
  s.reset();
  LOG(INFO) << "reset done";
}

IPC_SERVER(test_service_simple) {
  std::string base_name = "test_session_simple";
  neb::core::exception_handler eh;
  eh.run();
  shm_server_t s(base_name, 128, 128);
  s.run();
  s.wait_until_client_start();
  LOG(INFO) << "got client start!";

  s.wait_till_finish();
}
IPC_CLIENT(test_service_simple) {
  std::string base_name = "test_session_simple";
  neb::core::exception_handler eh;
  eh.run();

  shm_client_t c(base_name, 128, 128);
  c.run();

  std::this_thread::sleep_for(std::chrono::seconds(10));
  neb::core::command_queue::instance().send_command(
      std::make_shared<neb::core::exit_command>());

  c.wait_till_finish();
  eh.kill();
}
