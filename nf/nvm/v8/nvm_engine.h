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
// <http://www.gnu.org/licenses/>.
//
// Author: Samuel Chen <samuel.chen@nebulas.io>

#pragma once

#include "engine.h"
#include "engine_int.h"
#include "engine_conf.h"
#include "lru_map.h"
#include "compatibility.h"
#include "lib/tracing.h"
#include "lib/typescript.h"
#include "lib/logger.h"
#include "lib/nvm_error.h"
#include "lib/blockchain.h"
#include "lib/file.h"
#include "lib/log_callback.h"
#include "pb/nvm.grpc.pb.h"
#include "pb/nvm.pb.h"
#include "callback/memory_modules.h"
#include "callback/memory_storage.h"
#include "callback/blockchain_modules.h"
#include "callback/event_trigger.h"
#include "callback/crypto_modules.h"

#include <unistd.h>
#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <assert.h>
#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/time.h>
#include <unistd.h>

#include <vector>
#include <stack>
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

#define FG_DEBUG true
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

  const std::string TS_TYPE = "ts";
  const std::string JS_TYPE = "js";
  const std::string DATA_EXHG_START = "start";
  const std::string DATA_EXHG_CALL_BACK = "callback";
  const std::string DATA_EXHG_FINAL = "final";
  const std::string DATA_EXHG_INNER_CALL = "innercall";
  const uint32_t SRC_MODULE_CACHE_SIZE = 128;

  class NVMDaemon;

  // Each context is bundled with a smart contract execution instance
  class SCContext {
    public:
      explicit SCContext(NVMDaemon* daemon){
        m_daemon = daemon;
        m_src_offset = 0;
        engineSrcModules = std::unique_ptr<std::unordered_map<std::string, SourceInfo>>(new std::unordered_map<std::string, SourceInfo>());
      }

      ~SCContext(){
        engineSrcModules->clear();
      }

      int GetRunnableSourceCode(V8Engine*, const std::string&, const std::string&, const std::string&);
      void ReadExeStats(NVMStatsBundle *);
      int StartScriptExecution(const std::string&, const std::string&, const std::string&, const std::string&,
              const std::string&, char*&);

      NVMCallbackResult* Callback(void*, NVMCallbackResponse*, bool);
      std::string FetchEngineSrc(const std::string&, size_t*);
      void AddEngineSrcModule(void *, const char*, const char*, size_t);

      V8Engine* GetCurrentV8EngineInstance();
      uintptr_t GetCurrentEngineLcsHandler();
      uintptr_t GetCurrentEngineGcsHandler();
      const std::string ConfigBundleToString(NVMConfigBundle&);

      V8Engine* CreateInnerContractEngine(const std::string&, const std::string&, 
                  const std::string&, const std::string&, uint64_t, std::string&);

      void PopInnerEngine(V8Engine*);

      inline void SetChainID(uint32_t chain_id){ m_chain_id = chain_id; }
      inline uint32_t GetChainID(){ return m_chain_id; }

      inline void SetStream(grpc::ServerReaderWriter<NVMDataResponse, NVMDataRequest> *stream){m_stream=stream;}
      inline grpc::ServerReaderWriter<NVMDataResponse, NVMDataRequest>* GetStream(){return m_stream;}

      inline uint64_t GetBlockHeight(){
        if(config_bundle != nullptr)
          return config_bundle->block_height();
        return 0L;
      }
      
      inline std::string GetMetaVersion(){
        if(config_bundle != nullptr)
          return config_bundle->meta_version();
        // return the default lowest version
        return "1.0.0";
      }

      inline void SetV8Engine(V8Engine* engine){this->engine=engine;}
      inline V8Engine* GetV8Engine(){return this->engine;}

      inline void SetConfigBundle(NVMConfigBundle* bundle){config_bundle = bundle;}
      inline NVMConfigBundle* GetConfigBundle(){return config_bundle;}


    private:
      int m_src_offset = 0;                                                         // default source code offset
      int m_allow_usage = 1;                                                        // default allow usage
      int m_response_indx = 0;                                                      // index of the data request/response pair
      uint32_t m_chain_id = MainNetID;                                              // default chaind id

      V8Engine* engine = nullptr;                                                   // default engine
      char* m_exe_result = nullptr;                                                 // contract execution result
      char* m_inner_exe_result = nullptr;                                           // inner contract call execution result
      NVMConfigBundle* config_bundle = nullptr;                                     // default config bundle
      NVMDaemon* m_daemon = nullptr;                                                // backward pointer to daemon

      std::unique_ptr<std::stack<V8Engine*>> m_inner_engines = nullptr;                 // stack for keeping engines created because of inner contract calls
      grpc::ServerReaderWriter<NVMDataResponse, NVMDataRequest> *m_stream = nullptr;    // stream used to send/receive request
      std::unique_ptr<std::unordered_map<std::string, SourceInfo>> engineSrcModules;    // clear it before each smart contract call
  };



  class NVMDaemon final: public NVMService::Service, public CompatManager{

    public:
      explicit NVMDaemon(const int concurrency):
        m_concurrency_scale(concurrency){
        
        sc_ctxs = std::unique_ptr<std::unordered_map<V8Engine*, SCContext*>>(new std::unordered_map<V8Engine*, SCContext*>());
        traceable_src_cache = std::unique_ptr<std::unordered_map<std::string, CacheSrcItem>>(new std::unordered_map<std::string, CacheSrcItem>());
        lib_content_cache = std::unique_ptr<std::unordered_map<std::string, std::string>>(new std::unordered_map<std::string, std::string>());
        InitializeLibVersionManager();
      }

      ~NVMDaemon(){
        sc_ctxs->clear();
        traceable_src_cache->clear();
        lib_content_cache->clear();
      }

      void LocalTest();
      grpc::Status SmartContractCall(grpc::ServerContext*, grpc::ServerReaderWriter<NVMDataResponse, NVMDataRequest>*) override;
      std::string FetchLibContent(const char*);
      std::string FetchContractSrc(const std::string&, size_t*);
      std::string AttachNativeJSLibVersion(const char*, uint64_t, std::string, uint32_t);
      std::string AddTraceableSrcModule(const std::string&, const char*, size_t);

      inline size_t TraceableSrcModuleCacheSize(){
        if(traceable_src_cache != nullptr)
          return traceable_src_cache->size();
        else
          return (size_t)0;
      }

      inline bool IsTraceableSrcModuleExist(const std::string& srcHash){
        if(traceable_src_cache == nullptr)
          return false;
        if(traceable_src_cache->find(srcHash) == traceable_src_cache->end())
          return false;
        return true;
      }

      inline CacheSrcItem FindTraceableSrcModuleFromCache(const std::string& srcHash){
        auto item = traceable_src_cache->find(srcHash);
        return item->second;
      }

      inline void AddIntoTraceableSrcModuleCache(std::string scriptHash, CacheSrcItem newItem){
        if(traceable_src_cache != nullptr){
          m_mutex.lock();
          traceable_src_cache->insert(std::pair<std::string, CacheSrcItem>(scriptHash, newItem));
          m_mutex.unlock();
        }
      }

      inline void ClearTraceableSrcModuleCache(){
        if(traceable_src_cache != nullptr){
          m_mutex.lock();
          traceable_src_cache->clear();
          m_mutex.unlock();
        }
      }

      inline SCContext* GetSCContext(V8Engine* engine){
        if(sc_ctxs != nullptr){
          auto item = sc_ctxs->find(engine);
          if(item != sc_ctxs->end())
            return item->second;
        }
        return nullptr;
      }

      inline void AddSCContext(V8Engine* engine, SCContext* ctx){
        if(sc_ctxs != nullptr){
          m_mutex.lock();
          sc_ctxs->insert(std::pair<V8Engine*, SCContext*>(engine, ctx));
          m_mutex.unlock();
        }
      }

      inline void RemoveSCContext(V8Engine* engine){
        if(engine != nullptr && sc_ctxs != nullptr){
          m_mutex.lock();
          sc_ctxs->erase(engine);
          m_mutex.unlock();
        }
      }

      inline V8Engine* CreateV8Engine(uint64_t block_height, uint32_t chain_id){
        V8Engine* new_engine = CreateEngine();
        new_engine->timeout = GetTimeoutConfig(block_height, chain_id);
        return new_engine;
      }

    private:
      std::mutex m_mutex;
      int m_concurrency_scale = 1;                                                              // default concurrency number
      bool is_sc_running = false;                                                               // flag indicating if a sc is running already
      std::unique_ptr<std::unordered_map<V8Engine*, SCContext*>> sc_ctxs;                       // smart contract contexts
      std::unique_ptr<std::unordered_map<std::string, CacheSrcItem>> traceable_src_cache;       // LRU map: src code hash --> traceable js src code & offset
      std::unique_ptr<std::unordered_map<std::string, std::string>> lib_content_cache;          // source code cache for js libs
  };


  // utilities
  const NVMCallbackResult* DataExchangeCallback(V8Engine*, void*, NVMCallbackResponse*, bool inner_call_flag=false);
  std::string FetchEngineContractSrcFromModules(V8Engine*, const std::string&, size_t*);
  std::string FetchNativeJSLibContentFromCache(const char*);
  std::string AttachNativeJSLibVersion(V8Engine*, const char*);
}

extern SNVM::NVMDaemon* gNVMDaemon;