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
#include "common/ipc/shm_service.h"
#include "fs/util.h"

namespace neb {
namespace core {
#define SHM_SIZE 1 << 30
typedef neb::ipc::shm_service_server<SHM_SIZE> ipc_server_t;
typedef neb::ipc::shm_service_client<SHM_SIZE> ipc_client_t;
typedef neb::ipc::shm_service_util<SHM_SIZE> ipc_util_t;
static std::string shm_service_name_str =
    std::string("nbre.") + neb::fs::get_user_name();
static const char *shm_service_name = shm_service_name_str.c_str();
}
} // namespace neb
