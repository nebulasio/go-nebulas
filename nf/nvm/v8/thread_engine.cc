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

#include "engine.h"
#include "lib/tracing.h"
#include "lib/typescript.h"

#include <assert.h>
#include <string.h>
#include <stdio.h>
// #include <stdlib.h>
#include <stdlib.h>
#include <thread>

void SetRunScriptArgs(V8Engine *e, int opt, const char *source, int lineOffset, int allowUsage) {
  e->source = source;
  e->opt = (OptType)opt;
  e->allowUsage = allowUsage;
  e->lineOffset = lineOffset;
}

char *InjectTracingInstructionsThread(V8Engine *e, const char *source,
                                int *source_line_offset,
                                int strictDisallowUsage) {
  // TracingContext tContext;
  // tContext.source_line_offset = 0;
  // tContext.tracable_source = NULL;
  // tContext.strictDisallowUsage = strictDisallowUsage;

  // Execute(NULL, e, source, 0, 0L, 0L, InjectTracingInstructionDelegate,
  //         (void *)&tContext);

  // *source_line_offset = tContext.source_line_offset;
  // return static_cast<char *>(tContext.tracable_source);

  SetRunScriptArgs(e, INSTRUCTION, source, *source_line_offset, strictDisallowUsage);
	RunScriptThread(e);
  *source_line_offset = e->lineOffset;
  return e->result;
}

char *TranspileTypeScriptModuleThread(V8Engine *e, const char *source,
                                int *source_line_offset) {
  // TypeScriptContext tContext;
  // tContext.source_line_offset = 0;
  // tContext.js_source = NULL;

  // Execute(NULL, e, source, 0, 0L, 0L, TypeScriptTranspileDelegate,
  //         (void *)&tContext);
  SetRunScriptArgs(e, INSTRUCTIONTS, source, *source_line_offset, 1);
	RunScriptThread(e);
  *source_line_offset = e->lineOffset;
  return e->result;
  // *source_line_offset = tContext.source_line_offset;
  // return static_cast<char *>(tContext.js_source);
}
int RunScriptSourceThread(char **result, V8Engine *e, const char *source,
                    int source_line_offset, uintptr_t lcsHandler,
                    uintptr_t gcsHandler) {
  e->lcs = lcsHandler;
  e->gcs = gcsHandler;                    
  SetRunScriptArgs(e, RUNSCRIPT, source, source_line_offset, 1);
	RunScriptThread(e);
  *result = e->result;
  return e->ret;
}

void DecoratorOutPut(V8Engine *e) {
  if (e->result != NULL) {
    free(e->result);
    e->result = NULL;
  }
    
  e->lineOffset = 0;
  e->ret = 0;
}