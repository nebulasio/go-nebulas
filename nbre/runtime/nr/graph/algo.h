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
#include "common/common.h"
#include "runtime/nr/graph/graph.h"

namespace neb {
namespace rt {

struct in_out_val_t {
  wei_t m_in_val;
  wei_t m_out_val;
};

struct in_out_degree_t {
  uint32_t m_in_degree;
  uint32_t m_out_degree;
};

class graph_algo {
public:
  static void remove_cycles_based_on_time_sequence(
      transaction_graph::internal_graph_t &graph);

  static void non_recursive_remove_cycles_based_on_time_sequence(
      transaction_graph::internal_graph_t &graph);

  static void merge_edges_with_same_from_and_same_to(
      transaction_graph::internal_graph_t &graph);

  static transaction_graph *
  merge_graphs(const std::vector<transaction_graph_ptr_t> &graphs);

  static void merge_topk_edges_with_same_from_and_same_to(
      transaction_graph::internal_graph_t &graph, uint32_t k = 3);

  static auto get_in_out_vals(const transaction_graph::internal_graph_t &graph)
      -> std::unique_ptr<std::unordered_map<address_t, in_out_val_t>>;

  static auto get_stakes(const transaction_graph::internal_graph_t &graph)
      -> std::unique_ptr<std::unordered_map<address_t, wei_t>>;

  static auto
  get_in_out_degrees(const transaction_graph::internal_graph_t &graph)
      -> std::unique_ptr<std::unordered_map<address_t, in_out_degree_t>>;

  static auto get_degree_sum(const transaction_graph::internal_graph_t &graph)
      -> std::unique_ptr<std::unordered_map<address_t, uint32_t>>;

  static bool decrease_graph_edges(
      const transaction_graph::internal_graph_t &graph,
      std::unordered_set<transaction_graph::vertex_descriptor_t> &dead_v,
      std::unordered_map<transaction_graph::vertex_descriptor_t, size_t>
          &dead_to,
      std::unordered_map<transaction_graph::vertex_descriptor_t, size_t>
          &to_dead);

#ifdef NDEBUG
private:
#else
public:
#endif
  static void dfs_find_a_cycle_from_vertex_based_on_time_sequence(
      const transaction_graph::internal_graph_t &graph,
      const transaction_graph::vertex_descriptor_t &start_vertex,
      const transaction_graph::vertex_descriptor_t &v,
      const std::unordered_set<transaction_graph::vertex_descriptor_t> &dead_v,
      bool &has_cycle,
      std::unordered_map<transaction_graph::vertex_descriptor_t, bool> &visited,
      std::vector<transaction_graph::edge_descriptor_t> &edges,
      std::vector<transaction_graph::edge_descriptor_t> &ret);

  static auto find_a_cycle_from_vertex_based_on_time_sequence(
      const transaction_graph::internal_graph_t &graph,
      const transaction_graph::vertex_descriptor_t &v,
      const std::unordered_set<transaction_graph::vertex_descriptor_t> &dead_v)
      -> std::vector<transaction_graph::edge_descriptor_t>;

  static auto find_a_cycle_based_on_time_sequence(
      const transaction_graph::internal_graph_t &graph,
      const std::unordered_set<transaction_graph::vertex_descriptor_t> &dead_v)
      -> std::vector<transaction_graph::edge_descriptor_t>;

  static void bfs_decrease_graph_edges(
      const transaction_graph::internal_graph_t &graph,
      const std::unordered_set<transaction_graph::vertex_descriptor_t> &dead_v,
      std::unordered_set<transaction_graph::vertex_descriptor_t> &tmp_dead,
      std::unordered_map<transaction_graph::vertex_descriptor_t, size_t>
          &dead_to,
      std::unordered_map<transaction_graph::vertex_descriptor_t, size_t>
          &to_dead);

  static void remove_a_cycle(
      transaction_graph::internal_graph_t &graph,
      const std::vector<transaction_graph::edge_descriptor_t> &edges);

  static transaction_graph *merge_two_graphs(transaction_graph *tg,
                                             const transaction_graph *sg);
};

} // namespace rt
} // namespace neb
