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

#include "typescript.h"
#include "logger.h"
#include "util.h"
#include <iostream>

#include <string.h>

extern void PrintException(Local<Context> context, TryCatch &trycatch);

static char ts_transpile_source_template[] =
    "(function(){\n"
    "const tsc = require(\"tsc.js\");\n"
    "const source = \"%s\";\n"
    "return tsc.transpileModule(source);\n"
    "})();";

int TypeScriptTranspileDelegate(char **result, Isolate *isolate,
                                const char *source, int source_line_offset,
                                Local<Context> context, TryCatch &trycatch,
                                void *delegateContext) {
  TypeScriptContext *tContext =
      static_cast<TypeScriptContext *>(delegateContext);
  tContext->js_source = NULL;

  std::string s(source);
  s = ReplaceAll(s, "\\", "\\\\");
  s = ReplaceAll(s, "\n", "\\n");
  s = ReplaceAll(s, "\r", "\\r");
  s = ReplaceAll(s, "\"", "\\\"");

  std::cout<<">>>>>> Created a string containing JS source code ONE"<<std::endl;

  char *runnableSource = NULL;
  asprintf(&runnableSource, ts_transpile_source_template, s.c_str());

  // Create a string containing the JavaScript source code.
  Local<String> src =
      String::NewFromUtf8(isolate, runnableSource, NewStringType::kNormal)
          .ToLocalChecked();
  free(runnableSource);

  std::cout<<">>>>>>  Created a string containing JS source code"<<std::endl;

  // Compile the source code.
  ScriptOrigin sourceSrcOrigin(
      String::NewFromUtf8(isolate, "_tsc_execution.js"),
      Integer::New(isolate, source_line_offset));
  MaybeLocal<Script> script = Script::Compile(context, src, &sourceSrcOrigin);

  std::cout<<">>>>>>  Compiling the source code, while the src is: "<<std::endl;

  if (script.IsEmpty()) {
    PrintException(context, trycatch);
    return 1;
  }

  std::cout<<">>>>>>  Script is not empty, which is: "<<std::endl;

  // Run the script to get the result.
  MaybeLocal<Value> ret = script.ToLocalChecked()->Run(context);

  std::cout<<">>>>>>  ret is empty? "<<ret.IsEmpty()<<std::endl;

  if (ret.IsEmpty()) {
    PrintException(context, trycatch);
    return 1;
  }

  std::cout<<">>>>>>  before local check"<<std::endl;

  Local<Value> checked_ret = ret.ToLocalChecked();
  if (!checked_ret->IsObject()) {
    return 1;
  }

  std::cout<<">>>>>>  After running the source code"<<std::endl;

  Local<Object> obj = Local<Object>::Cast(checked_ret);
  Local<Value> jsSource = obj->Get(String::NewFromUtf8(isolate, "jsSource"));
  Local<Value> lineOffset =
      obj->Get(String::NewFromUtf8(isolate, "lineOffset"));

  if (!jsSource->IsString() || !lineOffset->IsNumber()) {
    LogErrorf("_tsc_execution.js:transpileModule() should return object "
              "with jsSource and lineOffset keys.");
    return 1;
  }

  String::Utf8Value str(jsSource);
  tContext->js_source = (char *)malloc(str.length() + 1);
  strcpy(tContext->js_source, *str);

  tContext->source_line_offset = (int)lineOffset->IntegerValue();

  return 0;
}
