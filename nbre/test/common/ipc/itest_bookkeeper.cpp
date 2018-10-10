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
#include "test/common/ipc/ipc_test.h"
#include <chrono>
#include <exception>
#include <thread>

IPC_SERVER(test_bookkeeper_simple) {

  //! In case test failure, we may need remove this to avoid deadlock.
  boost::interprocess::named_mutex::remove("t1");
  boost::interprocess::named_semaphore::remove("t2");
  boost::interprocess::named_semaphore::remove("t3");
  boost::interprocess::shared_memory_object::remove("test_bookkeeper_simple");

  neb::ipc::internal::shm_bookkeeper sb("test_bookkeeper_simple");
  auto mutex = sb.acquire_named_mutex("t1");
  bool v = mutex->try_lock();
  IPC_EXPECT(v);
  auto sema = sb.acquire_named_semaphore("t2");
  auto sema2 = sb.acquire_named_semaphore("t3");
  sema->post();
  sema2->wait();
  sb.release_named_mutex("t1");
  sb.release_named_semaphore("t2");
  sb.release_named_semaphore("t3");
}

IPC_CLIENT(test_bookkeeper_simple) {
  neb::ipc::internal::shm_bookkeeper sb("test_bookkeeper_simple");
  auto sema = sb.acquire_named_semaphore("t2");
  sema->wait();
  auto mutex = sb.acquire_named_mutex("t1");
  bool v = mutex->try_lock();
  IPC_EXPECT(v == false);
  auto sema2 = sb.acquire_named_semaphore("t3");
  sema2->post();
  sb.release_named_mutex("t1");
  sb.release_named_semaphore("t2");
  sb.release_named_semaphore("t3");
}

IPC_SERVER(test_bookkeeper_simple_cond) {

  //! In case test failure, we may need remove this to avoid deadlock.
  boost::interprocess::named_mutex::remove("tc1");
  boost::interprocess::named_semaphore::remove("tc2");
  boost::interprocess::named_semaphore::remove("tc3");
  boost::interprocess::shared_memory_object::remove(
      "test_bookkeeper_simple_cond");

  neb::ipc::internal::shm_bookkeeper sb("test_bookkeeper_simple_cond");
  auto mutex = sb.acquire_named_mutex("tc1");
  bool v = mutex->try_lock();
  IPC_EXPECT(v);
  mutex->unlock();
  boost::interprocess::scoped_lock<boost::interprocess::named_mutex> _l(
      *mutex.get());
  auto sema = sb.acquire_named_semaphore("tc2");
  auto sema2 = sb.acquire_named_semaphore("tc3");
  auto cond = sb.acquire_named_condition("c1");
  sema->post();
  cond->wait(_l);
  sema2->wait();
  sb.release_named_mutex("tc1");
  sb.release_named_semaphore("tc2");
  sb.release_named_semaphore("tc3");
  sb.release_named_condition("c1");
}

IPC_CLIENT(test_bookkeeper_simple_cond) {
  neb::ipc::internal::shm_bookkeeper sb("test_bookkeeper_simple_cond");
  auto sema = sb.acquire_named_semaphore("tc2");
  auto cond = sb.acquire_named_condition("c1");
  sema->wait();
  auto mutex = sb.acquire_named_mutex("tc1");
  cond->notify_one();
  auto sema2 = sb.acquire_named_semaphore("tc3");
  sema2->post();
  sb.release_named_mutex("tc1");
  sb.release_named_semaphore("tc2");
  sb.release_named_semaphore("tc3");
  sb.release_named_condition("c1");
}
