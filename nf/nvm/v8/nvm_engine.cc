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

uint32_t CurrChainID = 1;   // default value

uintptr_t NVMEngine::GetCurrentEngineLcsHandler(){
  if(this->m_inner_engines->size() > 0){
    V8Engine* curr_engine = this->m_inner_engines->top();
    return curr_engine->lcs;

  }else if(this->engine != nullptr){
    return this->engine->lcs;

  }else{
    return (uintptr_t)0;
  }
}

uintptr_t NVMEngine::GetCurrentEngineGcsHandler(){
  if(this->m_inner_engines->size() > 0){
    V8Engine* curr_engine = this->m_inner_engines->top();
    return curr_engine->gcs;

  }else if(this->engine != nullptr){
    return this->engine->gcs;

  }else{
    return (uintptr_t)0;
  }
}

int NVMEngine::GetRunnableSourceCode(V8Engine* engine, const std::string& sourceType, const std::string& scriptHash, const std::string& originalSource){

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

  auto searchRecord = srcModuleCache->find(scriptHash);
  if(searchRecord != srcModuleCache->end()){
    CacheSrcItem cachedSourceItem = searchRecord->second;
    engine->traceable_src = cachedSourceItem.traceableSource;
    engine->traceable_line_offset = cachedSourceItem.traceableSourceLineOffset;
    return 0;

  }else{
    char* traceableSource = InjectTracingInstructionsThread(engine, jsSource, &m_src_offset, this->m_allow_usage);
    if(traceableSource != nullptr){
      engine->traceable_src = std::string(traceableSource);
      engine->traceable_line_offset = 0;
      CacheSrcItem newItem = {originalSource, originalSourceLineOffset, traceableSource, engine->traceable_line_offset};
      srcModuleCache->insert({scriptHash, newItem});
      return 0;
    }

    if(FG_DEBUG && traceableSource!=nullptr){
      std::cout<<">>>>Traceable source code is: "<<traceableSource<<std::endl;
    }
  }
 
  return NVM_INJECT_TRACING_INSTRUCTION_ERR;
}



int NVMEngine::StartScriptExecution(
    V8Engine* engine, 
    const std::string& contractSource, 
    const std::string& scriptType,
    const std::string& scriptHash,
    const std::string& runnableSrc, 
    const std::string& moduleID, 
    const NVMConfigBundle& configBundle,
    char* exeResult){

    // transpile script and inject tracing code if necessary
    int runnableSourceResult = this->GetRunnableSourceCode(engine, scriptType, scriptHash, contractSource);
    if(runnableSourceResult != 0){
      LogErrorf("Failed to get runnable source code");
      return runnableSourceResult;
    }

    AddModule(engine, moduleID.c_str(), engine->traceable_src.c_str(), engine->traceable_line_offset);

    // check limitations
    if(this->config_bundle->block_height() >= GetNVMMemoryLimitWithoutInjectHeight()) {
      ReadMemoryStatistics(engine);
      uint64_t actualTotalMemorySize = (uint64_t)engine->stats.total_memory_size;
      uint64_t mem = actualTotalMemorySize + DefaultLimitsOfTotalMemorySize;
      LogInfof("mem limit reset in V8, actualTotalMemorySize: %ld, limit: %ld, tx.hash: %s", actualTotalMemorySize, DefaultLimitsOfTotalMemorySize, configBundle.tx_json().c_str());
      engine->limits_of_total_memory_size = mem;
    }

    v8ThreadContext ctx;
    memset(&ctx, 0x00, sizeof(ctx));
    SetRunScriptArgs(&ctx, engine, RUNSCRIPT, runnableSrc.c_str(), engine->traceable_line_offset, 1);
    //ctx.input.lcs = lcsHandler;
    //ctx.input.gcs = gcsHandler;
    bool btn = CreateScriptThread(&ctx);      // start exe thread
    if (btn == false) {
      return NVM_UNEXPECTED_ERR;
    }

    if(ctx.output.result != nullptr && ctx.output.result != NULL){
      size_t strLength = strlen(ctx.output.result) + 1;
      exeResult = (char*)calloc(strLength, sizeof(char));
      strcpy(exeResult, ctx.output.result);
    }

    if(FG_DEBUG && ctx.output.result!=nullptr)
      std::cout<<"$$$$$$$$$$$$  The ret is: "<<ctx.output.ret<<", while the result is: "<<ctx.output.result<<std::endl;

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
    stream->Read(request);

    std::string requestType = request->request_type();
    //google::protobuf::uint32 requestIndx = request->request_indx();
    google::protobuf::uint64 lcsHandler = request->lcs_handler();
    google::protobuf::uint64 gcsHandler = request->gcs_handler();

    if(requestType.compare(DATA_EXHG_START) == 0){

      NVMConfigBundle configBundle = request->config_bundle();
      this->config_bundle = &configBundle;
      CurrChainID = (uint32_t)configBundle.chain_id();
      std::string scriptSrc = configBundle.script_src();
      std::string scriptType = configBundle.script_type();
      std::string scriptHash = configBundle.script_hash();
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

      this->m_module_id = moduleID;
      //this->m_lcs_handler = (uintptr_t)lcsHandler;
      //this->m_gcs_handler = (uintptr_t)gcsHandler;

      // create engine and inject tracing instructions
      this->engine = CreateEngine();
      this->engine->limits_of_executed_instructions = configBundle.limits_exe_instruction();
      this->engine->limits_of_total_memory_size = configBundle.limits_total_mem_size();
      this->engine->lcs = lcsHandler;
      this->engine->gcs = gcsHandler;

      char* exeResult = nullptr;
      int ret = this->StartScriptExecution(this->engine, scriptSrc, scriptType, scriptHash, runnableSrc, moduleID, configBundle, exeResult);
      this->m_exe_result = exeResult;

      NVMDataResponse *response = new NVMDataResponse();
      NVMFinalResponse *finalResponse = new NVMFinalResponse();
      finalResponse->set_result(ret);
      finalResponse->set_not_null(this->m_exe_result != nullptr);

      if(this->m_exe_result == nullptr){
        this->m_exe_result = (char*)calloc(1, sizeof(char));
      }
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

  return grpc::Status::OK;
}

NVMCallbackResult* NVMEngine::Callback(void* handler, NVMCallbackResponse* callback_response){
    if(this->m_stm != nullptr){
        bool getResultFlag = false;
        NVMDataResponse *response = new NVMDataResponse();
        response->set_response_type(DATA_EXHG_CALL_BACK);
        response->set_response_indx(++this->m_response_indx);
        //response->set_lcs_handler(google::protobuf::uint64((uintptr_t)handler));
        response->set_lcs_handler((google::protobuf::uint64)this->GetCurrentEngineLcsHandler());
        response->set_gcs_handler((google::protobuf::uint64)this->GetCurrentEngineGcsHandler()); // gcs handler is not used by now
        response->set_allocated_callback_response(callback_response);
        this->m_stm->Write(*response);

        // wait for the result and return
        NVMDataRequest *request = new NVMDataRequest();        
        NVMCallbackResult* callback_result;
        std::string requestType;
        while(this->m_stm->Read(request)){
          requestType = request->request_type();
          if(requestType.compare(DATA_EXHG_CALL_BACK) == 0 || requestType.compare(DATA_EXHG_INNER_CALL) == 0){
            callback_result = (NVMCallbackResult*)&(request->callback_result());
            getResultFlag = true;
            if(FG_DEBUG){
              std::cout<<"-- Callback response from the GOLANG side with function: "<<callback_response->func_name()<<std::endl;
              // check gas cnt one by one
              std::cout<<"********** CALLBACK result of "<<callback_response->func_name()<<" is: "<<callback_result->result()<<std::endl;
              for(int i=0; i<callback_result->extra_size(); i++){
                std::cout<<"************* CALLBACK extra: "<<callback_result->extra(i)<<std::endl;
              }
            }

          }else if(requestType.compare(DATA_EXHG_INNER_CALL) == 0){
            // this is NOT the final result of this call, we need to create new engine and execute the contract firstly
            std::cout<<">>>>>>>NOW START to call the new contract"<<std::endl;
            callback_result = new NVMCallbackResult();
            NVMConfigBundle configBundle = request->config_bundle();
            std::string scriptSrc = configBundle.script_src();
            std::string scriptType = configBundle.script_type();
            std::string scriptHash = configBundle.script_hash();
            std::string runnableSrc = configBundle.runnable_src();
            std::string moduleID = configBundle.module_id();
            google::protobuf::uint64 maxLimitsOfExecutionInstructions = configBundle.max_limits_of_execution_instruction();
            google::protobuf::uint64 defaultTotalMemSize = configBundle.default_limits_of_total_mem_size();
            google::protobuf::uint64 limitsOfExecutionInstructions = configBundle.limits_exe_instruction();
            google::protobuf::uint64 totalMemSize = configBundle.limits_total_mem_size();
            std::string blockJson = configBundle.block_json();
            std::string txJson = configBundle.tx_json();

            std::string createResult;
            int ret = NVM_SUCCESS;
            char* exeResult = nullptr;
            V8Engine* newEngine = this->CreateInnerContractEngine(scriptType, scriptHash, scriptSrc, moduleID, createResult);
            if(newEngine != nullptr){
              newEngine->limits_of_executed_instructions = configBundle.limits_exe_instruction();
              newEngine->limits_of_total_memory_size = configBundle.limits_total_mem_size();
              newEngine->lcs = (uintptr_t)request->lcs_handler();
              newEngine->gcs = (uintptr_t)request->gcs_handler();

              // NOTE: this creates a new thread to execute the new contract
              ret = this->StartScriptExecution(newEngine, scriptSrc, scriptType, scriptHash, runnableSrc, moduleID, configBundle, exeResult);
            }
            
            getResultFlag = true;
            callback_result->set_result(std::string(exeResult));
          }
          break;
        }

        free(request);
        if(!getResultFlag)
          callback_result = nullptr;
        return callback_result;

    }else{
        LogErrorf("Streaming object is not NULL");
    }
    return nullptr;
}



void NVMEngine::LocalTest(){
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


V8Engine* NVMEngine::CreateInnerContractEngine(
    const std::string& scriptType, 
    const std::string& scriptHash, 
    const std::string& innerContractSrc, 
    const std::string& moduleID, 
    std::string& createResult){

    if(this->m_inner_engines == nullptr)
      this->m_inner_engines = std::unique_ptr<std::stack<V8Engine*>>(new std::stack<V8Engine*>());

    V8Engine* inner_engine = CreateEngine();
    // set limitation for the new engine
    if(inner_engine != nullptr)
      this->m_inner_engines->push(engine);
    ReadMemoryStatistics(this->engine);

    // transpile script and inject tracing code if necessary
    int runnableSourceResult = this->GetRunnableSourceCode(inner_engine, scriptType, scriptHash, innerContractSrc);
    if(runnableSourceResult != 0){
      LogErrorf("Failed to get runnable source code");
      createResult = std::string("Failed to get runnable source code");
      return nullptr;
    }

    AddModule(this->engine, moduleID.c_str(), this->m_traceable_src.c_str(), this->m_traceale_src_line_offset);

    return inner_engine;
}

// start one level of inner contract call, check the resource usage firstly
//const void NVMEngine::StartInnerContractCall(const std::string& address, const std::string& valueStr, const std::string& funcName, const std::string& args){

const std::string NVMEngine::ConfigBundleToString(NVMConfigBundle& configBundle){
  std::string res("chain id: " + std::to_string(configBundle.chain_id()) + ", script src: " + configBundle.script_src() + ", script type: " + configBundle.script_type()
    +  ", script hash: " + configBundle.script_hash() + ", runnable src: " + configBundle.runnable_src() + ", module id: " + configBundle.module_id());
  return res;
}