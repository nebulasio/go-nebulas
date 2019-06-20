// Copyright (C) 2017-2019 go-nebulas authors
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


#include "crypto_modules.h"

char *Sha256(V8Engine* engine, const char *data, size_t *counterVal){
  NVMCallbackResponse *res = new NVMCallbackResponse();
  res->set_func_name(std::string(SHA_256_FUNC));
  res->add_func_params(std::string(data));

  const NVMCallbackResult *callback_res = SNVM::DataExchangeCallback(engine, nullptr, res);
  *counterVal = (size_t)std::stoull(callback_res->extra(0));
  std::string resString = callback_res->result();
  bool not_null_flag = callback_res->not_null();
  if(callback_res != nullptr)
    delete callback_res;

  if(!not_null_flag)
    return NULL;
  char* value = (char*)calloc(resString.length()+1, sizeof(char));
  strcpy(value, resString.c_str());

  return value;
}

char *Sha3256(V8Engine* engine, const char *data, size_t *counterVal){
  NVMCallbackResponse *res = new NVMCallbackResponse();
  res->set_func_name(std::string(SHA_3256_FUNC));
  res->add_func_params(std::string(data));

  const NVMCallbackResult *callback_res = SNVM::DataExchangeCallback(engine, nullptr, res);
  *counterVal = (size_t)std::stoull(callback_res->extra(0));
  std::string resString = callback_res->result();
  bool not_null_flag = callback_res->not_null();
  if(callback_res != nullptr)
    delete callback_res;

  if(!not_null_flag)
    return NULL;
  char* value = (char*)calloc(resString.length()+1, sizeof(char));
  strcpy(value, resString.c_str());

  return value;
}

char *Ripemd160(V8Engine* engine, const char *data, size_t *counterVal){
  NVMCallbackResponse *res = new NVMCallbackResponse();
  res->set_func_name(std::string(RIPEMD_160_FUNC));
  res->add_func_params(std::string(data));

  const NVMCallbackResult *callback_res = SNVM::DataExchangeCallback(engine, nullptr, res);
  *counterVal = (size_t)std::stoull(callback_res->extra(0));
  std::string resString = callback_res->result();
  bool not_null_flag = callback_res->not_null();
   if(callback_res != nullptr)
    delete callback_res;
 
  if(!not_null_flag)
    return NULL;
  char* value = (char*)calloc(resString.length()+1, sizeof(char));
  strcpy(value, resString.c_str());

  return value;
}

char *RecoverAddress(V8Engine* engine, int alg, const char *data, const char *sign, size_t *counterVal){
  NVMCallbackResponse *res = new NVMCallbackResponse();
  res->set_func_name(std::string(RECOVER_ADDRESS_FUNC));
  res->add_func_params(std::to_string(alg));
  res->add_func_params(std::string(data));
  res->add_func_params(std::string(sign));

  const NVMCallbackResult *callback_res = SNVM::DataExchangeCallback(engine, nullptr, res);
  *counterVal = (size_t)std::stoull(callback_res->extra(0));
  std::string resString = callback_res->result();
  bool not_null_flag = callback_res->not_null();
  if(callback_res != nullptr)
    delete callback_res;

  if(!not_null_flag)
    return NULL;
  char* value = (char*)calloc(resString.length()+1, sizeof(char));
  strcpy(value, resString.c_str());

  return value;
}

char *Md5(V8Engine* engine, const char *data, size_t *counterVal){
  NVMCallbackResponse *res = new NVMCallbackResponse();
  res->set_func_name(std::string(MD5_FUNC));
  res->add_func_params(std::string(data));

  const NVMCallbackResult *callback_res = SNVM::DataExchangeCallback(engine, nullptr, res);
  *counterVal = (size_t)std::stoull(callback_res->extra(0));
  std::string resString = callback_res->result();
  bool not_null_flag = callback_res->not_null();
  if(callback_res != nullptr)
    delete callback_res; 

  if(!not_null_flag)
    return NULL;
  char* value = (char*)calloc(resString.length()+1, sizeof(char));
  strcpy(value, resString.c_str());

  return value;
}

char *Base64(V8Engine* engine, const char *data, size_t *counterVal){
  NVMCallbackResponse *res = new NVMCallbackResponse();
  res->set_func_name(std::string(BASE64_FUNC));
  res->add_func_params(std::string(data));

  const NVMCallbackResult *callback_res = SNVM::DataExchangeCallback(engine, nullptr, res);
  *counterVal = (size_t)std::stoull(callback_res->extra(0));
  std::string resString = callback_res->result();
  bool not_null_flag = callback_res->not_null();
  if(callback_res != nullptr)
    delete callback_res;

  if(!not_null_flag)
    return NULL;
  char* value = (char*)calloc(resString.length()+1, sizeof(char));
  strcpy(value, resString.c_str());

  return value;
}