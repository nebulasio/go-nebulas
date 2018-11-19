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

void ExecuteScript(const char *filename, V8ExecutionDelegate delegate) {
  void *lcsHandler = CreateStorageHandler();
  void *gcsHandler = CreateStorageHandler();

  V8Engine *e = CreateEngine();

  size_t size = 0;
  int lineOffset = 0;
  char *source = readFile(filename, &size);
  if (source == NULL) {
    fprintf(stderr, "%s is not found.\n", filename);
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


class NVMEngine final: public nvm::NVMService::Service{
  
  public:
    explicit NVMEngine(const int currency){
      //TODO: specify how many threads we should start
    }

    grpc::Status DeploySmartContract(grpc::ServerContext* ctx, const nvm::NVMDeployRequest* request, nvm::NVMResponse* response) override {
      std::cout<<"Request script source is: "<<request->script_src()<<std::endl;
      std::cout<<"Request address is: "<<request->from_addr()<<std::endl;

      nvm::NVMResponse* new_response = new nvm::NVMResponse();

      new_response->set_result(101);
      new_response->set_msg("Deployed successfully!");

      return grpc::Status::OK;
    }

    grpc::Status NVMDataExchange(grpc::ServerContext* ctx, grpc::ServerReaderWriter<nvm::NVMResponse, nvm::NVMDataRequest> *stm) override {
      //read the request firstly
      nvm::NVMDataRequest *request = new nvm::NVMDataRequest();
      stm->Read(request);
      std::cout<<"The request script souce "<<request->script_src()<<std::endl;
      std::cout<<"The function name is: "<<request->function_name()<<std::endl;

      //send request to client if necessary
      nvm::NVMResponse* response = new nvm::NVMResponse();
      response->set_result(202);
      response->set_msg("Send request to client!");

      stm->Write(*response);

      return grpc::Status::OK;
    }

  private:
    int m_currency = 1;   // default concurrency number

};

void Initialization(){
  Initialize();
  InitializeLogger(logFunc);
  InitializeRequireDelegate(RequireDelegateFunc, AttachLibVersionDelegateFunc);
  InitializeExecutionEnvDelegate(AttachLibVersionDelegateFunc);
  InitializeStorage(StorageGet, StoragePut, StorageDel);
  InitializeBlockchain(GetTxByHash, GetAccountState, Transfer, VerifyAddress, GetPreBlockHash, GetPreBlockSeed);
  InitializeEvent(eventTriggerFunc);
}

void RunServer(){

    std::string engine_addr("127.0.0.1:11199");

    NVMEngine* engine = new NVMEngine(1);

    grpc::ServerBuilder builder;
    builder.AddListeningPort(engine_addr, grpc::InsecureServerCredentials());
    builder.RegisterService(engine);
    std::unique_ptr<grpc::Server> server(builder.BuildAndStart());
    std::cout<<"V8 engine server is listening on: "<<engine_addr<<std::endl;
    
    server->Wait();

}

int main(int argc, const char *argv[]) {
  
  Initialization();

  RunServer();

  return 0;
}
