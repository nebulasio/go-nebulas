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
#include "fs/util.h"
#include "test/common/ipc/ipc_test.h"
#include <chrono>
#include <exception>
#include <thread>

std::string user_name = neb::fs::get_user_name();

auto t1_str = user_name + "t1";
auto t2_str = user_name + "t2";
auto t3_str = user_name + "t3";
auto bk_str = user_name + "test_bookkeeper_simple";
const char *t1_name = t1_str.c_str();
const char *t2_name = t2_str.c_str();
const char *t3_name = t3_str.c_str();
static const char *bk_name = bk_str.c_str();

IPC_PRELUDE(test_bookkeeper_simple) {
  boost::interprocess::named_mutex::remove(t1_name);
  boost::interprocess::named_semaphore::remove(t2_name);
  boost::interprocess::named_semaphore::remove(t3_name);
  boost::interprocess::shared_memory_object::remove(bk_name);
}

IPC_SERVER(test_bookkeeper_simple) {

  neb::ipc::internal::shm_bookkeeper sb(bk_name);
  auto mutex = sb.acquire_named_mutex(t1_name);
  bool v = mutex->try_lock();
  IPC_EXPECT(v);
  auto sema = sb.acquire_named_semaphore(t2_name);
  auto sema2 = sb.acquire_named_semaphore(t3_name);
  sema->post();
  sema2->wait();
  sb.release_named_mutex(t1_name);
  sb.release_named_semaphore(t2_name);
  sb.release_named_semaphore(t3_name);
}

IPC_CLIENT(test_bookkeeper_simple) {
  neb::ipc::internal::shm_bookkeeper sb(bk_name);
  auto sema = sb.acquire_named_semaphore(t2_name);
  auto mutex = sb.acquire_named_mutex(t1_name);
  sema->wait();
  bool v = mutex->try_lock();
  IPC_EXPECT(v == false);
  auto sema2 = sb.acquire_named_semaphore(t3_name);
  sema2->post();
  sb.release_named_mutex(t1_name);
  sb.release_named_semaphore(t2_name);
  sb.release_named_semaphore(t3_name);
}
auto tc1_str = (user_name + "tc1");
auto tc2_str = (user_name + "tc2");
auto tc3_str = (user_name + "tc3");
auto c1_str = (user_name + "c1");
auto bkc_str = (user_name + "test_bookkeeper_simple_cond");

const char *tc1_name = tc1_str.c_str();
const char *tc2_name = tc2_str.c_str();
const char *tc3_name = tc3_str.c_str();
const char *c1_name = c1_str.c_str();
const char *bkc_name = bkc_str.c_str();

IPC_PRELUDE(test_bookkeeper_simple_cond) {
  boost::interprocess::named_mutex::remove(tc1_name);
  boost::interprocess::named_semaphore::remove(tc2_name);
  boost::interprocess::named_semaphore::remove(tc3_name);
  boost::interprocess::shared_memory_object::remove(bkc_name);
}
IPC_SERVER(test_bookkeeper_simple_cond) {
  neb::ipc::internal::shm_bookkeeper sb(bkc_name);
  auto mutex = sb.acquire_named_mutex(tc1_name);
  bool v = mutex->try_lock();
  IPC_EXPECT(v);
  mutex->unlock();
  boost::interprocess::scoped_lock<boost::interprocess::named_mutex> _l(
      *mutex.get());
  auto sema = sb.acquire_named_semaphore(tc2_name);
  auto sema2 = sb.acquire_named_semaphore(tc3_name);
  auto cond = sb.acquire_named_condition(c1_name);
  sema->post();
  cond->wait(_l);
  sema2->wait();
  sb.release_named_mutex(tc1_name);
  sb.release_named_semaphore(tc2_name);
  sb.release_named_semaphore(tc3_name);
  sb.release_named_condition(c1_name);
}

IPC_CLIENT(test_bookkeeper_simple_cond) {
  neb::ipc::internal::shm_bookkeeper sb(bkc_name);
  auto sema = sb.acquire_named_semaphore(tc2_name);
  auto cond = sb.acquire_named_condition(c1_name);
  sema->wait();
  auto mutex = sb.acquire_named_mutex(tc1_name);
  cond->notify_one();
  auto sema2 = sb.acquire_named_semaphore(tc3_name);
  sema2->post();
  sb.release_named_mutex(tc1_name);
  sb.release_named_semaphore(tc2_name);
  sb.release_named_semaphore(tc3_name);
  sb.release_named_condition(c1_name);
}

IPC_PRELUDE(test_bookkeeper_release) {
  boost::interprocess::named_mutex::remove("tr1");
  boost::interprocess::named_semaphore::remove("tr2");
  boost::interprocess::named_condition::remove("cr1");
  boost::interprocess::shared_memory_object::remove(
      "test_bookkeeper_simple_release");
}

IPC_SERVER(test_bookkeeper_release) {

  neb::ipc::internal::shm_bookkeeper sb("test_bookkeeper_simple_release");
  auto mutex = sb.acquire_named_mutex("tr1");
  auto sema = sb.acquire_named_semaphore("tr2");
  auto cond = sb.acquire_named_condition("cr1");
  sb.release_named_mutex("tr1");
  sb.release_named_semaphore("tr2");
  sb.release_named_condition("cr1");

  IPC_EXPECT(neb::ipc::check_exists<boost::interprocess::named_mutex>("tr1") ==
             false);
  IPC_EXPECT(neb::ipc::check_exists<boost::interprocess::named_condition>(
                 "cr1") == false);
  IPC_EXPECT(neb::ipc::check_exists<boost::interprocess::named_semaphore>(
                 "tr2") == false);
}

IPC_CLIENT(test_bookkeeper_release) {}
