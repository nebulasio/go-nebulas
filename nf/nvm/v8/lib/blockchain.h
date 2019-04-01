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

#ifndef _NEBULAS_NF_NVM_V8_LIB_BLOCKCHAIN_H_
#define _NEBULAS_NF_NVM_V8_LIB_BLOCKCHAIN_H_

#include <v8.h>

using namespace v8;

void NewBlockchainInstance(Isolate *isolate, Local<Context> context,
                           void *handler, uint64_t build_flag);

void BlockchainConstructor(const FunctionCallbackInfo<Value> &info);
void GetTransactionByHashCallback(const FunctionCallbackInfo<Value> &info);
void GetAccountStateCallback(const FunctionCallbackInfo<Value> &info);
void TransferCallback(const FunctionCallbackInfo<Value> &info);
void VerifyAddressCallback(const FunctionCallbackInfo<Value> &info);
void GetPreBlockHashCallback(const FunctionCallbackInfo<Value> &info); 
void GetPreBlockSeedCallback(const FunctionCallbackInfo<Value> &info); 
void GetContractSourceCallback(const FunctionCallbackInfo<Value> &info);
void RunInnerContractSourceCallBack(const FunctionCallbackInfo<Value> &info);
void GetLatestNebulasRankCallback(const FunctionCallbackInfo<Value> &info);
void GetLatestNebulasRankSummaryCallback(const FunctionCallbackInfo<Value> &info);

#endif //_NEBULAS_NF_NVM_V8_LIB_BLOCKCHAIN_H_
