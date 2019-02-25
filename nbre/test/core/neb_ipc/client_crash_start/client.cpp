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
#include "common/common.h"
#include "common/configuration.h"
#include "core/driver.h"
#include "core/ir_warden.h"
#include "core/neb_ipc/ipc_interface.h"
#include "core/neb_ipc/server/ipc_server_endpoint.h"
#include "fs/util.h"
#include <ff/functionflow.h>

class simple_driver : public neb::core::internal::driver_base {
public:
  virtual bool init() {
    m_client = std::unique_ptr<neb::core::ipc_client_endpoint>(
        new neb::core::ipc_client_endpoint());
    add_handlers();

    //! we should make share wait_until_sync first

    bool ret = m_client->start();
    if (!ret)
      return ret;

    throw std::exception();
    m_ipc_conn = m_client->ipc_connection();

    auto p = m_ipc_conn->construct<neb::core::ipc_pkg::nbre_init_req>(
        nullptr, m_ipc_conn->default_allocator());
    m_ipc_conn->push_back(p);

    return true;
  }
  virtual void add_handlers() {
    m_client->add_handler<neb::core::ipc_pkg::nbre_version_req>(
        [this](neb::core::ipc_pkg::nbre_version_req *req) {
            LOG(INFO) << " to start jit driver for data";
            using mi = neb::core::pkg_type_to_module_info<
                neb::core::ipc_pkg::nbre_version_req>;

            neb::core::ipc_pkg::nbre_version_ack *ack =
                m_ipc_conn->construct<neb::core::ipc_pkg::nbre_version_ack>(
                    req->m_holder, m_ipc_conn->default_allocator());
            ack->set<neb::core::ipc_pkg::major>(1);
            ack->set<neb::core::ipc_pkg::minor>(1);
            ack->set<neb::core::ipc_pkg::patch>(1);
            throw std::exception();
            // m_ipc_conn->push_back(ack);

        });

    m_client->add_handler<neb::core::ipc_pkg::nbre_init_ack>(
        [](neb::core::ipc_pkg::nbre_init_ack *ack) {
          std::string s =
              ack->get<neb::core::ipc_pkg::admin_pub_addr>().c_str();
          LOG(INFO) << "got admin_pub_addr: " << s;
          neb::configuration::instance().admin_pub_addr() = s;
        });
  }
};

int main(int argc, char *argv[]) {
  FLAGS_logtostderr = true;

  ::google::InitGoogleLogging(argv[0]);
  neb::program_name = "nbre";

  simple_driver d;
  d.init();
  d.run();
  LOG(INFO) << "to quit nbre";

  return 0;
}
