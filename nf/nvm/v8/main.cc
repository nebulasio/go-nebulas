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

#include <libplatform/libplatform.h>
#include <v8.h>

#include <assert.h>
#include <stdio.h>
#include <stdlib.h>

using namespace v8;

void help(const char *name) {
  printf("%s [Javascript File] [Func Name] [Args...]\n", name);
  exit(1);
}

void readSource(const char *filename, char **data, size_t *size) {
  FILE *f = fopen(filename, "r");
  if (f == NULL) {
    fprintf(stderr, "file %s does not found.\n", filename);
    exit(1);
    return;
  }

  *size = 10 * 1024 * 1024;
  *data = (char *)malloc(*size);

  size_t len = 0;
  size_t idx = 0;
  while ((len = fread(*data + idx, sizeof(char), *size - idx, f)) > 0) {
    idx += len;
    if (*size - idx < 1024) {
      *size *= 1.5;
      *data = (char *)realloc(*data, *size);
    }
  }
  // size_t ret = fread(*data, sizeof(char), *size, f);
  // assert(ret == *size);
  // // printf("size = %zu, ret = %zu\n", *size, ret);

  if (feof(f) == 0) {
    fprintf(stderr, "read file %s error.\n", filename);
    exit(1);
  }

  fclose(f);
}

class ArrayBufferAllocator : public v8::ArrayBuffer::Allocator {
public:
  virtual void *Allocate(size_t length) {
    void *data = AllocateUninitialized(length);
    return data == NULL ? data : memset(data, 0, length);
  }
  virtual void *AllocateUninitialized(size_t length) { return malloc(length); }
  virtual void Free(void *data, size_t length) { free(data); }
};

/*
int main1(int argc, const char *argv[]) {
  if (argc < 3) {
    help(argv[0]);
  }

  const char *filename = argv[1];
  const char *funcName = argv[2];

  // Initialize v8.
  V8::InitializeICU();
  V8::InitializeExternalStartupData(argv[0]);
  Platform *platform = platform::CreateDefaultPlatform();
  V8::InitializePlatform(platform);
  V8::Initialize();

  // Create a new Isolate and make it the current one.
  ArrayBufferAllocator allocator;
  Isolate::CreateParams create_params;
  create_params.array_buffer_allocator = &allocator;
  Isolate *isolate = Isolate::New(create_params);

  char *data = NULL;
  size_t size = 0;
  readSource(filename, &data, &size);

  data = "'Hello' + ', World!'";
  {
    Isolate::Scope isolate_scope(isolate);

    HandleScope handle_scope(isolate);
    Local<Context> context = Context::New(isolate);
    Local<String> source =
        String::NewFromUtf8(isolate, data, NewStringType::kNormal)
            .ToLocalChecked();
    String::Utf8Value s(source);
    printf("source is %s, len is %d\n", *s, source->Length());

    Local<Script> script = Script::Compile(context, source).ToLocalChecked();
    Local<Value> result = script->Run(context).ToLocalChecked();

    // printf("result is %s\n", *utf8);
  }

  // Dispose.
  isolate->Dispose();
  V8::Dispose();
  V8::ShutdownPlatform();
  delete platform;

  return 0;
}
*/

int main(int argc, char *argv[]) {
  if (argc < 3) {
    help(argv[0]);
  }
  const char *filename = argv[1];
  const char *funcName = argv[2];

  // Initialize V8.
  V8::InitializeICU();
  V8::InitializeExternalStartupData("");
  Platform *platform = platform::CreateDefaultPlatform();
  V8::InitializePlatform(platform);
  V8::Initialize();
  // Create a new Isolate and make it the current one.
  ArrayBufferAllocator *allocator = new ArrayBufferAllocator();
  Isolate::CreateParams create_params;
  create_params.array_buffer_allocator = allocator;
  Isolate *isolate = Isolate::New(create_params);

  char *data = NULL;
  size_t size = 0;
  readSource(filename, &data, &size);

  {
    Isolate::Scope isolate_scope(isolate);
    // Create a stack-allocated handle scope.
    HandleScope handle_scope(isolate);
    // Create a new context.
    Local<Context> context = Context::New(isolate);
    // Enter the context for compiling and running the hello world script.
    Context::Scope context_scope(context);
    // Create a string containing the JavaScript source code.
    Local<String> source =
        String::NewFromUtf8(isolate, data, NewStringType::kNormal)
            .ToLocalChecked();
    // Compile the source code.
    MaybeLocal<Script> s = Script::Compile(context, source);
    if (s.IsEmpty()) {
      fprintf(stderr, "compile error.\n");
      return 1;
    }
    Local<Script> script = s.ToLocalChecked();
    // Run the script to get the result.
    MaybeLocal<Value> ret = script->Run(context);
    if (!ret.IsEmpty()) {
      Local<Value> ret_str = ret.ToLocalChecked();
      String::Utf8Value s(ret_str);
      fprintf(stdout, "ret value: %s\n", *s);
    }
  }

  free(data);

  // Dispose the isolate and tear down V8.
  isolate->Dispose();
  V8::Dispose();
  V8::ShutdownPlatform();
  delete platform;
  delete allocator;
  return 0;
}
