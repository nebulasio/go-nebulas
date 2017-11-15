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

#include "allocator.h"
#include <stdio.h>
#include <stdlib.h>

ArrayBufferAllocator::ArrayBufferAllocator()
    : total_allocated_size_(0L), peak_allocated_size_(0L) {}

ArrayBufferAllocator::~ArrayBufferAllocator() {}

/**
 * Allocate |length| bytes. Return NULL if allocation is not successful.
 * Memory should be initialized to zeroes.
 */
void *ArrayBufferAllocator::Allocate(size_t length) {
  this->total_allocated_size_ += length;
  if (this->total_allocated_size_ > this->peak_allocated_size_) {
    this->peak_allocated_size_ = this->total_allocated_size_;
  }
  return calloc(length, 1);
}

/**
 * Allocate |length| bytes. Return NULL if allocation is not successful.
 * Memory does not have to be initialized.
 */
void *ArrayBufferAllocator::AllocateUninitialized(size_t length) {
  this->total_allocated_size_ += length;
  if (this->total_allocated_size_ > this->peak_allocated_size_) {
    this->peak_allocated_size_ = this->total_allocated_size_;
  }
  return malloc(length);
}

/**
 * Free the memory block of size |length|, pointed to by |data|.
 * That memory is guaranteed to be previously allocated by |Allocate|.
 */
void ArrayBufferAllocator::Free(void *data, size_t length) {
  this->total_allocated_size_ -= length;
  free(data);
}

size_t ArrayBufferAllocator::total_available_size() {
  return this->total_allocated_size_;
}

size_t ArrayBufferAllocator::peak_allocated_size() {
  return this->peak_allocated_size_;
}
