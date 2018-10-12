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
#include "common/ipc/shm_queue.h"
#include "common/quitable_thread.h"
#include "shm_service_op_queue.h"

namespace neb {
namespace ipc {
namespace internal {

class shm_queue_watcher : public quitable_thread {
public:
  shm_queue_watcher(shm_queue *queue, shm_service_op_queue *op_queue);

protected:
  virtual void thread_func();

protected:
  shm_queue *m_queue;
  shm_service_op_queue *m_op_queue;
};

} // namespace internal
} // namespace ipc
} // namespace neb
