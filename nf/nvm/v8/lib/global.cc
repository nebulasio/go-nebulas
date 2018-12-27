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

#include "global.h"
#include "blockchain.h"
#include "event.h"
#include "instruction_counter.h"
#include "log_callback.h"
#include "require_callback.h"
#include "storage_object.h"
#include "crypto.h"
#include "random.h"

Local<ObjectTemplate> CreateGlobalObjectTemplate(Isolate *isolate) {
  Local<ObjectTemplate> globalTpl = ObjectTemplate::New(isolate);
  globalTpl->SetInternalFieldCount(1);

  NewNativeRequireFunction(isolate, globalTpl);
  NewNativeLogFunction(isolate, globalTpl);
  NewNativeEventFunction(isolate, globalTpl);
  // NewNativeRandomFunction(isolate, globalTpl);

  NewStorageType(isolate, globalTpl);

  return globalTpl;
}

void SetGlobalObjectProperties(Isolate *isolate, Local<Context> context,
                               V8Engine *e, void *lcsHandler,
                               void *gcsHandler) {
  // set e to global.
  Local<Object> global = context->Global();
  global->SetInternalField(0, External::New(isolate, e));

  NewStorageTypeInstance(isolate, context, lcsHandler, gcsHandler);
  NewInstructionCounterInstance(isolate, context,
                                &(e->stats.count_of_executed_instructions), e);
  uint64_t build_flag = e->ver;
  if (BUILD_MATH == (build_flag & BUILD_MATH)) {
    NewRandomInstance(isolate, context, lcsHandler);
  }
  if (BUILD_BLOCKCHAIN == (build_flag & BUILD_BLOCKCHAIN)) {
    NewBlockchainInstance(isolate, context, lcsHandler, build_flag);
  }
  
  NewCryptoInstance(isolate, context);
}

V8Engine *GetV8EngineInstance(Local<Context> context) {
  Local<Object> global = context->Global();
  Local<Value> val = global->GetInternalField(0);

  if (!val->IsExternal()) {
    return NULL;
  }

  return static_cast<V8Engine *>(Local<External>::Cast(val)->Value());
}
