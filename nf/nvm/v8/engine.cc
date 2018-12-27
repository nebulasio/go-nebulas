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
#include "allocator.h"
#include "engine_int.h"
#include "lib/execution_env.h"
#include "lib/global.h"
#include "lib/instruction_counter.h"
#include "lib/logger.h"
#include "lib/tracing.h"
#include "lib/typescript.h"
#include "v8_data_inc.h"

#include <libplatform/libplatform.h>


#include <assert.h>
#include <string.h>
#include <thread>
#include <stdlib.h>
#include <stdio.h>

using namespace v8;

static Platform *platformPtr = NULL;

void PrintException(Local<Context> context, TryCatch &trycatch);
void PrintAndReturnException(char **exception, Local<Context> context,
                             TryCatch &trycatch);
void EngineLimitsCheckDelegate(Isolate *isolate, size_t count,
                               void *listenerContext);

#define ExecuteTimeOut  5*1000*1000
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

  // Initialize V8Engine.
  SetInstructionCounterIncrListener(EngineLimitsCheckDelegate);
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
  ArrayBuffer::Allocator *allocator = new ArrayBufferAllocator();

  Isolate::CreateParams create_params;
  create_params.array_buffer_allocator = allocator;

  Isolate *isolate = Isolate::New(create_params);

  // fix bug: https://github.com/nebulasio/go-nebulas/issues/5
  // isolate->SetStackLimit(0x700000000000UL);

  V8Engine *e = (V8Engine *)calloc(1, sizeof(V8Engine));
  e->allocator = allocator;
  e->isolate = isolate;
  e->timeout = ExecuteTimeOut;
  e->ver = BUILD_DEFAULT_VER; //default load initial com
  return e;
}
void EnableInnerContract(V8Engine *e) {
  e->ver = BUILD_INNER_VER;
}

void DeleteEngine(V8Engine *e) {
  Isolate *isolate = static_cast<Isolate *>(e->isolate);
  isolate->Dispose();

  delete static_cast<ArrayBuffer::Allocator *>(e->allocator);

  free(e);
}

int ExecuteSourceDataDelegate(char **result, Isolate *isolate,
                              const char *source, int source_line_offset,
                              Local<Context> context, TryCatch &trycatch,
                              void *delegateContext) {
  // Create a string containing the JavaScript source code.
  Local<String> src =
      String::NewFromUtf8(isolate, source, NewStringType::kNormal)
          .ToLocalChecked();

  // Compile the source code.
  ScriptOrigin sourceSrcOrigin(
      String::NewFromUtf8(isolate, "_contract_runner.js"),
      Integer::New(isolate, source_line_offset));
  MaybeLocal<Script> script = Script::Compile(context, src, &sourceSrcOrigin);

  if (script.IsEmpty()) {
    PrintAndReturnException(result, context, trycatch);
    return NVM_EXCEPTION_ERR;
  }
  // Run the script to get the result.
  MaybeLocal<Value> ret = script.ToLocalChecked()->Run(context);
  if (ret.IsEmpty()) {
    PrintAndReturnException(result, context, trycatch);
    return NVM_EXCEPTION_ERR;
  }
  // set result.
  if (result != NULL) {
    Local<Object> obj = ret.ToLocalChecked().As<Object>();
    if (!obj->IsUndefined()) {
      MaybeLocal<String> json_result = v8::JSON::Stringify(context, obj);
      if (!json_result.IsEmpty()) {
        String::Utf8Value str(json_result.ToLocalChecked());
        *result = (char *)malloc(str.length() + 1);
        strcpy(*result, *str);
      }
    }
  }

  return NVM_SUCCESS;
}

char *InjectTracingInstructions(V8Engine *e, const char *source,
                                int *source_line_offset,
                                int strictDisallowUsage) {
  TracingContext tContext;
  tContext.source_line_offset = 0;
  tContext.tracable_source = NULL;
  tContext.strictDisallowUsage = strictDisallowUsage;

  Execute(NULL, e, source, 0, 0L, 0L, InjectTracingInstructionDelegate,
          (void *)&tContext);

  *source_line_offset = tContext.source_line_offset;
  return static_cast<char *>(tContext.tracable_source);
}

char *TranspileTypeScriptModule(V8Engine *e, const char *source,
                                int *source_line_offset) {
  TypeScriptContext tContext;
  tContext.source_line_offset = 0;
  tContext.js_source = NULL;

  Execute(NULL, e, source, 0, 0L, 0L, TypeScriptTranspileDelegate,
          (void *)&tContext);

  *source_line_offset = tContext.source_line_offset;
  return static_cast<char *>(tContext.js_source);
}

int RunScriptSource(char **result, V8Engine *e, const char *source,
                    int source_line_offset, uintptr_t lcsHandler,
                    uintptr_t gcsHandler) {
  return Execute(result, e, source, source_line_offset, (void *)lcsHandler,
                 (void *)gcsHandler, ExecuteSourceDataDelegate, NULL);
}

int Execute(char **result, V8Engine *e, const char *source,
            int source_line_offset, void *lcsHandler, void *gcsHandler,
            ExecutionDelegate delegate, void *delegateContext) {
  Isolate *isolate = static_cast<Isolate *>(e->isolate);
  Locker locker(isolate);
  assert(isolate);

  Isolate::Scope isolate_scope(isolate);
  // Create a stack-allocated handle scope.
  HandleScope handle_scope(isolate);

  // Create global object template.
  Local<ObjectTemplate> globalTpl = CreateGlobalObjectTemplate(isolate);

  // Create a new context.
  Local<Context> context = Context::New(isolate, NULL, globalTpl);

  // disable eval().
  context->AllowCodeGenerationFromStrings(false);

  // Enter the context for compiling and running the script.
  Context::Scope context_scope(context);

  TryCatch trycatch(isolate);

  // Continue put objects to global object.
  SetGlobalObjectProperties(isolate, context, e, lcsHandler, gcsHandler);

  // Setup execution env.
  if (SetupExecutionEnv(isolate, context)) {
    PrintAndReturnException(result, context, trycatch);
    return NVM_EXCEPTION_ERR;
  }

  int retTmp = delegate(result, isolate, source, source_line_offset, context,
                  trycatch, delegateContext);
  
  if (e->is_unexpected_error_happen) {
    return NVM_UNEXPECTED_ERR;
  }

  return retTmp;
}

void PrintException(Local<Context> context, TryCatch &trycatch) {
  PrintAndReturnException(NULL, context, trycatch);
}

void PrintAndReturnException(char **exception, Local<Context> context,
                             TryCatch &trycatch) {
  static char SOURCE_INFO_PLACEHOLDER[] = "";
  char *source_info = NULL;

  // print source line.
  Local<Message> message = trycatch.Message();
  if (!message.IsEmpty()) {
    // Print (filename):(line number): (message).
    ScriptOrigin origin = message->GetScriptOrigin();
    String::Utf8Value filename(message->GetScriptResourceName());
    int linenum = message->GetLineNumber();

    // Print line of source code.
    String::Utf8Value sourceline(message->GetSourceLine());
    int script_start = (linenum - origin.ResourceLineOffset()->Value()) == 1
                           ? origin.ResourceColumnOffset()->Value()
                           : 0;
    int start = message->GetStartColumn(context).FromMaybe(0);
    int end = message->GetEndColumn(context).FromMaybe(0);
    if (start >= script_start) {
      start -= script_start;
      end -= script_start;
    }

    char arrow[start + 1];
    for (int i = 0; i < start; i++) {
      char c = (*sourceline)[i];
      if (c == '\t') {
        arrow[i] = c;
      } else {
        arrow[i] = ' ';
      }
    }
    arrow[start] = '^';
    arrow[start + 1] = '\0';

    asprintf(&source_info, "%s:%d\n%s\n%s\n", *filename, linenum, *sourceline,
             arrow);
  }

  if (source_info == NULL) {
    source_info = SOURCE_INFO_PLACEHOLDER;
  }

  // get stack trace.
  MaybeLocal<Value> stacktrace_ret = trycatch.StackTrace(context);
  if (!stacktrace_ret.IsEmpty()) {
    // print full stack trace.
    String::Utf8Value stack_str(stacktrace_ret.ToLocalChecked());
    LogErrorf("V8 Exception:\n%s%s", source_info, *stack_str);
  }

  // exception message.
  Local<Value> exceptionValue = trycatch.Exception();
  String::Utf8Value exception_str(exceptionValue);
  if (stacktrace_ret.IsEmpty()) {
    // print exception when stack trace is not available.
    LogErrorf("V8 Exception:\n%s%s", source_info, *exception_str);
  }

  if (source_info != NULL && source_info != SOURCE_INFO_PLACEHOLDER) {
    free(source_info);
  }

  // return exception message.
  if (exception != NULL) {
    *exception = (char *)malloc(exception_str.length() + 1);
    //TODO: Branch code detected “(v8::String::Utf8Value) $0 = (str_ = <no value available>, length_ = 0)”
    if (exception_str.length() == 0) {
      strcpy(*exception, "");
    } else {
      strcpy(*exception, *exception_str);
    }
  }
}

void ReadMemoryStatistics(V8Engine *e) {
  Isolate *isolate = static_cast<Isolate *>(e->isolate);
  ArrayBufferAllocator *allocator =
      static_cast<ArrayBufferAllocator *>(e->allocator);

  HeapStatistics heap_stats;
  isolate->GetHeapStatistics(&heap_stats);

  V8EngineStats *stats = &(e->stats);
  stats->heap_size_limit = heap_stats.heap_size_limit();
  stats->malloced_memory = heap_stats.malloced_memory();
  stats->peak_malloced_memory = heap_stats.peak_malloced_memory();
  stats->total_available_size = heap_stats.total_available_size();
  stats->total_heap_size = heap_stats.total_heap_size();
  stats->total_heap_size_executable = heap_stats.total_heap_size_executable();
  stats->total_physical_size = heap_stats.total_physical_size();
  stats->used_heap_size = heap_stats.used_heap_size();
  stats->total_array_buffer_size = allocator->total_available_size();
  stats->peak_array_buffer_size = allocator->peak_allocated_size();

  stats->total_memory_size =
      stats->total_heap_size + stats->peak_array_buffer_size;
}

void TerminateExecution(V8Engine *e) {
  if (e->is_requested_terminate_execution) {
    return;
  }
  Isolate *isolate = static_cast<Isolate *>(e->isolate);
  isolate->TerminateExecution();
  e->is_requested_terminate_execution = true;
}
void SetInnerContractErrFlag(V8Engine *e) {
  e->is_inner_nvm_error_happen = true;
}
void EngineLimitsCheckDelegate(Isolate *isolate, size_t count,
                               void *listenerContext) {
  V8Engine *e = static_cast<V8Engine *>(listenerContext);

  if (IsEngineLimitsExceeded(e)) {
    TerminateExecution(e);
  }
}

int IsEngineLimitsExceeded(V8Engine *e) {
  // TODO: read memory stats everytime may impact the performance.
  ReadMemoryStatistics(e);
  if (e->limits_of_executed_instructions > 0 &&
      e->limits_of_executed_instructions <
          e->stats.count_of_executed_instructions) {
    // Reach instruction limits.
    return NVM_GAS_LIMIT_ERR;
  } else if (e->limits_of_total_memory_size > 0 &&
             e->limits_of_total_memory_size < e->stats.total_memory_size) {
    // reach memory limits.
    return NVM_MEM_LIMIT_ERR;
  }
  // 
  
  return 0;
}

//loopExecute迁移文件
