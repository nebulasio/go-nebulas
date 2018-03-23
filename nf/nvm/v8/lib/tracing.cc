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
#include "logger.h"
#include "util.h"

#include <stdio.h>
#include <string.h>

#include <string>

extern void PrintException(Local<Context> context, TryCatch &trycatch);

static char inject_tracer_source_template[] =
    "(function(){\n"
    "const instCounter = require(\"instruction_counter.js\");\n"
    "const source = \"%s\";\n"
    "return instCounter.processScript(source, %d);\n"
    "})();";

int InjectTracingInstructionDelegate(char **result, Isolate *isolate,
                                     const char *source, int source_line_offset,
                                     Local<Context> context, TryCatch &trycatch,
                                     void *delegateContext) {
  TracingContext *tContext = static_cast<TracingContext *>(delegateContext);
  tContext->tracable_source = NULL;

  std::string s(source);
  s = ReplaceAll(s, "\\", "\\\\");
  s = ReplaceAll(s, "\n", "\\n");
  s = ReplaceAll(s, "\r", "\\r");
  s = ReplaceAll(s, "\"", "\\\"");

  char *injectTracerSource = NULL;
  asprintf(&injectTracerSource, inject_tracer_source_template, s.c_str(),
           tContext->strictDisallowUsage);

  // Create a string containing the JavaScript source code.
  Local<String> src =
      String::NewFromUtf8(isolate, injectTracerSource, NewStringType::kNormal)
          .ToLocalChecked();
  free(injectTracerSource);

  // Compile the source code.
  ScriptOrigin sourceSrcOrigin(
      String::NewFromUtf8(isolate, "_inject_tracer.js"),
      Integer::New(isolate, source_line_offset));
  MaybeLocal<Script> script = Script::Compile(context, src, &sourceSrcOrigin);

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
  if (!checked_ret->IsObject()) {
    return 1;
  }

  Local<Object> obj = Local<Object>::Cast(checked_ret);
  Local<Value> traceableSource =
      obj->Get(String::NewFromUtf8(isolate, "traceableSource"));
  Local<Value> lineOffset =
      obj->Get(String::NewFromUtf8(isolate, "lineOffset"));

  if (!traceableSource->IsString() || !lineOffset->IsNumber()) {
    LogErrorf("instruction_counter.js:processScript() should return object "
              "with traceableSource and lineOffset keys.");
    return 1;
  }

  String::Utf8Value str(traceableSource);
  tContext->tracable_source = (char *)malloc(str.length() + 1);
  strcpy(tContext->tracable_source, *str);

  tContext->source_line_offset = (int)lineOffset->IntegerValue();

  return 0;
}
