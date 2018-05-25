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

using namespace v8;

static Platform *platformPtr = NULL;

void PrintException(Local<Context> context, TryCatch &trycatch);
void PrintAndReturnException(char **exception, Local<Context> context,
                             TryCatch &trycatch);
void EngineLimitsCheckDelegate(Isolate *isolate, size_t count,
                               void *listenerContext);

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

static uint32_t* ComputeStackLimit(uint32_t size) {
  uint32_t* answer = &size - (size / sizeof(size));
  // If the size is very large and the stack is very near the bottom of
  // memory then the calculation above may wrap around and give an address
  // that is above the (downwards-growing) stack.  In that case we return
  // a very low address.
  // printf("answer:%p-size:%p\n", answer, &size);
  if (answer > &size) {
      return reinterpret_cast<uint32_t*>(sizeof(size));
  }
  return answer;
}

V8Engine *CreateEngine() {
  ArrayBuffer::Allocator *allocator = new ArrayBufferAllocator();

  Isolate::CreateParams create_params;
  create_params.array_buffer_allocator = allocator;

  Isolate *isolate = Isolate::New(create_params);
  // char *newArgv[] = {"--timeout=2"};
  // v8::V8::SetFlagsFromCommandLine(&newArgc, newArgv, true);
  // uint32_t* set_limit = ComputeStackLimit(1 * 1024 * 1024);
  // printf("limit:%p\n", set_limit);

  // isolate->SetStackLimit(reinterpret_cast<uintptr_t>(set_limit));
  // Locker locker(isolate);
  // Isolate::Scope isolate_scope(isolate);
  // // Create a stack-allocated handle scope.
  // HandleScope handle_scope(isolate);
  // // fix bug: https://github.com/nebulasio/go-nebulas/issues/5
  // isolate->SetStackLimit(0x700000000000UL);
  // Local<ObjectTemplate> globalTpl = CreateGlobalObjectTemplate(isolate);

  // // Create a new context.

  // Local<Context> context = Context::New(isolate, NULL, globalTpl);
  // // disable eval().
  // context->AllowCodeGenerationFromStrings(false);

  V8Engine *e = (V8Engine *)calloc(1, sizeof(V8Engine));
  e->allocator = allocator;
  e->isolate = isolate;
  // e->context.Reset(isolate, context);
  // Persistent<Context> *pc = (Persistent<Context>*)e->context;
  // pc->Reset(isolate, context);
  // e->context = context;
  return e;
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
  // printf("test fe execute\n");                                
  // char str[3] = "fe";
  // *result = (char *)malloc(sizeof(str) + 1);
  // strcpy(*result, str);                                
  // return 0;
  // Create a string containing the JavaScript source code.
  Locker locker(isolate);
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
    return 1;
  }

  // Run the script to get the result.
  MaybeLocal<Value> ret = script.ToLocalChecked()->Run(context);
  if (ret.IsEmpty()) {
    PrintAndReturnException(result, context, trycatch);
    return 1;
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
  // isolate->Exit();
  return 0;
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
  // printf("return RunScriptSource\n");                      
  // *result = (char *)malloc(10);
  // strcpy(*result, "fe");                      
  // return 0;                      
  return Execute(result, e, source, source_line_offset, (void *)lcsHandler,
                (void *)gcsHandler, ExecuteSourceDataDelegate, NULL);
  // TracingContext tContext;
  // tContext.source_line_offset = 0;
  // tContext.tracable_source = NULL;
  // tContext.strictDisallowUsage = 1;
  // printf("test insert code\n");
  // char str[1 * 1024 * 1024];
  // memset(str, 0x01, sizeof(str));
  // uint32_t a= 0;
  // for (int i =0; i < sizeof(str); ++i) {
  //   a += str[i];
  // }
  // printf("a:%u\n", a);
  // Execute(NULL, e, source, 0, 0L, 0L, InjectTracingInstructionDelegate,
  //         (void *)&tContext);
  // return 0;
  // TracingContext tContext;
  // tContext.source_line_offset = 0;
  // tContext.tracable_source = NULL;
  // tContext.strictDisallowUsage = 1;

  // Execute(NULL, e, source, 0, 0L, 0L, InjectTracingInstructionDelegate,
  //         (void *)&tContext);
}

int Execute(char **result, V8Engine *e, const char *source,
            int source_line_offset, void *lcsHandler, void *gcsHandler,
            ExecutionDelegate delegate, void *delegateContext) {   
  
  Isolate *isolate = static_cast<Isolate *>(e->isolate);
  Locker locker(isolate);
  // Locker locker(isolate);
  // isolate->Enter();
  assert(isolate);
  
  Isolate::Scope isolate_scope(isolate);
  // Create a stack-allocated handle scope.
  HandleScope handle_scope(isolate);
  
  // return 0;
  // Create global object template.
  Local<ObjectTemplate> globalTpl = CreateGlobalObjectTemplate(isolate);
  if (result != 0x00) {
    printf("begin fe run\n");
    //   *result = (char *)malloc(10);
    // strcpy(*result, "fe");  
    // return 0;
  } else {
    printf("begin insert fe \n");
  }    
  // // Create a new context.
  Local<Context> context = Context::New(isolate, NULL, globalTpl);
  // if (result != 0x00) {
  //   printf("exit fe 3\n");
  //   *result = (char *)malloc(10);
  //   strcpy(*result, "fe");  
  //   return 0;
  // } else {
  //   printf("+++++insert code\n");
  // }    
  // isolate->
  // disable eval().
  context->AllowCodeGenerationFromStrings(false);

  // Enter the context for compiling and running the script.
  // Persistent<Context> *pc = (Persistent<Context>*)e->context;
  // Local<Context> context = Local<Context>::New(isolate, );
  // Local<Context> context(isolate->GetCurrentContext());
  Context::Scope context_scope(context);
  if (result != 0x00) {
    // printf("exit fe 2\n");
    // *result = (char *)malloc(10);
    // strcpy(*result, "fe");  
    // return 0;
  }     
  TryCatch trycatch(isolate);

  // Continue put objects to global object.
  SetGlobalObjectProperties(isolate, context, e, lcsHandler, gcsHandler);

  // Setup execution env.
  if (SetupExecutionEnv(isolate, context)) {
    PrintAndReturnException(result, context, trycatch);
    return 1;
  }
  if (result != 0x00) {
    // printf("exit fe 3\n");
    // *result = (char *)malloc(10);
    // strcpy(*result, "fe");  
    // return 0;
  }           

  return delegate(result, isolate, source, source_line_offset, context,
                  trycatch, delegateContext);
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
    strcpy(*exception, *exception_str);
  }
}

void ReadMemoryStatistics(V8Engine *e) {
  Isolate *isolate = static_cast<Isolate *>(e->isolate);
  // Locker locker(isolate);
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
  // Locker locker(isolate);
  isolate->TerminateExecution();
  e->is_requested_terminate_execution = 1;
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
    return 1;
  } else if (e->limits_of_total_memory_size > 0 &&
             e->limits_of_total_memory_size < e->stats.total_memory_size) {
    // reach memory limits.
    return 2;
  }

  return 0;
}
