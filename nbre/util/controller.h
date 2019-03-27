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
#include "core/net_ipc/nipc_pkg.h"
#include <ff/network.h>

enum controller_pkg_type {
  ctl_kill_req_pkg = nipc_last_pkg_id,

  ctl_last_pkg_id,
};

namespace neb {
namespace util {

typedef ff::net::ntpackage<ctl_kill_req_pkg> ctl_kill_req_t;

class elfin {
public:
  void run();

protected:
  void handle_kill_req();
};

class magic_wand {
public:
  void kill_nbre();

protected:
  void start_and_join();

protected:
  std::shared_ptr<ff::net::package> m_package;
};

} // namespace util
} // namespace neb

