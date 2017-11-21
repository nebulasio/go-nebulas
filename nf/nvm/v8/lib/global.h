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

#ifndef _NEBULAS_NF_NVM_V8_LIB_GLOBAL_H_
#define _NEBULAS_NF_NVM_V8_LIB_GLOBAL_H_

#include "../engine.h"
#include <v8.h>

using namespace v8;

Local<ObjectTemplate> CreateGlobalObjectTemplate(Isolate *isolate);

void SetGlobalObjectProperties(Isolate *isolate, Local<Context> context,
                               V8Engine *e, void *lcsHandler, void *gcsHandler);

V8Engine *GetV8EngineInstance(Local<Context> context);

#endif // _NEBULAS_NF_NVM_V8_LIB_GLOBAL_H_
