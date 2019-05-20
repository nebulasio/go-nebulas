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
// <http://www.gnu.org/licenses/>.

#pragma once

#include <unistd.h>
#include "engine.h"
#include "compatibility.h"
#include "lib/blockchain.h"
#include "lib/file.h"
#include "lib/log_callback.h"
#include "lib/logger.h"

#include "pb/nvm.grpc.pb.h"
#include "pb/nvm.pb.h"
#include "callback/memory_modules.h"
#include "callback/memory_storage.h"
#include "callback/blockchain_modules.h"
#include "callback/event_trigger.h"
#include "callback/crypto_modules.h"

#include <thread>
#include <vector>
#include <stack>
#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <string>
#include <iostream>
#include <fstream>
#include <sstream>
#include <mutex>

#include <grpc/grpc.h>
#include <grpcpp/server.h>
#include <grpcpp/server_builder.h>
#include <grpcpp/server_context.h>
#include <glog/logging.h>

#include "engine.h"
#include "engine_int.h"
#include "engine_conf.h"
#include "lru_map.h"
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

#define FG_DEBUG true

//static const uint32_t NVM_CURRENCY_LEVEL = 1;

typedef struct{
  std::string source;
  uint64_t sourceLineOffset;
  std::string traceableSource;
  uint64_t traceableSourceLineOffset;
}CacheSrcItem;

typedef struct {
  std::string source;
  int lineOffset;
}SourceInfo;

namespace SNVM{

  class NVMEngine final: public NVMService::Service{

    public:
      explicit NVMEngine(const int concurrency):
        m_concurrency_scale(concurrency),
        m_src_offset(0){

        srcModuleCache = std::unique_ptr<LRU_MAP<std::string, CacheSrcItem>>(new LRU_MAP<std::string, CacheSrcItem>());
        engineSrcModules = std::unique_ptr<LRU_MAP<std::string, SourceInfo>>(new LRU_MAP<std::string, SourceInfo>());
        lib_content_cache = std::unique_ptr<std::unordered_map<std::string, std::string>>(new std::unordered_map<std::string, std::string>);
        m_compat_manager = new SNVM::CompatManager(m_chain_id);
      }

      ~NVMEngine(){
        srcModuleCache->clear();
        engineSrcModules->clear();
        lib_content_cache->clear();
        if(m_compat_manager != nullptr)
          delete m_compat_manager;
      }

      // Reset data structures before the next time running to avoid status mismatch
      void ResetRuntimeStatus();

      int GetRunnableSourceCode(V8Engine*, const std::string&, const std::string&, const std::string&);

      void ReadExeStats(NVMStatsBundle *);

      int StartScriptExecution(V8Engine*, const std::string&, const std::string&, const std::string&, const std::string&,
              const std::string&, const NVMConfigBundle&, char*&);

      NVMCallbackResult* Callback(void*, NVMCallbackResponse*, bool);

      std::string FetchLibContent(const char*);

      std::string FetchContractSrc(const char*, size_t*);

      std::string AttachNativeJSLibVersion(const char*);

      void AddContractSrcToModules(const char*, const char*, size_t);


      // By default, it returns this->engine's lcshandler, or returns the lcshandler of the latest engine pushed in the inner engine stack
      uintptr_t GetCurrentEngineLcsHandler();

      uintptr_t GetCurrentEngineGcsHandler();

      const std::string ConfigBundleToString(NVMConfigBundle&);

      void LocalTest();

      grpc::Status SmartContractCall(grpc::ServerContext*, grpc::ServerReaderWriter<NVMDataResponse, NVMDataRequest>*) override;

      V8Engine* CreateInnerContractEngine(const std::string&, const std::string&, const std::string&, const std::string&, uint64_t, std::string&);

      inline void SetChainID(uint32_t chain_id){
        m_chain_id = chain_id;
        if(m_compat_manager != nullptr)
          m_compat_manager->SetChainID(chain_id);
      }
      

    private:
      const std::string TS_TYPE = "ts";
      const std::string JS_TYPE = "js";
      const std::string DATA_EXHG_START = "start";
      const std::string DATA_EXHG_CALL_BACK = "callback";
      const std::string DATA_EXHG_FINAL = "final";
      const std::string DATA_EXHG_INNER_CALL = "innercall";

      std::mutex m_mutex;
      int m_concurrency_scale = 1;              // default concurrency number
      int m_src_offset = 0;                     // default source code offset
      int m_allow_usage = 1;                    // default allow usage
      int m_response_indx = 0;                  // index of the data request/response pair
      uint32_t m_chain_id = MainNetID;          // default chaind id

      SNVM::CompatManager* m_compat_manager=nullptr;                      // compatibility manager

      V8Engine* engine = nullptr;                                         // default engine
      char* m_exe_result = nullptr;                                       // contract execution result
      char* m_inner_exe_result = nullptr;                                 // inner contract call execution result
      NVMConfigBundle* config_bundle = nullptr;                           // default config bundle
      std::unique_ptr<std::stack<V8Engine*>> m_inner_engines = nullptr;   // stack for keeping engines created because of inner contract calls
      grpc::ServerReaderWriter<NVMDataResponse, NVMDataRequest> *m_stm;   // stream used to send request from server
      std::unique_ptr<LRU_MAP<std::string, CacheSrcItem>> srcModuleCache; // LRU map: src code hash --> traceable js src code & offset
      std::unique_ptr<LRU_MAP<std::string, SourceInfo>> engineSrcModules; // LRU map: engine address + contract name --> SourceInfo

      std::unique_ptr<std::unordered_map<std::string, std::string>> lib_content_cache;  // source code cache for js libs
  };


  const NVMCallbackResult* DataExchangeCallback(void*, NVMCallbackResponse*, bool inner_call_flag=false);
  void AddContractSrcToModules(const char*, const char*, size_t);
  std::string FetchContractSrcFromModules(const char*, size_t*);
  std::string FetchNativeJSLibContentFromCache(const char*);
  std::string AttachNativeJSLibVersion(const char*);
}

extern SNVM::NVMEngine* gNVMEngine;