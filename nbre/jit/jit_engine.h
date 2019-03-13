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
#include "common/common.h"
#include "jit/OrcLazyJIT.h"

namespace neb {
namespace jit {
using namespace llvm;
class jit_engine {
public:

  template <typename RT, typename... ARGS> RT run(ARGS... args) {
    std::unique_lock<std::mutex> _l(m_mutex);
    using MainFnPtr = RT (*)(ARGS...);
    auto main_func = fromTargetAddress<MainFnPtr>(
        cantFail(m_main_sym->getAddress(), nullptr));
    if (nullptr == main_func)
      return RT();
    return main_func(args...);
  }

  void init(std::vector<std::unique_ptr<Module>> ms,
            const std::string &func_name);

protected:
  template <typename PtrTy>
  static PtrTy fromTargetAddress(llvm::JITTargetAddress Addr) {
    return reinterpret_cast<PtrTy>(static_cast<uintptr_t>(Addr));
  }

protected:
  std::vector<std::unique_ptr<Module>> m_modules;
  std::string m_func_name;
  std::unique_ptr<llvm::EngineBuilder> m_EB;
  std::unique_ptr<Triple> m_T;
  std::unique_ptr<OrcLazyJIT> m_jit;
  std::mutex m_mutex;

  std::unique_ptr<llvm::JITSymbol> m_main_sym;
}; // end class jit_engine
} // namespace jit
} // namespace neb
