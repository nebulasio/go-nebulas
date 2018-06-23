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
#include "lib/logger.h"
#include "lib/nvm_error.h"

#include <assert.h>
#include <string.h>
#include <stdio.h>
// #include <stdlib.h>
#include <stdlib.h>
#include <thread>
#include <sys/time.h>
#include <unistd.h>
#define MicroSecondDiff(newtv, oldtv) (1000000 * (unsigned long long)((newtv).tv_sec - (oldtv).tv_sec) + (newtv).tv_usec - (oldtv).tv_usec)  //微秒
#define CodeExecuteErr 1
#define CodeExecuteInnerNvmErr 2
#define CodeTimeOut    3

void SetRunScriptArgs(v8ThreadContext *ctx, V8Engine *e, int opt, const char *source, int line_offset, int allow_usage) {
  ctx->e = e;
  ctx->input.source = source;
  ctx->input.opt = (OptType)opt;
  ctx->input.allow_usage = allow_usage;
  ctx->input.line_offset = line_offset;
}

char *InjectTracingInstructionsThread(V8Engine *e, const char *source,
                                int *source_line_offset,
                                int allow_usage) {
  v8ThreadContext ctx;
  memset(&ctx, 0x00, sizeof(ctx));
  SetRunScriptArgs(&ctx, e, INSTRUCTION, source, *source_line_offset, allow_usage);
	bool btn = CreateScriptThread(&ctx);
  if (btn == false) {
    LogErrorf("Failed to create script thread");
    return NULL;
  }
  *source_line_offset = ctx.output.line_offset;
  return ctx.output.result;
}

char *TranspileTypeScriptModuleThread(V8Engine *e, const char *source,
                                int *source_line_offset) {
  v8ThreadContext ctx;
  memset(&ctx, 0x00, sizeof(ctx));
  SetRunScriptArgs(&ctx, e, INSTRUCTIONTS, source, *source_line_offset, 1);
	bool btn = CreateScriptThread(&ctx);
  if (btn == false) {
    return NULL;
  }
  *source_line_offset = ctx.output.line_offset;
  return ctx.output.result;
}
int RunScriptSourceThread(char **result, V8Engine *e, const char *source,
                    int source_line_offset, uintptr_t lcs_handler,
                    uintptr_t gcs_handler) {
  v8ThreadContext ctx;
  memset(&ctx, 0x00, sizeof(ctx));
  SetRunScriptArgs(&ctx, e, RUNSCRIPT, source, source_line_offset, 1);
	ctx.input.lcs = lcs_handler;
  ctx.input.gcs = gcs_handler;  

  bool btn = CreateScriptThread(&ctx);
  if (btn == false) {
    return NVM_UNEXPECTED_ERR;
  }

  *result = ctx.output.result;
  return ctx.output.ret;
}

void *ExecuteThread(void *args) {
  v8ThreadContext *ctx = (v8ThreadContext*)args;
  if (ctx->input.opt == INSTRUCTION) {
    TracingContext tContext;
    tContext.source_line_offset = 0;
    tContext.tracable_source = NULL;
    tContext.strictDisallowUsage = ctx->input.allow_usage;

    Execute(NULL, ctx->e, ctx->input.source, 0, 0L, 0L, InjectTracingInstructionDelegate,
            (void *)&tContext);

    ctx->output.line_offset = tContext.source_line_offset;
    ctx->output.result = static_cast<char *>(tContext.tracable_source);
  } else if (ctx->input.opt == INSTRUCTIONTS) {
    TypeScriptContext tContext;
    tContext.source_line_offset = 0;
    tContext.js_source = NULL;

    Execute(NULL, ctx->e, ctx->input.source, 0, 0L, 0L, TypeScriptTranspileDelegate,
            (void *)&tContext);

    ctx->output.line_offset = tContext.source_line_offset;
    ctx->output.result = static_cast<char *>(tContext.js_source);
  } else {
    ctx->output.ret = Execute(&ctx->output.result, ctx->e, ctx->input.source, ctx->input.line_offset, (void *)ctx->input.lcs,
                (void *)ctx->input.gcs, ExecuteSourceDataDelegate, NULL);
    // printf("iRtn:%d--result:%s\n", ctx->output.ret, ctx->output.result);
  }

  ctx->is_finished = true;
  return 0x00;
}
// return : success return true. if hava err ,then return false. and not need to free heap
// if gettimeofday hava err ,There is a risk of an infinite loop
bool CreateScriptThread(v8ThreadContext *ctx) {
  pthread_t thread;
  pthread_attr_t attribute;
  pthread_attr_init(&attribute);
  pthread_attr_setstacksize(&attribute, 2 * 1024 * 1024);
  pthread_attr_setdetachstate (&attribute, PTHREAD_CREATE_DETACHED);
  struct timeval tcBegin, tcEnd;
  int rtn = gettimeofday(&tcBegin, NULL);
  if (rtn != 0) {
    LogErrorf("CreateScriptThread get start time err:%d\n", rtn);
    return false;
  }
  rtn = pthread_create(&thread, &attribute, ExecuteThread, (void *)ctx);
  if (rtn != 0) {
    LogErrorf("CreateScriptThread pthread_create err:%d\n", rtn);
    return false;
  }
  
  int timeout = ctx->e->timeout;
  bool is_kill = false;
  //thread safe
  while(1) {
    if (ctx->is_finished == true) {
        if (is_kill == true) {
          ctx->output.ret = NVM_EXE_TIMEOUT_ERR; 
          // ctx->output.ret = CodeTimeOut; 
        } 
        else if (ctx->e->is_inner_nvm_error_happen == 1) {
          ctx->output.ret = NVM_INNER_EXE_ERR;
        }
        break;
    } else {
      usleep(10); //10 micro second loop .epoll_wait optimize
      rtn = gettimeofday(&tcEnd, NULL);
      if (rtn) {
        LogErrorf("CreateScriptThread get end time err:%d\n", rtn);
        continue;
      }
      int diff = MicroSecondDiff(tcEnd, tcBegin);
  
      if (diff >= timeout && is_kill == false) { 
        LogErrorf("CreateScriptThread timeout timeout:%d diff:%d\n", timeout, diff);
        TerminateExecution(ctx->e);
        is_kill = true;
      }
    }
  }
  return true;
}