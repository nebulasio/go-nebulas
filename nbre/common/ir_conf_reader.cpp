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
#include "common/ir_conf_reader.h"

#include <boost/property_tree/ptree.hpp>
#include <boost/property_tree/json_parser.hpp>
#include <boost/foreach.hpp>

namespace neb {

  template<typename T>
    void check_exception(T lambda) {
      try {
        lambda();
      } catch (const boost::property_tree::ptree_error &e) {
        throw json_general_failure(e.what());
      }
    }

  ir_conf_reader::ir_conf_reader(const std::string &conf_fp) {
    // TODO
    boost::property_tree::ptree json_root;
    read_json_file(conf_fp, json_root);
    get_ir_fp(json_root);
    get_self_ref(json_root);
    get_depends(json_root);
    get_available_height(json_root);
  }

  ir_conf_reader::~ir_conf_reader() = default;

  void ir_conf_reader::read_json_file(const std::string &conf_fp, boost::property_tree::ptree &json_root){
    auto lambda_fun = [&]() {
     boost::property_tree::read_json(conf_fp, json_root);
    };

    check_exception(lambda_fun);
  }

  void ir_conf_reader::get_ir_fp(const boost::property_tree::ptree &json_root){
    auto lambda_fun = [this, json_root]() {
      m_ir_fp = json_root.get<std::string>("ir_file_path");
    };

    check_exception(lambda_fun);
  }

  void ir_conf_reader::set_ir_ref_by_ptree(ir_ref &ir, const boost::property_tree::ptree &ptree) {
      ir.name() = ptree.get<std::string>("name");
      ir.version().major_version() = ptree.get<uint32_t>("version_major");
      ir.version().minor_version() = ptree.get<uint16_t>("version_minor");
      ir.version().patch_version() = ptree.get<uint16_t>("version_patch");
  }

  void ir_conf_reader::get_self_ref(const boost::property_tree::ptree &json_root){
    auto lambda_fun = [this, json_root]() {
      set_ir_ref_by_ptree(m_self_ref, json_root);
    };

    check_exception(lambda_fun);
  }

  void ir_conf_reader::get_depends(const boost::property_tree::ptree &json_root){
    auto lambda_fun = [this, json_root]() {
      boost::property_tree::ptree depends_node = json_root.get_child("depends");

      BOOST_FOREACH(boost::property_tree::ptree::value_type &child_node, depends_node) {
        ir_ref ir;
        set_ir_ref_by_ptree(ir, child_node.second);

        m_depends.push_back(ir);
      }
    };

    check_exception(lambda_fun);
  }

  void ir_conf_reader::get_available_height(const boost::property_tree::ptree &json_root){
    auto lambda_fun = [this, json_root]() {
      m_available_height = json_root.get<block_height_t>("available_height");
    };

    check_exception(lambda_fun);
  }
} // end namespace neb
