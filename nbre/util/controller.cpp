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
#include "util/controller.h"
#include <boost/process/environment.hpp>

namespace neb {
namespace util {
void elfin::run() {
  try {
    ff::net::net_nervure nn;
    ff::net::typed_pkg_hub hub;
    hub.to_recv_pkg<ctl_kill_req_t>(
        [this](std::shared_ptr<ctl_kill_req_t>) { handle_kill_req(); });

    nn.add_pkg_hub(hub);
    nn.add_tcp_server("127.0.0.1", 0x1969);

    nn.run();
  } catch (...) {
    LOG(ERROR) << "got exception";
  }
}

void elfin::handle_kill_req() {
  LOG(INFO) << boost::this_process::get_id();
  LOG(ERROR) << "got killed";
  throw std::invalid_argument("got kill command");
}

void magic_wand::kill_nbre() {
  m_package = std::make_shared<ctl_kill_req_t>();
  start_and_join();
}

void magic_wand::start_and_join() {
  ff::net::net_nervure nn;
  ff::net::typed_pkg_hub hub;
  ff::net::tcp_connection_base_ptr conn;
  nn.get_event_handler()->listen<::ff::net::event::tcp_get_connection>(
      [&, this](::ff::net::tcp_connection_base *conn) {
        conn->send(m_package);
      });
  nn.get_event_handler()->listen<::ff::net::event::tcp_lost_connection>(
      [&](::ff::net::tcp_connection_base *) { nn.ioservice().stop(); });

  nn.add_pkg_hub(hub);
  conn = nn.add_tcp_client("127.0.0.1", 0x1969);

  nn.run();
}
} // namespace util
} // namespace neb
