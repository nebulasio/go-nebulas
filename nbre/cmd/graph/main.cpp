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

#include "runtime/nr/graph/algo.h"

int main(int argc, char *argv[]) {

  neb::rt::transaction_graph tg;
  tg.add_edge("a", "b", 1, 2);
  tg.add_edge("a", "c", 3, 4);

  neb::rt::graph_algo::get_degree_sum(tg.internal_graph());

  return 0;
}
