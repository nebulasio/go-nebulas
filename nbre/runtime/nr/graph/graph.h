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
#include <boost/graph/adjacency_list.hpp>

namespace boost {
enum edge_timestamp_t { edge_timestamp };

BOOST_INSTALL_PROPERTY(edge, timestamp);
} // namespace boost

namespace neb {
namespace rt {

class transaction_graph {
public:
  typedef boost::adjacency_list<
      boost::vecS, boost::vecS, boost::bidirectionalS,
      boost::property<boost::vertex_name_t, std::string>,
      boost::property<boost::edge_weight_t, double,
                      boost::property<boost::edge_timestamp_t, double>>>
      internal_graph_t;

  typedef typename boost::graph_traits<internal_graph_t>::vertex_descriptor
      vertex_descriptor_t;
  typedef typename boost::graph_traits<internal_graph_t>::edge_descriptor
      edge_descriptor_t;

  typedef typename boost::graph_traits<
      transaction_graph::internal_graph_t>::vertex_iterator viterator_t;
  typedef typename boost::graph_traits<
      transaction_graph::internal_graph_t>::in_edge_iterator ieiterator_t;
  typedef typename boost::graph_traits<
      transaction_graph::internal_graph_t>::out_edge_iterator oeiterator_t;

  transaction_graph();

  void add_edge(const std::string &from, const std::string &to, double val,
                double ts);

  void write_to_graphviz(const std::string &filename);

  bool read_from_graphviz(const std::string &filename);

  inline internal_graph_t &internal_graph() { return m_graph; }
  inline const internal_graph_t &internal_graph() const { return m_graph; }

protected:
  internal_graph_t m_graph;

  std::unordered_map<int64_t, std::string> m_vertex_to_addr;
  std::unordered_map<std::string, int64_t> m_addr_to_vertex;

  uint64_t m_cur_max_index;

}; // end class transaction_graph

typedef std::shared_ptr<transaction_graph> transaction_graph_ptr_t;

transaction_graph_ptr_t build_graph_from_internal(
    const transaction_graph::internal_graph_t &internal_graph);

} // namespace rt
} // namespace neb
