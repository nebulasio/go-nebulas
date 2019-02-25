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

#if 0
#include "common/common.h"
#include "core/driver.h"
#include "fs/util.h"

int main(int argc, char *argv[]) {
  // FLAGS_logtostderr = true;
  neb::program_name = "nbre";
  //::google::InitGoogleLogging(argv[0]);

  assert(argc > 1);
  neb::shm_configuration::instance().shm_name_identity() = argv[1];

  neb::core::driver d;
  d.init();
  d.run();
  LOG(INFO) << "to quit nbre";

  return 0;
}
#endif

#include "common/common.h"
#include "common/configuration.h"
#include "core/net_ipc/client/client_driver.h"
#include "fs/util.h"

int main(int argc, char *argv[]) {
  FLAGS_logtostderr = true;
  neb::program_name = "nbre";

  LOG(INFO) << "nbre started!";
  assert(argc > 2);
  LOG(INFO) << "pass args " << argv[1] << ',' << argv[2];
  neb::configuration::instance().nipc_listen() = argv[1];
  neb::configuration::instance().nipc_port() = std::stoi(argv[2]);

  neb::core::client_driver d;
  d.init();
  d.run();
  LOG(INFO) << "to quit nbre";

  return 0;
}
