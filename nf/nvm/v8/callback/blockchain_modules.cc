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

#include "blockchain_modules.h"

#include <stdio.h>
#include <stdlib.h>
#include <string>

#include <string.h>

using namespace std;

char *GetTxByHash(void *handler, const char *hash, size_t *gasCnt) {

  NVMCallbackResponse *res = new NVMCallbackResponse();
  res->set_func_name(std::string(GET_TX_BY_HASH));
  res->add_func_params(std::string(hash));

  const NVMCallbackResult *callback_res = DataExchangeCallback(handler, res);
  *gasCnt = (size_t)std::stoull(callback_res->extra(0));
  std::string resStr = callback_res->result();
  char* ret = (char*)calloc(resStr.length()+1, sizeof(char));
  strcpy(ret, resStr.c_str());

  return ret;
}

int GetAccountState(void *handler, const char *address, size_t *gasCnt, char **result, char **info) {
  
  NVMCallbackResponse *res = new NVMCallbackResponse();
  res->set_func_name(std::string(GET_ACCOUNT_STATE));
  res->add_func_params(std::string(address));

  const NVMCallbackResult *callback_res = DataExchangeCallback(handler, res);
  int ret = (int)std::stoi(callback_res->result());
  std::string resStr = callback_res->extra(0);
  std::string infoStr = callback_res->extra(1);
  *gasCnt = (size_t)std::stoull(callback_res->extra(2));

  *result = (char*)calloc(resStr.length()+1, sizeof(char));
  strcpy(*result, resStr.c_str());
  if(infoStr.length()>0){
    *info = (char*)calloc(infoStr.length()+1, sizeof(char));
    strcpy(*info, infoStr.c_str());
  }

  return ret;
}

int Transfer(void *handler, const char *to, const char *value, size_t *gasCnt) {

  NVMCallbackResponse *res = new NVMCallbackResponse();
  res->set_func_name(std::string(TRANSFER));
  res->add_func_params(std::string(to));
  res->add_func_params(std::string(value));

  const NVMCallbackResult *callback_res = DataExchangeCallback(handler, res);
  *gasCnt = (size_t)std::stoull(callback_res->extra(0));
  int ret = (int)std::stoi(callback_res->result());
  
  return ret;
}

int VerifyAddress(void *handler, const char *address, size_t *gasCnt) {

  NVMCallbackResponse *res = new NVMCallbackResponse();
  res->set_func_name(std::string(VERIFY_ADDR));
  res->add_func_params(std::string(address));

  const NVMCallbackResult *callback_res = DataExchangeCallback(handler, res);
  *gasCnt = (size_t)std::stoull(callback_res->extra(0));
  int ret = (int)std::stoi(callback_res->result());
  
  return ret;
}

int GetPreBlockHash(void *handler, unsigned long long offset, size_t *gasCnt, char **result, char **info) {

  NVMCallbackResponse *res = new NVMCallbackResponse();
  res->set_func_name(std::string(GET_PRE_BLOCK_HASH));
  res->add_func_params(std::to_string(offset));

  const NVMCallbackResult *callback_res = DataExchangeCallback(handler, res);
  int ret = (int)std::stoi(callback_res->result());
  std::string resStr = callback_res->extra(0);
  std::string infoStr = callback_res->extra(1);
  *result = (char*)calloc(resStr.length()+1, sizeof(char));
  strcpy(*result, resStr.c_str());
  if(infoStr.length() > 0){
    *info = (char*)calloc(infoStr.length()+1, sizeof(char));
    strcpy(*info, infoStr.c_str());
  }
  *gasCnt = (size_t)std::stoull(callback_res->extra(2));
  
  return ret;  
}

int GetPreBlockSeed(void *handler, unsigned long long offset, size_t *gasCnt, char **result, char **info) {

  NVMCallbackResponse *res = new NVMCallbackResponse();
  res->set_func_name(std::string(GET_PRE_BLOCK_SEED));
  res->add_func_params(std::to_string(offset));

  const NVMCallbackResult *callback_res = DataExchangeCallback(handler, res);
  int ret = (int)std::stoi(callback_res->result());
  std::string resStr = callback_res->extra(0);
  std::string infoStr = callback_res->extra(1);
  *result = (char*)calloc(resStr.length()+1, sizeof(char));
  strcpy(*result, resStr.c_str());
  if(infoStr.length() > 0){
    *info = (char*)calloc(infoStr.length()+1, sizeof(char));
    strcpy(*info, infoStr.c_str());
  }
  *gasCnt = (size_t)std::stoull(callback_res->extra(2));
  
  return ret;
}

