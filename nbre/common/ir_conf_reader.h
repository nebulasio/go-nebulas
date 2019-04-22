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
#include "common/common.h"
#include "common/ir_ref.h"
#include "common/version.h"

namespace boost{
namespace property_tree{
  template <typename T1, typename T2, typename Compare > class basic_ptree;
  typedef basic_ptree<std::string, std::string, std::less<std::string> > ptree;
}}

namespace neb {

struct json_general_failure : public std::exception {
  inline json_general_failure(const std::string &msg) : m_msg(msg) {}
  inline const char *what() const throw() { return m_msg.c_str(); }
protected:
  std::string m_msg;
};

class ir_conf_reader {
public:
  ir_conf_reader(const std::string &conf_fp);
  ~ir_conf_reader();
  ir_conf_reader(const ir_conf_reader &icr) = delete;
  ir_conf_reader &operator=(const ir_conf_reader &icr) = delete;

  inline const ir_ref &self_ref() const { return m_self_ref; }
  inline const std::vector<ir_ref> depends() const { return m_depends; }
  inline block_height_t available_height() const { return m_available_height; }
  inline const std::vector<std::string> cpp_files() const { return m_cpp_files; }
  inline const std::vector<std::string> include_header_files() const { return m_include_header_files; }
  inline const std::vector<std::string> link_files() const { return m_link_files; }
  inline const std::vector<std::string> link_path() const { return m_link_path; }
  inline const std::vector<std::string> flags() const { return m_flags; }

protected:
  ir_ref m_self_ref;
  std::vector<ir_ref> m_depends;
  block_height_t m_available_height;
  std::vector<std::string> m_cpp_files;
  std::vector<std::string> m_include_header_files;
  std::vector<std::string> m_link_files;
  std::vector<std::string> m_link_path;
  std::vector<std::string> m_flags;

private:
  void set_ir_ref_by_ptree(ir_ref &ir, const boost::property_tree::ptree &ptree);
  void read_json_file(const std::string &conf_fp, boost::property_tree::ptree &json_root);
  void get_ir_fp(const boost::property_tree::ptree &json_root);
  void get_self_ref(const boost::property_tree::ptree &json_root);
  void get_depends(const boost::property_tree::ptree &json_root);
  void get_available_height(const boost::property_tree::ptree &json_root);
  void get_clang_arguments(const boost::property_tree::ptree &json_root,
                                           const std::string &key,
                                           std::vector<std::string> &container);

};
} // namespace neb
