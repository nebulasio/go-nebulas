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
#include "log_callback.h"

#include <assert.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <v8.h>

using namespace v8;

static char source_require_begin[] = "(function(){"
                                     "var module = {};";

static char source_require_end[] = ";if (module.exports) {"
                                   " return module.exports;"
                                   "} else if (exports) {"
                                   " return exports;"
                                   "}"
                                   "})();";

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
  char *path = getValidFilePath(filename);

  // char cwd[1024];
  // getcwd(cwd, 1024);
  // logInfof("fullpath is %s/%s", cwd, path);

  FILE *f = fopen(path, "r");
  if (f == NULL) {
    logErrorf("file %s does not found.", filename);
    free(path);
    return 1;
  }
  free(path);

  *size = 10 * 1024 * 1024;
  *data = (char *)malloc(*size);
  size_t idx = 0;

  // Prepare the source.
  idx += sprintf(*data, "%s", source_require_begin);

  size_t len = 0;
  while ((len = fread(*data + idx, sizeof(char), *size - idx, f)) > 0) {
    idx += len;
    if (*size - idx < 128) {
      *size *= 2;
      *data = (char *)realloc(*data, *size);
    }
  }

  if (feof(f) == 0) {
    free(static_cast<void *>(*data));
    logErrorf("read file %s error.", filename);
    return 1;
  }

  fclose(f);

  size_t source_end_len = strlen(source_require_end);
  if (*size - idx < source_end_len) {
    *size = idx + source_end_len + 1;
    *data = (char *)realloc(*data, *size);
  }
  idx += sprintf(*data + idx, "%s", source_require_end);

  return 0;
}

void requireCallback(const v8::FunctionCallbackInfo<v8::Value> &info) {
  // logErrorf("require called.");
  Isolate *isolate = info.GetIsolate();

  if (info.Length() == 0) {
    isolate->ThrowException(
        String::NewFromUtf8(isolate, "require missing path"));
    return;
  }

  Local<Value> path = info[0];
  if (!path->IsString()) {
    isolate->ThrowException(
        String::NewFromUtf8(isolate, "require path must be string"));
    return;
  }

  String::Utf8Value filename(path);
  char *data = NULL;
  size_t size = 0;
  if (readSource(*filename, &data, &size)) {
    char msg[1024];
    snprintf(msg, 1024, "read content of path error. %s", *filename);
    isolate->ThrowException(String::NewFromUtf8(isolate, msg));
    return;
  }

  // logErrorf("source is: %s", data);
  ScriptOrigin sourceSrcOrigin(path);
  Local<Context> context = isolate->GetCurrentContext();
  Local<Script> script =
      Script::Compile(context, String::NewFromUtf8(isolate, data),
                      &sourceSrcOrigin)
          .ToLocalChecked();
  MaybeLocal<Value> ret = script->Run(context);
  if (!ret.IsEmpty()) {
    Local<Value> rr = ret.ToLocalChecked();
    info.GetReturnValue().Set(rr);
  }

  free(static_cast<void *>(data));
}

char *encapsulateSourceToModuleStyle(const char *source) {
  size_t size = strlen(source) + strlen(source_require_begin) +
                strlen(source_require_end) + 1;
  char *data = (char *)malloc(size);

  size_t count =
      sprintf(data, "%s%s%s", source_require_begin, source, source_require_end);
  assert(count + 1 == size);

  return data;
}
