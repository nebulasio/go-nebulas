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

#include "v8_server.h"

#define MicroSecondDiff(newtv, oldtv) (1000000 * (unsigned long long)((newtv).tv_sec - (oldtv).tv_sec) + (newtv).tv_usec - (oldtv).tv_usec)  //milliseconds

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
  //InitializeCrypto(Sha256Func, Sha3256Func, Ripemd160Func, RecoverAddressFunc, Md5Func, Base64Func);
}

void InitializeDataStruct(){
  srcModuleCache = std::unique_ptr<std::map<std::string, CacheSrcItem>>(new std::map<std::string, CacheSrcItem>());
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

  std::cout<<">>>>>>>>>>>>>Result is: "<<*result<<", source is: "<<source<<", source lineoffset: "<<source_line_offset
    <<", lcs_handler: "<<lcs_handler<<", gcs_handler: "<<gcs_handler<<std::endl;

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
    printf("iRtn:%d--result:%s\n", ctx->output.ret, ctx->output.result);
  }

  ctx->is_finished = true;
  std::cout<<">>>>is_finished has been set to be true"<<std::endl;

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

  std::cout<<"Now is in create script thread"<<std::endl;

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


// NVMEngine related interfaces

bool NVMEngine::GetRunnableSourceCode(const std::string& sourceType, std::string& originalSource){
  const char* jsSource;
  uint64_t originalSourceLineOffset = 0;

  if(sourceType.compare(this->TS_TYPE) == 0){
    jsSource = TranspileTypeScriptModuleThread(this->engine, originalSource.c_str(), &this->m_src_offset);
  }else{
    jsSource = originalSource.c_str();
  }

  char* runnableSource;
  std::string sourceHash = sha256(std::string(jsSource));
  auto searchRecord = srcModuleCache->find(sourceHash);
  if(searchRecord != srcModuleCache->end()){
    CacheSrcItem cachedSourceItem = searchRecord->second;
    this->m_traceable_src = cachedSourceItem.traceableSource;
    this->m_traceale_src_line_offset = cachedSourceItem.traceableSourceLineOffset;
    return true;

  }else{
    char* traceableSource = InjectTracingInstructionsThread(this->engine, jsSource, &this->m_src_offset, this->m_allow_usage);
    this->m_traceable_src = std::string(traceableSource);
    this->m_traceale_src_line_offset = 0;
    CacheSrcItem newItem = {originalSource, originalSourceLineOffset, traceableSource, this->m_traceale_src_line_offset};
    srcModuleCache->insert({sourceHash, newItem});
    return true;
  }

  return false;
}

int NVMEngine::StartScriptExecution(std::string& contractSource, const std::string& scriptType, 
    const std::string& runnableSrc, const std::string& moduleID, const NVMConfigBundle& configBundle){

    // create engine and inject tracing instructions
    if(!this->engine){
      this->engine = CreateEngine();
    }

    this->engine->limits_of_executed_instructions = configBundle.limits_exe_instruction();
    this->engine->limits_of_total_memory_size = configBundle.limits_total_mem_size();

    std::cout<<">>>>Script type is: "<<scriptType<<", source: "<<contractSource<<std::endl;

    if(this->GetRunnableSourceCode(scriptType, contractSource)){           // transpile the source code if necessary, only if the source code is ts
      std::cout<<"Failed to get runnable source code"<<std::endl;
    }


    AddModule(this->engine, moduleID.c_str(), this->m_traceable_src.c_str(), this->m_traceale_src_line_offset);
    /*
    std::string data ("require(\"" + moduleID + "\")");
    std::cout<<"++ Start running source code"<<std::endl;
    RunScriptSourceDelegate(this->engine, data.c_str(), lcsHandler, gcsHandler);
    std::cout<<"++ Finished unnign source code"<<std::endl;

    /*
    // clean up
    DeleteEngine(this->engine);
    this->engine = NULL;
    std::cout<<">>>>After delete engine"<<std::endl;
    NVMRPCResponse* new_response = new NVMRPCResponse();
    new_response->set_result(101);
    new_response->set_msg("Deployed successfully!");
    */

    std::cout<<">>>>Now starting script execution!!!"<<std::endl;

    v8ThreadContext ctx;
    memset(&ctx, 0x00, sizeof(ctx));
    SetRunScriptArgs(&ctx, this->engine, RUNSCRIPT, this->m_runnable_src.c_str(), this->m_traceale_src_line_offset, 1);
    ctx.input.lcs = this->m_lcs_handler;
    ctx.input.gcs = this->m_gcs_handler;
    bool btn = CreateScriptThread(&ctx);
    if (btn == false) {
      return NVM_UNEXPECTED_ERR;
    }

    std::cout<<">>>>>>Get result "<<std::endl;

    if(ctx.output.result != NULL){
      this->m_exe_result = (char*)calloc(strlen(ctx.output.result)+1, sizeof(char));
      strcpy(this->m_exe_result, ctx.output.result);

      std::cout<<">>>>The running result is: "<<this->m_exe_result<<std::endl;

    }else{
      this->m_exe_result = (char*)calloc(1, sizeof(char));
      memset(this->m_exe_result, '\0', 1);
    }

    std::cout<<">>>>Finished running startscriptexecution"<<std::endl;

    return ctx.output.ret;
}

void NVMEngine::ReadExeStats(NVMStatsBundle *statsBundle){
  ReadMemoryStatistics(this->engine);

  statsBundle->set_actual_count_of_execution_instruction((google::protobuf::uint64)this->engine->stats.count_of_executed_instructions);
  statsBundle->set_actual_used_mem_size((google::protobuf::uint64)this->engine->stats.total_memory_size);
}

grpc::Status NVMEngine::SmartContractCall(grpc::ServerContext* context, grpc::ServerReaderWriter<NVMDataResponse, NVMDataRequest>* stream){

  this->m_stm = stream;

  try{
    bool terminate = false;
    NVMDataRequest *request = new NVMDataRequest();

    while(stream->Read(request)){

      std::string requestType = request->request_type();
      google::protobuf::uint32 requestIndx = request->request_indx();

      if(requestType.compare(DATA_REQUEST_START) == 0){

        NVMConfigBundle configBundle = request->config_bundle();
        std::string scriptSrc = configBundle.script_src();
        std::string scriptType = configBundle.script_type();
        std::string runnableSrc = configBundle.runnable_src();
        std::string moduleID = configBundle.module_id();
        google::protobuf::uint64 maxLimitsOfExecutionInstructions = configBundle.max_limits_of_execution_instruction();
        google::protobuf::uint64 defaultTotalMemSize = configBundle.default_limits_of_total_mem_size();
        google::protobuf::uint64 limitsOfExecutionInstructions = configBundle.limits_exe_instruction();
        google::protobuf::uint64 totalMemSize = configBundle.limits_total_mem_size();

        bool enableLimits = configBundle.enable_limits();
        std::string blockJson = configBundle.block_json();
        std::string txJson = configBundle.tx_json();
        google::protobuf::uint64 lcsHandler = configBundle.lcs_handler();
        google::protobuf::uint64 gcsHandler = configBundle.gcs_handler();
        
        std::cout<<">>>Script source is: "<<scriptSrc<<std::endl;
        std::cout<<">>>Script type is: "<<scriptType<<std::endl;
        std::cout<<">>>Runnable src is: "<<runnableSrc<<std::endl;
        std::cout<<">>>Module id is: "<<moduleID<<std::endl;
        std::cout<<">>>blockJson is: "<<blockJson<<std::endl;
        std::cout<<">>>>>tx json is: "<<txJson<<std::endl;
        std::cout<<">>>>lcsHandler is: "<<lcsHandler<<", gcshandler is: "<<gcsHandler<<std::endl;
        std::cout<<">>>>>>>The limit of exe instructions: "<<limitsOfExecutionInstructions<<std::endl;
        std::cout<<">>>>>>>The limit of mem usage: "<<totalMemSize<<std::endl;

        this->m_runnable_src = runnableSrc;
        this->m_module_id = moduleID;
        this->m_lcs_handler = (uintptr_t)lcsHandler;
        this->m_gcs_handler = (uintptr_t)gcsHandler;
        int ret = this->StartScriptExecution(scriptSrc, scriptType, runnableSrc, moduleID, configBundle);
        
        if(this->m_exe_result != nullptr)
          std::cout<<">>>>Hey running is done, and running result is: "<<this->m_exe_result<<std::endl;
        else
          std::cout<<">>>>Hey running is done, and the running result is null!"<<std::endl;

        NVMDataResponse *response = new NVMDataResponse();
        NVMFinalResponse *finalResponse = new NVMFinalResponse();
        finalResponse->set_result(ret);
        finalResponse->set_msg(this->m_exe_result);

        NVMStatsBundle *statsBundle = new NVMStatsBundle();
        ReadExeStats(statsBundle);
        finalResponse->set_allocated_stats_bundle(statsBundle);

        response->set_allocated_final_response(finalResponse);
        response->set_response_type(DATA_RESPONSE_FINAL);
        response->set_response_indx(0);
      
        stream->Write(*response);

        if(this->m_exe_result != nullptr){
          free(this->m_exe_result);
        }
        
      }else if(requestType.compare(DATA_REQUEST_CALL_BACK) == 0){
        // get result from the request index
        std::string metaData = request->meta_data();

        
      }else{
        // throw exception since the request type is not allowed
        std::cout<<"Illegal request type"<<std::endl;
      }
      
    }

  }catch(const std::exception& e){
    std::cout<<e.what()<<std::endl;
  }

  return grpc::Status::OK;
}

void RunServer(const char* addr_str){

  std::string engine_addr(addr_str);

  if(gNVMEngine != NULL)
    free(gNVMEngine);

  gNVMEngine = new NVMEngine(NVM_CURRENCY_LEVEL);

  grpc::ServerBuilder builder;
  builder.AddListeningPort(engine_addr, grpc::InsecureServerCredentials());
  builder.RegisterService(gNVMEngine);
  std::unique_ptr<grpc::Server> server(builder.BuildAndStart());
  //LOG(INFO)<<"V8 engine is listening on: "<<engine_addr;
  
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


// =================== Original Interfaces =====================
void ExecuteScript(const char* filename, V8ExecutionDelegate delegate) {
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