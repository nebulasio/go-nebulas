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

#include "fs/storage.h"
#include "util/thread_safe_map.h"

namespace neb {
namespace fs {
class memory_storage : public storage {
public:
  memory_storage();
  virtual ~memory_storage();
  memory_storage(const memory_storage &ms) = delete;
  memory_storage &operator=(const memory_storage &) = delete;

  virtual bytes get_bytes(const bytes &key);
  virtual void put_bytes(const bytes &key, const bytes &val);
  virtual void del_by_bytes(const bytes &key);

  virtual void enable_batch();
  virtual void disable_batch();
  virtual void flush();

protected:
  thread_safe_map<bytes, bytes> m_memory;
};
} // namespace fs
} // namespace neb
