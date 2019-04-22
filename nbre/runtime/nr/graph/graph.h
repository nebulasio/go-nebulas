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
#include <boost/graph/adjacency_list.hpp>

namespace boost {
enum edge_timestamp_t { edge_timestamp };
enum edge_sort_id_t { edge_sort_id };
enum edge_check_id_t { edge_check_id };

BOOST_INSTALL_PROPERTY(edge, timestamp);
BOOST_INSTALL_PROPERTY(edge, sort_id);
BOOST_INSTALL_PROPERTY(edge, check_id);
} // namespace boost

namespace neb {
namespace rt {

class transaction_graph {
public:
  typedef boost::property<
      boost::edge_weight_t, wei_t,
      boost::property<boost::edge_timestamp_t, int64_t,
                      boost::property<boost::edge_sort_id_t, int64_t>>>
      edge_property_t;

  typedef boost::adjacency_list<
      boost::vecS, boost::vecS, boost::bidirectionalS,
      boost::property<boost::vertex_name_t, std::string>, edge_property_t>
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

  void add_edge(const address_t &from, const address_t &to, wei_t val,
                int64_t ts);

  void write_to_graphviz(const std::string &filename);

  bool read_from_graphviz(const std::string &filename);

  inline internal_graph_t &internal_graph() { return m_graph; }
  inline const internal_graph_t &internal_graph() const { return m_graph; }
  inline int64_t edge_num() const { return m_edge_index; }
  inline int64_t vertex_num() const { return m_cur_max_index; }

protected:
  internal_graph_t m_graph;

  std::unordered_map<int64_t, address_t> m_vertex_to_addr;
  std::unordered_map<address_t, int64_t> m_addr_to_vertex;

  uint64_t m_cur_max_index;
  uint64_t m_edge_index;
}; // end class transaction_graph

using transaction_graph_ptr_t = std::unique_ptr<transaction_graph>;

transaction_graph_ptr_t build_graph_from_internal(
    const transaction_graph::internal_graph_t &internal_graph);

} // namespace rt
} // namespace neb
