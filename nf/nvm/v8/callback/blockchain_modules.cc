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
  bool not_null_flag = callback_res->not_null();

  if(!not_null_flag)
    return NULL;

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
  bool not_null_flag = callback_res->not_null();
  std::string resStr = callback_res->extra(0);
  std::string exceptionInfoStr = callback_res->extra(1);
  *gasCnt = (size_t)std::stoull(callback_res->extra(2));

  // since the string value assigned to exceptionInfoStr is determined on golang side, we can just use length to check if it's NULL or not
  if(exceptionInfoStr.length() > 0){
    *info = (char*)calloc(exceptionInfoStr.length()+1, sizeof(char));
    strcpy(*info, exceptionInfoStr.c_str());
  }else{
    *info = NULL;
  }

  if(not_null_flag){
    *result = (char*)calloc(resStr.length()+1, sizeof(char));
    strcpy(*result, resStr.c_str());
  }else{
    *result = NULL;
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
  bool not_null_flag = callback_res->not_null();
  std::string resStr = callback_res->extra(0);
  std::string exceptionInfoStr = callback_res->extra(1);
  *gasCnt = (size_t)std::stoull(callback_res->extra(2));

  // The string value assigned to exceptionInfoStr is determined on golang side, just use length to check if it's NULL or not
  if(exceptionInfoStr.length() > 0){
    *info = (char*)calloc(exceptionInfoStr.length()+1, sizeof(char));
    strcpy(*info, exceptionInfoStr.c_str());
  }else{
    *info = NULL;
  }

  if(not_null_flag){
    *result = (char*)calloc(resStr.length()+1, sizeof(char));
    strcpy(*result, resStr.c_str());
  }else{
    *result = NULL;
  }
  
  return ret;  
}

int GetPreBlockSeed(void *handler, unsigned long long offset, size_t *gasCnt, char **result, char **info) {

  NVMCallbackResponse *res = new NVMCallbackResponse();
  res->set_func_name(std::string(GET_PRE_BLOCK_SEED));
  res->add_func_params(std::to_string(offset));

  const NVMCallbackResult *callback_res = DataExchangeCallback(handler, res);
  int ret = (int)std::stoi(callback_res->result());
  bool not_null_flag = callback_res->not_null();
  std::string resStr = callback_res->extra(0);
  std::string exceptionInfoStr = callback_res->extra(1);
  *gasCnt = (size_t)std::stoull(callback_res->extra(2));

  // The string value assigned to exceptionInfoStr is determined on golang side, just use length to check if it's NULL or not
  if(exceptionInfoStr.length() > 0){
    *info = (char*)calloc(exceptionInfoStr.length()+1, sizeof(char));
    strcpy(*info, exceptionInfoStr.c_str());
  }else{
    *info = NULL;
  }

  if(not_null_flag){
    *result = (char*)calloc(resStr.length()+1, sizeof(char));
    strcpy(*result, resStr.c_str());
  }else{
    *result = NULL;
  }
    
  return ret;
}

char *GetContractSource(void *handler, 
    const char *address, 
    size_t *gasCnt){

  NVMCallbackResponse *res = new NVMCallbackResponse();
  res->set_func_name(std::string(GET_CONTRACT_SRC));
  res->add_func_params(std::string(address));

  const NVMCallbackResult *callback_res = DataExchangeCallback(handler, res);
  *gasCnt = (size_t)std::stoull(callback_res->extra(0));
  std::string srcStr = callback_res->result();
  bool not_null_flag = callback_res->not_null();

  if(!not_null_flag)
    return nullptr;

  char* ret = (char*)calloc(srcStr.length()+1, sizeof(char));
  strcpy(ret, srcStr.c_str());
  std::cout<<"++++++ The fetched contract source is: "<<ret<<std::endl;

  return ret;
}

char *InnerContract(void *handler, 
    const char *address,
    const char *funcName,
    const char *v,
    const char *args,
    size_t *gasCnt){

  NVMCallbackResponse *res = new NVMCallbackResponse();
  res->set_func_name(std::string(INNER_CONTRACT_CALL));
  res->add_func_params(std::string(address));
  res->add_func_params(std::string(funcName));
  res->add_func_params(std::string(v));
  res->add_func_params(std::string(args));

  const NVMCallbackResult *callback_res = DataExchangeCallback(handler, res);
  *gasCnt = (size_t)std::stoull(callback_res->extra(0));
  std::string resStr = callback_res->result();
  bool not_null_flag = callback_res->not_null();

  if(!not_null_flag)
    return nullptr;

  char* ret = (char*)calloc(resStr.length()+1, sizeof(char));
  strcpy(ret, resStr.c_str());

  // After this, start to create engine and run new contract
  if(resStr.compare("success") == 0){
    // Call handler defined in inner_contract.cc
    
  }

  return ret;
}

int GetLatestNebulasRank(void *handler, 
    const char *addres, 
    size_t *counterVal, 
    char **result, 
    char **info){
  
  int ret = 0;

  return ret;
}

int GetLatestNebulasRankSummary(void *handler, 
    size_t *gasCnt, 
    char **result, 
    char **info){

  int ret = 0;

  return ret;
}