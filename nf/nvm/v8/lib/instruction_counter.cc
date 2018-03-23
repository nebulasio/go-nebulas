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
#include "logger.h"

static char sInstructionCounter[] = "_instruction_counter";

static InstructionCounterIncrListener sListener = NULL;

void NewInstructionCounterInstance(Isolate *isolate, Local<Context> context,
                                   size_t *counter, void *listenerContext) {
  Local<ObjectTemplate> counterTpl = ObjectTemplate::New(isolate);
  counterTpl->SetInternalFieldCount(2);

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
  instance->SetInternalField(1, External::New(isolate, listenerContext));

  context->Global()->DefineOwnProperty(
      context, String::NewFromUtf8(isolate, sInstructionCounter), instance,
      static_cast<PropertyAttribute>(PropertyAttribute::DontDelete |
                                     PropertyAttribute::ReadOnly));
}

void IncrCounterCallback(const FunctionCallbackInfo<Value> &info) {
  Isolate *isolate = info.GetIsolate();
  Local<Object> thisArg = info.Holder();
  Local<External> count = Local<External>::Cast(thisArg->GetInternalField(0));
  Local<External> listenerContext =
      Local<External>::Cast(thisArg->GetInternalField(1));

  if (info.Length() < 1) {
    isolate->ThrowException(
        Exception::Error(String::NewFromUtf8(isolate, "incr: mssing params")));
    return;
  }

  Local<Value> arg = info[0];
  if (!arg->IsNumber()) {
    isolate->ThrowException(Exception::Error(
        String::NewFromUtf8(isolate, "incr: value must be number")));
    return;
  }

  // always return true.
  info.GetReturnValue().Set(true);

  int32_t val = arg->Int32Value();
  if (val < 0) {
    return;
  }

  size_t *cnt = static_cast<size_t *>(count->Value());
  *cnt += val;

  if (sListener != NULL) {
    sListener(isolate, *cnt, listenerContext->Value());
  }
}

void CountGetterCallback(Local<String> property,
                         const PropertyCallbackInfo<Value> &info) {
  Isolate *isolate = info.GetIsolate();
  Local<Object> thisArg = info.Holder();
  Local<External> count = Local<External>::Cast(thisArg->GetInternalField(0));

  size_t *cnt = static_cast<size_t *>(count->Value());
  info.GetReturnValue().Set(Number::New(isolate, (double)*cnt));
}

void IncrCounter(Isolate *isolate, Local<Context> context, size_t val) {
  if (val == 0) {
    return;
  }

  Local<Object> global = context->Global();
  HandleScope handle_scope(isolate);

  Local<Object> counter = Local<Object>::Cast(
      global->Get(String::NewFromUtf8(isolate, sInstructionCounter)));

  Local<External> count = Local<External>::Cast(counter->GetInternalField(0));
  Local<External> listenerContext =
      Local<External>::Cast(counter->GetInternalField(1));

  size_t *cnt = static_cast<size_t *>(count->Value());
  *cnt += val;

  if (sListener != NULL) {
    sListener(isolate, *cnt, listenerContext->Value());
  }
}

void SetInstructionCounterIncrListener(
    InstructionCounterIncrListener listener) {
  sListener = listener;
}
