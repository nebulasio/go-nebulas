// Copyright (C) 2018 go-nebulas authors
//
//
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
#include "common/util/byte.h"
#include "fs/proto/ir.pb.h"
#include "fs/util.h"
#include <boost/format.hpp>
#include <boost/program_options.hpp>
#include <boost/process.hpp>
#include <fstream>
#include <iostream>
#include <algorithm>

namespace po = boost::program_options;
namespace bp = boost::process;

typedef struct clang_flags_t {
  std::string root_path;
  std::string flag_command;
}clang_flags;

void merge_clang_arguments(const std::vector<std::string> &flag_value_list,
    const clang_flags &flags, 
    std::string &command_string,
    bool is_add_root_path) {
  if (!flag_value_list.empty()) {
    command_string = command_string + flags.flag_command;
    std::for_each(flag_value_list.begin(), flag_value_list.end(),
        [&command_string, &flags, &is_add_root_path](const std::string &value) {
        if (is_add_root_path) {
        command_string = command_string + neb::fs::join_path(flags.root_path, value) + " ";
        } else {
        command_string = command_string + value + " ";
        }
        });
  }
}

int execute_command(const std::string &command_string) {
  bp::ipstream pipe_stream;
  bp::child c(command_string, bp::std_out > pipe_stream);

  std::string line;
  while(pipe_stream && std::getline(pipe_stream, line) && !line.empty()) {
    std::cerr << line << std::endl;
  }

  c.wait();
  return c.exit_code();
}

void make_ir_bitcode(neb::ir_conf_reader &reader, std::string &ir_bc_file, bool isPayload) {
  int result = -1;

  std::string current_path = neb::fs::cur_dir();
  std::string command_string(neb::fs::join_path(current_path, "lib/bin/clang") + 
                             " -O3 -emit-llvm ");

  clang_flags flags;

  flags.flag_command = "";
  merge_clang_arguments(reader.flags(), flags, command_string, false);

  flags.root_path = reader.root_path();
  flags.flag_command = " -I";
  merge_clang_arguments(reader.include_header_files(), flags, command_string, true);

  flags.flag_command = " -L";
  merge_clang_arguments(reader.link_path(), flags, command_string, true);

  flags.flag_command = " -l";
  merge_clang_arguments(reader.link_files(), flags, command_string, false);

  flags.flag_command = " -c ";
  merge_clang_arguments(reader.cpp_files(), flags, command_string, true);

  if (isPayload) {
    std::string temp_path = neb::fs::tmp_dir();
    ir_bc_file = neb::fs::join_path(temp_path, reader.self_ref().name() + "_ir.bc");
  }

  command_string += " -o " + ir_bc_file;

  LOG(INFO) << command_string;

  result = execute_command(command_string);
  if (result != 0) {
    LOG(INFO) << "error: executed by boost::process::system.";
    LOG(INFO) << "result code = " << result;
    exit(1);
  }
}

po::variables_map get_variables_map(int argc, char *argv[]) {
  po::options_description desc("Generate IR Payload");
  desc.add_options()("help", "show help message")(
      "input", po::value<std::string>(), "IR configuration file")(
      "output", po::value<std::string>(), "output file")(
      "model", po::value<std::string>()->default_value("payload"), 
      "Generate ir bitcode or ir payload. - [bitcode | payload], default:payload");

  po::variables_map vm;
  po::store(po::parse_command_line(argc, argv, desc), vm);
  po::notify(vm);
  if (vm.count("help")) {
    LOG(INFO) << desc << "\n";
    exit(1);
  }

  if (!vm.count("input")) {
    LOG(INFO) << "You must specify \"input\"!";
    exit(1);
  }

  return vm;
}

void make_ir_payload(std::ifstream &ifs, 
    const neb::ir_conf_reader &reader,
    const std::string &ir_bc_file,
    const std::string &output_file
    ) {
  ifs.open(ir_bc_file.c_str(), std::ios::in | std::ios::binary);
  if (!ifs.is_open()) {
    throw std::invalid_argument(
        boost::str(boost::format("can't open file %1%") % ir_bc_file));
  }

  ifs.seekg(0, ifs.end);
  std::ifstream::pos_type size = ifs.tellg();
  if (size > 128 * 1024) {
    throw std::invalid_argument("IR file too large!");
  }

  neb::util::bytes buf(size);

  ifs.seekg(0, ifs.beg);
  ifs.read((char *)buf.value(), buf.size());
  if (!ifs)
    throw std::invalid_argument(boost::str(
          boost::format("Read IR file error: only %1% could be read") %
          ifs.gcount()));

  nbre::NBREIR ir_info;
  ir_info.set_name(reader.self_ref().name());
  ir_info.set_version(reader.self_ref().version().data());
  ir_info.set_height(reader.available_height());
  for (size_t i = 0; i < reader.depends().size(); ++i) {
    nbre::NBREIRDepend *d = ir_info.add_depends();
    d->set_name(reader.depends()[i].name());
    d->set_version(reader.depends()[i].version().data());
  }
  ir_info.set_ir(neb::util::byte_to_string(buf));

  auto bytes_long = ir_info.ByteSizeLong();
  if (bytes_long > 128 * 1024) {
    throw std::invalid_argument("bytes too long !");
  }

  std::ofstream ofs;
  ofs.open(output_file,
      std::ios::out | std::ios::binary | std::ios::trunc);
  if (!ofs.is_open()) {
    throw std::invalid_argument("can't open output file");
  }
  neb::util::bytes out_bytes(bytes_long);
  ir_info.SerializeToArray((void *)out_bytes.value(), out_bytes.size());

  ofs.write((const char *)out_bytes.value(), out_bytes.size());
  ofs.close();
}

int main(int argc, char *argv[]) {
  po::variables_map vm = get_variables_map(argc, argv);
  std::ifstream ifs;

  try {
    std::string ir_fp = vm["input"].as<std::string>();
    neb::ir_conf_reader reader(ir_fp);

    std::string model = vm["model"].as<std::string>();
    std::string ir_bc_file;

    if (model == "payload") {
      make_ir_bitcode(reader, ir_bc_file, true);
      make_ir_payload(ifs, reader, ir_bc_file, vm["output"].as<std::string>());
    } else if (model == "bitcode") {
      ir_bc_file = vm["output"].as<std::string>();
      make_ir_bitcode(reader, ir_bc_file, false);
    } else {
      LOG(INFO) << "Error arguments of model, please show help message.";
      return 1;
    }
  } catch (std::exception &e) {
    ifs.close();
    LOG(INFO) << e.what() << std::endl;
  }

  return 0;
}
