// Copyright (C) 2017-2019 go-nebulas authors
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
// Author: Samuel Chen <samuel.chen@nebulas.io>

#include "execution_env.h"
#include "../engine.h"
#include "file.h"
#include "logger.h"
#include "global.h"

#include <iostream>
#include <string>
#include <cstring>

static AttachLibVersionDelegateFunc alvDelegate = NULL;

int SetupExecutionEnv(Isolate *isolate, Local<Context> &context) {
  std::string verlib;

  if (alvDelegate != NULL) {
    V8Engine *e = GetV8EngineInstance(context);
    verlib = alvDelegate(e, "lib/execution_env.js");
  }
  if (verlib.length() == 0) {
    return 1;
  }

  char path[64] = {0};
  strcat(path, verlib.c_str());
  char *data = readFile(path, NULL);
  if (data == NULL) {
    isolate->ThrowException(Exception::Error(
        String::NewFromUtf8(isolate, "execution_env.js is not found.")));
    return 1;
  }

  Local<String> source =
      String::NewFromUtf8(isolate, data, NewStringType::kNormal)
          .ToLocalChecked();
  free(data);

  // Compile the source code.
  ScriptOrigin sourceSrcOrigin(
      String::NewFromUtf8(isolate, "execution_env.js"));
  MaybeLocal<Script> script =
      Script::Compile(context, source, &sourceSrcOrigin);

  if (script.IsEmpty()) {
    return 1;
  }

  // Run the script to get the result.
  MaybeLocal<Value> v = script.ToLocalChecked()->Run(context);
  if (v.IsEmpty()) {
    return 1;
  }

  return 0;
}

void InitializeExecutionEnvDelegate(AttachLibVersionDelegateFunc aDelegate) {
  alvDelegate = aDelegate;
}