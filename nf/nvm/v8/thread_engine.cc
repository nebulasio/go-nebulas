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
#include "engine_int.h"
#include "lib/tracing.h"
#include "lib/typescript.h"

#include <assert.h>
#include <string.h>
#include <stdio.h>
// #include <stdlib.h>
#include <stdlib.h>
#include <thread>
#include <sys/time.h>
#include <unistd.h>
#define KillTimeMicros  1000 * 1000 * 2  
#define MicroSecondDiff(newtv, oldtv) (1000000 * (unsigned long long)((newtv).tv_sec - (oldtv).tv_sec) + (newtv).tv_usec - (oldtv).tv_usec)  //微秒
#define CodeTimeOut   2
void SetRunScriptArgs(v8ThreadContext *pc, V8Engine *e, int opt, const char *source, int line_offset, int allow_usage) {
  // e->source = source;
  // e->opt = (OptType)opt;
  // e->allowUsage = allowUsage;
  // e->lineOffset = lineOffset;
 
  pc->te = e;
  pc->input.source = source;
  pc->input.opt = (OptType)opt;
  pc->input.allow_usage = allow_usage;
  pc->input.line_offset = line_offset;
}

char *InjectTracingInstructionsThread(V8Engine *e, const char *source,
                                int *source_line_offset,
                                int allow_usage) {
  // TracingContext tContext;
  // tContext.source_line_offset = 0;
  // tContext.tracable_source = NULL;
  // tContext.strictDisallowUsage = strictDisallowUsage;

  // Execute(NULL, e, source, 0, 0L, 0L, InjectTracingInstructionDelegate,
  //         (void *)&tContext);

  // *source_line_offset = tContext.source_line_offset;
  // return static_cast<char *>(tContext.tracable_source);
  v8ThreadContext *pc = (v8ThreadContext *)calloc(1, sizeof(v8ThreadContext));
  SetRunScriptArgs(pc, e, INSTRUCTION, source, *source_line_offset, allow_usage);
	CreateScriptThread(pc);
  *source_line_offset = pc->input.line_offset;
  char *result = pc->output.result;
  free(pc);
  return result;
}

char *TranspileTypeScriptModuleThread(V8Engine *e, const char *source,
                                int *source_line_offset) {
  // TypeScriptContext tContext;
  // tContext.source_line_offset = 0;
  // tContext.js_source = NULL;

  // Execute(NULL, e, source, 0, 0L, 0L, TypeScriptTranspileDelegate,
  //         (void *)&tContext);
  v8ThreadContext *pc = (v8ThreadContext *)calloc(1, sizeof(v8ThreadContext));
  SetRunScriptArgs(pc, e, INSTRUCTIONTS, source, *source_line_offset, 1);
	CreateScriptThread(pc);
  *source_line_offset = pc->input.line_offset;
  char *result = pc->output.result;
  free(pc);
  return result;
  // *source_line_offset = tContext.source_line_offset;
  // return static_cast<char *>(tContext.js_source);
}
int RunScriptSourceThread(char **result, V8Engine *e, const char *source,
                    int source_line_offset, uintptr_t lcs_handler,
                    uintptr_t gcs_handler) {
 
  v8ThreadContext *pc = (v8ThreadContext *)calloc(1, sizeof(v8ThreadContext));                    
  SetRunScriptArgs(pc, e, RUNSCRIPT, source, source_line_offset, 1);
	pc->input.lcs = lcs_handler;
  pc->input.gcs = gcs_handler;  

  CreateScriptThread(pc);

  // char *result = pc->output.result;
  *result = pc->output.result;
  int ret = pc->output.ret;
  return ret;
}

// void DecoratorOutPut(V8Engine *e) {
//   if (e->result != NULL) {
//     free(e->result);
//     e->result = NULL;
//   }
    
//   e->lineOffset = 0;
//   e->ret = 0;
// }
void *ExecuteThread(void *args) {
  v8ThreadContext *pc = (v8ThreadContext*)args;
  if (pc->input.opt == INSTRUCTION) {
    // printf("begin instruct\n");
    TracingContext tContext;
    tContext.source_line_offset = 0;
    tContext.tracable_source = NULL;
    tContext.strictDisallowUsage = pc->input.allow_usage;

    Execute(NULL, pc->te, pc->input.source, 0, 0L, 0L, InjectTracingInstructionDelegate,
            (void *)&tContext);

    pc->input.line_offset = tContext.source_line_offset;
    pc->output.result = static_cast<char *>(tContext.tracable_source);
  } else if (pc->input.opt == INSTRUCTIONTS) {
    TypeScriptContext tContext;
    tContext.source_line_offset = 0;
    tContext.js_source = NULL;

    Execute(NULL, pc->te, pc->input.source, 0, 0L, 0L, TypeScriptTranspileDelegate,
            (void *)&tContext);

    pc->input.line_offset = tContext.source_line_offset;
    pc->output.result = static_cast<char *>(tContext.js_source);
  } else {
    pc->output.ret = Execute(&pc->output.result, pc->te, pc->input.source, pc->input.line_offset, (void *)pc->input.lcs,
                (void *)pc->input.gcs, ExecuteSourceDataDelegate, NULL);
    printf("iRtn:%d--result:%s\n", pc->output.ret, pc->output.result);
  }

  pc->isRunEnd = true;
  return 0x00;
}

void CreateScriptThread(v8ThreadContext *pc) {
  pthread_t thread;
  pthread_attr_t attribute;
  pthread_attr_init(&attribute);
  pthread_attr_setstacksize(&attribute, 2 * 1024 * 1024);
  pthread_attr_setdetachstate (&attribute, PTHREAD_CREATE_DETACHED);
  // char *file = "test_fe1.js";
  // V8Engine *pe = (V8Engine*)args;
  pthread_create(&thread, &attribute, ExecuteThread, (void *)pc);
  // int count = 0;
  struct timeval tcBegin, tcEnd;
  gettimeofday(&tcBegin, NULL);
  bool isKill = false;
  //thread safe
  while(1) {  //TODO: 可以考虑迁移
    // V8Engine *pe = (V8Engine*)e;
    if (pc->isRunEnd == true) {
      pc->isRunEnd = false;
      // printf("e->stats.count_of_executed_instructions:%lu\n", e->stats.count_of_executed_instructions);
      // if (e->opt == RUN) {
        if (isKill == true) {
          pc->output.ret = CodeTimeOut; 
        }
        break;
      // }
    } else {
      usleep(10); //10 micro second loop .epoll_wait optimize
      gettimeofday(&tcEnd, NULL);
      int diff = MicroSecondDiff(tcEnd, tcBegin);
      if (diff >= KillTimeMicros && isKill == false) { 
        TerminateExecution(pc->te);
        isKill = true;
      }

    }
  }

    // pthread_join(thread, 0);
}