// Copyright (C) 2017 go-nebulas authors
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
#include "common/common.h"
#include <boost/noncopyable.hpp>

namespace neb {
namespace util {

template <typename T> class singleton : boost::noncopyable {
public:
  static T &instance() {
    std::call_once(s_oOnce, std::bind(singleton<T>::init));
    return *s_pInstance;
  }
protected:
  singleton() {}

private:
  static void init() { s_pInstance = std::shared_ptr<T>(new T()); }

protected:
  static std::shared_ptr<T> s_pInstance;
  static std::once_flag s_oOnce;
};
template <typename T> std::shared_ptr<T> singleton<T>::s_pInstance;
template <typename T> std::once_flag singleton<T>::s_oOnce;
}
}
