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

char* StorageGet(void* handler, const char *key, size_t *cnt){
  NVMCallbackResponse *res = new NVMCallbackResponse();
  res->set_func_name(std::string(STORAGE_GET));
  res->add_func_params(std::string(key));

  const NVMCallbackResult *result = DataExchangeCallback(res);
  const char* resString = result->res().c_str();
  char* cstr = new char[result->res().length() + 1];
  strcpy(cstr, resString);
  *cnt = (size_t)std::stoll(result->extra(0));

  return cstr;
}

int StoragePut(void* handler, const char* key, const char *value, size_t *cnt){
  NVMCallbackResponse *res = new NVMCallbackResponse();
  res->set_func_name(std::string(STORAGE_PUT));
  res->add_func_params(std::string(key));
  res->add_func_params(std::string(value));

  const NVMCallbackResult* result = DataExchangeCallback(res);
  *cnt = (size_t)std::stoll(result->extra(0));
  int resCode = std::stoi(result->res());

  return resCode;
}

int StorageDel(void* handler, const char* key, size_t *cnt){
  NVMCallbackResponse *res = new NVMCallbackResponse();
  res->set_func_name(std::string(STORAGE_DEL));
  res->add_func_params(std::string(key));

  const NVMCallbackResult* result = DataExchangeCallback(res);
  *cnt = (size_t)std::stoll(result->extra(0));
  int resCode = std::stoi(result->res());

  return resCode;
}
