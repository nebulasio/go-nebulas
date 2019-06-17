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
#include "core/execution_context.h"

namespace neb {
namespace core {

class client_context : public execution_context {
public:
  std::unique_ptr<fs::storage> m_bc_storage;
  std::unique_ptr<fs::storage> m_nbre_storage;
  std::unique_ptr<compatible::compatible_check_interface> m_compatible_checker;
  std::unique_ptr<fs::ir_processor> m_ir_processor;
  std::unique_ptr<fs::blockchain> m_blockchain;

  client_context();
  virtual ~client_context();
  virtual fs::storage *blockchain_storage();
  virtual fs::storage *nbre_storage();
  virtual compatible::compatible_check_interface *compatible_checker();
  virtual fs::ir_processor *ir_processor();
  virtual fs::blockchain *blockchain();
  virtual void shutdown();
};
} // namespace core
} // namespace neb
