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
#include <boost/algorithm/string/replace.hpp>
#include <boost/program_options.hpp>
#include <boost/property_tree/json_parser.hpp>
#include <boost/property_tree/ptree.hpp>

namespace po = boost::program_options;

int main(int argc, char *argv[]) {

  po::options_description desc("Params data structure to json string");
  desc.add_options()("help", "show help message")(
      "start_block", po::value<std::string>(), "start block height")(
      "end_block", po::value<std::string>(), "end block height")(
      "major_version", po::value<std::string>(), "major version")(
      "minor_version", po::value<std::string>(), "minor version")(
      "patch_version", po::value<std::string>(),
      "patch version")("a", po::value<std::string>(),
                       "params a")("b", po::value<std::string>(), "params b")(
      "c", po::value<std::string>(), "params c")("d", po::value<std::string>(),
                                                 "params d")(
      "theta", po::value<std::string>(),
      "params theta")("mu", po::value<std::string>(), "params mu")(
      "lambda", po::value<std::string>(), "params lambda");

  po::variables_map vm;
  po::store(po::parse_command_line(argc, argv, desc), vm);
  po::notify(vm);

  if (vm.count("help")) {
    std::cout << desc << "\n";
    return 1;
  }

  std::string start_block = vm["start_block"].as<std::string>();
  std::string end_block = vm["end_block"].as<std::string>();
  std::string major_version = vm["major_version"].as<std::string>();
  std::string minor_version = vm["minor_version"].as<std::string>();
  std::string patch_version = vm["patch_version"].as<std::string>();
  std::string a = vm["a"].as<std::string>();
  std::string b = vm["b"].as<std::string>();
  std::string c = vm["c"].as<std::string>();
  std::string d = vm["d"].as<std::string>();
  std::string theta = vm["theta"].as<std::string>();
  std::string mu = vm["mu"].as<std::string>();
  std::string lambda = vm["lambda"].as<std::string>();

  boost::property_tree::ptree pt;
  pt.put("start_block", start_block);
  pt.put("end_block", end_block);
  pt.put("major_version", major_version);
  pt.put("minor_version", minor_version);
  pt.put("patch_version", patch_version);
  pt.put("a", a);
  pt.put("b", b);
  pt.put("c", c);
  pt.put("d", d);
  pt.put("theta", theta);
  pt.put("mu", mu);
  pt.put("lambda", lambda);

  std::stringstream ss;
  boost::property_tree::json_parser::write_json(ss, pt, false);
  std::string tmp = ss.str();
  boost::replace_all(tmp, "\"", "\\\"");
  std::cout << tmp << std::endl;

  return 0;
}
