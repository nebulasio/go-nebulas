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
#include "require_callback.h"
#include "../engine.h"
#include "file.h"
#include "global.h"
#include "logger.h"

#include <assert.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <v8.h>

using namespace v8;

static char source_require_format[] =
    "(function(){\n"
    "return function (exports, module, require) {\n"
    "%s\n"
    "};\n"
    "})();\n";

static RequireDelegate sRequireDelegate = NULL;

static int readSource(Local<Context> context, const char *filename, char **data,
                      size_t *lineOffset) {
  if (strstr(filename, "\"") != NULL) {
    return -1;
  }

  *lineOffset = 0;

  char *content = NULL;

  // try sRequireDelegate.
  if (sRequireDelegate != NULL) {
    V8Engine *e = GetV8EngineInstance(context);
    content = sRequireDelegate(e, filename, lineOffset);
  }

  if (content == NULL) {
    size_t file_size = 0;
    content = readFile(filename, &file_size);
    if (content == NULL) {
      return 1;
    }
  }

  asprintf(data, source_require_format, content);
  *lineOffset += -2;
  free(content);

  return 0;
}

void NewNativeRequireFunction(Isolate *isolate,
                              Local<ObjectTemplate> globalTpl) {
  globalTpl->Set(String::NewFromUtf8(isolate, "_native_require"),
                 FunctionTemplate::New(isolate, RequireCallback),
                 static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                                PropertyAttribute::ReadOnly));
}

void RequireCallback(const v8::FunctionCallbackInfo<v8::Value> &info) {
  Isolate *isolate = info.GetIsolate();
  Local<Context> context = isolate->GetCurrentContext();

  if (info.Length() == 0) {
    isolate->ThrowException(
        Exception::Error(String::NewFromUtf8(isolate, "require missing path")));
    return;
  }

  Local<Value> path = info[0];
  if (!path->IsString()) {
    isolate->ThrowException(Exception::Error(
        String::NewFromUtf8(isolate, "require path must be string")));
    return;
  }

  String::Utf8Value filename(path);
  char *data = NULL;
  size_t lineOffset = 0;
  if (readSource(context, *filename, &data, &lineOffset)) {
    char msg[512];
    snprintf(msg, 512, "require cannot find module '%s'", *filename);
    isolate->ThrowException(
        Exception::Error(String::NewFromUtf8(isolate, msg)));
    return;
  }

  ScriptOrigin sourceSrcOrigin(path, Integer::New(isolate, lineOffset));
  MaybeLocal<Script> script = Script::Compile(
      context, String::NewFromUtf8(isolate, data), &sourceSrcOrigin);
  if (!script.IsEmpty()) {
    MaybeLocal<Value> ret = script.ToLocalChecked()->Run(context);
    if (!ret.IsEmpty()) {
      Local<Value> rr = ret.ToLocalChecked();
      info.GetReturnValue().Set(rr);
    }
  }

  free(static_cast<void *>(data));
}

void InitializeRequireDelegate(RequireDelegate delegate) {
  sRequireDelegate = delegate;
}
