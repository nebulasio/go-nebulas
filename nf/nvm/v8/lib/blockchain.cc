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

#include "blockchain.h"
#include "global.h"
#include "../engine.h"
#include "global.h"
#include "instruction_counter.h"
#include "logger.h"
#include "limits.h"

static GetTxByHashFunc sGetTxByHash = NULL;
static GetAccountStateFunc sGetAccountState = NULL;
static TransferFunc sTransfer = NULL;
static VerifyAddressFunc sVerifyAddress = NULL;
static GetPreBlockHashFunc sGetPreBlockHash = NULL;
static GetPreBlockSeedFunc sGetPreBlockSeed = NULL;
static GetContractSourceFunc sGetContractSource = NULL;
static InnerContractFunc sRunInnerContract = NULL;
static GetLatestNebulasRankFunc sGetLatestNR = NULL;
static GetLatestNebulasRankSummaryFunc sGetLatestNRSum = NULL;

void InitializeBlockchain(GetTxByHashFunc getTx, GetAccountStateFunc getAccount,
                          TransferFunc transfer,
                          VerifyAddressFunc verifyAddress,
                          GetPreBlockHashFunc getPreBlockHash,
                          GetPreBlockSeedFunc getPreBlockSeed,
                          GetContractSourceFunc contractSource,
                          InnerContractFunc rMultContract,
                          GetLatestNebulasRankFunc getLatestNR,
                          GetLatestNebulasRankSummaryFunc getLatestNRSum) {
  sGetTxByHash = getTx;
  sGetAccountState = getAccount;
  sTransfer = transfer;
  sVerifyAddress = verifyAddress;
  sGetPreBlockHash = getPreBlockHash;
  sGetPreBlockSeed = getPreBlockSeed;
  sGetContractSource = contractSource;
  sRunInnerContract = rMultContract;
  sGetLatestNR = getLatestNR;
  sGetLatestNRSum = getLatestNRSum;
}

void NewBlockchainInstance(Isolate *isolate, Local<Context> context,
                           void *handler, uint64_t build_flag) {
  Local<ObjectTemplate> blockTpl = ObjectTemplate::New(isolate);
  blockTpl->SetInternalFieldCount(1);

  /* disable getTransactionByHash() function.
    blockTpl->Set(String::NewFromUtf8(isolate, "getTransactionByHash"),
                  FunctionTemplate::New(isolate, GetTransactionByHashCallback),
                  static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                                 PropertyAttribute::ReadOnly));
  */

  blockTpl->Set(String::NewFromUtf8(isolate, "getAccountState"),
                FunctionTemplate::New(isolate, GetAccountStateCallback),
                static_cast<PropertyAttribute>(PropertyAttribute::DontDelete|
                                                PropertyAttribute::ReadOnly));


  blockTpl->Set(String::NewFromUtf8(isolate, "transfer"),
                FunctionTemplate::New(isolate, TransferCallback),
                static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                               PropertyAttribute::ReadOnly));

  blockTpl->Set(String::NewFromUtf8(isolate, "verifyAddress"),
                FunctionTemplate::New(isolate, VerifyAddressCallback),
                static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                               PropertyAttribute::ReadOnly));
  if (BUILD_BLOCKCHAIN_GET_RUN_SOURCE == (build_flag & BUILD_BLOCKCHAIN_GET_RUN_SOURCE)) {
    // printf("load getContractSource\n");
    blockTpl->Set(String::NewFromUtf8(isolate, "getContractSource"),
                FunctionTemplate::New(isolate, GetContractSourceCallback),
                static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                               PropertyAttribute::ReadOnly | 
                                               PropertyAttribute::DontEnum));
  }
  
  if (BUILD_BLOCKCHAIN_RUN_CONTRACT == (build_flag & BUILD_BLOCKCHAIN_RUN_CONTRACT)) {
    // printf("load runContractSource\n");
    blockTpl->Set(String::NewFromUtf8(isolate, "runContractSource"),
                FunctionTemplate::New(isolate, RunInnerContractSourceCallBack),
                static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                               PropertyAttribute::ReadOnly |
                                               PropertyAttribute::DontEnum));
  }
  

  blockTpl->Set(String::NewFromUtf8(isolate, "getPreBlockHash"),
              FunctionTemplate::New(isolate, GetPreBlockHashCallback),
              static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                              PropertyAttribute::ReadOnly));

  blockTpl->Set(String::NewFromUtf8(isolate, "getPreBlockSeed"),
              FunctionTemplate::New(isolate, GetPreBlockSeedCallback),
              static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                              PropertyAttribute::ReadOnly));

  blockTpl->Set(String::NewFromUtf8(isolate, "getLatestNebulasRank"),
              FunctionTemplate::New(isolate, GetLatestNebulasRankCallback),
              static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                              PropertyAttribute::ReadOnly));

  blockTpl->Set(String::NewFromUtf8(isolate, "getLatestNebulasRankSummary"),
              FunctionTemplate::New(isolate, GetLatestNebulasRankSummaryCallback),
              static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                              PropertyAttribute::ReadOnly));


  Local<Object> instance = blockTpl->NewInstance(context).ToLocalChecked();
  instance->SetInternalField(0, External::New(isolate, handler));

  context->Global()->DefineOwnProperty(
      context, String::NewFromUtf8(isolate, "_native_blockchain"), instance,
      static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                     PropertyAttribute::ReadOnly));
}

// GetTransactionByHashCallback
void GetTransactionByHashCallback(const FunctionCallbackInfo<Value> &info) {
  Isolate *isolate = info.GetIsolate();
  Local<Object> thisArg = info.Holder();
  Local<External> handler = Local<External>::Cast(thisArg->GetInternalField(0));

  if (info.Length() != 1) {
    isolate->ThrowException(String::NewFromUtf8(
        isolate, "Blockchain.getTransactionByHash() requires only 1 argument"));
    return;
  }

  Local<Value> key = info[0];
  if (!key->IsString()) {
    isolate->ThrowException(String::NewFromUtf8(isolate, "key must be string"));
    return;
  }

  size_t cnt = 0;

  char *value =
      sGetTxByHash(handler->Value(), *String::Utf8Value(key->ToString()), &cnt);
  if (value == NULL) {
    info.GetReturnValue().SetNull();
  } else {
    info.GetReturnValue().Set(String::NewFromUtf8(isolate, value));
    free(value);
  }

  // record storage usage.
  IncrCounter(isolate, isolate->GetCurrentContext(), cnt);
}

// GetAccountStateCallback
void GetAccountStateCallback(const FunctionCallbackInfo<Value> &info) {
  int err = NVM_SUCCESS;
  Isolate *isolate = info.GetIsolate();
  if (NULL == isolate) {
    LogFatalf("Unexpected error: failed to get ioslate");
  }
  Local<Object> thisArg = info.Holder();
  Local<External> handler = Local<External>::Cast(thisArg->GetInternalField(0));//TODO:

  if (info.Length() != 1) {
    isolate->ThrowException(String::NewFromUtf8(
        isolate, "Blockchain.getAccountState() requires only 1 argument"));
    return;
  }

  Local<Value> key = info[0];
  if (!key->IsString()) {
    isolate->ThrowException(String::NewFromUtf8(isolate, "Blockchain.getAccountState(), argument must be a string"));
    return;
  }


  size_t cnt = 0;
  char *result = NULL;
  char *exceptionInfo = NULL;
  err = sGetAccountState(handler->Value(), *String::Utf8Value(key->ToString()), &cnt, &result, &exceptionInfo);

  DEAL_ERROR_FROM_GOLANG(err);

  if (result != NULL) {
    free(result);
    result = NULL;
  }

  if (exceptionInfo != NULL) {
    free(exceptionInfo);
    exceptionInfo = NULL;
  }

  // record storage usage.
  IncrCounter(isolate, isolate->GetCurrentContext(), cnt);
}

// TransferCallback
void TransferCallback(const FunctionCallbackInfo<Value> &info) {
  Isolate *isolate = info.GetIsolate();
  Local<Object> thisArg = info.Holder();
  Local<External> handler = Local<External>::Cast(thisArg->GetInternalField(0));

  if (info.Length() != 2) {
    isolate->ThrowException(
        String::NewFromUtf8(isolate, "Blockchain.send() requires 2 arguments"));
    return;
  }

  Local<Value> address = info[0];
  if (!address->IsString()) {
    isolate->ThrowException(
        String::NewFromUtf8(isolate, "address must be string"));
    return;
  }

  Local<Value> amount = info[1];
  if (!amount->IsString()) {
    isolate->ThrowException(
        String::NewFromUtf8(isolate, "value must be string"));
    return;
  }

  size_t cnt = 0;

  int ret = sTransfer(handler->Value(), *String::Utf8Value(address->ToString()),
                      *String::Utf8Value(amount->ToString()), &cnt);
  info.GetReturnValue().Set(ret);

  // record storage usage.
  IncrCounter(isolate, isolate->GetCurrentContext(), cnt);
}

// VerifyAddressCallback
void VerifyAddressCallback(const FunctionCallbackInfo<Value> &info) {
  Isolate *isolate = info.GetIsolate();
  Local<Object> thisArg = info.Holder();
  Local<External> handler = Local<External>::Cast(thisArg->GetInternalField(0));

  if (info.Length() != 1) {
    isolate->ThrowException(String::NewFromUtf8(
        isolate, "Blockchain.verifyAddress() requires 1 argument"));
    return;
  }

  Local<Value> address = info[0];
  if (!address->IsString()) {
    isolate->ThrowException(
        String::NewFromUtf8(isolate, "address must be string"));
    return;
  }

  size_t cnt = 0;

  int ret = sVerifyAddress(handler->Value(),
                           *String::Utf8Value(address->ToString()), &cnt);
  info.GetReturnValue().Set(ret);

  // record storage usage.
  IncrCounter(isolate, isolate->GetCurrentContext(), cnt);
}

// GetPreBlockHashCallBack
void GetPreBlockHashCallback(const FunctionCallbackInfo<Value> &info) {
  int err = NVM_SUCCESS;
  Isolate *isolate = info.GetIsolate();
  if (NULL == isolate) {
    LogFatalf("Unexpected error: failed to get isolate");
  }
  Local<Object> thisArg = info.Holder();
  Local<External> handler = Local<External>::Cast(thisArg->GetInternalField(0));

  if (info.Length() != 1) {
    isolate->ThrowException(String::NewFromUtf8(
        isolate, "Blockchain.GetPreBlockHash() requires 1 arguments"));
    return;
  } 

  Local<Value> offset = info[0];
  if (!offset->IsNumber()) {
    isolate->ThrowException(
        String::NewFromUtf8(isolate, "Blockchain.GetPreBlockHash(), the argument must be a number")); 
    return;
  }

  double v = Number::Cast(*offset)->Value();
  if (v > ULLONG_MAX || v <= 0) {
    isolate->ThrowException(
        String::NewFromUtf8(isolate, "Blockchain.GetPreBlockHash(), argument out of range"));
    return;
  }

  if (v != (double)(unsigned long long)v) {
        isolate->ThrowException(
        String::NewFromUtf8(isolate, "Blockchain.GetPreBlockHash(), argument must be integer"));
    return;
  }

  size_t cnt = 0;
  char *result = NULL;
  char *exceptionInfo = NULL;
  err = sGetPreBlockHash(handler->Value(), (unsigned long long)(v), &cnt, &result, &exceptionInfo);

  DEAL_ERROR_FROM_GOLANG(err);  

  if (result != NULL) {
    free(result);
    result = NULL;
  }

  if (exceptionInfo != NULL) {
    free(exceptionInfo);
    exceptionInfo = NULL;
  }

  // record storage usage.
  IncrCounter(isolate, isolate->GetCurrentContext(), cnt);
}

// GetPreBlockSeedCallBack
void GetPreBlockSeedCallback(const FunctionCallbackInfo<Value> &info) {
  int err = NVM_SUCCESS;
  Isolate *isolate = info.GetIsolate();
  if (NULL == isolate) {
    LogFatalf("Unexpected error: failed to get isolate");
  }
  Local<Object> thisArg = info.Holder();
  Local<External> handler = Local<External>::Cast(thisArg->GetInternalField(0));

  if (info.Length() != 1) {
    isolate->ThrowException(String::NewFromUtf8(
        isolate, "Blockchain.GetPreBlockSeed() requires 1 arguments"));
    return;
  }
  
  Local<Value> offset = info[0];
  if (!offset->IsNumber()) {
    isolate->ThrowException(
        String::NewFromUtf8(isolate, "Blockchain.GetPreBlockSeed(), the argument must be a number")); 
    return;
  }

  double v = Number::Cast(*offset)->Value();
  if (v > ULLONG_MAX || v <= 0) {
    isolate->ThrowException(
        String::NewFromUtf8(isolate, "Blockchain.GetPreBlockSeed(), argument out of range"));
    return;
  }

 if (v != (double)(unsigned long long)v) {
    isolate->ThrowException(
        String::NewFromUtf8(isolate, "Blockchain.GetPreBlockSeed(), argument must be integer"));
    return;
  }

  size_t cnt = 0;
  char *result = NULL;
  char *exceptionInfo = NULL;
  err = sGetPreBlockSeed(handler->Value(), (unsigned long long)(v), &cnt, &result, &exceptionInfo);

  DEAL_ERROR_FROM_GOLANG(err);

  if (result != NULL) {
    free(result);
    result = NULL;
  }

  if (exceptionInfo != NULL) {
    free(exceptionInfo);
    exceptionInfo = NULL;
  }

  // record storage usage.
  IncrCounter(isolate, isolate->GetCurrentContext(), cnt);
}

void GetContractSourceCallback(const FunctionCallbackInfo<Value> &info) {
  Isolate *isolate = info.GetIsolate();
  Local<Object> thisArg = info.Holder();
  Local<External> handler = Local<External>::Cast(thisArg->GetInternalField(0));

  if (info.Length() != 1) {
    isolate->ThrowException(String::NewFromUtf8(
        isolate, "Blockchain.GetContractSource() requires 1 arguments"));
    return;
  }

  Local<Value> address = info[0];
  if (!address->IsString()) {
    isolate->ThrowException(
        String::NewFromUtf8(isolate, "address must be string"));
    return;
  }

  size_t cnt = 0;

  char *value = sGetContractSource(handler->Value(),
                           *String::Utf8Value(address->ToString()), &cnt);
  if (value == NULL) {
    info.GetReturnValue().SetNull();  //TODO: throw err
  } else {
    info.GetReturnValue().Set(String::NewFromUtf8(isolate, value));
    free(value);
  }

  // record storage usage.
  IncrCounter(isolate, isolate->GetCurrentContext(), cnt);
}

void RunInnerContractSourceCallBack(const FunctionCallbackInfo<Value> &info) {
  Isolate *isolate = info.GetIsolate();
  Local<Object> thisArg = info.Holder();
  Local<External> handler = Local<External>::Cast(thisArg->GetInternalField(0));

  if (info.Length() != 4) {
    char msg[128];
    snprintf(msg, 128, "Blockchain.RunInnerContractSourceCallBack() requires 4 arguments,args:%d", info.Length());
    isolate->ThrowException(String::NewFromUtf8(
        isolate, msg));
    return;
  }

  Local<Value> address = info[0];
  if (!address->IsString()) {
    isolate->ThrowException(
        String::NewFromUtf8(isolate, "address must be string"));
    return;
  }
  Local<Value> funcName = info[1];
  if (!address->IsString()) {
    isolate->ThrowException(
        String::NewFromUtf8(isolate, "func must be string"));
    return;
  }
  Local<Value> val = info[2];
  if (!val->IsString()) {
    isolate->ThrowException(
        String::NewFromUtf8(isolate, "val must be string"));
    return;
  }
  Local<Value> args = info[3];
  if (!args->IsString()) {
    isolate->ThrowException(
        String::NewFromUtf8(isolate, "args must be string"));
    return;
  }

  size_t cnt = 0;
  char *value = sRunInnerContract(handler->Value(),
                           *String::Utf8Value(address->ToString()), *String::Utf8Value(funcName->ToString()),
                           *String::Utf8Value(val->ToString()), *String::Utf8Value(args->ToString()),
                           &cnt);
  // Local<Object> rObj = v8::Object::New(isolate);
  if (value == NULL) {
    Local<Context> context = isolate->GetCurrentContext();
    V8Engine *e = GetV8EngineInstance(context);
    SetInnerContractErrFlag(e);
    TerminateExecution(e);
  } else {
    // Local<Boolean> flag = Boolean::New(isolate, true);
    // rObj->Set(v8::String::NewFromUtf8(isolate, "code"), flag);
    // Local<String> valueStr = v8::String::NewFromUtf8(isolate, value);

    // rObj->Set(v8::String::NewFromUtf8(isolate, "data"), valueStr);
    // info.GetReturnValue().Set(rObj);

    info.GetReturnValue().Set(String::NewFromUtf8(isolate, value));

    free(value);
  }

  // record storage usage.
  IncrCounter(isolate, isolate->GetCurrentContext(), cnt);
}

// Get Latest Nebulas Rank Callback
void GetLatestNebulasRankCallback(const FunctionCallbackInfo<Value> &info) {
  int err = NVM_SUCCESS;

  Isolate *isolate = info.GetIsolate();
  Local<Object> thisArg = info.Holder();
  Local<External> handler = Local<External>::Cast(thisArg->GetInternalField(0));

  if (info.Length() != 1) {
    isolate->ThrowException(String::NewFromUtf8(
        isolate, "Blockchain.getLatestNebulasRank() requires 1 argument"));
    return;
  }

  Local<Value> address = info[0];
  if (!address->IsString()) {
    isolate->ThrowException(
        String::NewFromUtf8(isolate, "address must be string"));
    return;
  }

  size_t cnt = 0;
  char* result = NULL;
  char* exceptionInfo = NULL;
  err = sGetLatestNR(handler->Value(), *String::Utf8Value(address->ToString()), &cnt, &result, &exceptionInfo);

  DEAL_ERROR_FROM_GOLANG(err);

  if (result != NULL) {
    free(result);
    result = NULL;
  }

  if (exceptionInfo != NULL) {
    free(exceptionInfo);
    exceptionInfo = NULL;
  }

  // record storage usage.
  IncrCounter(isolate, isolate->GetCurrentContext(), cnt);
}

// Get Latest Nebulas Rank Summary Callback
void GetLatestNebulasRankSummaryCallback(const FunctionCallbackInfo<Value> &info) {
  int err = NVM_SUCCESS;

  Isolate *isolate = info.GetIsolate();
  Local<Object> thisArg = info.Holder();
  Local<External> handler = Local<External>::Cast(thisArg->GetInternalField(0));

  if (info.Length() != 0) {
    isolate->ThrowException(String::NewFromUtf8(
        isolate, "Blockchain.getLatestNebulasRankSummary() needs not argument"));
    return;
  }

  size_t cnt = 0;
  char* result = NULL;
  char* exceptionInfo = NULL;
  err = sGetLatestNRSum(handler->Value(), &cnt, &result, &exceptionInfo);

  DEAL_ERROR_FROM_GOLANG(err);

  if (result != NULL) {
    free(result);
    result = NULL;
  }

  if (exceptionInfo != NULL) {
    free(exceptionInfo);
    exceptionInfo = NULL;
  }

  // record storage usage.
  IncrCounter(isolate, isolate->GetCurrentContext(), cnt);
}
