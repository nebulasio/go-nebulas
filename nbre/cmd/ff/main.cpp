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

#include <ff/ff.h>
#include <iostream>

void f() {}

int main() {
  ff::para<> a;
  a([]() {
    std::this_thread::sleep_for(std::chrono::seconds(1));
    std::cout << "this is ff" << std::endl;
  });
  std::cout << "ret" << std::endl;
  std::this_thread::sleep_for(std::chrono::seconds(10));
  return 0;
}
