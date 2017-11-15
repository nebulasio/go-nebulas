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

static GetBlockByHashFunc GETBLOCKBYHASH = NULL;
static GetTxByHashFunc GETTXBYHASH = NULL;
static GetAccountStateFunc GETACCOUNTSTATE = NULL;
static SendFunc SEND = NULL;

void InitializeBlockchain(GetBlockByHashFunc getBlock, GetTxByHashFunc getTx, GetAccountStateFunc getAccount, SendFunc send) {
  GETBLOCKBYHASH = getBlock;
  GETTXBYHASH = getTx;
  GETACCOUNTSTATE = getAccount;
  SEND = send;
}

void NewBlockchain(Isolate *isolate, Local<ObjectTemplate> globalTpl) {
  Local<FunctionTemplate> type =
      FunctionTemplate::New(isolate, BlockchainConstructor);
  Local<String> className = String::NewFromUtf8(isolate, "NativeBlockchain");
  type->SetClassName(className);

  Local<ObjectTemplate> instanceTpl = type->InstanceTemplate();
  instanceTpl->SetInternalFieldCount(1);

  instanceTpl->Set(String::NewFromUtf8(isolate, "getBlockByHash"),
    FunctionTemplate::New(isolate, GetBlockByHashCallback),
    static_cast<PropertyAttribute>(PropertyAttribute::DontDelete | PropertyAttribute::ReadOnly));

  instanceTpl->Set(String::NewFromUtf8(isolate, "getTransactionByHash"),
    FunctionTemplate::New(isolate, GetTransactionByHashCallback),
    static_cast<PropertyAttribute>(PropertyAttribute::DontDelete | PropertyAttribute::ReadOnly));

  instanceTpl->Set(String::NewFromUtf8(isolate, "getAccountState"),
    FunctionTemplate::New(isolate, GetAccountStateCallback),
    static_cast<PropertyAttribute>(PropertyAttribute::DontDelete | PropertyAttribute::ReadOnly));

  instanceTpl->Set(String::NewFromUtf8(isolate, "send"),
    FunctionTemplate::New(isolate, SendCallback),
    static_cast<PropertyAttribute>(PropertyAttribute::DontDelete | PropertyAttribute::ReadOnly));
}

void BlockchainConstructor(const FunctionCallbackInfo<Value> &info) {
  Isolate *isolate = info.GetIsolate();
  Local<Object> thisArg = info.Holder();

  if (info.Length() != 1) {
    isolate->ThrowException(String::NewFromUtf8(
        isolate, "Blockchain constructor requires only 1 argument"));
    return;
  }

  Local<Value> handler = info[0];
  if (!handler->IsExternal()) {
    isolate->ThrowException(String::NewFromUtf8(
        isolate,
        "Blockchain constructor requires a member of _native_blockchain_handlers"));
    return;
  }

  thisArg->SetInternalField(0, handler);
}

void GetBlockByHashCallback(const FunctionCallbackInfo<Value> &info) {
  Isolate *isolate = info.GetIsolate();
  Local<Object> thisArg = info.Holder();
  Local<External> handler = Local<External>::Cast(thisArg->GetInternalField(0));

  if (info.Length() != 1) {
    isolate->ThrowException(
        String::NewFromUtf8(isolate, "Blockchain.getBlockByHash() requires only 1 argument"));
    return;
  }

  Local<Value> key = info[0];
  if (!key->IsString()) {
    isolate->ThrowException(String::NewFromUtf8(isolate, "key must be string"));
    return;
  }

  char *value = GETBLOCKBYHASH(handler->Value(), *String::Utf8Value(key->ToString()));
  if (value == NULL) {
    info.GetReturnValue().SetNull();
  } else {
    info.GetReturnValue().Set(String::NewFromUtf8(isolate, value));
    free(value);
  }
}

void GetTransactionByHashCallback(const FunctionCallbackInfo<Value> &info) {
  Isolate *isolate = info.GetIsolate();
  Local<Object> thisArg = info.Holder();
  Local<External> handler = Local<External>::Cast(thisArg->GetInternalField(0));

  if (info.Length() != 1) {
    isolate->ThrowException(
        String::NewFromUtf8(isolate, "Blockchain.getTransactionByHash() requires only 1 argument"));
    return;
  }

  Local<Value> key = info[0];
  if (!key->IsString()) {
    isolate->ThrowException(String::NewFromUtf8(isolate, "key must be string"));
    return;
  }

  char *value = GETTXBYHASH(handler->Value(), *String::Utf8Value(key->ToString()));
  if (value == NULL) {
    info.GetReturnValue().SetNull();
  } else {
    info.GetReturnValue().Set(String::NewFromUtf8(isolate, value));
    free(value);
  }
}

void GetAccountStateCallback(const FunctionCallbackInfo<Value> &info) {
  Isolate *isolate = info.GetIsolate();
  Local<Object> thisArg = info.Holder();
  Local<External> handler = Local<External>::Cast(thisArg->GetInternalField(0));

  if (info.Length() != 1) {
    isolate->ThrowException(
        String::NewFromUtf8(isolate, "Blockchain.getAccountState() requires only 1 argument"));
    return;
  }

  Local<Value> key = info[0];
  if (!key->IsString()) {
    isolate->ThrowException(String::NewFromUtf8(isolate, "key must be string"));
    return;
  }

  char *value = GETACCOUNTSTATE(handler->Value(), *String::Utf8Value(key->ToString()));
  if (value == NULL) {
    info.GetReturnValue().SetNull();
  } else {
    info.GetReturnValue().Set(String::NewFromUtf8(isolate, value));
    free(value);
  }
}

void SendCallback(const FunctionCallbackInfo<Value> &info) {
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
    isolate->ThrowException(String::NewFromUtf8(isolate, "address must be string"));
    return;
  }

    Local<Value> amount = info[1];
    if (!amount->IsString()) {
      isolate->ThrowException(String::NewFromUtf8(isolate, "value must be string"));
      return;
    }

  int ret = SEND(handler->Value(), *String::Utf8Value(address->ToString()),  *String::Utf8Value(amount->ToString()));
  info.GetReturnValue().Set(ret);
}
