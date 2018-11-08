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
#include "core/neb_ipc/ipc_common.h"

//! According to boost.interprocess document, we should never use reference,
//! pointer to store data, and we should never use virtual functions here. We
//! must make sure each class here is POD, thus we can pass it with interprocess
//! communication.
//
//! Also, our IPC framework needs *pkg_identifier* in each class.
namespace neb {
namespace core {
using ipc_pkg_type_id_t = neb::ipc::shm_type_id_t;

namespace internal {
template <size_t id, typename T, bool is_prim = std::is_arithmetic<T>::value>
struct ipc_elem_base {
  typedef T type;
  ipc_elem_base(const neb::ipc::default_allocator_t &alloc) : __value(alloc) {}
  type __value;
};

template <size_t id, typename T> struct ipc_elem_base<id, T, true> {
  typedef T type;
  ipc_elem_base(const neb::ipc::default_allocator_t &) : __value() {}
  type __value;
};

template <ipc_pkg_type_id_t id_t, typename... ARGS>
struct define_ipc_pkg : public ARGS... {
  define_ipc_pkg(void *holder, const neb::ipc::default_allocator_t &alloc)
      : ARGS(alloc)..., m_holder(holder), m_alloc(alloc) {}
  const static ipc_pkg_type_id_t pkg_identifier = id_t;
  template <typename T> const typename T::type &get() const {
    return T::__value;
  }
  template <typename T> typename T::type &get() { return T::__value; }
  template <typename T> void set(const typename T::type &t) { T::__value = t; }
  template <typename T>
  auto set(const char *t) -> typename std::enable_if<
      std::is_same<typename T::type, neb::ipc::char_string_t>::value,
      void>::type {
    T::__value = neb::ipc::char_string_t(t, m_alloc);
  }

  void *m_holder; // used by C-GO
  const neb::ipc::default_allocator_t &m_alloc;
};

template <ipc_pkg_type_id_t id_t, typename... ARGS>
const ipc_pkg_type_id_t define_ipc_pkg<id_t, ARGS...>::pkg_identifier;

template <typename PkgType> struct pkg_to_module_info {
  constexpr static const bool is_for_module = false;
  constexpr static const char *module_name = "default";
  constexpr static const char *func_name = "default";
};

} // namespace internal

#define add_pkg_module_info(pkg_type, module, func)                            \
  template <> struct internal::pkg_to_module_info<pkg_type> {                  \
    constexpr static const bool is_for_module = true;                          \
    constexpr static const char *module_name = module;                         \
    constexpr static const char *func_name = func;                             \
  };

template <typename T>
using pkg_type_to_module_info = internal::pkg_to_module_info<T>;

} // namespace core
} // namespace neb
