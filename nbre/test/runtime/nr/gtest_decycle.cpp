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
  neb::rt::transaction_graph tg;
  tg.add_edge(neb::to_address("a"), neb::to_address("a"), 1, 1);
  tg.add_edge(neb::to_address("a"), neb::to_address("b"), 2, 2);
  tg.add_edge(neb::to_address("b"), neb::to_address("c"), 3, 3);
  tg.add_edge(neb::to_address("c"), neb::to_address("a"), 4, 4);

  auto graph = tg.internal_graph();
  neb::rt::graph_algo::non_recursive_remove_cycles_based_on_time_sequence(
      graph);

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
      if (source_name.compare("b") == 0 && target_name.compare("c") == 0) {
        EXPECT_TRUE(val == 1 && ts == 3);
      } else if (source_name.compare("c") == 0 &&
                 target_name.compare("a") == 0) {
        EXPECT_TRUE(val == 2 && ts == 4);
      }
    }
  }
}

TEST(test_decycle, non_recursive_remove_cycles_based_on_time_sequence_case2) {
  neb::rt::transaction_graph tg;
  tg.add_edge(neb::to_address("a"), neb::to_address("a"), 1, 1);
  tg.add_edge(neb::to_address("a"), neb::to_address("b"), 2, 3);
  tg.add_edge(neb::to_address("b"), neb::to_address("a"), 3, 2);
  tg.add_edge(neb::to_address("b"), neb::to_address("c"), 4, 1);
  tg.add_edge(neb::to_address("c"), neb::to_address("a"), 5, 5);

  auto graph = tg.internal_graph();
  neb::rt::graph_algo::non_recursive_remove_cycles_based_on_time_sequence(
      graph);
  tg.write_to_graphviz("case2.txt");

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
        EXPECT_TRUE(val == 1 && ts == 2) << "val: " << val;
      } else if (source_name.compare("c") == 0 &&
                 target_name.compare("b") == 0) {
        EXPECT_TRUE(val == 4 && ts == 1);
      } else if (source_name.compare("b") == 0 &&
                 target_name.compare("a") == 0) {
        EXPECT_TRUE(val == 5 && ts == 5);
      }
    }
  }
}

TEST(test_decycle, non_recursive_remove_cycles_based_on_time_sequence_case3) {
  neb::rt::transaction_graph tg;
  tg.add_edge(neb::to_address("a"), neb::to_address("a"), 1, 1);
  tg.add_edge(neb::to_address("a"), neb::to_address("b"), 3, 2);
  tg.add_edge(neb::to_address("b"), neb::to_address("a"), 1, 3);
  tg.add_edge(neb::to_address("b"), neb::to_address("c"), 2, 4);
  tg.add_edge(neb::to_address("c"), neb::to_address("a"), 2, 5);

  auto graph = tg.internal_graph();
  neb::rt::graph_algo::non_recursive_remove_cycles_based_on_time_sequence(
      graph);

  neb::rt::transaction_graph::viterator_t vi, vi_end;
  tg.write_to_graphviz("case3.txt");

  for (boost::tie(vi, vi_end) = boost::vertices(graph); vi != vi_end; vi++) {
    neb::rt::transaction_graph::oeiterator_t oei, oei_end;
    boost::tie(oei, oei_end) = boost::out_edges(*vi, graph);
    EXPECT_TRUE(oei == oei_end);
  }
}

TEST(test_decycle, non_recursive_remove_cycles_based_on_time_sequence_case4) {
  neb::rt::transaction_graph tg;
  tg.add_edge(neb::to_address("a"), neb::to_address("b"), 2, 2);
  tg.add_edge(neb::to_address("b"), neb::to_address("c"), 3, 1);
  tg.add_edge(neb::to_address("c"), neb::to_address("d"), 1, 4);
  tg.add_edge(neb::to_address("d"), neb::to_address("b"), 2, 4);

  auto graph = tg.internal_graph();
  neb::rt::graph_algo::non_recursive_remove_cycles_based_on_time_sequence(
      graph);

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
        EXPECT_TRUE(val == 2 && ts == 1);
      } else if (source_name.compare("d") == 0 &&
                 target_name.compare("b") == 0) {
        EXPECT_TRUE(val == 1 && ts == 4);
      }
    }
  }
}

TEST(test_decycle, non_recursive_remove_cycles_based_on_time_sequence_case5) {
  neb::rt::transaction_graph tg;
  char cc = 'z';
  int32_t n = 1000;
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
  LOG(INFO) << boost::num_vertices(graph);
  LOG(INFO) << boost::num_edges(graph);
  neb::rt::graph_algo::non_recursive_remove_cycles_based_on_time_sequence(
      graph);
}

TEST(test_decycle, non_recursive_remove_cycles_based_on_time_sequence_case6) {
  neb::rt::transaction_graph tg;
  char cc = 'z';
  int32_t n = 100;
  for (char ch = 'a'; ch <= cc; ch++) {
    for (int32_t i = 0; i < n; i++) {
      tg.add_edge(neb::to_address(std::string(1, ch)),
                  neb::to_address(std::string(1, ch)), ch - 'a' + 1,
                  ch - 'a' + 1);
    }
  }

  auto graph = tg.internal_graph();
  neb::rt::graph_algo::non_recursive_remove_cycles_based_on_time_sequence(
      graph);
  EXPECT_EQ(boost::num_vertices(graph), cc - 'a' + 1);
  EXPECT_EQ(boost::num_edges(graph), 0);
}

TEST(test_decycle, non_recursive_remove_cycles_based_on_time_sequence_case7) {
  neb::rt::transaction_graph tg;
  tg.add_edge(neb::to_address("a"), neb::to_address("b"), 1, 1);
  tg.add_edge(neb::to_address("b"), neb::to_address("c"), 2, 1);
  tg.add_edge(neb::to_address("c"), neb::to_address("d"), 3, 2);
  tg.add_edge(neb::to_address("d"), neb::to_address("a"), 4, 2);

  auto graph = tg.internal_graph();
  neb::rt::graph_algo::non_recursive_remove_cycles_based_on_time_sequence(
      graph);

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
      if (source_name.compare("b") == 0 && target_name.compare("c") == 0) {
        EXPECT_TRUE(val == 1 && ts == 1);
      } else if (source_name.compare("c") == 0 &&
                 target_name.compare("d") == 0) {
        EXPECT_TRUE(val == 2 && ts == 2);
      } else if (source_name.compare("d") == 0 &&
                 target_name.compare("a") == 0) {
        EXPECT_TRUE(val == 3 && ts == 2);
      }
    }
  }
}

TEST(test_decycle, non_recursive_remove_cycles_based_on_time_sequence_case8) {
  neb::rt::transaction_graph tg;
  tg.add_edge(neb::to_address("b"), neb::to_address("c"), 1, 1);
  tg.add_edge(neb::to_address("c"), neb::to_address("d"), 2, 2);
  tg.add_edge(neb::to_address("d"), neb::to_address("b"), 3, 3);
  tg.add_edge(neb::to_address("a"), neb::to_address("d"), 1, 1);
  tg.add_edge(neb::to_address("b"), neb::to_address("a"), 2, 2);

  auto graph = tg.internal_graph();
  neb::rt::graph_algo::non_recursive_remove_cycles_based_on_time_sequence(
      graph);

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
      if (source_name.compare("c") == 0 && target_name.compare("d") == 0) {
        EXPECT_TRUE(val == 1 && ts == 2);
      } else if (source_name.compare("d") == 0 &&
                 target_name.compare("b") == 0) {
        EXPECT_TRUE(val == 2 && ts == 3);
      } else if (source_name.compare("a") == 0 &&
                 target_name.compare("d") == 0) {
        EXPECT_TRUE(val == 1 && ts == 1);
      } else if (source_name.compare("b") == 0 &&
                 target_name.compare("a") == 0) {
        EXPECT_TRUE(val == 2 && ts == 2);
      }
    }
  }
}

TEST(test_decycle, non_recursive_remove_cycles_based_on_time_sequence_case9) {
  neb::rt::transaction_graph tg;
  char cc = 'z';
  int32_t n = 100;
  for (char ch = 'a'; ch < cc; ch++) {
    for (int32_t i = 0; i < n; i++) {
      tg.add_edge(neb::to_address(std::string(1, ch)),
                  neb::to_address(std::string(1, ch + 1)), ch - 'a' + 1,
                  ch - 'a' + 1);
    }
  }

  auto &graph = tg.internal_graph();
  auto before_v_nums = boost::num_vertices(graph);
  auto before_e_nums = boost::num_edges(graph);
  neb::rt::graph_algo::non_recursive_remove_cycles_based_on_time_sequence(
      graph);
  auto after_v_nums = boost::num_vertices(graph);
  auto after_e_nums = boost::num_edges(graph);
  EXPECT_EQ(before_v_nums, after_v_nums);
  EXPECT_EQ(before_e_nums, after_e_nums);
}

TEST(test_decycle, non_recursive_remove_cycles_based_on_time_sequence_case10) {
  neb::rt::transaction_graph tg;
  std::queue<char> q;
  char cc = 'z';
  int32_t n = 2;
  q.push('a');
  char tail = 'a';
  while (!q.empty()) {
    auto &ele = q.front();
    q.pop();
    char tmp_tail = tail;
    if (tail + n < cc) {
      for (auto i = 1; i <= n; i++) {
        tg.add_edge(neb::to_address(std::string(1, ele)),
                    neb::to_address(std::string(1, tmp_tail + i)),
                    ele - 'a' + 1, ele - 'a' + 1);
        q.push(tmp_tail + i);
        tail++;
      }
      tmp_tail = tail;
    }
  }

  auto graph = tg.internal_graph();
  auto before_v_nums = boost::num_vertices(graph);
  auto before_e_nums = boost::num_edges(graph);
  LOG(INFO) << before_v_nums << ',' << before_e_nums;
  neb::rt::graph_algo::non_recursive_remove_cycles_based_on_time_sequence(
      graph);
  auto after_v_nums = boost::num_vertices(graph);
  auto after_e_nums = boost::num_edges(graph);
  EXPECT_EQ(before_v_nums, after_v_nums);
  EXPECT_EQ(before_e_nums, after_e_nums);
}

TEST(test_decycle, non_recursive_remove_cycles_based_on_time_sequence_case11) {
  neb::rt::transaction_graph tg;
  char cc = 'z';
  int32_t n = 5;
  int32_t tmp = cc - 'a' + 1;

  for (char s = 'a'; s <= cc; s++) {
    for (char t = 'a'; t <= cc; t++) {
      int32_t s_ch = s - 'a' + 1;
      int32_t t_ch = t - 'a' + 1;
      for (int32_t i = 0; i < n; i++) {
        tg.add_edge(neb::to_address(std::string(1, s)),
                    neb::to_address(std::string(1, t)), s_ch + tmp * t_ch + 1,
                    s_ch + tmp * t_ch + 1);
      }
    }
  }

  auto graph = tg.internal_graph();
  LOG(INFO) << "edge num: " << tg.edge_num();
  neb::rt::graph_algo::non_recursive_remove_cycles_based_on_time_sequence(
      graph);
}

TEST(test_decycle, non_recursive_remove_cycles_based_on_time_sequence_case12) {
  neb::rt::transaction_graph tg;
  char cc = 'z';
  char ch_half = cc / 2;
  int32_t n = 10;
  int32_t tmp = cc - 'a' + 1;

  for (char s = 'a'; s < ch_half; s++) {
    for (char t = ch_half; t <= cc; t++) {
      int32_t s_ch = s - 'a' + 1;
      int32_t t_ch = t - 'a' + 1;
      for (int32_t i = 0; i < n; i++) {
        tg.add_edge(neb::to_address(std::string(1, s)),
                    neb::to_address(std::string(1, t)), s_ch + tmp * t_ch + 1,
                    s_ch + tmp * t_ch + 1);
      }
    }
  }

  auto graph = tg.internal_graph();
  neb::rt::graph_algo::non_recursive_remove_cycles_based_on_time_sequence(
      graph);
}

