// Copyright (C) 2017-2019 go-nebulas authors
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

#include "nvm_engine.h"
#include "compatibility.h"

//uint32_t CurrChainID = 1;   // default value

SNVM::NVMEngine* gNVMEngine = nullptr;

V8Engine* SNVM::NVMEngine::GetCurrentV8EngineInstance(){
  if(this->m_inner_engines != nullptr && this->m_inner_engines->size() > 0){
    V8Engine* top_engine = this->m_inner_engines->top();
    if(top_engine != nullptr)
      return top_engine;
  }
  return this->engine;
}

uintptr_t SNVM::NVMEngine::GetCurrentEngineLcsHandler(){
  uintptr_t curr_lcs = (uintptr_t)0;

  if(this->m_inner_engines != nullptr && this->m_inner_engines->size() > 0){
    V8Engine* curr_engine = this->m_inner_engines->top();
    curr_lcs = curr_engine->lcs;
    std::cout<<">>>>>>>>>>>>Got lcs handler from inner engine"<<std::endl;

  }else if(this->engine != nullptr){
    curr_lcs = this->engine->lcs;
    std::cout<<">>>>>>>>>>>>Got lcs handler from THIS engine"<<std::endl;
  }

  std::cout<<">>>>> Current lcs handler is: "<<(uint64)curr_lcs<<std::endl;
  return curr_lcs;
}

uintptr_t SNVM::NVMEngine::GetCurrentEngineGcsHandler(){
  uintptr_t curr_gcs = (uintptr_t)0;

  if(this->m_inner_engines != nullptr && this->m_inner_engines->size() > 0){
    V8Engine* curr_engine = this->m_inner_engines->top();
    curr_gcs = curr_engine->gcs;

  }else if(this->engine != nullptr){
    curr_gcs = this->engine->gcs;

  }
  return curr_gcs;
}

void SNVM::NVMEngine::ResetRuntimeStatus(){
    if(this->m_inner_engines != nullptr){
      while(!this->m_inner_engines->empty()){
        this->m_inner_engines->pop();
      }
    }
    if(this->m_exe_result != nullptr){
      free(this->m_exe_result);
      this->m_exe_result = nullptr;
    }
    if(this->m_inner_exe_result != nullptr){
      free(this->m_inner_exe_result);
      this->m_inner_exe_result = nullptr;
    }
}

int SNVM::NVMEngine::GetRunnableSourceCode(
  V8Engine* engine, 
  const std::string& sourceType, 
  const std::string& scriptHash, 
  const std::string& originalSource){

  const char* jsSource;
  uint64_t originalSourceLineOffset = 0;
  int m_src_offset = 0;

  if(sourceType.compare(this->TS_TYPE) == 0){
    jsSource = TranspileTypeScriptModuleThread(engine, originalSource.c_str(), &m_src_offset);
    if(jsSource == nullptr){
      return NVM_TRANSPILE_SCRIPT_ERR;
    }
  }else{
    jsSource = originalSource.c_str();
  }

  std::cout<<">>>>Hey, checking modules, with script hash: "<<scriptHash<<std::endl;
  if(this->srcModuleCache == nullptr)
    std::cout<<"<<<<<<<<<src module cache is null"<<std::endl;
  std::cout<<", srcmodulecache size: "<<this->srcModuleCache->size()<<std::endl;
  if(this->srcModuleCache->find(scriptHash)){
    CacheSrcItem cachedSourceItem = this->srcModuleCache->get(scriptHash);
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
      this->srcModuleCache->set(scriptHash, newItem);
      return 0;
    }

    if(FG_DEBUG && traceableSource!=nullptr){
      std::cout<<">>>>Traceable source code is: "<<traceableSource<<std::endl;
    }
  }
 
  return NVM_INJECT_TRACING_INSTRUCTION_ERR;
}


int SNVM::NVMEngine::StartScriptExecution(
    V8Engine* engine, 
    const std::string& contractSource, 
    const std::string& scriptType,
    const std::string& scriptHash,
    const std::string& runnableSrc, 
    const std::string& moduleID, 
    const NVMConfigBundle& configBundle,
    char* &exeResult){

    // transpile script and inject tracing code if necessary
    int runnableSourceResult = this->GetRunnableSourceCode(engine, scriptType, scriptHash, contractSource);
    if(runnableSourceResult != 0){
      LogErrorf("Failed to get runnable source code");
      return runnableSourceResult;
    }

    std::cout<<">>>>After getting runnable source code"<<std::endl;
    AddModule(engine, moduleID.c_str(), engine->traceable_src.c_str(), engine->traceable_line_offset);

    std::cout<<">>>>After adding module with ID: "<<moduleID<<std::endl;

    // check limitations
    if(this->config_bundle->block_height() >= this->m_compat_manager->GetNVMMemoryLimitWithoutInjectHeight()) {
      ReadMemoryStatistics(engine);
      uint64_t actualTotalMemorySize = (uint64_t)engine->stats.total_memory_size;
      uint64_t mem = actualTotalMemorySize + DefaultLimitsOfTotalMemorySize;
      LogInfof("mem limit reset in V8, actualTotalMemorySize: %ld, limit: %ld, tx.hash: %s", actualTotalMemorySize, DefaultLimitsOfTotalMemorySize, configBundle.tx_json().c_str());
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


void SNVM::NVMEngine::ReadExeStats(V8Engine* currEngine, NVMStatsBundle *statsBundle){
  ReadMemoryStatistics(currEngine);
  statsBundle->set_actual_count_of_execution_instruction((google::protobuf::uint64)currEngine->stats.count_of_executed_instructions);
  statsBundle->set_actual_used_mem_size((google::protobuf::uint64)currEngine->stats.total_memory_size);
}


grpc::Status SNVM::NVMEngine::SmartContractCall(
  grpc::ServerContext* context, 
  grpc::ServerReaderWriter<NVMDataResponse, NVMDataRequest>* stream){

  this->m_stm = stream;
  this->ResetRuntimeStatus();

  try{

    NVMDataRequest *request = new NVMDataRequest();
    stream->Read(request);

    std::string requestType = request->request_type();
    google::protobuf::uint64 lcsHandler = request->lcs_handler();
    google::protobuf::uint64 gcsHandler = request->gcs_handler();

    if(requestType.compare(DATA_EXHG_START) == 0){

      NVMConfigBundle configBundle = request->config_bundle();
      this->config_bundle = &configBundle;
      SetChainID((uint32_t)configBundle.chain_id());
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

      // create engine and inject tracing instructions
      this->engine = CreateEngine();
      this->engine->limits_of_executed_instructions = configBundle.limits_exe_instruction();
      this->engine->limits_of_total_memory_size = configBundle.limits_total_mem_size();
      this->engine->lcs = (uintptr_t)lcsHandler;
      this->engine->gcs = (uintptr_t)gcsHandler;
      if(blockHeight >= this->m_compat_manager->InnerContractCallAvailableHeight()){
        EnableInnerContract(this->engine);
      }

      std::cout<<">>>>After creating engine, the original lcs is: "<<lcsHandler<<std::endl;
      std::cout<<">>>>After creating engine, the lcs handler is: "<<this->engine->lcs<<std::endl;

      char* exeResult = nullptr;
      int ret = this->StartScriptExecution(this->engine, scriptSrc, scriptType, 
                            scriptHash, runnableSrc, moduleID, configBundle, exeResult);
      this->m_exe_result = exeResult;

      std::cout<<">>>>After starting script execution"<<std::endl;

      NVMDataResponse *response = new NVMDataResponse();
      NVMFinalResponse *finalResponse = new NVMFinalResponse();
      finalResponse->set_result(ret);
      finalResponse->set_not_null(this->m_exe_result != nullptr);

      if(this->m_exe_result == nullptr){
        this->m_exe_result = (char*)calloc(1, sizeof(char));
      }
      finalResponse->set_msg(this->m_exe_result);

      NVMStatsBundle *statsBundle = new NVMStatsBundle();
      ReadExeStats(this->engine, statsBundle);
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
      
    }else{
      // throw exception since the request type is not allowed
      if(FG_DEBUG)
        std::cout<<"Illegal request type"<<std::endl;
    }

  }catch(const std::exception& e){
    if(FG_DEBUG)
      std::cout<<e.what()<<std::endl;
  }

  DeleteEngine(this->engine);

  std::cout<<std::endl<<std::endl<<std::endl<<"&&&&&&&&&&& Finished EXE &&&&&&&&&&&"<<std::endl<<std::endl<<std::endl;

  return grpc::Status::OK;
}

NVMCallbackResult* SNVM::NVMEngine::Callback(void* handler, NVMCallbackResponse* callback_response, bool inner_call_flag=false){
    if(this->m_stm != nullptr){
        NVMDataResponse response;
        if(inner_call_flag){
          response.set_response_type(DATA_EXHG_INNER_CALL);
          NVMFinalResponse* innerFinalResponse = new NVMFinalResponse();
          NVMStatsBundle* innerStatsBundle = new NVMStatsBundle();
          V8Engine* topEngine = GetCurrentV8EngineInstance();
          ReadExeStats(topEngine, innerStatsBundle);
          innerFinalResponse->set_allocated_stats_bundle(innerStatsBundle);
          response.set_allocated_final_response(innerFinalResponse);

        }else{
          response.set_response_type(DATA_EXHG_CALL_BACK);
        }

        response.set_response_indx(++this->m_response_indx);
        response.set_lcs_handler((google::protobuf::uint64)this->GetCurrentEngineLcsHandler());
        response.set_gcs_handler((google::protobuf::uint64)this->GetCurrentEngineGcsHandler()); // gcs handler is not used by now
        std::cout<<"<><><><><> After setting handlers"<<std::endl;
        response.set_allocated_callback_response(callback_response);
        this->m_stm->Write(response);

        std::cout<<">>>>>After sending request: "<<std::endl;
        std::cout<<callback_response->func_name()<<std::endl;

        // wait for the result and return
        NVMDataRequest request;
        NVMCallbackResult* callback_result = nullptr;
        std::string requestType;
        while(this->m_stm->Read(&request)){
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
                ret = this->StartScriptExecution(newEngine, scriptSrc, scriptType, 
                                                  scriptHash, runnableSrc, moduleID, configBundle, exeResult);
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

void SNVM::NVMEngine::LocalTest(){
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

  this->engine = CreateEngine();
  NVMConfigBundle* configBundle = new NVMConfigBundle();
  configBundle->set_limits_exe_instruction(400000000);
  configBundle->set_limits_total_mem_size(40000000);

  char* exeResult = nullptr;
  int ret = this->StartScriptExecution(this->engine, scriptSrc, scriptType, scriptHash, runnableSrc, moduleID, *configBundle, exeResult);
  this->m_exe_result = exeResult;

  if(FG_DEBUG){
    if(this->m_exe_result != nullptr)
      std::cout<<">>>>Hey running is done, and running result is: "<<this->m_exe_result<<std::endl;
    else
      std::cout<<">>>>Hey running is done, and the running result is null! and the ret is: "<<ret<<std::endl;
  }

  if(this->m_exe_result != nullptr)
    free(this->m_exe_result);

}


V8Engine* SNVM::NVMEngine::CreateInnerContractEngine(
    const std::string& scriptType, 
    const std::string& scriptHash, 
    const std::string& innerContractSrc, 
    const std::string& moduleID,
    const uint64_t blockHeight,
    std::string& createResult){

    if(this->m_inner_engines == nullptr)
      this->m_inner_engines = std::unique_ptr<std::stack<V8Engine*>>(new std::stack<V8Engine*>());

    V8Engine* inner_engine = CreateEngine();
    // set limitation for the new engine
    if(inner_engine != nullptr)
      this->m_inner_engines->push(inner_engine);
    ReadMemoryStatistics(this->engine);
    if(blockHeight >= m_compat_manager->InnerContractCallAvailableHeight()){
      EnableInnerContract(inner_engine);
    }

    return inner_engine;
}

void SNVM::NVMEngine::PopInnerEngine(V8Engine* engine){
  if(engine == nullptr || m_inner_engines == nullptr)
    return;
  if(m_inner_engines->top() == engine)
    m_inner_engines->pop();
}

// start one level of inner contract call, check the resource usage firstly
//const void NVMEngine::StartInnerContractCall(const std::string& address, const std::string& valueStr, const std::string& funcName, const std::string& args){
const std::string SNVM::NVMEngine::ConfigBundleToString(NVMConfigBundle& configBundle){
  std::string res("chain id: " + std::to_string(configBundle.chain_id()) + ", script src: " + configBundle.script_src() + ", script type: " + configBundle.script_type()
    +  ", script hash: " + configBundle.script_hash() + ", runnable src: " + configBundle.runnable_src() + ", module id: " + configBundle.module_id());
  return res;
}

std::string SNVM::NVMEngine::FetchLibContent(const char* version_file_path){
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
  lib_content_cache->insert(std::pair<std::string, std::string>(std::string(version_file_path), std::string(content)));
  return std::string(content);
}

std::string SNVM::NVMEngine::FetchContractSrc(const char* sid, size_t* line_offset){
  LogInfof("NVMEngine::FetchContractSrc: %s", sid);
  std::cout<<"[ ----- CALLBACK ------ ] FetchContractSrc: "<<sid<<std::endl;

  if(engineSrcModules->find(std::string(sid))){
    m_mutex.lock();
    SourceInfo srcInfo = engineSrcModules->get(std::string(sid));
    *line_offset = srcInfo.lineOffset;
    m_mutex.unlock();
    return srcInfo.source;
  }
  return "";
}

void SNVM::NVMEngine::AddContractSrcToModules(const char* sid, const char* source_code, size_t line_offset){
  if(engineSrcModules->find(std::string(sid)) == false){
    m_mutex.lock();
    SourceInfo srcInfo;
    srcInfo.lineOffset = line_offset;
    srcInfo.source = std::string(source_code);
    engineSrcModules->set(std::string(sid), srcInfo);
    m_mutex.unlock();
    std::cout<<">>>>>>> Add contract source code into engine src modules"<<std::endl;
  }
}

std::string SNVM::NVMEngine::AttachNativeJSLibVersion(const char* lib_name){
  std::string lib_name_str(lib_name);
  std::string meta_version = config_bundle->meta_version();
  std::cout<<"$$$$$$$$$$ ********* Meta version: "<<meta_version<<std::endl;
  std::string res = m_compat_manager->AttachVersionForLib(lib_name_str, (uint64_t)config_bundle->block_height(), meta_version);
  std::cout<<"$$$$$$$$$$ Attached version for lib: "<<res<<std::endl;

  return res;
}



std::string SNVM::AttachNativeJSLibVersion(const char* lib_name){
  if(lib_name == nullptr)
    return nullptr;

  if(gNVMEngine != nullptr){
    try{
      return gNVMEngine->AttachNativeJSLibVersion(lib_name);
    }catch(const std::exception& e){
      LogErrorf("NVM engine is nil when attaching native js lib version");
    }
  }
  return nullptr;
}

void SNVM::AddContractSrcToModules(const char* sid, const char* source_code, size_t line_offset){
  if(sid == nullptr || source_code == nullptr)
    return;

  if(gNVMEngine != nullptr){
    try{
      gNVMEngine->AddContractSrcToModules(sid, source_code, line_offset);
    }catch(const std::exception& e){
      LogErrorf("NVM engine is nil when adding contract source code to modules");
    }
  }
}

std::string SNVM::FetchContractSrcFromModules(const char* sid, size_t* line_offset){
  if(sid == nullptr)
    return "";
  
  if(gNVMEngine != nullptr){
    try{
      return gNVMEngine->FetchContractSrc(sid, line_offset);
    }catch(const std::exception& e){
      LogErrorf("NVM engine is nil when fetching contract source code from cache");
    }
  }
  return "";
}

std::string SNVM::FetchNativeJSLibContentFromCache(const char* file_path){
  if(file_path == nullptr)
    return "";

  if(gNVMEngine != nullptr){
    try{
      return gNVMEngine->FetchLibContent(file_path);
    }catch(const std::exception& e){
      LogErrorf("NVM engine is nil when fetching native js lib content from cache");
    }
  }
  return "";
}

const NVMCallbackResult* SNVM::DataExchangeCallback(void* handler, NVMCallbackResponse* response, bool inner_call_flag){

    if(gNVMEngine != nullptr){
        std::cout<<">>>>nvmengine is not null!"<<std::endl;
        return gNVMEngine->Callback(handler, response, inner_call_flag);
    }else{
        std::cout<<">>>>>>> Failed to exchange data"<<std::endl;
        LogErrorf("Failed to exchange data");
    }
    return nullptr;
}