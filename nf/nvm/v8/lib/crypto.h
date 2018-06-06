// Copyright (C) 2018 go-nebulas authors
// 
// This file is part of the go-nebulas library.
// 
// the go-nebulas library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
// 
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
// 
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see <http://www.gnu.org/licenses/>.
// 

#ifndef _NEBULAS_NF_NVM_V8_LIB_CRYPTO_H_
#define _NEBULAS_NF_NVM_V8_LIB_CRYPTO_H_

#include <v8.h>

using namespace v8;

void NewCryptoInstance(Isolate *isolate, Local<Context> context);

void Sha256Callback(const FunctionCallbackInfo<Value> &info);
void Sha3256Callback(const FunctionCallbackInfo<Value> &info);
void Ripemd160Callback(const FunctionCallbackInfo<Value> &info);
void RecoverAddressCallback(const FunctionCallbackInfo<Value> &info);
void Md5Callback(const FunctionCallbackInfo<Value> &info);
void Base64Callback(const FunctionCallbackInfo<Value> &info);

#endif //_NEBULAS_NF_NVM_V8_LIB_CRYPTO_H_