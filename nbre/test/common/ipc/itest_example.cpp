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
#include "test/common/ipc/ipc_test.h"
#include <iostream>

void func() { throw std::runtime_error("tx"); }
void foo() { throw 2; }
void bar() {}

IPC_SERVER(test_example) {
  int a = 0, b = 0;
  IPC_CHECK_ANY_THROW(foo());
  IPC_CHECK_THROW(func(), std::runtime_error);
  IPC_CHECK_NO_THROW(bar());
  IPC_EXPECT(a == b) << "xxx";
  IPC_EXPECT(a == b);
  IPC_EXPECT_EQ(a, b);
}

IPC_CLIENT(test_example) { std::cout << "this is client" << std::endl; }

