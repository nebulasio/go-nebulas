// Copyright (C) 2017-2019 go-nebulas
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
// <http://www.gnu.org/licenses/>
//
// Author: Samuel Chen <samuel.chen@nebulas.io>

#include "nvm_engine.h"

SNVM::NVMDaemon* gNVMDaemon = nullptr;

// Basically, this function is used to get rid of "." and ".." in the file path
void reformatModuleId(char *dst, const char *src) {
  std::string s(src);
  std::string delimiter("/");
  std::vector<std::string> paths;

  size_t pos = 0;
  while ((pos = s.find(delimiter)) != std::string::npos) {
    std::string p = s.substr(0, pos);
    s.erase(0, pos + delimiter.length());

    if (p.length() == 0 || p.compare(".") == 0) {
      continue;
    }

    if (p.compare("..") == 0) {
      if (paths.size() > 0) {
        paths.pop_back();
        continue;
      }
    }
    paths.push_back(p);
  }
  paths.push_back(s);

  std::stringstream ss;
  for (size_t i = 0; i < paths.size(); ++i) {
    if (i != 0)
      ss << "/";
    ss << paths[i];
  }

  strcpy(dst, ss.str().c_str());
}

V8Engine* SNVM::SCContext::GetCurrentV8EngineInstance(){
  if(this->m_inner_engines != nullptr && this->m_inner_engines->size() > 0){
    V8Engine* top_engine = this->m_inner_engines->top();
    if(top_engine != nullptr)
      return top_engine;
  }
  return this->engine;
}

uintptr_t SNVM::SCContext::GetCurrentEngineLcsHandler(){
  uintptr_t curr_lcs = (uintptr_t)0;

  if(this->m_inner_engines != nullptr && this->m_inner_engines->size() > 0){
    V8Engine* curr_engine = this->m_inner_engines->top();
    curr_lcs = curr_engine->lcs;
  }else if(this->engine != nullptr){
    curr_lcs = this->engine->lcs;
  }
  return curr_lcs;
}

uintptr_t SNVM::SCContext::GetCurrentEngineGcsHandler(){
  uintptr_t curr_gcs = (uintptr_t)0;

  if(this->m_inner_engines != nullptr && this->m_inner_engines->size() > 0){
    V8Engine* curr_engine = this->m_inner_engines->top();
    curr_gcs = curr_engine->gcs;
  }else if(this->engine != nullptr){
    curr_gcs = this->engine->gcs;
  }
  return curr_gcs;
}

int SNVM::SCContext::GetRunnableSourceCode(
  V8Engine* engine, 
  const std::string& sourceType, 
  const std::string& scriptHash, 
  const std::string& originalSource){

  const char* jsSource;
  uint64_t originalSourceLineOffset = 0;
  int m_src_offset = 0;

  if(sourceType.compare(SNVM::TS_TYPE) == 0){
    jsSource = TranspileTypeScriptModuleThread(engine, originalSource.c_str(), &m_src_offset);
    if(jsSource == nullptr){
      return NVM_TRANSPILE_SCRIPT_ERR;
    }
  }else{
    jsSource = originalSource.c_str();
  }

  std::cout<<">>>>Hey, checking modules, with script hash: "<<scriptHash<<std::endl;
  std::cout<<", srcmodulecache size: "<<this->m_daemon->TraceableSrcModuleCacheSize()<<std::endl;
  if(this->m_daemon->IsTraceableSrcModuleExist(scriptHash)){
    CacheSrcItem cachedSourceItem = this->m_daemon->FindTraceableSrcModuleFromCache(scriptHash);
    engine->traceable_src = cachedSourceItem.traceableSource;
    engine->traceable_line_offset = cachedSourceItem.traceableSourceLineOffset;
    std::cout<<">>>>Found src from src module cache"<<std::endl;
    return 0;

  }else{
    std::cout<<">>>>Before injecting tracing instruction"<<std::endl;
    char* traceableSource = InjectTracingInstructionsThread(engine, jsSource, &m_src_offset, this->m_allow_usage);
    if(traceableSource != nullptr){
      engine->traceable_src = std::string(traceableSource);
      engine->traceable_line_offset = 0;
      CacheSrcItem newItem = {originalSource, originalSourceLineOffset, traceableSource, engine->traceable_line_offset};
      // Check size and clear the cache if necessary, we actually do not need to hold all the code in cache
      if(this->m_daemon->TraceableSrcModuleCacheSize() > SNVM::SRC_MODULE_CACHE_SIZE){
        this->m_daemon->ClearTraceableSrcModuleCache();
      }
      this->m_daemon->AddIntoTraceableSrcModuleCache(scriptHash, newItem);
      return 0;
    }

    if(FG_DEBUG && traceableSource!=nullptr){
      std::cout<<">>>>Traceable source code is: "<<traceableSource<<std::endl;
    }
  }
 
  return NVM_INJECT_TRACING_INSTRUCTION_ERR;
}


int SNVM::SCContext::StartScriptExecution(
    const std::string& contractSource, 
    const std::string& scriptType,
    const std::string& scriptHash,
    const std::string& runnableSrc, 
    const std::string& moduleID,
    char* &exeResult){

    // transpile script and inject tracing code if necessary
    int runnableSourceResult = this->GetRunnableSourceCode(engine, scriptType, scriptHash, contractSource);
    if(runnableSourceResult != 0){
      LogErrorf("Failed to get runnable source code");
      return runnableSourceResult;
    }

    std::cout<<">>>>After getting runnable source code"<<std::endl;
    AddEngineSrcModule(engine, moduleID.c_str(), engine->traceable_src.c_str(), engine->traceable_line_offset);
    std::cout<<">>>>After adding module with ID: "<<moduleID<<std::endl;

    // check limitations
    if(this->config_bundle->block_height() >= this->m_daemon->GetNVMMemoryLimitWithoutInjectHeight(this->m_chain_id)) {
      ReadMemoryStatistics(engine);
      uint64_t actualTotalMemorySize = (uint64_t)engine->stats.total_memory_size;
      uint64_t mem = actualTotalMemorySize + DefaultLimitsOfTotalMemorySize;
      LogInfof("mem limit reset in V8, actualTotalMemorySize: %ld, limit: %ld, tx.hash: %s", 
        actualTotalMemorySize, DefaultLimitsOfTotalMemorySize, config_bundle->tx_json().c_str());
      engine->limits_of_total_memory_size = mem;
    }

    v8ThreadContext ctx;
    memset(&ctx, 0x00, sizeof(ctx));
    SetRunScriptArgs(&ctx, engine, RUNSCRIPT, runnableSrc.c_str(), engine->traceable_line_offset, 1);

    std::cout<<">>>>After setting runscript args"<<std::endl;

    bool btn = CreateScriptThread(&ctx);      // start exe thread
    if (btn == false) {
      return NVM_UNEXPECTED_ERR;
    }

    std::cout<<">>>>After creating script thread"<<std::endl;

    if(ctx.output.result != nullptr && ctx.output.result != NULL){
      size_t strLength = strlen(ctx.output.result) + 1;
      exeResult = (char*)calloc(strLength, sizeof(char));
      strcpy(exeResult, ctx.output.result);
    }

    if(FG_DEBUG && ctx.output.result!=nullptr)
      std::cout<<"$$$$$$$$$$$$  The ret is: "<<ctx.output.ret<<", while the result is: "
          <<ctx.output.result<<", exeresult: "<<exeResult<<std::endl;

    return ctx.output.ret;
}


void SNVM::SCContext::ReadExeStats(NVMStatsBundle *statsBundle){
  if(this->engine != nullptr){
    V8Engine* curr_engine = this->engine;
    ReadMemoryStatistics(curr_engine);
    statsBundle->set_actual_count_of_execution_instruction((google::protobuf::uint64)curr_engine->stats.count_of_executed_instructions);
    statsBundle->set_actual_used_mem_size((google::protobuf::uint64)curr_engine->stats.total_memory_size);
  }
}


grpc::Status SNVM::NVMDaemon::SmartContractCall(
  grpc::ServerContext* context, 
  grpc::ServerReaderWriter<NVMDataResponse, NVMDataRequest>* stream){

  V8Engine* new_engine = nullptr;
  SCContext* sc_ctx = nullptr;
  try{
    NVMDataRequest *request = new NVMDataRequest();
    stream->Read(request);

    std::string requestType = request->request_type();
    google::protobuf::uint64 lcsHandler = request->lcs_handler();
    google::protobuf::uint64 gcsHandler = request->gcs_handler();

    if(requestType.compare(DATA_EXHG_START) == 0){
      this->is_sc_running = true;
      NVMConfigBundle configBundle = request->config_bundle(); 
      std::string scriptSrc = configBundle.script_src();
      std::string scriptType = configBundle.script_type();
      std::string scriptHash = configBundle.script_hash();
      std::string runnableSrc = configBundle.runnable_src();
      std::string moduleID = configBundle.module_id();
      uint64_t blockHeight = configBundle.block_height();
      google::protobuf::uint64 maxLimitsOfExecutionInstructions = configBundle.max_limits_of_execution_instruction();
      google::protobuf::uint64 defaultTotalMemSize = configBundle.default_limits_of_total_mem_size();
      google::protobuf::uint64 limitsOfExecutionInstructions = configBundle.limits_exe_instruction();
      google::protobuf::uint64 totalMemSize = configBundle.limits_total_mem_size();

      //bool enableLimits = configBundle.enable_limits();
      std::string blockJson = configBundle.block_json();
      std::string txJson = configBundle.tx_json();
      std::string metaVersion = configBundle.meta_version();

      std::cout<<">>>>The configBundle is: "<<scriptHash<<", module id: "<<moduleID<<std::endl;
      std::cout<<">>>>Now receiving request from golang side, with lcshandler: "<<lcsHandler<<std::endl;
      std::cout<<">>>>*************** The meta version is: "<<metaVersion<<std::endl;
      
      LogInfof(">>>>Script source is: %s", scriptSrc.c_str());
      LogInfof(">>>>Script type is: %s", scriptType.c_str());
      LogInfof(">>>>Runnable source is: %s", runnableSrc.c_str());
      LogInfof(">>>>Module id is: %s", moduleID.c_str());
      LogInfof(">>>>Blockjson is: %s", blockJson.c_str());
      LogInfof(">>>>TX json is: %s", txJson.c_str());
      LogInfof(">>>>lcsHandler %ld, gcsHandler %ld", lcsHandler, gcsHandler);
      LogInfof(">>>>The limit of exe instruction %ld", limitsOfExecutionInstructions);
      LogInfof(">>>>The limit of mem usage is: %ld", totalMemSize);

      sc_ctx = new SCContext(this);
      uint32_t curr_chain_id = (uint32_t)configBundle.chain_id();
      sc_ctx->SetChainID(curr_chain_id);
      sc_ctx->SetStream(stream);
      new_engine = CreateV8Engine(blockHeight, curr_chain_id);
      new_engine->limits_of_executed_instructions = configBundle.limits_exe_instruction();
      new_engine->limits_of_total_memory_size = configBundle.limits_total_mem_size();
      new_engine->lcs = (uintptr_t)lcsHandler;
      new_engine->gcs = (uintptr_t)gcsHandler;
      if(blockHeight >= InnerContractCallAvailableHeight(curr_chain_id)){
        EnableInnerContract(new_engine);
      }
      sc_ctx->SetV8Engine(new_engine);
      std::cout<<">>>>After creating engine, the original lcs is: "<<lcsHandler<<std::endl;
      std::cout<<">>>>After creating engine, the lcs handler is: "<<new_engine->lcs<<std::endl;

      sc_ctx->SetConfigBundle(&configBundle);
      AddSCContext(new_engine, sc_ctx);
      char* exeResult = nullptr;
      int ret = sc_ctx->StartScriptExecution(scriptSrc, scriptType, scriptHash, runnableSrc, moduleID, exeResult);

      if(exeResult == nullptr)
        std::cout<<">>>>the exe result is NULL"<<std::endl;
      else
        std::cout<<">>>>After starting script execution, exe result is: "<<exeResult<<std::endl;
      NVMDataResponse *response = new NVMDataResponse();
      NVMFinalResponse *finalResponse = new NVMFinalResponse();
      finalResponse->set_result(ret);
      finalResponse->set_not_null(exeResult != nullptr);

      if(exeResult == nullptr){
        exeResult = (char*)calloc(1, sizeof(char));
      }
      finalResponse->set_msg(exeResult);

      NVMStatsBundle *statsBundle = new NVMStatsBundle();
      sc_ctx->ReadExeStats(statsBundle);
      finalResponse->set_allocated_stats_bundle(statsBundle);
      response->set_allocated_final_response(finalResponse);
      response->set_response_type(DATA_EXHG_FINAL);
      response->set_response_indx(0);
      response->set_lcs_handler(lcsHandler);
      response->set_gcs_handler(gcsHandler);
    
      stream->Write(*response);

      if(new_engine != nullptr){
        RemoveSCContext(new_engine);
        DeleteEngine(new_engine);
        new_engine = nullptr;
        sc_ctx = nullptr;
      }
      if(exeResult != nullptr){
        free(exeResult);
      }
      if(sc_ctx != nullptr)
        delete sc_ctx;
      
    }else{
      // throw exception since the request type is not allowed
      if(FG_DEBUG)
        std::cout<<"Illegal request type"<<std::endl;
    }

  }catch(const std::exception& e){
    if(FG_DEBUG)
      std::cout<<e.what()<<std::endl;
  }

  if(new_engine != nullptr)
    DeleteEngine(new_engine);
  if(sc_ctx != nullptr)
    delete sc_ctx;
  this->is_sc_running = false;
  std::cout<<std::endl<<std::endl<<std::endl<<"&&&&&&&&&&& Finished EXE &&&&&&&&&&&"<<std::endl<<std::endl<<std::endl;

  return grpc::Status::OK;
}

NVMCallbackResult* SNVM::SCContext::Callback(void* handler, NVMCallbackResponse* callback_response, bool inner_call_flag=false){
    if(this->m_stream != nullptr){
        NVMDataResponse response;
        if(inner_call_flag){
          response.set_response_type(DATA_EXHG_INNER_CALL);
          NVMFinalResponse* innerFinalResponse = new NVMFinalResponse();
          NVMStatsBundle* innerStatsBundle = new NVMStatsBundle();
          V8Engine* topEngine = GetCurrentV8EngineInstance();
          ReadExeStats(innerStatsBundle);
          innerFinalResponse->set_allocated_stats_bundle(innerStatsBundle);
          response.set_allocated_final_response(innerFinalResponse);
        }else{
          response.set_response_type(DATA_EXHG_CALL_BACK);
        }

        response.set_response_indx(++this->m_response_indx);
        response.set_lcs_handler((google::protobuf::uint64)this->GetCurrentEngineLcsHandler());
        response.set_gcs_handler((google::protobuf::uint64)this->GetCurrentEngineGcsHandler());     // gcs handler is not used by now
        std::cout<<"<><><><><> After setting handlers"<<std::endl;
        response.set_allocated_callback_response(callback_response);
        this->m_stream->Write(response);

        std::cout<<">>>>>After sending request: "<<std::endl;
        std::cout<<callback_response->func_name()<<std::endl;

        // wait for the result and return
        NVMDataRequest request;
        NVMCallbackResult* callback_result = nullptr;
        std::string requestType;
        while(this->m_stream->Read(&request)){
          requestType = request.request_type();
          callback_result = new NVMCallbackResult(request.callback_result());

          if(requestType.compare(DATA_EXHG_CALL_BACK) == 0){
            std::cout<<"********** CALLBACK result of "<<callback_response->func_name()<<" is: "<<callback_result->result()<<std::endl;

          }else if(requestType.compare(DATA_EXHG_INNER_CALL) == 0){
            // this is NOT the final result of this call, we need to create new engine and execute the contract firstly
            std::cout<<">>>>>>>NOW START to call the new contract"<<std::endl;
            if(callback_result->not_null()){
              NVMConfigBundle configBundle = request.config_bundle();
              callback_result->clear_extra();
              std::string scriptSrc = configBundle.script_src();
              std::string scriptType = configBundle.script_type();
              std::string scriptHash = configBundle.script_hash();
              std::string runnableSrc = configBundle.runnable_src();
              std::string moduleID = configBundle.module_id();
              uint64_t blockHeight = configBundle.block_height();
              google::protobuf::uint64 maxLimitsOfExecutionInstructions = configBundle.max_limits_of_execution_instruction();
              google::protobuf::uint64 defaultTotalMemSize = configBundle.default_limits_of_total_mem_size();
              google::protobuf::uint64 limitsOfExecutionInstructions = configBundle.limits_exe_instruction();
              google::protobuf::uint64 totalMemSize = configBundle.limits_total_mem_size();
              std::string blockJson = configBundle.block_json();
              std::string txJson = configBundle.tx_json();
              std::string metaVersion = configBundle.meta_version();

              std::string createResult;
              int ret = NVM_SUCCESS;
              char* exeResult = nullptr;
              V8Engine* newEngine = this->CreateInnerContractEngine(scriptType, scriptHash, scriptSrc, moduleID, blockHeight, createResult);
              if(newEngine != nullptr){
                newEngine->limits_of_executed_instructions = configBundle.limits_exe_instruction();
                newEngine->limits_of_total_memory_size = configBundle.limits_total_mem_size();
                newEngine->lcs = (uintptr_t)request.lcs_handler();
                newEngine->gcs = (uintptr_t)request.gcs_handler();
                std::cout<<">>>>>New engine set lcs: "<<(uint64)newEngine->lcs<<std::endl;

                // NOTE: this creates a new thread to execute the new contract
                std::cout<<"@@@@@@@@@@@@ Before inner contract exe"<<std::endl;
                ret = this->StartScriptExecution(scriptSrc, scriptType, 
                                                  scriptHash, runnableSrc, moduleID, exeResult);
                std::cout<<"@@@@@@@@@@@@ Inner contract exe result is: "<<ret<<std::endl;
                if(exeResult == nullptr)
                  std::cout<<"Exeresult is null"<<std::endl;
                else
                  std::cout<<"ExeResult is: "<<exeResult<<std::endl;

                // Check execution result and return to the caller
                if(ret != NVM_SUCCESS){
                  ret = NVM_INNER_EXE_ERR;
                  if(exeResult != nullptr)
                    callback_result->set_result("");
                  else
                    callback_result->set_result(std::string(exeResult));
                  callback_result->set_not_null(false);
                }else{
                  callback_result->set_result(std::string(exeResult));
                  callback_result->set_not_null(true);
                }
                // calculate gas fee
                ReadMemoryStatistics(newEngine);
                size_t gasCnt = newEngine->stats.count_of_executed_instructions;
                size_t callbackGasCnt = (size_t)std::stoull((request.callback_result()).extra(0));
                std::cout<<">>>>> new engine execution gas cnt: "<<gasCnt<<", callback gas cnt: "<<callbackGasCnt<<std::endl;
                gasCnt += callbackGasCnt;
                callback_result->add_extra(std::to_string(gasCnt));
                // pop engine from the stack
                this->PopInnerEngine(newEngine);
                DeleteEngine(newEngine);

              }else{
                callback_result->set_result("");
                callback_result->add_extra(request.callback_result().extra(0));
                callback_result->set_not_null(false);

              }

            }
          }
          break;
        }

        if(callback_result != nullptr)
          std::cout<<">>>> Return a none callback result!!!!"<<callback_result<<std::endl;
        return callback_result;

    }else{
        LogErrorf("Streaming object is not NULL");
    }
    return nullptr;
}

void SNVM::NVMDaemon::LocalTest(){
  // compose testing data
  std::string moduleID ("contract.js");
  std::string scriptType ("js");
  std::string scriptHash ("0x123456789");
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

  SCContext* sc_ctx = new SCContext(gNVMDaemon);
  V8Engine* engine = CreateV8Engine(0L, LocalNetID);
  NVMConfigBundle* configBundle = new NVMConfigBundle();
  configBundle->set_limits_exe_instruction(400000000);
  configBundle->set_limits_total_mem_size(40000000);
  sc_ctx->SetV8Engine(engine);
  sc_ctx->SetConfigBundle(configBundle);

  char* exeResult = nullptr;
  int ret = sc_ctx->StartScriptExecution(scriptSrc, scriptType, scriptHash, runnableSrc, moduleID, exeResult);

  if(FG_DEBUG){
    if(exeResult != nullptr)
      std::cout<<">>>>Hey running is done, and running result is: "<<exeResult<<std::endl;
    else
      std::cout<<">>>>Hey running is done, and the running result is null! and the ret is: "<<ret<<std::endl;
  }

  if(exeResult != nullptr)
    free(exeResult);
  if(engine != nullptr)
    DeleteEngine(engine);
  if(sc_ctx != nullptr)
    delete sc_ctx;
}


V8Engine* SNVM::SCContext::CreateInnerContractEngine(
    const std::string& scriptType, 
    const std::string& scriptHash, 
    const std::string& innerContractSrc, 
    const std::string& moduleID,
    const uint64_t blockHeight,
    std::string& createResult){

    if(this->m_inner_engines == nullptr)
      this->m_inner_engines = std::unique_ptr<std::stack<V8Engine*>>(new std::stack<V8Engine*>());
    V8Engine* inner_engine = m_daemon->CreateV8Engine(blockHeight, m_chain_id);
    if(inner_engine != nullptr)
      this->m_inner_engines->push(inner_engine);
    ReadMemoryStatistics(this->engine);
    if(blockHeight >= this->m_daemon->InnerContractCallAvailableHeight(this->m_chain_id)){
      EnableInnerContract(inner_engine);
    }
    return inner_engine;
}

void SNVM::SCContext::PopInnerEngine(V8Engine* engine){
  if(engine == nullptr || m_inner_engines == nullptr)
    return;
  if(m_inner_engines->top() == engine){
    m_inner_engines->pop();
  }
}

const std::string SNVM::SCContext::ConfigBundleToString(NVMConfigBundle& configBundle){
  std::string res("chain id: " + std::to_string(configBundle.chain_id()) + ", script src: " + configBundle.script_src() + ", script type: " + configBundle.script_type()
    +  ", script hash: " + configBundle.script_hash() + ", runnable src: " + configBundle.runnable_src() + ", module id: " + configBundle.module_id());
  return res;
}


std::string SNVM::SCContext::FetchEngineSrc(const std::string& sid, size_t* line_offset){
  LogInfof("NVMEngine::FetchContractSrc: %s", sid.c_str());
  std::cout<<"[ ----- CALLBACK ------ ] FetchContractSrc: "<<sid<<", engine module size: "<<engineSrcModules->size()<<std::endl;

  auto engineSrc = engineSrcModules->find(sid);
  if(engineSrc != engineSrcModules->end()){
    SourceInfo srcInfo = engineSrc->second;
    *line_offset = srcInfo.lineOffset;
    std::cout<<">>>>> FetchContractSrc: found the source code"<<std::endl;
    return srcInfo.source;
  }
  std::cout<<">>>>Failed to find the engine source code with ID: "<<sid<<std::endl;
  for(auto const& iter: *engineSrcModules){
    std::cout<<">>>>[key-value]: "<<iter.first<<std::endl;
  }
  return "";
}

void SNVM::SCContext::AddEngineSrcModule(void *engineptr, const char *filename, const char *source, size_t line_offset) {
  char filepath[128];
  if (strncmp(filename, "/", 1) != 0 && strncmp(filename, "./", 2) != 0 &&
      strncmp(filename, "../", 3) != 0) {
    sprintf(filepath, "lib/%s", filename);
    reformatModuleId(filepath, filepath);
  } else {
    reformatModuleId(filepath, filename);
  }
  std::cout<<"<><><>><>< Adding module: engine - "<<(uintptr_t)engineptr<<", filename-"<<filename<<std::endl;

  char sid[128];
  // 140568151729856:lib/blockchain.js
  sprintf(sid, "%zu:%s", (uintptr_t)engineptr, filepath);
  std::string ssid(sid);

  if(engineSrcModules->find(ssid) == engineSrcModules->end()){
    SourceInfo srcInfo;
    srcInfo.lineOffset = line_offset;
    srcInfo.source = std::string(source);
    engineSrcModules->insert(std::pair<std::string, SourceInfo>(std::string(sid), srcInfo));
    std::cout<<">>>>>>> Add contract source code into engine src modules with ID: "
      <<sid<<", engine source module size: "<<engineSrcModules->size()<<std::endl;
  }
}


std::string SNVM::NVMDaemon::FetchLibContent(const char* version_file_path){
  if(version_file_path == nullptr)
    return "";
  std::cout<<"&&&&&&&&&& version file path: "<<version_file_path<<std::endl;

  auto find_iter = lib_content_cache->find(std::string(version_file_path));
  if( find_iter != lib_content_cache->end())
    return find_iter->second;
  
  // read content from file path
  size_t file_size = 0;
  char* content = readFile(version_file_path, &file_size);
  if (content == NULL) {
    return "";
  }
  m_mutex.lock();
  lib_content_cache->insert(std::pair<std::string, std::string>(std::string(version_file_path), std::string(content)));
  m_mutex.unlock();

  return std::string(content);
}

std::string SNVM::NVMDaemon::AttachNativeJSLibVersion(
  const char* lib_name, 
  uint64_t block_height, 
  std::string meta_version,
  uint32_t chain_id){
  std::string lib_name_str(lib_name);
  std::cout<<"$$$$$$$$$$ ********* Meta version: "<<meta_version<<std::endl;
  std::string res = AttachVersionForLib(lib_name_str, block_height, meta_version, chain_id);
  std::cout<<"$$$$$$$$$$ Attached version for lib: "<<res<<std::endl;
  return res;
}


std::string SNVM::AttachNativeJSLibVersion(V8Engine* engine, const char* lib_name){
  if(lib_name == nullptr || engine == nullptr)
    return nullptr;

  if(gNVMDaemon != nullptr){
    try{
      SCContext* curr_ctx = gNVMDaemon->GetSCContext(engine);
      if(curr_ctx != nullptr){
        std::cout<<">>>>>SCContext is not nil"<<std::endl;
        return gNVMDaemon->AttachNativeJSLibVersion(lib_name, 
                  curr_ctx->GetBlockHeight(), curr_ctx->GetMetaVersion(), curr_ctx->GetChainID());
      }
      std::cout<<">>>>SCContext is NULL"<<std::endl;
    }catch(const std::exception& e){
      LogErrorf("NVM engine is nil when attaching native js lib version");
    }
  }
  return nullptr;
}

std::string SNVM::FetchEngineContractSrcFromModules(V8Engine* engine, const std::string& sid, size_t* line_offset){
  if(sid.compare("") == 0 || engine == nullptr)
    return "";
  
  if(gNVMDaemon != nullptr){
    try{
      SCContext* curr_ctx = gNVMDaemon->GetSCContext(engine);
      if(curr_ctx != nullptr){
        return curr_ctx->FetchEngineSrc(sid, line_offset);
      }
    }catch(const std::exception& e){
      LogErrorf("NVM engine is nil when fetching contract source code from cache");
    }
  }
  return "";
}

std::string SNVM::FetchNativeJSLibContentFromCache(const char* file_path){
  if(file_path == nullptr)
    return "";

  if(gNVMDaemon != nullptr){
    try{
      return gNVMDaemon->FetchLibContent(file_path);
    }catch(const std::exception& e){
      LogErrorf("NVM engine is nil when fetching native js lib content from cache");
    }
  }
  return "";
}

const NVMCallbackResult* SNVM::DataExchangeCallback(
  V8Engine* engine, 
  void* handler, 
  NVMCallbackResponse* response, 
  bool inner_call_flag){

    if(gNVMDaemon != nullptr){
        SCContext* curr_ctx = gNVMDaemon->GetSCContext(engine);
        if(curr_ctx != nullptr){
          std::cout<<">>>>nvmengine is not null!"<<std::endl;
          return curr_ctx->Callback(handler, response, inner_call_flag);
        }
    }else{
        std::cout<<">>>>>>> Failed to exchange data"<<std::endl;
        LogErrorf("Failed to exchange data");
    }
    return nullptr;
}