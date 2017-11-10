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

#include "instruction_counter.h"
#include "log_callback.h"

void NewInstructionCounterInstance(Isolate *isolate, Local<Context> context,
                                   size_t *counter) {
  Local<ObjectTemplate> counterTpl = ObjectTemplate::New(isolate);
  counterTpl->SetInternalFieldCount(1);

  counterTpl->Set(String::NewFromUtf8(isolate, "incr"),
                  FunctionTemplate::New(isolate, IncrCounterCallback),
                  static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                                 PropertyAttribute::ReadOnly));
  counterTpl->SetAccessor(
      String::NewFromUtf8(isolate, "count"), CountGetterCallback, 0,
      Local<Value>(), DEFAULT,
      static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                     PropertyAttribute::ReadOnly));

  Local<Object> instance = counterTpl->NewInstance(context).ToLocalChecked();
  instance->SetInternalField(0, External::New(isolate, counter));

  context->Global()->DefineOwnProperty(
      context, String::NewFromUtf8(isolate, "_instruction_counter"), instance,
      static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                     PropertyAttribute::ReadOnly));
}

void IncrCounterCallback(const FunctionCallbackInfo<Value> &info) {
  Isolate *isolate = info.GetIsolate();
  Local<Object> thisArg = info.Holder();
  Local<External> count = Local<External>::Cast(thisArg->GetInternalField(0));

  if (info.Length() < 1) {
    isolate->ThrowException(
        Exception::Error(String::NewFromUtf8(isolate, "incr: mssing params")));
    return;
  }

  Local<Value> arg = info[0];
  if (!arg->IsString()) {
    isolate->ThrowException(Exception::Error(
        String::NewFromUtf8(isolate, "incr: Expression must be string")));
    return;
  }

  String::Utf8Value expr(arg);

  size_t *cnt = static_cast<size_t *>(count->Value());
  *cnt += 1;
  LogInfof("Incr: count = %zu, %s", *cnt, *expr);
}

void CountGetterCallback(Local<String> property,
                         const PropertyCallbackInfo<Value> &info) {
  Isolate *isolate = info.GetIsolate();
  Local<Object> thisArg = info.Holder();
  Local<External> count = Local<External>::Cast(thisArg->GetInternalField(0));

  size_t *cnt = static_cast<size_t *>(count->Value());
  info.GetReturnValue().Set(Number::New(isolate, (double)*cnt));
}
