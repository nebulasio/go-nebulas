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
#include "v8_data_inc.h"
#include "lib/execution_env.h"
#include "lib/log_callback.h"
#include "lib/require_callback.h"
#include "lib/storage_object.h"

#include <libplatform/libplatform.h>
#include <v8.h>

#include <assert.h>

using namespace v8;

static Platform *platformPtr = NULL;

void PrintException(Local<Context> context, TryCatch &trycatch);

#define STRINGIZE2(s) #s
#define STRINGIZE(s) STRINGIZE2(s)
#define V8VERSION_STRING STRINGIZE(V8_MAJOR_VERSION) \
                        "." STRINGIZE(V8_MINOR_VERSION) \
                        "." STRINGIZE(V8_BUILD_NUMBER) \
                        "." STRINGIZE(V8_PATCH_LEVEL)

static char V8VERSION[] = V8VERSION_STRING;
char *GetV8Version() { return V8VERSION; }

void Initialize() {
  platformPtr = platform::CreateDefaultPlatform();

  // Initialize V8.
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

  V8Engine *e = (V8Engine *)calloc(1, sizeof(V8Engine));
  // e->allocator = allocator;
  e->isolate = isolate;
  return e;
}

void DeleteEngine(V8Engine *e) {
  Isolate *isolate = static_cast<Isolate *>(e->isolate);
  isolate->Dispose();

  delete static_cast<ArrayBuffer::Allocator *>(e->allocator);

  free(e);
}

int RunScriptSource2(V8Engine *e, const char *data, uintptr_t lcsHandler,
                     uintptr_t gcsHandler) {
  return RunScriptSource(e, data, (void *)lcsHandler, (void *)gcsHandler);
}

int RunScriptSource(V8Engine *e, const char *data, void *lcsHandler,
                    void *gcsHandler) {
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

  // Create a new context.
  Local<Context> context = Context::New(isolate, NULL, globalTpl);

  // Disable eval().
  context->AllowCodeGenerationFromStrings(false);

  // Enter the context for compiling and running the hello world script.
  Context::Scope context_scope(context);

  TryCatch trycatch(isolate);

  // Continue put objects to global object.
  NewStorageObject(isolate, context, lcsHandler, gcsHandler);

  // Setup execution env.
  if (SetupExecutionEnv(isolate, context)) {
    // logErrorf("setup execution env failed.");
    PrintException(context, trycatch);
    return 1;
  }

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
    // logErrorf("contract_wrapper.js: compilation fail.");
    PrintException(context, trycatch);
    return 1;
  }

  // Run the script to get the result.
  MaybeLocal<Value> ret = script.ToLocalChecked()->Run(context);
  if (ret.IsEmpty()) {
    // logErrorf("contract_wrapper.js: execution fail.");
    PrintException(context, trycatch);
    return 1;
  }

  // Local<Value> ret_str = ret.ToLocalChecked();
  // String::Utf8Value s(ret_str);
  // logInfof("ret value: %s", *s);
  return 0;
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
