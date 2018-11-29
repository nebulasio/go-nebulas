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

#include "runtime/nr/impl/nebulas_rank.h"

int main(int argc, char *argv[]) {

  neb::rt::nr::transaction_db_ptr_t tdb_ptr;
  neb::rt::nr::account_db_ptr_t adb_ptr;
  neb::rt::nr::rank_params_t para;
  neb::block_height_t start_block = 0;
  neb::block_height_t end_block = 0;

  neb::rt::nr::nebulas_rank nr(tdb_ptr, adb_ptr, para, start_block, end_block);
  return 0;
}
