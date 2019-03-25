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
#include <gtest/gtest.h>

TEST(test_decycle, non_recursive_remove_cycles_based_on_time_sequence_case1) {
  neb::rt::transaction_graph tg1;
  tg1.add_edge(neb::to_address("a"), neb::to_address("a"), 1, 1);
  tg1.add_edge(neb::to_address("a"), neb::to_address("b"), 2, 2);
  tg1.add_edge(neb::to_address("b"), neb::to_address("a"), 3, 3);
  tg1.add_edge(neb::to_address("b"), neb::to_address("c"), 4, 4);
  tg1.add_edge(neb::to_address("c"), neb::to_address("a"), 5, 5);

  auto graph = tg1.internal_graph();
  neb::rt::opt::graph_algo::remove_cycles_based_on_time_sequence(graph);

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

TEST(test_decycle, non_recursive_remove_cycles_based_on_time_sequence_case2) {
  neb::rt::transaction_graph tg2;
  tg2.add_edge(neb::to_address("a"), neb::to_address("a"), 1, 1);
  tg2.add_edge(neb::to_address("a"), neb::to_address("c"), 2, 3);
  tg2.add_edge(neb::to_address("c"), neb::to_address("a"), 3, 2);
  tg2.add_edge(neb::to_address("c"), neb::to_address("b"), 4, 4);
  tg2.add_edge(neb::to_address("b"), neb::to_address("a"), 5, 5);

  auto graph = tg2.internal_graph();
  neb::rt::opt::graph_algo::remove_cycles_based_on_time_sequence(graph);

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
        EXPECT_TRUE(val == 1 && ts == 2);
      } else if (source_name.compare("c") == 0 &&
                 target_name.compare("b") == 0) {
        EXPECT_TRUE(val == 4 && ts == 4);
      } else if (source_name.compare("b") == 0 &&
                 target_name.compare("a") == 0) {
        EXPECT_TRUE(val == 5 && ts == 5);
      }
    }
  }
}

TEST(test_decycle, non_recursive_remove_cycles_based_on_time_sequence_case3) {
  neb::rt::transaction_graph tg3;
  tg3.add_edge(neb::to_address("a"), neb::to_address("a"), 1, 1);
  tg3.add_edge(neb::to_address("a"), neb::to_address("b"), 3, 2);
  tg3.add_edge(neb::to_address("b"), neb::to_address("a"), 1, 3);
  tg3.add_edge(neb::to_address("b"), neb::to_address("c"), 2, 4);
  tg3.add_edge(neb::to_address("c"), neb::to_address("a"), 2, 5);

  auto graph = tg3.internal_graph();
  neb::rt::opt::graph_algo::remove_cycles_based_on_time_sequence(graph);

  neb::rt::transaction_graph::viterator_t vi, vi_end;

  for (boost::tie(vi, vi_end) = boost::vertices(graph); vi != vi_end; vi++) {
    neb::rt::transaction_graph::oeiterator_t oei, oei_end;
    boost::tie(oei, oei_end) = boost::out_edges(*vi, graph);
    EXPECT_TRUE(oei == oei_end);
  }
}

TEST(test_decycle, non_recursive_remove_cycles_based_on_time_sequence_case4) {
  neb::rt::transaction_graph tg4;
  tg4.add_edge(neb::to_address("a"), neb::to_address("a"), 1, 1);
  tg4.add_edge(neb::to_address("a"), neb::to_address("b"), 2, 2);
  tg4.add_edge(neb::to_address("b"), neb::to_address("a"), 3, 1);
  tg4.add_edge(neb::to_address("b"), neb::to_address("c"), 2, 3);
  tg4.add_edge(neb::to_address("c"), neb::to_address("a"), 1, 4);
  tg4.add_edge(neb::to_address("c"), neb::to_address("a"), 2, 4);

  auto graph = tg4.internal_graph();
  neb::rt::opt::graph_algo::remove_cycles_based_on_time_sequence(graph);

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
        EXPECT_TRUE(val == 1 && ts == 1);
      } else if (source_name.compare("b") == 0 &&
                 target_name.compare("c") == 0) {
        EXPECT_TRUE(val == 2 && ts == 3);
      } else if (source_name.compare("c") == 0 &&
                 target_name.compare("a") == 0) {
        EXPECT_TRUE((val == 1 && ts == 4) || (val == 2 && ts == 4));
      }
    }
  }
}

TEST(test_decycle, non_recursive_remove_cycles_based_on_time_sequence_case5) {
  neb::rt::transaction_graph tg5;
  tg5.add_edge(neb::to_address("a"), neb::to_address("b"), 2, 2);
  tg5.add_edge(neb::to_address("b"), neb::to_address("c"), 3, 1);
  tg5.add_edge(neb::to_address("c"), neb::to_address("d"), 1, 4);
  tg5.add_edge(neb::to_address("d"), neb::to_address("b"), 2, 4);

  auto graph = tg5.internal_graph();
  neb::rt::opt::graph_algo::remove_cycles_based_on_time_sequence(graph);

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
      if (source_name.compare("a") == 0 && target_name.compare("b") == 0) {
        EXPECT_TRUE(val == 2 && ts == 2);
      } else if (source_name.compare("b") == 0 &&
                 target_name.compare("c") == 0) {
        EXPECT_TRUE(val == 3 && ts == 1);
      } else if (source_name.compare("c") == 0 &&
                 target_name.compare("d") == 0) {
        EXPECT_TRUE(val == 1 && ts == 4);
      } else if (source_name.compare("d") == 0 &&
                 target_name.compare("b") == 0) {
        EXPECT_TRUE(val == 2 && ts == 4);
      }
    }
  }
}

TEST(test_decycle, non_recursive_remove_cycles_based_on_time_sequence_case6) {
  neb::rt::transaction_graph tg6;
  char cc = 'b';
  int32_t n = 3;
  for (char ch = 'a'; ch < cc; ch++) {
    for (int32_t i = 0; i < n; i++) {
      tg6.add_edge(neb::to_address(std::string(1, ch)),
                   neb::to_address(std::string(1, ch + 1)), ch - 'a' + 1,
                   ch - 'a' + 1);
    }
  }
  for (int32_t i = 0; i < n; i++) {
    tg6.add_edge(neb::to_address(std::string(1, cc)),
                 neb::to_address(std::string(1, 'a')), cc - 'a' + 1,
                 cc - 'a' + 1);
  }

  auto graph = tg6.internal_graph();
  LOG(INFO) << boost::num_vertices(graph);
  LOG(INFO) << boost::num_edges(graph);
  neb::rt::opt::graph_algo::remove_cycles_based_on_time_sequence(graph);
  return;

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
      EXPECT_EQ(source_name, "b");
      EXPECT_EQ(target_name, "a");

      neb::wei_t val = boost::get(boost::edge_weight_t(), graph, *oei);
      int64_t ts = boost::get(boost::edge_timestamp_t(), graph, *oei);
      EXPECT_EQ(val, 1);
      EXPECT_EQ(ts, 2);
    }
  }
}

