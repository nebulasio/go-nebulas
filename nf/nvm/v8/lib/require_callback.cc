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
#include "logger.h"

#include <assert.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <v8.h>

using namespace v8;

#define SOURCE_REQUIRE_LINE_OFFSET -6
static char source_require_format[] = "(function () {\n"
                                      "    var module = {\n"
                                      "        exports: {},\n"
                                      "        id: \"%s\"\n"
                                      "    };\n"
                                      "    (function (exports, module) {\n"
                                      "%s"
                                      ";\n"
                                      "})(module.exports, module);\n"
                                      "    return module.exports;\n"
                                      "})();\n";

static char *getValidFilePath(const char *filename) {
  size_t len = strlen(filename);
  char *ret = NULL;

  if (strncmp(filename, "./", 2) == 0) {
    // Load file from local package.
    ret = (char *)malloc(len + 1);
    stpcpy(ret, filename);
    return ret;

  } else {
    // Load file from lib.
    ret = (char *)malloc(len + 1 + 6);
    strcpy(ret, "./lib/");
    strcpy(ret + 6, filename);
  }

  return ret;
}

static int readSource(const char *filename, char **data, size_t *size) {
  if (strstr(filename, "\"") != NULL) {
    return -1;
  }

  char *filepath = getValidFilePath(filename);
  size_t file_size = 0;
  char *content = readFile(filepath, &file_size);

  if (content == NULL) {
    free(filepath);
    return 1;
  }

  *size = asprintf(data, source_require_format, filepath, content);
  free(filepath);
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
  // logErrorf("require called.");
  Isolate *isolate = info.GetIsolate();

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
  size_t size = 0;
  if (readSource(*filename, &data, &size)) {
    char msg[512];
    snprintf(msg, 512, "require cannot find module '%s'", *filename);
    isolate->ThrowException(
        Exception::Error(String::NewFromUtf8(isolate, msg)));
    return;
  }

  // LogInfof("source is: %s", data);
  Local<Context> context = isolate->GetCurrentContext();

  ScriptOrigin sourceSrcOrigin(
      path, Integer::New(isolate, SOURCE_REQUIRE_LINE_OFFSET));
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

char *EncapsulateSourceToModuleStyle(const char *source,
                                     int *source_line_offset) {
  static const char charset[] =
      "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";

  char genfilename[10];
  genfilename[0] = '_';
  size_t i = 1;
  for (; i < sizeof(genfilename) - 4; i++) {
    genfilename[i] = charset[rand() % (int)(sizeof charset - 1)];
  }
  strncpy(genfilename + i, ".js", 3);

  char *data = NULL;
  asprintf(&data, source_require_format, genfilename, source);
  *source_line_offset = SOURCE_REQUIRE_LINE_OFFSET;
  return data;
}
