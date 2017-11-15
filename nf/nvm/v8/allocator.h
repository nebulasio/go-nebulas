// Copyright (C) 2017 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or
// modify it under the terms of the GNU General Public License as published by
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
#ifndef _NEBULAS_NF_NVM_V8_ALLOCATOR_H_
#define _NEBULAS_NF_NVM_V8_ALLOCATOR_H_

#include <stdint.h>
#include <v8.h>

using namespace v8;

class ArrayBufferAllocator : public ArrayBuffer::Allocator {
public:
  ArrayBufferAllocator();
  virtual ~ArrayBufferAllocator();

  /**
   * Allocate |length| bytes. Return NULL if allocation is not successful.
   * Memory should be initialized to zeroes.
   */
  virtual void *Allocate(size_t length);

  /**
   * Allocate |length| bytes. Return NULL if allocation is not successful.
   * Memory does not have to be initialized.
   */
  virtual void *AllocateUninitialized(size_t length);

  /**
   * Free the memory block of size |length|, pointed to by |data|.
   * That memory is guaranteed to be previously allocated by |Allocate|.
   */
  virtual void Free(void *data, size_t length);

  size_t total_available_size();

  size_t peak_allocated_size();

private:
  size_t total_allocated_size_;
  size_t peak_allocated_size_;
};

#endif // _NEBULAS_NF_NVM_V8_ALLOCATOR_H_
