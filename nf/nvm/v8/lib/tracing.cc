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

#include "tracing.h"
#include "log_callback.h"

#include <stdio.h>
#include <string.h>

#include <string>

extern void PrintException(Local<Context> context, TryCatch &trycatch);

static char inject_tracer_source_template[] =
    "(function(){\n"
    "const instCounter = require(\"instruction_counter.js\");\n"
    "const source = \"%s\";\n"
    "var traceableSource = instCounter.processScript(source);\n"
    "return traceableSource;\n"
    "})();";

std::string ReplaceAll(std::string str, const std::string &from,
                       const std::string &to) {
  size_t start_pos = 0;
  while ((start_pos = str.find(from, start_pos)) != std::string::npos) {
    str.replace(start_pos, from.length(), to);
    start_pos +=
        to.length(); // Handles case where 'to' is a substring of 'from'
  }
  return str;
}

int InjectTracingInstructionDelegate(Isolate *isolate, const char *data,
                                     Local<Context> context, TryCatch &trycatch,
                                     void *delegateContext) {
  char **output = static_cast<char **>(delegateContext);
  *output = NULL;

  std::string s(data);
  s = ReplaceAll(s, "\\", "\\\\");
  s = ReplaceAll(s, "\n", "\\n");
  s = ReplaceAll(s, "\"", "\\\"");

  char *injectTracerSource = NULL;
  asprintf(&injectTracerSource, inject_tracer_source_template, s.c_str());

  // Create a string containing the JavaScript source code.
  Local<String> source =
      String::NewFromUtf8(isolate, injectTracerSource, NewStringType::kNormal)
          .ToLocalChecked();
  free(injectTracerSource);

  // Compile the source code.
  ScriptOrigin sourceSrcOrigin(
      String::NewFromUtf8(isolate, "_inject_tracer.js"));
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

  Local<Value> checked_ret = ret.ToLocalChecked();
  if (!checked_ret->IsString()) {
    return 1;
  }

  String::Utf8Value str(checked_ret->ToString(isolate));
  *output = (char *)malloc(str.length() + 1);
  strcpy(*output, *str);

  return 0;
}
