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
#include "runtime/nr/graph/graph.h"
#include <boost/graph/graphviz.hpp>
#include <map>

namespace neb {
namespace rt {

transaction_graph::transaction_graph() : m_cur_max_index(0), m_edge_index(0) {}

void transaction_graph::add_edge(const address_t &from, const address_t &to,
                                 wei_t val, int64_t ts) {
  uint64_t from_vertex, to_vertex;

  auto tmp_func = [&](const address_t &addr) {
    uint64_t ret;
    if (m_addr_to_vertex.find(addr) != m_addr_to_vertex.end()) {
      ret = m_addr_to_vertex[addr];
    } else {
      ret = m_cur_max_index;
      m_cur_max_index++;
      m_vertex_to_addr.insert(std::make_pair(ret, addr));
      m_addr_to_vertex.insert(std::make_pair(addr, ret));
    }
    return ret;
  };

  from_vertex = tmp_func(from);
  to_vertex = tmp_func(to);

  // boost::add_edge(from_vertex, to_vertex, {val, ts, m_edge_index}, m_graph);
  edge_property_t prop;
  boost::get_property_value(prop, boost::edge_weight) = val;
  boost::get_property_value(prop, boost::edge_timestamp) = ts;
  boost::get_property_value(prop, boost::edge_sort_id) = m_edge_index;
  boost::add_edge(from_vertex, to_vertex, prop, m_graph);
  m_edge_index++;

  boost::put(boost::vertex_name_t(), m_graph, from_vertex,
             std::to_string(from));
  boost::put(boost::vertex_name_t(), m_graph, to_vertex, std::to_string(to));
}

void transaction_graph::write_to_graphviz(const std::string &filename) {
  std::map<std::string, std::string> graph_attr, vertex_attr, edge_attr;

  std::ofstream of;
  of.open(filename);
  boost::dynamic_properties dp;
  dp.property("node_id", boost::get(boost::vertex_name, m_graph));

  boost::write_graphviz(
      of, m_graph,
      boost::make_label_writer(boost::get(boost::vertex_name, m_graph)),
      boost::make_label_writer(boost::get(boost::edge_weight, m_graph)),
      boost::make_graph_attributes_writer(graph_attr, vertex_attr, edge_attr));

  of.close();
}

bool transaction_graph::read_from_graphviz(const std::string &filename) {
  std::ifstream ifs(filename);
  if (!ifs) {
    return false;
  }

  std::stringstream ss;
  ss << ifs.rdbuf();
  ifs.close();

  boost::dynamic_properties dp(boost::ignore_other_properties);
  dp.property("label", boost::get(boost::vertex_name, m_graph));
  dp.property("label", boost::get(boost::edge_weight, m_graph));
  return boost::read_graphviz(ss, m_graph, dp);
}

transaction_graph_ptr_t build_graph_from_internal(
    const transaction_graph::internal_graph_t &internal_graph) {

  transaction_graph_ptr_t tg_ptr = std::make_unique<transaction_graph>();
  auto sgi = internal_graph;

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
      int64_t t = boost::get(boost::edge_timestamp_t(), sgi, *oei);

      tg_ptr->add_edge(from, to, w, t);
    }
  }
  return tg_ptr;
}
} // namespace rt

} // namespace neb
