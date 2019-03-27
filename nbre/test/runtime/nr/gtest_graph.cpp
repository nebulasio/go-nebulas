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
#include "runtime/nr/graph/graph.h"
#include <gtest/gtest.h>
#include <random>

TEST(test_runtime_graph, add_edge_simple) {
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

TEST(test_runtime_graph, add_edge_self_cycle) {
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

TEST(test_runtime_graph, add_edge_multi) {
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

TEST(test_runtime_graph, add_edge_random) {
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

TEST(test_runtime_graph, build_graph_from_internal) {
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

