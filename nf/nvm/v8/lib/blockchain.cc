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
#include "instruction_counter.h"
#include "logger.h"
#include "limits.h"

static GetTxByHashFunc sGetTxByHash = NULL;
static GetAccountStateFunc sGetAccountState = NULL;
static TransferFunc sTransfer = NULL;
static VerifyAddressFunc sVerifyAddress = NULL;
static GetPreBlockHashFunc sGetPreBlockHash = NULL;
static GetPreBlockSeedFunc sGetPreBlockSeed = NULL;

void InitializeBlockchain(GetTxByHashFunc getTx, GetAccountStateFunc getAccount,
                          TransferFunc transfer,
                          VerifyAddressFunc verifyAddress,
                          GetPreBlockHashFunc getPreBlockHash,
                          GetPreBlockSeedFunc getPreBlockSeed) {
  sGetTxByHash = getTx;
  sGetAccountState = getAccount;
  sTransfer = transfer;
  sVerifyAddress = verifyAddress;
  sGetPreBlockHash = getPreBlockHash;
  sGetPreBlockSeed = getPreBlockSeed;
}

void NewBlockchainInstance(Isolate *isolate, Local<Context> context,
                           void *handler) {
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

  blockTpl->Set(String::NewFromUtf8(isolate, "getPreBlockHash"),
              FunctionTemplate::New(isolate, GetPreBlockHashCallback),
              static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                              PropertyAttribute::ReadOnly));

  blockTpl->Set(String::NewFromUtf8(isolate, "getPreBlockSeed"),
              FunctionTemplate::New(isolate, GetPreBlockSeedCallback),
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
        isolate, "Blockchain.verifyAddress() requires 1 arguments"));
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
