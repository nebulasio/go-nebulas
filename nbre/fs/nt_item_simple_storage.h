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
#include "core/net_ipc/nipc_pkg.h"
#include "fs/nt_items_storage.h"
#include "fs/storage_holder.h"
#include "util/singleton.h"

namespace neb {
namespace fs {
template <class ItemType, class ItemTraits>
class nt_items_simple_storage : public nt_items_storage<ItemType> {
public:
  typedef ItemType item_type;
  typedef ItemTraits item_traits;

  nt_items_simple_storage(storage *db)
      : nt_items_storage<ItemType>(db, ItemTraits::key_prefix,
                                   ItemTraits::last_item_key,
                                   ItemTraits::block_trunk_size) {}
};

namespace internal {
struct nr_param_storage_traits {
  constexpr static const char *key_prefix = "nr_param_";
  constexpr static const char *last_item_key = "nr_param_last_block";
  constexpr static size_t block_trunk_size = 16;
};

struct dip_param_storage_traits {
  constexpr static const char *key_prefix = "dip_param_";
  constexpr static const char *last_item_key = "dip_param_last_block";
  constexpr static size_t block_trunk_size = 16;
};

struct auth_param_storage_traints {
  constexpr static const char *key_prefix = "auth_param_";
  constexpr static const char *last_item_key = "auth_param_last_block";
  constexpr static size_t block_trunk_size = 32;
};
} // namespace internal
typedef nt_items_simple_storage<nr_param_storage_t,
                                internal::nr_param_storage_traits>
    nr_params_storage_t;

typedef nt_items_simple_storage<dip_param_storage_t,
                                internal::dip_param_storage_traits>
    dip_params_storage_t;

typedef nt_items_simple_storage<auth_param_storage_t,
                                internal::auth_param_storage_traints>
    auth_params_storage_t;
} // namespace fs
} // namespace neb
