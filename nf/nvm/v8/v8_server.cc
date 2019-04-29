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

#include "nvm_engine.h"
#include "compatibility.h"

#define MicroSecondDiff(newtv, oldtv) (1000000 * (unsigned long long)((newtv).tv_sec - (oldtv).tv_sec) + (newtv).tv_usec - (oldtv).tv_usec)  //milliseconds

static int enable_tracer_injection = 0;
static int strict_disallow_usage = 0;
static size_t limits_of_executed_instructions = 0;
static size_t limits_of_total_memory_size = 0;


class NVMEngine;
NVMEngine *gNVMEngine = nullptr;

void logFuncOld(int level, const char *msg) {
  std::thread::id tid = std::this_thread::get_id();
  std::hash<std::thread::id> hasher;

  FILE *f = stdout;
  if (level >= LogLevel::ERROR) {
    f = stderr;
  }
  fprintf(f, "[tid-%020zu] [%s] %s\n", hasher(tid), GetLogLevelText(level), msg);
}

void Logger(int level, const char *msg){
  switch(level){
    case LogLevel::DEBUG:
      LOG(WARNING)<<msg;
      break;
    case LogLevel::WARN:
      LOG(WARNING)<<msg;
      break;
    case LogLevel::INFO:
      LOG(INFO)<<msg;
      break;
    case LogLevel::ERROR:
      LOG(ERROR)<<msg;
      break;
    default:
      break;
  }
}

void Initialization(){
  Initialize();
  InitializeLogger(Logger);
  InitializeRequireDelegate(RequireDelegate, AttachLibVersionDelegate);
  InitializeExecutionEnvDelegate(AttachLibVersionDelegate);

  InitializeStorage(StorageGet, StoragePut, StorageDel);
  InitializeBlockchain(GetTxByHash, GetAccountState, Transfer, VerifyAddress,
          GetPreBlockHash, GetPreBlockSeed, GetContractSource, InnerContract, GetLatestNebulasRank, GetLatestNebulasRankSummary);
  InitializeEvent(EventTrigger);
  InitializeCrypto(Sha256, Sha3256, Ripemd160, RecoverAddress, Md5, Base64);
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
  
  std::cout<<">>>>Creating script thread!!!!"<<std::endl;

	bool btn = CreateScriptThread(&ctx);
  if (btn == false) {
    LogErrorf("Failed to create script thread");
    return nullptr;
  }
  *source_line_offset = ctx.output.line_offset;

  if(FG_DEBUG && ctx.output.result != nullptr)
    std::cout<<"*********** injected traceable source code: "<<std::endl<<ctx.output.result<<std::endl;

  return ctx.output.result;
}

char *TranspileTypeScriptModuleThread(V8Engine *e, const char *source,
                                int *source_line_offset) {
  v8ThreadContext ctx;
  memset(&ctx, 0x00, sizeof(ctx));
  SetRunScriptArgs(&ctx, e, INSTRUCTIONTS, source, *source_line_offset, 1);
	bool btn = CreateScriptThread(&ctx);

  if (btn == false) {
    return nullptr;
  }
  *source_line_offset = ctx.output.line_offset;
  return ctx.output.result;
}

void *ExecuteThread(void *args) {
  v8ThreadContext *ctx = (v8ThreadContext*)args;
  if (ctx->input.opt == INSTRUCTION) {
    TracingContext tContext;
    tContext.source_line_offset = 0;
    tContext.tracable_source = nullptr;
    tContext.strictDisallowUsage = ctx->input.allow_usage;

    Execute(nullptr, ctx->e, ctx->input.source, 0, 0L, 0L, InjectTracingInstructionDelegate,
            (void *)&tContext);

    ctx->output.line_offset = tContext.source_line_offset;
    ctx->output.result = static_cast<char *>(tContext.tracable_source);

  } else if (ctx->input.opt == INSTRUCTIONTS) {
    TypeScriptContext tContext;
    tContext.source_line_offset = 0;
    tContext.js_source = nullptr;

    Execute(nullptr, ctx->e, ctx->input.source, 0, 0L, 0L, TypeScriptTranspileDelegate,
            (void *)&tContext);

    ctx->output.line_offset = tContext.source_line_offset;
    ctx->output.result = static_cast<char *>(tContext.js_source);

  } else {
    ctx->output.ret = Execute(&ctx->output.result, ctx->e, ctx->input.source, ctx->input.line_offset, (void *)ctx->e->lcs,
                (void *)ctx->e->gcs, ExecuteSourceDataDelegate, nullptr);
    LogInfof("iRtn:%d--result:%s\n", ctx->output.ret, ctx->output.result);
    fprintf(stderr, "iRtn:%d--result:%s\n", ctx->output.ret, ctx->output.result);
    
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
  int rtn = gettimeofday(&tcBegin, nullptr);
  if (rtn != 0) {
    LogErrorf("CreateScriptThread get start time err:%d\n", rtn);
    fprintf(stderr, "CreateScriptThread get start time err:%d\n", rtn);

    return false;
  }
  rtn = pthread_create(&thread, &attribute, ExecuteThread, (void *)ctx);
  if (rtn != 0) {
    LogErrorf("CreateScriptThread pthread_create err:%d\n", rtn);
    fprintf(stderr, "CreateScriptThread pthread_create err:%d\n", rtn);
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
      rtn = gettimeofday(&tcEnd, nullptr);
      if (rtn) {
        LogErrorf("CreateScriptThread get end time err:%d\n", rtn);
        continue;
      }
      int diff = MicroSecondDiff(tcEnd, tcBegin);
      if (diff >= timeout && is_kill == false) {
        if(FG_DEBUG)
          std::cout<<"%%%%%%%%%%%% checking termination condition"<<std::endl;
        LogErrorf("CreateScriptThread timeout timeout:%d diff:%d\n", timeout, diff);
        fprintf(stderr, "CreateScriptThread timeout timeout:%d diff:%d\n", timeout, diff);
        TerminateExecution(ctx->e);
        is_kill = true;
      }
    }
  }

  return true;
}


const NVMCallbackResult* DataExchangeCallback(void* handler, NVMCallbackResponse* response){

    std::cout<<"Now is executing callback: "<<response->func_name()<<std::endl;

    if(gNVMEngine != nullptr){
        std::cout<<">>>>nvmengine is not null!"<<std::endl;
        return gNVMEngine->Callback(handler, response);
    }else{
        std::cout<<">>>>>>> Failed to exchange data"<<std::endl;
        LogErrorf("Failed to exchange data");
    }
    return nullptr;
}

void RunServer(const char* addr_str){

  std::string engine_addr(addr_str);

  /*
  if(gNVMEngine != nullptr){
    free(gNVMEngine);
  }
  */

  if(gNVMEngine == nullptr){
    gNVMEngine = new NVMEngine(NVM_CURRENCY_LEVEL);
    std::cout<<"%%%%%%%%%%%% nvm engine is initialized!!!!!!!!!!!"<<std::endl;
  }

  grpc::ServerBuilder builder;
  builder.AddListeningPort(engine_addr, grpc::InsecureServerCredentials());
  builder.RegisterService(gNVMEngine);
  std::unique_ptr<grpc::Server> server(builder.BuildAndStart());
  
  server->Wait();
}

int main(int argc, const char *argv[]) {

  FLAGS_log_dir = "logs";
  ::google::InitGoogleLogging(argv[0]);

  Initialization();

  bool DEBUG = false;

  if(DEBUG){
    NVMEngine* engine = new NVMEngine(NVM_CURRENCY_LEVEL);
    engine->LocalTest();

  }else{
    if(argc > 1){
      RunServer(argv[1]);
    }else{
      std::cout<<"Please specify the port"<<std::endl;
      LogErrorf("Please specify the port");
    }
  }

  return 0;
}
