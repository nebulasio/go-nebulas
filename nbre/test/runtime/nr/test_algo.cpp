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
  int32_t n = std::stoi(argv[1]);
  std::string s = argv[2];
  char cc = s[0];

  for (char ch = 'a'; ch < cc; ch++) {
    for (int32_t i = 0; i < n; i++) {
      tg.add_edge(neb::to_address(std::string(1, ch)),
                  neb::to_address(std::string(1, ch + 1)), ch - 'a' + 1,
                  ch - 'a' + 1);
    }
  }
  for (int32_t i = 0; i < n; i++) {
    tg.add_edge(neb::to_address(std::string(1, cc)),
                neb::to_address(std::string(1, 'a')), cc - 'a' + 1,
                cc - 'a' + 1);
  }

  auto graph = tg.internal_graph();
  LOG(INFO) << "start to remove cycle";
  neb::rt::graph_algo::remove_cycles_based_on_time_sequence(graph);
  LOG(INFO) << "done with remove cycle";

  return 0;
}
