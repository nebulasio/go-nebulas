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
#include "util/timer_loop.h"
#include <iostream>

void func1() { std::cout << "func 1" << std::endl; }
void func2() { std::cout << "func 2" << std::endl; }
int main(int argc, char *argv[]) {
  boost::asio::io_service io_service;
  neb::util::timer_loop tl(&io_service);
  tl.register_timer_and_callback(1, func1);
  tl.register_timer_and_callback(5, func2);

  io_service.run();
  return 0;
}
