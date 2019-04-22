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
#include "util/command.h"

namespace neb {
namespace util {

int command_executor::execute_command(const std::string &command_string) {
  namespace bp = boost::process;
  bp::ipstream pipe_stream;
  bp::child c(command_string, bp::std_out > pipe_stream);

  std::string line;
  while (pipe_stream && std::getline(pipe_stream, line) && !line.empty()) {
    std::cerr << line << std::endl;
  }
  c.wait();
  return c.exit_code();
}

int command_executor::execute_command(const std::string &command_string,
                                      std::string &output_string) {
  namespace bp = boost::process;
  bp::ipstream pipe_stream;
  bp::child c(command_string, bp::std_out > pipe_stream);
  output_string = std::string();

  while (pipe_stream && std::getline(pipe_stream, output_string) &&
         !output_string.empty()) {
    std::cerr << output_string << std::endl;
  }
  c.wait();
  return c.exit_code();
}

} // namespace util
} // namespace neb
