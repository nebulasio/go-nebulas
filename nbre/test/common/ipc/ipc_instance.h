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
#include "common/util/singleton.h"
#include <unordered_set>

namespace neb {
class ipc_instance_base {
public:
  virtual std::string get_fixture_name() = 0;
  virtual std::string get_instance_name() = 0;
  virtual void run() const = 0;
}; // end class ipc_instance_base

typedef std::shared_ptr<ipc_instance_base> ipc_instance_base_ptr;

class ipc_instances : public util::singleton<ipc_instances> {
public:
  void init_ipc_instances(int argc, char *argv[]);

  virtual ~ipc_instances();

  int run_all_ipc_instances();

  int run_one_ipc_instance(const std::string &fixture,
                           const std::string &instance);

  size_t register_ipc_instance(const ipc_instance_base_ptr &b);

protected:
  void show_all_instances();

protected:
  std::vector<ipc_instance_base_ptr> m_all_instances;
  std::string m_enabled_instance_name;
  std::string m_enabled_fixture_name;
  std::string m_exe_name;
  bool m_debug_mode;
}; // end class ipc_instances
} // namespace neb

#define GEN_NAME_VAR(name) _##name##_nouse
#define CLASS_NAME(f, n) _class_##f##_##n##_
#define TMP_VAR_NAME(f, n) _tmp_var_##f##_##n##_

#define _IPC_INSTANCE(fixture, name)                                           \
  class CLASS_NAME(fixture, name) : public neb::ipc_instance_base {            \
  public:                                                                      \
    virtual std::string get_fixture_name() { return #fixture; }                \
    virtual std::string get_instance_name() { return #name; }                  \
    virtual void run() const;                                                  \
  };                                                                           \
  static int TMP_VAR_NAME(fixture, name) =                                     \
      neb::ipc_instances::instance().register_ipc_instance(                    \
          std::static_pointer_cast<neb::ipc_instance_base>(                    \
              std::make_shared<CLASS_NAME(fixture, name)>()));                 \
  void CLASS_NAME(fixture, name)::run() const

#define IPC_PRELUDE(fixture) _IPC_INSTANCE(fixture, prelude)

#define IPC_SERVER(fixture) _IPC_INSTANCE(fixture, server)

#define IPC_CLIENT(fixture) _IPC_INSTANCE(fixture, client)
