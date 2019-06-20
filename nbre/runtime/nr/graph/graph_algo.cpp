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

#include "runtime/nr/graph/graph_algo.h"
#include "runtime/nr/graph/graph_algo_helper.h"
#include "util/chrono.h"

namespace neb {
namespace rt {

void graph_algo::non_recursive_remove_cycles_based_on_time_sequence(
    transaction_graph::internal_graph_t &graph) {
  non_recursive_remove_cycles_based_on_time_sequence_helper nh(graph);
  nh.remove_cycles_based_on_time_sequence(graph);
}

void graph_algo::merge_edges_with_same_from_and_same_to(
    transaction_graph::internal_graph_t &graph) {

  transaction_graph::viterator_t vi, vi_end;

  for (boost::tie(vi, vi_end) = boost::vertices(graph); vi != vi_end; vi++) {
    transaction_graph::oeiterator_t oei, oei_end;
    std::unordered_map<transaction_graph::vertex_descriptor_t, wei_t>
        target_and_vals;
    for (boost::tie(oei, oei_end) = boost::out_edges(*vi, graph);
         oei != oei_end; oei++) {
      auto target = boost::target(*oei, graph);
      wei_t val = boost::get(boost::edge_weight_t(), graph, *oei);
      if (target_and_vals.find(target) == target_and_vals.end()) {
        target_and_vals.insert(std::make_pair(target, val));
      } else {
        target_and_vals[target] += val;
      }
    }

    bool removed_all_edges = false;
    while (!removed_all_edges) {
      removed_all_edges = true;
      for (boost::tie(oei, oei_end) = boost::out_edges(*vi, graph);
           oei != oei_end; oei++) {
        removed_all_edges = false;
        boost::remove_edge(oei, graph);
        break;
      }
    }

    for (auto it = target_and_vals.begin(); it != target_and_vals.end(); it++) {
      boost::add_edge(*vi, it->first, {it->second, 0}, graph);
    }
  }
  return;
}

transaction_graph *graph_algo::merge_two_graphs(transaction_graph *tg,
                                                const transaction_graph *sg) {

  transaction_graph::internal_graph_t sgi = sg->internal_graph();
  transaction_graph::viterator_t vi, vi_end;

  for (boost::tie(vi, vi_end) = boost::vertices(sgi); vi != vi_end; vi++) {
    transaction_graph::oeiterator_t oei, oei_end;
    for (boost::tie(oei, oei_end) = boost::out_edges(*vi, sgi); oei != oei_end;
         oei++) {
      auto source = boost::source(*oei, sgi);
      auto target = boost::target(*oei, sgi);
      auto from = to_address(boost::get(boost::vertex_name_t(), sgi, source));
      auto to = to_address(boost::get(boost::vertex_name_t(), sgi, target));
      wei_t w = boost::get(boost::edge_weight_t(), sgi, *oei);

      tg->add_edge(from, to, w, 0);
    }
  }
  return tg;
}

transaction_graph *
graph_algo::merge_graphs(const std::vector<transaction_graph_ptr_t> &graphs) {
  if (!graphs.empty()) {
    transaction_graph *ret = graphs.begin()->get();
    for (auto it = graphs.begin() + 1; it != graphs.end(); it++) {
      transaction_graph *ptr = it->get();
      ret = merge_two_graphs(ret, ptr);
    }
    return ret;
  }
  return nullptr;
}

void graph_algo::merge_topk_edges_with_same_from_and_same_to(
    transaction_graph::internal_graph_t &graph, uint32_t k) {

  transaction_graph::viterator_t vi, vi_end;

  struct edge_st {
    wei_t weight;
    transaction_graph::edge_descriptor_t edescriptor;
  };

  auto cmp = [](const edge_st &e1, const edge_st &e2) -> bool {
    return e1.weight > e2.weight;
  };

  typedef std::priority_queue<edge_st, std::vector<edge_st>, decltype(cmp)>
      pq_t;

  for (boost::tie(vi, vi_end) = boost::vertices(graph); vi != vi_end; vi++) {
    std::unordered_map<transaction_graph::vertex_descriptor_t, wei_t>
        target_and_vals;
    std::unordered_map<transaction_graph::vertex_descriptor_t, pq_t>
        target_and_minheap;

    transaction_graph::oeiterator_t oei, oei_end;
    for (boost::tie(oei, oei_end) = boost::out_edges(*vi, graph);
         oei != oei_end; oei++) {
      auto target = boost::target(*oei, graph);
      wei_t val = boost::get(boost::edge_weight_t(), graph, *oei);
      if (target_and_vals.find(target) == target_and_vals.end()) {
        target_and_vals.insert(std::make_pair(target, val));
        pq_t min_heap(cmp);
        min_heap.push(edge_st{val, *oei});
        target_and_minheap.insert(std::make_pair(target, min_heap));
      } else {
        pq_t &min_heap = target_and_minheap.find(target)->second;
        if (min_heap.size() < k) {
          min_heap.push(edge_st{val, *oei});
          target_and_vals[target] += val;
        } else {
          edge_st e = min_heap.top();
          if (val > e.weight) {
            // boost::remove_edge(e.edescriptor, graph);
            min_heap.pop();
            target_and_vals[target] -= e.weight;
            min_heap.push(edge_st{val, *oei});
            target_and_vals[target] += val;
          } else {
            // boost::remove_edge(oei, graph);
          }
        }
      }
    }

    bool removed_all_edges = false;
    while (!removed_all_edges) {
      removed_all_edges = true;
      for (boost::tie(oei, oei_end) = boost::out_edges(*vi, graph);
           oei != oei_end; oei++) {
        removed_all_edges = false;
        boost::remove_edge(oei, graph);
        break;
      }
    }

    for (auto it = target_and_vals.begin(); it != target_and_vals.end(); it++) {
      boost::add_edge(*vi, it->first, {it->second, 0}, graph);
    }
  }
}

std::unordered_map<address_t, in_out_val_t>
graph_algo::get_in_out_vals(const transaction_graph::internal_graph_t &graph) {

  std::unordered_map<address_t, in_out_val_t> ret;

  transaction_graph::viterator_t vi, vi_end;

  for (boost::tie(vi, vi_end) = boost::vertices(graph); vi != vi_end; vi++) {
    transaction_graph::ieiterator_t iei, iei_end;
    wei_t in_val = 0;
    for (boost::tie(iei, iei_end) = boost::in_edges(*vi, graph); iei != iei_end;
         iei++) {
      wei_t val = boost::get(boost::edge_weight_t(), graph, *iei);
      in_val += val;
    }

    transaction_graph::oeiterator_t oei, oei_end;
    wei_t out_val = 0;
    for (boost::tie(oei, oei_end) = boost::out_edges(*vi, graph);
         oei != oei_end; oei++) {
      wei_t val = boost::get(boost::edge_weight_t(), graph, *oei);
      out_val += val;
    }

    auto addr = to_address(boost::get(boost::vertex_name_t(), graph, *vi));
    ret.insert(std::make_pair(addr, in_out_val_t{in_val, out_val}));
  }
  return ret;
}

} // namespace rt
} // namespace neb
