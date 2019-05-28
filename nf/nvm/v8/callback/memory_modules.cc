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
#include "../lib/file.h"

#include <atomic>
#include <sstream>
#include <string>
#include <unordered_map>
#include <vector>
#include <iostream>

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// Basically, this function is used to get rid of "." and ".." in the file path
void reformatModuleId(char *dst, const char *src) {
  std::string s(src);
  std::string delimiter("/");
  std::vector<std::string> paths;

  size_t pos = 0;
  while ((pos = s.find(delimiter)) != std::string::npos) {
    std::string p = s.substr(0, pos);
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

std::string RequireDelegate(void *engineptr, const char *filepath, size_t *lineOffset){
  char sid[128];
  sprintf(sid, "%zu:%s", (uintptr_t)engineptr, filepath);
  std::string ssid(sid);
  *lineOffset = 0;
  std::string res = SNVM::FetchContractSrcFromModules(ssid, lineOffset);
  return res;
}

std::string FetchNativeJSLibContentDelegate(const char* file_path){
  std::string res = SNVM::FetchNativeJSLibContentFromCache(file_path);
  return res;
}

std::string AttachLibVersionDelegate(const char *lib_name) {
  std::string resStr = SNVM::AttachNativeJSLibVersion(lib_name);
  std::cout<<">>>>> attachlib callback: "<<resStr<<std::endl;
  return resStr;
}

void AddModule(void *engineptr, const char *filename, const char *source, size_t lineOffset) {
  char filepath[128];
  if (strncmp(filename, "/", 1) != 0 && strncmp(filename, "./", 2) != 0 &&
      strncmp(filename, "../", 3) != 0) {
    sprintf(filepath, "lib/%s", filename);
    reformatModuleId(filepath, filepath);
  } else {
    reformatModuleId(filepath, filename);
  }

  char sid[128];
  // 140568151729856:lib/blockchain.js
  sprintf(sid, "%zu:%s", (uintptr_t)engineptr, filepath);
  std::string ssid(sid);
  SNVM::AddContractSrcToModules(ssid, source, lineOffset);
}