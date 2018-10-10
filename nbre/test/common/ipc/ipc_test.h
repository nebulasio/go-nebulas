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
      std::cerr << internal_fail_msg << std::endl;                             \
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
      std::cerr << internal_fail_msg << std::endl;                             \
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
      std::cerr << internal_fail_msg << std::endl;                             \
      exit(1);                                                                 \
    }                                                                          \
  }

class IPCCmpExpectHelper {
public:
  inline IPCCmpExpectHelper(bool val, const std::string &label)
      : m_value(val), m_label(label), m_ss() {}
  ~IPCCmpExpectHelper() {
    if (!m_value) {
      std::cerr << m_label << " expected true.\n Actual: it's false. "
                << m_ss.str() << std::endl;
      exit(1);
    }
  }

  template <typename T1, typename T2>
  IPCCmpExpectHelper(T1 &&t1, T2 &&t2, const std::string &label)
      : m_value(t1 == t2), m_label(label) {}

  template <typename T> IPCCmpExpectHelper &operator<<(T &&t) {
    m_ss << t;
    return *this;
  }

protected:
  bool m_value;
  std::string m_label;
  std::stringstream m_ss;
};

#define IPC_EXPECT(statement) IPCCmpExpectHelper(statement, #statement)

#define IPC_EXPECT_EQ(a, b)                                                    \
  IPCCmpExpectHelper(a, b, std::string(#a) + " == " + std::string(#b))
