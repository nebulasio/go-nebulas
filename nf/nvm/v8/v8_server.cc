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

#include "v8_util.h"
#include "compatibility.h"
#include <unistd.h>

#define MicroSecondDiff(newtv, oldtv) (1000000 * (unsigned long long)((newtv).tv_sec - (oldtv).tv_sec) + (newtv).tv_usec - (oldtv).tv_usec)  //milliseconds

static int enable_tracer_injection = 0;
static int strict_disallow_usage = 0;
static size_t limits_of_executed_instructions = 0;
static size_t limits_of_total_memory_size = 0;
//static int print_injection_result = 0;

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

void RunScriptSourceDelegate(V8Engine *e, const char *data,
                             uintptr_t lcsHandler, uintptr_t gcsHandler) {
  int lineOffset = 0;

  if (enable_tracer_injection) {
    e->limits_of_executed_instructions = limits_of_executed_instructions;
    e->limits_of_total_memory_size = limits_of_total_memory_size;

    char *traceableSource =
        InjectTracingInstructions(e, data, &lineOffset, strict_disallow_usage);

    if(FG_DEBUG)
      std::cout<<"The tracebale source code is: "<<traceableSource<<std::endl;

    if (traceableSource == nullptr) {
      LogErrorf("Inject tracing instructions failed.\n");
      fprintf(stderr, "Inject tracing instructions failed.\n");
    } else {
      char *out = nullptr;
      int ret = RunScriptSource(&out, e, traceableSource, lineOffset,
                                (uintptr_t)lcsHandler, (uintptr_t)gcsHandler);
      if(traceableSource != nullptr)
        free(traceableSource);

      LogInfof("[V8] Execution ret = %d, out = %s\n", ret, out);
      fprintf(stderr, "[V8] Execution ret = %d, out = %s\n", ret, out);

      if(out != nullptr)
        free(out);

      ret = IsEngineLimitsExceeded(e);
      if (ret) {
        LogErrorf("[V8Error] Exceed %s limits, ret = %d\n",
                ret == 1 ? "Instructions" : "Memory", ret);

        fprintf(stderr, "[V8Error] Exceed %s limits, ret = %d\n",
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

      LogInfof("\nStats of V8Engine:\n"
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
    char *out = nullptr;
    int ret = RunScriptSource(&out, e, data, lineOffset, (uintptr_t)lcsHandler,
                              (uintptr_t)gcsHandler);
    LogInfof("[V8] Execution ret = %d, out = %s\n", ret, out);
    fprintf(stderr, "[V8] Execution ret = %d, out = %s\n", ret, out);

    if(out != nullptr)
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
  if(source != nullptr)
    free(source);
}

void Initialization(){
  Initialize();
  InitializeLogger(Logger);
  InitializeRequireDelegate(RequireDelegate, AttachLibVersionDelegate);
  InitializeExecutionEnvDelegate(AttachLibVersionDelegate);

  InitializeStorage(StorageGet, StoragePut, StorageDel);
  InitializeBlockchain(GetTxByHash, GetAccountState, Transfer, VerifyAddress, GetPreBlockHash, GetPreBlockSeed);
  InitializeEvent(EventTrigger);
  InitializeCrypto(Sha256, Sha3256, Ripemd160, RecoverAddress, Md5, Base64);
}

void InitializeDataStructure(){
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
  if(FG_DEBUG)
    std::cout<<">>>>>> Finishing injecting tracing instructions thread"<<std::endl;
  if (btn == false) {
    LogErrorf("Failed to create script thread");
    return nullptr;
  }
  *source_line_offset = ctx.output.line_offset;

  if(FG_DEBUG)
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
    ctx->output.ret = Execute(&ctx->output.result, ctx->e, ctx->input.source, ctx->input.line_offset, (void *)ctx->input.lcs,
                (void *)ctx->input.gcs, ExecuteSourceDataDelegate, nullptr);
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

  if(FG_DEBUG)
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


// NVMEngine related interfaces
int NVMEngine::GetRunnableSourceCode(const std::string& sourceType, std::string& originalSource){
  const char* jsSource;
  uint64_t originalSourceLineOffset = 0;

  if(sourceType.compare(this->TS_TYPE) == 0){
    jsSource = TranspileTypeScriptModuleThread(this->engine, originalSource.c_str(), &this->m_src_offset);
    if(jsSource == nullptr){
      return NVM_TRANSPILE_SCRIPT_ERR;
    }
  }else{
    jsSource = originalSource.c_str();
  }

  std::string sourceHash = sha256(std::string(jsSource));
  auto searchRecord = srcModuleCache->find(sourceHash);
  if(searchRecord != srcModuleCache->end()){
    if(FG_DEBUG)
      std::cout<<">>>>>> Found existing runnable source module"<<std::endl;
    CacheSrcItem cachedSourceItem = searchRecord->second;
    this->m_traceable_src = cachedSourceItem.traceableSource;
    this->m_traceale_src_line_offset = cachedSourceItem.traceableSourceLineOffset;
    return 0;

  }else{
    if(FG_DEBUG)
      std::cout<<">>>>>>>>>> Injecting tracking instructions in thread"<<std::endl;
    char* traceableSource = InjectTracingInstructionsThread(this->engine, jsSource, &this->m_src_offset, this->m_allow_usage);

    if(traceableSource != nullptr){
      this->m_traceable_src = std::string(traceableSource);
      this->m_traceale_src_line_offset = 0;
      CacheSrcItem newItem = {originalSource, originalSourceLineOffset, traceableSource, this->m_traceale_src_line_offset};
      srcModuleCache->insert({sourceHash, newItem});
      return 0;
    }

    if(FG_DEBUG && traceableSource!=nullptr){
      std::cout<<">>>>Traceable source code is: "<<traceableSource<<std::endl;
    }
  }
 
  return NVM_INJECT_TRACING_INSTRUCTION_ERR;
}

int NVMEngine::StartScriptExecution(std::string& contractSource, const std::string& scriptType, 
    const std::string& runnableSrc, const std::string& moduleID, const NVMConfigBundle& configBundle){

    // create engine and inject tracing instructions
    if(this->engine == nullptr){
      this->engine = CreateEngine();
    }

    this->engine->limits_of_executed_instructions = configBundle.limits_exe_instruction();
    this->engine->limits_of_total_memory_size = configBundle.limits_total_mem_size();

    // transpile script and inject tracing code if necessary
    int runnableSourceResult = this->GetRunnableSourceCode(scriptType, contractSource);
    if(runnableSourceResult != 0){
      LogErrorf("Failed to get runnable source code");
      return runnableSourceResult;
    }

    AddModule(this->engine, moduleID.c_str(), this->m_traceable_src.c_str(), this->m_traceale_src_line_offset);

    // check limitations
    if(configBundle.block_height() >= GetNVMMemoryLimitWithoutInjectHeight()) {
      ReadMemoryStatistics(this->engine);
      uint64_t actualTotalMemorySize = (uint64_t)this->engine->stats.total_memory_size;
      uint64_t mem = actualTotalMemorySize + DefaultLimitsOfTotalMemorySize;
      LogInfof("mem limit reset in V8, actualTotalMemorySize: %ld, limit: %ld, tx.hash: %s", actualTotalMemorySize, DefaultLimitsOfTotalMemorySize, configBundle.tx_json().c_str());
      this->engine->limits_of_total_memory_size = mem;
    }

    v8ThreadContext ctx;
    memset(&ctx, 0x00, sizeof(ctx));
    SetRunScriptArgs(&ctx, this->engine, RUNSCRIPT, this->m_runnable_src.c_str(), this->m_traceale_src_line_offset, 1);
    ctx.input.lcs = this->m_lcs_handler;
    ctx.input.gcs = this->m_gcs_handler;
    bool btn = CreateScriptThread(&ctx);
    if (btn == false) {
      return NVM_UNEXPECTED_ERR;
    }

    if(ctx.output.result != nullptr){
      size_t strLength = strlen(ctx.output.result) + 1;
      this->m_exe_result = (char*)calloc(strLength, sizeof(char));
      strcpy(this->m_exe_result, ctx.output.result);
    }

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

    NVMDataRequest *request = new NVMDataRequest();
    //while(stream->Read(request)){

    stream->Read(request);

    std::string requestType = request->request_type();
    //google::protobuf::uint32 requestIndx = request->request_indx();
    google::protobuf::uint64 lcsHandler = request->lcs_handler();
    google::protobuf::uint64 gcsHandler = request->gcs_handler();

    if(requestType.compare(DATA_EXHG_START) == 0){

      NVMConfigBundle configBundle = request->config_bundle();
      CurrChainID = (uint32_t)configBundle.chain_id();
      std::string scriptSrc = configBundle.script_src();
      std::string scriptType = configBundle.script_type();
      std::string runnableSrc = configBundle.runnable_src();
      std::string moduleID = configBundle.module_id();
      google::protobuf::uint64 maxLimitsOfExecutionInstructions = configBundle.max_limits_of_execution_instruction();
      google::protobuf::uint64 defaultTotalMemSize = configBundle.default_limits_of_total_mem_size();
      google::protobuf::uint64 limitsOfExecutionInstructions = configBundle.limits_exe_instruction();
      google::protobuf::uint64 totalMemSize = configBundle.limits_total_mem_size();

      //bool enableLimits = configBundle.enable_limits();
      std::string blockJson = configBundle.block_json();
      std::string txJson = configBundle.tx_json();
      
      LogInfof(">>>>Script source is: %s", scriptSrc.c_str());
      LogInfof(">>>>Script type is: %s", scriptType.c_str());
      LogInfof(">>>>Runnable source is: %s", runnableSrc.c_str());
      LogInfof(">>>>Module id is: %s", moduleID.c_str());
      LogInfof(">>>>Blockjson is: %s", blockJson.c_str());
      LogInfof(">>>>TX json is: %s", txJson.c_str());
      LogInfof(">>>>lcsHandler %ld, gcsHandler %ld", lcsHandler, gcsHandler);
      LogInfof(">>>>The limit of exe instruction %ld", limitsOfExecutionInstructions);
      LogInfof(">>>>The limit of mem usage is: %ld", totalMemSize);

      this->m_runnable_src = runnableSrc;
      this->m_module_id = moduleID;
      this->m_lcs_handler = (uintptr_t)lcsHandler;
      this->m_gcs_handler = (uintptr_t)gcsHandler;
      int ret = this->StartScriptExecution(scriptSrc, scriptType, runnableSrc, moduleID, configBundle);
      
      if(this->m_exe_result == nullptr){
        this->m_exe_result = (char*)calloc(1, sizeof(char));
      }

      NVMDataResponse *response = new NVMDataResponse();
      NVMFinalResponse *finalResponse = new NVMFinalResponse();
      finalResponse->set_result(ret);
      finalResponse->set_msg(this->m_exe_result);

      NVMStatsBundle *statsBundle = new NVMStatsBundle();
      ReadExeStats(statsBundle);
      finalResponse->set_allocated_stats_bundle(statsBundle);

      response->set_allocated_final_response(finalResponse);
      response->set_response_type(DATA_EXHG_FINAL);
      response->set_response_indx(0);
      response->set_lcs_handler(lcsHandler);
      response->set_gcs_handler(gcsHandler);
    
      stream->Write(*response);

      if(this->m_exe_result != nullptr){
        free(this->m_exe_result);
        this->m_exe_result = nullptr;
      }

      if(FG_DEBUG)
        std::cout<<"\n\n\n\n\n\n"<<std::endl;
      
    }else{
      // throw exception since the request type is not allowed
      if(FG_DEBUG)
        std::cout<<"Illegal request type"<<std::endl;
    }

  }catch(const std::exception& e){
    if(FG_DEBUG)
      std::cout<<e.what()<<std::endl;
  }

  if(this->engine != nullptr){
    // release and delete engine
    DeleteEngine(this->engine);
    this->engine = nullptr;
  }

  return grpc::Status::OK;
}

void NVMEngine::LocalTest(){
  // compose testing data
  std::string moduleID ("contract.js");
  std::string scriptType ("js");
  std::string scriptSrc, runnableSrc;

  std::ifstream sourceFP("testcase/vault/source.txt");
  if(sourceFP.is_open()){
    std::stringstream buffer;
    buffer<<sourceFP.rdbuf();
    scriptSrc = buffer.str();
    sourceFP.close();
  }

  std::ifstream runnableSrcFP("testcase/vault/runnable_source.txt");
  if(runnableSrcFP.is_open()){
    std::stringstream buffer;
    buffer<<runnableSrcFP.rdbuf();
    runnableSrc = buffer.str();
    runnableSrcFP.close();
  }

  NVMConfigBundle* configBundle = new NVMConfigBundle();
  configBundle->set_limits_exe_instruction(400000000);
  configBundle->set_limits_total_mem_size(40000000);

  int ret = this->StartScriptExecution(scriptSrc, scriptType, runnableSrc, moduleID, *configBundle);

  if(FG_DEBUG){
    if(this->m_exe_result != nullptr)
      std::cout<<">>>>Hey running is done, and running result is: "<<this->m_exe_result<<std::endl;
    else
      std::cout<<">>>>Hey running is done, and the running result is null! and the ret is: "<<ret<<std::endl;
  }

  if(this->m_exe_result != nullptr)
    free(this->m_exe_result);

}

const NVMCallbackResult* NVMEngine::Callback(void* handler, NVMCallbackResponse* callback_response){
    if(this->m_stm != nullptr){
        const NVMCallbackResult *result;
        bool getResultFlag = false;
        NVMDataResponse *response = new NVMDataResponse();
        response->set_response_type(DATA_EXHG_CALL_BACK);
        response->set_response_indx(++this->m_response_indx);
        //response->set_lcs_handler(google::protobuf::uint64((uintptr_t)handler));
        response->set_lcs_handler((google::protobuf::uint64)this->m_lcs_handler);
        response->set_gcs_handler((google::protobuf::uint64)this->m_gcs_handler); // gcs handler is not used by now
        response->set_allocated_callback_response(callback_response);
        this->m_stm->Write(*response);

        // wait for the result and return
        NVMDataRequest *request = new NVMDataRequest();
        while(this->m_stm->Read(request)){
          std::string requestType = request->request_type();
          if(requestType.compare(DATA_EXHG_CALL_BACK) == 0){
            result = &(request->callback_result());
            getResultFlag = true;
            if(FG_DEBUG){
              std::cout<<"-- Callback response from the GOLANG side with function: "<<callback_response->func_name()<<std::endl;
              // check gas cnt one by one
              std::cout<<"********** CALLBACK result of "<<callback_response->func_name()<<" is: "<<result->result()<<std::endl;
              for(int i=0; i<result->extra_size(); i++){
                std::cout<<"************* CALLBACK extra: "<<result->extra(i)<<std::endl;
              }
            }
          }
          break;
        }
        free(request);
        if(!getResultFlag)
          result = nullptr;
        return result;

    }else{
        LogErrorf("Streaming object is not NULL");
    }
    return nullptr;
}

const NVMCallbackResult* DataExchangeCallback(void* handler, NVMCallbackResponse* response){
    if(gNVMEngine != nullptr){
        return gNVMEngine->Callback(handler, response);
    }else{
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
