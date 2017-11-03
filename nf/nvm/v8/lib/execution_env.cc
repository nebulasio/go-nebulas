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

#include "execution_env.h"
#include "log_callback.h"

int SetupExecutionEnv(Isolate *isolate, Local<Context> &context) {
  char data[] =
      "const require = (function() {"
      "    var requiredLibs = {};"
      "    return function(filename) {"
      "        if (!(filename in requiredLibs)) {"
      "            requiredLibs[filename] = _native_require(filename);"
      "        }"
      "        return requiredLibs[filename];"
      "    };"
      "})();"
      "const console = require('console.js');"
      "const ContractStorage = require('storage.js');"
      "const LocalContractStorage = ContractStorage.lcs;"
      "const GlobalContractStorage = ContractStorage.gcs;";

  Local<String> source =
      String::NewFromUtf8(isolate, data, NewStringType::kNormal)
          .ToLocalChecked();

  // Compile the source code.
  MaybeLocal<Script> script = Script::Compile(context, source);

  if (script.IsEmpty()) {
    logErrorf("execution-env.js: compile error.");
    return 1;
  }

  // Run the script to get the result.
  MaybeLocal<Value> v = script.ToLocalChecked()->Run(context);
  if (v.IsEmpty()) {
    return 1;
  }
  return 0;
}
