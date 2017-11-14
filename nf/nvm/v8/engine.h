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

#ifndef _NEBULAS_NF_NVM_V8_ENGINE_H_
#define _NEBULAS_NF_NVM_V8_ENGINE_H_

#if BUILDING_DLL
#define EXPORT __attribute__((__visibility__("default")))
#else
#define EXPORT
#endif

#ifdef __cplusplus
extern "C" {
#endif // __cplusplus

#include <stddef.h>
#include <stdint.h>

enum LogLevel {
  DEBUG = 1,
  WARN = 2,
  INFO = 3,
  ERROR = 4,
};

// log
typedef void (*LogFunc)(int level, const char *msg);
EXPORT const char *GetLogLevelText(int level);
EXPORT void InitializeLogger(LogFunc f);

// storage
typedef char *(*StorageGetFunc)(void *handler, const char *key);
typedef int (*StoragePutFunc)(void *handler, const char *key,
                              const char *value);
typedef int (*StorageDelFunc)(void *handler, const char *key);
EXPORT void InitializeStorage(StorageGetFunc get, StoragePutFunc put,
                              StorageDelFunc del);

// blockchain
typedef char *(*GetBlockByHashFunc)(void *handler, const char *hash);
typedef char *(*GetTxByHashFunc)(void *handler, const char *hash);
typedef char *(*GetAccountStateFunc)(void *handler, const char *address);
typedef int  (*SendFunc)(void *handler, const char *to, const char *value);
EXPORT void InitializeBlockchain(GetBlockByHashFunc getBlock, GetTxByHashFunc getTx, GetAccountStateFunc getAccount,
                        SendFunc send);

// version
EXPORT char *GetV8Version();

typedef struct V8Engine {
  void *isolate;
  void *allocator;
  size_t count_of_executed_instruction;
} V8Engine;

typedef struct V8EngineStats {
  size_t total_heap_size;
  size_t total_heap_size_executable;
  size_t total_physical_size;
  size_t total_available_size;
  size_t used_heap_size;
  size_t heap_size_limit;
  size_t malloced_memory;
  size_t peak_malloced_memory;
} V8EngineStats;

EXPORT void Initialize();
EXPORT void Dispose();

EXPORT V8Engine *CreateEngine();

EXPORT int RunScriptSource(V8Engine *e, const char *data, uintptr_t lcsHandler,
                           uintptr_t gcsHandler);

EXPORT char *InjectTracingInstructions(V8Engine *e, const char *source);

EXPORT V8EngineStats *GetV8EngineStatistics(V8Engine *e);

EXPORT void TerminateExecution(V8Engine *e);

EXPORT void DeleteEngine(V8Engine *e);

EXPORT char *EncapsulateSourceToModuleStyle(const char *source);

#ifdef __cplusplus
}
#endif // __cplusplus

#endif // _NEBULAS_NF_NVM_V8_ENGINE_H_
