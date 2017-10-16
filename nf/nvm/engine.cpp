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

#include "engine.h"
#include "_cgo_export.h"

#include <llvm/ExecutionEngine/ExecutionEngine.h>
#include <llvm/ExecutionEngine/GenericValue.h>
#include <llvm/ExecutionEngine/SectionMemoryManager.h>
#include <llvm/IR/LLVMContext.h>
#include <llvm/IR/LegacyPassManager.h>
#include <llvm/IR/Module.h>
#include <llvm/IRReader/IRReader.h>
#include <llvm/Support/SourceMgr.h>
#include <llvm/Support/TargetSelect.h>
#include <llvm/Support/raw_ostream.h>
#include <llvm/Transforms/Scalar.h>
#include <llvm/Transforms/Scalar/GVN.h>

#include <stdio.h>
#include <stdlib.h>

using namespace llvm;

void Initialize() {
  // Initialization.
  InitializeNativeTarget();
  InitializeNativeTargetAsmParser();
  InitializeNativeTargetAsmPrinter();
  LLVMLinkInMCJIT();
}

Engine *CreateEngine(const char *irPath) {
  // Parse IR.
  LLVMContext *context = new LLVMContext();
  SMDiagnostic err;

  std::unique_ptr<Module> pModule =
      parseIRFile(StringRef(irPath), err, *context);
  Module *module = pModule.get();
  if (module == nullptr) {
    LogError(err.getMessage().data());
    delete context;
    return NULL;
  }

  // Create PassManager.
  legacy::PassManager *passMgr = new legacy::PassManager();
  passMgr->add(createConstantPropagationPass());
  passMgr->add(createInstructionCombiningPass());
  passMgr->add(createPromoteMemoryToRegisterPass());
  passMgr->add(createCFGSimplificationPass());
  passMgr->add(createDeadCodeEliminationPass());
  passMgr->add(createGVNPass());
  passMgr->run(*module);

  if (false) {
    // TODO: @robin, invalid IR file should be quit.
    LogError("running pass failed.");
    delete passMgr;
    delete context;
    return NULL;
  }

  // Create EngineBuilder.
  std::string errMsg;

  EngineBuilder *builder = new EngineBuilder(std::move(pModule));
  builder->setErrorStr(&errMsg);
  builder->setEngineKind(EngineKind::JIT);
  builder->setUseOrcMCJITReplacement(false);

  // Enable MCJIT.
  SectionMemoryManager *rtDyldMM = new SectionMemoryManager();
  builder->setMCJITMemoryManager(
      std::unique_ptr<RTDyldMemoryManager>(rtDyldMM));

  builder->setOptLevel(CodeGenOpt::Default);

  // Create ExecutionEngine.
  ExecutionEngine *engine = builder->create();
  if (engine == nullptr) {
    LogError(errMsg.c_str());
    delete rtDyldMM;
    delete builder;
    delete passMgr;
    delete context;
    return NULL;
  }

  // Run static contructors.
  engine->finalizeObject();
  engine->runStaticConstructorsDestructors(false);

  Engine *e = static_cast<Engine *>(calloc(1, sizeof(Engine)));
  e->llvm_context = context;
  e->llvm_pass_manager = passMgr;
  e->llvm_builder = builder;
  e->llvm_mem_manager = rtDyldMM;

  e->llvm_engine = engine;
  e->llvm_module = module;
  return e;
}

void DeleteEngine(Engine *e) {
  // TODO: release llvm resource by call proper llvm dispose function.
  delete static_cast<SectionMemoryManager *>(e->llvm_mem_manager);
  delete static_cast<EngineBuilder *>(e->llvm_builder);
  delete static_cast<legacy::PassManager *>(e->llvm_pass_manager);
  delete static_cast<LLVMContext *>(e->llvm_context);
  free(e);
}

int RunFunction(Engine *e, const char *funcName, size_t len,
                const uint8_t *data) {
  Module *module = static_cast<Module *>(e->llvm_module);
  ExecutionEngine *engine = static_cast<ExecutionEngine *>(e->llvm_engine);

  Function *func = module->getFunction(StringRef(funcName));
  if (func == nullptr) {
    char msg[128];
    snprintf(msg, 128, "%s function not found.", funcName);
    LogError(msg);
    return -1;
  }

  (void)engine->getPointerToFunction(func);

  // Run func.
  std::vector<GenericValue> args;

  GenericValue argLen;
  argLen.IntVal = APInt(sizeof(size_t) * 8, len);
  args.push_back(argLen);

  GenericValue argData;
  argData.PointerVal = (PointerTy)data;
  args.push_back(argData);
  engine->runFunction(func, args);

  return 0;
}
