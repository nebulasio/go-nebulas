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
#include "memory_storage.h"

#include <atomic>
#include <mutex>
#include <string>
#include <unordered_map>

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

using namespace std;

static std::mutex mapMutex;
static unordered_map<string, string> memoryMap;
static atomic<uintptr_t> handlerCounter(1234);

string genKey(void *handler, const char *key) {
  uintptr_t prefix = (uintptr_t)handler;
  string sKey = to_string(prefix);
  sKey.append("-");
  sKey.append(key);
  return sKey;
}

void *CreateStorageHandler() { return (void *)handlerCounter.fetch_add(1); }

void DeleteStorageHandler(void *handler) {}

char *StorageGet(void *handler, const char *key, size_t *cnt) {
  char *ret = NULL;
  string sKey = genKey(handler, key);

  mapMutex.lock();
  auto it = memoryMap.find(sKey);
  if (it != memoryMap.end()) {
    string &value = it->second;
    ret = (char *)calloc(value.length() + 1, sizeof(char));
    strncpy(ret, value.c_str(), value.length());
  }
  mapMutex.unlock();

  *cnt = 0;

  return ret;
}

int StoragePut(void *handler, const char *key, const char *value, size_t *cnt) {
  string sKey = genKey(handler, key);

  mapMutex.lock();
  memoryMap[sKey] = string(value);
  mapMutex.unlock();

  *cnt = strlen(key) + strlen(value);
  return 0;
}

int StorageDel(void *handler, const char *key, size_t *cnt) {
  string sKey = genKey(handler, key);

  mapMutex.lock();
  memoryMap.erase(sKey);
  mapMutex.unlock();

  *cnt = 0;

  return 0;
}
