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

#ifdef __cplusplus
extern "C" {
#endif

#include <stddef.h>
#include <stdint.h>

typedef struct EngineStruct {
  void *llvm_engine;
  void *llvm_builder;
  void *llvm_main_module;
  void *llvm_context;
  void *llvm_pass_manager;
  void *llvm_mem_manager;
} Engine;

Engine *CreateEngine();
int AddModuleFile(Engine *e, const char *irFile);
void DeleteEngine(Engine *e);
int RunFunction(Engine *e, const char *funcName, size_t len,
                const uint8_t *data);
void AddNamedFunction(Engine *e, const char *funcName, void *address);

void Initialize();

#ifdef __cplusplus
}
#endif
