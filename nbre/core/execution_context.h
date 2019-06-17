// Copyright (C) 2017 go-nebulas authors
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

namespace neb {
namespace fs {
class storage;
class ir_processor;
class blockchain;
}
namespace compatible {
class compatible_check_interface;
}
namespace core {
class execution_context {
public:
  execution_context();
  virtual ~execution_context();
  virtual fs::storage *blockchain_storage() = 0;
  virtual fs::storage *nbre_storage() = 0;
  virtual compatible::compatible_check_interface *compatible_checker() = 0;
  virtual fs::ir_processor *ir_processor() = 0;
  virtual fs::blockchain *blockchain() = 0;

  virtual void shutdown() = 0;

  virtual bool is_ready() const;
  virtual void wait_until_ready();
  virtual void set_ready();

protected:
  std::condition_variable m_cond_var;
  mutable std::mutex m_mutex;
  bool m_ready_flag;
};

extern std::unique_ptr<execution_context> context;
} // namespace core
} // namespace neb
