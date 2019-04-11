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
#include "fs/util.h"
#include "test/common/ipc/ipc_test.h"

std::string base_name = neb::fs::get_user_name() + "test_session_simple";
static std::string bk_name = base_name + ".bookkeeper";
std::string sema_name = bk_name + ".test.sema";
std::string quit_sema_name = bk_name + ".test.quit.sema";

IPC_PRELUDE(test_session_simple) {

  neb::ipc::internal::shm_session_util ss(base_name);
  ss.reset();
  neb::ipc::internal::shm_bookkeeper sb(bk_name);
  sb.reset();

  boost::interprocess::named_semaphore::remove(sema_name.c_str());
  boost::interprocess::named_semaphore::remove(quit_sema_name.c_str());
  boost::interprocess::shared_memory_object::remove(bk_name.c_str());
}

IPC_SERVER(test_session_simple) {

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

IPC_PRELUDE(test_session_full) {

  neb::ipc::internal::shm_session_util ss(base_name);
  ss.reset();
  neb::ipc::internal::shm_bookkeeper sb(bk_name);
  sb.reset();
}

IPC_SERVER(test_session_full) {

  std::shared_ptr<neb::ipc::internal::shm_session_server> ss =
      std::make_shared<neb::ipc::internal::shm_session_server>(base_name);
  ss->start_session();
  ss->wait_until_client_start();
  std::this_thread::sleep_for(std::chrono::seconds(3));
  bool ret = ss->is_client_alive();
  IPC_EXPECT(ret);
  LOG(INFO) << "server done";
}

IPC_CLIENT(test_session_full) {
  std::shared_ptr<neb::ipc::internal::shm_session_client> ss =
      std::make_shared<neb::ipc::internal::shm_session_client>(base_name);
  LOG(INFO) << "sleep to wait server start";
  std::this_thread::sleep_for(std::chrono::seconds(3));
  LOG(INFO) << "to start session";
  ss->start_session();
  std::this_thread::sleep_for(std::chrono::seconds(3));
  bool ret = ss->is_server_alive();
  IPC_EXPECT(ret);

  LOG(INFO) << "sleep to wait server got ";
  std::this_thread::sleep_for(std::chrono::seconds(3));
  neb::core::command_queue::instance().send_command(
      std::make_shared<neb::core::exit_command>());
  LOG(INFO) << "client done";
}

#if 0
typedef struct session_name_t {
  std::string base_name;
  std::string bk_name;
  std::string sema_name;
  std::string quit_sema_name;
} session_name;

session_name get_session_name(const std::string &num) {
  session_name sn;
  sn.base_name = neb::fs::get_user_name() + "test_multi_session_" + num;
  sn.bk_name = sn.base_name + ".bookkeeper";
  sn.sema_name = sn.bk_name + ".test.sema";
  sn.quit_sema_name = sn.bk_name + ".test.quit.sema";
  return sn;
}

void do_ipc_prelude(const session_name &sn) {
  neb::ipc::internal::shm_session_util ss(sn.base_name);
  ss.reset();
  neb::ipc::internal::shm_bookkeeper sb(sn.bk_name);
  sb.reset();

  boost::interprocess::named_semaphore::remove(sn.sema_name.c_str());
  boost::interprocess::named_semaphore::remove(sn.quit_sema_name.c_str());
  boost::interprocess::shared_memory_object::remove(sn.bk_name.c_str());
}

void build_server_session(const session_name &sn) {
  std::shared_ptr<neb::ipc::internal::shm_session_server> ss =
      std::make_shared<neb::ipc::internal::shm_session_server>(sn.base_name);
  ss->start_session();

  LOG(INFO) << "server session started " << sn.base_name;

  neb::ipc::internal::shm_bookkeeper sb(sn.bk_name);
  auto sema = sb.acquire_named_semaphore(sn.sema_name);
  auto quit_sema = sb.acquire_named_semaphore(sn.quit_sema_name);
  sema->post();
  LOG(INFO) << "server to wait" << sn.base_name;
  quit_sema->wait();

  ss.reset();
  sb.release_named_semaphore(sn.sema_name);
  sb.release_named_semaphore(sn.quit_sema_name);
}

void build_client_session(const session_name &sn) {
  neb::ipc::internal::shm_bookkeeper sb(sn.bk_name);
  auto sema = sb.acquire_named_semaphore(sn.sema_name);
  auto quit_sema = sb.acquire_named_semaphore(sn.quit_sema_name);
  LOG(INFO) << "client to wait";
  sema->wait();
  LOG(INFO) << "client session to start";
  std::shared_ptr<neb::ipc::internal::shm_session_client> ss =
      std::make_shared<neb::ipc::internal::shm_session_client>(sn.base_name);
  ss->start_session();

  LOG(INFO) << "client session started";
  quit_sema->post();

  std::this_thread::sleep_for(std::chrono::seconds(3));
  neb::core::command_queue::instance().send_command(
      std::make_shared<neb::core::exit_command>());
  ss.reset();
  sb.release_named_semaphore(sn.sema_name);
  sb.release_named_semaphore(sn.quit_sema_name);
  LOG(INFO) << "client done";
}

IPC_PRELUDE(test_multi_session) {
  session_name sn = get_session_name("1");
  do_ipc_prelude(sn);
}

IPC_SERVER(test_multi_session) {
  session_name sn1 = get_session_name("1");
  session_name sn2 = get_session_name("2");
  LOG(INFO) << "server start ";

  build_server_session(sn1);
  build_server_session(sn2);

  LOG(INFO) << "server done";
}

IPC_CLIENT(test_multi_session) {
  session_name sn = get_session_name("1");
  build_client_session(sn);
}

#endif
