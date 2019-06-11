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
#include "fs/proto/ir.pb.h"
#include "fs/rocksdb_storage.h"

namespace neb {
namespace fs {
namespace internal {

class ir_item_list_interface {
public:
  virtual void write_ir(const nbre::NBREIR &raw_ir,
                        const nbre::NBREIR &compiled_ir) = 0;

  virtual nbre::NBREIR get_raw_ir(version_t v) = 0;
  virtual nbre::NBREIR get_ir(version_t v) = 0;
};

template <class PST> class ir_item_list_base {
public:
  typedef PST param_storage_t;
  typedef typename param_storage_t::item_type item_type;
  typedef typename param_storage_t::item_traits item_traits;

  ir_item_list_base(rocksdb_storage *storage, const std::string &name)
      : m_storage(storage), m_ir_name(name){};

  virtual void write_ir(const nbre::NBREIR &raw_ir,
                        const nbre::NBREIR &compiled_ir) {
    std::string raw_str = raw_ir.SerializeAsString();
    std::string ir_str = compiled_ir.SerializeAsString();
    item_type it = get_ir_param(compiled_ir);
    auto raw_key = std::to_string(raw_ir_key(raw_ir.version()));
    auto ir_key = std::to_string(compiled_ir_key(compiled_ir.version()));
    m_storage->put(raw_key, raw_str);
    m_storage->put(ir_key, ir_str);
    it.template set<p_raw_key>(raw_key);
    it.template set<p_ir_key>(ir_key);
    m_param_storage.append_item(it);
  }

  virtual nbre::NBREIR get_raw_ir(version_t v) {
    nbre::NBREIR ir;
    bytes bs = m_storage->get_bytes(raw_ir_key(v));
    ir.ParseFromArray(bs.value(), bs.size());
    return ir;
  }
  virtual nbre::NBREIR get_ir(version_t v) {
    nbre::NBREIR ir;
    bytes bs = m_storage->get_bytes(compiled_ir_key(v));
    ir.ParseFromArray(bs.value(), bs.size());
    return ir;
  }

  virtual item_type get_ir_param(const nbre::NBREIR &compiled_ir) = 0;

protected:
  bytes raw_ir_key(version_t v) { return gen_key(v) + std::string("r_ir"); }
  bytes compiled_ir_key(version_t v) {
    return gen_key(v) + std::string("c_ir");
  }
  bytes gen_key(version_t v) {
    auto b1 = number_to_byte<bytes>(v);
    return b1 + m_ir_name;
  }

protected:
  rocksdb_storage *m_storage;
  std::string m_ir_name;
  param_storage_t m_param_storage;
};
} // namespace internal
} // namespace fs
} // namespace neb
