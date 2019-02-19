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
#include "ff/net/framework/application.h"
#include "ff/net/common/defines.h"
#include "ff/net/framework/routine.h"

namespace ff {
namespace net {
application::application(const std::string &app_name)
    : m_app_name(app_name), m_app_desc("Allowed options") {
  m_app_desc.add_options()("help", "show help messages")(
      "net_mode", boost::program_options::value<int>(),
      "set net mode, real_net:1, simu_net:2, single_net:3")(
      "routine", boost::program_options::value<std::string>(),
      "to run routine")("list", "list routines");
}

void application::initialize(int argc, char *argv[]) {
  boost::program_options::store(
      boost::program_options::parse_command_line(argc, argv, m_app_desc),
      m_app_vm);
}

void application::register_routine(routine *rp) {
  assert(rp != NULL);
  m_routines.push_back(rp);
}

void application::run() {
  if (m_app_vm.count("help")) {
    std::cout << m_app_desc << std::endl;
    return;
  }
  if (m_app_vm.count("list")) {
    list_routines();
    return;
  }
  if (m_app_vm["routine"].as<std::string>() != std::string("")) {
    m_routine_name.push_back(m_app_vm["routine"].as<std::string>());
  }
  m_nm = net_mode::real_net;
  if (m_app_vm.count("net_mode")) {
    m_nm = static_cast<net_mode>(m_app_vm["net_mode"].as<int>());
  }
  run_routine();
}

void application::list_routines() {
  std::cout << "\tall:  represent all routines" << std::endl;
  for (size_t i = 0; i < m_routines.size(); ++i) {
    std::cout << "\t" << m_routines[i]->get_name() << std::endl;
  }
}

void application::run_routine() {
  std::map<std::string, routine *> rnames;
  for (size_t i = 0; i < m_routines.size(); ++i) {
    std::string n = m_routines[i]->get_name();
    if (rnames.find(n) != rnames.end()) {
      std::cout << "More than one routine with same name, " << n << std::endl;
      return;
    }
    rnames.insert(std::make_pair(m_routines[i]->get_name(), m_routines[i]));
  }

  for (size_t i = 0; i < m_routine_name.size(); ++i) {
    std::string s = m_routine_name[i];
    if (rnames.find(s) == rnames.end()) {
      std::cout << "Cannot find routine " << s << std::endl;
      return;
    }
  }

  if (m_routine_name.size() == 0) {
    std::cout << "No available routines to run!" << std::endl;
  }
  for (size_t i = 0; i < m_routine_name.size(); ++i) {
    std::string s = m_routine_name[i];
    routine *pr = rnames[s];

    start_routine(pr);
  }
}

void application::start_routine(routine *r) {
  r->initialize(m_nm, m_routine_args);
  r->run();
  /*
  if(m_nm == real_net){
    r->initialize(m_nm, m_routine_args);
    r->run();
  }else if(m_nm == simu_net){

    std::cout<<"TODO start_routine:do something here!"<<std::endl;

  }*/
}

} // namespace net
} // namespace ff

