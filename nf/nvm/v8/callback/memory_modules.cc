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

#include "memory_modules.h"
#include "../lib/logger.h"

#include <atomic>
#include <mutex>
#include <sstream>
#include <string>
#include <unordered_map>
#include <vector>

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include <iostream>

using namespace std;

typedef struct {
  string source;
  int lineOffset;
} SourceInfo;

static std::mutex m;
static std::unordered_map<string, SourceInfo> modules;

void reformatModuleId(char *dst, const char *src) {
  string s = src;
  string delimiter = "/";

  vector<string> paths;

  size_t pos = 0;
  while ((pos = s.find(delimiter)) != std::string::npos) {
    string p = s.substr(0, pos);
    s.erase(0, pos + delimiter.length());

    if (p.length() == 0 || p.compare(".") == 0) {
      continue;
    }

    if (p.compare("..") == 0) {
      if (paths.size() > 0) {
        paths.pop_back();
        continue;
      }
    }
    paths.push_back(p);
  }
  paths.push_back(s);

  std::stringstream ss;
  for (size_t i = 0; i < paths.size(); ++i) {
    if (i != 0)
      ss << "/";
    ss << paths[i];
  }

  strcpy(dst, ss.str().c_str());
}

char *RequireDelegate(void *handler, const char *filepath,
                          size_t *lineOffset) {
  
  LogInfof(">>>>> RequireDelegateFunc: %s -> %s", "", filepath);
  std::cout<<"[ ----- CALLBACK ------ ] RequireDelegateFunc: "<<filepath<<std::endl;

  char id[128];
  sprintf(id, "%zu:%s", (uintptr_t)handler, filepath);

  char *ret = NULL;
  *lineOffset = 0;

  m.lock();
  auto it = modules.find(string(id));
  if (it != modules.end()) {
    SourceInfo &srcInfo = it->second;
    ret = (char *)calloc(srcInfo.source.length() + 1, sizeof(char));
    strncpy(ret, srcInfo.source.c_str(), srcInfo.source.length());
    *lineOffset = srcInfo.lineOffset;
  }
  m.unlock();

  return ret;
}

char *AttachLibVersionDelegate(void *handler, const char *libname) {

  NVMCallbackResponse *res = new NVMCallbackResponse();
  res->set_func_name(std::string(ATTACH_LIB_VERSION_DELEGATE_FUNC));
  res->add_func_params(std::string(libname));

  const NVMCallbackResult *callback_res = DataExchangeCallback(handler, res);
  const std::string func_name = callback_res->func_name();
  if(func_name.compare(ATTACH_LIB_VERSION_DELEGATE_FUNC) != 0){
    return nullptr;
  }
  std::string resString = callback_res->result();
  char* path = (char*)calloc(resString.length(), sizeof(char));
  strcpy(path, resString.c_str());

  return path;
}

void AddModule(void *handler, const char *filename, const char *source, int lineOffset) {
  char filepath[128];
  if (strncmp(filename, "/", 1) != 0 && strncmp(filename, "./", 2) != 0 &&
      strncmp(filename, "../", 3) != 0) {
    sprintf(filepath, "lib/%s", filename);
    reformatModuleId(filepath, filepath);
  } else {
    reformatModuleId(filepath, filename);
  }

  char id[128];
  sprintf(id, "%zu:%s", (uintptr_t)handler, filepath);

  if(modules.find(std::string(id)) == modules.end()){
    m.lock();
    SourceInfo srcInfo;
    srcInfo.lineOffset = lineOffset;
    srcInfo.source = source;
    modules[string(id)] = srcInfo;
    m.unlock();
    
    LogDebugf("AddModule: %s -> %s %d", filename, filepath, lineOffset);
    
    if(FG_DEBUG)
      std::cout<<"[ ---- Addmodule ---- ] AddModule: "<<filename<<" --> "<<filepath<<" "<<lineOffset<<std::endl;
  }
}

char *GetModuleSource(void *handler, const char *filename) {
  char filepath[128];
  if (strncmp(filename, "/", 1) != 0 && strncmp(filename, "./", 2) != 0 &&
      strncmp(filename, "../", 3) != 0) {
    sprintf(filepath, "lib/%s", filename);
    reformatModuleId(filepath, filepath);
  } else {
    reformatModuleId(filepath, filename);
  }

  size_t size = 0;
  return RequireDelegate(handler, filepath, &size);
}