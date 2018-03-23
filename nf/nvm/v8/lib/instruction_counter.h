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

#ifndef _NEBULAS_NF_NVM_V8_LIB_INSTRUCTION_COUNTER_H_
#define _NEBULAS_NF_NVM_V8_LIB_INSTRUCTION_COUNTER_H_

#include <v8.h>

using namespace v8;

typedef void (*InstructionCounterIncrListener)(Isolate *isolate, size_t count,
                                               void *context);
void SetInstructionCounterIncrListener(InstructionCounterIncrListener listener);

void NewInstructionCounterInstance(Isolate *isolate, Local<Context> context,
                                   size_t *counter, void *listenerContext);

void IncrCounterCallback(const FunctionCallbackInfo<Value> &info);
void CountGetterCallback(Local<String> property,
                         const PropertyCallbackInfo<Value> &info);

void IncrCounter(Isolate *isolate, Local<Context> context, size_t count);

#endif // _NEBULAS_NF_NVM_V8_LIB_INSTRUCTION_COUNTER_H_
