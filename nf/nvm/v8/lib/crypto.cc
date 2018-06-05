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

#include "crypto.h"
#include "../engine.h"
#include "instruction_counter.h"

static Sha256Func sSha256 = NULL;
static Sha3256Func sSha3256 = NULL;
static Ripemd160Func sRipemd160 = NULL;
static RecoverAddressFunc sRecoverAddress = NULL;

void InitializeCrypto(Sha256Func sha256,
                                 Sha3256Func sha3256,
                                 Ripemd160Func ripemd160,
                                 RecoverAddressFunc recoverAddress) {
    sSha256 = sha256;
    sSha3256 = sha3256;
    sRipemd160 = ripemd160;
    sRecoverAddress = recoverAddress;
}

void NewCryptoInstance(Isolate *isolate, Local<Context> context) {
  Local<ObjectTemplate> cryptoTpl = ObjectTemplate::New(isolate);
//   cryptoTpl->SetInternalFieldCount(1);

  cryptoTpl->Set(String::NewFromUtf8(isolate, "sha256"),
                FunctionTemplate::New(isolate, Sha256Callback),
                static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                               PropertyAttribute::ReadOnly));

  cryptoTpl->Set(String::NewFromUtf8(isolate, "sha3256"),
                FunctionTemplate::New(isolate, Sha3256Callback),
                static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                               PropertyAttribute::ReadOnly));

  cryptoTpl->Set(String::NewFromUtf8(isolate, "ripemd160"),
                FunctionTemplate::New(isolate, Ripemd160Callback),
                static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                               PropertyAttribute::ReadOnly));

  cryptoTpl->Set(String::NewFromUtf8(isolate, "recoverAddress"),
                FunctionTemplate::New(isolate, RecoverAddressCallback),
                static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                               PropertyAttribute::ReadOnly));

  Local<Object> instance = cryptoTpl->NewInstance(context).ToLocalChecked();
//   instance->SetInternalField(0, External::New(isolate, handler));

  context->Global()->DefineOwnProperty(
      context, String::NewFromUtf8(isolate, "_native_crypto"), instance,
      static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                     PropertyAttribute::ReadOnly));
}

// Sha256Callback
void Sha256Callback(const FunctionCallbackInfo<Value> &info) {
  Isolate *isolate = info.GetIsolate();

  if (info.Length() != 1) {
    isolate->ThrowException(String::NewFromUtf8(
        isolate, "_native_crypto.sha256() requires only 1 argument"));
    return;
  }

  Local<Value> data = info[0];
  if (!data->IsString()) {
    isolate->ThrowException(String::NewFromUtf8(isolate, "_native_crypto.sha256(): argument must be string"));
    return;
  }

  size_t cnt = 0;

  char *value = sSha256(*String::Utf8Value(data->ToString()), &cnt);
  if (value == NULL) {
    info.GetReturnValue().SetNull();
  } else {
    info.GetReturnValue().Set(String::NewFromUtf8(isolate, value));
    free(value);
  }

  // record storage usage.
  IncrCounter(isolate, isolate->GetCurrentContext(), cnt);
}

// Sha3256Callback
void Sha3256Callback(const FunctionCallbackInfo<Value> &info) {
  Isolate *isolate = info.GetIsolate();

  if (info.Length() != 1) {
    isolate->ThrowException(String::NewFromUtf8(
        isolate, "_native_crypto.sha3256() requires only 1 argument"));
    return;
  }

  Local<Value> data = info[0];
  if (!data->IsString()) {
    isolate->ThrowException(String::NewFromUtf8(isolate, "_native_crypto.sha3256(): argument must be string"));
    return;
  }

  size_t cnt = 0;

  char *value = sSha3256(*String::Utf8Value(data->ToString()), &cnt);
  if (value == NULL) {
    info.GetReturnValue().SetNull();
  } else {
    info.GetReturnValue().Set(String::NewFromUtf8(isolate, value));
    free(value);
  }

  // record storage usage.
  IncrCounter(isolate, isolate->GetCurrentContext(), cnt);
}

// Ripemd160Callback
void Ripemd160Callback(const FunctionCallbackInfo<Value> &info) {
  Isolate *isolate = info.GetIsolate();

  if (info.Length() != 1) {
    isolate->ThrowException(String::NewFromUtf8(
        isolate, "_native_crypto.ripemd() requires only 1 argument"));
    return;
  }

  Local<Value> data = info[0];
  if (!data->IsString()) {
    isolate->ThrowException(String::NewFromUtf8(isolate, "_native_crypto.ripemd(): argument must be string"));
    return;
  }

  size_t cnt = 0;

  char *value = sRipemd160(*String::Utf8Value(data->ToString()), &cnt);
  if (value == NULL) {
    info.GetReturnValue().SetNull();
  } else {
    info.GetReturnValue().Set(String::NewFromUtf8(isolate, value));
    free(value);
  }

  // record storage usage.
  IncrCounter(isolate, isolate->GetCurrentContext(), cnt);
}

// RecoverAddressCallback
void RecoverAddressCallback(const FunctionCallbackInfo<Value> &info) {
  Isolate *isolate = info.GetIsolate();

  if (info.Length() != 3) {
    isolate->ThrowException(String::NewFromUtf8(
        isolate, "_native_crypto.recoverAddress() requires 3 arguments"));
    return;
  }

  Local<Value> alg = info[0];
  if (!alg->IsInt32()) {
    isolate->ThrowException(
        String::NewFromUtf8(isolate, "_native_crypto.recoverAddress(): alg must be integer"));
    return;
  }

  Local<Value> data = info[1];
  if (!data->IsString()) {
    isolate->ThrowException(
        String::NewFromUtf8(isolate, "_native_crypto.recoverAddress(): data must be string"));
    return;
  }

  Local<Value> sign = info[2];
  if (!sign->IsString()) {
    isolate->ThrowException(
        String::NewFromUtf8(isolate, "_native_crypto.recoverAddress(): sign must be string"));
    return;
  }

  size_t cnt = 0;

  char *value = sRecoverAddress(alg->ToInt32()->Int32Value(), *String::Utf8Value(data->ToString()), 
                               *String::Utf8Value(sign->ToString()), &cnt);
  if (value == NULL) {
    info.GetReturnValue().SetNull();
  } else {
    info.GetReturnValue().Set(String::NewFromUtf8(isolate, value));
    free(value);
  }

  // record storage usage.
  IncrCounter(isolate, isolate->GetCurrentContext(), cnt);
}
