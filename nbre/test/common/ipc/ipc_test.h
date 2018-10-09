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
#include "test/common/ipc/ipc_instance.h"

#define IPC_CHECK_THROW(statement, expected_exception)                         \
  {                                                                            \
    std::string internal_fail_msg = "";                                        \
    bool ipc_test_caught_expected = false;                                     \
    try {                                                                      \
      internal_fail_msg =                                                      \
          "Expected: " #statement " throws " #expected_exception               \
          ".\n Actual: it throws nothing. ";                                   \
      statement;                                                               \
    } catch (expected_exception const &) {                                     \
      ipc_test_caught_expected = true;                                         \
    } catch (...) {                                                            \
      ipc_test_caught_expected = false;                                        \
      internal_fail_msg =                                                      \
          "Expected: " #statement " throws " #expected_exception               \
          ".\n Actual: it throws a different type";                            \
    }                                                                          \
    if (!ipc_test_caught_expected) {                                           \
      std::cerr << internal_fail_msg;                                          \
      exit(1);                                                                 \
    }                                                                          \
  }

#define IPC_CHECK_NO_THROW(statement)                                          \
  {                                                                            \
    std::string internal_fail_msg = "";                                        \
    bool ipc_test_caught_nothing = true;                                       \
    try {                                                                      \
      statement;                                                               \
    } catch (...) {                                                            \
      ipc_test_caught_nothing = false;                                         \
      internal_fail_msg = "Expected: " #statement " throws nothing"            \
                          ".\n Actual: it throws an exception";                \
    }                                                                          \
    if (!ipc_test_caught_nothing) {                                            \
      std::cerr << internal_fail_msg;                                          \
      exit(1);                                                                 \
    }                                                                          \
  }

#define IPC_CHECK_ANY_THROW(statement)                                         \
  {                                                                            \
    std::string internal_fail_msg = "";                                        \
    bool ipc_test_caught_nothing = true;                                       \
    try {                                                                      \
      statement;                                                               \
    } catch (...) {                                                            \
      ipc_test_caught_nothing = false;                                         \
    }                                                                          \
    if (ipc_test_caught_nothing) {                                             \
      internal_fail_msg = "Expected: " #statement " throws anything"           \
                          ".\n Actual: it throws nothing.";                    \
      std::cerr << internal_fail_msg;                                          \
      exit(1);                                                                 \
    }                                                                          \
  }

//#define IPC_EXPECT(statement) \
