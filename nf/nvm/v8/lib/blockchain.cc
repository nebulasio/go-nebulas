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
#include "../engine.h"
#include "instruction_counter.h"

static GetTxByHashFunc sGetTxByHash = NULL;
static GetAccountStateFunc sGetAccountState = NULL;
static TransferFunc sTransfer = NULL;
static VerifyAddressFunc sVerifyAddress = NULL;
static GetContractSourceFunc sGetContractSource = NULL;
static RunMultilevelContractSourceFunc sRunMultContract = NULL;

void InitializeBlockchain(GetTxByHashFunc getTx, GetAccountStateFunc getAccount,
                          TransferFunc transfer,
                          VerifyAddressFunc verifyAddress,
                          GetContractSourceFunc contractSource,
                          RunMultilevelContractSourceFunc rMultContract) {
  sGetTxByHash = getTx;
  sGetAccountState = getAccount;
  sTransfer = transfer;
  sVerifyAddress = verifyAddress;
  sGetContractSource = contractSource;
  sRunMultContract = rMultContract;
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

  /* disable getAccountState() function.
    blockTpl->Set(String::NewFromUtf8(isolate, "getAccountState"),
                  FunctionTemplate::New(isolate, GetAccountStateCallback),
                  static_cast<PropertyAttribute>(PropertyAttribute::DontDelete
                  |
                                                 PropertyAttribute::ReadOnly));
  */

  blockTpl->Set(String::NewFromUtf8(isolate, "transfer"),
                FunctionTemplate::New(isolate, TransferCallback),
                static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                               PropertyAttribute::ReadOnly));

  blockTpl->Set(String::NewFromUtf8(isolate, "verifyAddress"),
                FunctionTemplate::New(isolate, VerifyAddressCallback),
                static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                               PropertyAttribute::ReadOnly));
  
  blockTpl->Set(String::NewFromUtf8(isolate, "getContractSource"),
                FunctionTemplate::New(isolate, GetContractSourceCallback),
                static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                               PropertyAttribute::ReadOnly));

  blockTpl->Set(String::NewFromUtf8(isolate, "runContractSource"),
                FunctionTemplate::New(isolate, RunMultilevelContractSourceCallBack),
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
  Isolate *isolate = info.GetIsolate();
  Local<Object> thisArg = info.Holder();
  Local<External> handler = Local<External>::Cast(thisArg->GetInternalField(0));

  if (info.Length() != 1) {
    isolate->ThrowException(String::NewFromUtf8(
        isolate, "Blockchain.getAccountState() requires only 1 argument"));
    return;
  }

  Local<Value> key = info[0];
  if (!key->IsString()) {
    isolate->ThrowException(String::NewFromUtf8(isolate, "key must be string"));
    return;
  }

  size_t cnt = 0;

  char *value = sGetAccountState(handler->Value(),
                                 *String::Utf8Value(key->ToString()), &cnt);
  if (value == NULL) {
    info.GetReturnValue().SetNull();
  } else {
    info.GetReturnValue().Set(String::NewFromUtf8(isolate, value));
    free(value);
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
    info.GetReturnValue().SetNull();
  } else {
    info.GetReturnValue().Set(String::NewFromUtf8(isolate, value));
    free(value);
  }

  // record storage usage.
  IncrCounter(isolate, isolate->GetCurrentContext(), cnt);
}
void RunMultilevelContractSourceCallBack(const FunctionCallbackInfo<Value> &info) {
  Isolate *isolate = info.GetIsolate();
  Local<Object> thisArg = info.Holder();
  Local<External> handler = Local<External>::Cast(thisArg->GetInternalField(0));

  if (info.Length() != 4) {
    char msg[512];
    snprintf(msg, 512, "Blockchain.RunMultilevelContractSourceCallBack() requires 14 arguments,args:%d", info.Length());
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
  size_t rerrType = 0;
  char *rerr = NULL;
  char *value = sRunMultContract(handler->Value(),
                           *String::Utf8Value(address->ToString()), *String::Utf8Value(funcName->ToString()),
                           *String::Utf8Value(val->ToString()), *String::Utf8Value(args->ToString()),
                           &cnt, &rerrType, &rerr);
  Local<Object> rObj = v8::Object::New(isolate);
    // 对象的赋值
    // obj->Set(v8::String::NewFromUtf8(isolate, "arg1"), str);
    // obj->Set(v8::String::NewFromUtf8(isolate, "arg2"), retval);                         
  if (value == NULL) {
    //info.GetReturnValue().SetNull();
    if (rerrType <= 10000) {
      free(rerr);
      isolate->ThrowException(
        String::NewFromUtf8(isolate, "mult run nvm err!!!"));
        return;
    }
    Local<Boolean> flag = Boolean::New(isolate, false);
    rObj->Set(v8::String::NewFromUtf8(isolate, "code"), flag);
    char msg[512];
    snprintf(msg, 512, "blockchain.sRunMultContract err:%s", rerr);
    //rObj->Set(v8::String::NewFromUtf8(isolate, "ret"), Null);
    Local<String> errStr = v8::String::NewFromUtf8(isolate, msg);
    rObj->Set(v8::String::NewFromUtf8(isolate, "msg"), errStr);
    info.GetReturnValue().Set(rObj);
    free(rerr);
    return;
  } else {
    Local<Boolean> flag = Boolean::New(isolate, true);
    rObj->Set(v8::String::NewFromUtf8(isolate, "code"), flag);
    Local<String> valueStr = v8::String::NewFromUtf8(isolate, value);

    rObj->Set(v8::String::NewFromUtf8(isolate, "data"), valueStr);
    info.GetReturnValue().Set(rObj);

    //info.GetReturnValue().Set(String::NewFromUtf8(isolate, value));

    free(value);
  }

  // record storage usage.
  IncrCounter(isolate, isolate->GetCurrentContext(), cnt);
}