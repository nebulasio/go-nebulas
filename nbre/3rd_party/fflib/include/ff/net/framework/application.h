/***********************************************
The MIT License (MIT)

Copyright (c) 2012 Athrun Arthur <athrunarthur@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*************************************************/
#pragma once

#include "ff/net/common/common.h"
#include <boost/program_options.hpp>

namespace ff {
namespace net {
class routine;
class application {
public:
  application(const std::string &app_name);

  void initialize(int argc, char *argv[]);

  void register_routine(routine *rp);

  void run();

protected:
  void print_help();
  void list_routines();
  void run_routine();
  void start_routine(routine *r);

  std::vector<routine *> m_routines;
  std::function<void()> m_to_run_func;
  net_mode m_nm;
  std::vector<std::string> m_routine_name;
  std::vector<std::string> m_routine_args;
  std::string m_app_name;

  boost::program_options::options_description m_app_desc;
  boost::program_options::variables_map m_app_vm;
}; // end class application

} // namespace net
} // namespace ff
