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
#include "common/ipc/shm_bookkeeper.h"
#include "common/ipc/shm_session.h"
#include "core/command.h"
#include "test/common/ipc/ipc_test.h"

IPC_PRELUDE(test_session_simple) {
  std::string base_name = "test_session_simple";
  std::string bk_name = base_name + ".bookkeeper";
  std::string sema_name = bk_name + ".test.sema";
  std::string quit_sema_name = bk_name + ".test.quit.sema";

  neb::ipc::internal::shm_session_util ss(base_name);
  ss.reset();
  neb::ipc::internal::shm_bookkeeper sb(bk_name);
  sb.reset();

  boost::interprocess::named_semaphore::remove(sema_name.c_str());
  boost::interprocess::named_semaphore::remove(quit_sema_name.c_str());
  boost::interprocess::shared_memory_object::remove(bk_name.c_str());
}

IPC_SERVER(test_session_simple) {
  std::string base_name = "test_session_simple";
  std::string bk_name = base_name + ".bookkeeper";
  std::string sema_name = bk_name + ".test.sema";
  std::string quit_sema_name = bk_name + ".test.quit.sema";

  LOG(INFO) << "server start ";
  std::shared_ptr<neb::ipc::internal::shm_session_server> ss =
      std::make_shared<neb::ipc::internal::shm_session_server>(base_name);
  ss->start_session();
  LOG(INFO) << "server session started ";
  neb::ipc::internal::shm_bookkeeper sb(bk_name);
  auto sema = sb.acquire_named_semaphore(sema_name);
  auto quit_sema = sb.acquire_named_semaphore(quit_sema_name);
  sema->post();
  LOG(INFO) << "server to wait";
  quit_sema->wait();

  ss.reset();
  sb.release_named_semaphore(sema_name);
  sb.release_named_semaphore(quit_sema_name);
  LOG(INFO) << "server done";
}

IPC_CLIENT(test_session_simple) {
  std::string base_name = "test_session_simple";
  std::string bk_name = base_name + ".bookkeeper";
  std::string sema_name = bk_name + ".test.sema";
  std::string quit_sema_name = bk_name + ".test.quit.sema";

  neb::ipc::internal::shm_bookkeeper sb(bk_name);
  auto sema = sb.acquire_named_semaphore(sema_name);
  auto quit_sema = sb.acquire_named_semaphore(quit_sema_name);
  LOG(INFO) << "client to wait";
  sema->wait();
  LOG(INFO) << "client session to start";
  std::shared_ptr<neb::ipc::internal::shm_session_client> ss =
      std::make_shared<neb::ipc::internal::shm_session_client>(base_name);
  ss->start_session();

  LOG(INFO) << "client session started";
  quit_sema->post();

  std::this_thread::sleep_for(std::chrono::seconds(3));
  neb::core::command_queue::instance().send_command(
      std::make_shared<neb::core::exit_command>());
  ss.reset();
  sb.release_named_semaphore(sema_name);
  sb.release_named_semaphore(quit_sema_name);
  LOG(INFO) << "client done";
}
