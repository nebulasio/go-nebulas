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
  char data[] = "const require = (function() {\n"
                "    var requiredLibs = {};\n"
                "    return function(id) {\n"
                "        if (!(id in requiredLibs)) {\n"
                "            requiredLibs[id] = _native_require(id);\n"
                "        }\n"
                "        return requiredLibs[id];\n"
                "    };\n"
                "})();\n"
                "const console = require('console.js');\n"
                "const ContractStorage = require('storage.js');\n"
                "const LocalContractStorage = ContractStorage.lcs;\n"
                "const GlobalContractStorage = ContractStorage.gcs;\n"
                "const BigNumber = require('bignumber.min.js');\n"
                "const Blockchain = require('blockchain.js');\n";

  Local<String> source =
      String::NewFromUtf8(isolate, data, NewStringType::kNormal)
          .ToLocalChecked();

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
