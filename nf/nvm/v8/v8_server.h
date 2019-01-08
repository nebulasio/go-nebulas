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

#pragma once

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
#include "engine_conf.h"
#include "lib/tracing.h"
#include "lib/typescript.h"
#include "lib/logger.h"
#include "lib/nvm_error.h"
#include "sha256.h"
//#include "lru_cache.h"

#include <assert.h>
#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include <thread>
#include <sys/time.h>
#include <unistd.h>

// constants
static const uint32_t NVM_CURRENCY_LEVEL = 1;

class NVMEngine;

NVMEngine *gNVMEngine;

typedef struct{
  std::string source;
  uint64_t sourceLineOffset;
  std::string traceableSource;
  uint64_t traceableSourceLineOffset;
}CacheSrcItem;

std::unique_ptr<std::map<std::string, CacheSrcItem>> srcModuleCache;

class NVMEngine final: public NVMService::Service{

  public:

    explicit NVMEngine(const int concurrency){
      //TODO: specify how many threads we should start and do Initialization

      m_concurrency_scale = concurrency;
      m_src_offset = 0;
      srcModuleCache = std::unique_ptr<std::map<std::string, CacheSrcItem>>(new std::map<std::string, CacheSrcItem>());

      std::cout<<"Now we're starting the engine"<<std::endl;
    }

    bool GetRunnableSourceCode(const std::string&, std::string&);

    void ReadExeStats(NVMStatsBundle *);

    int StartScriptExecution(std::string&, const std::string&, const std::string&, const std::string&, const NVMConfigBundle&);

    grpc::Status SmartContractCall(grpc::ServerContext*, grpc::ServerReaderWriter<NVMDataResponse, NVMDataRequest>*) override;

    inline void* Callback(char* func_name, char** args){
        if(this->m_stm != NULL){
            // compose server request
            NVMDataRequest* request = new NVMDataRequest();
            //this->m_stm->Write(request);
        }
    }

  private:
    int m_concurrency_scale = 1;              // default concurrency number

    int m_src_offset = 0;                     // default source code offset
    int m_allow_usage = 1;                    // default allow usage
    std::string m_module_id = "contract.js";  // default module ID to be used
    std::string m_traceable_src;              // source code after injection
    std::string m_runnable_src;               // runnable source code
    uint64_t m_traceale_src_line_offset = 0;   // set to be 0 by default
    uintptr_t m_lcs_handler = 0;                // lcs handler
    uintptr_t m_gcs_handler = 0;                // gcs handler
    
    V8Engine* engine = nullptr;                    // default engine
    char* m_exe_result = nullptr;                  // contract execution result
    
    // constants for defining contract source type
    const std::string TS_TYPE = "ts";
    const std::string JS_TYPE = "js";
    const std::string DATA_REQUEST_START = "start";
    const std::string DATA_REQUEST_CALL_BACK = "callback";
    const std::string DATA_RESPONSE_FINAL = "final";

    //static lru_cache<std::string, std::string, CacheCleanCounter=3000> m_lru_cache;
    grpc::ServerReaderWriter<NVMDataResponse, NVMDataRequest> *m_stm;    // stream used to send request from server

};

inline void* DataExchangeCallback(char* func_name, char** args){
    if(gNVMEngine != NULL){
        void* result = gNVMEngine->Callback(func_name, args);
        return result;
    }
    return NULL;
}