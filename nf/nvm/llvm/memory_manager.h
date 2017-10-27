// Copyright (C) 2017 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or
// modify it under the terms of the GNU General Public License as published by
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

#include <llvm/ExecutionEngine/SectionMemoryManager.h>
#include <string>
#include <unordered_map>

using namespace llvm;

class MemoryManager : public SectionMemoryManager {
  MemoryManager(const MemoryManager &) = delete;
  void operator=(const MemoryManager &) = delete;

public:
  MemoryManager();
  virtual ~MemoryManager();

  void addNamedFunction(const char *Name, void *Address);
  void addNamedFunction(const std::string &Name, void *Address);

  /// This method returns a RuntimeDyld::SymbolInfo for the specified function
  /// or variable. It is used to resolve symbols during module linking.
  ///
  /// By default this falls back on the legacy lookup method:
  /// 'getSymbolAddress'. The address returned by getSymbolAddress is treated as
  /// a strong, exported symbol, consistent with historical treatment by
  /// RuntimeDyld.
  ///
  /// Clients writing custom RTDyldMemoryManagers are encouraged to override
  /// this method and return a SymbolInfo with the flags set correctly. This is
  /// necessary for RuntimeDyld to correctly handle weak and non-exported
  /// symbols.
  virtual JITSymbol findSymbol(const std::string &Name);

private:
  std::unordered_map<std::string, uint64_t> namedFunctionMap;
};
