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
#include "common/address.h"
#include "runtime/nr/graph/data_type.h"
#include "runtime/nr/graph/graph.h"

namespace neb {
namespace rt {

class graph_algo {
public:
  typedef transaction_graph::internal_graph_t graph_t;

  virtual void
  non_recursive_remove_cycles_based_on_time_sequence(graph_t &graph);

  virtual void merge_edges_with_same_from_and_same_to(graph_t &graph);

  virtual transaction_graph *
  merge_graphs(const std::vector<transaction_graph_ptr_t> &graphs);

  virtual void merge_topk_edges_with_same_from_and_same_to(graph_t &graph,
                                                           uint32_t k = 3);

  virtual std::unordered_map<address_t, in_out_val_t>
  get_in_out_vals(const graph_t &graph);

protected:
  virtual transaction_graph *merge_two_graphs(transaction_graph *tg,
                                              const transaction_graph *sg);
};

} // namespace rt
} // namespace neb
