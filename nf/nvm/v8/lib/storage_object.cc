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

#include "storage_object.h"
#include "../engine.h"
#include "instruction_counter.h"
#include "logger.h"
#include <math.h>

static StorageGetFunc GET = NULL;
static StoragePutFunc PUT = NULL;
static StorageDelFunc DEL = NULL;

void NewStorageType(Isolate *isolate, Local<ObjectTemplate> globalTpl) {
  Local<FunctionTemplate> type =
      FunctionTemplate::New(isolate, StorageConstructor);
  Local<String> className = String::NewFromUtf8(isolate, "NativeStorage");
  type->SetClassName(className);

  Local<ObjectTemplate> instanceTpl = type->InstanceTemplate();
  instanceTpl->SetInternalFieldCount(1);
  instanceTpl->Set(
      String::NewFromUtf8(isolate, "get"),
      FunctionTemplate::New(isolate, StorageGetCallback),
      static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                     PropertyAttribute::ReadOnly));
  instanceTpl->Set(
      String::NewFromUtf8(isolate, "set"),
      FunctionTemplate::New(isolate, StoragePutCallback),
      static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                     PropertyAttribute::ReadOnly));
  instanceTpl->Set(
      String::NewFromUtf8(isolate, "put"),
      FunctionTemplate::New(isolate, StoragePutCallback),
      static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                     PropertyAttribute::ReadOnly));
  instanceTpl->Set(
      String::NewFromUtf8(isolate, "del"),
      FunctionTemplate::New(isolate, StorageDelCallback),
      static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                     PropertyAttribute::ReadOnly));

  globalTpl->Set(className, type,
                 static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                                PropertyAttribute::ReadOnly));
}

void NewStorageTypeInstance(Isolate *isolate, Local<Context> context,
                            void *lcsHandler, void *gcsHandler) {
  Local<ObjectTemplate> storageHandlerTpl = ObjectTemplate::New(isolate);
  Local<Object> handlers =
      storageHandlerTpl->NewInstance(context).ToLocalChecked();

  handlers->DefineOwnProperty(
      context, String::NewFromUtf8(isolate, "lcs"),
      External::New(isolate, lcsHandler),
      static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                     PropertyAttribute::ReadOnly));
  handlers->DefineOwnProperty(
      context, String::NewFromUtf8(isolate, "gcs"),
      External::New(isolate, gcsHandler),
      static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                     PropertyAttribute::ReadOnly));

  context->Global()->DefineOwnProperty(
      context, String::NewFromUtf8(isolate, "_native_storage_handlers"),
      handlers,
      static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                     PropertyAttribute::ReadOnly));
}

void InitializeStorage(StorageGetFunc get, StoragePutFunc put,
                       StorageDelFunc del) {
  GET = get;
  PUT = put;
  DEL = del;
}

void StorageConstructor(const FunctionCallbackInfo<Value> &info) {
  Isolate *isolate = info.GetIsolate();
  Local<Object> thisArg = info.Holder();

  if (info.Length() != 1) {
    isolate->ThrowException(String::NewFromUtf8(
        isolate, "Storage constructor requires only 1 argument"));
    return;
  }

  Local<Value> handler = info[0];
  if (!handler->IsExternal()) {
    isolate->ThrowException(String::NewFromUtf8(
        isolate,
        "Storage constructor requires a member of _native_storage_handlers"));
    return;
  }

  thisArg->SetInternalField(0, handler);
}

void StorageGetCallback(const FunctionCallbackInfo<Value> &info) {
  Isolate *isolate = info.GetIsolate();
  Local<Object> thisArg = info.Holder();
  Local<External> handler = Local<External>::Cast(thisArg->GetInternalField(0));

  if (info.Length() != 1) {
    isolate->ThrowException(
        String::NewFromUtf8(isolate, "Storage.get() requires only 1 argument"));
    return;
  }

  Local<Value> key = info[0];
  if (!key->IsString()) {
    isolate->ThrowException(String::NewFromUtf8(isolate, "key must be string"));
    return;
  }

  size_t cnt = 0;
  char *value =
      GET(handler->Value(), *String::Utf8Value(key->ToString()), &cnt);
  if (value == NULL) {
    info.GetReturnValue().SetNull();
  } else {
    info.GetReturnValue().Set(String::NewFromUtf8(isolate, value));
    free(value);
  }

  // record storage usage.
  IncrCounter(isolate, isolate->GetCurrentContext(), cnt);
}

void StoragePutCallback(const FunctionCallbackInfo<Value> &info) {
  Isolate *isolate = info.GetIsolate();
  Local<Object> thisArg = info.Holder();
  Local<External> handler = Local<External>::Cast(thisArg->GetInternalField(0));

  if (info.Length() != 2) {
    isolate->ThrowException(
        String::NewFromUtf8(isolate, "Storage.put() requires only 2 argument"));
    return;
  }

  Local<Value> key = info[0];
  if (!key->IsString()) {
    isolate->ThrowException(String::NewFromUtf8(isolate, "key must be string"));
    return;
  }

  Local<Value> value = info[1];
  if (!value->IsString()) {
    isolate->ThrowException(
        String::NewFromUtf8(isolate, "value must be string"));
    return;
  }

  Local<String> key_str = key->ToString();
  Local<String> val_str = value->ToString();

  size_t cnt = 0;
  int ret = PUT(handler->Value(), *String::Utf8Value(key_str),
                *String::Utf8Value(val_str), &cnt);
  info.GetReturnValue().Set(ret);

  // record storage usage.
  IncrCounter(isolate, isolate->GetCurrentContext(), cnt);
}

void StorageDelCallback(const FunctionCallbackInfo<Value> &info) {
  Isolate *isolate = info.GetIsolate();
  Local<Object> thisArg = info.Holder();
  Local<External> handler = Local<External>::Cast(thisArg->GetInternalField(0));

  if (info.Length() != 1) {
    isolate->ThrowException(
        String::NewFromUtf8(isolate, "Storage.del() requires only 1 argument"));
    return;
  }

  Local<Value> key = info[0];
  if (!key->IsString()) {
    isolate->ThrowException(String::NewFromUtf8(isolate, "key must be string"));
    return;
  }

  size_t cnt = 0;
  int ret = DEL(handler->Value(), *String::Utf8Value(key->ToString()), &cnt);
  info.GetReturnValue().Set(ret);

  // record storage usage.
  IncrCounter(isolate, isolate->GetCurrentContext(), cnt);
}
