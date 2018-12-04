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
#include <unistd.h>
#include "engine.h"
#include "lib/blockchain.h"
#include "lib/fake_blockchain.h"
#include "lib/file.h"
#include "lib/log_callback.h"
#include "lib/logger.h"

#include "pb/nvm.grpc.pb.h"
#include "pb/nvm.pb.h"
#include "samples/memory_modules.h"
#include "samples/memory_storage.h"

#include <thread>
#include <vector>
#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <iostream>

#include <grpc/grpc.h>
#include <grpcpp/server.h>
#include <grpcpp/server_builder.h>
#include <grpcpp/server_context.h>

#include <glog/logging.h>

#include "engine.h"
#include "engine_int.h"
#include "lib/tracing.h"
#include "lib/typescript.h"
#include "lib/logger.h"
#include "lib/nvm_error.h"

#include <assert.h>
#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include <thread>
#include <sys/time.h>
#include <unistd.h>


#define MicroSecondDiff(newtv, oldtv) (1000000 * (unsigned long long)((newtv).tv_sec - (oldtv).tv_sec) + (newtv).tv_usec - (oldtv).tv_usec)  //milliseconds

static int concurrency = 1;
static int enable_tracer_injection = 0;
static int strict_disallow_usage = 0;
static size_t limits_of_executed_instructions = 0;
static size_t limits_of_total_memory_size = 0;
static int print_injection_result = 0;

void logFunc(int level, const char *msg) {
  std::thread::id tid = std::this_thread::get_id();
  std::hash<std::thread::id> hasher;

  FILE *f = stdout;
  if (level >= LogLevel::ERROR) {
    f = stderr;
  }
  fprintf(f, "[tid-%020zu] [%s] %s\n", hasher(tid), GetLogLevelText(level),
          msg);
}

void eventTriggerFunc(void *handler, const char *topic, const char *data,
                      size_t *cnt) {
  fprintf(stdout, "[Event] [%s] %s\n", topic, data);
  *cnt = 20 + strlen(topic) + strlen(data);
}


typedef void (*V8ExecutionDelegate)(V8Engine *e, const char *data,
                                    uintptr_t lcsHandler, uintptr_t gcsHandler);

void RunScriptSourceDelegate(V8Engine *e, const char *data,
                             uintptr_t lcsHandler, uintptr_t gcsHandler) {
  int lineOffset = 0;

  if (enable_tracer_injection) {
    e->limits_of_executed_instructions = limits_of_executed_instructions;
    e->limits_of_total_memory_size = limits_of_total_memory_size;

    char *traceableSource =
        InjectTracingInstructions(e, data, &lineOffset, strict_disallow_usage);
    if (traceableSource == NULL) {
      fprintf(stderr, "Inject tracing instructions failed.\n");
    } else {
      char *out = NULL;
      int ret = RunScriptSource(&out, e, traceableSource, lineOffset,
                                (uintptr_t)lcsHandler, (uintptr_t)gcsHandler);
      free(traceableSource);

      fprintf(stdout, "[V8] Execution ret = %d, out = %s\n", ret, out);
      free(out);

      ret = IsEngineLimitsExceeded(e);
      if (ret) {
        fprintf(stdout, "[V8Error] Exceed %s limits, ret = %d\n",
                ret == 1 ? "Instructions" : "Memory", ret);
      }

      // print tracing stats.
      fprintf(stdout,
              "\nStats of V8Engine:\n"
              "  count_of_executed_instructions: \t%zu\n"
              "  total_memory_size: \t\t\t%zu\n"
              "  total_heap_size: \t\t\t%zu\n"
              "  total_heap_size_executable: \t\t%zu\n"
              "  total_physical_size: \t\t\t%zu\n"
              "  total_available_size: \t\t%zu\n"
              "  used_heap_size: \t\t\t%zu\n"
              "  heap_size_limit: \t\t\t%zu\n"
              "  malloced_memory: \t\t\t%zu\n"
              "  peak_malloced_memory: \t\t%zu\n"
              "  total_array_buffer_size: \t\t%zu\n"
              "  peak_array_buffer_size: \t\t%zu\n",
              e->stats.count_of_executed_instructions,
              e->stats.total_memory_size, e->stats.total_heap_size,
              e->stats.total_heap_size_executable, e->stats.total_physical_size,
              e->stats.total_available_size, e->stats.used_heap_size,
              e->stats.heap_size_limit, e->stats.malloced_memory,
              e->stats.peak_malloced_memory, e->stats.total_array_buffer_size,
              e->stats.peak_array_buffer_size);
    }
  } else {
    char *out = NULL;
    int ret = RunScriptSource(&out, e, data, lineOffset, (uintptr_t)lcsHandler,
                              (uintptr_t)gcsHandler);
    fprintf(stdout, "[V8] Execution ret = %d, out = %s\n", ret, out);
    free(out);
  }
}

void InjectTracingInstructionsAndPrintDelegate(V8Engine *e, const char *data,
                                               uintptr_t lcsHandler,
                                               uintptr_t gcsHandler) {
  const char *begin = strstr(data, "\"") + 1;
  const char *end = strstr(begin, "\"");

  char id[128];
  int idx = 0;
  while (begin + idx != end) {
    id[idx] = begin[idx];
    idx++;
  }
  id[idx] = '\0';

  char *source = GetModuleSource(e, id);
  fprintf(stdout, "%s", source);
  free(source);
}

void Initialization(){
  Initialize();
  InitializeLogger(logFunc);
  InitializeRequireDelegate(RequireDelegateFunc, AttachLibVersionDelegateFunc);
  InitializeExecutionEnvDelegate(AttachLibVersionDelegateFunc);
  InitializeStorage(StorageGet, StoragePut, StorageDel);
  InitializeBlockchain(GetTxByHash, GetAccountState, Transfer, VerifyAddress, GetPreBlockHash, GetPreBlockSeed);
  InitializeEvent(eventTriggerFunc);
}


//===================== v8 engine thread interfaces =======================

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





class NVMEngine final: public NVMService::Service{

  public:

    explicit NVMEngine(const int concurrency){
      //TODO: specify how many threads we should start and do Initialization

      m_concurrency_scale = concurrency;
      m_src_offset = 0;
    }

    void ConfigureEngine(){
      
    }

    char* GetRunnableSourceCode(const std::string& sourceType, std::string& originalSource){
      const char* jsSource;

      if(sourceType.compare(this->TS_TYPE) == 0){
        jsSource = TranspileTypeScriptModuleThread(this->engine, originalSource.c_str(), &this->m_src_offset);
      }else{
        jsSource = originalSource.c_str();
      }

      // prepare runnable contract source

      return nullptr;
    }

    grpc::Status DeploySmartContract(grpc::ServerContext* ctx, const NVMCallRequest* request, NVMRPCResponse* response) override {

      std::string scriptSrc = request->script_src();
      std::string scriptType = request->script_type();
      std::string functionName = request->func_name();

      LOG(INFO)<<"Request script source is: "<<scriptSrc;
      LOG(INFO)<<"Request script source type is: "<<scriptType;
      LOG(INFO)<<"Request function name is: "<<functionName;
      
      LOG(INFO)<<"Request address is: "<<request->from_addr();
      LOG(INFO)<<"Request block height is: "<<request->block_height();

      // create engine and inject tracing instructions
      if(!this->engine){
        this->engine = CreateEngine();
      }
      std::string contractSource = request->script_src();
      char* runnableSourceCode = this->GetRunnableSourceCode(scriptType, scriptSrc);
      InjectTracingInstructionsThread(this->engine, runnableSourceCode, &this->m_src_offset, this->m_allow_usage);



      // clean up
      DeleteEngine(this->engine);

      NVMRPCResponse* new_response = new NVMRPCResponse();
      new_response->set_result(101);
      new_response->set_msg("Deployed successfully!");

      return grpc::Status::OK;
    }

    grpc::Status NVMDataExchange(grpc::ServerContext* ctx, grpc::ServerReaderWriter<NVMRPCResponse, NVMDataRequest> *stm) override {

      //read the request firstly
      NVMDataRequest *request = new NVMDataRequest();
      stm->Read(request);
      LOG(INFO)<<"The request script souce "<<request->script_src()<<std::endl;
      LOG(INFO)<<"The function name is: "<<request->function_name()<<std::endl;

      //send request to client if necessary
      NVMRPCResponse* response = new NVMRPCResponse();
      response->set_result(202);
      response->set_msg("Send request to client!");

      stm->Write(*response);

      return grpc::Status::OK;
    }    

  private:
    int m_concurrency_scale = 1;              // default concurrency number
    int m_src_offset = 0;                     // default source code offset
    int m_allow_usage = 1;                    // default allow usage
    V8Engine* engine = nullptr;              // default engine
    
    // constants for defining contract source type
    const std::string TS_TYPE = "ts";
    const std::string JS_TYPE = "js"; 
};


void RunServer(const char* addr_str){

    std::string engine_addr(addr_str);
    NVMEngine* engine = new NVMEngine(1);

    grpc::ServerBuilder builder;
    builder.AddListeningPort(engine_addr, grpc::InsecureServerCredentials());
    builder.RegisterService(engine);
    std::unique_ptr<grpc::Server> server(builder.BuildAndStart());
    LOG(INFO)<<"V8 engine is listening on: "<<engine_addr;
    
    server->Wait();
}

int main(int argc, const char *argv[]) {

  FLAGS_log_dir = "logs";
  ::google::InitGoogleLogging(argv[0]);

  Initialization();

  if(argc > 1){
    RunServer(argv[1]);
  }else{
    std::cout<<"Please specify the port"<<std::endl;
  }

  return 0;
}


// =================== Sample Program =====================
void ExecuteScript(const char *filename, V8ExecutionDelegate delegate) {
  void *lcsHandler = CreateStorageHandler();
  void *gcsHandler = CreateStorageHandler();

  V8Engine *e = CreateEngine();

  size_t size = 0;
  int lineOffset = 0;
  char *source = readFile(filename, &size);
  if (source == NULL) {
    LOG(FATAL)<<filename<<" is not found."<<std::endl;
    exit(1);
  }

  // convert TS to js if needed.
  size_t filenameLen = strlen(filename);
  if (filenameLen > 3 && filename[filenameLen - 3] == '.' &&
      filename[filenameLen - 2] == 't' && filename[filenameLen - 1] == 's') {
    size = 0;
    char *jsSource = TranspileTypeScriptModule(e, source, &lineOffset);
    if (jsSource == NULL) {
      fprintf(stderr, "%s is not a valid TypeScript file.\n", filename);
      free(source);
      exit(1);
    }
    free(source);
    source = jsSource;
  }

  // inject tracing code.
  if (enable_tracer_injection) {
    char *traceableSource = InjectTracingInstructions(e, source, &lineOffset,
                                                      strict_disallow_usage);
    if (traceableSource == NULL) {
      fprintf(stderr, "Inject tracing instructions failed.\n");
      free(source);
      return;
    }
    free(source);
    source = traceableSource;
  }

  char id[128];
  sprintf(id, "./%s", filename);

  AddModule(e, id, source, lineOffset);

  char data[128];
  sprintf(data, "require(\"%s\");", id);

  delegate(e, data, (uintptr_t)lcsHandler, (uintptr_t)gcsHandler);

  free(source);
  DeleteEngine(e);

  DeleteStorageHandler(lcsHandler);
  DeleteStorageHandler(gcsHandler);
}

void *loop(void *arg) {
  ExecuteScript((const char*)arg, RunScriptSourceDelegate);
  return 0x00;
}

void ExecuteScriptSource(const char *filename) {
  ExecuteScript(filename, RunScriptSourceDelegate);
}