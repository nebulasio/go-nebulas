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

#include "engine.h"
#include "engine_int.h"
#include "lib/execution_env.h"
#include "lib/instruction_counter.h"
#include "lib/log_callback.h"
#include "lib/require_callback.h"
#include "lib/storage_object.h"
#include "lib/tracing.h"
#include "lib/blockchain.h"
#include "v8_data_inc.h"

#include <libplatform/libplatform.h>
#include <v8.h>

#include <assert.h>

using namespace v8;

static Platform *platformPtr = NULL;

void PrintException(Local<Context> context, TryCatch &trycatch);

#define STRINGIZE2(s) #s
#define STRINGIZE(s) STRINGIZE2(s)
#define V8VERSION_STRING                                                       \
  STRINGIZE(V8_MAJOR_VERSION)                                                  \
  "." STRINGIZE(V8_MINOR_VERSION) "." STRINGIZE(                               \
      V8_BUILD_NUMBER) "." STRINGIZE(V8_PATCH_LEVEL)

static char V8VERSION[] = V8VERSION_STRING;
char *GetV8Version() { return V8VERSION; }

void Initialize() {
  // Initialize V8.
  platformPtr = platform::CreateDefaultPlatform();

  V8::InitializeICU();
  V8::InitializePlatform(platformPtr);

  StartupData nativesData, snapshotData;
  nativesData.data = reinterpret_cast<char *>(natives_blob_bin);
  nativesData.raw_size = natives_blob_bin_len;
  snapshotData.data = reinterpret_cast<char *>(snapshot_blob_bin);
  snapshotData.raw_size = snapshot_blob_bin_len;
  V8::SetNativesDataBlob(&nativesData);
  V8::SetSnapshotDataBlob(&snapshotData);

  V8::Initialize();
}

void Dispose() {
  V8::Dispose();
  V8::ShutdownPlatform();
  if (platformPtr) {
    delete platformPtr;
    platformPtr = NULL;
  }
}

V8Engine *CreateEngine() {
  ArrayBuffer::Allocator *allocator =
      ArrayBuffer::Allocator::NewDefaultAllocator();

  Isolate::CreateParams create_params;
  create_params.array_buffer_allocator = allocator;

  Isolate *isolate = Isolate::New(create_params);

  // fix bug: https://github.com/nebulasio/go-nebulas/issues/5
  isolate->SetStackLimit(0x700000000000UL);

  V8Engine *e = (V8Engine *)calloc(1, sizeof(V8Engine));
  e->allocator = allocator;
  e->isolate = isolate;
  e->count_of_executed_instruction = 0;
  return e;
}

void DeleteEngine(V8Engine *e) {
  Isolate *isolate = static_cast<Isolate *>(e->isolate);
  isolate->Dispose();

  delete static_cast<ArrayBuffer::Allocator *>(e->allocator);

  free(e);
}

int ExecuteSourceDataDelegate(Isolate *isolate, const char *data,
                              Local<Context> context, TryCatch &trycatch,
                              void *delegateContext) {
  // Create a string containing the JavaScript source code.
  Local<String> source =
      String::NewFromUtf8(isolate, data, NewStringType::kNormal)
          .ToLocalChecked();

  // Compile the source code.
  ScriptOrigin sourceSrcOrigin(
      String::NewFromUtf8(isolate, "_contract_runner.js"));
  MaybeLocal<Script> script =
      Script::Compile(context, source, &sourceSrcOrigin);

  if (script.IsEmpty()) {
    PrintException(context, trycatch);
    return 1;
  }

  // Run the script to get the result.
  MaybeLocal<Value> ret = script.ToLocalChecked()->Run(context);
  if (ret.IsEmpty()) {
    PrintException(context, trycatch);
    return 1;
  }

  return 0;
}

char *InjectTracingInstructions(V8Engine *e, const char *source) {
  char *traceableSource = NULL;
  Execute(e, source, 0L, 0L, InjectTracingInstructionDelegate,
          (void *)&traceableSource);
  return traceableSource;
}

int RunScriptSource(V8Engine *e, const char *data, uintptr_t lcsHandler,
                    uintptr_t gcsHandler) {
  return Execute(e, data, (void *)lcsHandler, (void *)gcsHandler,
                 ExecuteSourceDataDelegate, NULL);
}

int Execute(V8Engine *e, const char *data, void *lcsHandler, void *gcsHandler,
            ExecutionDelegate delegate, void *delegateContext) {
  Isolate *isolate = static_cast<Isolate *>(e->isolate);
  assert(isolate);

  Isolate::Scope isolate_scope(isolate);
  // Create a stack-allocated handle scope.
  HandleScope handle_scope(isolate);

  // Create global object template.
  Local<ObjectTemplate> globalTpl = ObjectTemplate::New(isolate);
  NewNativeRequireFunction(isolate, globalTpl);
  NewNativeLogFunction(isolate, globalTpl);
  NewStorageType(isolate, globalTpl);
  NewBlockchain(isolate, globalTpl);

  // Create a new context.
  Local<Context> context = Context::New(isolate, NULL, globalTpl);

  // Disable eval().
  context->AllowCodeGenerationFromStrings(false);

  // Enter the context for compiling and running the script.
  Context::Scope context_scope(context);

  TryCatch trycatch(isolate);

  // Continue put objects to global object.
  NewStorageTypeInstance(isolate, context, lcsHandler, gcsHandler);
  NewInstructionCounterInstance(isolate, context,
                                &(e->count_of_executed_instruction));

  // Setup execution env.
  if (SetupExecutionEnv(isolate, context)) {
    // logErrorf("setup execution env failed.");
    PrintException(context, trycatch);
    return 1;
  }

  return delegate(isolate, data, context, trycatch, delegateContext);
}

void PrintException(Local<Context> context, TryCatch &trycatch) {
  // get stack trace.
  MaybeLocal<Value> stacktrace_ret = trycatch.StackTrace(context);

  if (stacktrace_ret.IsEmpty()) {
    // print exception only.
    Local<Value> exception = trycatch.Exception();
    String::Utf8Value exception_str(exception);
    LogErrorf("[V8 Exception] %s", *exception_str);
  } else {
    // print full stack trace.
    String::Utf8Value stack_str(stacktrace_ret.ToLocalChecked());
    LogErrorf("[V8 Exception] %s", *stack_str);
  }
}

V8EngineStats *GetV8EngineStatistics(V8Engine *e) {
  Isolate *isolate = static_cast<Isolate *>(e->isolate);
  HeapStatistics heap_stats;
  isolate->GetHeapStatistics(&heap_stats);

  V8EngineStats *stats =
      static_cast<V8EngineStats *>(calloc(1, sizeof(V8EngineStats)));

  stats->heap_size_limit = heap_stats.heap_size_limit();
  stats->malloced_memory = heap_stats.malloced_memory();
  stats->peak_malloced_memory = heap_stats.peak_malloced_memory();
  stats->total_available_size = heap_stats.total_available_size();
  stats->total_heap_size = heap_stats.total_heap_size();
  stats->total_heap_size_executable = heap_stats.total_heap_size_executable();
  stats->total_physical_size = heap_stats.total_physical_size();
  stats->used_heap_size = heap_stats.used_heap_size();

  return stats;
}

void TerminateExecution(V8Engine *e) {
  Isolate *isolate = static_cast<Isolate *>(e->isolate);
  isolate->TerminateExecution();
}
