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
#include <gtest/gtest.h>

TEST(test_command, simple) {
  EXPECT_EQ(0, neb::util::command_executor::execute_command("touch hah"));
  EXPECT_EQ(0, neb::util::command_executor::execute_command("rm -rf hah"));
  EXPECT_THROW(neb::util::command_executor::execute_command("hah"),
               std::exception);
}

TEST(test_command, complex) {
  std::string output;
  EXPECT_EQ(0,
            neb::util::command_executor::execute_command("touch hah", output));
  EXPECT_EQ(output, std::string());

  EXPECT_EQ(0,
            neb::util::command_executor::execute_command("rm -rf hah", output));
  EXPECT_EQ(output, std::string());

  EXPECT_THROW(neb::util::command_executor::execute_command("hah", output),
               std::exception);
  EXPECT_EQ(output, std::string());
}
