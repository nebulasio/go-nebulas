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
#include "common/math.h"
#include "util/chrono.h"
#include <stack>

namespace neb {
namespace rt {

void graph_algo::dfs_find_a_cycle_from_vertex_based_on_time_sequence(
    const transaction_graph::internal_graph_t &graph,
    const transaction_graph::vertex_descriptor_t &s,
    const transaction_graph::vertex_descriptor_t &v,
    const std::unordered_set<transaction_graph::vertex_descriptor_t> &dead_v,
    bool &has_cycle,
    std::unordered_map<transaction_graph::vertex_descriptor_t, bool> &visited,
    std::vector<transaction_graph::edge_descriptor_t> &edges,
    std::vector<transaction_graph::edge_descriptor_t> &ret) {

  auto in_time_order =
      [&graph, &edges](const transaction_graph::oeiterator_t &oei) -> bool {
    if (!edges.empty()) {
      int64_t ts = boost::get(boost::edge_timestamp_t(), graph, edges.back());
      int64_t ts_next = boost::get(boost::edge_timestamp_t(), graph, *oei);
      if (ts >= ts_next) {
        return false;
      }
    }
    return true;
  };
  auto exit_cond = [&has_cycle, &edges,
                    &ret](const transaction_graph::oeiterator_t &oei) {
    edges.push_back(*oei);
    has_cycle = true;

    for (auto it = edges.begin(); it != edges.end(); it++) {
      ret.push_back(*it);
    }
  };

  transaction_graph::oeiterator_t oei, oei_end;
  for (boost::tie(oei, oei_end) = boost::out_edges(v, graph); oei != oei_end;
       oei++) {
    if (has_cycle) {
      return;
    }
    auto target = boost::target(*oei, graph);
    if (dead_v.find(target) != dead_v.end()) {
      continue;
    }

    if (target == s) {
      if (!in_time_order(oei)) {
        continue;
      }
      exit_cond(oei);
      return;
    }

    if (visited[target]) {
      if (!in_time_order(oei)) {
        continue;
      }

      for (auto it = edges.begin(); it != edges.end();) {
        auto t = boost::target(*it, graph);
        it = edges.erase(it);
        if (t == target) {
          break;
        }
      }

      exit_cond(oei);
      return;
    }

    visited[target] = true;

    if (edges.empty() || (!edges.empty() && in_time_order(oei))) {
      edges.push_back(*oei);
      dfs_find_a_cycle_from_vertex_based_on_time_sequence(
          graph, s, target, dead_v, has_cycle, visited, edges, ret);
      edges.pop_back();
    }

    visited[target] = false;
  }
  return;
}

std::vector<transaction_graph::edge_descriptor_t>
graph_algo::find_a_cycle_from_vertex_based_on_time_sequence(
    const transaction_graph::internal_graph_t &graph,
    const transaction_graph::vertex_descriptor_t &v,
    const std::unordered_set<transaction_graph::vertex_descriptor_t> &dead_v) {

  std::vector<transaction_graph::edge_descriptor_t> ret;
  std::vector<transaction_graph::edge_descriptor_t> edges;
  std::unordered_map<transaction_graph::vertex_descriptor_t, bool> visited;
  bool has_cycle = false;

  visited[v] = true;
  dfs_find_a_cycle_from_vertex_based_on_time_sequence(
      graph, v, v, dead_v, has_cycle, visited, edges, ret);
  return ret;
}

std::vector<transaction_graph::edge_descriptor_t>
graph_algo::find_a_cycle_based_on_time_sequence(
    const transaction_graph::internal_graph_t &graph,
    const std::unordered_set<transaction_graph::vertex_descriptor_t> &dead_v) {
  std::vector<transaction_graph::edge_descriptor_t> ret;

  std::vector<transaction_graph::vertex_descriptor_t> to_visit;
  transaction_graph::viterator_t vi, vi_end;
  for (boost::tie(vi, vi_end) = boost::vertices(graph); vi != vi_end; vi++) {
    if (dead_v.find(*vi) == dead_v.end()) {
      to_visit.push_back(*vi);
    }
  }

  for (auto it = to_visit.begin(); it != to_visit.end(); it++) {
    auto ret =
        find_a_cycle_from_vertex_based_on_time_sequence(graph, *it, dead_v);
    if (!ret.empty()) {
      return ret;
    }
  }
  return ret;
}

void graph_algo::bfs_decrease_graph_edges(
    const transaction_graph::internal_graph_t &graph,
    const std::unordered_set<transaction_graph::vertex_descriptor_t> &dead_v,
    std::unordered_set<transaction_graph::vertex_descriptor_t> &tmp_dead,
    std::unordered_map<transaction_graph::vertex_descriptor_t, size_t> &dead_to,
    std::unordered_map<transaction_graph::vertex_descriptor_t, size_t>
        &to_dead) {

  std::queue<transaction_graph::vertex_descriptor_t> q;

  auto update_dead_to = [&graph, &dead_v, &tmp_dead, &dead_to](
                            const transaction_graph::vertex_descriptor_t &v) {
    transaction_graph::oeiterator_t oei, oei_end;
    for (boost::tie(oei, oei_end) = boost::out_edges(v, graph); oei != oei_end;
         oei++) {
      auto target = boost::target(*oei, graph);
      if (dead_v.find(target) == dead_v.end() &&
          tmp_dead.find(target) == tmp_dead.end()) {
        if (dead_to.find(target) != dead_to.end()) {
          dead_to[target]++;
        } else {
          dead_to.insert(std::make_pair(target, 1));
        }
      }
    }
  };
  auto update_to_dead = [&graph, &dead_v, &tmp_dead, &to_dead](
                            const transaction_graph::vertex_descriptor_t &v) {
    transaction_graph::ieiterator_t iei, iei_end;
    for (boost::tie(iei, iei_end) = boost::in_edges(v, graph); iei != iei_end;
         iei++) {
      auto source = boost::source(*iei, graph);
      if (dead_v.find(source) == dead_v.end() &&
          tmp_dead.find(source) == tmp_dead.end()) {
        if (to_dead.find(source) != to_dead.end()) {
          to_dead[source]++;
        } else {
          to_dead.insert(std::make_pair(source, 1));
        }
      }
    }
  };

  for (auto &v : tmp_dead) {
    q.push(v);
    update_dead_to(v);
    update_to_dead(v);
  }

  while (!q.empty()) {
    auto &v = q.front();
    q.pop();

    transaction_graph::oeiterator_t oei, oei_end;
    for (boost::tie(oei, oei_end) = boost::out_edges(v, graph); oei != oei_end;
         oei++) {
      auto target = boost::target(*oei, graph);

      if (dead_v.find(target) == dead_v.end() &&
          tmp_dead.find(target) == tmp_dead.end()) {
        auto ret = boost::in_degree(target, graph);
        if (ret && dead_to.find(target) != dead_to.end() &&
            ret == dead_to[target]) {
          q.push(target);
          tmp_dead.insert(target);
          update_dead_to(target);
        }
      }
    }

    transaction_graph::ieiterator_t iei, iei_end;
    for (boost::tie(iei, iei_end) = boost::in_edges(v, graph); iei != iei_end;
         iei++) {
      auto source = boost::source(*iei, graph);

      if (dead_v.find(source) == dead_v.end() &&
          tmp_dead.find(source) == tmp_dead.end()) {
        auto ret = boost::out_degree(source, graph);
        if (ret && to_dead.find(source) != to_dead.end() &&
            ret == to_dead[source]) {
          q.push(source);
          tmp_dead.insert(source);
          update_to_dead(source);
        }
      }
    }
  }
}

bool graph_algo::decrease_graph_edges(
    const transaction_graph::internal_graph_t &graph,
    std::unordered_set<transaction_graph::vertex_descriptor_t> &dead_v,
    std::unordered_map<transaction_graph::vertex_descriptor_t, size_t> &dead_to,
    std::unordered_map<transaction_graph::vertex_descriptor_t, size_t>
        &to_dead) {

  std::unordered_set<transaction_graph::vertex_descriptor_t> tmp_dead;
  transaction_graph::viterator_t vi, vi_end;
  for (boost::tie(vi, vi_end) = boost::vertices(graph); vi != vi_end; vi++) {
    if (dead_v.find(*vi) == dead_v.end()) {
      auto ins = boost::in_degree(*vi, graph);
      auto outs = boost::out_degree(*vi, graph);
      if (!ins || !outs) {
        tmp_dead.insert(*vi);
      }
    }
  }
  bfs_decrease_graph_edges(graph, dead_v, tmp_dead, dead_to, to_dead);
  for (auto &tmp : tmp_dead) {
    dead_v.insert(tmp);
  }
  return boost::num_vertices(graph) != dead_v.size();
}

void graph_algo::remove_a_cycle(
    transaction_graph::internal_graph_t &graph,
    const std::vector<transaction_graph::edge_descriptor_t> &edges) {

  wei_t min_w = -1;
  for (auto it = edges.begin(); it != edges.end(); it++) {
    wei_t w = boost::get(boost::edge_weight_t(), graph, *it);
    min_w = (min_w == -1 ? w : math::min(min_w, w));
  }

  for (auto it = edges.begin(); it != edges.end(); it++) {
    wei_t w = boost::get(boost::edge_weight_t(), graph, *it);
    boost::put(boost::edge_weight_t(), graph, *it, w - min_w);
    if (w == min_w) {
      boost::remove_edge(*it, graph);
    }
  }
}

void graph_algo::remove_cycles_based_on_time_sequence(
    transaction_graph::internal_graph_t &graph) {

  std::vector<transaction_graph::edge_descriptor_t> ret;
  std::unordered_set<transaction_graph::vertex_descriptor_t> dead_v;
  std::unordered_map<transaction_graph::vertex_descriptor_t, size_t> dead_to;
  std::unordered_map<transaction_graph::vertex_descriptor_t, size_t> to_dead;

  while (true) {
    if (!decrease_graph_edges(graph, dead_v, dead_to, to_dead)) {
      break;
    }
    ret = find_a_cycle_based_on_time_sequence(graph, dead_v);
    if (ret.empty()) {
      break;
    }
    remove_a_cycle(graph, ret);
  }
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

std::unique_ptr<std::unordered_map<address_t, in_out_val_t>>
graph_algo::get_in_out_vals(const transaction_graph::internal_graph_t &graph) {

  auto ret = std::make_unique<std::unordered_map<address_t, in_out_val_t>>();

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
    ret->insert(std::make_pair(addr, in_out_val_t{in_val, out_val}));
  }
  return ret;
}

std::unique_ptr<std::unordered_map<address_t, wei_t>>
graph_algo::get_stakes(const transaction_graph::internal_graph_t &graph) {

  auto ret = std::make_unique<std::unordered_map<address_t, wei_t>>();

  auto it_in_out_vals = get_in_out_vals(graph);
  auto in_out_vals = *it_in_out_vals;
  for (auto it = in_out_vals.begin(); it != in_out_vals.end(); it++) {
    ret->insert(
        std::make_pair(it->first, it->second.m_in_val - it->second.m_out_val));
  }
  return ret;
}

std::unique_ptr<std::unordered_map<address_t, in_out_degree_t>>
graph_algo::get_in_out_degrees(
    const transaction_graph::internal_graph_t &graph) {

  auto ret = std::make_unique<std::unordered_map<address_t, in_out_degree_t>>();

  transaction_graph::viterator_t vi, vi_end;

  for (boost::tie(vi, vi_end) = boost::vertices(graph); vi != vi_end; vi++) {
    transaction_graph::ieiterator_t iei, iei_end;
    uint32_t in_degree = 0;
    for (boost::tie(iei, iei_end) = boost::in_edges(*vi, graph); iei != iei_end;
         iei++) {
      in_degree++;
    }

    transaction_graph::oeiterator_t oei, oei_end;
    uint32_t out_degree = 0;
    for (boost::tie(oei, oei_end) = boost::out_edges(*vi, graph);
         oei != oei_end; oei++) {
      out_degree++;
    }

    auto addr = to_address(boost::get(boost::vertex_name_t(), graph, *vi));
    ret->insert(std::make_pair(addr, in_out_degree_t{in_degree, out_degree}));
  }
  return ret;
}

std::unique_ptr<std::unordered_map<address_t, uint32_t>>
graph_algo::get_degree_sum(const transaction_graph::internal_graph_t &graph) {

  auto ret = std::make_unique<std::unordered_map<address_t, uint32_t>>();

  auto it_in_out_degrees = get_in_out_degrees(graph);
  auto in_out_degrees = *it_in_out_degrees;
  for (auto it = in_out_degrees.begin(); it != in_out_degrees.end(); it++) {
    ret->insert(std::make_pair(it->first, it->second.m_in_degree +
                                              it->second.m_out_degree));
  }
  return ret;
}

namespace opt {

std::vector<transaction_graph::edge_descriptor_t>
graph_algo::find_a_cycle_from_vertex_based_on_time_sequence(
    const transaction_graph::internal_graph_t &graph,
    const std::unordered_map<transaction_graph::vertex_descriptor_t,
                             std::vector<transaction_graph::edge_descriptor_t>>
        &adj,
    const transaction_graph::vertex_descriptor_t &v,
    const std::unordered_set<transaction_graph::vertex_descriptor_t> &dead_v,
    const std::unordered_map<transaction_graph::vertex_descriptor_t, size_t>
        &dead_to,
    const std::unordered_map<transaction_graph::vertex_descriptor_t, size_t>
        &to_dead) {

  std::vector<transaction_graph::edge_descriptor_t> edges;
  typedef std::pair<transaction_graph::vertex_descriptor_t, size_t> st_t;
  std::stack<st_t> st;
  std::unordered_map<transaction_graph::vertex_descriptor_t, bool> visited;
  st.push(std::make_pair(v, 0));
  visited[v] = true;

  auto backtrace_cond = [&adj, &to_dead](const st_t &ele) {
    size_t to_dead_cnt = 0;
    auto it = to_dead.find(ele.first);
    if (it != to_dead.end()) {
      to_dead_cnt = it->second;
    }
    auto ite = adj.find(ele.first);
    if (ite == adj.end()) {
      return true;
    }
    return ele.second + to_dead_cnt == ite->second.size();
  };
  auto in_time_order = [&graph, &edges](
                           const transaction_graph::edge_descriptor_t &ed) {
    if (!edges.empty()) {
      int64_t ts = boost::get(boost::edge_timestamp_t(), graph, edges.back());
      int64_t ts_next = boost::get(boost::edge_timestamp_t(), graph, ed);
      if (ts >= ts_next) {
        return false;
      }
    }
    return true;
  };
  auto erase_tail =
      [&edges, &graph](const transaction_graph::vertex_descriptor_t &target) {
        for (auto it = edges.begin(); it != edges.end();) {
          auto t = boost::target(*it, graph);
          it = edges.erase(it);
          if (t == target) {
            break;
          }
        }
      };

  while (!st.empty()) {
    auto &ele = st.top();
    if (backtrace_cond(ele)) {
      st.pop();
      visited[ele.first] = false;
      if (!edges.empty()) {
        edges.pop_back();
      }
    } else {
      auto it = adj.find(ele.first);
      auto &tmp = it->second;
      auto &nxt = tmp[ele.second++];
      if (in_time_order(nxt)) {
        auto target = boost::target(nxt, graph);
        if (dead_v.find(target) == dead_v.end()) {
          if (!visited[target]) {
            st.push(std::make_pair(target, 0));
            visited[target] = true;
            edges.push_back(nxt);
          } else {
            if (!edges.empty()) {
              auto &e = edges.front();
              auto source = boost::source(e, graph);
              if (source != target) {
                erase_tail(target);
              }
            }
            edges.push_back(nxt);
            break;
          }
        }
      }
    }
  }
  return edges;
}

std::vector<transaction_graph::edge_descriptor_t>
graph_algo::find_a_cycle_based_on_time_sequence(
    const transaction_graph::internal_graph_t &graph,
    const std::unordered_map<transaction_graph::vertex_descriptor_t,
                             std::vector<transaction_graph::edge_descriptor_t>>
        &adj,
    const std::unordered_set<transaction_graph::vertex_descriptor_t> &dead_v,
    const std::unordered_map<transaction_graph::vertex_descriptor_t, size_t>
        &dead_to,
    const std::unordered_map<transaction_graph::vertex_descriptor_t, size_t>
        &to_dead) {

  std::vector<transaction_graph::edge_descriptor_t> ret;
  for (auto &ele : adj) {
    if (dead_v.find(ele.first) == dead_v.end()) {
      ret = find_a_cycle_from_vertex_based_on_time_sequence(
          graph, adj, ele.first, dead_v, dead_to, to_dead);
      if (!ret.empty()) {
        break;
      }
    }
  }
  return ret;
}

void graph_algo::build_adj_graph(
    const transaction_graph::internal_graph_t &graph,
    std::unordered_map<transaction_graph::vertex_descriptor_t,
                       std::vector<transaction_graph::edge_descriptor_t>>
        &adj) {
  transaction_graph::viterator_t vi, vi_end;
  for (boost::tie(vi, vi_end) = boost::vertices(graph); vi != vi_end; vi++) {
    transaction_graph::oeiterator_t oei, oei_end;
    for (boost::tie(oei, oei_end) = boost::out_edges(*vi, graph);
         oei != oei_end; oei++) {
      if (adj.find(*vi) != adj.end()) {
        auto &tmp = adj.find(*vi)->second;
        tmp.push_back(*oei);
      } else {
        std::vector<transaction_graph::edge_descriptor_t> tmp;
        tmp.push_back(*oei);
        adj.insert(std::make_pair(*vi, tmp));
      }
    }
  }
}

void graph_algo::remove_cycles_based_on_time_sequence(
    transaction_graph::internal_graph_t &graph) {

  std::unordered_map<transaction_graph::vertex_descriptor_t,
                     std::vector<transaction_graph::edge_descriptor_t>>
      adj;
  opt::graph_algo::build_adj_graph(graph, adj);

  std::unordered_set<transaction_graph::vertex_descriptor_t> dead_v;
  std::unordered_map<transaction_graph::vertex_descriptor_t, size_t> dead_to;
  std::unordered_map<transaction_graph::vertex_descriptor_t, size_t> to_dead;

  while (true) {
    if (!::neb::rt::graph_algo::decrease_graph_edges(graph, dead_v, dead_to,
                                                     to_dead)) {
      break;
    }
    auto ret = find_a_cycle_based_on_time_sequence(graph, adj, dead_v, dead_to,
                                                   to_dead);
    if (ret.empty()) {
      break;
    }
    remove_a_cycle(graph, adj, ret);
  }
}

void graph_algo::remove_a_cycle(
    transaction_graph::internal_graph_t &graph,
    std::unordered_map<transaction_graph::vertex_descriptor_t,
                       std::vector<transaction_graph::edge_descriptor_t>> &adj,
    const std::vector<transaction_graph::edge_descriptor_t> &edges) {

  wei_t min_w = -1;
  for (auto &e : edges) {
    wei_t w = boost::get(boost::edge_weight_t(), graph, e);
    min_w = (min_w == -1 ? w : math::min(min_w, w));
  }

  for (auto &e : edges) {
    wei_t w = boost::get(boost::edge_weight_t(), graph, e);
    boost::put(boost::edge_weight_t(), graph, e, w - min_w);
    if (w == min_w) {
      boost::remove_edge(e, graph);

      auto source = boost::source(e, graph);
      if (adj.find(source) != adj.end()) {
        for (auto it_e = adj[source].begin(); it_e != adj[source].end();) {
          if (*it_e == e) {
            it_e = adj[source].erase(it_e);
          } else {
            it_e++;
          }
        }
      }
    }
  }
}
}

class non_recursive_remove_cycles_based_on_time_sequence_helper {
public:
  typedef transaction_graph::internal_graph_t graph_t;
  non_recursive_remove_cycles_based_on_time_sequence_helper(
      transaction_graph &g)
      : m_graph(g.internal_graph()), m_sorted_vs(g.vertex_num(), 0),
        m_dead_vs(g.vertex_num(), 0), m_dead_es(g.edge_num(), 0),
        m_sorted_ves(g.vertex_num(),
                     std::vector<transaction_graph::edge_descriptor_t>()) {
    m_path_vs_has_cycle = false;
  }

  inline void add_dead_vertex(const transaction_graph::vertex_descriptor_t &v) {
    m_dead_vs[v] = true;
  }
  inline bool is_dead_vertex(const transaction_graph::vertex_descriptor_t &v) {
    return m_dead_vs[v];
  }
  inline void add_dead_edge(const transaction_graph::edge_descriptor_t &e) {
    m_dead_es[boost::get(boost::edge_sort_id_t(), m_graph, e)] = true;
  }
  inline bool is_dead_edge(const transaction_graph::edge_descriptor_t &e) {
    return m_dead_es[boost::get(boost::edge_sort_id_t(), m_graph, e)];
  }

  std::vector<transaction_graph::vertex_descriptor_t>
  possible_start_nodes_of_cycles() {
    std::vector<transaction_graph::vertex_descriptor_t> nodes;
    transaction_graph::viterator_t vi, vi_end;
    for (boost::tie(vi, vi_end) = boost::vertices(m_graph); vi != vi_end;
         vi++) {
      auto ins = boost::in_degree(*vi, m_graph);
      auto outs = boost::in_degree(*vi, m_graph);
      if (!ins || !outs) {
        add_dead_vertex(*vi);
      } else {
        int64_t ts_in_max = std::numeric_limits<int64_t>::min();
        int64_t ts_out_min = std::numeric_limits<int64_t>::max();

        transaction_graph::ieiterator_t iei, iei_end;
        for (boost::tie(iei, iei_end) = boost::in_edges(*vi, m_graph);
             iei != iei_end; ++iei) {
          int64_t ts = boost::get(boost::edge_timestamp_t(), m_graph, *iei);
          if (ts > ts_in_max)
            ts_in_max = ts;
        }
        transaction_graph::oeiterator_t oei, oei_end;
        for (boost::tie(oei, oei_end) = boost::out_edges(*vi, m_graph);
             oei != oei_end; ++oei) {
          int64_t ts = boost::get(boost::edge_timestamp_t(), m_graph, *oei);
          if (ts < ts_out_min)
            ts_out_min = ts;
        }
        if (ts_in_max >= ts_out_min) {
          nodes.push_back(*vi);
        }
      }
    }
    return nodes;
}

std::pair<bool, transaction_graph::edge_descriptor_t>
larger_min_out_edge(const transaction_graph::vertex_descriptor_t &v,
                    int64_t larger_than) {
  transaction_graph::edge_descriptor_t edge;
  transaction_graph::oeiterator_t oei, oei_end;
  std::vector<transaction_graph::edge_descriptor_t> &es = m_sorted_ves[v];
  auto it = std::lower_bound(
      es.begin(), es.end(), larger_than,
      [this](transaction_graph::edge_descriptor_t e1, int64_t m2) {
        auto m1 = boost::get(boost::edge_timestamp_t(), m_graph, e1);
        return m1 < m2;
      });
  return std::make_pair(it != es.end(), *it);
  // LOG(INFO) << "v:" << v << ", es size: " << es.size();
  int s = 0;
  int e = es.size() - 1;
  while (s < e) {
    int mid = (s + e) / 2;
    auto ms = boost::get(boost::edge_timestamp_t(), m_graph, es[mid]);
    if (ms < larger_than) {
      s = mid + 1;
    } else if (ms > larger_than) {
      e = mid - 1;
    } else {
      int ret = mid;
      while (mid >= 0 && ms == larger_than) {
        ret = mid;
        mid--;
        if (mid >= 0) {
          ms = boost::get(boost::edge_timestamp_t(), m_graph, es[mid]);
        }
      }
      return std::make_pair(true, es[ret]);
    }
    // LOG(INFO) << "v:" << v << ",s: " << s << ", e: " << e;
  }
  if (s > e && s < es.size()) {
    return std::make_pair(true, es[s]);
  }
  // LOG(INFO) << "larger_min_out_edge: " << es[s];
  return std::make_pair(true, es[s]);
  // return std::make_pair(found, edge);
}

std::pair<bool, transaction_graph::edge_descriptor_t>
min_out_edge(const transaction_graph::vertex_descriptor_t &v) {
  sort_out_edges_if_not(v);
  if (m_sorted_ves.empty()) {
    return std::make_pair(false, transaction_graph::edge_descriptor_t());
  }
  return std::make_pair(true, m_sorted_ves[v][0]);
}

std::vector<transaction_graph::vertex_descriptor_t>
ascend_sort_vertex_by_min_out_edge(
    const std::vector<transaction_graph::vertex_descriptor_t> &vs) {

  std::vector<transaction_graph::vertex_descriptor_t> ret = vs;
  std::sort(ret.begin(), ret.end(),
            [this](transaction_graph::vertex_descriptor_t v1,
                   transaction_graph::vertex_descriptor_t v2) {
              auto t1 = boost::get(boost::edge_timestamp_t(), m_graph,
                                   min_out_edge(v1).second);
              auto t2 = boost::get(boost::edge_timestamp_t(), m_graph,
                                   min_out_edge(v2).second);
              return t1 < t2;
            });
  return ret;
}
bool path_is_cycle() {
  return m_path_vs_has_cycle;
}

bool check_and_update_dead_edge(
    const transaction_graph::edge_descriptor_t &edge) {

  if (is_dead_edge(edge))
    return true;

  auto target = boost::target(edge, m_graph);
  auto source = boost::source(edge, m_graph);
  if (is_dead_vertex(source))
    return true;

  if (is_dead_vertex(target)) {
    add_dead_edge(edge);

    // if all the edges from soure are dead, the the source is a dead node.

    transaction_graph::oeiterator_t oei, oei_end;
    bool found_one_live_edges = false;
    for (boost::tie(oei, oei_end) = boost::out_edges(source, m_graph);
         oei != oei_end; ++oei) {
      if (!is_dead_edge(*oei)) {
        found_one_live_edges = true;
        break;
      }
    }
    if (!found_one_live_edges) {
      add_dead_vertex(source);
    }
    return true;
  }
  return false;
}

std::pair<bool, transaction_graph::edge_descriptor_t>
get_next_edge_with_the_same_source(
    const transaction_graph::edge_descriptor_t &e) {
  auto source = boost::source(e, m_graph);
  auto check_id = boost::get(boost::edge_check_id_t(), m_graph, e);
  transaction_graph::edge_descriptor_t ret;
  transaction_graph::oeiterator_t oei, oei_end;
  bool found = false;
  if (check_id >= m_sorted_ves[source].size() - 1) {
    return std::make_pair(found, ret);
  } else {
    return std::make_pair(true, m_sorted_ves[source][check_id + 1]);
  }
}

void remove_a_cycle(
    const std::vector<transaction_graph::edge_descriptor_t> &edges) {

  // if (edges.empty())
  // return;
  std::vector<transaction_graph::edge_descriptor_t> cycle;
  bool found = false;
  bool target = boost::target(edges.back(), m_graph);
  bool source = boost::source(edges.front(), m_graph);
  if (target == source) {
    cycle = edges;
  } else {
    for (size_t i = 0; i < edges.size(); ++i) {
      auto edge = edges[i];
      // LOG(INFO) << "edge: " << edge;
      if (boost::source(edge, m_graph) == target) {
        found = true;
      }
      if (found) {
        cycle.push_back(edge);
      }
    }
  }
  wei_t min_w = std::numeric_limits<wei_t>::max();
  for (auto it = cycle.begin(); it != cycle.end(); it++) {
    wei_t w = boost::get(boost::edge_weight_t(), m_graph, *it);
    min_w = std::min(w, min_w);
  }

  // LOG(INFO) << "cycle size: " << cycle.size();
  for (auto it = cycle.begin(); it != cycle.end(); it++) {
    // LOG(INFO) << "to change edge: " << *it;
    wei_t w = boost::get(boost::edge_weight_t(), m_graph, *it);
    boost::put(boost::edge_weight_t(), m_graph, *it, w - min_w);
    if (w == min_w) {
      boost::remove_edge(*it, m_graph);
      // LOG(INFO) << "to remove " << *it;
      auto source = boost::source(*it, m_graph);
      auto target = boost::target(*it, m_graph);
      // m_sorted_vs[source] = false;
      // sort_out_edges_if_not(source);
      auto &tes = m_sorted_ves[source];
      auto tid = boost::get(boost::edge_sort_id_t(), m_graph, *it);
      auto rit = std::remove_if(
          tes.begin(), tes.end(),
          [this, tid](transaction_graph::edge_descriptor_t t) {
            return boost::get(boost::edge_sort_id_t(), m_graph, t) == tid;
          });
      std::for_each(rit, tes.end(),
                    [this](transaction_graph::edge_descriptor_t &t) {
                      auto v = boost::get(boost::edge_check_id_t(), m_graph, t);
                      boost::put(boost::edge_check_id_t(), m_graph, t, v - 1);
                    });

      // We may recursively mark dead nodes, yet cyccles should be rare.
      // Thus, may delay this when we meet them again.
      if (boost::out_degree(source, m_graph) == 0) {
        add_dead_vertex(source);
      }
      if (boost::in_degree(target, m_graph) == 0) {
        add_dead_vertex(target);
      }
    }
  }
}

void path_add_vertex(const transaction_graph::vertex_descriptor_t &v) {
  if (m_path_vs.find(v) != m_path_vs.end()) {
    m_path_vs_has_cycle = true;
  } else {
    m_path_vs.insert(v);
  }
}
void path_remove_vertex(const transaction_graph::vertex_descriptor_t &v) {
  m_path_vs.erase(v);
}
void sort_out_edges_if_not(const transaction_graph::vertex_descriptor_t &v) {
  if (m_sorted_vs[v])
    return;

  m_sorted_vs[v] = true;
  std::vector<transaction_graph::edge_descriptor_t> es;
  transaction_graph::oeiterator_t oei, oei_end;
  for (boost::tie(oei, oei_end) = boost::out_edges(v, m_graph); oei != oei_end;
       oei++) {
    es.push_back(*oei);
  }
  std::sort(es.begin(), es.end(),
            [this](transaction_graph::edge_descriptor_t e1,
                   transaction_graph::edge_descriptor_t e2) {
              auto t1 = boost::get(boost::edge_timestamp_t(), m_graph, e1);
              auto t2 = boost::get(boost::edge_timestamp_t(), m_graph, e2);
              return t1 < t2;
            });
  for (size_t i = 0; i < es.size(); ++i) {
    boost::put(boost::edge_check_id_t(), m_graph, es[i], i);
  }
  // LOG(INFO) << "sorted for v: " << v << ", edges: " << es.size();
  m_sorted_ves[v] = es;
}
void remove_cycles_based_on_time_sequence() {
  std::vector<transaction_graph::vertex_descriptor_t> to_visit =
      possible_start_nodes_of_cycles();

  to_visit = ascend_sort_vertex_by_min_out_edge(to_visit);

  std::vector<transaction_graph::edge_descriptor_t> to_visit_path;
  std::vector<transaction_graph::edge_descriptor_t> cur_path;
  while (!to_visit.empty()) {
    transaction_graph::vertex_descriptor_t v = to_visit.back();
    to_visit.pop_back();
    if (is_dead_vertex(v))
      continue;
    auto ins = boost::in_degree(v, m_graph);
    auto outs = boost::in_degree(v, m_graph);
    if (!ins || !outs) {
      add_dead_vertex(v);
      continue;
    }
    sort_out_edges_if_not(v);

    // LOG(INFO) << "for v: " << v;
    auto ts_min = std::numeric_limits<int64_t>::min();
    auto ts_in_max =
        boost::get(boost::edge_timestamp_t(), m_graph, m_sorted_ves[v].back());
    while (true) {
      to_visit_path.clear();
      cur_path.clear();
      m_path_vs.clear();
      auto e = larger_min_out_edge(v, ts_min);
      if (!e.first) {
        // LOG(INFO) << "got break";
        break;
      }
      ts_min = boost::get(boost::edge_timestamp_t(), m_graph, e.second);
      if (ts_min > ts_in_max)
        break;
      to_visit_path.push_back(e.second);
      path_add_vertex(boost::source(e.second, m_graph));
      while (!path_is_cycle() && !to_visit_path.empty()) {
        if (!to_visit_path.empty()) {
          auto edge = to_visit_path.back();
          to_visit_path.pop_back();
          if (check_and_update_dead_edge(edge)) {
            continue;
          }

          // LOG(INFO) << "add edge: " << edge;
          // path_add_vertex(boost::target(edge, m_graph));
          sort_out_edges_if_not(boost::source(edge, m_graph));
          auto new_edge = std::make_pair(true, edge);

          while (new_edge.first && !path_is_cycle()) {
            cur_path.push_back(new_edge.second);
            // LOG(INFO) << "new edge: " << new_edge.second;
            path_add_vertex(boost::target(new_edge.second, m_graph));
            sort_out_edges_if_not(boost::target(new_edge.second, m_graph));
            auto cur_ts =
                boost::get(boost::edge_timestamp_t(), m_graph, new_edge.second);
            new_edge = larger_min_out_edge(
                boost::target(new_edge.second, m_graph), cur_ts);
            // LOG(INFO) << "tnew edge: " << new_edge.second;
          }

          // LOG(INFO) << "quit";
        } else if (!path_is_cycle()) {
          auto edge = cur_path.back();
          cur_path.pop_back();
          path_remove_vertex(boost::target(edge, m_graph));
          if (check_and_update_dead_edge(edge)) {
            continue;
          }
          sort_out_edges_if_not(boost::source(edge, m_graph));
          auto new_edge = get_next_edge_with_the_same_source(edge);
          if (new_edge.first) {
            to_visit_path.push_back(new_edge.second);
          }
        }
      }
      if (path_is_cycle()) {
        // LOG(INFO) << "got cycle";
        remove_a_cycle(cur_path);
        m_path_vs_has_cycle = false;
      } else {
        break;
      }
    }
  }
}

protected:
graph_t &m_graph;
std::vector<bool> m_sorted_vs;
std::vector<bool> m_dead_vs;
std::vector<bool> m_dead_es;
std::unordered_set<transaction_graph::vertex_descriptor_t> m_path_vs;
bool m_path_vs_has_cycle;
std::vector<std::vector<transaction_graph::edge_descriptor_t>> m_sorted_ves;
};

void graph_algo::non_recursive_remove_cycles_based_on_time_sequence(
    transaction_graph &graph) {
  non_recursive_remove_cycles_based_on_time_sequence_helper nh(graph);
  nh.remove_cycles_based_on_time_sequence();
}

} // namespace rt
} // namespace neb
