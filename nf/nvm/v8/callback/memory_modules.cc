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
// Author: Samuel Chen <samuel.chen@nebulas.io>

#include "memory_modules.h"
#include "../lib/logger.h"
#include "../lib/file.h"

#include <atomic>
#include <sstream>
#include <string>
#include <unordered_map>
#include <vector>
#include <iostream>
#include <stdio.h>
#include <stdlib.h>

std::string RequireDelegate(V8Engine *engine, const char *filepath, size_t *lineOffset){
  char sid[128];
  sprintf(sid, "%zu:%s", (uintptr_t)engine, filepath);
  std::string ssid(sid);
  *lineOffset = 0;
  std::string res = SNVM::FetchEngineContractSrcFromModules(engine, ssid, lineOffset);
  return res;
}

std::string FetchNativeJSLibContentDelegate(const char* file_path){
  std::string res = SNVM::FetchNativeJSLibContentFromCache(file_path);
  return res;
}

std::string AttachLibVersionDelegate(V8Engine* engine, const char *lib_name) {
  std::string resStr = SNVM::AttachNativeJSLibVersion(engine, lib_name);
  std::cout<<">>>>> attachlib callback: "<<resStr<<std::endl;
  return resStr;
}