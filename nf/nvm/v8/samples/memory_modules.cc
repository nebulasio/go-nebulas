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

char *RequireDelegateFunc(void *handler, const char *filepath,
                          size_t *lineOffset) {
  // LogInfof("RequireDelegateFunc: %s -> %s", "", filepath);

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

char *AttachLibVersionDelegateFunc(void *handler, const char *libname) {
  char *path = (char *)calloc(128, sizeof(char));
  if (strncmp(libname, "lib/", 4) == 0) {
    sprintf(path, "lib/1.0.0/%s", libname + 4);
  } else {
    sprintf(path, "1.0.0/%s", libname);
  }
   LogDebugf("AttachLibVersion: %s -> %s", libname, path);
   return path;
}

void AddModule(void *handler, const char *filename, const char *source,
               int lineOffset) {
  char filepath[128];
  if (strncmp(filename, "/", 1) != 0 && strncmp(filename, "./", 2) != 0 &&
      strncmp(filename, "../", 3) != 0) {
    sprintf(filepath, "lib/%s", filename);
    reformatModuleId(filepath, filepath);
  } else {
    reformatModuleId(filepath, filename);
  }
  LogDebugf("AddModule: %s -> %s %d", filename, filepath, lineOffset);

  char id[128];
  sprintf(id, "%zu:%s", (uintptr_t)handler, filepath);

  m.lock();
  SourceInfo srcInfo;
  srcInfo.lineOffset = lineOffset;
  srcInfo.source = source;
  modules[string(id)] = srcInfo;
  m.unlock();
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
  // LogInfof("GetModule: %s -> %s", filename, filepath);

  size_t size = 0;
  return RequireDelegateFunc(handler, filepath, &size);
}
