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

#include "common/common.h"
#include "runtime/nr/graph/algo.h"
#include "runtime/nr/graph/graph.h"
#include <gtest/gtest.h>

TEST(test_runtime_graph, test_add_edge_simple) {
  neb::rt::transaction_graph tg;
  auto s = neb::to_address(std::string("s"));
  auto t = neb::to_address(std::string("t"));
  neb::wei_t val = 1;
  int64_t ts = 1;
  tg.add_edge(s, t, val, ts);

  auto g = tg.internal_graph();
  neb::rt::transaction_graph::viterator_t vi, vi_end;
  boost::tie(vi, vi_end) = boost::vertices(g);
  neb::rt::transaction_graph::oeiterator_t oei, oei_end;
  boost::tie(oei, oei_end) = boost::out_edges(*vi, g);

  auto target = boost::target(*oei, g);
  std::string target_name = boost::get(boost::vertex_name_t(), g, target);
  EXPECT_EQ(neb::to_address(target_name), t);

  auto source = boost::source(*oei, g);
  std::string source_name = boost::get(boost::vertex_name_t(), g, source);
  EXPECT_EQ(neb::to_address(source_name), s);

  auto w = boost::get(boost::edge_weight_t(), g, *oei);
  EXPECT_EQ(w, val);
  auto timestamp = boost::get(boost::edge_timestamp_t(), g, *oei);
  EXPECT_EQ(timestamp, ts);
}

TEST(test_runtime_graph, test_add_edge_self_cycle) {
  neb::rt::transaction_graph tg;
  auto s = neb::to_address(std::string("s"));
  neb::wei_t val = 1;
  int64_t ts = 1;
  tg.add_edge(s, s, val, ts);

  auto g = tg.internal_graph();
  neb::rt::transaction_graph::viterator_t vi, vi_end;
  boost::tie(vi, vi_end) = boost::vertices(g);
  neb::rt::transaction_graph::oeiterator_t oei, oei_end;
  boost::tie(oei, oei_end) = boost::out_edges(*vi, g);

  auto target = boost::target(*oei, g);
  std::string target_name = boost::get(boost::vertex_name_t(), g, target);
  EXPECT_EQ(neb::to_address(target_name), s);

  auto source = boost::source(*oei, g);
  std::string source_name = boost::get(boost::vertex_name_t(), g, source);
  EXPECT_EQ(neb::to_address(source_name), s);

  auto w = boost::get(boost::edge_weight_t(), g, *oei);
  EXPECT_EQ(w, val);
  auto timestamp = boost::get(boost::edge_timestamp_t(), g, *oei);
  EXPECT_EQ(timestamp, ts);
}

TEST(test_runtime_graph, test_add_edge_multi) {
  neb::rt::transaction_graph tg;
  auto s = neb::to_address(std::string("s"));
  auto t = neb::to_address(std::string("t"));

  struct edge_t {
    neb::wei_t m_val;
    int64_t m_ts;
  };
  std::vector<edge_t> v({{1, 2}, {3, 4}, {5, 6}, {7, 8}});
  for (size_t i = 0; i < v.size(); i++) {
    tg.add_edge(s, t, v[i].m_val, v[i].m_ts);
  }
  auto it_edge = v.begin();

  auto g = tg.internal_graph();
  neb::rt::transaction_graph::viterator_t vi, vi_end;

  for (boost::tie(vi, vi_end) = boost::vertices(g); vi != vi_end; vi++) {
    neb::rt::transaction_graph::oeiterator_t oei, oei_end;
    for (boost::tie(oei, oei_end) = boost::out_edges(*vi, g); oei != oei_end;
         oei++) {
      auto source = boost::source(*oei, g);
      auto target = boost::target(*oei, g);
      auto ss = neb::to_address(boost::get(boost::vertex_name_t(), g, source));
      auto tt = neb::to_address(boost::get(boost::vertex_name_t(), g, target));
      neb::wei_t w = boost::get(boost::edge_weight_t(), g, *oei);
      int64_t ts = boost::get(boost::edge_timestamp_t(), g, *oei);
      EXPECT_EQ(ss, s);
      EXPECT_EQ(tt, t);
      EXPECT_EQ(w, it_edge->m_val);
      EXPECT_EQ(ts, it_edge->m_ts);
      it_edge++;
    }
  }
}

struct edge_t {
  neb::address_t m_s;
  neb::address_t m_t;
  neb::wei_t m_val;
  int64_t m_ts;
};

TEST(test_runtime_graph, test_add_edge_random) {
  neb::rt::transaction_graph tg;

  size_t ch_size = 'z' - 'a';
  std::random_device rd;
  std::mt19937 mt(rd());
  std::uniform_int_distribution<> dis(0, ch_size);
  size_t edge_size = dis(mt) * dis(mt);
  edge_size++;

  std::vector<edge_t> v;
  for (size_t i = 0; i < edge_size; i++) {
    char ch_s = 'a' + dis(mt);
    char ch_t = 'a' + dis(mt);
    v.push_back(edge_t{neb::to_address(std::string(1, ch_s)),
                       neb::to_address(std::string(1, ch_t)), dis(mt),
                       dis(mt)});
    tg.add_edge(v[i].m_s, v[i].m_t, v[i].m_val, v[i].m_ts);
  }

  auto g = tg.internal_graph();
  neb::rt::transaction_graph::viterator_t vi, vi_end;

  for (boost::tie(vi, vi_end) = boost::vertices(g); vi != vi_end; vi++) {
    neb::rt::transaction_graph::oeiterator_t oei, oei_end;
    for (boost::tie(oei, oei_end) = boost::out_edges(*vi, g); oei != oei_end;
         oei++) {
      auto source = boost::source(*oei, g);
      auto target = boost::target(*oei, g);
      auto ss = neb::to_address(boost::get(boost::vertex_name_t(), g, source));
      auto tt = neb::to_address(boost::get(boost::vertex_name_t(), g, target));
      neb::wei_t w = boost::get(boost::edge_weight_t(), g, *oei);
      int64_t ts = boost::get(boost::edge_timestamp_t(), g, *oei);

      bool is_found = false;
      for (auto &e : v) {
        if (e.m_s == ss && e.m_t == tt && e.m_val == w && e.m_ts == ts) {
          is_found = true;
          break;
        }
      }
      EXPECT_TRUE(is_found);
    }
  }
}

TEST(test_runtime_graph, test_build_graph_from_internal) {
  neb::rt::transaction_graph tg;

  size_t ch_size = 'z' - 'a';
  std::random_device rd;
  std::mt19937 mt(rd());
  std::uniform_int_distribution<> dis(0, ch_size);
  size_t edge_size = dis(mt) * dis(mt);
  edge_size++;

  std::vector<edge_t> v;
  for (size_t i = 0; i < edge_size; i++) {
    char ch_s = 'a' + dis(mt);
    char ch_t = 'a' + dis(mt);
    v.push_back(edge_t{neb::to_address(std::string(1, ch_s)),
                       neb::to_address(std::string(1, ch_t)), dis(mt),
                       dis(mt)});
    tg.add_edge(v[i].m_s, v[i].m_t, v[i].m_val, v[i].m_ts);
  }

  auto g = tg.internal_graph();
  auto tg_ptr = neb::rt::build_graph_from_internal(g);
  auto gg = tg_ptr->internal_graph();

  neb::rt::transaction_graph::viterator_t vi, vi_end;
  for (boost::tie(vi, vi_end) = boost::vertices(g); vi != vi_end; vi++) {
    neb::rt::transaction_graph::oeiterator_t oei, oei_end;
    for (boost::tie(oei, oei_end) = boost::out_edges(*vi, g); oei != oei_end;
         oei++) {
      auto source = boost::source(*oei, g);
      auto target = boost::target(*oei, g);
      auto ss = neb::to_address(boost::get(boost::vertex_name_t(), g, source));
      auto tt = neb::to_address(boost::get(boost::vertex_name_t(), g, target));
      neb::wei_t w = boost::get(boost::edge_weight_t(), g, *oei);
      int64_t ts = boost::get(boost::edge_timestamp_t(), g, *oei);

      bool is_found = false;
      neb::rt::transaction_graph::viterator_t _vi, _vi_end;
      for (boost::tie(_vi, _vi_end) = boost::vertices(gg); _vi != _vi_end;
           _vi++) {
        neb::rt::transaction_graph::oeiterator_t _oei, _oei_end;
        for (boost::tie(_oei, _oei_end) = boost::out_edges(*_vi, gg);
             _oei != _oei_end; _oei++) {
          auto _source = boost::source(*_oei, gg);
          auto _target = boost::target(*_oei, gg);
          auto _ss =
              neb::to_address(boost::get(boost::vertex_name_t(), gg, _source));
          auto _tt =
              neb::to_address(boost::get(boost::vertex_name_t(), gg, _target));
          neb::wei_t _w = boost::get(boost::edge_weight_t(), gg, *_oei);
          int64_t _ts = boost::get(boost::edge_timestamp_t(), gg, *_oei);
          if (ss == _ss && tt == _tt & w == _w && ts == _ts) {
            is_found = true;
            break;
          }
        }
      }
      EXPECT_TRUE(is_found);
    }
  }
}

TEST(test_algo, test_get_in_out_vals) {
  neb::rt::transaction_graph tg;
  tg.add_edge(neb::to_address("a"), neb::to_address("b"), 1, 1);
  tg.add_edge(neb::to_address("b"), neb::to_address("c"), 2, 2);
  tg.add_edge(neb::to_address("b"), neb::to_address("d"), 3, 3);
  tg.add_edge(neb::to_address("c"), neb::to_address("d"), 4, 4);

  auto graph = tg.internal_graph();

  auto addr_and_in_out_vals_ptr = neb::rt::graph_algo::get_in_out_vals(graph);
  auto addr_and_in_out_vals = *addr_and_in_out_vals_ptr;
  EXPECT_TRUE(addr_and_in_out_vals[neb::to_address("a")].m_in_val == 0);
  EXPECT_TRUE(addr_and_in_out_vals[neb::to_address("a")].m_out_val == 1);

  EXPECT_TRUE(addr_and_in_out_vals[neb::to_address("b")].m_in_val == 1);
  EXPECT_TRUE(addr_and_in_out_vals[neb::to_address("b")].m_out_val == 5);

  EXPECT_TRUE(addr_and_in_out_vals[neb::to_address("c")].m_in_val == 2);
  EXPECT_TRUE(addr_and_in_out_vals[neb::to_address("c")].m_out_val == 4);

  EXPECT_TRUE(addr_and_in_out_vals[neb::to_address("d")].m_in_val == 7);
  EXPECT_TRUE(addr_and_in_out_vals[neb::to_address("d")].m_out_val == 0);
}

TEST(test_algo, test_get_stakes) {
  neb::rt::transaction_graph tg;
  tg.add_edge(neb::to_address("a"), neb::to_address("b"), 4, 4);
  tg.add_edge(neb::to_address("b"), neb::to_address("c"), 3, 3);
  tg.add_edge(neb::to_address("b"), neb::to_address("d"), 2, 2);
  tg.add_edge(neb::to_address("c"), neb::to_address("d"), 1, 1);

  auto graph = tg.internal_graph();
  auto addr_and_stakes_ptr = neb::rt::graph_algo::get_stakes(graph);
  auto addr_and_stakes = *addr_and_stakes_ptr;

  EXPECT_TRUE(addr_and_stakes[neb::to_address("a")] == -4);
  EXPECT_TRUE(addr_and_stakes[neb::to_address("b")] == -1);
  EXPECT_TRUE(addr_and_stakes[neb::to_address("c")] == 2);
  EXPECT_TRUE(addr_and_stakes[neb::to_address("d")] == 3);
}

TEST(test_algo, test_get_in_out_degree) {
  neb::rt::transaction_graph tg;
  tg.add_edge(neb::to_address("a"), neb::to_address("b"), 1, 1);
  tg.add_edge(neb::to_address("a"), neb::to_address("c"), 2, 2);
  tg.add_edge(neb::to_address("b"), neb::to_address("c"), 3, 3);
  tg.add_edge(neb::to_address("b"), neb::to_address("d"), 4, 4);
  tg.add_edge(neb::to_address("c"), neb::to_address("d"), 5, 5);

  auto graph = tg.internal_graph();

  auto addr_and_in_out_degree_ptr =
      neb::rt::graph_algo::get_in_out_degrees(graph);
  auto addr_and_in_out_degree = *addr_and_in_out_degree_ptr;
  EXPECT_TRUE(addr_and_in_out_degree[neb::to_address("a")].m_in_degree == 0);
  EXPECT_TRUE(addr_and_in_out_degree[neb::to_address("a")].m_out_degree == 2);

  EXPECT_TRUE(addr_and_in_out_degree[neb::to_address("b")].m_in_degree == 1);
  EXPECT_TRUE(addr_and_in_out_degree[neb::to_address("b")].m_out_degree == 2);

  EXPECT_TRUE(addr_and_in_out_degree[neb::to_address("c")].m_in_degree == 2);
  EXPECT_TRUE(addr_and_in_out_degree[neb::to_address("c")].m_out_degree == 1);

  EXPECT_TRUE(addr_and_in_out_degree[neb::to_address("d")].m_in_degree == 2);
  EXPECT_TRUE(addr_and_in_out_degree[neb::to_address("d")].m_out_degree == 0);
}

TEST(test_algo, test_get_degree_sum) {
  neb::rt::transaction_graph tg;
  tg.add_edge(neb::to_address("a"), neb::to_address("b"), 1, 1);
  tg.add_edge(neb::to_address("a"), neb::to_address("c"), 2, 2);
  tg.add_edge(neb::to_address("b"), neb::to_address("c"), 3, 3);
  tg.add_edge(neb::to_address("b"), neb::to_address("d"), 4, 4);
  tg.add_edge(neb::to_address("c"), neb::to_address("d"), 5, 5);

  auto graph = tg.internal_graph();
  auto addr_and_degrees_ptr = neb::rt::graph_algo::get_degree_sum(graph);
  auto addr_and_degrees = *addr_and_degrees_ptr;

  EXPECT_TRUE(addr_and_degrees[neb::to_address("a")] == 2);
  EXPECT_TRUE(addr_and_degrees[neb::to_address("b")] == 3);
  EXPECT_TRUE(addr_and_degrees[neb::to_address("c")] == 3);
  EXPECT_TRUE(addr_and_degrees[neb::to_address("d")] == 2);
}

TEST(test_algo, test_merge_edges_with_same_from_and_same_to) {
  neb::rt::transaction_graph tg;
  tg.add_edge(neb::to_address("a"), neb::to_address("a"), 1, 1);
  tg.add_edge(neb::to_address("a"), neb::to_address("b"), 2, 2);
  tg.add_edge(neb::to_address("a"), neb::to_address("b"), 3, 3);
  tg.add_edge(neb::to_address("b"), neb::to_address("a"), 4, 4);

  auto graph = tg.internal_graph();
  neb::rt::graph_algo::merge_edges_with_same_from_and_same_to(graph);

  neb::rt::transaction_graph::viterator_t vi, vi_end;

  for (boost::tie(vi, vi_end) = boost::vertices(graph); vi != vi_end; vi++) {
    neb::rt::transaction_graph::oeiterator_t oei, oei_end;
    for (boost::tie(oei, oei_end) = boost::out_edges(*vi, graph);
         oei != oei_end; oei++) {
      auto source = boost::source(*oei, graph);
      auto target = boost::target(*oei, graph);
      std::string source_name =
          boost::get(boost::vertex_name_t(), graph, source);
      std::string target_name =
          boost::get(boost::vertex_name_t(), graph, target);

      neb::wei_t val = boost::get(boost::edge_weight_t(), graph, *oei);
      if (source_name.compare("a") == 0 && target_name.compare("a") == 0) {
        EXPECT_TRUE(val == 1);
      } else if (source_name.compare("a") == 0 &&
                 target_name.compare("b") == 0) {
        EXPECT_TRUE(val == 5);
      } else if (source_name.compare("b") == 0 &&
                 target_name.compare("a") == 0) {
        EXPECT_TRUE(val == 4);
      }
    }
  }
}

TEST(test_algo, test_merge_topk_edges_with_same_from_and_same_to_less_than_k) {
  neb::rt::transaction_graph tg;
  tg.add_edge(neb::to_address("a"), neb::to_address("b"), 1, 1);
  tg.add_edge(neb::to_address("a"), neb::to_address("b"), 2, 2);

  auto graph = tg.internal_graph();
  neb::rt::graph_algo::merge_topk_edges_with_same_from_and_same_to(graph);

  neb::rt::transaction_graph::viterator_t vi, vi_end;

  boost::tie(vi, vi_end) = boost::vertices(graph);
  neb::rt::transaction_graph::oeiterator_t oei, oei_end;
  boost::tie(oei, oei_end) = boost::out_edges(*vi, graph);
  auto source = boost::source(*oei, graph);
  auto target = boost::target(*oei, graph);
  std::string source_name = boost::get(boost::vertex_name_t(), graph, source);
  std::string target_name = boost::get(boost::vertex_name_t(), graph, target);
  EXPECT_EQ(source_name, "a");
  EXPECT_EQ(target_name, "b");

  int64_t timestamp = boost::get(boost::edge_timestamp_t(), graph, *oei);
  EXPECT_EQ(timestamp, 0);

  neb::wei_t val = boost::get(boost::edge_weight_t(), graph, *oei);
  EXPECT_TRUE(val == 3);
}

TEST(test_algo, test_merge_topk_edges_with_same_from_and_same_to_equal_to_k) {
  neb::rt::transaction_graph tg1;
  tg1.add_edge(neb::to_address("a"), neb::to_address("b"), 1, 1);
  tg1.add_edge(neb::to_address("a"), neb::to_address("b"), 2, 2);
  tg1.add_edge(neb::to_address("a"), neb::to_address("b"), 3, 3);

  auto graph = tg1.internal_graph();
  neb::rt::graph_algo::merge_topk_edges_with_same_from_and_same_to(graph);

  neb::rt::transaction_graph::viterator_t vi, vi_end;
  boost::tie(vi, vi_end) = boost::vertices(graph);
  neb::rt::transaction_graph::oeiterator_t oei, oei_end;
  boost::tie(oei, oei_end) = boost::out_edges(*vi, graph);
  auto source = boost::source(*oei, graph);
  auto target = boost::target(*oei, graph);
  std::string source_name = boost::get(boost::vertex_name_t(), graph, source);
  std::string target_name = boost::get(boost::vertex_name_t(), graph, target);
  EXPECT_EQ(source_name, "a");
  EXPECT_EQ(target_name, "b");

  int64_t timestamp = boost::get(boost::edge_timestamp_t(), graph, *oei);
  EXPECT_EQ(timestamp, 0);

  neb::wei_t val = boost::get(boost::edge_weight_t(), graph, *oei);
  EXPECT_TRUE(val == 6);
}

TEST(test_algo,
     test_merge_topk_edges_with_same_from_and_same_to_larger_than_k) {
  neb::rt::transaction_graph tg2;
  tg2.add_edge(neb::to_address("a"), neb::to_address("b"), 1, 1);
  tg2.add_edge(neb::to_address("a"), neb::to_address("b"), 2, 2);
  tg2.add_edge(neb::to_address("a"), neb::to_address("b"), 3, 3);
  tg2.add_edge(neb::to_address("a"), neb::to_address("b"), 4, 4);

  auto graph = tg2.internal_graph();
  neb::rt::graph_algo::merge_topk_edges_with_same_from_and_same_to(graph);

  neb::rt::transaction_graph::viterator_t vi, vi_end;
  boost::tie(vi, vi_end) = boost::vertices(graph);
  neb::rt::transaction_graph::oeiterator_t oei, oei_end;
  boost::tie(oei, oei_end) = boost::out_edges(*vi, graph);

  auto source = boost::source(*oei, graph);
  auto target = boost::target(*oei, graph);
  std::string source_name = boost::get(boost::vertex_name_t(), graph, source);
  std::string target_name = boost::get(boost::vertex_name_t(), graph, target);
  EXPECT_EQ(source_name, "a");
  EXPECT_EQ(target_name, "b");

  int64_t timestamp = boost::get(boost::edge_timestamp_t(), graph, *oei);
  EXPECT_EQ(timestamp, 0);

  neb::wei_t val = boost::get(boost::edge_weight_t(), graph, *oei);
  EXPECT_TRUE(val != 6);
  EXPECT_TRUE(val == 9);
}

// class private func
TEST(test_algo, test_merge_two_graphs) {
  neb::rt::transaction_graph tg1;
  tg1.add_edge(neb::to_address("a"), neb::to_address("b"), 1, 1);
  tg1.add_edge(neb::to_address("a"), neb::to_address("c"), 2, 2);

  neb::rt::transaction_graph tg2;
  tg2.add_edge(neb::to_address("a"), neb::to_address("c"), 3, 3);
  tg2.add_edge(neb::to_address("a"), neb::to_address("d"), 4, 4);
  tg2.add_edge(neb::to_address("c"), neb::to_address("d"), 5, 5);

  neb::rt::transaction_graph_ptr_t ptr1 =
      std::make_unique<neb::rt::transaction_graph>(tg1);
  neb::rt::transaction_graph_ptr_t ptr2 =
      std::make_unique<neb::rt::transaction_graph>(tg2);
  neb::rt::graph_algo::merge_two_graphs(ptr1.get(), ptr2.get());

  neb::rt::transaction_graph::viterator_t vi, vi_end;

  auto graph = ptr1->internal_graph();
  for (boost::tie(vi, vi_end) = boost::vertices(graph); vi != vi_end; vi++) {
    neb::rt::transaction_graph::oeiterator_t oei, oei_end;
    for (boost::tie(oei, oei_end) = boost::out_edges(*vi, graph);
         oei != oei_end; oei++) {
      auto source = boost::source(*oei, graph);
      auto target = boost::target(*oei, graph);
      std::string source_name =
          boost::get(boost::vertex_name_t(), graph, source);
      std::string target_name =
          boost::get(boost::vertex_name_t(), graph, target);

      neb::wei_t val = boost::get(boost::edge_weight_t(), graph, *oei);
      if (source_name.compare("a") == 0 && target_name.compare("b") == 0) {
        EXPECT_TRUE(val == 1);
      } else if (source_name.compare("a") == 0 &&
                 target_name.compare("c") == 0) {
        EXPECT_TRUE(val == 2 || val == 3);
      } else if (source_name.compare("a") == 0 &&
                 target_name.compare("d") == 0) {
        EXPECT_TRUE(val == 4);
      } else if (source_name.compare("c") == 0 &&
                 target_name.compare("d") == 0) {
        EXPECT_TRUE(val == 5);
      }
    }
  }
}

TEST(test_algo, test_merge_graphs) {
  neb::rt::transaction_graph tg1;
  tg1.add_edge(neb::to_address("a"), neb::to_address("b"), 1, 1);
  tg1.add_edge(neb::to_address("a"), neb::to_address("c"), 2, 2);

  neb::rt::transaction_graph tg2;
  tg2.add_edge(neb::to_address("a"), neb::to_address("c"), 3, 3);
  tg2.add_edge(neb::to_address("a"), neb::to_address("d"), 4, 4);
  tg2.add_edge(neb::to_address("c"), neb::to_address("d"), 5, 5);

  neb::rt::transaction_graph tg3;
  tg3.add_edge(neb::to_address("e"), neb::to_address("f"), 6, 6);
  tg3.add_edge(neb::to_address("f"), neb::to_address("b"), 7, 7);
  tg3.add_edge(neb::to_address("c"), neb::to_address("f"), 8, 8);

  neb::rt::transaction_graph_ptr_t ptr1 =
      std::make_unique<neb::rt::transaction_graph>(tg1);
  neb::rt::transaction_graph_ptr_t ptr2 =
      std::make_unique<neb::rt::transaction_graph>(tg2);
  neb::rt::transaction_graph_ptr_t ptr3 =
      std::make_unique<neb::rt::transaction_graph>(tg3);
  std::vector<neb::rt::transaction_graph_ptr_t> v;
  v.push_back(std::move(ptr1));
  v.push_back(std::move(ptr2));
  v.push_back(std::move(ptr3));
  auto ptr = neb::rt::graph_algo::merge_graphs(v);

  neb::rt::transaction_graph::viterator_t vi, vi_end;

  auto graph = ptr->internal_graph();
  for (boost::tie(vi, vi_end) = boost::vertices(graph); vi != vi_end; vi++) {
    neb::rt::transaction_graph::oeiterator_t oei, oei_end;
    for (boost::tie(oei, oei_end) = boost::out_edges(*vi, graph);
         oei != oei_end; oei++) {
      auto source = boost::source(*oei, graph);
      auto target = boost::target(*oei, graph);
      std::string source_name =
          boost::get(boost::vertex_name_t(), graph, source);
      std::string target_name =
          boost::get(boost::vertex_name_t(), graph, target);

      neb::wei_t val = boost::get(boost::edge_weight_t(), graph, *oei);
      if (source_name.compare("a") == 0 && target_name.compare("b") == 0) {
        EXPECT_TRUE(val == 1);
      } else if (source_name.compare("a") == 0 &&
                 target_name.compare("c") == 0) {
        EXPECT_TRUE(val == 2 || val == 3);
      } else if (source_name.compare("a") == 0 &&
                 target_name.compare("d") == 0) {
        EXPECT_TRUE(val == 4);
      } else if (source_name.compare("c") == 0 &&
                 target_name.compare("d") == 0) {
        EXPECT_TRUE(val == 5);
      } else if (source_name.compare("e") == 0 &&
                 target_name.compare("f") == 0) {
        EXPECT_TRUE(val == 6);
      } else if (source_name.compare("f") == 0 &&
                 target_name.compare("b") == 0) {
        EXPECT_TRUE(val == 7);
      } else if (source_name.compare("c") == 0 &&
                 target_name.compare("f") == 0) {
        EXPECT_TRUE(val == 8);
      }
    }
  }
}

TEST(test_algo, test_find_a_cycle_based_on_time_sequence) {
  neb::rt::transaction_graph tg;
  tg.add_edge(neb::to_address("a"), neb::to_address("b"), 1, 5);
  tg.add_edge(neb::to_address("d"), neb::to_address("a"), 2, 1);
  tg.add_edge(neb::to_address("b"), neb::to_address("d"), 3, 2);
  tg.add_edge(neb::to_address("c"), neb::to_address("b"), 4, 4);
  tg.add_edge(neb::to_address("d"), neb::to_address("c"), 5, 3);

  auto graph = tg.internal_graph();
  auto cycle = neb::rt::graph_algo::find_a_cycle_based_on_time_sequence(graph);

  auto it = cycle.begin();
  auto source = boost::source(*it, graph);
  auto target = boost::target(*it, graph);
  std::string source_name = boost::get(boost::vertex_name_t(), graph, source);
  std::string target_name = boost::get(boost::vertex_name_t(), graph, target);
  neb::wei_t val = boost::get(boost::edge_timestamp_t(), graph, *it);
  EXPECT_TRUE(source_name.compare("b") == 0 && target_name.compare("d") == 0 &&
              val == 2);
  it++;

  source = boost::source(*it, graph);
  target = boost::target(*it, graph);
  source_name = boost::get(boost::vertex_name_t(), graph, source);
  target_name = boost::get(boost::vertex_name_t(), graph, target);
  val = boost::get(boost::edge_timestamp_t(), graph, *it);
  EXPECT_TRUE(source_name.compare("d") == 0 && target_name.compare("c") == 0 &&
              val == 3);
  it++;

  source = boost::source(*it, graph);
  target = boost::target(*it, graph);
  source_name = boost::get(boost::vertex_name_t(), graph, source);
  target_name = boost::get(boost::vertex_name_t(), graph, target);
  val = boost::get(boost::edge_timestamp_t(), graph, *it);
  EXPECT_TRUE(source_name.compare("c") == 0 && target_name.compare("b") == 0 &&
              val == 4);
  it++;
}

TEST(test_algo, test_remove_cycles_based_on_time_sequence_case1) {
  neb::rt::transaction_graph tg1;
  tg1.add_edge(neb::to_address("a"), neb::to_address("a"), 1, 1);
  tg1.add_edge(neb::to_address("a"), neb::to_address("b"), 2, 2);
  tg1.add_edge(neb::to_address("b"), neb::to_address("a"), 3, 3);
  tg1.add_edge(neb::to_address("b"), neb::to_address("c"), 4, 4);
  tg1.add_edge(neb::to_address("c"), neb::to_address("a"), 5, 5);

  auto graph = tg1.internal_graph();
  neb::rt::graph_algo::remove_cycles_based_on_time_sequence(graph);

  neb::rt::transaction_graph::viterator_t vi, vi_end;

  for (boost::tie(vi, vi_end) = boost::vertices(graph); vi != vi_end; vi++) {
    neb::rt::transaction_graph::oeiterator_t oei, oei_end;
    for (boost::tie(oei, oei_end) = boost::out_edges(*vi, graph);
         oei != oei_end; oei++) {
      auto source = boost::source(*oei, graph);
      auto target = boost::target(*oei, graph);
      std::string source_name =
          boost::get(boost::vertex_name_t(), graph, source);
      std::string target_name =
          boost::get(boost::vertex_name_t(), graph, target);

      neb::wei_t val = boost::get(boost::edge_weight_t(), graph, *oei);
      int64_t ts = boost::get(boost::edge_timestamp_t(), graph, *oei);
      if (source_name.compare("b") == 0 && target_name.compare("a") == 0) {
        EXPECT_TRUE(val == 1 && ts == 3);
      } else if (source_name.compare("b") == 0 &&
                 target_name.compare("c") == 0) {
        EXPECT_TRUE(val == 4 && ts == 4);
      } else if (source_name.compare("c") == 0 &&
                 target_name.compare("a") == 0) {
        EXPECT_TRUE(val == 5 && ts == 5);
      }
    }
  }
}

TEST(test_algo, test_remove_cycles_based_on_time_sequence_case2) {
  neb::rt::transaction_graph tg2;
  tg2.add_edge(neb::to_address("a"), neb::to_address("a"), 1, 1);
  tg2.add_edge(neb::to_address("a"), neb::to_address("c"), 2, 3);
  tg2.add_edge(neb::to_address("c"), neb::to_address("a"), 3, 2);
  tg2.add_edge(neb::to_address("c"), neb::to_address("b"), 4, 4);
  tg2.add_edge(neb::to_address("b"), neb::to_address("a"), 5, 5);

  auto graph = tg2.internal_graph();
  neb::rt::graph_algo::remove_cycles_based_on_time_sequence(graph);

  neb::rt::transaction_graph::viterator_t vi, vi_end;

  for (boost::tie(vi, vi_end) = boost::vertices(graph); vi != vi_end; vi++) {
    neb::rt::transaction_graph::oeiterator_t oei, oei_end;
    for (boost::tie(oei, oei_end) = boost::out_edges(*vi, graph);
         oei != oei_end; oei++) {
      auto source = boost::source(*oei, graph);
      auto target = boost::target(*oei, graph);
      std::string source_name =
          boost::get(boost::vertex_name_t(), graph, source);
      std::string target_name =
          boost::get(boost::vertex_name_t(), graph, target);

      neb::wei_t val = boost::get(boost::edge_weight_t(), graph, *oei);
      int64_t ts = boost::get(boost::edge_timestamp_t(), graph, *oei);
      if (source_name.compare("c") == 0 && target_name.compare("a") == 0) {
        EXPECT_TRUE(val == 3 && ts == 2);
      } else if (source_name.compare("c") == 0 &&
                 target_name.compare("b") == 0) {
        EXPECT_TRUE(val == 2 && ts == 4);
      } else if (source_name.compare("b") == 0 &&
                 target_name.compare("a") == 0) {
        EXPECT_TRUE(val == 3 && ts == 5);
      }
    }
  }
}

TEST(test_algo, test_remove_cycles_based_on_time_sequence_case3) {
  neb::rt::transaction_graph tg3;
  tg3.add_edge(neb::to_address("a"), neb::to_address("a"), 1, 1);
  tg3.add_edge(neb::to_address("a"), neb::to_address("b"), 3, 2);
  tg3.add_edge(neb::to_address("b"), neb::to_address("a"), 1, 3);
  tg3.add_edge(neb::to_address("b"), neb::to_address("c"), 2, 4);
  tg3.add_edge(neb::to_address("c"), neb::to_address("a"), 2, 5);

  auto graph = tg3.internal_graph();
  neb::rt::graph_algo::remove_cycles_based_on_time_sequence(graph);

  neb::rt::transaction_graph::viterator_t vi, vi_end;

  for (boost::tie(vi, vi_end) = boost::vertices(graph); vi != vi_end; vi++) {
    neb::rt::transaction_graph::oeiterator_t oei, oei_end;
    boost::tie(oei, oei_end) = boost::out_edges(*vi, graph);
    EXPECT_TRUE(oei == oei_end);
  }
}

TEST(test_algo, test_remove_cycles_based_on_time_sequence_case4) {
  neb::rt::transaction_graph tg4;
  tg4.add_edge(neb::to_address("a"), neb::to_address("a"), 1, 1);
  tg4.add_edge(neb::to_address("a"), neb::to_address("b"), 2, 2);
  tg4.add_edge(neb::to_address("b"), neb::to_address("a"), 3, 1);
  tg4.add_edge(neb::to_address("b"), neb::to_address("c"), 2, 3);
  tg4.add_edge(neb::to_address("c"), neb::to_address("a"), 1, 4);
  tg4.add_edge(neb::to_address("c"), neb::to_address("a"), 2, 4);

  auto graph = tg4.internal_graph();
  neb::rt::graph_algo::remove_cycles_based_on_time_sequence(graph);

  neb::rt::transaction_graph::viterator_t vi, vi_end;

  for (boost::tie(vi, vi_end) = boost::vertices(graph); vi != vi_end; vi++) {
    neb::rt::transaction_graph::oeiterator_t oei, oei_end;
    for (boost::tie(oei, oei_end) = boost::out_edges(*vi, graph);
         oei != oei_end; oei++) {
      auto source = boost::source(*oei, graph);
      auto target = boost::target(*oei, graph);
      std::string source_name =
          boost::get(boost::vertex_name_t(), graph, source);
      std::string target_name =
          boost::get(boost::vertex_name_t(), graph, target);

      neb::wei_t val = boost::get(boost::edge_weight_t(), graph, *oei);
      int64_t ts = boost::get(boost::edge_timestamp_t(), graph, *oei);
      if (source_name.compare("b") == 0 && target_name.compare("a") == 0) {
        EXPECT_TRUE(val == 3 && ts == 1);
      } else if (source_name.compare("c") == 0 &&
                 target_name.compare("a") == 0) {
        EXPECT_TRUE(val == 1 && ts == 4);
      }
    }
  }
}

