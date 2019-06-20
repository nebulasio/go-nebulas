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
#include "runtime/nr/graph/graph.h"

namespace neb {
namespace rt {

class non_recursive_remove_cycles_based_on_time_sequence_helper {
public:
  typedef transaction_graph::internal_graph_t graph_t;

  non_recursive_remove_cycles_based_on_time_sequence_helper(graph_t &tg);

  virtual void remove_cycles_based_on_time_sequence(graph_t &graph);

protected:
  virtual void
  show_path(const std::vector<transaction_graph::edge_descriptor_t> &edges);

  virtual void build_adj_graph(
      const graph_t &graph,
      std::unordered_map<transaction_graph::vertex_descriptor_t,
                         std::vector<transaction_graph::edge_descriptor_t>>
          &adj);

  virtual std::vector<transaction_graph::vertex_descriptor_t>
  possible_start_nodes_of_cycles(const graph_t &graph);

  virtual std::vector<transaction_graph::edge_descriptor_t>
  find_a_cycle_from_vertex_based_on_time_sequence(
      const graph_t &graph,
      const std::unordered_map<
          transaction_graph::vertex_descriptor_t,
          std::vector<transaction_graph::edge_descriptor_t>> &adj,
      const transaction_graph::vertex_descriptor_t &v,
      const std::unordered_set<transaction_graph::vertex_descriptor_t> &dead_v,
      const std::unordered_map<transaction_graph::vertex_descriptor_t, size_t>
          &dead_to,
      const std::unordered_map<transaction_graph::vertex_descriptor_t, size_t>
          &to_dead);

  virtual std::vector<transaction_graph::edge_descriptor_t>
  find_a_cycle_based_on_time_sequence(
      const graph_t &graph,
      const std::unordered_map<
          transaction_graph::vertex_descriptor_t,
          std::vector<transaction_graph::edge_descriptor_t>> &adj,
      const std::vector<transaction_graph::vertex_descriptor_t> &start_nodes,
      const std::unordered_set<transaction_graph::vertex_descriptor_t> &dead_v,
      const std::unordered_map<transaction_graph::vertex_descriptor_t, size_t>
          &dead_to,
      const std::unordered_map<transaction_graph::vertex_descriptor_t, size_t>
          &to_dead);

  virtual void remove_a_cycle(
      graph_t &graph,
      std::unordered_map<transaction_graph::vertex_descriptor_t,
                         std::vector<transaction_graph::edge_descriptor_t>>
          &adj,
      const std::vector<transaction_graph::edge_descriptor_t> &edges);

  virtual bool decrease_graph_edges(
      const graph_t &graph,
      std::unordered_set<transaction_graph::vertex_descriptor_t> &dead_v,
      std::unordered_map<transaction_graph::vertex_descriptor_t, size_t>
          &dead_to,
      std::unordered_map<transaction_graph::vertex_descriptor_t, size_t>
          &to_dead);

  virtual void bfs_decrease_graph_edges(
      const graph_t &graph,
      const std::unordered_set<transaction_graph::vertex_descriptor_t> &dead_v,
      std::unordered_set<transaction_graph::vertex_descriptor_t> &tmp_dead,
      std::unordered_map<transaction_graph::vertex_descriptor_t, size_t>
          &dead_to,
      std::unordered_map<transaction_graph::vertex_descriptor_t, size_t>
          &to_dead);

protected:
  graph_t m_graph;
};

} // namespace rt
} // namespace neb
