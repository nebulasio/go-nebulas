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
#include "fs/util.h"
#include "test/common/ipc/ipc_test.h"
#include <thread>

typedef neb::ipc::shm_service_server<128 * 1024> shm_server_t;
typedef neb::ipc::shm_service_util<128 * 1024> shm_util_t;
typedef neb::ipc::shm_service_client<128 * 1024> shm_client_t;

static std::string base_name = neb::fs::get_user_name() + "test_session_simple";

IPC_PRELUDE(test_service_simple) {
  boost::interprocess::named_mutex::remove(base_name.c_str());
  shm_util_t s(base_name, 128, 128);
  LOG(INFO) << "to reset";
  s.reset();
  LOG(INFO) << "reset done";
}

IPC_SERVER(test_service_simple) {
  neb::core::exception_handler eh;
  eh.run();
  std::thread thrd([]() {
    std::this_thread::sleep_for(std::chrono::seconds(20));
    neb::core::command_queue::instance().send_command(
        std::make_shared<neb::core::exit_command>());
  });
  shm_util_t us(base_name, 128, 128);
  us.reset();
  LOG(INFO) << "reset done";

  shm_server_t s(base_name, 128, 128);
  s.init_local_env();
  s.run();
  thrd.join();
  eh.kill();
  LOG(INFO) << "got client start!";

}
IPC_CLIENT(test_service_simple) {
  neb::core::exception_handler eh;
  eh.run();
  std::thread thrd([]() {
    std::this_thread::sleep_for(std::chrono::seconds(10));
    neb::core::command_queue::instance().send_command(
        std::make_shared<neb::core::exit_command>());
  });

  shm_client_t c(base_name, 128, 128);
  c.init_local_env();
  c.run();

  LOG(INFO) << " c done!";
  thrd.join();

  eh.kill();
}

struct SamplePkg {
  static constexpr neb::ipc::shm_type_id_t pkg_identifier = 12;

  SamplePkg(uint64_t v) : m_value(v) {}
  uint64_t m_value;
};
const neb::ipc::shm_type_id_t SamplePkg::pkg_identifier;

IPC_SERVER(test_service_message) {
  neb::core::exception_handler eh;
  eh.run();
  shm_util_t us(base_name, 128, 128);
  us.reset();
  shm_server_t s(base_name, 128, 128);
  s.init_local_env();
  s.add_handler<SamplePkg>([](SamplePkg *p) {
    LOG(INFO) << "got data from client " << p->m_value;
    IPC_EXPECT(p->m_value == 2);
    // neb::core::command_queue::instance().send_command(
    // std::make_shared<neb::core::exit_command>());
  });

  s.run();
  LOG(INFO) << "got client start!";
  eh.kill();

  // s.wait_till_finish();
}

IPC_CLIENT(test_service_message) {
  neb::core::exception_handler eh;
  eh.run();

  shm_client_t c(base_name, 128, 128);
  c.init_local_env();
  std::thread thrd([&c]() {
    std::this_thread::sleep_for(std::chrono::seconds(1));
    SamplePkg *pkg = c.construct<SamplePkg>(2);
    c.push_back(pkg);

    std::this_thread::sleep_for(std::chrono::seconds(10));
    neb::core::command_queue::instance().send_command(
        std::make_shared<neb::core::exit_command>());
  });
  c.run();
  thrd.join();

  // c.wait_till_finish();
  eh.kill();
}

IPC_SERVER(test_service_message_pingpong) {
  neb::core::exception_handler eh;
  eh.run();
  shm_util_t us(base_name, 128, 128);
  us.reset();
  shm_server_t s(base_name, 128, 128);
  s.init_local_env();

  uint64_t v = 0;
  s.add_handler<SamplePkg>([&s, &v](SamplePkg *p) {
    // std::cout << "got data from client " << p->m_value;
    IPC_EXPECT(p->m_value == v);

    v++;
    SamplePkg *pkg = s.construct<SamplePkg>(v);
    s.push_back(pkg);
    // LOG(INFO) << "push back data " << v;
    // neb::core::command_queue::instance().send_command(
    // std::make_shared<neb::core::exit_command>());
  });

  s.run();
  LOG(INFO) << "got client start!";
  eh.kill();
}

IPC_CLIENT(test_service_message_pingpong) {
  neb::core::exception_handler eh;
  eh.run();

  shm_client_t c(base_name, 128, 128);
  c.init_local_env();
  uint64_t v = 0;
  std::thread thrd([&c, v]() {
    std::this_thread::sleep_for(std::chrono::seconds(3));
    SamplePkg *pkg = c.construct<SamplePkg>(v);
    c.push_back(pkg);
  });

  c.add_handler<SamplePkg>([&c, &v](SamplePkg *p) {

    SamplePkg *pkg = c.construct<SamplePkg>(p->m_value);
    c.push_back(pkg);

    if (p->m_value > 10000) {
      neb::core::command_queue::instance().send_command(
          std::make_shared<neb::core::exit_command>());
    }
  });
  c.run();
  thrd.join();

  // c.wait_till_finish();
  eh.kill();
}

struct SampleStringPkg {
  static constexpr neb::ipc::shm_type_id_t pkg_identifier = 13;

  SampleStringPkg(uint64_t v, const char *str,
                  const neb::ipc::default_allocator_t &alloc)
      : m_value(v), m_string_val(str, alloc) {}
  uint64_t m_value;
  neb::ipc::char_string_t m_string_val;
};

const neb::ipc::shm_type_id_t SampleStringPkg::pkg_identifier;

IPC_SERVER(test_service_string_message) {
  neb::core::exception_handler eh;
  eh.run();
  shm_util_t us(base_name, 128, 128);
  us.reset();
  shm_server_t s(base_name, 128, 128);
  s.init_local_env();
  s.add_handler<SampleStringPkg>([](SampleStringPkg *p) {
    LOG(INFO) << "got data from client " << p->m_value << ", "
              << p->m_string_val;
    IPC_EXPECT(p->m_string_val == "test xxx");
  });

  s.run();
  LOG(INFO) << "got client start!";
  eh.kill();
}

IPC_CLIENT(test_service_string_message) {
  neb::core::exception_handler eh;
  eh.run();

  shm_client_t c(base_name, 128, 128);
  c.init_local_env();
  std::thread thrd([&c]() {
    std::this_thread::sleep_for(std::chrono::seconds(1));
    SampleStringPkg *pkg =
        c.construct<SampleStringPkg>(2, "test xxx", c.default_allocator());
    c.push_back(pkg);

    std::this_thread::sleep_for(std::chrono::seconds(10));
    neb::core::command_queue::instance().send_command(
        std::make_shared<neb::core::exit_command>());
  });
  c.run();
  thrd.join();

  // c.wait_till_finish();
  eh.kill();
}
